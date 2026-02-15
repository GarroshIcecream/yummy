package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// llmIngredient mirrors the JSON shape we ask the LLM to produce.
type llmIngredient struct {
	Amount   string `json:"amount"`
	Unit     string `json:"unit"`
	Name     string `json:"name"`
	Details  string `json:"details"`
	BaseName string `json:"base_name"`
}

// maxParallelLLM limits the number of concurrent Ollama requests to avoid
// overwhelming the local model server.
const maxParallelLLM = 4

const singleIngredientPrompt = `Extract the following raw ingredient string into a JSON object with these fields:
- "amount"    : the numeric quantity (e.g. "2", "0.5", "1/2"). Empty string if none.
- "unit"      : the measurement unit normalised to a short form (e.g. "cup", "tsp", "tbl", "gram", "ounce", "pound", "ml"). Empty string if none.
- "name"      : the ingredient name without quantity, unit, or parenthetical notes.
- "details"   : any parenthetical or extra qualifier (e.g. "divided", "at room temperature"). Empty string if none.
- "base_name" : the core ingredient noun(s) stripped of adjectives, preparation methods and modifiers. Used for highlighting in instructions. Examples: "dried thyme" → "thyme", "freshly grated parmesan cheese" → "parmesan cheese", "lean ground beef" → "ground beef", "ground black pepper to taste" → "black pepper", "onions, chopped" → "onions", "olive oil" → "olive oil", "all-purpose flour" → "flour".

Rules:
1. Return ONLY a single JSON object — no markdown fences, no explanation.
2. Normalise unicode fractions (½ → 0.5, ¼ → 0.25, ¾ → 0.75, ⅓ → 0.33, ⅔ → 0.67).

Ingredient: %s

JSON output:`

const baseNamePrompt = `You are a structured-data extraction assistant.
Given a list of ingredient names from a recipe, extract the core ingredient noun(s) for each — stripped of adjectives, preparation methods, and modifiers. The base name should be what you would search for in an instruction text to highlight that ingredient.

Examples:
- "dried thyme" → "thyme"
- "freshly grated parmesan cheese" → "parmesan cheese"
- "lean ground beef" → "ground beef"
- "ground black pepper to taste" → "black pepper"
- "onions, chopped" → "onions"
- "olive oil" → "olive oil"
- "all-purpose flour" → "flour"
- "fines herbs" → "fines herbs"
- "salt to taste" → "salt"
- "garlic, minced" → "garlic"
- "fresh lime juice" → "lime juice"

Return ONLY a JSON array of strings (one base name per input), in the same order. No markdown fences, no explanation.

Input ingredient names:
%s

JSON output:`

// parseResult holds the result of a single parallel ingredient parse.
type parseResult struct {
	index int
	ing   Ingredient
	err   error
}

// parseSingleIngredient sends a single raw ingredient string to the LLM and
// returns a structured Ingredient. Uses JSON mode for guaranteed valid output.
func parseSingleIngredient(ctx context.Context, llm llms.Model, raw string) (llmIngredient, error) {
	prompt := fmt.Sprintf(singleIngredientPrompt, raw)

	resp, err := llm.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	},
		llms.WithTemperature(0.0),
		llms.WithMaxTokens(512),
		llms.WithJSONMode(),
	)
	if err != nil {
		return llmIngredient{}, fmt.Errorf("ollama generate: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Content == "" {
		return llmIngredient{}, fmt.Errorf("empty response from LLM")
	}

	body := strings.TrimSpace(resp.Choices[0].Content)
	body = stripCodeFences(body)

	var parsed llmIngredient
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return llmIngredient{}, fmt.Errorf("json decode: %w (body: %s)", err, body)
	}

	return parsed, nil
}

// ParseIngredientsWithLLM sends parallel requests to a local Ollama model —
// one per ingredient — and returns structured Ingredient values. Each request
// uses JSON mode so the model is forced to return valid JSON.
//
// If an individual ingredient fails to parse via LLM, it falls back to the
// regex-based ParseIngredient for that ingredient only.
//
// modelName is the Ollama model to use (e.g. "gemma3:4b").
func ParseIngredientsWithLLM(ctx context.Context, rawIngredients []string, modelName string) ([]Ingredient, error) {
	if len(rawIngredients) == 0 {
		return []Ingredient{}, nil
	}

	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return nil, fmt.Errorf("create ollama client: %w", err)
	}

	// Filter out empty lines upfront, keeping track of original indices.
	type indexedRaw struct {
		idx int
		raw string
	}
	var work []indexedRaw
	for i, s := range rawIngredients {
		line := strings.TrimSpace(s)
		if line != "" {
			work = append(work, indexedRaw{idx: i, raw: line})
		}
	}
	if len(work) == 0 {
		return []Ingredient{}, nil
	}

	results := make([]parseResult, len(work))
	sem := make(chan struct{}, maxParallelLLM)
	var wg sync.WaitGroup

	for i, item := range work {
		wg.Add(1)
		go func(i int, raw string, origIdx int) {
			defer wg.Done()
			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			parsed, err := parseSingleIngredient(ctx, llm, raw)
			if err != nil {
				slog.Warn("LLM parse failed for ingredient, falling back to regex",
					"ingredient", raw, "error", err)
				// Fall back to regex for this ingredient.
				regexIng, regexErr := ParseIngredient(raw)
				if regexErr != nil {
					results[i] = parseResult{index: origIdx, err: fmt.Errorf("both LLM and regex failed for %q: llm: %w, regex: %v", raw, err, regexErr)}
					return
				}
				results[i] = parseResult{index: origIdx, ing: regexIng}
				return
			}

			ing := Ingredient{
				Amount:   strings.TrimSpace(parsed.Amount),
				Unit:     normaliseUnit(strings.TrimSpace(parsed.Unit)),
				Name:     strings.TrimSpace(parsed.Name),
				Details:  strings.TrimSpace(parsed.Details),
				BaseName: strings.TrimSpace(parsed.BaseName),
			}
			results[i] = parseResult{index: origIdx, ing: ing}
		}(i, item.raw, item.idx)
	}

	wg.Wait()

	// Collect results in original order, skipping total failures.
	ingredients := make([]Ingredient, 0, len(results))
	var errs []string
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err.Error())
			continue
		}
		if r.ing.Name == "" {
			continue
		}
		ingredients = append(ingredients, r.ing)
	}

	if len(errs) > 0 {
		slog.Warn("Some ingredients failed to parse", "failures", len(errs), "succeeded", len(ingredients))
	}

	slog.Info("Parsed ingredients with LLM", "count", len(ingredients), "model", modelName)
	return ingredients, nil
}

// ExtractBaseNamesWithLLM takes already-parsed ingredients and asks an Ollama
// model to extract the core ingredient noun(s) for each. It mutates the
// BaseName field in-place on the provided slice.
//
// This is the lighter-weight alternative: it only extracts base names rather
// than doing a full structured parse, so it can be used even when ingredients
// were parsed via the regex path.
func ExtractBaseNamesWithLLM(ctx context.Context, ingredients []Ingredient, modelName string) error {
	if len(ingredients) == 0 {
		return nil
	}

	llm, err := ollama.New(ollama.WithModel(modelName))
	if err != nil {
		return fmt.Errorf("create ollama client: %w", err)
	}

	// Build numbered list of ingredient names.
	var sb strings.Builder
	for i, ing := range ingredients {
		name := strings.TrimSpace(ing.Name)
		if name == "" {
			name = "(empty)"
		}
		fmt.Fprintf(&sb, "%d. %s\n", i+1, name)
	}

	prompt := fmt.Sprintf(baseNamePrompt, sb.String())

	resp, err := llm.GenerateContent(ctx, []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	},
		llms.WithTemperature(0.0),
		llms.WithMaxTokens(2048),
		llms.WithJSONMode(),
	)
	if err != nil {
		return fmt.Errorf("ollama generate: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Content == "" {
		return fmt.Errorf("empty response from LLM")
	}

	body := stripCodeFences(strings.TrimSpace(resp.Choices[0].Content))

	var baseNames []string
	if err := json.Unmarshal([]byte(body), &baseNames); err != nil {
		slog.Error("LLM base name extraction failed", "error", err, "body", body)
		return fmt.Errorf("json decode: %w", err)
	}

	// Apply base names — only where we got a response.
	for i := range ingredients {
		if i < len(baseNames) {
			bn := strings.TrimSpace(baseNames[i])
			if bn != "" {
				ingredients[i].BaseName = bn
			}
		}
	}

	slog.Info("Extracted base names with LLM", "count", len(baseNames), "model", modelName)
	return nil
}

// stripCodeFences removes ```json ... ``` wrappers.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		// Remove opening fence line.
		if idx := strings.Index(s, "\n"); idx != -1 {
			s = s[idx+1:]
		}
		// Remove closing fence.
		if idx := strings.LastIndex(s, "```"); idx != -1 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}

// normaliseUnit maps common unit strings to the short forms used by the app.
func normaliseUnit(u string) string {
	if u == "" {
		return ""
	}
	low := strings.ToLower(u)
	if mapped, ok := CorpusMeasuresMap[low]; ok {
		return mapped
	}
	// Accept the LLM output as-is if it's not in our map.
	return low
}

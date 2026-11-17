package chatgpt

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type streamResult struct {
	text         string
	stopReason   string
	inputTokens  int
	outputTokens int
}

type streamEvent struct {
	Type     string `json:"type"`
	Delta    string `json:"delta"`
	Message  string `json:"message"`
	Response *struct {
		Usage *struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
		IncompleteDetails *struct {
			Reason string `json:"reason"`
		} `json:"incomplete_details"`
	} `json:"response"`
}

// readStream consumes a Responses API SSE stream. Text deltas are forwarded
// to stream (if set) and accumulated. When the text reaches one of the stop
// words it is cut there and reading ends; the stream is held back by the
// length of the longest stop word so a stop word split across deltas never
// reaches the stream callback.
func readStream(ctx context.Context, body io.Reader, stopWords []string, stream func(context.Context, []byte) error) (streamResult, error) {
	var res streamResult
	var full strings.Builder
	emitted := 0
	holdBack := 0
	for _, s := range stopWords {
		holdBack = max(holdBack, len(s)-1)
	}

	emit := func(upTo int) error {
		if stream == nil || upTo <= emitted {
			return nil
		}
		text := full.String()
		// Never split a UTF-8 sequence across chunks.
		for upTo > emitted && upTo < len(text) && !utf8.RuneStart(text[upTo]) {
			upTo--
		}
		if upTo <= emitted {
			return nil
		}
		if err := stream(ctx, []byte(text[emitted:upTo])); err != nil {
			return err
		}
		emitted = upTo
		return nil
	}

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var ev streamEvent
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			continue // tolerate events we don't understand
		}
		switch ev.Type {
		case "response.output_text.delta":
			full.WriteString(ev.Delta)
			text := full.String()
			if idx := firstStop(text, stopWords); idx >= 0 {
				res.text = text[:idx]
				res.stopReason = "stop"
				if err := emit(idx); err != nil {
					return res, err
				}
				return res, nil
			}
			if err := emit(len(text) - holdBack); err != nil {
				return res, err
			}
		case "response.completed", "response.incomplete":
			if ev.Response != nil && ev.Response.Usage != nil {
				res.inputTokens = ev.Response.Usage.InputTokens
				res.outputTokens = ev.Response.Usage.OutputTokens
			}
			if ev.Response != nil && ev.Response.IncompleteDetails != nil {
				res.stopReason = ev.Response.IncompleteDetails.Reason
			}
		case "response.failed":
			msg := "response failed"
			if ev.Response != nil && ev.Response.Error != nil && ev.Response.Error.Message != "" {
				msg = ev.Response.Error.Message
			}
			return res, fmt.Errorf("ChatGPT: %s", msg)
		case "error":
			msg := ev.Message
			if msg == "" {
				msg = data
			}
			return res, fmt.Errorf("ChatGPT: %s", msg)
		}
	}
	if err := scanner.Err(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return res, ctxErr
		}
		return res, fmt.Errorf("read ChatGPT stream: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return res, err
	}
	res.text = full.String()
	if res.stopReason == "" {
		res.stopReason = "stop"
	}
	if err := emit(len(res.text)); err != nil {
		return res, err
	}
	return res, nil
}

// firstStop returns the index of the earliest stop word in text, or -1.
func firstStop(text string, stopWords []string) int {
	best := -1
	for _, s := range stopWords {
		if s == "" {
			continue
		}
		if i := strings.Index(text, s); i >= 0 && (best < 0 || i < best) {
			best = i
		}
	}
	return best
}

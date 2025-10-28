package list

import (
	"encoding/json"
	"strings"

	consts "github.com/GarroshIcecream/yummy/yummy/consts"
	"github.com/charmbracelet/bubbles/list"
)

// CustomFilter handles filtering with special commands like @author and @category
func CustomFilter(query string, targets []string) []list.Rank {
	query = strings.TrimSpace(query)

	// Handle special filter commands
	if strings.HasPrefix(query, "@author ") {
		author := strings.TrimSpace(strings.TrimPrefix(query, "@author "))
		if author == "" {
			return []list.Rank{}
		}
		return filterByJSONField(author, targets, consts.AuthorField)
	}

	if strings.HasPrefix(query, "@category ") {
		categoryInput := strings.TrimSpace(strings.TrimPrefix(query, "@category "))
		if categoryInput == "" {
			return []list.Rank{}
		}
		return filterByArrayField(categoryInput, targets, consts.CategoryField)
	}

	if strings.HasPrefix(query, "@title ") {
		title := strings.TrimSpace(strings.TrimPrefix(query, "@title "))
		if title == "" {
			return []list.Rank{}
		}
		return filterByJSONField(title, targets, consts.TitleField)
	}

	if strings.HasPrefix(query, "@description ") {
		description := strings.TrimSpace(strings.TrimPrefix(query, "@description "))
		if description == "" {
			return []list.Rank{}
		}
		return filterByJSONField(description, targets, consts.DescriptionField)
	}

	if strings.HasPrefix(query, "@ingredients ") {
		ingredientsInput := strings.TrimSpace(strings.TrimPrefix(query, "@ingredients "))
		if ingredientsInput == "" {
			return []list.Rank{}
		}
		return filterByArrayField(ingredientsInput, targets, consts.IngredientsField)
	}

	if strings.HasPrefix(query, "@url ") {
		url := strings.TrimSpace(strings.TrimPrefix(query, "@url "))
		if url == "" {
			return []list.Rank{}
		}
		return filterByJSONField(url, targets, consts.URLField)
	}

	// Default to fuzzy search on title field for regular text
	return filterByJSONField(query, targets, consts.TitleField)
}

// filterByArrayField filters recipes by a comma-separated list for array fields (categories, ingredients)
func filterByArrayField(input string, targets []string, fieldName consts.FilterField) []list.Rank {
	var ranks []list.Rank

	// Split the input by commas and trim whitespace
	parts := strings.Split(input, ",")
	var searchTerms []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			searchTerms = append(searchTerms, trimmed)
		}
	}

	if len(searchTerms) == 0 {
		return []list.Rank{}
	}

	for i, target := range targets {
		// Try to parse the target as JSON
		var filterData map[string]interface{}
		if err := json.Unmarshal([]byte(target), &filterData); err != nil {
			// If JSON parsing fails, fall back to simple string matching with AND logic
			allTermsMatched := true
			var firstMatchedTerm string
			for _, searchTerm := range searchTerms {
				if !strings.Contains(strings.ToLower(target), strings.ToLower(searchTerm)) {
					allTermsMatched = false
					break
				}
				if firstMatchedTerm == "" {
					firstMatchedTerm = searchTerm
				}
			}

			if allTermsMatched && firstMatchedTerm != "" {
				matchedIndexes := findMatchedIndices(target, firstMatchedTerm)
				ranks = append(ranks, list.Rank{
					Index:          i,
					MatchedIndexes: matchedIndexes,
				})
			}
			continue
		}

		// Get the specified field from the JSON
		fieldValue, exists := filterData[string(fieldName)]
		if !exists {
			continue
		}

		// Handle array fields
		var fieldArray []string
		switch v := fieldValue.(type) {
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok {
					fieldArray = append(fieldArray, str)
				}
			}
		default:
			continue
		}

		// Check if ALL search terms match any of the field values (AND logic)
		allTermsMatched := true
		var matchedTerms []string

		for _, searchTerm := range searchTerms {
			termFound := false
			for _, fieldItem := range fieldArray {
				if strings.Contains(strings.ToLower(fieldItem), strings.ToLower(searchTerm)) {
					termFound = true
					matchedTerms = append(matchedTerms, searchTerm)
					break
				}
			}
			if !termFound {
				allTermsMatched = false
				break
			}
		}

		if allTermsMatched {
			// Use the first matched term for highlighting
			matchedIndexes := findMatchedIndices(strings.Join(fieldArray, " "), matchedTerms[0])
			ranks = append(ranks, list.Rank{
				Index:          i,
				MatchedIndexes: matchedIndexes,
			})
		}
	}

	return ranks
}

// filterByJSONField filters recipes by a specific field in the JSON filter data
func filterByJSONField(searchTerm string, targets []string, fieldName consts.FilterField) []list.Rank {
	var ranks []list.Rank

	for i, target := range targets {
		// Try to parse the target as JSON
		var filterData map[string]interface{}
		if err := json.Unmarshal([]byte(target), &filterData); err != nil {
			matchedIndexes := findMatchedIndices(target, searchTerm)
			ranks = append(ranks, list.Rank{
				Index:          i,
				MatchedIndexes: matchedIndexes,
			})
			continue
		}

		// Get the field value from the JSON
		fieldValue, exists := filterData[string(fieldName)]
		if !exists {
			continue
		}

		// Handle different field types
		var searchableValue string
		switch v := fieldValue.(type) {
		case string:
			searchableValue = v
		case []interface{}:
			// Handle arrays (like categories)
			var stringSlice []string
			for _, item := range v {
				if str, ok := item.(string); ok {
					stringSlice = append(stringSlice, str)
				}
			}
			searchableValue = strings.Join(stringSlice, " ")
		case bool:
			// Handle boolean fields (like favourite)
			if v {
				searchableValue = "true"
			} else {
				searchableValue = "false"
			}
		default:
			searchableValue = ""
		}

		// Check if the search term matches the field value
		if strings.Contains(strings.ToLower(searchableValue), strings.ToLower(searchTerm)) {
			// For JSON-based filtering, we'll use simple matching indices
			// since the original target is JSON, not the field value
			matchedIndexes := findMatchedIndices(searchableValue, searchTerm)
			ranks = append(ranks, list.Rank{
				Index:          i,
				MatchedIndexes: matchedIndexes,
			})
		}
	}

	return ranks
}

// findMatchedIndices finds the indices of matched characters in the target string
func findMatchedIndices(target string, search string) []int {
	var indices []int
	targetLower := strings.ToLower(target)
	searchLower := strings.ToLower(search)

	start := 0
	for {
		index := strings.Index(targetLower[start:], searchLower)
		if index == -1 {
			break
		}
		actualIndex := start + index
		for j := 0; j < len(searchLower); j++ {
			indices = append(indices, actualIndex+j)
		}
		start = actualIndex + len(searchLower)
	}

	return indices
}

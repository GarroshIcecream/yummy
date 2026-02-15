package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// wrapText wraps text to fit within the specified width
func WrapTextToWidth(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		// If adding this word would exceed the width, start a new line
		if len(currentLine)+len(word)+1 > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Word is longer than width, add it anyway
				lines = append(lines, word)
			}
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// parseInstructions extracts instructions from the markdown
func ParseInstructionsFromMarkdown(text string) ([]string, error) {
	instructionsStart := strings.Index(text, "## 👩‍🍳 Cooking Instructions")
	if instructionsStart == -1 {
		return []string{}, fmt.Errorf("no instructions found")
	}

	// Find the end of instructions section
	instructionsEnd := strings.Index(text[instructionsStart:], "\n## ")
	if instructionsEnd == -1 {
		instructionsEnd = len(text)
	} else {
		instructionsEnd += instructionsStart
	}

	instructionsSection := text[instructionsStart:instructionsEnd]

	// Parse numbered instructions
	lines := strings.Split(instructionsSection, "\n")
	instructions := []string{}
	numberedItem := regexp.MustCompile(`^(?:\*\*)?\d+\.(?:\*\*)?\s*(.+)$`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if match := numberedItem.FindStringSubmatch(line); len(match) > 1 {
			instructions = append(instructions, strings.TrimSpace(match[1]))
		}
	}

	return instructions, nil
}

// parseCategories extracts categories from the markdown
func ParseCategoriesFromMarkdown(text string) ([]string, error) {
	// Find the categories section
	categoriesStart := strings.Index(text, "## 🏷️ Recipe Type")
	if categoriesStart == -1 {
		return []string{}, fmt.Errorf("no categories found")
	}

	// Find the end of categories section
	categoriesEnd := strings.Index(text[categoriesStart:], "\n## ")
	if categoriesEnd == -1 {
		categoriesEnd = len(text)
	} else {
		categoriesEnd += categoriesStart
	}

	categoriesSection := text[categoriesStart:categoriesEnd]

	// Extract categories from backticks
	re := regexp.MustCompile("`([^`]+)`")
	matches := re.FindAllStringSubmatch(categoriesSection, -1)
	categories := []string{}
	for _, match := range matches {
		if len(match) > 1 {
			categories = append(categories, strings.TrimSpace(match[1]))
		}
	}

	return categories, nil
}

// parseSourceURL extracts the source URL from the markdown
func ParseSourceURLFromMarkdown(text string) (string, error) {
	urlMatch := regexp.MustCompile(`🔗 \[View Original Recipe\]\((.+?)\)`).FindStringSubmatch(text)
	if len(urlMatch) > 1 {
		return strings.TrimSpace(urlMatch[1]), nil
	}
	return "", fmt.Errorf("no URL found")
}

// parseDuration parses a duration string into time.Duration
func ParseDurationFromString(durationStr string) time.Duration {
	if durationStr == "" || durationStr == "N/A" {
		return 0
	}

	// Handle common duration formats
	durationStr = strings.ToLower(strings.TrimSpace(durationStr))

	// Try to parse as Go duration first
	if duration, err := time.ParseDuration(durationStr); err == nil {
		return duration
	}

	// Handle "X hours Y minutes" format
	if match := regexp.MustCompile(`(\d+)\s*hours?\s*(\d+)\s*minutes?`).FindStringSubmatch(durationStr); len(match) > 2 {
		hours, _ := strconv.Atoi(match[1])
		minutes, _ := strconv.Atoi(match[2])
		return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute
	}

	// Handle "X hours" format
	if match := regexp.MustCompile(`(\d+)\s*hours?`).FindStringSubmatch(durationStr); len(match) > 1 {
		hours, _ := strconv.Atoi(match[1])
		return time.Duration(hours) * time.Hour
	}

	// Handle "X minutes" format
	if match := regexp.MustCompile(`(\d+)\s*minutes?`).FindStringSubmatch(durationStr); len(match) > 1 {
		minutes, _ := strconv.Atoi(match[1])
		return time.Duration(minutes) * time.Minute
	}

	return 0
}

func ValidateURL(urlStr string) error {
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return err
	}
	return nil
}

func ValidateInteger(str string) error {
	if str == "" {
		return nil
	}
	if _, err := strconv.Atoi(str); err != nil {
		return fmt.Errorf("must be a valid number")
	}
	return nil
}

func ValidateRequired(str string) error {
	if str == "" {
		return fmt.Errorf("required field")
	}
	return nil
}

func ValidateDuration(str string) error {
	if str == "" {
		return nil
	}
	if _, err := time.ParseDuration(str); err != nil {
		return fmt.Errorf("must be a valid duration")
	}
	return nil
}

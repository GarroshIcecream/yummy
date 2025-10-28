package utils

import (
	"fmt"
	"net/url"
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

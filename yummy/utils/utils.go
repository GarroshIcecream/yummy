package utils

import (
	"fmt"
	"net/url"
	"strconv"
)

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

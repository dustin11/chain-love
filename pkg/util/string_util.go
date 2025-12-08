package util

import "strings"

// IsBlank returns true if s is empty or contains only whitespace.
func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotBlank returns true if s contains non-whitespace characters.
func IsNotBlank(s string) bool {
	return !IsBlank(s)
}

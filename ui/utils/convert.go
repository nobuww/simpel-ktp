// Package utils provides utility functions for templ templates
package utils

import "strconv"

// IntToStr converts an integer to string using strconv.Itoa
// This provides a consistent way to convert integers in templates
func IntToStr(n int) string {
	return strconv.Itoa(n)
}

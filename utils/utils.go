// Package utils provides common utilities for the Sui Go SDK, such as address parsing and type tag handling.
package utils

// StringPtr returns a pointer to the given string.
func StringPtr(value string) *string {
	return &value
}

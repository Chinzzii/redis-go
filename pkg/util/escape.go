package util

import "strings"

// EscapeNewlines replaces real newlines with the literal "\n".
func EscapeNewlines(s string) string {
    return strings.ReplaceAll(s, "\n", "\\n")
}

// UnescapeNewlines restores "\n" sequences to actual newlines.
func UnescapeNewlines(s string) string {
    return strings.ReplaceAll(s, "\\n", "\n")
}

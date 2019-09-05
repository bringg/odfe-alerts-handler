package handlers

import "strings"

func fields(s string, sep rune) []string {
	return strings.FieldsFunc(s, func(c rune) bool {
		return c == sep
	})
}

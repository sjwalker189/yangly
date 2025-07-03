package casing

import (
	"strings"
	"unicode"
)

func ToPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = string(unicode.ToUpper(rune(part[0]))) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

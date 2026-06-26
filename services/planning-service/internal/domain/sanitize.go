package domain

import (
	"strings"
	"unicode"
)

func SanitizeRequestName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "untitled-request"
	}
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ' ' {
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "untitled-request"
	}
	if len(out) > 80 {
		return out[:80]
	}
	return out
}

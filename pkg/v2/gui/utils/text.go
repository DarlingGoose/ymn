package utils

import "strings"

func Acronym(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	parts := strings.Fields(s)
	if len(parts) == 0 {
		return ""
	}

	// Single word: use first 1-2 runes.
	if len(parts) == 1 {
		runes := []rune(parts[0])
		if len(runes) == 0 {
			return ""
		}
		if len(runes) == 1 {
			return strings.ToUpper(string(runes[0]))
		}
		return strings.ToUpper(string(runes[:2]))
	}

	var out []rune
	for _, part := range parts {
		runes := []rune(part)
		if len(runes) == 0 {
			continue
		}

		out = append(out, runes[0])
		if len(out) >= 2 {
			break
		}
	}

	return strings.ToUpper(string(out))
}

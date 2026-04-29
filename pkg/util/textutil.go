package util

import (
	"path/filepath"
	"strings"
	"unicode"
)

func TruncateForWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	return string(runes[:width])
}

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func DeriveGameName(inputPath, executablePath string, inputWasDir bool) string {
	if inputWasDir {
		return filepath.Base(inputPath)
	}
	return strings.TrimSuffix(filepath.Base(executablePath), filepath.Ext(executablePath))
}

func SanitizeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "_", "-")
	name = replacer.Replace(name)
	var builder strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	sanitized := strings.Trim(builder.String(), "-")
	if sanitized == "" {
		return "game"
	}
	return sanitized
}

func DesktopExecEscape(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		" ", "\\ ",
		"\t", "\\\t",
		"\n", "\\\n",
		"\"", "\\\"",
		"'", "\\'",
		">", "\\>",
		"<", "\\<",
		"~", "\\~",
		"|", "\\|",
		"&", "\\&",
		";", "\\;",
		"$", "\\$",
		"*", "\\*",
		"?", "\\?",
		"#", "\\#",
		"(", "\\(",
		")", "\\)",
		"`", "\\`",
	)
	return replacer.Replace(strings.TrimSpace(value))
}

func AnkiDeckName(gameName string) string {
	return strings.TrimSpace(gameName)
}

func ContainsKanji(text string) bool {
	for _, r := range text {
		if unicode.In(r, unicode.Han) {
			return true
		}
	}
	return false
}

func HtmlEscapedSingleLine(text string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&#39;")
	return replacer.Replace(strings.TrimSpace(text))
}

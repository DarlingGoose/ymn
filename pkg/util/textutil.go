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

func NormalizeGUISelectionText(text string) string {
	fields := strings.Fields(strings.ReplaceAll(text, "\x00", " "))
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, " ")
}

func BoolSettingLabel(value bool) string {
	if value {
		return "On"
	}
	return "Off"
}

func FindFlashcardSourceLine(displayBuffer, selectedText string) string {
	selectedText = NormalizeGUISelectionText(selectedText)
	if selectedText == "" {
		return ""
	}

	lines := strings.Split(displayBuffer, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := NormalizeGUISelectionText(lines[i])
		if line == "" {
			continue
		}
		if strings.Contains(line, selectedText) {
			return lines[i]
		}
	}
	return ""
}

func LimitLines(text string, limit int) string {
	if limit <= 0 || strings.TrimSpace(text) == "" {
		return text
	}

	lines := strings.Split(text, "\n")
	if len(lines) <= limit {
		return text
	}
	return strings.Join(lines[len(lines)-limit:], "\n")
}

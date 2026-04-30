package gameconfig

import (
	"os"
	"path/filepath"
	"strings"
)

type CompatibilityProfile struct {
	Locale        string
	StageToPrefix bool
	Reason        string
}

func detectCompatibilityProfile(root string) CompatibilityProfile {
	var hasTJS bool
	var hasXP3 bool
	var hasMojibakeName bool
	var hasJapaneseName bool

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		name := filepath.Base(path)
		lower := strings.ToLower(name)

		if !info.IsDir() {
			switch filepath.Ext(lower) {
			case ".tjs":
				hasTJS = true
			case ".xp3":
				hasXP3 = true
			}
		}

		if looksLikeMojibake(name) {
			hasMojibakeName = true
		}
		if containsJapanese(name) {
			hasJapaneseName = true
		}

		return nil
	})

	// Kirikiri/KAG games are commonly .tjs + .xp3.
	// They often dislike running from Wine's Z: drive.
	if hasTJS || hasXP3 || hasMojibakeName || hasJapaneseName {
		reasons := make([]string, 0, 4)
		if hasTJS {
			reasons = append(reasons, "TJS script files detected")
		}
		if hasXP3 {
			reasons = append(reasons, "XP3 archive files detected")
		}
		if hasMojibakeName {
			reasons = append(reasons, "mojibake-looking filenames detected")
		}
		if hasJapaneseName {
			reasons = append(reasons, "Japanese filenames detected")
		}

		return CompatibilityProfile{
			Locale:        "ja_JP.UTF-8",
			StageToPrefix: true,
			Reason:        strings.Join(reasons, ", "),
		}
	}

	return CompatibilityProfile{}
}

func containsJapanese(s string) bool {
	for _, r := range s {
		switch {
		case r >= 0x3040 && r <= 0x309F: // Hiragana
			return true
		case r >= 0x30A0 && r <= 0x30FF: // Katakana
			return true
		case r >= 0x4E00 && r <= 0x9FFF: // CJK
			return true
		}
	}
	return false
}

func looksLikeMojibake(s string) bool {
	// Common mojibake markers from CP932/Shift-JIS interpreted incorrectly.
	return strings.Contains(s, "ƒ") ||
		strings.Contains(s, "‚") ||
		strings.Contains(s, "„") ||
		strings.Contains(s, "‰") ||
		strings.Contains(s, "Œ") ||
		strings.Contains(s, "Ž") ||
		strings.Contains(s, "Р") ||
		strings.Contains(s, "Ц")
}

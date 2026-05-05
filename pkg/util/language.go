package util

import (
	"os"
	"strings"

	"golang.org/x/text/language"
)

const (
	DefaultTranslationLanguage       = "English"
	SystemTranslationLanguageSetting = "system"
)

func ResolveTranslationLanguage(setting string) string {
	setting = strings.TrimSpace(setting)
	if setting == "" || strings.EqualFold(setting, SystemTranslationLanguageSetting) || strings.EqualFold(setting, "System Default") {
		return SystemTranslationLanguage()
	}
	if label := TranslationLanguageLabel(setting); label != "" {
		return label
	}
	return DefaultTranslationLanguage
}

func SystemTranslationLanguage() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG", "LANGUAGE"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			continue
		}
		if label := TranslationLanguageLabel(value); label != "" {
			return label
		}
	}
	return DefaultTranslationLanguage
}

func TranslationLanguageLabel(value string) string {
	value = cleanLanguageTag(value)
	if value == "" {
		return ""
	}
	if label := translationLanguageLabelFromCode(value); label != "" {
		return label
	}
	tag, err := language.Parse(value)
	if err != nil {
		return ""
	}
	base, _ := tag.Base()
	return translationLanguageLabelFromCode(base.String())
}

func cleanLanguageTag(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if i := strings.IndexAny(value, ".@"); i >= 0 {
		value = value[:i]
	}
	if i := strings.Index(value, ":"); i >= 0 {
		value = value[:i]
	}
	value = strings.ReplaceAll(value, "_", "-")
	return strings.ToLower(strings.TrimSpace(value))
}

func translationLanguageLabelFromCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" || code == "c" || code == "posix" {
		return ""
	}
	if i := strings.Index(code, "-"); i >= 0 {
		code = code[:i]
	}
	switch code {
	case "en":
		return "English"
	case "ja", "jpn":
		return "Japanese"
	case "es":
		return "Spanish"
	case "fr":
		return "French"
	case "de":
		return "German"
	case "ko":
		return "Korean"
	case "zh":
		return "Chinese"
	case "it":
		return "Italian"
	case "pt":
		return "Portuguese"
	case "ru":
		return "Russian"
	default:
		return ""
	}
}

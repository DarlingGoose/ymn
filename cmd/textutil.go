package cmd

func truncateForWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	return string(runes[:width])
}

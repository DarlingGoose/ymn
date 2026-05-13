package transcript

import (
	"fmt"
	"image/color"
	"regexp"
	"strings"
	"unicode"

	"github.com/DarlingGoose/vntext/pkg/engine"
	"github.com/google/uuid"
)

func cleanTranscriptText(text string) string {
	text = strings.ReplaceAll(text, `\n`, " ")
	text = strings.ReplaceAll(text, "\r\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func compactHookLabel(hook string) string {
	hook = strings.TrimSpace(hook)
	if hook == "" {
		return ""
	}

	// Example: @14748B:KSH_dl.exe -> @14748B
	if idx := strings.Index(hook, ":"); idx > 0 {
		return hook[:idx]
	}

	return hook
}

func transcriptRowKey(index int, line *engine.Line) string {
	if line == nil {
		return fmt.Sprintf("line:%d:nil", index)
	}

	hook := strings.TrimSpace(line.Hook)
	speaker := cleanTranscriptText(line.Speaker)
	text := cleanTranscriptText(line.Text)
	raw := strings.TrimSpace(line.Raw)

	if raw != "" {
		return fmt.Sprintf("line:%d:%s:%s", index, hook, raw)
	}

	return fmt.Sprintf("line:%d:%s:%s:%s", index, hook, speaker, text)
}

func mixNRGBA(a, b color.NRGBA, amount float32) color.NRGBA {
	if amount < 0 {
		amount = 0
	}
	if amount > 1 {
		amount = 1
	}

	ia := 1 - amount

	return color.NRGBA{
		R: uint8(float32(a.R)*amount + float32(b.R)*ia),
		G: uint8(float32(a.G)*amount + float32(b.G)*ia),
		B: uint8(float32(a.B)*amount + float32(b.B)*ia),
		A: uint8(float32(a.A)*amount + float32(b.A)*ia),
	}
}

var cleanRe = regexp.MustCompile(`^(\[[a-zA-Z_]*\s*])+`)

func transcriptRowFromEngineLine(line engine.Line) transcriptRow {
	text := strings.TrimSpace(line.Text)

	if text == "" {
		text = strings.TrimSpace(fmt.Sprint(line))
	}
	var info bool
	if strings.Contains(text, "[system]") {
		text = cleanRe.ReplaceAllString(text, "")
		info = true
	}

	return transcriptRow{
		Key:     uuid.NewString(),
		Hook:    line.Hook,
		Text:    text,
		Speaker: strings.TrimSpace(line.Speaker),
		Raw:     line.Raw,
		Info:    info,
		Time:    line.Time.Format("15:04:05"),
	}
}

func transcriptRowLooksLikeLanguage(row transcriptRow) bool {
	if row.Info {
		return true
	}

	text := strings.TrimSpace(row.Text + " " + row.Speaker)
	if text == "" || containsReplacementGlyph(text) {
		return false
	}

	for _, r := range text {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func containsReplacementGlyph(text string) bool {
	for _, r := range text {
		switch r {
		case unicode.ReplacementChar, '\u25a1', '\u25a0', '\u25fb', '\u25fc':
			return true
		}
	}
	return false
}

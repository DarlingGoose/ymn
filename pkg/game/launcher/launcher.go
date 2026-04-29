package launcher

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Seann-Moser/wgl/pkg/game/gameconfig"
	"github.com/Seann-Moser/wgl/pkg/util"
)

func SaveDesktopEntry(config *gameconfig.GameConfig, launcherDir string) (path string, err error) {
	if config == nil {
		return "", errors.New("config is nil")
	}
	if strings.TrimSpace(config.Name) == "" {
		return "", errors.New("config name is required")
	}

	if strings.TrimSpace(launcherDir) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}
		launcherDir = filepath.Join(home, ".local", "share", "applications")
	}

	if err := os.MkdirAll(launcherDir, 0o755); err != nil {
		return "", fmt.Errorf("create launcher dir: %w", err)
	}

	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	self, err = filepath.Abs(self)
	if err != nil {
		return "", fmt.Errorf("resolve executable abs path: %w", err)
	}

	fileName := util.SanitizeName(config.Name) + ".desktop"
	path = filepath.Join(launcherDir, fileName)

	execLine := desktopExec(self, "run", config.Name)

	var b strings.Builder
	b.WriteString("[Desktop Entry]\n")
	b.WriteString("Type=Application\n")
	b.WriteString("Version=1.0\n")
	b.WriteString("Name=" + desktopValue(config.Name) + "\n")
	b.WriteString("Comment=Launch " + desktopValue(config.Name) + "\n")
	b.WriteString("Exec=" + execLine + "\n")
	b.WriteString("Terminal=false\n")
	b.WriteString("Categories=Game;\n")
	b.WriteString("StartupNotify=true\n")

	if strings.TrimSpace(config.WorkingDir) != "" {
		b.WriteString("Path=" + desktopValue(config.WorkingDir) + "\n")
	}

	if strings.TrimSpace(config.IconPath) != "" {
		b.WriteString("Icon=" + desktopValue(config.IconPath) + "\n")
	}

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("write desktop entry: %w", err)
	}

	return path, nil
}

func desktopExec(parts ...string) string {
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		escaped = append(escaped, desktopQuote(part))
	}
	return strings.Join(escaped, " ")
}

func desktopQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "$", "\\$")
	return `"` + s + `"`
}

func desktopValue(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}

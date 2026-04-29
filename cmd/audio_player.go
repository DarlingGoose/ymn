package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func playAudioFile(path string) error {
	path = strings.TrimSpace(path)
	if !isExistingFile(path) {
		return fmt.Errorf("audio file not found: %s", path)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("afplay"); err == nil {
			cmd = exec.Command("afplay", path)
			break
		}
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		for _, candidate := range [][]string{
			{"xdg-open", path},
			{"mpv", path},
			{"ffplay", "-nodisp", "-autoexit", path},
			{"vlc", "--play-and-exit", path},
		} {
			if _, err := exec.LookPath(candidate[0]); err == nil {
				cmd = exec.Command(candidate[0], candidate[1:]...)
				break
			}
		}
	}
	if cmd == nil {
		return fmt.Errorf("no supported audio player found")
	}
	return cmd.Start()
}

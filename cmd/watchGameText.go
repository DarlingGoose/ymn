package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Seann-Moser/wgl/pkg/game/gameconfig"
	"github.com/Seann-Moser/wgl/pkg/game/launcher"
	"github.com/Seann-Moser/wgl/pkg/util"
	"github.com/spf13/cobra"
)

const rpgMakerTranscriptFilename = "wgl-dialogue.log"

var ansiEscapeSequencePattern = regexp.MustCompile(`\x1b(?:\[[0-?]*[ -/]*[@-~]|[@-_])`)

var watchGameTextName string
var watchGameTextPrintExisting bool
var watchGameTextPollInterval time.Duration

var watchGameTextCmd = &cobra.Command{
	Use:     "watch-game-text [game-name]",
	Aliases: []string{"watchGameText"},
	Short:   "Watch the RPG Maker dialogue transcript for a saved game",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		selectedName := strings.TrimSpace(watchGameTextName)
		if selectedName == "" && len(args) > 0 {
			selectedName = args[0]
		}

		var cfg *gameconfig.GameConfig
		var err error
		if selectedName == "" {
			picker, err := launcher.NewPicker("Select a game transcript to watch", "watch")
			if err != nil {
				return err
			}
			cfg, err = picker.SelectGameConfig()
			if err != nil {
				return err
			}
		} else {
			cfg, err = gameconfig.FindConfig(selectedName)
			if err != nil {
				return err
			}
		}

		logPath, err := resolveRPGMakerTranscriptPath(util.FirstNonEmpty(cfg.GamePath, cfg.Executable, cfg.WorkingDir))
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "watching transcript: %s\n", logPath)
		if !watchGameTextPrintExisting {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "waiting for new dialogue; pass --print-existing to dump the current log first")
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		return tailTranscriptFile(ctx, cmd, logPath, watchGameTextPrintExisting, watchGameTextPollInterval)
	},
}

func init() {
	rootCmd.AddCommand(watchGameTextCmd)
	watchGameTextCmd.Flags().StringVarP(&watchGameTextName, "game", "g", "", "name of the saved game transcript to watch")
	watchGameTextCmd.Flags().BoolVar(&watchGameTextPrintExisting, "print-existing", false, "print the current transcript contents before waiting for new dialogue")
	watchGameTextCmd.Flags().DurationVar(&watchGameTextPollInterval, "poll-interval", 750*time.Millisecond, "how often to poll the transcript log for new text")
}

func resolveRPGMakerTranscriptPath(inputPath string) (string, error) {
	projectRoot, _, err := resolveRPGMakerProjectRoot(inputPath)
	if err != nil {
		return "", err
	}

	candidates := transcriptPathCandidates(projectRoot)
	for _, candidate := range candidates {
		if isExistingFile(candidate) {
			return candidate, nil
		}
	}
	if len(candidates) == 0 {
		return "", errors.New("could not derive an RPG Maker transcript path")
	}
	return candidates[0], nil
}

func transcriptPathCandidates(projectRoot string) []string {
	root := filepath.Clean(strings.TrimSpace(projectRoot))
	if root == "" {
		return nil
	}

	var candidates []string
	addCandidate := func(path string) {
		cleaned := filepath.Clean(strings.TrimSpace(path))
		if cleaned == "" {
			return
		}
		for _, existing := range candidates {
			if existing == cleaned {
				return
			}
		}
		candidates = append(candidates, cleaned)
	}

	addCandidate(filepath.Join(root, rpgMakerTranscriptFilename))
	if strings.EqualFold(filepath.Base(root), "www") {
		addCandidate(filepath.Join(filepath.Dir(root), rpgMakerTranscriptFilename))
	}
	addCandidate(filepath.Join(root, "www", rpgMakerTranscriptFilename))

	return candidates
}

func tailTranscriptFile(ctx context.Context, cmd *cobra.Command, logPath string, printExisting bool, pollInterval time.Duration) error {
	if pollInterval <= 0 {
		pollInterval = 750 * time.Millisecond
	}

	var offset int64
	if !printExisting {
		info, err := os.Stat(logPath)
		if err == nil {
			offset = info.Size()
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("stat transcript log: %w", err)
		}
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	waitingNotified := false
	for {
		delta, err := readTranscriptDelta(logPath, &offset)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if !waitingNotified {
					fmt.Fprintln(cmd.ErrOrStderr(), "transcript log not found yet; start the game and advance dialogue")
					waitingNotified = true
				}
			} else {
				return err
			}
		} else if delta != "" {
			if _, err := cmd.OutOrStdout().Write([]byte(sanitizeTranscriptForDisplay(delta))); err != nil {
				return fmt.Errorf("write transcript output: %w", err)
			}
		} else {
			waitingNotified = false
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func readTranscriptDelta(logPath string, offset *int64) (string, error) {
	info, err := os.Stat(logPath)
	if err != nil {
		return "", err
	}
	if info.Size() < *offset {
		*offset = 0
	}
	if info.Size() == *offset {
		return "", nil
	}

	file, err := os.Open(logPath)
	if err != nil {
		return "", fmt.Errorf("open transcript log: %w", err)
	}
	defer file.Close()

	if _, err := file.Seek(*offset, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek transcript log: %w", err)
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read transcript log: %w", err)
	}
	*offset = info.Size()
	return string(data), nil
}

func sanitizeTranscriptForDisplay(text string) string {
	if text == "" {
		return ""
	}

	sanitized := ansiEscapeSequencePattern.ReplaceAllString(text, "")
	var builder strings.Builder
	builder.Grow(len(sanitized))

	for len(sanitized) > 0 {
		r, size := utf8.DecodeRuneInString(sanitized)
		if r == utf8.RuneError && size == 1 {
			sanitized = sanitized[size:]
			continue
		}

		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			builder.WriteRune(r)
		}
		sanitized = sanitized[size:]
	}

	return builder.String()
}

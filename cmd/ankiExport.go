package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var ankiExportGameName string

var ankiExportCmd = &cobra.Command{
	Use:   "anki-export [game-name]",
	Short: "Export flashcards for manual Anki import",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gameName, err := resolveSelectedGameName(ankiExportGameName, args, "Select a game deck to export", "export")
		if err != nil {
			return err
		}

		cards, err := loadFlashcards(gameName)
		if err != nil {
			return err
		}
		if len(cards) == 0 {
			return fmt.Errorf("no flashcards saved for %s", gameName)
		}

		path, err := exportFlashcardsToTSV(gameName, cards)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "exported %d cards to %s\n", len(cards), path)
		fmt.Fprintf(cmd.OutOrStdout(), "import into deck %q in Anki\n", ankiDeckName(gameName))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ankiExportCmd)
	ankiExportCmd.Flags().StringVarP(&ankiExportGameName, "game", "g", "", "name of the saved game deck to export")
}

func resolveSelectedGameName(flagValue string, args []string, title, action string) (string, error) {
	selectedName := strings.TrimSpace(flagValue)
	if selectedName == "" && len(args) > 0 {
		selectedName = args[0]
	}
	if selectedName != "" {
		cfg, err := findGameConfig(selectedName)
		if err != nil {
			return "", err
		}
		return cfg.Name, nil
	}

	cfg, err := selectGameConfigWithTUI(title, action)
	if err != nil {
		return "", err
	}
	return cfg.Name, nil
}

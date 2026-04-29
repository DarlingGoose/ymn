package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var ankiSyncGameName string
var ankiSyncURL string
var ankiSyncPushSync bool

var ankiSyncCmd = &cobra.Command{
	Use:   "anki-sync [game-name]",
	Short: "Sync flashcards to AnkiConnect using the game name as the deck name",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		gameName, err := resolveSelectedGameName(ankiSyncGameName, args, "Select a game deck to sync", "sync")
		if err != nil {
			return err
		}

		result, err := syncFlashcardsToAnki(gameName, ankiSyncURL, ankiSyncPushSync)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "synced %d cards to deck %q (%d created, %d updated)\n", result.Total, result.DeckName, result.Created, result.Updated)
		if strings.TrimSpace(ankiSyncURL) != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "anki endpoint: %s\n", strings.TrimSpace(ankiSyncURL))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(ankiSyncCmd)
	ankiSyncCmd.Flags().StringVarP(&ankiSyncGameName, "game", "g", "", "name of the saved game deck to sync")
	ankiSyncCmd.Flags().StringVar(&ankiSyncURL, "anki-url", defaultAnkiConnectURL, "AnkiConnect URL")
	ankiSyncCmd.Flags().BoolVar(&ankiSyncPushSync, "sync-collection", true, "call AnkiConnect sync after uploading notes")
}

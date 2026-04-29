package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var addGameRunner string
var addGameSkipVerify bool
var addGameRequiresSteam bool
var addGameSteamAppID string
var addGameIconPath string
var addGameImagePath string

var addGameCmd = &cobra.Command{
	Use:     "add-game <path-to-game-dir-or-exe>",
	Aliases: []string{"addGame"},
	Short:   "Create a launcher config for a Windows game",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := buildGameConfig(
			args[0],
			addGameRunner,
			addGameRequiresSteam,
			addGameSteamAppID,
			addGameIconPath,
			addGameImagePath,
		)
		if err != nil {
			return err
		}

		if !addGameSkipVerify {
			fmt.Fprintln(cmd.OutOrStdout(), "verifying game launch...")
			cfg, err = verifyAndAutofixGameConfig(cfg)
			if err != nil {
				printVerificationAttempts(cmd, cfg.Verification.Attempts)
				return err
			}
			printVerificationAttempts(cmd, cfg.Verification.Attempts)
		}

		configPath, err := saveGameConfig(cfg)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "saved config: %s\n", configPath)
		fmt.Fprintf(cmd.OutOrStdout(), "game: %s\n", cfg.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "runner: %s\n", cfg.Runner)
		fmt.Fprintf(cmd.OutOrStdout(), "requires steam: %t\n", cfg.RequiresSteam)
		if strings.TrimSpace(cfg.SteamAppID) != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "steam app id: %s\n", cfg.SteamAppID)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "executable: %s\n", cfg.Executable)
		if strings.TrimSpace(cfg.IconPath) != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "icon: %s\n", cfg.IconPath)
		}
		if strings.TrimSpace(cfg.ImagePath) != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "image: %s\n", cfg.ImagePath)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addGameCmd)
	addGameCmd.Flags().StringVar(&addGameRunner, "runner", "auto", "runner to use: auto, wine, proton, or steam")
	addGameCmd.Flags().BoolVar(&addGameSkipVerify, "skip-verify", false, "skip the launch verification smoke test")
	addGameCmd.Flags().BoolVar(&addGameRequiresSteam, "requires-steam", false, "mark the game as requiring Steam to launch")
	addGameCmd.Flags().StringVar(&addGameSteamAppID, "steam-app-id", "", "Steam app id used when launching through Steam")
	addGameCmd.Flags().StringVar(&addGameIconPath, "icon", "", "path to a game icon file (defaults to auto-discovery)")
	addGameCmd.Flags().StringVar(&addGameImagePath, "image", "", "path to a game image/cover file (defaults to auto-discovery)")
}

func printVerificationAttempts(cmd *cobra.Command, attempts []VerificationAttempt) {
	for _, attempt := range attempts {
		status := "failed"
		if attempt.Success {
			status = "ok"
		}
		fmt.Fprintf(
			cmd.OutOrStdout(),
			"verification [%s:%s] %s: %s",
			attempt.Runner,
			attempt.Strategy,
			status,
			attempt.Message,
		)
		if strings.TrimSpace(attempt.LogPath) != "" {
			fmt.Fprintf(cmd.OutOrStdout(), " (log: %s)", attempt.LogPath)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}
}

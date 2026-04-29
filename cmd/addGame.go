package cmd

import (
	"fmt"
	"strings"

	"github.com/Seann-Moser/wgl/pkg/game/gameconfig"
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
		cfg, err := gameconfig.BuildGameConfig(
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

		//if !addGameSkipVerify {
		//	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "verifying game launch...")
		//	cfg, err = verifyAndAutofixGameConfig(cfg)
		//	if err != nil {
		//		printVerificationAttempts(cmd, cfg.Verification.Attempts)
		//		return err
		//	}
		//	printVerificationAttempts(cmd, cfg.Verification.Attempts)
		//}
		configPath, err := gameconfig.SaveGameConfig(cfg)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()

		_, _ = fmt.Fprintf(out, "\nGame Config\n")
		_, _ = fmt.Fprintf(out, "------------\n")

		_, _ = fmt.Fprintf(out, "%-16s %s\n", "Name:", cfg.Name)
		_, _ = fmt.Fprintf(out, "%-16s %s\n", "Runner:", cfg.Runner)
		_, _ = fmt.Fprintf(out, "%-16s %t\n", "Requires Steam:", cfg.RequiresSteam)

		if cfg.SteamAppID != "" {
			_, _ = fmt.Fprintf(out, "%-16s %s\n", "Steam App ID:", cfg.SteamAppID)
		}

		_, _ = fmt.Fprintf(out, "%-16s %s\n", "Executable:", cfg.Executable)

		if cfg.IconPath != "" {
			_, _ = fmt.Fprintf(out, "%-16s %s\n", "Icon:", cfg.IconPath)
		}
		if cfg.ImagePath != "" {
			_, _ = fmt.Fprintf(out, "%-16s %s\n", "Image:", cfg.ImagePath)
		}

		_, _ = fmt.Fprintf(out, "\nSaved to: %s\n", configPath)
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

func printVerificationAttempts(cmd *cobra.Command, attempts []gameconfig.VerificationAttempt) {
	for _, attempt := range attempts {
		status := "failed"
		if attempt.Success {
			status = "ok"
		}
		_, _ = fmt.Fprintf(
			cmd.OutOrStdout(),
			"verification [%s:%s] %s: %s",
			attempt.Runner,
			attempt.Strategy,
			status,
			attempt.Message,
		)
		if strings.TrimSpace(attempt.LogPath) != "" {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), " (log: %s)", attempt.LogPath)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}
}

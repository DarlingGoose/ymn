package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	//"github.com/DarlingGoose/vntext/pkg/runner"
	"github.com/DarlingGoose/wgl/pkg/util"
	"github.com/spf13/cobra"
)

var runGameName string
var runGameSaveLauncher bool
var runGameLauncherDir string
var runGameVirtualDesktop string
var runGameDisableVirtualDesktop bool

var runGameCmd = &cobra.Command{
	Use:     "run-game [game-name]",
	Aliases: []string{"runGame"},
	Short:   "Launch a previously added game",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		//name := strings.TrimSpace(runGameName)
		//if name == "" && len(args) > 0 {
		//	name = strings.TrimSpace(args[0])
		//}
		//if name == "" {
		//	return fmt.Errorf("game name is required")
		//}
		//
		//configs, err := loadInstalledGameConfigs()
		//if err != nil {
		//	return err
		//}
		//cfg, err := gameConfig.FindInstalledGame(configs, name)
		//if err != nil {
		//	return err
		//}
		//
		//if cmd.Flags().Changed("virtual-desktop") {
		//	cfg.VirtualDesktop = strings.TrimSpace(runGameVirtualDesktop)
		//}
		//if runGameDisableVirtualDesktop {
		//	cfg.VirtualDesktop = "off"
		//}
		//
		//if runGameSaveLauncher {
		//	path, err := writeGameDesktopLauncher(cfg.Name, cfg.IconPath, cfg.VirtualDesktop)
		//	if err != nil {
		//		return err
		//	}
		//	fmt.Fprintf(cmd.OutOrStdout(), "Wrote launcher: %s\n", path)
		//}
		//
		//status, err := runner.New().Run(cfg)
		//if err != nil {
		//	return err
		//}
		//if status != nil && strings.TrimSpace(status.Message) != "" {
		//	fmt.Fprintln(cmd.OutOrStdout(), status.Message)
		//}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runGameCmd)
	runGameCmd.Flags().StringVarP(&runGameName, "game", "g", "", "name of the saved game to launch")
	runGameCmd.Flags().BoolVar(&runGameSaveLauncher, "save-launcher", false, "write a .desktop launcher so rofi and Linux app launchers can find the game")
	runGameCmd.Flags().StringVar(&runGameLauncherDir, "launcher-dir", "", "directory where the .desktop launcher is written (default: ~/.local/share/applications)")
	runGameCmd.Flags().StringVar(&runGameVirtualDesktop, "virtual-desktop", "", "Wine/Proton virtual desktop size, for example 1280x720; use off to disable")
	runGameCmd.Flags().BoolVar(&runGameDisableVirtualDesktop, "no-virtual-desktop", false, "disable the Wine/Proton virtual desktop for this launch")
}

func writeGameDesktopLauncher(name, iconPath, virtualDesktop string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("game name is required")
	}

	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	exe, err = filepath.Abs(exe)
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}

	dir := strings.TrimSpace(runGameLauncherDir)
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home dir: %w", err)
		}
		dir = filepath.Join(home, ".local", "share", "applications")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create launcher dir: %w", err)
	}

	path := filepath.Join(dir, "yomuna-"+util.SanitizeName(name)+".desktop")
	var b strings.Builder
	b.WriteString("[Desktop Entry]\n")
	b.WriteString("Type=Application\n")
	b.WriteString("Name=" + desktopEntryValue("Yomuna - "+name) + "\n")
	b.WriteString("Comment=" + desktopEntryValue("Launch "+name+" with Yomuna") + "\n")
	b.WriteString("Exec=" + shellQuote(exe) + " run-game --game " + shellQuote(name))
	if strings.TrimSpace(virtualDesktop) != "" {
		b.WriteString(" --virtual-desktop " + shellQuote(strings.TrimSpace(virtualDesktop)))
	}
	b.WriteString("\n")
	if strings.TrimSpace(iconPath) != "" {
		b.WriteString("Icon=" + desktopEntryValue(strings.TrimSpace(iconPath)) + "\n")
	}
	b.WriteString("Terminal=false\n")
	b.WriteString("Categories=Game;\n")

	if err := os.WriteFile(path, []byte(b.String()), 0o755); err != nil {
		return "", fmt.Errorf("write launcher: %w", err)
	}
	return path, nil
}

func desktopEntryValue(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	return strings.TrimSpace(value)
}

func shellQuote(value string) string {
	return strconv.Quote(value)
}

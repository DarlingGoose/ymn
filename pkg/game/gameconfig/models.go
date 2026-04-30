package gameconfig

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Seann-Moser/wgl/pkg/util"
)

type RunnerType string

const (
	RunnerWine   RunnerType = "wine"
	RunnerProton RunnerType = "proton"
	RunnerSteam  RunnerType = "steam"
)

type RuntimeStatus struct {
	WinePath           string   `json:"wine_path,omitempty"`
	WineBootPath       string   `json:"wineboot_path,omitempty"`
	SteamPath          string   `json:"steam_path,omitempty"`
	SteamRoot          string   `json:"steam_root,omitempty"`
	AvailableProton    []string `json:"available_proton,omitempty"`
	SelectedProtonPath string   `json:"selected_proton_path,omitempty"`
}

type VerificationAttempt struct {
	Runner    RunnerType `json:"runner"`
	Strategy  string     `json:"strategy"`
	Success   bool       `json:"success"`
	Message   string     `json:"message"`
	LogPath   string     `json:"log_path,omitempty"`
	CheckedAt time.Time  `json:"checked_at"`
}

type VerificationStatus struct {
	Verified   bool                  `json:"verified"`
	VerifiedAt time.Time             `json:"verified_at,omitempty"`
	Attempts   []VerificationAttempt `json:"attempts,omitempty"`
}

type GameConfig struct {
	EnableProxyInjection bool               `json:"enable_proxy_injection"`
	Name                 string             `json:"name"`
	GamePath             string             `json:"game_path"`
	Executable           string             `json:"executable"`
	WorkingDir           string             `json:"working_dir"`
	IconPath             string             `json:"icon_path,omitempty"`
	ImagePath            string             `json:"image_path,omitempty"`
	Runner               RunnerType         `json:"runner"`
	RunnerPath           string             `json:"runner_path"`
	PrefixPath           string             `json:"prefix_path,omitempty"`
	RequiresSteam        bool               `json:"requires_steam"`
	SteamAppID           string             `json:"steam_app_id,omitempty"`
	CreatedAt            time.Time          `json:"created_at"`
	RuntimeInfo          RuntimeStatus      `json:"runtime_info"`
	Verification         VerificationStatus `json:"verification"`

	Locale        string `json:"locale,omitempty"`
	StageToPrefix bool   `json:"stage_to_prefix,omitempty"`
	StagedPath    string `json:"staged_path,omitempty"`
}

func (c *GameConfig) driveCPath() string {
	switch c.Runner {
	case RunnerProton:
		return filepath.Join(c.PrefixPath, "pfx", "drive_c")
	default:
		return filepath.Join(c.PrefixPath, "drive_c")
	}
}

func (c *GameConfig) Launch() error {
	cmd, err := c.launchCommand()
	if err != nil {
		return err
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (c *GameConfig) LaunchInBackground() error {
	cmd, err := c.launchCommand()
	if err != nil {
		return err
	}
	logPath := filepath.Join(c.prefixOrGameDir(), "launch.log")

	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open launch log: %w", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil

	// Detach from parent process group.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return err
	}

	// Let Go release process resources after it exits.
	go func() {
		_ = cmd.Wait()
		_ = logFile.Close()
	}()

	return nil
}

func (c *GameConfig) launchCommand() (*exec.Cmd, error) {
	if err := c.validateLaunchConfig(); err != nil {
		return nil, err
	}

	switch c.Runner {
	case RunnerWine:
		return c.wineCommand(), nil

	case RunnerProton:
		return c.protonCommand()

	case RunnerSteam:
		return c.steamCommand(), nil

	default:
		return nil, fmt.Errorf("unsupported runner: %q", c.Runner)
	}
}

func (c *GameConfig) wineCommand() *exec.Cmd {
	wine := util.FirstNonEmpty(c.RunnerPath, c.RuntimeInfo.WinePath, "wine")

	exe := c.windowsExecutablePath()
	cmd := exec.Command(wine, exe)
	cmd.Dir = c.workingDir()
	cmd.Env = c.baseEnv()

	if strings.TrimSpace(c.PrefixPath) != "" {
		cmd.Env = append(cmd.Env, "WINEPREFIX="+c.PrefixPath)
	}

	return cmd
}

func (c *GameConfig) protonCommand() (*exec.Cmd, error) {
	proton := util.FirstNonEmpty(
		c.RunnerPath,
		c.RuntimeInfo.SelectedProtonPath,
	)

	if strings.TrimSpace(proton) == "" {
		return nil, errors.New("proton path is required")
	}

	exe := c.windowsExecutablePath()

	prefix := c.PrefixPath
	if strings.TrimSpace(prefix) == "" {
		prefix = filepath.Join(util.ConfigBaseDir(), "prefixes", util.SanitizeName(c.Name))
	}

	if err := os.MkdirAll(prefix, 0o755); err != nil {
		return nil, fmt.Errorf("create proton prefix dir: %w", err)
	}

	cmd := exec.Command(proton, "run", exe)
	cmd.Dir = c.workingDir()
	cmd.Env = c.baseEnv()

	cmd.Env = append(cmd.Env,
		"STEAM_COMPAT_DATA_PATH="+prefix,
		"STEAM_COMPAT_CLIENT_INSTALL_PATH="+util.FirstNonEmpty(
			c.RuntimeInfo.SteamRoot,
			filepath.Join(os.Getenv("HOME"), ".steam", "steam"),
		),
	)

	return cmd, nil
}

func (c *GameConfig) steamCommand() *exec.Cmd {
	steam := util.FirstNonEmpty(c.RunnerPath, c.RuntimeInfo.SteamPath, "steam")

	var target string
	if strings.TrimSpace(c.SteamAppID) != "" {
		target = "steam://rungameid/" + c.SteamAppID
	} else {
		// Fallback for non-Steam shortcut/exe launch.
		target = c.executablePath()
	}

	cmd := exec.Command(steam, target)
	cmd.Dir = c.workingDir()
	cmd.Env = c.baseEnv()

	return cmd
}

func (c *GameConfig) validateLaunchConfig() error {
	if strings.TrimSpace(c.Name) == "" {
		return errors.New("game name is required")
	}

	if strings.TrimSpace(c.GamePath) == "" {
		return errors.New("game path is required")
	}

	switch c.Runner {
	case RunnerWine:
		if strings.TrimSpace(c.Executable) == "" {
			return errors.New("executable is required for wine runner")
		}

	case RunnerProton:
		if strings.TrimSpace(c.Executable) == "" {
			return errors.New("executable is required for proton runner")
		}

		if strings.TrimSpace(c.RunnerPath) == "" &&
			strings.TrimSpace(c.RuntimeInfo.SelectedProtonPath) == "" {
			return errors.New("proton runner path is required")
		}

	case RunnerSteam:
		if strings.TrimSpace(c.SteamAppID) == "" && strings.TrimSpace(c.Executable) == "" {
			return errors.New("steam_app_id or executable is required for steam runner")
		}

	default:
		return fmt.Errorf("unknown runner: %q", c.Runner)
	}

	return nil
}

func (c *GameConfig) executablePath() string {
	exe := strings.TrimSpace(c.Executable)
	if filepath.IsAbs(exe) {
		return exe
	}
	return filepath.Join(c.GamePath, exe)
}

func (c *GameConfig) workingDir() string {
	if strings.TrimSpace(c.WorkingDir) != "" {
		return c.WorkingDir
	}

	exe := c.executablePath()
	if exe != "" {
		return filepath.Dir(exe)
	}

	return c.GamePath
}

func (c *GameConfig) prefixOrGameDir() string {
	if strings.TrimSpace(c.PrefixPath) != "" {
		return c.PrefixPath
	}
	if strings.TrimSpace(c.GamePath) != "" {
		return c.GamePath
	}
	return os.TempDir()
}

func (c *GameConfig) baseEnv() []string {
	env := cleanWineEnv(os.Environ())

	// Standard overrides
	overrides := "winemenubuilder.exe=d"

	// If the toggle is on, add the version.dll override
	if c.EnableProxyInjection {
		overrides += ";version=n,b"
	}

	env = append(env, "WINEDLLOVERRIDES="+overrides)

	if strings.TrimSpace(c.Locale) != "" {
		env = append(env,
			"LANG="+c.Locale,
			"LC_ALL="+c.Locale,
		)
	}

	return env
}

func (c *GameConfig) windowsExecutablePath() string {
	exe := c.executablePath()

	if strings.TrimSpace(c.PrefixPath) == "" {
		return exe
	}

	driveC := c.driveCPath()

	rel, err := filepath.Rel(driveC, exe)
	if err != nil {
		return exe
	}

	if strings.HasPrefix(rel, "..") {
		return exe
	}

	return `C:\` + strings.ReplaceAll(filepath.ToSlash(rel), "/", `\`)
}

func cleanWineEnv(env []string) []string {
	out := make([]string, 0, len(env))

	for _, e := range env {
		switch {
		case strings.HasPrefix(e, "WINEARCH="):
			continue
		case strings.HasPrefix(e, "WINEPREFIX="):
			continue
		case strings.HasPrefix(e, "STEAM_COMPAT_DATA_PATH="):
			continue
		case strings.HasPrefix(e, "STEAM_COMPAT_CLIENT_INSTALL_PATH="):
			continue
		}

		out = append(out, e)
	}

	return out
}

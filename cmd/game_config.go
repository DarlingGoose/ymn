package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const configDirName = ".local/gl"

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
	Name          string             `json:"name"`
	GamePath      string             `json:"game_path"`
	Executable    string             `json:"executable"`
	WorkingDir    string             `json:"working_dir"`
	IconPath      string             `json:"icon_path,omitempty"`
	ImagePath     string             `json:"image_path,omitempty"`
	Runner        RunnerType         `json:"runner"`
	RunnerPath    string             `json:"runner_path"`
	PrefixPath    string             `json:"prefix_path,omitempty"`
	RequiresSteam bool               `json:"requires_steam"`
	SteamAppID    string             `json:"steam_app_id,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	RuntimeInfo   RuntimeStatus      `json:"runtime_info"`
	Verification  VerificationStatus `json:"verification"`
}

const verificationTimeout = 12 * time.Second

func CheckInstallations() (RuntimeStatus, error) {
	status := RuntimeStatus{}

	if winePath, err := exec.LookPath(string(RunnerWine)); err == nil {
		status.WinePath = winePath
	}
	if wineBootPath, err := exec.LookPath("wineboot"); err == nil {
		status.WineBootPath = wineBootPath
	}

	if steamPath, err := exec.LookPath("steam"); err == nil {
		status.SteamPath = steamPath
	}
	status.SteamRoot = findSteamRoot()

	protonPaths := findInstalledProtonVersions()
	status.AvailableProton = protonPaths
	if len(protonPaths) > 0 {
		status.SelectedProtonPath = protonPaths[0]
	}

	if status.WinePath == "" && status.SelectedProtonPath == "" && status.SteamPath == "" {
		return status, errors.New("no supported launcher was found; install wine, proton, or steam")
	}

	return status, nil
}

func buildGameConfig(
	inputPath string,
	requestedRunner string,
	requiresSteam bool,
	steamAppID string,
	requestedIconPath string,
	requestedImagePath string,
) (GameConfig, error) {
	runtimeStatus, err := CheckInstallations()
	if err != nil {
		return GameConfig{}, err
	}

	resolvedPath, err := filepath.Abs(inputPath)
	if err != nil {
		return GameConfig{}, fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return GameConfig{}, fmt.Errorf("stat path: %w", err)
	}

	executablePath, workingDir, err := resolveExecutable(resolvedPath, info)
	if err != nil {
		return GameConfig{}, err
	}

	gameName := deriveGameName(resolvedPath, executablePath, info.IsDir())
	assetSearchRoot := workingDir
	if info.IsDir() {
		assetSearchRoot = resolvedPath
	}
	iconPath, imagePath, err := resolveGameAssets(assetSearchRoot, requestedIconPath, requestedImagePath)
	if err != nil {
		return GameConfig{}, err
	}
	runner, runnerPath, err := selectRunner(requestedRunner, runtimeStatus, requiresSteam)
	if err != nil {
		return GameConfig{}, err
	}
	if requiresSteam && strings.TrimSpace(steamAppID) == "" {
		return GameConfig{}, errors.New("steam app id is required when a game is marked as requiring steam")
	}

	return GameConfig{
		Name:          gameName,
		GamePath:      resolvedPath,
		Executable:    executablePath,
		WorkingDir:    workingDir,
		IconPath:      iconPath,
		ImagePath:     imagePath,
		Runner:        runner,
		RunnerPath:    runnerPath,
		PrefixPath:    filepath.Join(configBaseDir(), "prefixes", sanitizeName(gameName)),
		RequiresSteam: requiresSteam,
		SteamAppID:    strings.TrimSpace(steamAppID),
		CreatedAt:     time.Now().UTC(),
		RuntimeInfo: RuntimeStatus{
			WinePath:           runtimeStatus.WinePath,
			WineBootPath:       runtimeStatus.WineBootPath,
			SteamPath:          runtimeStatus.SteamPath,
			SteamRoot:          runtimeStatus.SteamRoot,
			AvailableProton:    runtimeStatus.AvailableProton,
			SelectedProtonPath: runtimeStatus.SelectedProtonPath,
		},
	}, nil
}

func saveGameConfig(cfg GameConfig) (string, error) {
	baseDir := configBaseDir()
	configDir := filepath.Join(baseDir, "configs")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", fmt.Errorf("create config directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(baseDir, "prefixes"), 0o755); err != nil {
		return "", fmt.Errorf("create prefix directory: %w", err)
	}

	configPath := filepath.Join(configDir, sanitizeName(cfg.Name)+".json")
	existingPath, existingCfg, foundExisting, err := findExistingConfigMatch(configDir, cfg)
	if err != nil {
		return "", err
	}
	if foundExisting {
		cfg = mergeGameConfig(existingCfg, cfg)
		configPath = existingPath
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, append(data, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}

	return configPath, nil
}

func saveDesktopEntry(cfg GameConfig, launcherDir string) (string, error) {
	if strings.TrimSpace(launcherDir) == "" {
		launcherDir = defaultApplicationsDir()
	}
	if err := os.MkdirAll(launcherDir, 0o755); err != nil {
		return "", fmt.Errorf("create launcher directory: %w", err)
	}

	binaryPath, err := resolveSelfExecutable()
	if err != nil {
		return "", fmt.Errorf("resolve wgl executable: %w", err)
	}

	desktopPath := filepath.Join(launcherDir, "wgl-"+sanitizeName(cfg.Name)+".desktop")
	entry := strings.Join([]string{
		"[Desktop Entry]",
		"Version=1.0",
		"Type=Application",
		"Name=" + desktopValueEscape(cfg.Name),
		"Comment=" + desktopValueEscape(fmt.Sprintf("Launch %s with wgl", cfg.Name)),
		"Exec=" + desktopExecEscape(binaryPath) + " run-game --game " + desktopExecEscape(sanitizeName(cfg.Name)),
		"Path=" + desktopValueEscape(cfg.WorkingDir),
		"Terminal=false",
		"Categories=Game;",
		"",
	}, "\n")
	if iconPath := firstNonEmpty(cfg.IconPath, cfg.ImagePath); iconPath != "" {
		entry = strings.Replace(entry, "Categories=Game;\n", "Icon="+desktopValueEscape(iconPath)+"\nCategories=Game;\n", 1)
	}

	if err := os.WriteFile(desktopPath, []byte(entry), 0o644); err != nil {
		return "", fmt.Errorf("write desktop entry: %w", err)
	}
	return desktopPath, nil
}

func listGameConfigs() ([]GameConfig, error) {
	configDir := filepath.Join(configBaseDir(), "configs")
	entries, err := os.ReadDir(configDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config directory: %w", err)
	}

	var configs []GameConfig
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		cfg, err := loadGameConfig(filepath.Join(configDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}

	sort.Slice(configs, func(i, j int) bool {
		return strings.ToLower(configs[i].Name) < strings.ToLower(configs[j].Name)
	})
	return configs, nil
}

func findGameConfig(name string) (GameConfig, error) {
	configs, err := listGameConfigs()
	if err != nil {
		return GameConfig{}, err
	}

	if len(configs) == 0 {
		return GameConfig{}, errors.New("no saved games found in ~/.local/gl/configs")
	}

	normalizedName := strings.ToLower(strings.TrimSpace(name))
	for _, cfg := range configs {
		if strings.ToLower(cfg.Name) == normalizedName || sanitizeName(cfg.Name) == sanitizeName(normalizedName) {
			return cfg, nil
		}
	}

	return GameConfig{}, fmt.Errorf("game %q was not found", name)
}

func loadGameConfig(path string) (GameConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return GameConfig{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg GameConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return GameConfig{}, fmt.Errorf("decode config %s: %w", path, err)
	}
	return cfg, nil
}

func findExistingConfigMatch(configDir string, candidate GameConfig) (string, GameConfig, bool, error) {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", GameConfig{}, false, nil
		}
		return "", GameConfig{}, false, fmt.Errorf("read config directory: %w", err)
	}

	candidateGamePath := cleanComparablePath(candidate.GamePath)
	candidateExecutable := cleanComparablePath(candidate.Executable)
	candidateName := sanitizeName(candidate.Name)

	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}

		path := filepath.Join(configDir, entry.Name())
		cfg, err := loadGameConfig(path)
		if err != nil {
			return "", GameConfig{}, false, err
		}

		if cleanComparablePath(cfg.GamePath) == candidateGamePath && candidateGamePath != "" {
			return path, cfg, true, nil
		}
		if cleanComparablePath(cfg.Executable) == candidateExecutable && candidateExecutable != "" {
			return path, cfg, true, nil
		}
		if sanitizeName(cfg.Name) == candidateName && candidateName != "" {
			return path, cfg, true, nil
		}
	}

	return "", GameConfig{}, false, nil
}

func mergeGameConfig(existing GameConfig, incoming GameConfig) GameConfig {
	if incoming.CreatedAt.IsZero() {
		incoming.CreatedAt = existing.CreatedAt
	}
	if !existing.CreatedAt.IsZero() {
		incoming.CreatedAt = existing.CreatedAt
	}
	if strings.TrimSpace(incoming.IconPath) == "" {
		incoming.IconPath = existing.IconPath
	}
	if strings.TrimSpace(incoming.ImagePath) == "" {
		incoming.ImagePath = existing.ImagePath
	}
	return incoming
}

func launchGame(cfg GameConfig) error {
	cmd, err := prepareLaunchCommand(cfg)
	if err != nil {
		return err
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launch %s: %w", cfg.Name, err)
	}
	return nil
}

func launchGameInBackground(cfg GameConfig) error {
	cmd, err := prepareLaunchCommand(cfg)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("launch %s: %w", cfg.Name, err)
	}
	go func() {
		_ = cmd.Wait()
	}()
	return nil
}

func prepareLaunchCommand(cfg GameConfig) (*exec.Cmd, error) {
	if cfg.Runner != RunnerSteam {
		if err := os.MkdirAll(cfg.PrefixPath, 0o755); err != nil {
			return nil, fmt.Errorf("create prefix path: %w", err)
		}
		if _, err := os.Stat(cfg.Executable); err != nil {
			return nil, fmt.Errorf("game executable is unavailable: %w", err)
		}
	}

	cmd, err := buildLaunchCommand(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Runner != RunnerSteam {
		cmd.Dir = cfg.WorkingDir
	}
	return cmd, nil
}

func verifyAndAutofixGameConfig(cfg GameConfig) (GameConfig, error) {
	candidates := candidateConfigs(cfg)
	verification := VerificationStatus{}

	for _, candidate := range candidates {
		verifiedCandidate, attempts, err := verifyCandidateConfig(candidate)
		verification.Attempts = append(verification.Attempts, attempts...)
		if err == nil {
			verifiedCandidate.Verification = VerificationStatus{
				Verified:   true,
				VerifiedAt: time.Now().UTC(),
				Attempts:   verification.Attempts,
			}
			return verifiedCandidate, nil
		}
	}

	cfg.Verification = VerificationStatus{
		Verified: false,
		Attempts: verification.Attempts,
	}
	return cfg, errors.New("unable to verify the game with the available launch modes")
}

func verifyCandidateConfig(cfg GameConfig) (GameConfig, []VerificationAttempt, error) {
	if cfg.Runner == RunnerSteam {
		attempt, err := verifySteamConfig(cfg)
		return cfg, []VerificationAttempt{attempt}, err
	}

	strategies := []string{"initial-launch", "clean-prefix-retry"}
	var attempts []VerificationAttempt

	for _, strategy := range strategies {
		if strategy == "clean-prefix-retry" {
			if err := os.RemoveAll(cfg.PrefixPath); err != nil {
				attempts = append(attempts, VerificationAttempt{
					Runner:    cfg.Runner,
					Strategy:  strategy,
					Success:   false,
					Message:   fmt.Sprintf("failed to reset prefix: %v", err),
					CheckedAt: time.Now().UTC(),
				})
				return cfg, attempts, err
			}
		}

		if err := initializeRunnerPrefix(cfg); err != nil {
			attempts = append(attempts, VerificationAttempt{
				Runner:    cfg.Runner,
				Strategy:  strategy,
				Success:   false,
				Message:   fmt.Sprintf("prefix init failed: %v", err),
				CheckedAt: time.Now().UTC(),
			})
			continue
		}

		attempt, err := smokeTestGame(cfg, strategy)
		attempts = append(attempts, attempt)
		if err == nil {
			return cfg, attempts, nil
		}
	}

	return cfg, attempts, errors.New("verification failed")
}

func verifySteamConfig(cfg GameConfig) (VerificationAttempt, error) {
	if strings.TrimSpace(cfg.SteamAppID) == "" {
		return VerificationAttempt{
			Runner:    cfg.Runner,
			Strategy:  "steam-app-check",
			Success:   false,
			Message:   "steam launch requires a steam app id",
			CheckedAt: time.Now().UTC(),
		}, errors.New("steam launch requires a steam app id")
	}
	if strings.TrimSpace(cfg.RunnerPath) == "" {
		return VerificationAttempt{
			Runner:    cfg.Runner,
			Strategy:  "steam-app-check",
			Success:   false,
			Message:   "steam executable was not found",
			CheckedAt: time.Now().UTC(),
		}, errors.New("steam executable was not found")
	}
	return VerificationAttempt{
		Runner:    cfg.Runner,
		Strategy:  "steam-app-check",
		Success:   true,
		Message:   fmt.Sprintf("steam launch configured for app id %s", cfg.SteamAppID),
		CheckedAt: time.Now().UTC(),
	}, nil
}

func smokeTestGame(cfg GameConfig, strategy string) (VerificationAttempt, error) {
	ctx, cancel := context.WithTimeout(context.Background(), verificationTimeout)
	defer cancel()

	cmd, err := buildLaunchCommand(ctx, cfg)
	if err != nil {
		return VerificationAttempt{
			Runner:    cfg.Runner,
			Strategy:  strategy,
			Success:   false,
			Message:   err.Error(),
			CheckedAt: time.Now().UTC(),
		}, err
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		logPath := writeVerificationLog(cfg.Name, cfg.Runner, strategy, output.Bytes())
		attempt := VerificationAttempt{
			Runner:    cfg.Runner,
			Strategy:  strategy,
			Success:   false,
			Message:   fmt.Sprintf("launch failed: %v", err),
			LogPath:   logPath,
			CheckedAt: time.Now().UTC(),
		}
		return attempt, err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		logPath := writeVerificationLog(cfg.Name, cfg.Runner, strategy, output.Bytes())
		if err != nil {
			attempt := VerificationAttempt{
				Runner:    cfg.Runner,
				Strategy:  strategy,
				Success:   false,
				Message:   fmt.Sprintf("process exited early: %v", err),
				LogPath:   logPath,
				CheckedAt: time.Now().UTC(),
			}
			return attempt, err
		}
		attempt := VerificationAttempt{
			Runner:    cfg.Runner,
			Strategy:  strategy,
			Success:   true,
			Message:   "process exited successfully during smoke test",
			LogPath:   logPath,
			CheckedAt: time.Now().UTC(),
		}
		return attempt, nil
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		<-done
		logPath := writeVerificationLog(cfg.Name, cfg.Runner, strategy, output.Bytes())
		attempt := VerificationAttempt{
			Runner:    cfg.Runner,
			Strategy:  strategy,
			Success:   true,
			Message:   fmt.Sprintf("process stayed alive for %s", verificationTimeout),
			LogPath:   logPath,
			CheckedAt: time.Now().UTC(),
		}
		return attempt, nil
	}
}

func initializeRunnerPrefix(cfg GameConfig) error {
	if err := os.MkdirAll(cfg.PrefixPath, 0o755); err != nil {
		return fmt.Errorf("create prefix path: %w", err)
	}

	switch cfg.Runner {
	case RunnerWine:
		if strings.TrimSpace(cfg.RuntimeInfo.WineBootPath) == "" {
			return errors.New("wineboot is required for wine prefix initialization")
		}
		cmd := exec.Command(cfg.RuntimeInfo.WineBootPath, "-u")
		cmd.Env = runnerEnv(cfg, os.Environ())
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Run()
	case RunnerProton:
		return nil
	case RunnerSteam:
		return nil
	default:
		return fmt.Errorf("unsupported runner %q", cfg.Runner)
	}
}

func buildLaunchCommand(ctx context.Context, cfg GameConfig) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	switch cfg.Runner {
	case RunnerWine:
		cmd = exec.CommandContext(ctx, cfg.RunnerPath, cfg.Executable)
	case RunnerProton:
		cmd = exec.CommandContext(ctx, cfg.RunnerPath, "run", cfg.Executable)
	case RunnerSteam:
		if strings.TrimSpace(cfg.SteamAppID) == "" {
			return nil, errors.New("steam launch requires a steam app id")
		}
		cmd = exec.CommandContext(ctx, cfg.RunnerPath, "-applaunch", cfg.SteamAppID)
	default:
		return nil, fmt.Errorf("unsupported runner %q", cfg.Runner)
	}
	cmd.Env = runnerEnv(cfg, os.Environ())
	if cfg.Runner != RunnerSteam {
		cmd.Dir = cfg.WorkingDir
	}
	return cmd, nil
}

func runnerEnv(cfg GameConfig, baseEnv []string) []string {
	env := append([]string{}, baseEnv...)
	switch cfg.Runner {
	case RunnerWine:
		env = append(env, "WINEPREFIX="+cfg.PrefixPath)
	case RunnerProton:
		env = append(env, "STEAM_COMPAT_DATA_PATH="+cfg.PrefixPath)
		if cfg.RuntimeInfo.SteamRoot != "" {
			env = append(env, "STEAM_COMPAT_CLIENT_INSTALL_PATH="+cfg.RuntimeInfo.SteamRoot)
		}
	case RunnerSteam:
		if strings.TrimSpace(cfg.SteamAppID) != "" {
			env = append(env, "SteamAppId="+cfg.SteamAppID, "SteamGameId="+cfg.SteamAppID)
		}
	}
	return env
}

func candidateConfigs(cfg GameConfig) []GameConfig {
	if cfg.RequiresSteam {
		return []GameConfig{cfg}
	}

	candidates := []GameConfig{cfg}
	alternates := []struct {
		runner RunnerType
		path   string
	}{
		{runner: RunnerProton, path: cfg.RuntimeInfo.SelectedProtonPath},
		{runner: RunnerWine, path: cfg.RuntimeInfo.WinePath},
	}

	for _, alternate := range alternates {
		if alternate.runner == cfg.Runner || alternate.path == "" {
			continue
		}
		alt := cfg
		alt.Runner = alternate.runner
		alt.RunnerPath = alternate.path
		alt.PrefixPath = cfg.PrefixPath + "-" + string(alternate.runner)
		candidates = append(candidates, alt)
	}
	return candidates
}

func writeVerificationLog(gameName string, runner RunnerType, strategy string, data []byte) string {
	logDir := filepath.Join(configBaseDir(), "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return ""
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s-%s-%s.log", sanitizeName(gameName), runner, strategy))
	if len(data) == 0 {
		data = []byte("no output captured\n")
	}
	if err := os.WriteFile(logPath, data, 0o644); err != nil {
		return ""
	}
	return logPath
}

func resolveExecutable(resolvedPath string, info os.FileInfo) (string, string, error) {
	if !info.IsDir() {
		if !isExeFile(resolvedPath) {
			return "", "", fmt.Errorf("path must point to a directory or .exe file: %s", resolvedPath)
		}
		return resolvedPath, filepath.Dir(resolvedPath), nil
	}

	var candidates []string
	err := filepath.Walk(resolvedPath, func(path string, fileInfo os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fileInfo == nil || fileInfo.IsDir() {
			return nil
		}
		if !isExeFile(path) {
			return nil
		}
		candidates = append(candidates, path)
		return nil
	})
	if err != nil {
		return "", "", fmt.Errorf("scan directory for executables: %w", err)
	}

	if len(candidates) == 0 {
		return "", "", fmt.Errorf("no .exe files found in %s", resolvedPath)
	}

	sort.Strings(candidates)
	return candidates[0], filepath.Dir(candidates[0]), nil
}

func resolveGameAssets(searchRoot, requestedIconPath, requestedImagePath string) (string, string, error) {
	iconPath, err := resolveOptionalAssetPath(requestedIconPath)
	if err != nil {
		return "", "", fmt.Errorf("resolve icon path: %w", err)
	}
	imagePath, err := resolveOptionalAssetPath(requestedImagePath)
	if err != nil {
		return "", "", fmt.Errorf("resolve image path: %w", err)
	}

	if iconPath != "" && imagePath != "" {
		return iconPath, imagePath, nil
	}

	autoIconPath, autoImagePath := findGameAssets(searchRoot)
	if iconPath == "" {
		iconPath = autoIconPath
	}
	if imagePath == "" {
		imagePath = autoImagePath
	}
	return iconPath, imagePath, nil
}

func resolveOptionalAssetPath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}

	resolvedPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("expected a file, got directory: %s", resolvedPath)
	}
	if !isImageFile(resolvedPath) {
		return "", fmt.Errorf("unsupported image file type: %s", resolvedPath)
	}
	return resolvedPath, nil
}

func findGameAssets(searchRoot string) (string, string) {
	type scoredAsset struct {
		path  string
		score int
	}

	var iconCandidates []scoredAsset
	var imageCandidates []scoredAsset

	_ = filepath.Walk(searchRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !isImageFile(path) {
			return nil
		}

		iconScore := scoreIconCandidate(searchRoot, path)
		if iconScore > 0 {
			iconCandidates = append(iconCandidates, scoredAsset{path: path, score: iconScore})
		}

		imageScore := scoreImageCandidate(searchRoot, path)
		if imageScore > 0 {
			imageCandidates = append(imageCandidates, scoredAsset{path: path, score: imageScore})
		}
		return nil
	})

	bestPath := func(candidates []scoredAsset) string {
		if len(candidates) == 0 {
			return ""
		}
		sort.Slice(candidates, func(i, j int) bool {
			if candidates[i].score == candidates[j].score {
				return candidates[i].path < candidates[j].path
			}
			return candidates[i].score > candidates[j].score
		})
		return candidates[0].path
	}

	iconPath := bestPath(iconCandidates)
	imagePath := bestPath(imageCandidates)
	if imagePath == "" {
		imagePath = iconPath
	}
	if iconPath == "" {
		iconPath = imagePath
	}
	return iconPath, imagePath
}

func scoreIconCandidate(searchRoot, path string) int {
	score := scoreSharedAssetCandidate(searchRoot, path)
	name := strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))

	switch {
	case strings.Contains(name, "icon"):
		score += 120
	case strings.Contains(name, "logo"):
		score += 100
	case strings.Contains(name, "favicon"):
		score += 90
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".ico":
		score += 60
	case ".png":
		score += 25
	case ".svg":
		score += 20
	}

	if strings.Contains(name, "banner") || strings.Contains(name, "cover") || strings.Contains(name, "screenshot") {
		score -= 40
	}
	return score
}

func scoreImageCandidate(searchRoot, path string) int {
	score := scoreSharedAssetCandidate(searchRoot, path)
	name := strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))

	switch {
	case strings.Contains(name, "cover"):
		score += 120
	case strings.Contains(name, "poster"):
		score += 100
	case strings.Contains(name, "banner"):
		score += 90
	case strings.Contains(name, "hero"):
		score += 80
	case strings.Contains(name, "art"):
		score += 60
	case strings.Contains(name, "image"):
		score += 40
	}

	if strings.Contains(name, "icon") || strings.Contains(name, "favicon") {
		score -= 35
	}
	if strings.Contains(name, "screenshot") || strings.Contains(name, "thumb") {
		score -= 20
	}
	return score
}

func scoreSharedAssetCandidate(searchRoot, path string) int {
	relativePath, err := filepath.Rel(searchRoot, path)
	if err != nil {
		relativePath = path
	}

	score := 10
	depth := strings.Count(filepath.Clean(relativePath), string(os.PathSeparator))
	score -= depth * 5

	lowerPath := strings.ToLower(relativePath)
	for _, token := range []string{"assets", "images", "img", "artwork", "media"} {
		if strings.Contains(lowerPath, token) {
			score += 15
			break
		}
	}

	return score
}

func selectRunner(requestedRunner string, status RuntimeStatus, requiresSteam bool) (RunnerType, string, error) {
	switch strings.ToLower(strings.TrimSpace(requestedRunner)) {
	case "", "auto":
		if requiresSteam {
			if status.SteamPath == "" {
				return "", "", errors.New("steam is required for this game but steam is not installed")
			}
			return RunnerSteam, status.SteamPath, nil
		}
		if status.SelectedProtonPath != "" {
			return RunnerProton, status.SelectedProtonPath, nil
		}
		if status.WinePath != "" {
			return RunnerWine, status.WinePath, nil
		}
		if status.SteamPath != "" {
			return RunnerSteam, status.SteamPath, nil
		}
	case string(RunnerProton):
		if status.SelectedProtonPath == "" {
			return "", "", errors.New("proton requested but no proton installation was found")
		}
		if requiresSteam {
			return "", "", errors.New("games marked as requiring steam must use the steam runner")
		}
		return RunnerProton, status.SelectedProtonPath, nil
	case string(RunnerWine):
		if status.WinePath == "" {
			return "", "", errors.New("wine requested but wine is not installed")
		}
		if requiresSteam {
			return "", "", errors.New("games marked as requiring steam must use the steam runner")
		}
		return RunnerWine, status.WinePath, nil
	case string(RunnerSteam):
		if status.SteamPath == "" {
			return "", "", errors.New("steam requested but steam is not installed")
		}
		return RunnerSteam, status.SteamPath, nil
	default:
		return "", "", fmt.Errorf("unsupported runner %q; expected auto, wine, proton, or steam", requestedRunner)
	}

	return "", "", errors.New("no compatible runner installation found")
}

func findInstalledProtonVersions() []string {
	searchRoots := steamCommonRoots()
	var protonPaths []string
	for _, root := range searchRoots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := strings.ToLower(entry.Name())
			if !strings.Contains(name, "proton") {
				continue
			}
			candidate := filepath.Join(root, entry.Name(), "proton")
			if fileInfo, err := os.Stat(candidate); err == nil && !fileInfo.IsDir() {
				protonPaths = append(protonPaths, candidate)
			}
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(protonPaths)))
	return protonPaths
}

func findSteamRoot() string {
	roots := findSteamRoots()
	if len(roots) == 0 {
		return ""
	}
	return roots[0]
}

func findSteamRoots() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(homeDir, ".steam", "steam"),
		filepath.Join(homeDir, ".local", "share", "Steam"),
		filepath.Join(homeDir, ".var", "app", "com.valvesoftware.Steam", ".local", "share", "Steam"),
	}
	var roots []string
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			roots = append(roots, candidate)
		}
	}
	return roots
}

func steamCommonRoots() []string {
	steamRoots := findSteamRoots()
	if len(steamRoots) == 0 {
		return nil
	}
	roots := make([]string, 0, len(steamRoots))
	for _, steamRoot := range steamRoots {
		roots = append(roots, filepath.Join(steamRoot, "steamapps", "common"))
	}
	return roots
}

func configBaseDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return configDirName
	}
	return filepath.Join(homeDir, configDirName)
}

func defaultApplicationsDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".local", "share", "applications")
	}
	return filepath.Join(homeDir, ".local", "share", "applications")
}

func resolveSelfExecutable() (string, error) {
	if executablePath, err := os.Executable(); err == nil && strings.TrimSpace(executablePath) != "" {
		return executablePath, nil
	}
	return exec.LookPath(os.Args[0])
}

func desktopValueEscape(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\n", "\\n")
	return replacer.Replace(strings.TrimSpace(value))
}

func desktopExecEscape(value string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		" ", "\\ ",
		"\t", "\\\t",
		"\n", "\\\n",
		"\"", "\\\"",
		"'", "\\'",
		">", "\\>",
		"<", "\\<",
		"~", "\\~",
		"|", "\\|",
		"&", "\\&",
		";", "\\;",
		"$", "\\$",
		"*", "\\*",
		"?", "\\?",
		"#", "\\#",
		"(", "\\(",
		")", "\\)",
		"`", "\\`",
	)
	return replacer.Replace(strings.TrimSpace(value))
}

func deriveGameName(inputPath, executablePath string, inputWasDir bool) string {
	if inputWasDir {
		return filepath.Base(inputPath)
	}
	return strings.TrimSuffix(filepath.Base(executablePath), filepath.Ext(executablePath))
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", "_", "-")
	name = replacer.Replace(name)
	var builder strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			builder.WriteRune(r)
		}
	}
	sanitized := strings.Trim(builder.String(), "-")
	if sanitized == "" {
		return "game"
	}
	return sanitized
}

func isExeFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".exe")
}

func isImageFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".webp", ".bmp", ".gif", ".ico", ".svg":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func cleanComparablePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

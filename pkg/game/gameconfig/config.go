package gameconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Seann-Moser/wgl/pkg/util"
)

func ListConfigs() ([]GameConfig, error) {
	configDir := filepath.Join(util.ConfigBaseDir(), "games")

	entries, err := os.ReadDir(configDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config dir: %w", err)
	}

	configs := make([]GameConfig, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}

		path := filepath.Join(configDir, entry.Name())

		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config %s: %w", path, err)
		}

		var cfg GameConfig
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}

		configs = append(configs, cfg)
	}

	sort.Slice(configs, func(i, j int) bool {
		return strings.ToLower(configs[i].Name) < strings.ToLower(configs[j].Name)
	})

	return configs, nil
}

func FindConfig(name string) (*GameConfig, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("config name is required")
	}

	configDir := filepath.Join(util.ConfigBaseDir(), "games")

	entries, err := os.ReadDir(configDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no game configs found in %s", configDir)
		}
		return nil, fmt.Errorf("read config dir: %w", err)
	}

	wanted := util.SanitizeName(name)

	var matches []GameConfig

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}

		path := filepath.Join(configDir, entry.Name())

		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config %s: %w", path, err)
		}

		var cfg GameConfig
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", path, err)
		}

		fileName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		if strings.EqualFold(fileName, wanted) ||
			strings.EqualFold(util.SanitizeName(cfg.Name), wanted) ||
			strings.EqualFold(cfg.Name, name) {
			matches = append(matches, cfg)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("config %q not found in %s", name, configDir)
	}

	if len(matches) > 1 {
		names := make([]string, 0, len(matches))
		for _, match := range matches {
			names = append(names, match.Name)
		}
		sort.Strings(names)

		return nil, fmt.Errorf("config %q is ambiguous; matched: %s", name, strings.Join(names, ", "))
	}

	return &matches[0], nil
}

func SaveGameConfig(config GameConfig) (path string, err error) {
	if strings.TrimSpace(config.Name) == "" {
		return "", errors.New("config name is required")
	}

	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now().UTC()
	}

	configDir := filepath.Join(util.ConfigBaseDir(), "games")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	fileName := util.SanitizeName(config.Name) + ".json"
	path = filepath.Join(configDir, fileName)

	raw, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal config: %w", err)
	}

	raw = append(raw, '\n')

	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}

	return path, nil
}

func BuildGameConfig(
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

	gameName := util.DeriveGameName(resolvedPath, executablePath, info.IsDir())
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
		PrefixPath:    filepath.Join(util.ConfigBaseDir(), "prefixes", util.SanitizeName(gameName)),
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

func resolveExecutable(resolvedPath string, info os.FileInfo) (string, string, error) {
	if !info.IsDir() {
		if !util.IsExeFile(resolvedPath) {
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
		if !util.IsExeFile(path) {
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
	if !util.IsImageFile(resolvedPath) {
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
		if !util.IsImageFile(path) {
			return nil
		}

		iconScore := util.ScoreIconCandidate(searchRoot, path)
		if iconScore > 0 {
			iconCandidates = append(iconCandidates, scoredAsset{path: path, score: iconScore})
		}

		imageScore := util.ScoreImageCandidate(searchRoot, path)
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

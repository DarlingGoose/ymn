package gameconfig

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Seann-Moser/wgl/pkg/util"
)

type ExecutableKind string

const (
	ExecutableGame      ExecutableKind = "game"
	ExecutableInstaller ExecutableKind = "installer"
	ExecutableRuntime   ExecutableKind = "runtime"
)

type ExecutableCandidate struct {
	Path   string
	Kind   ExecutableKind
	Reason string
	Score  int
}

func classifyExecutable(path string) ExecutableCandidate {
	name := strings.ToLower(filepath.Base(path))
	dir := strings.ToLower(filepath.ToSlash(filepath.Dir(path)))

	c := ExecutableCandidate{
		Path:  path,
		Kind:  ExecutableGame,
		Score: 100,
	}

	switch {
	case strings.Contains(dir, "/codec"):
		c.Kind = ExecutableRuntime
		c.Reason = "inside codec directory"
		c.Score = -100
		return c

	case strings.Contains(name, "codec") ||
		strings.Contains(name, "wmp") ||
		strings.Contains(name, "wm9") ||
		strings.Contains(name, "wmv") ||
		strings.Contains(name, "wmfdist"):
		c.Kind = ExecutableRuntime
		c.Reason = "looks like Windows Media codec/runtime installer"
		c.Score = -100
		return c

	case name == "setup.exe" ||
		name == "install.exe" ||
		name == "installer.exe":
		c.Kind = ExecutableInstaller
		c.Reason = "looks like installer executable"
		c.Score = -100
		return c

	case name == "instmsia.exe" ||
		name == "instmsiw.exe" ||
		name == "dxsetup.exe" ||
		strings.HasPrefix(name, "vcredist"):
		c.Kind = ExecutableRuntime
		c.Reason = "looks like dependency/runtime installer"
		c.Score = -100
		return c

	case strings.HasPrefix(name, "unins"):
		c.Kind = ExecutableInstaller
		c.Reason = "looks like uninstaller"
		c.Score = -100
		return c
	}

	return c
}

func resolveExecutable(resolvedPath string, info os.FileInfo) (string, string, error) {
	if !info.IsDir() {
		if !util.IsExeFile(resolvedPath) {
			return "", "", fmt.Errorf("path must point to a directory or .exe file: %s", resolvedPath)
		}

		candidate := classifyExecutable(resolvedPath)
		if candidate.Kind != ExecutableGame {
			return "", "", fmt.Errorf(
				"%s does not look like the game executable: %s",
				resolvedPath,
				candidate.Reason,
			)
		}

		return resolvedPath, filepath.Dir(resolvedPath), nil
	}

	var candidates []ExecutableCandidate

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

		candidates = append(candidates, classifyExecutable(path))
		return nil
	})
	if err != nil {
		return "", "", fmt.Errorf("scan directory for executables: %w", err)
	}

	if len(candidates) == 0 {
		return "", "", fmt.Errorf("no .exe files found in %s", resolvedPath)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Path < candidates[j].Path
		}
		return candidates[i].Score > candidates[j].Score
	})

	best := candidates[0]
	if best.Kind != ExecutableGame {
		var found []string
		for _, c := range candidates {
			found = append(found, fmt.Sprintf("%s [%s: %s]", c.Path, c.Kind, c.Reason))
		}

		return "", "", fmt.Errorf(
			"no likely game executable found; only installer/runtime executables were detected:\n%s",
			strings.Join(found, "\n"),
		)
	}

	return best.Path, filepath.Dir(best.Path), nil
}

func scanInstallerArtifacts(root string) InstallerScanResult {
	var result InstallerScanResult

	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			return nil
		}

		name := strings.ToLower(filepath.Base(path))
		ext := strings.ToLower(filepath.Ext(path))

		if util.IsExeFile(path) {
			c := classifyExecutable(path)
			if c.Kind == ExecutableGame {
				result.GameExecutables = append(result.GameExecutables, c)
			}
		}

		switch {
		case name == "setup.exe":
			result.Installers = append(result.Installers, InstallerArtifact{
				Path:  path,
				Type:  "setup",
				Score: 100,
			})

		case ext == ".msi":
			result.Installers = append(result.Installers, InstallerArtifact{
				Path:  path,
				Type:  "msi",
				Score: 90,
			})

		case name == "install.exe" || strings.Contains(name, "installer"):
			result.Installers = append(result.Installers, InstallerArtifact{
				Path:  path,
				Type:  "setup",
				Score: 80,
			})

		case name == "instmsia.exe" || name == "instmsiw.exe":
			result.Installers = append(result.Installers, InstallerArtifact{
				Path:  path,
				Type:  "runtime",
				Score: 10,
			})
		}

		return nil
	})

	return result
}

type InstallerArtifact struct {
	Path  string
	Type  string // "setup", "msi", "runtime", "codec"
	Score int
}

type InstallerScanResult struct {
	HasInstaller     bool
	HasMSI           bool
	HasCodec         bool
	HasRuntime       bool
	MSIFiles         []string
	RuntimeFiles     []string // vcredist, dxsetup, etc
	CodecFiles       []string
	SuspiciousReason []string // human readable
	Installers       []InstallerArtifact
	GameExecutables  []ExecutableCandidate
}

func (r InstallerScanResult) HasAny() bool {
	return r.HasInstaller || r.HasMSI || r.HasCodec || r.HasRuntime
}

func (r InstallerScanResult) LooksInstallerOnly() bool {
	return (r.HasMSI || r.HasInstaller) && len(r.GameExecutables) == 0
}

func (r InstallerScanResult) BestInstaller() (InstallerArtifact, bool) {
	var candidates []InstallerArtifact

	for _, installer := range r.Installers {
		switch installer.Type {
		case "setup", "msi":
			candidates = append(candidates, installer)
		}
	}

	if len(candidates) == 0 {
		return InstallerArtifact{}, false
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return candidates[0], true
}

func RunInstaller(
	cfg GameConfig,
	installer InstallerArtifact,
) error {
	switch cfg.Runner {
	case RunnerWine:
		return runWineInstaller(cfg, installer)

	case RunnerProton:
		return runProtonInstaller(cfg, installer)

	default:
		return fmt.Errorf("installer flow only supports wine/proton, got %q", cfg.Runner)
	}
}
func runWineInstaller(cfg GameConfig, installer InstallerArtifact) error {
	wine := util.FirstNonEmpty(cfg.RunnerPath, cfg.RuntimeInfo.WinePath, "wine")

	args := []string{installer.Path}
	if installer.Type == "msi" {
		args = []string{"msiexec", "/i", installer.Path}
	}

	cmd := exec.Command(wine, args...)
	cmd.Dir = filepath.Dir(installer.Path)
	cmd.Env = cfg.baseEnv()
	cmd.Env = append(cmd.Env, "WINEPREFIX="+cfg.PrefixPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
func runProtonInstaller(cfg GameConfig, installer InstallerArtifact) error {
	proton := util.FirstNonEmpty(
		cfg.RunnerPath,
		cfg.RuntimeInfo.SelectedProtonPath,
	)

	if proton == "" {
		return errors.New("proton path is required")
	}

	args := []string{"run", installer.Path}
	if installer.Type == "msi" {
		args = []string{"run", "msiexec", "/i", installer.Path}
	}

	cmd := exec.Command(proton, args...)
	cmd.Dir = filepath.Dir(installer.Path)
	cmd.Env = cfg.baseEnv()
	cmd.Env = append(cmd.Env,
		"STEAM_COMPAT_DATA_PATH="+cfg.PrefixPath,
		"STEAM_COMPAT_CLIENT_INSTALL_PATH="+util.FirstNonEmpty(
			cfg.RuntimeInfo.SteamRoot,
			filepath.Join(os.Getenv("HOME"), ".steam", "steam"),
		),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func findInstalledExecutable(prefixPath string) (string, error) {
	roots := []string{
		filepath.Join(prefixPath, "pfx", "drive_c", "Program Files"),
		filepath.Join(prefixPath, "pfx", "drive_c", "Program Files (x86)"),
		filepath.Join(prefixPath, "drive_c", "Program Files"),
		filepath.Join(prefixPath, "drive_c", "Program Files (x86)"),
	}

	var candidates []ExecutableCandidate

	for _, root := range roots {
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if !util.IsExeFile(path) {
				return nil
			}

			c := classifyExecutable(path)
			if c.Kind == ExecutableGame {
				candidates = append(candidates, c)
			}

			return nil
		})
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("no installed game executable found in prefix: %s", prefixPath)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Path < candidates[j].Path
		}
		return candidates[i].Score > candidates[j].Score
	})

	return candidates[0].Path, nil
}

func shouldRunInstaller(info os.FileInfo, scan InstallerScanResult, selectedExe string, root string) bool {
	if !info.IsDir() {
		return false
	}

	if len(scan.Installers) == 0 {
		return false
	}

	selected := classifyExecutable(selectedExe)
	if selected.Kind == ExecutableGame && selected.Score > 0 {
		return false
	}

	if hasLikelyInstalledLayout(root, selectedExe) {
		return false
	}

	return true
}

func hasLikelyInstalledLayout(root, exe string) bool {
	root = strings.TrimSpace(root)
	exe = strings.TrimSpace(exe)

	if root == "" || exe == "" {
		return false
	}

	rel, err := filepath.Rel(root, exe)
	if err != nil {
		return false
	}

	rel = strings.ToLower(filepath.ToSlash(rel))

	return strings.HasPrefix(rel, "program files/") ||
		strings.HasPrefix(rel, "program files (x86)/") ||
		strings.Contains(rel, "/program files/") ||
		strings.Contains(rel, "/program files (x86)/")
}

func stageGameIntoPrefix(cfg *GameConfig) error {
	if strings.TrimSpace(cfg.PrefixPath) == "" {
		return errors.New("prefix path is required for staging")
	}
	if strings.TrimSpace(cfg.Executable) == "" {
		return errors.New("executable path is required for staging")
	}

	sourceDir := cfg.WorkingDir
	if strings.TrimSpace(sourceDir) == "" {
		sourceDir = filepath.Dir(cfg.Executable)
	}

	if sourceDir == "" || sourceDir == "." {
		return fmt.Errorf("invalid source dir for staging: %q", sourceDir)
	}

	stagedDir := filepath.Join(
		cfg.driveCPath(),
		"Games",
		util.SanitizeName(cfg.Name),
	)

	if err := os.RemoveAll(stagedDir); err != nil {
		return fmt.Errorf("remove old staged game: %w", err)
	}
	if err := copyDir(sourceDir, stagedDir); err != nil {
		return fmt.Errorf("stage game into prefix: %w", err)
	}

	stagedExe := filepath.Join(stagedDir, filepath.Base(cfg.Executable))
	if _, err := os.Stat(stagedExe); err != nil {
		return fmt.Errorf("staged executable missing: %s: %w", stagedExe, err)
	}

	cfg.StagedPath = stagedDir
	cfg.GamePath = stagedDir
	cfg.Executable = stagedExe
	cfg.WorkingDir = stagedDir

	return nil
}

func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()

	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

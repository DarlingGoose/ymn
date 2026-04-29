package gameconfig

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

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

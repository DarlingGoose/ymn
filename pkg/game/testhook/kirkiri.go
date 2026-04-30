package testhook

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type KirikiriHook struct{}

func (h *KirikiriHook) Detect(inputPath string) (TextHookStatus, error) {
	projectRoot, err := resolveKirikiriProjectRoot(inputPath)
	if err != nil {
		return TextHookStatus{
			Supported: false,
			Installed: false,
			Loaded:    false,
			Engine:    string(EngineUnknown),
			Method:    string(MethodScriptPatch),
			Message:   err.Error(),
		}, nil
	}

	report := inspectKirikiriCompatibility(projectRoot)

	pluginPath := filepath.Join(projectRoot, "wgl_text_hook.tjs")
	outputPath := filepath.Join(projectRoot, "wgl-dialogue.log")

	installed := fileExists(pluginPath)
	loaded := installed && kirikiriHookAppearsLoaded(projectRoot)

	message := "Kirikiri/KAG text hook is not installed."
	switch {
	case installed && loaded:
		message = "Kirikiri/KAG text hook is installed and appears to be loaded by the game scripts."
	case installed && !loaded:
		message = "Kirikiri/KAG text hook file exists, but no script load/injection point was detected."
		report.RiskLevel = maxRisk(report.RiskLevel, "risky")
		report.Findings = append(report.Findings,
			"wgl_text_hook.tjs exists but no known script references it",
		)
	}

	return TextHookStatus{
		Supported:     true,
		Installed:     installed,
		Loaded:        loaded,
		Engine:        string(EngineKirikiri),
		Method:        string(MethodScriptPatch),
		ProjectRoot:   projectRoot,
		PluginPath:    pluginPath,
		OutputPath:    outputPath,
		Compatibility: report,
		Message:       message,
	}, nil
}

func (h *KirikiriHook) IsInstalled(inputPath string) (bool, error) {
	status, err := h.Detect(inputPath)
	if err != nil {
		return false, err
	}
	return status.Supported && status.Installed && status.Loaded, nil
}

func (h *KirikiriHook) InstallHook(inputPath string) (TextHookInstallResult, error) {
	projectRoot, err := resolveKirikiriProjectRoot(inputPath)
	if err != nil {
		return TextHookInstallResult{}, err
	}

	report := inspectKirikiriCompatibility(projectRoot)

	pluginPath := filepath.Join(projectRoot, "wgl_text_hook.tjs")
	outputPath := filepath.Join(projectRoot, "wgl-dialogue.log")

	if err := os.WriteFile(pluginPath, []byte(kirikiriTextHookTJS(outputPath)), 0o644); err != nil {
		return TextHookInstallResult{}, fmt.Errorf("write Kirikiri hook: %w", err)
	}

	report.Findings = append(report.Findings,
		"wrote wgl_text_hook.tjs; game still needs a TJS load/injection point",
	)

	return TextHookInstallResult{
		Engine:        string(EngineKirikiri),
		Method:        string(MethodScriptPatch),
		PluginPath:    pluginPath,
		OutputPath:    outputPath,
		Compatibility: report,
	}, nil
}

func inspectKirikiriCompatibility(projectRoot string) TextHookCompatibilityReport {
	report := TextHookCompatibilityReport{
		ProjectRoot: filepath.Clean(projectRoot),
		RiskLevel:   "warn",
	}

	var hasXP3 bool
	var hasLooseTJS bool
	var hasLooseKS bool
	var hasStartupTJS bool
	var hasMainWindowTJS bool

	_ = filepath.WalkDir(projectRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}

		name := strings.ToLower(filepath.Base(path))
		ext := strings.ToLower(filepath.Ext(path))

		switch ext {
		case ".xp3":
			hasXP3 = true
		case ".tjs":
			hasLooseTJS = true
		case ".ks":
			hasLooseKS = true
		}

		if name == "startup.tjs" {
			hasStartupTJS = true
		}
		if name == "mainwindow.tjs" {
			hasMainWindowTJS = true
		}

		return nil
	})

	if hasXP3 {
		report.Findings = append(report.Findings, "XP3 archive detected; scripts may be packed and harder to patch safely")
	}
	if hasLooseTJS {
		report.Findings = append(report.Findings, "loose TJS scripts detected")
	}
	if hasLooseKS {
		report.Findings = append(report.Findings, "loose KAG scenario scripts detected")
	}
	if hasStartupTJS {
		report.Findings = append(report.Findings, "startup.tjs detected; possible script injection point")
	}
	if hasMainWindowTJS {
		report.Findings = append(report.Findings, "MainWindow.tjs detected; possible KAG message-layer hook point")
	}

	if hasStartupTJS || hasMainWindowTJS {
		report.RiskLevel = "warn"
	} else {
		report.RiskLevel = "risky"
		report.Findings = append(report.Findings, "no obvious loose TJS injection point found; external hooker may be safer")
	}

	return report
}

func resolveKirikiriProjectRoot(inputPath string) (string, error) {
	resolvedPath, err := filepath.Abs(strings.TrimSpace(inputPath))
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return "", fmt.Errorf("stat path: %w", err)
	}

	root := resolvedPath
	if !info.IsDir() {
		root = filepath.Dir(resolvedPath)
	}

	found := false

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}

		name := strings.ToLower(filepath.Base(path))
		ext := strings.ToLower(filepath.Ext(path))

		switch {
		case name == "data.xp3":
			found = true
			return filepath.SkipAll
		case name == "startup.tjs":
			found = true
			return filepath.SkipAll
		case name == "mainwindow.tjs":
			found = true
			return filepath.SkipAll
		case ext == ".xp3":
			found = true
			return filepath.SkipAll
		case ext == ".ks":
			found = true
			return filepath.SkipAll
		case ext == ".tjs":
			found = true
			return filepath.SkipAll
		}

		return nil
	})

	if !found {
		return "", fmt.Errorf("could not find Kirikiri/KAG project under %s", resolvedPath)
	}

	return root, nil
}

//go:embed plugins/wgl_text_hook.tjs
var kirikiriTextHook string

func kirikiriTextHookTJS(outputPath string) string {
	escaped := strings.ReplaceAll(outputPath, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)

	return fmt.Sprintf(kirikiriTextHook, escaped)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func kirikiriHasObviousInjectionPoint(report TextHookCompatibilityReport) bool {
	for _, finding := range report.Findings {
		lower := strings.ToLower(finding)
		if strings.Contains(lower, "startup.tjs detected") ||
			strings.Contains(lower, "mainwindow.tjs detected") {
			return true
		}
	}
	return false
}

func maxRisk(current, next string) string {
	rank := map[string]int{
		"safe":        0,
		"warn":        1,
		"risky":       2,
		"unsupported": 3,
	}

	if rank[next] > rank[current] {
		return next
	}
	return current
}

func kirikiriHookAppearsLoaded(projectRoot string) bool {
	const hookName = "wgl_text_hook.tjs"

	found := false

	_ = filepath.WalkDir(projectRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil || d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".tjs" && ext != ".ks" {
			return nil
		}

		// Do not count the hook file referencing itself.
		if strings.EqualFold(filepath.Base(path), hookName) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := strings.ToLower(string(data))
		if strings.Contains(content, strings.ToLower(hookName)) ||
			strings.Contains(content, "wgltexthookappend") {
			found = true
			return filepath.SkipAll
		}

		return nil
	})

	return found
}

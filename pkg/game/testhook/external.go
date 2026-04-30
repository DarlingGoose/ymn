package testhook

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ExternalTextHook struct{}

func (h *ExternalTextHook) Detect(inputPath string) (TextHookStatus, error) {
	resolvedPath, err := filepath.Abs(strings.TrimSpace(inputPath))
	if err != nil {
		return TextHookStatus{}, fmt.Errorf("resolve path: %w", err)
	}

	projectRoot := resolvedPath
	if info, statErr := os.Stat(resolvedPath); statErr == nil && !info.IsDir() {
		projectRoot = filepath.Dir(resolvedPath)
	}

	report := TextHookCompatibilityReport{
		ProjectRoot: filepath.Clean(projectRoot),
		RiskLevel:   "unsupported",
		Findings: []string{
			"no supported in-engine text hook target was detected",
			"external process text hooking may be required",
			"in-engine transcript logging is unavailable for this engine",
		},
	}

	return TextHookStatus{
		Supported:     false,
		Installed:     false,
		Loaded:        false,
		Engine:        string(EngineUnknown),
		Method:        string(MethodExternalHook),
		ProjectRoot:   filepath.Clean(projectRoot),
		Compatibility: report,
		Message:       "No supported in-engine text hook was detected; use an external hooker fallback.",
	}, nil
}

func (h *ExternalTextHook) IsInstalled(inputPath string) (bool, error) {
	status, err := h.Detect(inputPath)
	if err != nil {
		return false, err
	}
	return status.Supported && status.Installed && status.Loaded, nil
}

func (h *ExternalTextHook) InstallHook(inputPath string) (TextHookInstallResult, error) {
	status, err := h.Detect(inputPath)
	if err != nil {
		return TextHookInstallResult{}, err
	}

	return TextHookInstallResult{
		Engine:        status.Engine,
		Method:        status.Method,
		OutputPath:    status.OutputPath,
		Compatibility: status.Compatibility,
	}, nil
}

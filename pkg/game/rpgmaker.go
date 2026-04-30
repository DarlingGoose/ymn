package game

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Seann-Moser/wgl/pkg/util"
)

const (
	textHookMethodAuto       = "auto"
	textHookMethodMod        = "mod"
	textHookMethodTextractor = "textractor"
	rpgMakerClipboardPlugin  = "WGLClipboardText"
)

//go:embed plugins/rpgmakerPlugin.js
var rpgMakerClipboardPluginSource string

type RPGMakerHook struct {
}

//func (h *RPGMakerHook) resolveInstallTextHookTarget(args []string) (string, error) {
//	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
//		return args[0], nil
//	}
//
//	selectedName := strings.TrimSpace(installTextHookGame)
//	if selectedName != "" {
//		cfg, err := findGameConfig(selectedName)
//		if err != nil {
//			return "", err
//		}
//		return firstNonEmpty(cfg.GamePath, cfg.Executable), nil
//	}
//
//	cfg, err := selectGameConfigWithTUI("Select a game to install a text hook", "install the text hook")
//	if err != nil {
//		return "", err
//	}
//	return firstNonEmpty(cfg.GamePath, cfg.Executable), nil
//}

func (h *RPGMakerHook) InstallHook(inputPath string) (TextHookInstallResult, error) {
	projectRoot, engine, err := resolveRPGMakerProjectRoot(inputPath)
	if err != nil {
		return TextHookInstallResult{}, err
	}

	compatibility, err := inspectRPGMakerTextHookCompatibility(projectRoot)
	if err != nil {
		return TextHookInstallResult{}, err
	}

	jsDir := filepath.Join(projectRoot, "js")
	pluginsDir := filepath.Join(jsDir, "plugins")
	pluginsConfigPath := filepath.Join(jsDir, "plugins.js")
	pluginPath := filepath.Join(pluginsDir, rpgMakerClipboardPlugin+".js")

	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return TextHookInstallResult{}, fmt.Errorf("create plugins directory: %w", err)
	}

	if err := os.WriteFile(pluginPath, []byte(rpgMakerClipboardPluginSource), 0o644); err != nil {
		return TextHookInstallResult{}, fmt.Errorf("write plugin: %w", err)
	}

	if err := ensureRPGMakerPluginEnabled(pluginsConfigPath, rpgMakerClipboardPlugin); err != nil {
		return TextHookInstallResult{}, err
	}

	return TextHookInstallResult{
		Engine:            engine,
		PluginPath:        pluginPath,
		PluginsConfigPath: pluginsConfigPath,
		Compatibility:     compatibility,
	}, nil
}

func resolveRPGMakerProjectRoot(inputPath string) (string, string, error) {
	resolvedPath, err := filepath.Abs(strings.TrimSpace(inputPath))
	if err != nil {
		return "", "", fmt.Errorf("resolve path: %w", err)
	}

	info, err := os.Stat(resolvedPath)
	if err != nil {
		return "", "", fmt.Errorf("stat path: %w", err)
	}

	searchRoots := []string{}
	if info.IsDir() {
		searchRoots = append(searchRoots, resolvedPath)
	} else {
		searchRoots = append(searchRoots, filepath.Dir(resolvedPath))
	}
	searchRoots = append(searchRoots, filepath.Join(searchRoots[0], "www"))

	seen := map[string]bool{}
	for _, root := range searchRoots {
		root = filepath.Clean(root)
		if seen[root] {
			continue
		}
		seen[root] = true

		engine, ok := detectRPGMakerEngine(root)
		if ok {
			return root, engine, nil
		}
	}

	return "", "", fmt.Errorf("could not find an RPG Maker MV/MZ project under %s", resolvedPath)
}

func (h *RPGMakerHook) InspectHook(inputPath string) (TextHookStatus, error) {
	projectRoot, engine, err := resolveRPGMakerProjectRoot(inputPath)
	if err != nil {
		return TextHookStatus{Message: err.Error()}, nil
	}

	compatibility, err := inspectRPGMakerTextHookCompatibility(projectRoot)
	if err != nil {
		return TextHookStatus{}, err
	}

	pluginPath := filepath.Join(projectRoot, "js", "plugins", rpgMakerClipboardPlugin+".js")
	pluginsConfigPath := filepath.Join(projectRoot, "js", "plugins.js")
	configs, err := readRPGMakerPluginConfigs(pluginsConfigPath)
	if err != nil {
		return TextHookStatus{}, err
	}

	enabled := false
	for _, cfg := range configs {
		if strings.TrimSpace(cfg.Name) == rpgMakerClipboardPlugin && cfg.Status {
			enabled = true
			break
		}
	}

	installed := enabled && util.IsExistingFile(pluginPath)
	message := "Text hook plugin is not installed."
	if installed {
		message = "Text hook plugin is installed and enabled."
	}

	return TextHookStatus{
		Supported:         true,
		Installed:         installed,
		Engine:            engine,
		ProjectRoot:       projectRoot,
		PluginPath:        pluginPath,
		PluginsConfigPath: pluginsConfigPath,
		Compatibility:     compatibility,
		Message:           message,
	}, nil
}

func detectRPGMakerEngine(projectRoot string) (string, bool) {
	jsDir := filepath.Join(projectRoot, "js")
	pluginsDir := filepath.Join(jsDir, "plugins")
	pluginsConfigPath := filepath.Join(jsDir, "plugins.js")

	if !util.IsExistingDir(pluginsDir) || !util.IsExistingFile(pluginsConfigPath) {
		return "", false
	}

	switch {
	case util.IsExistingFile(filepath.Join(jsDir, "rmmz_core.js")):
		return "RPG Maker MZ", true
	case util.IsExistingFile(filepath.Join(jsDir, "rpg_core.js")):
		return "RPG Maker MV", true
	default:
		return "RPG Maker MV/MZ", true
	}
}

func inspectRPGMakerTextHookCompatibility(projectRoot string) (TextHookCompatibilityReport, error) {
	pluginsConfigPath := filepath.Join(projectRoot, "js", "plugins.js")
	configs, err := readRPGMakerPluginConfigs(pluginsConfigPath)
	if err != nil {
		return TextHookCompatibilityReport{}, err
	}

	report := TextHookCompatibilityReport{
		ProjectRoot: filepath.Clean(projectRoot),
		RiskLevel:   "safe",
	}

	for _, cfg := range configs {
		if !cfg.Status || strings.TrimSpace(cfg.Name) == "" || cfg.Name == rpgMakerClipboardPlugin {
			continue
		}
		report.EnabledPlugins = append(report.EnabledPlugins, cfg.Name)
	}
	sort.Slice(report.EnabledPlugins, func(i, j int) bool {
		return strings.ToLower(report.EnabledPlugins[i]) < strings.ToLower(report.EnabledPlugins[j])
	})

	findings := map[string]bool{}
	addFinding := func(riskLevel, finding string) {
		if strings.TrimSpace(finding) == "" || findings[finding] {
			return
		}
		findings[finding] = true
		report.Findings = append(report.Findings, finding)
		if report.RiskLevel == "safe" && riskLevel == "warn" {
			report.RiskLevel = "warn"
		}
	}

	for _, pluginName := range report.EnabledPlugins {
		lowerName := strings.ToLower(pluginName)
		if looksLikeCustomMessagePlugin(lowerName) {
			addFinding("warn", fmt.Sprintf("plugin %q looks message-related and may replace the stock dialogue flow", pluginName))
		}

		pluginPath := filepath.Join(projectRoot, "js", "plugins", pluginName+".js")
		data, err := os.ReadFile(pluginPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				addFinding("warn", fmt.Sprintf("enabled plugin %q is listed in plugins.js but its file was not found", pluginName))
				continue
			}
			return TextHookCompatibilityReport{}, fmt.Errorf("read plugin %s: %w", pluginPath, err)
		}

		if signature := detectMessageHookSignature(string(data)); signature != "" {
			addFinding("warn", fmt.Sprintf("plugin %q appears to override %s", pluginName, signature))
		}
	}

	if len(report.Findings) == 0 {
		report.Findings = append(report.Findings, "no obvious custom message-hook conflicts were detected in enabled plugin files")
	}

	return report, nil
}

func readRPGMakerPluginConfigs(pluginsConfigPath string) ([]rpgMakerPluginConfig, error) {
	data, err := os.ReadFile(pluginsConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read plugins.js: %w", err)
	}

	content := string(data)
	firstBracket := strings.Index(content, "[")
	lastBracket := strings.LastIndex(content, "]")
	if firstBracket == -1 || lastBracket == -1 || lastBracket < firstBracket {
		return nil, fmt.Errorf("plugins.js did not contain a recognizable plugin array: %s", pluginsConfigPath)
	}

	var configs []rpgMakerPluginConfig
	if err := json.Unmarshal([]byte(content[firstBracket:lastBracket+1]), &configs); err != nil {
		return nil, fmt.Errorf("decode plugins.js: %w", err)
	}
	return configs, nil
}

func ensureRPGMakerPluginEnabled(pluginsConfigPath, pluginName string) error {
	data, err := os.ReadFile(pluginsConfigPath)
	if err != nil {
		return fmt.Errorf("read plugins.js: %w", err)
	}

	content := string(data)
	if strings.Contains(content, "\"name\":\""+pluginName+"\"") || strings.Contains(content, "\"name\": \""+pluginName+"\"") {
		return nil
	}

	lastBracket := strings.LastIndex(content, "]")
	firstBracket := strings.Index(content, "[")
	if firstBracket == -1 || lastBracket == -1 || lastBracket < firstBracket {
		return fmt.Errorf("plugins.js did not contain a recognizable plugin array: %s", pluginsConfigPath)
	}

	inner := strings.TrimSpace(content[firstBracket+1 : lastBracket])
	entry := fmt.Sprintf(`{"name":"%s","status":true,"description":"Capture the latest dialogue and copy it with Ctrl+C.","parameters":{}}`, pluginName)

	var updated string
	if inner == "" {
		updated = content[:firstBracket+1] + "\n" + entry + "\n" + content[lastBracket:]
	} else {
		updated = content[:lastBracket] + ",\n" + entry + "\n" + content[lastBracket:]
	}

	if err := os.WriteFile(pluginsConfigPath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write plugins.js: %w", err)
	}
	return nil
}

func looksLikeCustomMessagePlugin(pluginName string) bool {
	indicators := []string{
		"message",
		"msg",
		"text",
		"dialog",
		"novel",
		"namebox",
		"balloon",
		"speech",
		"adv",
	}
	for _, indicator := range indicators {
		if strings.Contains(pluginName, indicator) {
			return true
		}
	}
	return false
}

func detectMessageHookSignature(content string) string {
	lowerContent := strings.ToLower(content)
	signatures := []struct {
		pattern string
		label   string
	}{
		{"window_message.prototype.startmessage", "Window_Message.startMessage"},
		{"window_message.prototype.updatemessage", "Window_Message.updateMessage"},
		{"window_message.prototype.terminatemessage", "Window_Message.terminateMessage"},
		{"game_message.prototype.add", "Game_Message.add"},
		{"game_message.prototype.setspeakername", "Game_Message.setSpeakerName"},
		{"game_message.prototype.alltext", "Game_Message.allText"},
		{"scene_message.prototype", "Scene_Message methods"},
	}
	for _, signature := range signatures {
		if strings.Contains(lowerContent, signature.pattern) {
			return signature.label
		}
	}
	return ""
}

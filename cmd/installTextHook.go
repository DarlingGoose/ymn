package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/DarlingGoose/wgl/pkg/game/gameconfig"
	"github.com/DarlingGoose/wgl/pkg/game/launcher"
	"github.com/DarlingGoose/wgl/pkg/util"
	"github.com/spf13/cobra"
)

const (
	textHookMethodAuto       = "auto"
	textHookMethodMod        = "mod"
	textHookMethodTextractor = "textractor"
	rpgMakerClipboardPlugin  = "WGLClipboardText"
)

var installTextHookMethod string
var installTextHookGame string

var installTextHookCmd = &cobra.Command{
	Use:   "install-text-hook [path-to-game-dir-or-exe]",
	Short: "Install text extraction support for a game",
	Long: strings.Join([]string{
		"Install one of the supported text extraction options for a game.",
		"",
		"Option 1 (recommended for RPG Maker MV/MZ): mod the game directly.",
		"wgl writes a plugin into the game, enables it in plugins.js, and lets you press Ctrl+C in-game to copy the latest dialogue to the clipboard.",
		"",
		"Option 2: Textractor.",
		"That path is reserved for non-MV/MZ games, but this CLI does not automate Textractor installation yet.",
		"",
		"When no path is provided, wgl opens a picker for previously added games.",
	}, "\n"),
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		method := strings.ToLower(strings.TrimSpace(installTextHookMethod))
		if method == "" {
			method = textHookMethodAuto
		}

		inputPath, err := resolveInstallTextHookTarget(args)
		if err != nil {
			return err
		}

		switch method {
		case textHookMethodAuto, textHookMethodMod:
			result, err := installRPGMakerClipboardHook(inputPath)
			if err != nil {
				if method == textHookMethodAuto {
					return fmt.Errorf("auto mode currently supports RPG Maker MV/MZ direct mods only: %w", err)
				}
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "installed plugin: %s\n", result.PluginPath)
			fmt.Fprintf(cmd.OutOrStdout(), "updated plugins.js: %s\n", result.PluginsConfigPath)
			fmt.Fprintf(cmd.OutOrStdout(), "engine: %s\n", result.Engine)
			printTextHookCompatibilityReport(cmd, result.Compatibility)
			fmt.Fprintln(cmd.OutOrStdout(), "use Ctrl+C in-game to copy the latest visible dialogue to the clipboard")
			return nil
		case textHookMethodTextractor:
			return errors.New("Textractor installation is not automated yet; use --method mod for RPG Maker MV/MZ games")
		default:
			return fmt.Errorf("unsupported method %q; expected auto, mod, or textractor", installTextHookMethod)
		}
	},
}

type textHookInstallResult struct {
	Engine            string
	PluginPath        string
	PluginsConfigPath string
	Compatibility     textHookCompatibilityReport
}

type textHookCompatibilityReport struct {
	ProjectRoot    string
	RiskLevel      string
	EnabledPlugins []string
	Findings       []string
}

type textHookStatus struct {
	Supported         bool
	Installed         bool
	Engine            string
	ProjectRoot       string
	PluginPath        string
	PluginsConfigPath string
	Compatibility     textHookCompatibilityReport
	Message           string
}

type rpgMakerPluginConfig struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
}

func init() {
	rootCmd.AddCommand(installTextHookCmd)
	installTextHookCmd.Flags().StringVar(&installTextHookMethod, "method", textHookMethodAuto, "text hook method: auto, mod, or textractor")
	installTextHookCmd.Flags().StringVarP(&installTextHookGame, "game", "g", "", "name of a previously added game to install the hook into")
}

func resolveInstallTextHookTarget(args []string) (string, error) {
	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		return args[0], nil
	}

	selectedName := strings.TrimSpace(installTextHookGame)
	if selectedName != "" {
		cfg, err := gameconfig.FindConfig(selectedName)
		if err != nil {
			return "", err
		}
		return util.FirstNonEmpty(cfg.GamePath, cfg.Executable), nil
	}
	picker, err := launcher.NewPicker("Select a game to install a text hook", "install the text hook")
	if err != nil {
		return "", err
	}
	cfg, err := picker.SelectGameConfig()
	if err != nil {
		return "", err
	}
	return util.FirstNonEmpty(cfg.GamePath, cfg.Executable), nil
}

func printTextHookCompatibilityReport(cmd *cobra.Command, report textHookCompatibilityReport) {
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "project root: %s\n", report.ProjectRoot)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "text hook compatibility: %s\n", report.RiskLevel)
	if len(report.EnabledPlugins) > 0 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "enabled plugins (%d): %s\n", len(report.EnabledPlugins), strings.Join(report.EnabledPlugins, ", "))
	}
	for _, finding := range report.Findings {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "compatibility note: %s\n", finding)
	}
}

func installRPGMakerClipboardHook(inputPath string) (textHookInstallResult, error) {
	projectRoot, engine, err := resolveRPGMakerProjectRoot(inputPath)
	if err != nil {
		return textHookInstallResult{}, err
	}

	compatibility, err := inspectRPGMakerTextHookCompatibility(projectRoot)
	if err != nil {
		return textHookInstallResult{}, err
	}

	jsDir := filepath.Join(projectRoot, "js")
	pluginsDir := filepath.Join(jsDir, "plugins")
	pluginsConfigPath := filepath.Join(jsDir, "plugins.js")
	pluginPath := filepath.Join(pluginsDir, rpgMakerClipboardPlugin+".js")

	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return textHookInstallResult{}, fmt.Errorf("create plugins directory: %w", err)
	}

	if err := os.WriteFile(pluginPath, []byte(rpgMakerClipboardPluginSource), 0o644); err != nil {
		return textHookInstallResult{}, fmt.Errorf("write plugin: %w", err)
	}

	if err := ensureRPGMakerPluginEnabled(pluginsConfigPath, rpgMakerClipboardPlugin); err != nil {
		return textHookInstallResult{}, err
	}

	return textHookInstallResult{
		Engine:            engine,
		PluginPath:        pluginPath,
		PluginsConfigPath: pluginsConfigPath,
		Compatibility:     compatibility,
	}, nil
}

func inspectRPGMakerClipboardHook(inputPath string) (textHookStatus, error) {
	projectRoot, engine, err := resolveRPGMakerProjectRoot(inputPath)
	if err != nil {
		return textHookStatus{Message: err.Error()}, nil
	}

	compatibility, err := inspectRPGMakerTextHookCompatibility(projectRoot)
	if err != nil {
		return textHookStatus{}, err
	}

	pluginPath := filepath.Join(projectRoot, "js", "plugins", rpgMakerClipboardPlugin+".js")
	pluginsConfigPath := filepath.Join(projectRoot, "js", "plugins.js")
	configs, err := readRPGMakerPluginConfigs(pluginsConfigPath)
	if err != nil {
		return textHookStatus{}, err
	}

	enabled := false
	for _, cfg := range configs {
		if strings.TrimSpace(cfg.Name) == rpgMakerClipboardPlugin && cfg.Status {
			enabled = true
			break
		}
	}

	installed := enabled && isExistingFile(pluginPath)
	message := "Text hook plugin is not installed."
	if installed {
		message = "Text hook plugin is installed and enabled."
	}

	return textHookStatus{
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

func detectRPGMakerEngine(projectRoot string) (string, bool) {
	jsDir := filepath.Join(projectRoot, "js")
	pluginsDir := filepath.Join(jsDir, "plugins")
	pluginsConfigPath := filepath.Join(jsDir, "plugins.js")

	if !isExistingDir(pluginsDir) || !isExistingFile(pluginsConfigPath) {
		return "", false
	}

	switch {
	case isExistingFile(filepath.Join(jsDir, "rmmz_core.js")):
		return "RPG Maker MZ", true
	case isExistingFile(filepath.Join(jsDir, "rpg_core.js")):
		return "RPG Maker MV", true
	default:
		return "RPG Maker MV/MZ", true
	}
}

func inspectRPGMakerTextHookCompatibility(projectRoot string) (textHookCompatibilityReport, error) {
	pluginsConfigPath := filepath.Join(projectRoot, "js", "plugins.js")
	configs, err := readRPGMakerPluginConfigs(pluginsConfigPath)
	if err != nil {
		return textHookCompatibilityReport{}, err
	}

	report := textHookCompatibilityReport{
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
			return textHookCompatibilityReport{}, fmt.Errorf("read plugin %s: %w", pluginPath, err)
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

func isExistingDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isExistingFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

const rpgMakerClipboardPluginSource = `(function() {
  "use strict";

  const COPY_KEY = "c";
  const LOG_FILENAME = "wgl-dialogue.log";

  window.__wglLastMessage = "";
  window.__wglLastChoices = [];
  window.__wglLastChoiceText = "";
  window.__wglTranscriptPath = "";
  let lastLoggedMessage = "";
  let lastLoggedChoicesKey = "";

  function setLatestMessage(text) {
    if (typeof text !== "string") {
      return;
    }
    window.__wglLastMessage = text;
  }

  function latestMessage() {
    return String(window.__wglLastMessage || "").trim();
  }

  function setLatestChoices(choices) {
    if (!Array.isArray(choices)) {
      window.__wglLastChoices = [];
      window.__wglLastChoiceText = "";
      return;
    }
    const normalized = choices.map(function(choice) {
      return String(choice || "").trim();
    }).filter(function(choice) {
      return choice.length > 0;
    });
    window.__wglLastChoices = normalized;
    window.__wglLastChoiceText = normalized.join("\n");
  }

  function resolveTranscriptPath() {
    try {
      if (typeof require !== "function") {
        return "";
      }

      const path = require("path");
      const roots = [];

      if (typeof __dirname === "string" && __dirname) {
        roots.push(path.resolve(__dirname, "..", ".."));
      }
      if (typeof process !== "undefined" && process && typeof process.cwd === "function") {
        roots.push(process.cwd());
      }
      if (typeof process !== "undefined" && process && typeof process.execPath === "string" && process.execPath) {
        roots.push(path.dirname(process.execPath));
      }
      if (typeof nw !== "undefined" && nw.App && typeof nw.App.startPath === "string" && nw.App.startPath) {
        roots.push(nw.App.startPath);
      }

      const seen = {};
      for (let i = 0; i < roots.length; i += 1) {
        const root = String(roots[i] || "").trim();
        if (!root) {
          continue;
        }
        const normalized = path.resolve(root);
        if (seen[normalized]) {
          continue;
        }
        seen[normalized] = true;
        return path.join(normalized, LOG_FILENAME);
      }
    } catch (error) {
      console.warn("WGLClipboardText transcript path lookup failed", error);
    }
    return "";
  }

  function currentSpeakerName() {
    try {
      if ($gameMessage && typeof $gameMessage.speakerName === "function") {
        return String($gameMessage.speakerName() || "").trim();
      }
      if ($gameMessage && typeof $gameMessage._speakerName !== "undefined") {
        return String($gameMessage._speakerName || "").trim();
      }
    } catch (error) {
      console.warn("WGLClipboardText speaker lookup failed", error);
    }
    return "";
  }

  function rawMessageText() {
    try {
      if ($gameMessage && typeof $gameMessage.allText === "function") {
        return String($gameMessage.allText() || "");
      }
      if ($gameMessage && Array.isArray($gameMessage._texts)) {
        return $gameMessage._texts.join("\n");
      }
    } catch (error) {
      console.warn("WGLClipboardText raw message lookup failed", error);
    }
    return "";
  }

  function rawChoices() {
    try {
      if ($gameMessage && typeof $gameMessage.choices === "function") {
        return $gameMessage.choices();
      }
      if ($gameMessage && Array.isArray($gameMessage._choices)) {
        return $gameMessage._choices;
      }
    } catch (error) {
      console.warn("WGLClipboardText raw choice lookup failed", error);
    }
    return [];
  }

  function normalizeMessageText(text, messageWindow) {
    let normalized = String(text || "");
    if (normalized && messageWindow && typeof messageWindow.convertEscapeCharacters === "function") {
      normalized = messageWindow.convertEscapeCharacters(normalized);
    }
    return normalized.replace(/\r\n/g, "\n").trim();
  }

  function normalizeChoiceText(text, messageWindow) {
    return normalizeMessageText(text, messageWindow);
  }

  function refreshLatestMessage(messageWindow) {
    const text = normalizeMessageText(rawMessageText(), messageWindow);
    setLatestMessage(text);
    appendToTranscript(text);
  }

  function refreshLatestChoices(messageWindow) {
    const choices = rawChoices().map(function(choice) {
      return normalizeChoiceText(choice, messageWindow);
    }).filter(function(choice) {
      return choice.length > 0;
    });
    setLatestChoices(choices);
    appendChoicesToTranscript(choices);
  }

  function currentMessageWindow() {
    if (typeof SceneManager === "undefined" || !SceneManager || !SceneManager._scene) {
      return null;
    }
    return SceneManager._scene._messageWindow || null;
  }

  function appendToTranscript(text) {
    const message = String(text || "").trim();
    if (!message || message === lastLoggedMessage) {
      return;
    }

    try {
      if (typeof require !== "function") {
        return;
      }

      const fs = require("fs");
      const logPath = resolveTranscriptPath();
      if (!logPath) {
        return;
      }

      const speaker = currentSpeakerName();
      const timestamp = new Date().toISOString();
      const header = speaker ? "[" + timestamp + "] " + speaker + "\n" : "[" + timestamp + "]\n";
      const entry = header + message + "\n\n";
      fs.appendFileSync(logPath, entry, "utf8");
      window.__wglTranscriptPath = logPath;
      lastLoggedMessage = message;
    } catch (error) {
      console.warn("WGLClipboardText transcript write failed", error);
    }
  }

  function appendChoicesToTranscript(choices) {
    if (!Array.isArray(choices) || choices.length === 0) {
      lastLoggedChoicesKey = "";
      return;
    }

    const normalized = choices.map(function(choice) {
      return String(choice || "").trim();
    }).filter(function(choice) {
      return choice.length > 0;
    });
    if (normalized.length === 0) {
      lastLoggedChoicesKey = "";
      return;
    }

    const choiceKey = normalized.join("\n");
    if (choiceKey === lastLoggedChoicesKey) {
      return;
    }

    try {
      if (typeof require !== "function") {
        return;
      }

      const fs = require("fs");
      const logPath = resolveTranscriptPath();
      if (!logPath) {
        return;
      }

      const timestamp = new Date().toISOString();
      const lines = normalized.map(function(choice, index) {
        return (index + 1) + ". " + choice;
      });
      const entry = "[" + timestamp + "] choices\n" + lines.join("\n") + "\n\n";
      fs.appendFileSync(logPath, entry, "utf8");
      window.__wglTranscriptPath = logPath;
      lastLoggedChoicesKey = choiceKey;
    } catch (error) {
      console.warn("WGLClipboardText choice transcript write failed", error);
    }
  }

  function copyToClipboard(text) {
    if (!text) {
      return Promise.resolve(false);
    }

    try {
      if (typeof nw !== "undefined" && nw.Clipboard && typeof nw.Clipboard.get === "function") {
        nw.Clipboard.get().set(text, "text");
        return Promise.resolve(true);
      }
    } catch (error) {
      console.warn("WGLClipboardText nw.js clipboard copy failed", error);
    }

    if (typeof navigator !== "undefined" && navigator.clipboard && typeof navigator.clipboard.writeText === "function") {
      return navigator.clipboard.writeText(text).then(function() {
        return true;
      }).catch(function(error) {
        console.warn("WGLClipboardText navigator clipboard copy failed", error);
        return false;
      });
    }

    return Promise.resolve(false);
  }

  const originalGameMessageClear = Game_Message.prototype.clear;
  Game_Message.prototype.clear = function() {
    originalGameMessageClear.call(this);
    setLatestMessage("");
    setLatestChoices([]);
  };

  const originalGameMessageAdd = Game_Message.prototype.add;
  Game_Message.prototype.add = function(text) {
    originalGameMessageAdd.call(this, text);
    refreshLatestMessage(currentMessageWindow());
  };

  if (typeof Game_Message.prototype.setSpeakerName === "function") {
    const originalGameMessageSetSpeakerName = Game_Message.prototype.setSpeakerName;
    Game_Message.prototype.setSpeakerName = function(speakerName) {
      originalGameMessageSetSpeakerName.call(this, speakerName);
      refreshLatestMessage(currentMessageWindow());
    };
  }

  if (typeof Game_Message.prototype.setChoices === "function") {
    const originalGameMessageSetChoices = Game_Message.prototype.setChoices;
    Game_Message.prototype.setChoices = function(choices, defaultType, cancelType) {
      originalGameMessageSetChoices.call(this, choices, defaultType, cancelType);
      refreshLatestChoices(currentMessageWindow());
    };
  }

  const originalWindowMessageStartMessage = Window_Message.prototype.startMessage;
  Window_Message.prototype.startMessage = function() {
    originalWindowMessageStartMessage.call(this);
    refreshLatestMessage(this);
    refreshLatestChoices(this);
  };

  if (typeof Window_ChoiceList !== "undefined" && Window_ChoiceList.prototype) {
    if (typeof Window_ChoiceList.prototype.start === "function") {
      const originalWindowChoiceListStart = Window_ChoiceList.prototype.start;
      Window_ChoiceList.prototype.start = function() {
        originalWindowChoiceListStart.call(this);
        refreshLatestChoices(currentMessageWindow() || this);
      };
    }

    if (typeof Window_ChoiceList.prototype.refresh === "function") {
      const originalWindowChoiceListRefresh = Window_ChoiceList.prototype.refresh;
      Window_ChoiceList.prototype.refresh = function() {
        originalWindowChoiceListRefresh.call(this);
        refreshLatestChoices(currentMessageWindow() || this);
      };
    }
  }

  document.addEventListener("keydown", function(event) {
    if (!event.ctrlKey || String(event.key || "").toLowerCase() !== COPY_KEY) {
      return;
    }

    const text = latestMessage();
    if (!text) {
      return;
    }

    copyToClipboard(text).then(function(copied) {
      if (copied) {
        console.log("WGLClipboardText copied:", text);
      }
    });
  });
})();`

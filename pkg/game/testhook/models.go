package testhook

type TextHookInstallResult struct {
	Engine            string                      `json:"engine"`
	Method            string                      `json:"method"` // mod, script_patch, external_hook
	PluginPath        string                      `json:"plugin_path,omitempty"`
	PluginsConfigPath string                      `json:"plugins_config_path,omitempty"`
	OutputPath        string                      `json:"output_path,omitempty"`
	Compatibility     TextHookCompatibilityReport `json:"compatibility"`
}

type TextHookCompatibilityReport struct {
	ProjectRoot    string   `json:"project_root"`
	RiskLevel      string   `json:"risk_level"` // safe, warn, risky, unsupported
	EnabledPlugins []string `json:"enabled_plugins,omitempty"`
	Findings       []string `json:"findings,omitempty"`
}

type TextHookStatus struct {
	Supported bool `json:"supported"`
	Installed bool `json:"installed"`

	Loaded            bool
	Engine            string                      `json:"engine"`
	Method            string                      `json:"method,omitempty"`
	ProjectRoot       string                      `json:"project_root"`
	PluginPath        string                      `json:"plugin_path,omitempty"`
	PluginsConfigPath string                      `json:"plugins_config_path,omitempty"`
	OutputPath        string                      `json:"output_path,omitempty"`
	Compatibility     TextHookCompatibilityReport `json:"compatibility"`
	Message           string                      `json:"message"`
}

type rpgMakerPluginConfig struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
}

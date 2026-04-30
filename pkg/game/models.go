package game

type TextHookInstallResult struct {
	Engine            string
	PluginPath        string
	PluginsConfigPath string
	Compatibility     TextHookCompatibilityReport
}

type TextHookCompatibilityReport struct {
	ProjectRoot    string
	RiskLevel      string
	EnabledPlugins []string
	Findings       []string
}

type TextHookStatus struct {
	Supported         bool
	Installed         bool
	Engine            string
	ProjectRoot       string
	PluginPath        string
	PluginsConfigPath string
	Compatibility     TextHookCompatibilityReport
	Message           string
}

type rpgMakerPluginConfig struct {
	Name   string `json:"name"`
	Status bool   `json:"status"`
}

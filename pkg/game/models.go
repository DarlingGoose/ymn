package game

type TextHookInstallResult struct {
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

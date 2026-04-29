package gameconfig

import "time"

type RunnerType string

const (
	RunnerWine   RunnerType = "wine"
	RunnerProton RunnerType = "proton"
	RunnerSteam  RunnerType = "steam"
)

type RuntimeStatus struct {
	WinePath           string   `json:"wine_path,omitempty"`
	WineBootPath       string   `json:"wineboot_path,omitempty"`
	SteamPath          string   `json:"steam_path,omitempty"`
	SteamRoot          string   `json:"steam_root,omitempty"`
	AvailableProton    []string `json:"available_proton,omitempty"`
	SelectedProtonPath string   `json:"selected_proton_path,omitempty"`
}

type VerificationAttempt struct {
	Runner    RunnerType `json:"runner"`
	Strategy  string     `json:"strategy"`
	Success   bool       `json:"success"`
	Message   string     `json:"message"`
	LogPath   string     `json:"log_path,omitempty"`
	CheckedAt time.Time  `json:"checked_at"`
}

type VerificationStatus struct {
	Verified   bool                  `json:"verified"`
	VerifiedAt time.Time             `json:"verified_at,omitempty"`
	Attempts   []VerificationAttempt `json:"attempts,omitempty"`
}

type GameConfig struct {
	Name          string             `json:"name"`
	GamePath      string             `json:"game_path"`
	Executable    string             `json:"executable"`
	WorkingDir    string             `json:"working_dir"`
	IconPath      string             `json:"icon_path,omitempty"`
	ImagePath     string             `json:"image_path,omitempty"`
	Runner        RunnerType         `json:"runner"`
	RunnerPath    string             `json:"runner_path"`
	PrefixPath    string             `json:"prefix_path,omitempty"`
	RequiresSteam bool               `json:"requires_steam"`
	SteamAppID    string             `json:"steam_app_id,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	RuntimeInfo   RuntimeStatus      `json:"runtime_info"`
	Verification  VerificationStatus `json:"verification"`
}

package gamepage

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gioui.org/widget/material"
	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/autorunner"
	"github.com/DarlingGoose/gr/gamescope"
	"github.com/DarlingGoose/gr/monitors"
	"github.com/DarlingGoose/gr/wine"
	vngame "github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/input"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/components/toggles"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
)

func runnerOptionLabel(name string) string {
	switch name {
	case "Width":
		return "Width (px)"
	case "Height":
		return "Height (px)"
	case "OutputWidth":
		return "Output width (px)"
	case "OutputHeight":
		return "Output height (px)"
	case "RefreshRate":
		return "Refresh rate (Hz)"
	case "UseWine":
		return "Use Wine"
	case "WineStartWait":
		return "Wait for Wine startup"
	case "KillWineOnExit":
		return "Kill Wine on exit"
	case "ForceGrab":
		return "Force grab"
	case "SteamDeckMode":
		return "Steam Deck mode"
	case "ExposeWayland":
		return "Expose Wayland"
	case "DefaultWinePrefix", "DefaultPrefix":
		return "Wine prefix"
	case "GamescopeBin":
		return "Gamescope binary"
	case "WineBin":
		return "Wine binary"
	case "WineServerBin":
		return "Wineserver binary"
	case "WineTricksBin":
		return "Winetricks binary"
	case "ExtraArgs":
		return "Extra arguments"
	default:
		return splitCamel(name)
	}
}

func runnerOptionPlaceholder(name string) string {
	switch name {
	case "Width":
		return "1920"
	case "Height":
		return "1080"
	case "RefreshRate":
		return "0"
	case "OutputWidth":
		return "3840"
	case "OutputHeight":
		return "2160"
	default:
		return "0"
	}
}

func runnerOptionDescription(name string) string {
	switch name {
	case "Name":
		return "Internal runner profile name."
	case "ExtraArgs":
		return "Additional runner arguments, separated by commas."
	case "DefaultPrefix", "DefaultWinePrefix":
		return "Used when the game does not provide a more specific prefix."
	case "UseWine":
		return "Launch the game through Wine inside the runner."
	case "WineStartWait":
		return "Wait briefly after Wine starts before continuing."
	case "KillWineOnExit":
		return "Stop Wine processes when the game exits."
	case "Fullscreen":
		return "Launch through the runner in fullscreen mode."
	case "Borderless":
		return "Use a borderless window when not fullscreen."
	case "ForceGrab":
		return "Keep pointer and keyboard input captured by the game."
	case "SteamDeckMode":
		return "Enable runner behavior intended for Steam Deck style sessions."
	case "ExposeWayland":
		return "Expose Wayland to the game when supported by the session."
	case "Scaler":
		return "Choose how the game image is scaled to the output resolution."
	case "Filter":
		return "Choose the upscaling filter used by Gamescope."
	case "Width", "Height", "OutputWidth", "OutputHeight":
		return "Pixel value passed to the runner. Use 0 to let the runner choose."
	case "RefreshRate":
		return "Target refresh rate. Use 0 for the monitor default."
	case "GamescopeBin":
		return "Binary override. Leave empty to use gamescope from PATH."
	case "WineBin":
		return "Binary override. Leave empty to use wine from PATH."
	case "WineServerBin":
		return "Binary override. Leave empty to use wineserver from PATH."
	case "WineTricksBin":
		return "Binary override. Leave empty to use winetricks from PATH."
	default:
		return ""
	}
}

func applyRunnerOptionFields(fields []*runnerOptionField, v reflect.Value) {
	for _, f := range fields {
		field := v.FieldByName(f.name)
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		switch {
		case f.toggle != nil && field.Kind() == reflect.Bool:
			field.SetBool(f.toggle.Checked)
		case f.input != nil && field.Kind() == reflect.String:
			field.SetString(strings.TrimSpace(f.input.Text()))
		case f.dropdown != nil && field.Kind() == reflect.String:
			if item, ok := f.dropdown.SelectedItem(); ok {
				field.SetString(strings.TrimSpace(item.Value))
			}
		case f.input != nil && field.Kind() >= reflect.Int && field.Kind() <= reflect.Int64:
			text := strings.TrimSpace(f.input.Text())
			if text == "" {
				field.SetInt(0)
				continue
			}
			if n, err := strconv.ParseInt(text, 10, 64); err == nil {
				field.SetInt(n)
			}
		case f.input != nil && field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String:
			field.Set(reflect.ValueOf(splitCommaList(f.input.Text())))
		}
	}
}

func setRunnerOptionFields(fields []*runnerOptionField, v reflect.Value) {
	for _, f := range fields {
		field := v.FieldByName(f.name)
		if !field.IsValid() {
			continue
		}
		switch {
		case f.toggle != nil && field.Kind() == reflect.Bool:
			f.toggle.JumpTo(field.Bool())
		case f.input != nil && field.Kind() == reflect.String:
			f.input.SetText(field.String())
		case f.dropdown != nil && field.Kind() == reflect.String:
			if !f.dropdown.SelectItem(field.String()) && len(f.dropdown.Items) > 0 {
				f.dropdown.Selected = 0
			}
		case f.input != nil && field.Kind() >= reflect.Int && field.Kind() <= reflect.Int64:
			f.input.SetText(strconv.FormatInt(field.Int(), 10))
		case f.input != nil && field.Kind() == reflect.Slice:
			parts := make([]string, 0, field.Len())
			for i := 0; i < field.Len(); i++ {
				parts = append(parts, fmt.Sprint(field.Index(i).Interface()))
			}
			f.input.SetText(strings.Join(parts, ", "))
		}
	}
}

func buildRunnerOptionFields(sample any, th *material.Theme, tc *theme.Client, onChange func()) []*runnerOptionField {
	t := reflect.TypeOf(sample)
	fields := make([]*runnerOptionField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		f := &runnerOptionField{
			name:        sf.Name,
			label:       runnerOptionLabel(sf.Name),
			description: runnerOptionDescription(sf.Name),
			kind:        sf.Type.Kind(),
			isSlice:     sf.Type.Kind() == reflect.Slice,
			onChange:    onChange,
		}
		switch {
		case f.kind == reflect.Bool:
			f.toggle = toggles.NewToggle(f.label, false).WithThemeClient(tc)
		case sf.Name == "Scaler":
			f.dropdown = dropdowns.NewDropdown([]dropdowns.DropdownItem{
				{Label: "Auto", Value: "auto"},
				{Label: "Integer", Value: "integer"},
				{Label: "Fit", Value: "fit"},
				{Label: "Fill", Value: "fill"},
				{Label: "Stretch", Value: "stretch"},
			}).WithThemeClient(tc).WithRole(theme.TextRoleLabel)
			f.dropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
				if valid && onChange != nil {
					onChange()
				}
			})
		case sf.Name == "Filter":
			f.dropdown = dropdowns.NewDropdown([]dropdowns.DropdownItem{
				{Label: "Linear", Value: "linear"},
				{Label: "Nearest", Value: "nearest"},
				{Label: "FSR", Value: "fsr"},
				{Label: "NIS", Value: "nis"},
				{Label: "Pixel", Value: "pixel"},
			}).WithThemeClient(tc).WithRole(theme.TextRoleLabel)
			f.dropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
				if valid && onChange != nil {
					onChange()
				}
			})
		default:
			f.input = input.NewTextInput(f.label, "").WithMaterialTheme(th).WithThemeClient(tc)
			if f.kind >= reflect.Int && f.kind <= reflect.Int64 {
				f.input.Kind = input.KindInteger
				f.input.LeadingIcon = "lucide:hash"
				f.input.Rules = append(f.input.Rules, nonNegativeIntegerRule(sf.Name))
				f.input.Hint = runnerOptionPlaceholder(sf.Name)
			}
			if f.isSlice {
				f.input.Hint = "Comma-separated values"
			}
			f.input.OnChange = func(string) {
				if onChange != nil {
					onChange()
				}
			}
		}
		fields = append(fields, f)
	}
	return fields
}

func hasWineOptions(cfg *wine.Options) bool {
	return cfg != nil && (cfg.Name != "" || cfg.WineBin != "" || cfg.WineTricksBin != "" || cfg.DefaultPrefix != "")
}

func hasGamescopeOptions(cfg *gamescope.Options) bool {
	return cfg != nil && (cfg.Name != "" ||
		cfg.GamescopeBin != "" ||
		cfg.WineBin != "" ||
		cfg.WineServerBin != "" ||
		cfg.DefaultWinePrefix != "" ||
		cfg.UseWine ||
		cfg.WineStartWait ||
		cfg.KillWineOnExit ||
		cfg.Width != 0 ||
		cfg.Height != 0 ||
		cfg.RefreshRate != 0 ||
		cfg.OutputWidth != 0 ||
		cfg.OutputHeight != 0 ||
		cfg.Fullscreen ||
		cfg.Borderless ||
		cfg.ForceGrab ||
		cfg.SteamDeckMode ||
		cfg.ExposeWayland ||
		cfg.Scaler != "" ||
		cfg.Filter != "" ||
		len(cfg.ExtraArgs) > 0)
}

func hasGRConfig(cfg gr.Config) bool {
	return cfg.Background ||
		cfg.WorkingDir != "" ||
		len(cfg.Args) > 0 ||
		len(cfg.Envs) > 0 ||
		cfg.SystemArch != "" ||
		cfg.WinePrefix != "" ||
		len(cfg.Dependencies) > 0 ||
		cfg.Name != "" ||
		cfg.PID != 0 ||
		cfg.Session != "" ||
		cfg.SessionID != "" ||
		cfg.LogFile != ""
}

func defaultGRConfigForGame(g *vngame.Game) gr.Config {
	if g == nil {
		return gr.Config{}
	}

	exe := strings.TrimSpace(g.Executable)
	opts := []gr.Option{
		gr.WithWinePrefix(strings.TrimSpace(g.PrefixPath)),
	}
	if workingDir := gameWorkingDir(g); workingDir != "" {
		opts = append(opts, gr.WithWorkingDir(workingDir))
	}

	if exe == "" {
		return gr.NewGameConfig(exe, opts...).Config
	}

	defaults, err := autorunner.AutoOptionsForExe(exe, autorunner.DefaultOptionsConfig{
		WinePrefix:          strings.TrimSpace(g.PrefixPath),
		WorkingDir:          gameWorkingDir(g),
		UseGamescope:        g.Runner == vngame.RunnerGamescope,
		WineBin:             wineBinForDefaults(g),
		GamescopeBin:        gamescopeBinForDefaults(g),
		SkipDependencyCheck: true,
	})
	if err == nil {
		return gr.NewGameConfig(defaults.ExePath, defaults.Options...).Config
	}

	return gr.NewGameConfig(exe, opts...).Config
}

func gameWorkingDir(g *vngame.Game) string {
	if g == nil {
		return ""
	}
	if dir := strings.TrimSpace(g.WorkingDir); dir != "" {
		return dir
	}
	if exe := strings.TrimSpace(g.Executable); exe != "" {
		return filepath.Dir(exe)
	}
	return strings.TrimSpace(g.GamePath)
}

func wineBinForDefaults(g *vngame.Game) string {
	if g == nil || g.Runner == vngame.RunnerGamescope {
		return ""
	}
	if strings.TrimSpace(g.RunnerPath) != "" {
		return strings.TrimSpace(g.RunnerPath)
	}
	if g.WineConfig != nil {
		return strings.TrimSpace(g.WineConfig.WineBin)
	}
	return ""
}

func gamescopeBinForDefaults(g *vngame.Game) string {
	if g == nil || g.Runner != vngame.RunnerGamescope {
		return ""
	}
	if strings.TrimSpace(g.RunnerPath) != "" {
		return strings.TrimSpace(g.RunnerPath)
	}
	if g.GamescopeConfig != nil {
		return strings.TrimSpace(g.GamescopeConfig.GamescopeBin)
	}
	return ""
}

func defaultWineOptions(prefix string) wine.Options {
	if cfg, err := autorunner.DefaultRunnerConfig(prefix); err == nil && cfg.Wine != nil {
		return *cfg.Wine
	}

	cfg := wine.ApplyOptions(wine.WithDefaultPrefix(prefix))
	return cfg
}

func defaultGamescopeOptions(prefix string) gamescope.Options {
	if cfg, err := autorunner.DefaultRunnerConfig(prefix); err == nil && cfg.Gamescope != nil {
		return *cfg.Gamescope
	}

	outW, outH := 1280, 720
	inW, inH := 1920, 1080
	if m, err := monitors.GetMonitors(); err == nil && len(m) > 0 {
		outW = m[0].CurrentMode.Width
		outH = m[0].CurrentMode.Height
	} else {
		inW = 1280
		inH = 720
	}
	cfg := gamescope.ApplyOptions(
		gamescope.WithWine(true),
		gamescope.WithDefaultWinePrefix(prefix),
		gamescope.WithResolution(inW, inH),
		gamescope.WithOutputResolution(outW, outH),
		gamescope.WithFullscreen(true),
		gamescope.WithScaler("fit"),
		gamescope.WithFilter("linear"),
		gamescope.WithExposeWayland(monitors.IsWayland()),
	)
	return cfg
}

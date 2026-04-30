package game

import (
	"context"
	"fmt"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareui "github.com/Seann-Moser/bare/pkg/ui"
	"github.com/Seann-Moser/bare/pkg/ui/filemanager"
	"github.com/Seann-Moser/bare/pkg/ui/icons"
	barethemes "github.com/Seann-Moser/bare/pkg/ui/themes"
	bareutils "github.com/Seann-Moser/bare/pkg/ui/utils"
	"github.com/Seann-Moser/wgl/pkg/game"
	"github.com/Seann-Moser/wgl/pkg/game/gameconfig"
	pkggui "github.com/Seann-Moser/wgl/pkg/gui"
	"github.com/Seann-Moser/wgl/pkg/util"
)

var _ pkggui.EvenHandler = &Page{}

type Page struct {
	theme   barethemes.Theme
	iconify *icons.Iconify

	fileBrowser *filemanager.FileBrowser
	hook        *game.RPGMakerHook
	gameList    layout.List
	configList  widget.List

	runnerDropdown bareui.Dropdown
	runnerOptions  []pkggui.DropdownOption

	gameSearchEditor widget.Editor
	pathEditor       widget.Editor
	steamAppIDEditor widget.Editor
	iconPathEditor   widget.Editor
	imagePathEditor  widget.Editor
	requiresSteam    widget.Bool

	useSelectionButton  widget.Clickable
	toggleBrowserButton widget.Clickable
	analyzeButton       widget.Clickable
	saveButton          widget.Clickable
	newGameButton       widget.Clickable
	deleteButton        widget.Clickable
	inspectHookButton   widget.Clickable
	installHookButton   widget.Clickable

	selectedRunner string
	statusText     string
	previewText    string
	hookStatusText string
	showBrowser    bool

	currentConfigName string
	loadedConfigName  string
	configs           []gameconfig.GameConfig
	gameSelectClicks  map[string]*widget.Clickable

	OnSaved    func(config *gameconfig.GameConfig)
	OnSelected func(config *gameconfig.GameConfig)
	OnNew      func()
	OnDeleted  func(name string)
	OnError    func(title, body string)
}

func New(theme barethemes.Theme) *Page {
	p := &Page{
		theme:            theme,
		fileBrowser:      filemanager.NewFileBrowser(""),
		hook:             &game.RPGMakerHook{},
		gameList:         layout.List{Axis: layout.Vertical},
		configList:       widget.List{List: layout.List{Axis: layout.Vertical}},
		selectedRunner:   "Auto",
		showBrowser:      true,
		statusText:       "Create or update a saved game config for transcript watching and launch flows.",
		hookStatusText:   "Select a game or path to inspect text hook support.",
		gameSelectClicks: map[string]*widget.Clickable{},
	}
	p.gameSearchEditor.SingleLine = true
	p.pathEditor.SingleLine = true
	p.steamAppIDEditor.SingleLine = true
	p.iconPathEditor.SingleLine = true
	p.imagePathEditor.SingleLine = true
	p.runnerOptions = pkggui.NewGameRunnerOptions()
	pkggui.NewDropDownLayout(&p.runnerDropdown, "mdi:rocket-launch-outline")
	p.fileBrowser.Extensions = []string{".exe"}
	return p
}

func (p *Page) WithTheme(theme barethemes.Theme) *Page {
	p.theme = theme
	return p
}

func (p *Page) WithIcon(icon *icons.Iconify) *Page {
	p.iconify = icon
	return p
}

func (p *Page) SetConfigs(configs []gameconfig.GameConfig) *Page {
	p.configs = append([]gameconfig.GameConfig(nil), configs...)
	valid := make(map[string]struct{}, len(p.configs))
	for _, cfg := range p.configs {
		valid[cfg.Name] = struct{}{}
		if p.gameSelectClicks[cfg.Name] == nil {
			p.gameSelectClicks[cfg.Name] = new(widget.Clickable)
		}
	}
	for name := range p.gameSelectClicks {
		if _, ok := valid[name]; !ok {
			delete(p.gameSelectClicks, name)
		}
	}
	return p
}

func (p *Page) SetCurrentConfig(cfg *gameconfig.GameConfig) *Page {
	if cfg == nil {
		return p
	}
	if strings.TrimSpace(cfg.Name) == "" {
		return p
	}
	if cfg.Name == p.loadedConfigName {
		p.currentConfigName = cfg.Name
		return p
	}
	p.currentConfigName = cfg.Name
	p.loadedConfigName = cfg.Name
	p.showBrowser = false
	p.pathEditor.SetText(util.FirstNonEmpty(cfg.GamePath, cfg.Executable))
	p.steamAppIDEditor.SetText(cfg.SteamAppID)
	p.iconPathEditor.SetText(cfg.IconPath)
	p.imagePathEditor.SetText(cfg.ImagePath)
	p.requiresSteam.Value = cfg.RequiresSteam
	switch cfg.Runner {
	case gameconfig.RunnerWine:
		p.selectedRunner = "Wine"
	case gameconfig.RunnerProton:
		p.selectedRunner = "Proton"
	case gameconfig.RunnerSteam:
		p.selectedRunner = "Steam"
	default:
		p.selectedRunner = "Auto"
	}
	return p
}

func (p *Page) HandleEvents(gtx layout.Context, _ context.Context, _ *app.Window) {
	p.runnerDropdown.Update(gtx)
	for i := range p.runnerOptions {
		opt := &p.runnerOptions[i]
		for opt.Clickable.Clicked(gtx) {
			p.selectedRunner = opt.Label
			p.runnerDropdown.Close()
		}
	}
	for _, cfg := range p.filteredConfigs() {
		click := p.gameSelectClicks[cfg.Name]
		for click.Clicked(gtx) {
			cfgCopy := cfg
			p.SetCurrentConfig(&cfgCopy)
			p.showBrowser = false
			p.statusText = fmt.Sprintf("Loaded saved game %q.", cfg.Name)
			if p.OnSelected != nil {
				p.OnSelected(&cfgCopy)
			}
		}
	}
	for p.newGameButton.Clicked(gtx) {
		p.resetConfigForm()
		p.showBrowser = true
		p.statusText = "Creating a new game config. Use the file browser to choose a folder or executable."
		p.hookStatusText = "Select a game or path to inspect text hook support."
		p.previewText = ""
		if p.OnNew != nil {
			p.OnNew()
		}
	}
	for p.toggleBrowserButton.Clicked(gtx) {
		p.showBrowser = !p.showBrowser
	}
	for p.useSelectionButton.Clicked(gtx) {
		p.useBrowserSelection()
	}
	for p.analyzeButton.Clicked(gtx) {
		p.analyzePath()
	}
	for p.saveButton.Clicked(gtx) {
		p.saveConfig()
	}
	for p.deleteButton.Clicked(gtx) {
		p.deleteConfig()
	}
	for p.inspectHookButton.Clicked(gtx) {
		p.inspectHook()
	}
	for p.installHookButton.Clicked(gtx) {
		p.installHook()
	}
}

func (p *Page) LayoutPage(gtx layout.Context) layout.Dimensions {
	if p.iconify == nil {
		p.iconify = icons.NewIconify()
	}
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.LG), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H5(p.theme.Gio(), "Game Setup")
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), "Save launchable game configs and install the RPG Maker clipboard hook from the same page.")
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					children := []layout.FlexChild{
						layout.Flexed(0.34, func(gtx layout.Context) layout.Dimensions {
							return p.layoutSavedGamesPanel(gtx)
						}),
						layout.Rigid(bareutils.SpacerW(unit.Dp(16))),
						layout.Flexed(0.66, func(gtx layout.Context) layout.Dimensions {
							return p.layoutConfigPanel(gtx)
						}),
					}
					if p.showBrowser {
						children = append(children,
							layout.Rigid(bareutils.SpacerW(unit.Dp(16))),
							layout.Flexed(0.52, func(gtx layout.Context) layout.Dimensions {
								return p.layoutBrowserPanel(gtx)
							}),
						)
					}
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, children...)
				}),
			)
		})
	})
}

func (p *Page) layoutSavedGamesPanel(gtx layout.Context) layout.Dimensions {
	search := material.Editor(p.theme.Gio(), &p.gameSearchEditor, "Find saved games")
	search.Color = p.theme.Color.Text
	search.HintColor = p.theme.Color.TextMuted
	filtered := p.filteredConfigs()
	newGame := bareui.Button{Clickable: &p.newGameButton, Text: "New Game", Prefix: "mdi:plus-circle-outline", Variant: bareui.ButtonPrimary}

	return bareutils.Panel(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H6(p.theme.Gio(), "Saved Games")
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), "Find an existing game config and load it into the editor.")
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return newGame.Layout(gtx, p.theme, p.iconify) }),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Rigid(search.Layout),
				layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if len(filtered) == 0 {
						lbl := material.Body1(p.theme.Gio(), "No saved games match this search.")
						lbl.Color = p.theme.Color.TextMuted
						return lbl.Layout(gtx)
					}
					return p.gameList.Layout(gtx, len(filtered), func(gtx layout.Context, index int) layout.Dimensions {
						cfg := filtered[index]
						return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return p.layoutSavedGameRow(gtx, cfg)
						})
					})
				}),
			)
		})
	})
}

func (p *Page) layoutSavedGameRow(gtx layout.Context, cfg gameconfig.GameConfig) layout.Dimensions {
	variant := bareui.ButtonSecondary
	if cfg.Name == p.currentConfigName {
		variant = bareui.ButtonPrimary
	}
	btn := bareui.Button{
		Clickable: p.gameSelectClicks[cfg.Name],
		Text:      cfg.Name,
		Prefix:    "mdi:controller-classic",
		Variant:   variant,
	}
	meta := []string{
		"Runner: " + string(cfg.Runner),
		"Path: " + util.FirstNonEmpty(cfg.GamePath, cfg.Executable),
	}
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return btn.Layout(gtx, p.theme, p.iconify)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(p.theme.Gio(), strings.Join(meta, "\n"))
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (p *Page) layoutConfigPanel(gtx layout.Context) layout.Dimensions {
	pathEditor := material.Editor(p.theme.Gio(), &p.pathEditor, "Path to game directory or executable")
	pathEditor.Color = p.theme.Color.Text
	pathEditor.HintColor = p.theme.Color.TextMuted

	steamAppIDEditor := material.Editor(p.theme.Gio(), &p.steamAppIDEditor, "Steam app id (optional unless requires Steam)")
	steamAppIDEditor.Color = p.theme.Color.Text
	steamAppIDEditor.HintColor = p.theme.Color.TextMuted

	iconPathEditor := material.Editor(p.theme.Gio(), &p.iconPathEditor, "Icon path (optional)")
	iconPathEditor.Color = p.theme.Color.Text
	iconPathEditor.HintColor = p.theme.Color.TextMuted

	imagePathEditor := material.Editor(p.theme.Gio(), &p.imagePathEditor, "Image path (optional)")
	imagePathEditor.Color = p.theme.Color.Text
	imagePathEditor.HintColor = p.theme.Color.TextMuted

	useSelection := bareui.Button{Clickable: &p.useSelectionButton, Text: "Use Browser Selection", Prefix: "mdi:folder-check-outline", Variant: bareui.ButtonSecondary}
	toggleBrowser := bareui.Button{Clickable: &p.toggleBrowserButton, Text: p.browserButtonLabel(), Prefix: p.browserButtonIcon(), Variant: bareui.ButtonSecondary}
	analyze := bareui.Button{Clickable: &p.analyzeButton, Text: "Auto Populate Info", Prefix: "mdi:auto-fix", Variant: bareui.ButtonSecondary}
	save := bareui.Button{Clickable: &p.saveButton, Text: "Save Game", Prefix: "mdi:content-save-outline", Variant: bareui.ButtonPrimary}
	deleteButton := bareui.Button{Clickable: &p.deleteButton, Text: "Delete Game", Prefix: "mdi:trash-can-outline", Variant: bareui.ButtonGhost}
	inspectHook := bareui.Button{Clickable: &p.inspectHookButton, Text: "Inspect Hook", Prefix: "mdi:puzzle-check-outline", Variant: bareui.ButtonSecondary}
	installHook := bareui.Button{Clickable: &p.installHookButton, Text: "Install Hook", Prefix: "mdi:puzzle-plus-outline", Variant: bareui.ButtonPrimary}

	return bareutils.Panel(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			rows := []layout.Widget{
				func(gtx layout.Context) layout.Dimensions {
					lbl := material.H6(p.theme.Gio(), "Config")
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				},
				bareutils.SpacerH(unit.Dp(8)),
				func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), p.statusText)
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				},
				bareutils.SpacerH(unit.Dp(14)),
				func(gtx layout.Context) layout.Dimensions { return toggleBrowser.Layout(gtx, p.theme, p.iconify) },
				bareutils.SpacerH(unit.Dp(12)),
				pathEditor.Layout,
				bareutils.SpacerH(unit.Dp(12)),
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return analyze.Layout(gtx, p.theme, p.iconify) }),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if !p.showBrowser {
								return layout.Dimensions{}
							}
							return layout.Inset{Left: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								return useSelection.Layout(gtx, p.theme, p.iconify)
							})
						}),
					)
				},
				bareutils.SpacerH(unit.Dp(12)),
				func(gtx layout.Context) layout.Dimensions {
					return p.layoutInfoRow(gtx, "Runner", p.selectedRunner, func(gtx layout.Context) layout.Dimensions {
						return p.runnerDropdown.Layout(gtx, p.theme, p.iconify, p.selectedRunner, func(gtx layout.Context) layout.Dimensions {
							return pkggui.LayoutOptionMenu(gtx, p.runnerOptions, p.selectedRunner, p.theme, p.iconify)
						})
					})
				},
				bareutils.SpacerH(unit.Dp(12)),
				func(gtx layout.Context) layout.Dimensions {
					check := material.CheckBox(p.theme.Gio(), &p.requiresSteam, "Requires Steam")
					check.Color = p.theme.Color.Text
					return check.Layout(gtx)
				},
				bareutils.SpacerH(unit.Dp(12)),
				steamAppIDEditor.Layout,
				bareutils.SpacerH(unit.Dp(12)),
				iconPathEditor.Layout,
				bareutils.SpacerH(unit.Dp(12)),
				imagePathEditor.Layout,
				bareutils.SpacerH(unit.Dp(14)),
				func(gtx layout.Context) layout.Dimensions {
					return p.layoutSummaryPanel(gtx, "Preview", p.previewText)
				},
				bareutils.SpacerH(unit.Dp(14)),
				func(gtx layout.Context) layout.Dimensions {
					return p.layoutSummaryPanel(gtx, "Text Hook", p.hookStatusText)
				},
				bareutils.SpacerH(unit.Dp(14)),
				func(gtx layout.Context) layout.Dimensions { return save.Layout(gtx, p.theme, p.iconify) },
				func(gtx layout.Context) layout.Dimensions {
					if strings.TrimSpace(p.currentConfigName) == "" {
						return layout.Dimensions{}
					}
					return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return deleteButton.Layout(gtx, p.theme, p.iconify)
					})
				},
				bareutils.SpacerH(unit.Dp(10)),
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return inspectHook.Layout(gtx, p.theme, p.iconify) }),
						layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return installHook.Layout(gtx, p.theme, p.iconify) }),
					)
				},
			}
			return material.List(p.theme.Gio(), &p.configList).Layout(gtx, len(rows), func(gtx layout.Context, index int) layout.Dimensions {
				return rows[index](gtx)
			})
		})
	})
}

func (p *Page) layoutBrowserPanel(gtx layout.Context) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.Background, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.H6(p.theme.Gio(), "File Browser")
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), "Browse to a game folder or select a specific `.exe`, then copy that selection into the config form.")
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(14))),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return p.fileBrowser.Layout(gtx, p.theme, p.iconify)
				}),
			)
		})
	})
}

func (p *Page) layoutSummaryPanel(gtx layout.Context, title, body string) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), title)
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(bareutils.SpacerH(unit.Dp(6))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(p.theme.Gio(), util.FirstNonEmpty(body, "No data yet."))
					lbl.Color = p.theme.Color.Text
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (p *Page) layoutInfoRow(gtx layout.Context, label, current string, control layout.Widget) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(p.theme.Gio(), label)
							lbl.Color = p.theme.Color.TextMuted
							return lbl.Layout(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(4))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.H6(p.theme.Gio(), current)
							lbl.Color = p.theme.Color.Text
							return lbl.Layout(gtx)
						}),
					)
				}),
				layout.Rigid(control),
			)
		})
	})
}

func (p *Page) useBrowserSelection() {
	selected := strings.TrimSpace(p.fileBrowser.SelectedPath)
	if selected == "" {
		selected = strings.TrimSpace(p.fileBrowser.Dir)
	}
	if selected == "" {
		p.showError("Use Browser Selection Failed", "Select a folder or executable in the browser first.")
		return
	}
	p.pathEditor.SetText(selected)
	p.showBrowser = false
	p.statusText = "Copied browser selection into the game path field."
}

func (p *Page) analyzePath() {
	cfg, err := p.buildConfig()
	if err != nil {
		p.previewText = err.Error()
		p.statusText = err.Error()
		return
	}
	p.previewText = strings.Join([]string{
		"Name: " + cfg.Name,
		"Resolved Path: " + cfg.GamePath,
		"Executable: " + cfg.Executable,
		"Working Dir: " + cfg.WorkingDir,
		"Runner: " + string(cfg.Runner),
		"Icon: " + util.FirstNonEmpty(cfg.IconPath, "Unavailable"),
		"Image: " + util.FirstNonEmpty(cfg.ImagePath, "Unavailable"),
	}, "\n")
	p.statusText = fmt.Sprintf("Ready to save %q.", cfg.Name)
}

func (p *Page) saveConfig() {
	cfg, err := p.buildConfig()
	if err != nil {
		p.statusText = err.Error()
		p.showError("Save Game Failed", err.Error())
		return
	}
	if _, err := gameconfig.SaveGameConfig(cfg); err != nil {
		p.statusText = err.Error()
		p.showError("Save Game Failed", err.Error())
		return
	}
	p.currentConfigName = cfg.Name
	p.statusText = fmt.Sprintf("Saved %q.", cfg.Name)
	p.loadedConfigName = cfg.Name
	if p.OnSaved != nil {
		p.OnSaved(&cfg)
	}
}

func (p *Page) deleteConfig() {
	name := strings.TrimSpace(p.currentConfigName)
	if name == "" {
		p.showError("Delete Game Failed", "Load a saved game config before deleting it.")
		return
	}
	if err := gameconfig.DeleteGameConfig(name); err != nil {
		p.statusText = err.Error()
		p.showError("Delete Game Failed", err.Error())
		return
	}
	p.resetConfigForm()
	p.showBrowser = true
	p.previewText = ""
	p.hookStatusText = "Select a game or path to inspect text hook support."
	p.statusText = fmt.Sprintf("Deleted %q.", name)
	if p.OnDeleted != nil {
		p.OnDeleted(name)
	}
}

func (p *Page) inspectHook() {
	inputPath := strings.TrimSpace(p.pathEditor.Text())
	if inputPath == "" {
		p.showError("Inspect Hook Failed", "Enter or browse to a game path first.")
		return
	}
	status, err := p.hook.InspectHook(inputPath)
	if err != nil {
		p.hookStatusText = err.Error()
		p.showError("Inspect Hook Failed", err.Error())
		return
	}
	lines := []string{
		status.Message,
		"Engine: " + util.FirstNonEmpty(status.Engine, "Unknown"),
		"Project Root: " + util.FirstNonEmpty(status.ProjectRoot, "Unavailable"),
		"Plugin Path: " + util.FirstNonEmpty(status.PluginPath, "Unavailable"),
	}
	lines = append(lines, status.Compatibility.Findings...)
	p.hookStatusText = strings.Join(lines, "\n")
}

func (p *Page) installHook() {
	inputPath := strings.TrimSpace(p.pathEditor.Text())
	if inputPath == "" {
		p.showError("Install Hook Failed", "Enter or browse to a game path first.")
		return
	}
	result, err := p.hook.InstallHook(inputPath)
	if err != nil {
		p.hookStatusText = err.Error()
		p.showError("Install Hook Failed", err.Error())
		return
	}
	lines := []string{
		"Text hook plugin is installed and enabled.",
		"Engine: " + result.Engine,
		"Plugin Path: " + result.PluginPath,
		"Plugins Config: " + result.PluginsConfigPath,
	}
	lines = append(lines, result.Compatibility.Findings...)
	p.hookStatusText = strings.Join(lines, "\n")
	p.statusText = "Installed the clipboard text hook plugin for this game."
}

func (p *Page) buildConfig() (gameconfig.GameConfig, error) {
	inputPath := strings.TrimSpace(p.pathEditor.Text())
	if inputPath == "" {
		return gameconfig.GameConfig{}, fmt.Errorf("game path is required")
	}
	return gameconfig.BuildGameConfig(
		inputPath,
		strings.ToLower(strings.TrimSpace(p.selectedRunner)),
		p.requiresSteam.Value,
		strings.TrimSpace(p.steamAppIDEditor.Text()),
		strings.TrimSpace(p.iconPathEditor.Text()),
		strings.TrimSpace(p.imagePathEditor.Text()),
	)
}

func (p *Page) showError(title, body string) {
	if p.OnError != nil {
		p.OnError(title, body)
	}
}

func (p *Page) filteredConfigs() []gameconfig.GameConfig {
	query := strings.TrimSpace(strings.ToLower(p.gameSearchEditor.Text()))
	if query == "" {
		return p.configs
	}
	filtered := make([]gameconfig.GameConfig, 0, len(p.configs))
	for _, cfg := range p.configs {
		haystack := strings.ToLower(strings.Join([]string{
			cfg.Name,
			cfg.GamePath,
			cfg.Executable,
			string(cfg.Runner),
		}, "\n"))
		if strings.Contains(haystack, query) {
			filtered = append(filtered, cfg)
		}
	}
	return filtered
}

func (p *Page) resetConfigForm() {
	p.currentConfigName = ""
	p.loadedConfigName = ""
	p.pathEditor.SetText("")
	p.steamAppIDEditor.SetText("")
	p.iconPathEditor.SetText("")
	p.imagePathEditor.SetText("")
	p.requiresSteam.Value = false
	p.selectedRunner = "Auto"
}

func (p *Page) browserButtonLabel() string {
	if p.showBrowser {
		return "Hide Browser"
	}
	return "Browse Files"
}

func (p *Page) browserButtonIcon() string {
	if p.showBrowser {
		return "mdi:folder-remove-outline"
	}
	return "mdi:folder-search-outline"
}

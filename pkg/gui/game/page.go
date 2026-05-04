package game

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	bareui "github.com/DarlingGoose/bare/pkg/ui"
	"github.com/DarlingGoose/bare/pkg/ui/filemanager"
	"github.com/DarlingGoose/bare/pkg/ui/icons"
	"github.com/DarlingGoose/bare/pkg/ui/media"
	barethemes "github.com/DarlingGoose/bare/pkg/ui/themes"
	bareutils "github.com/DarlingGoose/bare/pkg/ui/utils"
	"github.com/DarlingGoose/vntext/pkg/engine/auto"
	vngame "github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/vntext/pkg/gameConfig"
	flashcards "github.com/DarlingGoose/wgl/pkg/flashcard"
	pkggui "github.com/DarlingGoose/wgl/pkg/gui"
	"github.com/DarlingGoose/wgl/pkg/util"
)

var _ pkggui.EvenHandler = &Page{}

type Page struct {
	theme   barethemes.Theme
	iconify *icons.Iconify

	fileBrowser  *filemanager.FileBrowser
	browserModal bareui.Modal
	gameList     layout.List
	configList   widget.List

	runnerDropdown  bareui.Dropdown
	runnerOptions   []pkggui.DropdownOption
	desktopDropdown bareui.Dropdown
	desktopOptions  []pkggui.DropdownOption

	gameSearchEditor widget.Editor
	nameEditor       widget.Editor
	pathEditor       widget.Editor
	steamAppIDEditor widget.Editor
	iconPathEditor   widget.Editor
	imagePathEditor  widget.Editor
	requiresSteam    widget.Bool

	useSelectionButton  widget.Clickable
	useIconButton       widget.Clickable
	useImageButton      widget.Clickable
	toggleBrowserButton widget.Clickable
	analyzeButton       widget.Clickable
	saveButton          widget.Clickable
	newGameButton       widget.Clickable
	deleteButton        widget.Clickable
	inspectHookButton   widget.Clickable
	installHookButton   widget.Clickable

	selectedRunner  string
	selectedDesktop string
	statusText      string
	previewText     string
	hookStatusText  string
	showBrowser     bool
	installing      bool
	installResult   chan gameInstallResult

	currentConfigName string
	loadedConfigName  string
	draftConfig       *vngame.Game
	configs           []*vngame.Game
	gameSelectClicks  map[string]*widget.Clickable
	gameImageViews    map[string]*media.ImageView

	OnSaved    func(config *vngame.Game)
	OnSelected func(config *vngame.Game)
	OnNew      func()
	OnDeleted  func(name string)
	OnError    func(title, body string)
}

type gameInstallResult struct {
	cfg     *vngame.Game
	preview string
	title   string
	status  string
	hook    string
	err     error
}

func New(theme barethemes.Theme) *Page {
	p := &Page{
		theme:            theme,
		fileBrowser:      filemanager.NewFileBrowser(""),
		browserModal:     bareui.Modal{CloseOnScrim: true},
		gameList:         layout.List{Axis: layout.Vertical},
		configList:       widget.List{List: layout.List{Axis: layout.Vertical}},
		selectedRunner:   "Auto",
		selectedDesktop:  "Default",
		showBrowser:      false,
		statusText:       "Create or update a saved game config for transcript watching and launch flows.",
		hookStatusText:   "Select a game or path to inspect text hook support.",
		gameSelectClicks: map[string]*widget.Clickable{},
		gameImageViews:   map[string]*media.ImageView{},
	}
	p.gameSearchEditor.SingleLine = true
	p.nameEditor.SingleLine = true
	p.pathEditor.SingleLine = true
	p.steamAppIDEditor.SingleLine = true
	p.iconPathEditor.SingleLine = true
	p.imagePathEditor.SingleLine = true
	p.runnerOptions = pkggui.NewGameRunnerOptions()
	p.desktopOptions = pkggui.NewVirtualDesktopOptions()
	pkggui.NewDropDownLayout(&p.runnerDropdown, "mdi:rocket-launch-outline")
	pkggui.NewDropDownLayout(&p.desktopDropdown, "mdi:monitor")
	p.fileBrowser.Extensions = []string{".exe", ".png", ".jpg", ".jpeg", ".webp", ".gif"}
	p.fileBrowser.ShowPreview = true
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

func (p *Page) SetConfigs(configs []*vngame.Game) *Page {
	p.configs = append([]*vngame.Game(nil), configs...)
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

func (p *Page) SetCurrentConfig(cfg *vngame.Game) *Page {
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
	p.draftConfig = cloneGameConfig(cfg)
	p.showBrowser = false
	p.nameEditor.SetText(cfg.Name)
	p.pathEditor.SetText(util.FirstNonEmpty(cfg.GamePath, cfg.Executable))
	p.steamAppIDEditor.SetText(cfg.SteamAppID)
	p.selectedDesktop = virtualDesktopLabel(cfg.VirtualDesktop)
	p.iconPathEditor.SetText(cfg.IconPath)
	p.imagePathEditor.SetText(cfg.ImagePath)
	p.requiresSteam.Value = cfg.RequiresSteam
	switch cfg.Runner {
	case vngame.RunnerWine:
		p.selectedRunner = "Wine"
	case vngame.RunnerProton:
		p.selectedRunner = "Proton"
	case vngame.RunnerSteam:
		p.selectedRunner = "Steam"
	default:
		p.selectedRunner = "Auto"
	}
	return p
}

func (p *Page) HandleEvents(gtx layout.Context, _ context.Context, w *app.Window) {
	p.consumeInstallResult(w)
	p.runnerDropdown.Update(gtx)
	p.desktopDropdown.Update(gtx)
	for i := range p.runnerOptions {
		opt := &p.runnerOptions[i]
		for opt.Clickable.Clicked(gtx) {
			p.selectedRunner = opt.Label
			p.runnerDropdown.Close()
		}
	}
	for i := range p.desktopOptions {
		opt := &p.desktopOptions[i]
		for opt.Clickable.Clicked(gtx) {
			p.selectedDesktop = opt.Label
			p.desktopDropdown.Close()
		}
	}
	for _, cfg := range p.filteredConfigs() {
		click := p.gameSelectClicks[cfg.Name]
		for click.Clicked(gtx) {
			if p.installing {
				continue
			}
			cfgCopy := cfg
			p.SetCurrentConfig(cfgCopy)
			p.showBrowser = false
			p.statusText = fmt.Sprintf("Loaded saved game %q.", cfg.Name)
			if p.OnSelected != nil {
				p.OnSelected(cfgCopy)
			}
		}
	}
	for p.newGameButton.Clicked(gtx) {
		if p.installing {
			continue
		}
		p.resetConfigForm()
		p.showBrowser = false
		p.browserModal.Open = false
		p.statusText = "Creating a new game config. Use the file browser to choose a folder or executable."
		p.hookStatusText = "Select a game or path to inspect text hook support."
		p.previewText = ""
		if p.OnNew != nil {
			p.OnNew()
		}
	}
	for p.toggleBrowserButton.Clicked(gtx) {
		if p.installing {
			continue
		}
		p.showBrowser = true
		p.browserModal.Open = true
	}
	for p.useSelectionButton.Clicked(gtx) {
		p.useBrowserSelection()
	}
	for p.useIconButton.Clicked(gtx) {
		p.useBrowserAssetSelection(&p.iconPathEditor, "icon")
	}
	for p.useImageButton.Clicked(gtx) {
		p.useBrowserAssetSelection(&p.imagePathEditor, "image")
	}
	for p.analyzeButton.Clicked(gtx) {
		p.startInstallGame(w, false)
	}
	for p.saveButton.Clicked(gtx) {
		p.saveConfig()
	}
	for p.deleteButton.Clicked(gtx) {
		if p.installing {
			continue
		}
		p.deleteConfig()
	}

	for p.installHookButton.Clicked(gtx) {
		p.startInstallHook(w)
	}
}

func (p *Page) LayoutPage(gtx layout.Context) layout.Dimensions {
	if p.iconify == nil {
		p.iconify = icons.NewIconify()
	}
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
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
							lbl := material.Body1(p.theme.Gio(), "Save launchable game configs and install supported text hook plugins from the same page.")
							lbl.Color = p.theme.Color.TextMuted
							return lbl.Layout(gtx)
						}),
						layout.Rigid(bareutils.SpacerH(unit.Dp(16))),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
								layout.Flexed(0.34, func(gtx layout.Context) layout.Dimensions {
									return p.layoutSavedGamesPanel(gtx)
								}),
								layout.Rigid(bareutils.SpacerW(unit.Dp(16))),
								layout.Flexed(0.66, func(gtx layout.Context) layout.Dimensions {
									return p.layoutConfigPanel(gtx)
								}),
							)
						}),
					)
				})
			})
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			dims := p.browserModal.Layout(gtx, p.theme, "Browse Game Files", func(gtx layout.Context) layout.Dimensions {
				return p.layoutBrowserModalContent(gtx)
			})
			p.showBrowser = p.browserModal.Open
			return dims
		}),
	)
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

func (p *Page) layoutSavedGameRow(gtx layout.Context, cfg *vngame.Game) layout.Dimensions {
	variant := bareui.ButtonSecondary
	if cfg.Name == p.currentConfigName {
		variant = bareui.ButtonPrimary
	}
	bg := p.theme.Color.Surface
	if cfg.Name == p.currentConfigName {
		bg = p.theme.Color.SurfaceAlt
	}
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return p.gameSelectClicks[cfg.Name].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return bareutils.RoundedSurface(gtx, bg, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return p.layoutGameArtwork(gtx, cfg)
						}),
						layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Body1(p.theme.Gio(), cfg.Name)
									lbl.Color = p.theme.Color.Text
									return lbl.Layout(gtx)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									meta := savedGameMeta(cfg)
									if meta == "" {
										return layout.Dimensions{}
									}
									return layout.Inset{Top: unit.Dp(3)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										lbl := material.Body2(p.theme.Gio(), meta)
										lbl.Color = p.theme.Color.TextMuted
										return lbl.Layout(gtx)
									})
								}),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if variant != bareui.ButtonPrimary {
								return layout.Dimensions{}
							}
							lbl := material.Body2(p.theme.Gio(), "Selected")
							lbl.Color = p.theme.Color.Primary
							return lbl.Layout(gtx)
						}),
					)
				})
			})
		})
	})
}

func (p *Page) layoutConfigPanel(gtx layout.Context) layout.Dimensions {
	nameEditor := material.Editor(p.theme.Gio(), &p.nameEditor, "Game name")
	nameEditor.Color = p.theme.Color.Text
	nameEditor.HintColor = p.theme.Color.TextMuted

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

	useSelection := bareui.Button{Clickable: &p.useSelectionButton, Text: "Use as Game Path", Prefix: "mdi:folder-check-outline", Variant: bareui.ButtonSecondary}
	useIconSelection := bareui.Button{Clickable: &p.useIconButton, Text: "Use as Icon", Prefix: "mdi:image-filter-center-focus", Variant: bareui.ButtonSecondary}
	useImageSelection := bareui.Button{Clickable: &p.useImageButton, Text: "Use as Image", Prefix: "mdi:image-outline", Variant: bareui.ButtonSecondary}
	toggleBrowser := bareui.Button{Clickable: &p.toggleBrowserButton, Text: p.browserButtonLabel(), Prefix: p.browserButtonIcon(), Variant: bareui.ButtonSecondary}
	install := bareui.Button{Clickable: &p.analyzeButton, Text: p.installActionLabel("Install"), Prefix: p.installActionIcon("mdi:download-box-outline"), Variant: bareui.ButtonPrimary}
	saveChanges := bareui.Button{Clickable: &p.saveButton, Text: "Save Changes", Prefix: "mdi:content-save-outline", Variant: bareui.ButtonSecondary}
	installHook := bareui.Button{Clickable: &p.installHookButton, Text: p.installActionLabel("Install/Reinstall Hook Plugin"), Prefix: p.installActionIcon("mdi:puzzle-plus-outline"), Variant: bareui.ButtonSecondary}
	deleteButton := bareui.Button{Clickable: &p.deleteButton, Text: "Delete Game", Prefix: "mdi:trash-can-outline", Variant: bareui.ButtonGhost}

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
				func(gtx layout.Context) layout.Dimensions {
					if p.installing {
						return toggleBrowser.Layout(gtx.Disabled(), p.theme, p.iconify)
					}
					return toggleBrowser.Layout(gtx, p.theme, p.iconify)
				},
				bareutils.SpacerH(unit.Dp(12)),
				func(gtx layout.Context) layout.Dimensions {
					return p.layoutEditorRow(gtx, "Game Name / Rename", "Saved config name", nameEditor.Layout)
				},
				bareutils.SpacerH(unit.Dp(12)),
				func(gtx layout.Context) layout.Dimensions {
					return p.layoutEditorRow(gtx, "Game Path", "Folder or executable used for install", pathEditor.Layout)
				},
				bareutils.SpacerH(unit.Dp(12)),
				func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.installing {
								return install.Layout(gtx.Disabled(), p.theme, p.iconify)
							}
							return install.Layout(gtx, p.theme, p.iconify)
						}),
						layout.Rigid(bareutils.SpacerW(unit.Dp(10))),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.installing || p.draftConfig == nil {
								return saveChanges.Layout(gtx.Disabled(), p.theme, p.iconify)
							}
							return saveChanges.Layout(gtx, p.theme, p.iconify)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if !p.showBrowser {
								return layout.Dimensions{}
							}
							return layout.Inset{Left: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								if p.installing {
									return useSelection.Layout(gtx.Disabled(), p.theme, p.iconify)
								}
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
				func(gtx layout.Context) layout.Dimensions {
					return p.layoutInfoRow(gtx, "Wine Virtual Desktop", p.selectedDesktop, func(gtx layout.Context) layout.Dimensions {
						return p.desktopDropdown.Layout(gtx, p.theme, p.iconify, p.selectedDesktop, func(gtx layout.Context) layout.Dimensions {
							return pkggui.LayoutOptionMenu(gtx, p.desktopOptions, p.selectedDesktop, p.theme, p.iconify)
						})
					})
				},
				bareutils.SpacerH(unit.Dp(12)),
				iconPathEditor.Layout,
				bareutils.SpacerH(unit.Dp(8)),
				func(gtx layout.Context) layout.Dimensions {
					if !p.showBrowser {
						return layout.Dimensions{}
					}
					return useIconSelection.Layout(gtx, p.theme, p.iconify)
				},
				bareutils.SpacerH(unit.Dp(12)),
				imagePathEditor.Layout,
				bareutils.SpacerH(unit.Dp(8)),
				func(gtx layout.Context) layout.Dimensions {
					if !p.showBrowser {
						return layout.Dimensions{}
					}
					return useImageSelection.Layout(gtx, p.theme, p.iconify)
				},
				bareutils.SpacerH(unit.Dp(14)),
				func(gtx layout.Context) layout.Dimensions {
					return p.layoutSummaryPanel(gtx, "Preview", p.previewText)
				},
				bareutils.SpacerH(unit.Dp(14)),
				func(gtx layout.Context) layout.Dimensions {
					return p.layoutSummaryPanel(gtx, "Text Hook", p.hookStatusText)
				},
				bareutils.SpacerH(unit.Dp(14)),
				func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} },
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
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if p.installing {
								return installHook.Layout(gtx.Disabled(), p.theme, p.iconify)
							}
							return installHook.Layout(gtx, p.theme, p.iconify)
						}),
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

func (p *Page) layoutBrowserModalContent(gtx layout.Context) layout.Dimensions {
	if gtx.Constraints.Max.Y > 0 {
		gtx.Constraints.Min.Y = min(gtx.Constraints.Max.Y, gtx.Dp(unit.Dp(560)))
	}
	usePath := bareui.Button{Clickable: &p.useSelectionButton, Text: "Use as Game Path", Prefix: "mdi:folder-check-outline", Variant: bareui.ButtonPrimary}
	useIcon := bareui.Button{Clickable: &p.useIconButton, Text: "Use as Icon", Prefix: "mdi:image-filter-center-focus", Variant: bareui.ButtonSecondary}
	useImage := bareui.Button{Clickable: &p.useImageButton, Text: "Use as Image", Prefix: "mdi:image-outline", Variant: bareui.ButtonSecondary}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(p.theme.Gio(), "Select a game folder, .exe, or image. Use the buttons below to copy the selection into the form.")
			lbl.Color = p.theme.Color.TextMuted
			return lbl.Layout(gtx)
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return p.fileBrowser.Layout(gtx, p.theme, p.iconify)
		}),
		layout.Rigid(bareutils.SpacerH(unit.Dp(12))),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return usePath.Layout(gtx, p.theme, p.iconify)
				}),
				layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return useIcon.Layout(gtx, p.theme, p.iconify)
				}),
				layout.Rigid(bareutils.SpacerW(unit.Dp(8))),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return useImage.Layout(gtx, p.theme, p.iconify)
				}),
			)
		}),
	)
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

func (p *Page) layoutEditorRow(gtx layout.Context, label, hint string, editor layout.Widget) layout.Dimensions {
	return bareutils.Panel(gtx, p.theme.Color.Surface, unit.Dp(p.theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(12)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body2(p.theme.Gio(), label)
					lbl.Color = p.theme.Color.TextMuted
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if strings.TrimSpace(hint) == "" {
						return layout.Dimensions{}
					}
					return layout.Inset{Top: unit.Dp(2), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body2(p.theme.Gio(), hint)
						lbl.Color = p.theme.Color.TextMuted
						return lbl.Layout(gtx)
					})
				}),
				layout.Rigid(editor),
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
	if p.installing {
		return
	}
	selected := p.browserSelectedPath()
	if selected == "" {
		p.showError("Use Browser Selection Failed", "Select a folder or executable in the browser first.")
		return
	}
	p.pathEditor.SetText(selected)
	p.showBrowser = false
	p.browserModal.Open = false
	p.statusText = "Copied browser selection into the game path field."
	p.inspectHook()
}

func (p *Page) useBrowserAssetSelection(editor *widget.Editor, label string) {
	if p.installing {
		return
	}
	selected := p.browserSelectedPath()
	if selected == "" {
		p.showError("Use Browser Selection Failed", "Select an image file in the browser first.")
		return
	}
	if !util.IsImageFile(selected) {
		p.showError("Use Browser Selection Failed", "The selected "+label+" must be an image file.")
		return
	}
	editor.SetText(selected)
	p.showBrowser = false
	p.browserModal.Open = false
	p.statusText = "Selected game " + label + "."
}

func (p *Page) browserSelectedPath() string {
	selected := strings.TrimSpace(p.fileBrowser.SelectedPath)
	if selected == "" {
		selected = strings.TrimSpace(p.fileBrowser.Dir)
	}
	return selected
}

func (p *Page) startInstallGame(w *app.Window, installHook bool) {
	if p.installing {
		return
	}
	inputPath := strings.TrimSpace(p.pathEditor.Text())
	if inputPath == "" {
		p.showError("Install Game Failed", "Enter or browse to a game path first.")
		return
	}
	overrides := p.draftFromForm(nil)
	oldName := p.currentConfigName

	p.installing = true
	p.statusText = "Installing game config..."
	if installHook {
		p.hookStatusText = "Installing text hook plugin..."
	}
	p.previewText = "Working in the background."
	p.installResult = make(chan gameInstallResult, 1)

	go func() {
		p.installResult <- installGameConfig(inputPath, installHook, overrides, oldName)
		if w != nil {
			w.Invalidate()
		}
	}()
}

func (p *Page) startInstallHook(w *app.Window) {
	if p.installing {
		return
	}
	cfg := p.draftFromForm(p.draftConfig)
	if cfg == nil || strings.TrimSpace(cfg.GamePath) == "" && strings.TrimSpace(cfg.Executable) == "" {
		p.showError("Install Hook Failed", "Install or select a game before installing the hook plugin.")
		return
	}
	oldName := p.currentConfigName
	p.installing = true
	p.statusText = "Saving game config before hook install..."
	p.hookStatusText = "Installing text hook plugin..."
	p.previewText = "Working in the background."
	p.installResult = make(chan gameInstallResult, 1)

	go func() {
		p.installResult <- installHookPlugin(cfg, oldName)
		if w != nil {
			w.Invalidate()
		}
	}()
}

func installGameConfig(inputPath string, installHook bool, overrides *vngame.Game, oldName string) gameInstallResult {
	eng, err := auto.SelectOrInstallEngine(context.Background(), inputPath)
	if err != nil {
		return gameInstallResult{title: "Engine Select Failed", err: err}
	}
	cfg, err := eng.InstallGame(inputPath)
	if err != nil {
		return gameInstallResult{title: "Install Game Failed", err: err}
	}
	if installHook {
		if err := eng.InstallTextHook(cfg); err != nil {
			return gameInstallResult{title: "Install Text Hook Failed", err: err}
		}
	}
	applyGameOverrides(cfg, overrides)
	if strings.TrimSpace(cfg.EngineName) == "" {
		cfg.EngineName = eng.Name()
	}
	if err := gameConfig.WriteGameConfig(gameConfig.DefaultGameConfigPath(cfg), cfg); err != nil {
		return gameInstallResult{title: "Save Game Failed", err: err}
	}
	if err := renameGameFlashcards(oldName, cfg.Name); err != nil {
		return gameInstallResult{title: "Move Flashcards Failed", err: err}
	}
	cleanupOldGameConfig(oldName, cfg)
	cleanupDuplicateGameConfigs(cfg)
	preview := gamePreviewText(cfg)
	hook := "Hook plugin not installed."
	if installHook {
		hook = "Text hook plugin installed."
	}
	return gameInstallResult{cfg: cfg, preview: preview, status: fmt.Sprintf("Saved config %q.", cfg.Name), hook: hook}
}

func installHookPlugin(cfg *vngame.Game, oldName string) gameInstallResult {
	if cfg == nil {
		return gameInstallResult{title: "Install Hook Failed", err: fmt.Errorf("game config is empty")}
	}
	inputPath := util.FirstNonEmpty(cfg.GamePath, cfg.Executable)
	eng, err := auto.SelectEngine(inputPath)
	if err != nil {
		return gameInstallResult{title: "Engine Select Failed", err: err}
	}
	if err := eng.InstallTextHook(cfg); err != nil {
		return gameInstallResult{title: "Install Hook Failed", err: err}
	}
	if err := gameConfig.WriteGameConfig(gameConfig.DefaultGameConfigPath(cfg), cfg); err != nil {
		return gameInstallResult{title: "Save Game Failed", err: err}
	}
	if err := renameGameFlashcards(oldName, cfg.Name); err != nil {
		return gameInstallResult{title: "Move Flashcards Failed", err: err}
	}
	cleanupOldGameConfig(oldName, cfg)
	cleanupDuplicateGameConfigs(cfg)
	return gameInstallResult{cfg: cfg, preview: gamePreviewText(cfg), status: fmt.Sprintf("Saved config %q.", cfg.Name), hook: "Text hook plugin installed."}
}

func (p *Page) consumeInstallResult(w *app.Window) {
	if p.installResult == nil {
		return
	}
	select {
	case result := <-p.installResult:
		p.installing = false
		p.installResult = nil
		if result.err != nil {
			p.statusText = result.err.Error()
			p.hookStatusText = "Text hook install did not complete."
			p.showError(util.FirstNonEmpty(result.title, "Install Game Failed"), result.err.Error())
			return
		}
		cfg := result.cfg
		if cfg == nil {
			p.statusText = "Install completed without returning a game config."
			return
		}
		p.previewText = result.preview
		p.hookStatusText = util.FirstNonEmpty(result.hook, p.hookStatusText)
		p.currentConfigName = cfg.Name
		p.draftConfig = cloneGameConfig(cfg)
		p.statusText = util.FirstNonEmpty(result.status, fmt.Sprintf("Saved config %q.", cfg.Name))
		p.loadedConfigName = cfg.Name
		p.nameEditor.SetText(cfg.Name)
		p.pathEditor.SetText(util.FirstNonEmpty(cfg.GamePath, cfg.Executable))
		p.steamAppIDEditor.SetText(cfg.SteamAppID)
		p.selectedDesktop = virtualDesktopLabel(cfg.VirtualDesktop)
		p.iconPathEditor.SetText(cfg.IconPath)
		p.imagePathEditor.SetText(cfg.ImagePath)
		if p.OnSaved != nil {
			p.OnSaved(cfg)
		}
		if w != nil {
			w.Invalidate()
		}
	default:
	}
}

func (p *Page) saveConfig() {
	if p.installing {
		return
	}
	if p.draftConfig == nil {
		p.showError("Save Game Failed", "Select or install a game before saving changes.")
		return
	}
	cfg := p.draftFromForm(p.draftConfig)
	if strings.TrimSpace(cfg.Name) == "" {
		p.showError("Save Game Failed", "Game name cannot be empty.")
		return
	}
	if strings.TrimSpace(cfg.GamePath) == "" && strings.TrimSpace(cfg.Executable) == "" {
		p.showError("Save Game Failed", "Game path cannot be empty.")
		return
	}
	oldName := p.currentConfigName
	if err := gameConfig.WriteGameConfig(gameConfig.DefaultGameConfigPath(cfg), cfg); err != nil {
		p.statusText = err.Error()
		p.showError("Save Game Failed", err.Error())
		return
	}
	if err := renameGameFlashcards(oldName, cfg.Name); err != nil {
		p.statusText = err.Error()
		p.showError("Move Flashcards Failed", err.Error())
		return
	}
	cleanupOldGameConfig(oldName, cfg)
	cleanupDuplicateGameConfigs(cfg)
	p.currentConfigName = cfg.Name
	p.loadedConfigName = cfg.Name
	p.draftConfig = cloneGameConfig(cfg)
	p.previewText = gamePreviewText(cfg)
	p.statusText = fmt.Sprintf("Saved changes to %q.", cfg.Name)
	if p.OnSaved != nil {
		p.OnSaved(cfg)
	}
}

func (p *Page) deleteConfig() {
	// todo
	//name := strings.TrimSpace(p.currentConfigName)
	//if name == "" {
	//	p.showError("Delete Game Failed", "Load a saved game config before deleting it.")
	//	return
	//}
	//if err := gameconfig.DeleteGameConfig(name); err != nil {
	//	p.statusText = err.Error()
	//	p.showError("Delete Game Failed", err.Error())
	//	return
	//}
	//p.resetConfigForm()
	//p.showBrowser = true
	//p.previewText = ""
	//p.hookStatusText = "Select a game or path to inspect text hook support."
	//p.statusText = fmt.Sprintf("Deleted %q.", name)
	//if p.OnDeleted != nil {
	//	p.OnDeleted(name)
	//}
}

func (p *Page) inspectHook() {
	inputPath := strings.TrimSpace(p.pathEditor.Text())
	if inputPath == "" {
		p.showError("Inspect Hook Failed", "Enter or browse to a game path first.")
		return
	}
	_, err := auto.SelectEngine(inputPath)
	if err != nil {
		p.hookStatusText = err.Error()
		p.showError("Inspect Hook Failed: engine select error ", err.Error())
		return
	}
	//
	//status, err := e.Status(inputPath)
	//if err != nil {
	//	p.hookStatusText = err.Error()
	//	p.showError("Inspect Hook Failed", err.Error())
	//	return
	//}

	//p.hookStatusText = p.formatHookStatus(status)
}

//todo
//func (p *Page) formatHookStatus(status testhook.TextHookStatus) string {
//	//lines := []string{
//	//	status.Message,
//	//	"Supported: " + boolLabel(status.Supported),
//	//	"Installed: " + boolLabel(status.Installed),
//	//	"Loaded: " + boolLabel(status.Loaded),
//	//	"Engine: " + util.FirstNonEmpty(status.Engine, "Unknown"),
//	//	"Method: " + util.FirstNonEmpty(status.Method, "Unavailable"),
//	//	"Project Root: " + util.FirstNonEmpty(status.ProjectRoot, "Unavailable"),
//	//}
//	//
//	//if strings.TrimSpace(status.PluginPath) != "" {
//	//	lines = append(lines, "Plugin Path: "+status.PluginPath)
//	//}
//	//if strings.TrimSpace(status.PluginsConfigPath) != "" {
//	//	lines = append(lines, "Plugins Config: "+status.PluginsConfigPath)
//	//}
//	//if strings.TrimSpace(status.OutputPath) != "" {
//	//	lines = append(lines, "Output: "+status.OutputPath)
//	//}
//	//
//	//if strings.TrimSpace(status.Compatibility.RiskLevel) != "" {
//	//	lines = append(lines, "Risk: "+status.Compatibility.RiskLevel)
//	//}
//	//
//	//lines = appendHookFindings(lines, status.Compatibility.Findings)
//	//
//	//return strings.Join(lines, "\n")
//}

//func (p *Page) formatHookInstallResult(result testhook.TextHookInstallResult) string {
//	lines := []string{
//		"Text hook install completed.",
//		"Engine: " + util.FirstNonEmpty(result.Engine, "Unknown"),
//		"Method: " + util.FirstNonEmpty(result.Method, "Unavailable"),
//	}
//
//	if strings.TrimSpace(result.PluginPath) != "" {
//		lines = append(lines, "Plugin Path: "+result.PluginPath)
//	}
//	if strings.TrimSpace(result.PluginsConfigPath) != "" {
//		lines = append(lines, "Plugins Config: "+result.PluginsConfigPath)
//	}
//	if strings.TrimSpace(result.OutputPath) != "" {
//		lines = append(lines, "Output: "+result.OutputPath)
//	}
//	if strings.TrimSpace(result.Compatibility.RiskLevel) != "" {
//		lines = append(lines, "Risk: "+result.Compatibility.RiskLevel)
//	}
//
//	lines = appendHookFindings(lines, result.Compatibility.Findings)
//
//	return strings.Join(lines, "\n")
//}

func appendHookFindings(lines []string, findings []string) []string {
	if len(findings) == 0 {
		return lines
	}

	lines = append(lines, "Findings:")
	for _, finding := range findings {
		finding = strings.TrimSpace(finding)
		if finding == "" {
			continue
		}
		lines = append(lines, "- "+finding)
	}

	return lines
}

func boolLabel(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func (p *Page) layoutGameArtwork(gtx layout.Context, cfg *vngame.Game) layout.Dimensions {
	const size = 44
	gtx.Constraints.Min.X = gtx.Dp(unit.Dp(size))
	gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(size))
	gtx.Constraints.Max.X = gtx.Dp(unit.Dp(size))
	gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(size))

	path := firstDrawableImagePath(cfg)
	if path != "" {
		view := p.gameImageView(path)
		view.Path = path
		return bareutils.RoundedSurface(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
			dims := view.Draw(gtx)
			if dims.Size.X == 0 || dims.Size.Y == 0 {
				return p.layoutGameArtworkFallback(gtx)
			}
			return dims
		})
	}
	return p.layoutGameArtworkFallback(gtx)
}

func (p *Page) layoutGameArtworkFallback(gtx layout.Context) layout.Dimensions {
	return bareutils.RoundedSurface(gtx, p.theme.Color.SurfaceAlt, unit.Dp(p.theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if p.iconify == nil {
				return layout.Dimensions{}
			}
			return p.iconify.LayoutWithSize(gtx, "mdi:controller-classic", unit.Dp(24), p.theme.Color.TextMuted)
		})
	})
}

func (p *Page) gameImageView(path string) *media.ImageView {
	if p.gameImageViews == nil {
		p.gameImageViews = map[string]*media.ImageView{}
	}
	if p.gameImageViews[path] == nil {
		p.gameImageViews[path] = &media.ImageView{Path: path}
	}
	return p.gameImageViews[path]
}

func firstDrawableImagePath(cfg *vngame.Game) string {
	if cfg == nil {
		return ""
	}
	for _, path := range []string{cfg.IconPath, cfg.ImagePath} {
		path = strings.TrimSpace(path)
		if isDrawableGameImage(path) && util.IsExistingFile(path) {
			return path
		}
	}
	return ""
}

func isDrawableGameImage(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".webp", ".gif":
		return true
	default:
		return false
	}
}

func savedGameMeta(cfg *vngame.Game) string {
	if cfg == nil {
		return ""
	}
	parts := make([]string, 0, 2)
	if strings.TrimSpace(cfg.EngineName) != "" {
		parts = append(parts, cfg.EngineName)
	}
	if cfg.RequiresSteam {
		parts = append(parts, "Steam")
	} else if strings.TrimSpace(string(cfg.Runner)) != "" {
		parts = append(parts, string(cfg.Runner))
	}
	return strings.Join(parts, " · ")
}

func (p *Page) draftFromForm(base *vngame.Game) *vngame.Game {
	cfg := cloneGameConfig(base)
	if cfg == nil {
		cfg = &vngame.Game{}
	}
	cfg.Name = strings.TrimSpace(p.nameEditor.Text())
	cfg.GamePath = strings.TrimSpace(p.pathEditor.Text())
	cfg.IconPath = strings.TrimSpace(p.iconPathEditor.Text())
	cfg.ImagePath = strings.TrimSpace(p.imagePathEditor.Text())
	cfg.SteamAppID = strings.TrimSpace(p.steamAppIDEditor.Text())
	cfg.VirtualDesktop = virtualDesktopValueFromLabel(p.selectedDesktop)
	cfg.RequiresSteam = p.requiresSteam.Value
	if runner := runnerFromLabel(p.selectedRunner); runner != "" {
		cfg.Runner = runner
	}
	return cfg
}

func applyGameOverrides(cfg, overrides *vngame.Game) {
	if cfg == nil || overrides == nil {
		return
	}
	if strings.TrimSpace(overrides.Name) != "" {
		cfg.Name = strings.TrimSpace(overrides.Name)
	}
	if strings.TrimSpace(overrides.IconPath) != "" {
		cfg.IconPath = strings.TrimSpace(overrides.IconPath)
	}
	if strings.TrimSpace(overrides.ImagePath) != "" {
		cfg.ImagePath = strings.TrimSpace(overrides.ImagePath)
	}
	if strings.TrimSpace(overrides.SteamAppID) != "" {
		cfg.SteamAppID = strings.TrimSpace(overrides.SteamAppID)
	}
	cfg.VirtualDesktop = strings.TrimSpace(overrides.VirtualDesktop)
	cfg.RequiresSteam = overrides.RequiresSteam
	if overrides.Runner != "" {
		cfg.Runner = overrides.Runner
	}
}

func cloneGameConfig(cfg *vngame.Game) *vngame.Game {
	if cfg == nil {
		return nil
	}
	copy := *cfg
	if cfg.EnvVars != nil {
		copy.EnvVars = append([]vngame.EnvVar(nil), cfg.EnvVars...)
	}
	return &copy
}

func runnerFromLabel(label string) vngame.RunnerType {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "wine":
		return vngame.RunnerWine
	case "proton":
		return vngame.RunnerProton
	case "steam":
		return vngame.RunnerSteam
	default:
		return ""
	}
}

func virtualDesktopLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Default"
	}
	switch strings.ToLower(value) {
	case "off", "false", "none", "disabled", "disable", "0":
		return "Off"
	default:
		return value
	}
}

func virtualDesktopValueFromLabel(label string) string {
	label = strings.TrimSpace(label)
	switch strings.ToLower(label) {
	case "", "default":
		return ""
	case "off":
		return "off"
	default:
		return label
	}
}

func gamePreviewText(cfg *vngame.Game) string {
	if cfg == nil {
		return ""
	}
	return strings.Join([]string{
		"Name: " + cfg.Name,
		"Resolved Path: " + cfg.GamePath,
		"Executable: " + cfg.Executable,
		"Working Dir: " + cfg.WorkingDir,
		"Runner: " + string(cfg.Runner),
		"Virtual Desktop: " + util.FirstNonEmpty(cfg.VirtualDesktop, "Default"),
		"Engine: " + util.FirstNonEmpty(cfg.EngineName, "Unknown"),
		"Icon: " + util.FirstNonEmpty(cfg.IconPath, "Unavailable"),
		"Image: " + util.FirstNonEmpty(cfg.ImagePath, "Unavailable"),
	}, "\n")
}

func renameGameFlashcards(oldName, newName string) error {
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	if oldName == "" || newName == "" || strings.EqualFold(oldName, newName) {
		return nil
	}
	return flashcards.RenameGameFlashcards(oldName, newName)
}

func cleanupOldGameConfig(oldName string, cfg *vngame.Game) {
	oldName = strings.TrimSpace(oldName)
	if oldName == "" || cfg == nil || strings.EqualFold(oldName, strings.TrimSpace(cfg.Name)) {
		return
	}
	newPath := cleanAbsPath(gameConfig.DefaultGameConfigPath(cfg))
	for _, dir := range gameConfigSearchDirs() {
		oldPath := cleanAbsPath(filepath.Join(dir, util.SanitizeName(oldName)+".json"))
		if oldPath == "" || oldPath == newPath {
			continue
		}
		_ = os.Remove(oldPath)
	}
}

func cleanupDuplicateGameConfigs(cfg *vngame.Game) {
	if cfg == nil {
		return
	}
	keepPath := cleanAbsPath(gameConfig.DefaultGameConfigPath(cfg))
	for _, dir := range gameConfigSearchDirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
				continue
			}
			path := cleanAbsPath(filepath.Join(dir, entry.Name()))
			if path == "" || path == keepPath {
				continue
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			var existing vngame.Game
			if err := json.Unmarshal(raw, &existing); err != nil {
				continue
			}
			if sameGameTarget(&existing, cfg) && !strings.EqualFold(strings.TrimSpace(existing.Name), strings.TrimSpace(cfg.Name)) {
				_ = os.Remove(path)
			}
		}
	}
}

func sameGameTarget(a, b *vngame.Game) bool {
	if a == nil || b == nil {
		return false
	}
	aPath := normalizedGameTarget(a)
	bPath := normalizedGameTarget(b)
	return aPath != "" && bPath != "" && aPath == bPath
}

func normalizedGameTarget(cfg *vngame.Game) string {
	if cfg == nil {
		return ""
	}
	return cleanAbsPath(util.FirstNonEmpty(cfg.Executable, cfg.GamePath))
}

func cleanAbsPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return filepath.Clean(path)
}

func gameConfigSearchDirs() []string {
	return []string{
		filepath.Join(gameConfig.ConfigBaseDir(), "games"),
		filepath.Join(util.ConfigBaseDir(), "games"),
	}
}

func (p *Page) showError(title, body string) {
	if p.OnError != nil {
		p.OnError(title, body)
	}
}

func (p *Page) filteredConfigs() []*vngame.Game {
	query := strings.TrimSpace(strings.ToLower(p.gameSearchEditor.Text()))
	if query == "" {
		return p.configs
	}
	filtered := make([]*vngame.Game, 0, len(p.configs))
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
	if p.installing {
		return
	}
	p.currentConfigName = ""
	p.loadedConfigName = ""
	p.draftConfig = nil
	p.pathEditor.SetText("")
	p.steamAppIDEditor.SetText("")
	p.iconPathEditor.SetText("")
	p.imagePathEditor.SetText("")
	p.requiresSteam.Value = false
	p.selectedRunner = "Auto"
	p.selectedDesktop = "Default"
}

func (p *Page) browserButtonLabel() string {
	if p.browserModal.Open {
		return "Browsing Files"
	}
	return "Browse Files"
}

func (p *Page) browserButtonIcon() string {
	if p.browserModal.Open {
		return "mdi:folder-open-outline"
	}
	return "mdi:folder-search-outline"
}

func (p *Page) installActionLabel(idle string) string {
	if p.installing {
		return "Installing..."
	}
	return idle
}

func (p *Page) installActionIcon(idle string) string {
	if p.installing {
		return "mdi:progress-clock"
	}
	return idle
}

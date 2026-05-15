package gamepage

import (
	"context"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/input"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/modal"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/pages/fileexplorer"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
	"github.com/DarlingGoose/ymn/pkg/yomuna/backend"
)

type AddGameUI struct {
	th      *material.Theme
	theme   *theme.Client
	backend backend.Backend

	pathInput      *input.TextInput
	installerInput *input.TextInput

	browseButton          *components.IconButton
	installerBrowseButton *components.IconButton
	createButton          *components.IconButton
	installButton         *components.IconButton
	clearButton           *components.IconButton
	installHook           widget.Bool
	installing            bool
	status                string
	resultChannel         chan addGameResult

	filePickerModal *modal.Modal
	filePicker      *fileexplorer.FileExplorer
}

type addGameResult struct {
	game *game.Game
	err  error
}

func NewAddGameUI(th *material.Theme, tc *theme.Client, b backend.Backend) *AddGameUI {
	if th == nil {
		th = material.NewTheme()
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	folderIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:folder-open")
	plusIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:plus")
	clearIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:x")
	downloadIcon, _ := iconify.DefaultIconify.Icon(context.Background(), "lucide:download")

	ui := &AddGameUI{
		th:                    th,
		theme:                 tc,
		backend:               b,
		pathInput:             input.NewPathInput("Game path", "/path/to/game-folder-or-exe").WithMaterialTheme(th).WithThemeClient(tc),
		installerInput:        input.NewPathInput("Installer path", "/path/to/setup.exe").WithMaterialTheme(th).WithThemeClient(tc),
		browseButton:          components.NewIconButton("Browse", nil, folderIcon).WithThemeClient(tc),
		installerBrowseButton: components.NewIconButton("Browse", nil, folderIcon).WithThemeClient(tc),
		createButton:          components.NewIconButton("Create Game", nil, plusIcon).WithThemeClient(tc),
		installButton:         components.NewIconButton("Run Installer", nil, downloadIcon).WithThemeClient(tc),
		clearButton:           components.NewIconButton("Clear", nil, clearIcon).WithThemeClient(tc),
		status:                "Select a game folder, executable, or launcher path.",
		resultChannel:         make(chan addGameResult, 1),
	}
	ui.installHook.Value = true
	ui.configureButton(ui.browseButton)
	ui.configureButton(ui.installerBrowseButton)
	ui.configureButton(ui.clearButton)
	ui.configureButton(ui.createButton)
	ui.configureButton(ui.installButton)
	ui.createButton.MinWidth = unit.Dp(148)
	ui.installButton.MinWidth = unit.Dp(158)
	ui.createButton.CollapseTextBelow = unit.Dp(180)
	ui.installButton.CollapseTextBelow = unit.Dp(190)
	return ui
}

func (ui *AddGameUI) configureButton(btn *components.IconButton) {
	if btn == nil {
		return
	}
	btn.FillWidth = false
	btn.TextCollapseMode = components.TextCollapseNever
	btn.CollapseTextBelow = unit.Dp(160)
	btn.MinWidth = unit.Dp(96)
	btn.Height = unit.Dp(44)
	btn.Radius = unit.Dp(10)
}

func (ui *AddGameUI) Layout(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if ui == nil {
		return layout.Dimensions{}
	}
	ui.consumeResults(gtx)
	if ui.installing {
		gtx.Execute(op.InvalidateCmd{})
	}

	dims := layout.UniformInset(unit.Dp(16)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return ui.layoutContentContainer(gtx, ui.layoutPage)
	})
	ui.layoutFilePickerOverlay(gtx, layer)
	return dims
}

func (ui *AddGameUI) consumeResults(gtx layout.Context) {
	for {
		select {
		case result := <-ui.resultChannel:
			ui.installing = false
			ui.createButton.SetLoading(false)
			ui.installButton.SetLoading(false)
			if result.err != nil {
				ui.status = "Add game failed: " + result.err.Error()
			} else if result.game != nil && strings.TrimSpace(result.game.Name) != "" {
				ui.status = "Added " + strings.TrimSpace(result.game.Name)
			} else {
				ui.status = "Game added"
			}
			gtx.Execute(op.InvalidateCmd{})
		default:
			return
		}
	}
}

func (ui *AddGameUI) layoutContentContainer(gtx layout.Context, content layout.Widget) layout.Dimensions {
	if content == nil {
		return layout.Dimensions{}
	}
	maxWidth := gtx.Dp(unit.Dp(960))
	available := gtx.Constraints.Max.X
	if available <= 0 || available <= maxWidth {
		return content(gtx)
	}
	side := (available - maxWidth) / 2
	return layout.Inset{
		Left:  unit.Dp(float32(side) / gtx.Metric.PxPerDp),
		Right: unit.Dp(float32(side) / gtx.Metric.PxPerDp),
	}.Layout(gtx, content)
}

func (ui *AddGameUI) layoutPage(gtx layout.Context) layout.Dimensions {
	ct := ui.theme.GetCurrentColorToken()
	return utils.SurfaceOutlined(gtx, ct.SurfaceNRGBA(), unit.Dp(8), utils.SurfaceBorder{Color: ct.BorderNRGBA(), Width: unit.Dp(1)}, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleH2, theme.ThemeColorTextPrimary, "Add Game")
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, "Create a game config using vntext installgame.")
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
				layout.Rigid(ui.layoutPathField),
				layout.Rigid(layout.Spacer{Height: unit.Dp(14)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return ui.layoutInputRow(gtx, "Installer path", ui.installerInput, ui.installerBrowseButton)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					check := material.CheckBox(ui.th, &ui.installHook, "Install text hook when supported")
					check.Color = ui.theme.GetCurrentColorToken().PrimaryNRGBA()
					return check.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(18)}.Layout),
				layout.Rigid(ui.layoutActions),
				layout.Rigid(layout.Spacer{Height: unit.Dp(18)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleBodySmall, theme.ThemeColorTextMuted, ui.status)
				}),
			)
		})
	})
}

func (ui *AddGameUI) layoutPathField(gtx layout.Context) layout.Dimensions {
	return ui.layoutInputRow(gtx, "Game path", ui.pathInput, ui.browseButton)
}

func (ui *AddGameUI) layoutInputRow(gtx layout.Context, label string, in *input.TextInput, browse *components.IconButton) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(gtx, ui.th, ui.theme, theme.TextRoleLabel, theme.ThemeColorTextPrimary, label)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.End}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if in == nil {
						return layout.Dimensions{}
					}
					oldLabel := in.Label
					in.Label = ""
					dims := in.Layout(gtx)
					in.Label = oldLabel
					return dims
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if browse.Clicked(gtx) {
						ui.openFilePicker(in)
						gtx.Execute(op.InvalidateCmd{})
					}
					return browse.Layout(gtx)
				}),
			)
		}),
	)
}

func (ui *AddGameUI) layoutActions(gtx layout.Context) layout.Dimensions {
	if ui.createButton != nil {
		ui.createButton.Disabled = ui.installing || strings.TrimSpace(ui.pathInput.Text()) == ""
		ui.createButton.SetLoading(ui.installing)
	}
	if ui.installButton != nil {
		ui.installButton.Disabled = ui.installing ||
			strings.TrimSpace(ui.pathInput.Text()) == "" ||
			strings.TrimSpace(ui.installerInput.Text()) == ""
		ui.installButton.SetLoading(ui.installing)
	}
	if ui.clearButton != nil {
		ui.clearButton.Disabled = ui.installing ||
			(strings.TrimSpace(ui.pathInput.Text()) == "" && strings.TrimSpace(ui.installerInput.Text()) == "")
	}

	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.createButton.Clicked(gtx) {
				ui.installGame(gtx)
			}
			return ui.createButton.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.installButton.Clicked(gtx) {
				ui.installGameWithInstaller(gtx)
			}
			return ui.installButton.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if ui.clearButton.Clicked(gtx) {
				if ui.pathInput != nil {
					ui.pathInput.SetText("")
				}
				if ui.installerInput != nil {
					ui.installerInput.SetText("")
				}
				ui.status = "Select a game folder, executable, or launcher path."
				gtx.Execute(op.InvalidateCmd{})
			}
			return ui.clearButton.Layout(gtx)
		}),
	)
}

func (ui *AddGameUI) installGame(gtx layout.Context) {
	if ui == nil || ui.backend == nil || ui.pathInput == nil || ui.installing {
		return
	}
	path := strings.TrimSpace(ui.pathInput.Text())
	if path == "" {
		ui.status = "Game path is required"
		gtx.Execute(op.InvalidateCmd{})
		return
	}

	ui.installing = true
	ui.status = "Creating game config..."
	gtx.Execute(op.InvalidateCmd{})

	installHook := ui.installHook.Value
	go func() {
		g, err := ui.backend.InstallGameConfig(context.Background(), path, installHook)
		ui.resultChannel <- addGameResult{game: g, err: err}
	}()
}

func (ui *AddGameUI) installGameWithInstaller(gtx layout.Context) {
	if ui == nil || ui.backend == nil || ui.pathInput == nil || ui.installerInput == nil || ui.installing {
		return
	}
	gamePath := strings.TrimSpace(ui.pathInput.Text())
	installerPath := strings.TrimSpace(ui.installerInput.Text())
	if gamePath == "" {
		ui.status = "Game path is required"
		gtx.Execute(op.InvalidateCmd{})
		return
	}
	if installerPath == "" {
		ui.status = "Installer path is required"
		gtx.Execute(op.InvalidateCmd{})
		return
	}

	ui.installing = true
	ui.status = "Running installer..."
	gtx.Execute(op.InvalidateCmd{})

	installHook := ui.installHook.Value
	go func() {
		g, err := ui.backend.InstallGameWithInstaller(context.Background(), installerPath, gamePath, installHook)
		ui.resultChannel <- addGameResult{game: g, err: err}
	}()
}

func (ui *AddGameUI) openFilePicker(target *input.TextInput) {
	if ui == nil || target == nil {
		return
	}
	startDir := filePickerStartDir(target.Text())
	explorer := fileexplorer.NewFileExplorer(startDir, media.DefaultRegistry, ui.theme).
		WithMaterialTheme(ui.th).
		WithThemeClient(ui.theme)
	explorer.SelectButtonText = "Use Path"
	explorer.OnChoose = func(path string) {
		target.SetText(path)
		if ui.filePickerModal != nil {
			ui.filePickerModal.Dismiss()
		}
	}
	ui.filePicker = explorer
	ui.filePickerModal = modal.New("add-game-file-picker", "Select File or Folder", func(gtx layout.Context) layout.Dimensions {
		if ui.filePicker == nil {
			return layout.Dimensions{}
		}
		return ui.filePicker.Layout(gtx)
	}).WithThemeClient(ui.theme).WithMaterialTheme(ui.th).WithSize(unit.Dp(1600), unit.Dp(0))
	ui.filePickerModal.MaxWidth = 0
	ui.filePickerModal.MinHeight = unit.Dp(720)
	ui.filePickerModal.Margin = unit.Dp(16)
	ui.filePickerModal.Padding = unit.Dp(12)
	ui.filePickerModal.Radius = unit.Dp(12)
	ui.filePickerModal.Open()
}

func (ui *AddGameUI) layoutFilePickerOverlay(gtx layout.Context, layer *overlay.Overlay) {
	if ui == nil || layer == nil || ui.filePickerModal == nil || !ui.filePickerModal.Visible {
		return
	}
	layer.Add(gtx, ui.filePickerModal)
}

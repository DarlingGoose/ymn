package fileexplorer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/layouts/split"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/dropdowns"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/input"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/panel"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
)

type FileExplorer struct {
	tc    *theme.Client
	Theme *material.Theme

	Overlay overlay.Overlay
	Split   *split.SplitV

	Root       string
	CurrentDir string
	Selected   string

	CommonPlaces []CommonPlace

	Preview *media.View

	Search       *input.TextInput
	PathInput    *input.TextInput
	SortDropdown *dropdowns.Dropdown

	List layout.List

	entries []entry
	err     error

	rowClicks   map[string]*widget.Clickable
	placeClicks map[string]*widget.Clickable

	upButton      widget.Clickable
	refreshButton widget.Clickable
	homeButton    widget.Clickable

	selectButton  widget.Clickable
	pathGoButton  widget.Clickable
	sortDirButton widget.Clickable

	// OnSelect fires when a file row is clicked and previewed.
	OnSelect func(path string)

	// OnChoose fires when the explicit select button is clicked.
	// If a file is selected, it emits Selected. Otherwise it emits CurrentDir.
	OnChoose func(path string)

	ShowSelectButton bool
	SelectButtonText string

	SortBy   SortBy
	SortDesc bool

	SidebarWidth unit.Dp
	SidebarGap   unit.Dp

	Gap       unit.Dp
	Padding   unit.Dp
	RowHeight unit.Dp

	showHidden bool
}

func NewFileExplorer(startDir string, registry *media.Registry, tc *theme.Client) *FileExplorer {
	if startDir == "" {
		if wd, err := os.Getwd(); err == nil {
			startDir = wd
		} else {
			startDir = "."
		}
	}

	abs, err := filepath.Abs(startDir)
	if err == nil {
		startDir = abs
	}

	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	sortDropdown := dropdowns.NewDropdown([]dropdowns.DropdownItem{
		{Label: "Name", Value: "name"},
		{Label: "Size", Value: "size"},
		{Label: "Modified", Value: "modified"},
		{Label: "Kind", Value: "kind"},
	}).
		WithThemeClient(tc).
		WithCompact()

	p := &FileExplorer{
		tc:    tc,
		Theme: material.NewTheme(),

		Overlay: overlay.Overlay{},

		Split: &split.SplitV{
			Ratio:    -0.25,
			Bar:      unit.Dp(8),
			MinRatio: -0.85,
			MaxRatio: 0.65,
		},

		Root:       startDir,
		CurrentDir: startDir,

		CommonPlaces: DefaultCommonPlaces(startDir),

		Preview: media.NewView(registry),

		Search: input.NewSearchInput("Search files...").
			WithThemeClient(tc),

		PathInput: input.NewPathInput("", startDir).
			WithThemeClient(tc),

		SortDropdown: sortDropdown,

		List: layout.List{Axis: layout.Vertical},

		rowClicks:   make(map[string]*widget.Clickable),
		placeClicks: make(map[string]*widget.Clickable),

		ShowSelectButton: true,
		SelectButtonText: "Select",

		SortBy:   SortByName,
		SortDesc: false,

		SidebarWidth: unit.Dp(190),
		SidebarGap:   unit.Dp(12),

		Gap:       unit.Dp(16),
		Padding:   unit.Dp(16),
		RowHeight: unit.Dp(40),
	}

	p.Search.OnChange = func(string) {
		p.reload()
	}

	p.PathInput.SetText(startDir)
	p.PathInput.OnSubmit = func(path string) {
		p.GoToPath(path)
	}

	p.SortDropdown.SelectItemEvent(func(item dropdowns.DropdownItem, valid bool) {
		if !valid {
			return
		}

		p.SortBy = sortByFromValue(item.Value)
		p.reload()
	})

	p.reload()

	return p
}

func (p *FileExplorer) WithThemeClient(tc *theme.Client) *FileExplorer {
	if p == nil {
		return p
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	p.tc = tc

	if p.Search != nil {
		p.Search.WithThemeClient(tc)
	}
	if p.PathInput != nil {
		p.PathInput.WithThemeClient(tc)
	}
	if p.SortDropdown != nil {
		p.SortDropdown.WithThemeClient(tc)
	}

	return p
}

func (p *FileExplorer) WithMaterialTheme(th *material.Theme) *FileExplorer {
	if p == nil {
		return p
	}
	if th != nil {
		p.Theme = th
	}
	if p.Search != nil {
		p.Search.WithMaterialTheme(th)
	}
	if p.PathInput != nil {
		p.PathInput.WithMaterialTheme(th)
	}

	return p
}

func (p *FileExplorer) WithCommonPlaces(places ...CommonPlace) *FileExplorer {
	if p == nil {
		return p
	}

	p.CommonPlaces = dedupePlaces(places)
	p.placeClicks = make(map[string]*widget.Clickable, len(p.CommonPlaces))

	return p
}

func (p *FileExplorer) AddCommonPlace(label, path, icon string) *FileExplorer {
	if p == nil {
		return p
	}

	if icon == "" {
		icon = "lucide:folder"
	}

	path = expandHome(path)
	abs, err := filepath.Abs(path)
	if err == nil {
		path = abs
	}

	p.CommonPlaces = dedupePlaces(append(p.CommonPlaces, CommonPlace{
		Label: label,
		Path:  path,
		Icon:  icon,
	}))

	return p
}

func (p *FileExplorer) Select(path string) {
	if p == nil || path == "" {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		p.err = err
		return
	}

	if info.IsDir() {
		p.CurrentDir = path
		p.Selected = ""

		if p.PathInput != nil {
			p.PathInput.SetText(path)
		}

		if p.Preview != nil {
			_ = p.Preview.Close()
		}

		p.reload()
		return
	}

	p.Selected = path
	p.err = nil

	if p.PathInput != nil {
		p.PathInput.SetText(path)
	}

	if p.Preview != nil && supportsMediaPreview(path) {
		_ = p.Preview.LoadPath(context.Background(), path)
	}

	if p.OnSelect != nil {
		p.OnSelect(path)
	}
}

func (p *FileExplorer) layoutPreview(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutPreviewHeader(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if p.Selected == "" {
				return p.layoutMessage(gtx, "Select a file to preview")
			}

			return p.layoutSelectedPreview(gtx)
		}),
	)
}

func (p *FileExplorer) layoutSelectedPreview(gtx layout.Context) layout.Dimensions {
	if p == nil || p.Selected == "" {
		return layout.Dimensions{}
	}

	if supportsMediaPreview(p.Selected) && p.Preview != nil {
		return p.Preview.Layout(gtx)
	}

	return p.layoutMetadataPreview(gtx, p.Selected)
}

func (p *FileExplorer) layoutMetadataPreview(gtx layout.Context, path string) layout.Dimensions {
	info, err := os.Stat(path)
	if err != nil {
		return p.layoutMessage(gtx, err.Error())
	}

	rows := []metadataRow{
		{Label: "Name", Value: info.Name()},
		{Label: "Type", Value: detectFileCategory(path, info)},
		{Label: "Path", Value: path},
		{Label: "Size", Value: formatSize(info.Size())},
		{Label: "Modified", Value: info.ModTime().Format("2006-01-02 15:04:05")},
		{Label: "Permissions", Value: info.Mode().String()},
	}

	if ext := filepath.Ext(path); ext != "" {
		rows = append(rows, metadataRow{Label: "Extension", Value: strings.ToLower(ext)})
	}

	if info.Mode().IsRegular() && info.Mode()&0o111 != 0 {
		rows = append(rows, metadataRow{Label: "Executable", Value: "Yes"})
	}

	children := make([]layout.FlexChild, 0, len(rows)*2+4)

	children = append(children,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				p.Theme,
				p.tc,
				theme.TextRoleH4,
				theme.ThemeColorTextPrimary,
				"File details",
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
	)

	for i := range rows {
		row := rows[i]

		if i > 0 {
			children = append(children, layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout))
		}

		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutMetadataRow(gtx, row.Label, row.Value)
		}))
	}
	if isArchive(path) {
		rows = append(rows, metadataRow{
			Label: "Archive",
			Value: "Preview not expanded yet",
		})
	}
	if isLikelyTextFile(path) {
		children = append(children,
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return p.layoutTextSnippet(gtx, path)
			}),
		)
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, children...)
}

type metadataRow struct {
	Label string
	Value string
}

func (p *FileExplorer) layoutMetadataRow(gtx layout.Context, label, value string) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Start,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Dp(unit.Dp(110))
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(110))

			return theme.ThemedLabel(
				gtx,
				p.Theme,
				p.tc,
				theme.TextRoleCaption,
				theme.ThemeColorTextMuted,
				label,
			)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				p.Theme,
				p.tc,
				theme.TextRoleBody,
				theme.ThemeColorTextPrimary,
				value,
			)
		}),
	)
}

func (p *FileExplorer) layoutTextSnippet(gtx layout.Context, path string) layout.Dimensions {
	data, err := os.ReadFile(path)
	if err != nil {
		return p.layoutMetadataRow(gtx, "Preview", err.Error())
	}

	const max = 4096
	if len(data) > max {
		data = data[:max]
	}

	text := string(data)
	text = strings.TrimSpace(text)

	if text == "" {
		text = "(empty file)"
	}

	return panel.NewBackgroundPanel(p.tc).
		WithRole(panel.BackgroundRoleSurfaceAlt).
		WithRadius(unit.Dp(12)).
		WithInset(layout.UniformInset(unit.Dp(12))).
		WithFillMax(false).
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(
						gtx,
						p.Theme,
						p.tc,
						theme.TextRoleLabel,
						theme.ThemeColorTextMuted,
						"Preview",
					)
				}),
				layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(
						gtx,
						p.Theme,
						p.tc,
						theme.TextRoleCode,
						theme.ThemeColorTextPrimary,
						text,
					)
				}),
			)
		})
}
func (p *FileExplorer) CurrentPath() string {
	if p == nil {
		return ""
	}
	if p.Selected != "" {
		return p.Selected
	}
	return p.CurrentDir
}

func (p *FileExplorer) Choose() {
	if p == nil {
		return
	}

	path := p.CurrentPath()
	if path == "" {
		return
	}

	if p.OnChoose != nil {
		p.OnChoose(path)
	}
}

func (p *FileExplorer) GoToPath(path string) {
	if p == nil {
		return
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return
	}

	expanded := expandHome(path)

	abs, err := filepath.Abs(expanded)
	if err != nil {
		p.err = err
		return
	}

	info, err := os.Stat(abs)
	if err != nil {
		p.err = err
		return
	}

	if info.IsDir() {
		p.CurrentDir = path
		p.Selected = ""

		if p.PathInput != nil {
			p.PathInput.SetText(path)
		}

		if p.Preview != nil {
			_ = p.Preview.Close()
		}

		p.reload()
		return
	}

	p.CurrentDir = filepath.Dir(abs)
	p.err = nil

	if p.PathInput != nil {
		p.PathInput.SetText(abs)
	}

	p.reload()
	p.Select(abs)
}

func (p *FileExplorer) Layout(gtx layout.Context) layout.Dimensions {
	if p == nil {
		return layout.Dimensions{}
	}

	if p.Theme == nil {
		p.Theme = material.NewTheme()
	}
	if p.tc == nil {
		p.tc = theme.DefaultThemeClient
	}
	if p.Split == nil {
		p.Split = &split.SplitV{
			Ratio:    -0.25,
			Bar:      unit.Dp(8),
			MinRatio: -0.85,
			MaxRatio: 0.65,
		}
	}

	// Only show preview pane when a file is selected.
	p.Split.HideRight = p.Selected == ""

	if p.tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	p.update(gtx)

	return panel.NewBackgroundPanel(p.tc).
		WithFillMax(true).
		Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return p.Overlay.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(p.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					if gtx.Constraints.Max.X < gtx.Dp(unit.Dp(820)) {
						children := []layout.FlexChild{
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return p.layoutSidebarPanel(gtx)
							}),
							layout.Rigid(layout.Spacer{Height: p.Gap}.Layout),
							layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
								return p.layoutExplorerPanel(gtx)
							}),
						}

						if p.Selected != "" {
							children = append(children,
								layout.Rigid(layout.Spacer{Height: p.Gap}.Layout),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return p.layoutPreviewPanel(gtx)
								}),
							)
						}

						return layout.Flex{
							Axis: layout.Vertical,
						}.Layout(gtx, children...)
					}

					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							w := gtx.Dp(p.SidebarWidth)
							gtx.Constraints.Min.X = w
							gtx.Constraints.Max.X = w
							return p.layoutSidebarPanel(gtx)
						}),
						layout.Rigid(layout.Spacer{Width: p.SidebarGap}.Layout),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return p.Split.Layout(
								gtx,
								func(gtx layout.Context) layout.Dimensions {
									return p.layoutExplorerPanel(gtx)
								},
								func(gtx layout.Context) layout.Dimensions {
									return p.layoutPreviewPanel(gtx)
								},
							)
						}),
					)
				})
			})
		})
}

func (p *FileExplorer) update(gtx layout.Context) {
	for p.upButton.Clicked(gtx) {
		parent := filepath.Dir(p.CurrentDir)
		if parent != "" && parent != p.CurrentDir {
			p.CurrentDir = parent
			p.Selected = ""

			if p.PathInput != nil {
				p.PathInput.SetText(parent)
			}

			p.reload()
		}
	}

	for p.refreshButton.Clicked(gtx) {
		p.reload()
	}

	for p.homeButton.Clicked(gtx) {
		if p.Root != "" {
			p.CurrentDir = p.Root
			p.Selected = ""

			if p.PathInput != nil {
				p.PathInput.SetText(p.Root)
			}

			p.reload()
		}
	}

	for _, place := range p.CommonPlaces {
		click := p.placeClickFor(place.Path)
		for click.Clicked(gtx) {
			p.GoToPath(place.Path)
		}
	}

	for p.pathGoButton.Clicked(gtx) {
		if p.PathInput != nil {
			p.GoToPath(p.PathInput.Text())
		}
	}

	for p.selectButton.Clicked(gtx) {
		p.Choose()
	}

	for p.sortDirButton.Clicked(gtx) {
		p.SortDesc = !p.SortDesc
		p.reload()
	}

	for _, e := range p.entries {
		click := p.clickFor(e.Path)
		for click.Clicked(gtx) {
			p.Select(e.Path)
		}
	}
}

func (p *FileExplorer) layoutSidebarPanel(gtx layout.Context) layout.Dimensions {
	return panel.NewBackgroundPanel(p.tc).
		WithRole(panel.BackgroundRoleSurface).
		WithRadius(unit.Dp(16)).
		WithInset(layout.UniformInset(unit.Dp(10))).
		WithFillMax(true).
		Layout(gtx, p.layoutSidebar)
}

func (p *FileExplorer) layoutSidebar(gtx layout.Context) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(p.CommonPlaces)*2+2)

	children = append(children,
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
	)

	for i := range p.CommonPlaces {
		place := p.CommonPlaces[i]

		if i > 0 {
			children = append(children, layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout))
		}

		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutPlace(gtx, place)
		}))
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, children...)
}

func (p *FileExplorer) layoutPlace(gtx layout.Context, place CommonPlace) layout.Dimensions {
	click := p.placeClickFor(place.Path)

	selected := samePath(p.CurrentDir, place.Path) || samePath(p.Selected, place.Path)
	hovered := click.Hovered()

	if hovered {
		gtx.Execute(op.InvalidateCmd{})
	}

	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		content := func(gtx layout.Context) layout.Dimensions {
			colorRole := theme.ThemeColorTextPrimary

			tokens := p.tc.GetCurrentColorToken()
			iconColor := tokens.TextSecondaryNRGBA()

			if selected {
				colorRole = theme.ThemeColorOnPrimary
				iconColor = tokens.OnPrimaryNRGBA()
			}

			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					icon := place.Icon
					if icon == "" {
						icon = "lucide:folder"
					}

					return iconify.DefaultIconify.Layout(
						gtx,
						icon,
						unit.Dp(17),
						iconColor,
					)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(
						gtx,
						p.Theme,
						p.tc,
						theme.TextRoleBody,
						colorRole,
						place.Label,
					)
				}),
			)
		}

		inset := layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
			Left:   unit.Dp(10),
			Right:  unit.Dp(10),
		}

		switch {
		case selected:
			return panel.NewBackgroundPanel(p.tc).
				WithRole(panel.BackgroundRolePrimary).
				WithRadius(unit.Dp(10)).
				WithInset(inset).
				WithFillMax(false).
				Layout(gtx, content)

		case hovered:
			return panel.NewBackgroundPanel(p.tc).
				WithRole(panel.BackgroundRoleSurfaceAlt).
				WithRadius(unit.Dp(10)).
				WithInset(inset).
				WithFillMax(false).
				Layout(gtx, content)

		default:
			return inset.Layout(gtx, content)
		}
	})
}

func (p *FileExplorer) layoutExplorerPanel(gtx layout.Context) layout.Dimensions {
	return panel.NewBackgroundPanel(p.tc).
		WithRole(panel.BackgroundRoleSurface).
		WithRadius(unit.Dp(16)).
		WithInset(layout.UniformInset(unit.Dp(12))).
		WithFillMax(true).
		Layout(gtx, p.layoutExplorer)
}

func (p *FileExplorer) layoutPreviewPanel(gtx layout.Context) layout.Dimensions {
	return panel.NewBackgroundPanel(p.tc).
		WithRole(panel.BackgroundRoleSurface).
		WithRadius(unit.Dp(16)).
		WithInset(layout.UniformInset(unit.Dp(12))).
		WithFillMax(true).
		Layout(gtx, p.layoutPreview)
}

func (p *FileExplorer) layoutExplorer(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutHeader(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		//layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		//	return p.Search.Layout(gtx)
		//}),
		//layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if p.err != nil {
				return p.layoutMessage(gtx, p.err.Error())
			}

			if len(p.entries) == 0 {
				return p.layoutMessage(gtx, "No files found")
			}

			return p.List.Layout(gtx, len(p.entries), func(gtx layout.Context, index int) layout.Dimensions {
				return p.layoutEntry(gtx, p.entries[index])
			})
		}),
	)
}

func (p *FileExplorer) layoutHeader(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.iconButton(gtx, &p.homeButton, "lucide:home")
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.iconButton(gtx, &p.upButton, "lucide:arrow-up")
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.iconButton(gtx, &p.refreshButton, "lucide:refresh-cw")
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if p.PathInput == nil {
						return theme.ThemedLabel(
							gtx,
							p.Theme,
							p.tc,
							theme.TextRoleCaption,
							theme.ThemeColorTextMuted,
							p.CurrentDir,
						)
					}
					return p.PathInput.Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.textButton(gtx, &p.pathGoButton, "Go")
				}),
			)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(10)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(
						gtx,
						p.Theme,
						p.tc,
						theme.TextRoleCaption,
						theme.ThemeColorTextMuted,
						"Sort",
					)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if p.SortDropdown == nil {
						return layout.Dimensions{}
					}

					p.SortDropdown.Width = unit.Dp(120)
					p.SortDropdown.Height = unit.Dp(24)
					p.SortDropdown.ItemHeight = unit.Dp(22)
					p.SortDropdown.MaxMenuHeight = unit.Dp(190)
					p.SortDropdown.Inset = unit.Dp(8)
					p.SortDropdown.ArrowIconSize = unit.Dp(14)
					p.SortDropdown.WithRole(theme.TextRoleCaption)

					return p.SortDropdown.Layout(gtx, &p.Overlay)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					icon := "lucide:arrow-up-a-z"
					text := "Asc"
					if p.SortDesc {
						icon = "lucide:arrow-down-z-a"
						text = "Desc"
					}

					return p.iconTextButton(gtx, &p.sortDirButton, icon, text)
				}),
				layout.Flexed(1, layout.Spacer{}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !p.ShowSelectButton {
						return layout.Dimensions{}
					}

					label := p.SelectButtonText
					if label == "" {
						label = "Select"
					}

					return p.iconTextButton(gtx, &p.selectButton, "lucide:check", label)
				}),
			)
		}),
	)
}

func (p *FileExplorer) layoutEntry(gtx layout.Context, e entry) layout.Dimensions {
	click := p.clickFor(e.Path)

	icon := "lucide:file"
	if e.IsDir {
		icon = "lucide:folder"
	}

	selected := p.Selected == e.Path
	hovered := click.Hovered()

	if hovered {
		gtx.Execute(op.InvalidateCmd{})
	}

	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		minHeight := gtx.Dp(p.RowHeight)
		if minHeight > 0 {
			gtx.Constraints.Min.Y = minHeight
		}

		content := func(gtx layout.Context) layout.Dimensions {
			colorRole := theme.ThemeColorTextPrimary
			metaColorRole := theme.ThemeColorTextMuted

			tokens := p.tc.GetCurrentColorToken()
			iconColor := tokens.TextSecondaryNRGBA()

			if selected {
				colorRole = theme.ThemeColorOnPrimary
				metaColorRole = theme.ThemeColorOnPrimary
				iconColor = tokens.OnPrimaryNRGBA()
			}

			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return iconify.DefaultIconify.Layout(
						gtx,
						icon,
						unit.Dp(18),
						iconColor,
					)
				}),
				layout.Rigid(layout.Spacer{Width: unit.Dp(10)}.Layout),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(
						gtx,
						p.Theme,
						p.tc,
						theme.TextRoleBody,
						colorRole,
						e.Name,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := "dir"
					if !e.IsDir {
						label = formatSize(e.Size)
					}

					return theme.ThemedLabel(
						gtx,
						p.Theme,
						p.tc,
						theme.TextRoleCaption,
						metaColorRole,
						label,
					)
				}),
			)
		}

		inset := layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
			Left:   unit.Dp(10),
			Right:  unit.Dp(10),
		}

		switch {
		case selected:
			return panel.NewBackgroundPanel(p.tc).
				WithRole(panel.BackgroundRolePrimary).
				WithRadius(unit.Dp(10)).
				WithInset(inset).
				WithFillMax(false).
				Layout(gtx, content)

		case hovered:
			return panel.NewBackgroundPanel(p.tc).
				WithRole(panel.BackgroundRoleSurfaceAlt).
				WithRadius(unit.Dp(10)).
				WithInset(inset).
				WithFillMax(false).
				Layout(gtx, content)

		default:
			return inset.Layout(gtx, content)
		}
	})
}

func (p *FileExplorer) layoutPreviewHeader(gtx layout.Context) layout.Dimensions {
	if p.Selected == "" {
		return theme.ThemedLabel(
			gtx,
			p.Theme,
			p.tc,
			theme.TextRoleH4,
			theme.ThemeColorTextPrimary,
			"Preview",
		)
	}

	name := filepath.Base(p.Selected)
	kind := media.DetectKind(p.Selected)

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return theme.ThemedLabel(
				gtx,
				p.Theme,
				p.tc,
				theme.TextRoleH4,
				theme.ThemeColorTextPrimary,
				name,
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return theme.ThemedLabel(
					gtx,
					p.Theme,
					p.tc,
					theme.TextRoleCaption,
					theme.ThemeColorTextMuted,
					fmt.Sprintf("%s • %s", kind, p.Selected),
				)
			})
		}),
	)
}

func (p *FileExplorer) layoutMessage(gtx layout.Context, msg string) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return theme.ThemedLabel(
			gtx,
			p.Theme,
			p.tc,
			theme.TextRoleBody,
			theme.ThemeColorTextMuted,
			msg,
		)
	})
}

func (p *FileExplorer) iconButton(gtx layout.Context, click *widget.Clickable, icon string) layout.Dimensions {
	tokens := p.tc.GetCurrentColorToken()

	size := gtx.Dp(unit.Dp(34))
	gtx.Constraints.Min.X = size
	gtx.Constraints.Min.Y = size
	gtx.Constraints.Max.X = size
	gtx.Constraints.Max.Y = size

	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return iconify.DefaultIconify.Layout(
				gtx,
				icon,
				unit.Dp(18),
				tokens.TextSecondaryNRGBA(),
			)
		})
	})
}

func (p *FileExplorer) iconTextButton(
	gtx layout.Context,
	click *widget.Clickable,
	iconName string,
	text string,
) layout.Dimensions {
	if click == nil {
		return layout.Dimensions{}
	}

	hovered := click.Hovered()
	pressed := false

	for _, h := range click.History() {
		if h.End.IsZero() && !h.Cancelled {
			pressed = true
			break
		}
	}

	if hovered || pressed {
		gtx.Execute(op.InvalidateCmd{})
	}

	if hovered || pressed {
		gtx.Execute(op.InvalidateCmd{})
	}

	tokens := p.tc.GetCurrentColorToken()

	role := panel.BackgroundRolePrimary
	textRole := theme.ThemeColorOnPrimary
	iconColor := tokens.OnPrimaryNRGBA()

	if hovered {
		role = panel.BackgroundRoleSurfaceAlt
		textRole = theme.ThemeColorPrimary
		iconColor = tokens.PrimaryNRGBA()
	}

	if pressed {
		role = panel.BackgroundRolePrimary
		textRole = theme.ThemeColorOnPrimary
		iconColor = tokens.OnPrimaryNRGBA()
	}
	return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return panel.NewBackgroundPanel(p.tc).
			WithRole(role).
			WithRadius(unit.Dp(10)).
			WithInset(layout.Inset{
				Top:    unit.Dp(8),
				Bottom: unit.Dp(8),
				Left:   unit.Dp(12),
				Right:  unit.Dp(12),
			}).
			WithFillMax(false).
			Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return iconify.DefaultIconify.Layout(
							gtx,
							iconName,
							unit.Dp(16),
							iconColor,
						)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return theme.ThemedLabel(
							gtx,
							p.Theme,
							p.tc,
							theme.TextRoleLabel,
							textRole,
							text,
						)
					}),
				)
			})
	})
}

func (p *FileExplorer) textButton(gtx layout.Context, click *widget.Clickable, text string) layout.Dimensions {
	btn := material.Button(p.Theme, click, text)

	if p.tc != nil {
		tokens := p.tc.GetCurrentColorToken()
		btn.Background = tokens.PrimaryNRGBA()
		btn.Color = tokens.OnPrimaryNRGBA()
	}

	return btn.Layout(gtx)
}

func (p *FileExplorer) reload() {
	if p == nil {
		return
	}

	entries, err := readDir(p.CurrentDir, p.SearchText(), p.showHidden, p.SortBy, p.SortDesc)
	if err != nil {
		p.err = err
		p.entries = nil
		return
	}

	p.err = nil
	p.entries = entries

	live := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		live[e.Path] = struct{}{}
	}
	for path := range p.rowClicks {
		if _, ok := live[path]; !ok {
			delete(p.rowClicks, path)
		}
	}
}

func (p *FileExplorer) SearchText() string {
	if p == nil || p.Search == nil {
		return ""
	}
	return strings.TrimSpace(p.Search.Text())
}

func (p *FileExplorer) clickFor(path string) *widget.Clickable {
	if p.rowClicks == nil {
		p.rowClicks = make(map[string]*widget.Clickable)
	}

	click := p.rowClicks[path]
	if click == nil {
		click = new(widget.Clickable)
		p.rowClicks[path] = click
	}

	return click
}

func (p *FileExplorer) placeClickFor(path string) *widget.Clickable {
	if p.placeClicks == nil {
		p.placeClicks = make(map[string]*widget.Clickable)
	}

	key := filepath.Clean(path)

	click := p.placeClicks[key]
	if click == nil {
		click = new(widget.Clickable)
		p.placeClicks[key] = click
	}

	return click
}

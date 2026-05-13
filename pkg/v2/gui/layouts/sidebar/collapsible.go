package sidebar

import (
	"context"
	"image"
	"image/color"
	"math"
	"strings"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/components/tabs"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

type CollapsibleSidebar struct {
	Flip *tween.Flip
	BG   *tween.ColorTween

	// Optional. Useful for chevrons/icons.
	Rotation *tween.RotationTween

	Tabs *tabs.Layout

	CollapsedWidth unit.Dp
	ExpandedWidth  unit.Dp
	MinWidth       unit.Dp
	MaxWidth       unit.Dp
	Bar            unit.Dp

	Inset               unit.Dp
	ButtonGap           unit.Dp
	ButtonInset         unit.Dp
	ButtonRad           unit.Dp
	IconSize            unit.Dp
	CollapsedButtonSize unit.Dp
	CollapsedHideTextAt float64

	CollapsedInset unit.Dp

	// ShowExitButton renders an optional exit action pinned to the bottom of the sidebar.
	ShowExitButton bool
	ExitButtonText string
	ExitButtonIcon string
	ExitButton     *widget.Clickable

	Title string
	Icon  string
	Image image.Image

	Header      *components.Brand
	HeaderGap   unit.Dp
	HeaderInset unit.Dp

	theme *theme.Client

	drag     bool
	dragID   pointer.ID
	dragX    float32
	exitFunc func(gtx layout.Context)
}

const defaultBarWidth = unit.Dp(8)

func NewCollapsibleSidebar(tabLayout *tabs.Layout) *CollapsibleSidebar {
	tc := theme.DefaultThemeClient
	tokens := tc.GetCurrentColorToken()

	return &CollapsibleSidebar{
		theme: tc,
		Tabs:  tabLayout,
		Flip: tween.NewFlip(
			240*time.Millisecond,
			tween.EaseOutCubic,
		),
		BG: tween.NewColorTween(
			180*time.Millisecond,
			tween.EaseOutCubic,
			tokens.SurfaceNRGBA(),
		),
		exitFunc: func(gtx layout.Context) {

		},
		Rotation: tween.NewRotationTweenDeg(
			180*time.Millisecond,
			tween.EaseOutCubic,
			0,
		),
		Inset:          unit.Dp(12),
		CollapsedInset: unit.Dp(0),

		ButtonGap:   unit.Dp(8),
		ButtonInset: unit.Dp(10),
		ButtonRad:   unit.Dp(12),

		IconSize:            unit.Dp(20),
		CollapsedButtonSize: unit.Dp(40),
		CollapsedHideTextAt: 0.35,
		CollapsedWidth:      unit.Dp(64),
		ExpandedWidth:       unit.Dp(280),
		MinWidth:            unit.Dp(180),
		MaxWidth:            unit.Dp(520),
		Bar:                 defaultBarWidth,
		Title:               "WGL",
		Icon:                "",
		Image:               nil,

		Header: components.NewBrand("WGL").
			WithThemeClient(tc),
		HeaderGap:   unit.Dp(18),
		HeaderInset: unit.Dp(12),

		ShowExitButton: false,
		ExitButtonText: "Exit",
		ExitButtonIcon: "lucide:log-out",
		ExitButton:     &widget.Clickable{},
	}
}
func (s *CollapsibleSidebar) WithExitFunc(ef func(gtx layout.Context)) *CollapsibleSidebar {
	if s == nil {
		return s
	}
	s.exitFunc = ef
	return s
}
func (s *CollapsibleSidebar) WithExitButton(show bool) *CollapsibleSidebar {
	if s == nil {
		return s
	}

	s.ShowExitButton = show
	if s.ExitButton == nil {
		s.ExitButton = &widget.Clickable{}
	}
	if s.ExitButtonText == "" {
		s.ExitButtonText = "Exit"
	}
	if s.ExitButtonIcon == "" {
		s.ExitButtonIcon = "lucide:log-out"
	}
	return s
}
func (s *CollapsibleSidebar) WithExitButtonText(text string) *CollapsibleSidebar {
	if s == nil {
		return s
	}
	s.ExitButtonText = strings.TrimSpace(text)
	return s
}

func (s *CollapsibleSidebar) WithExitButtonIcon(icon string) *CollapsibleSidebar {
	if s == nil {
		return s
	}
	s.ExitButtonIcon = strings.TrimSpace(icon)
	return s
}

func (s *CollapsibleSidebar) ExitClicked(gtx layout.Context) bool {
	if s == nil || s.ExitButton == nil || !s.ShowExitButton {
		return false
	}

	clicked := false
	for s.ExitButton.Clicked(gtx) {
		clicked = true
	}
	if clicked {
		s.exitFunc(gtx)
	}
	return clicked
}

func (s *CollapsibleSidebar) WithTitle(title string) *CollapsibleSidebar {
	if s == nil {
		return s
	}

	s.Title = title

	if s.Header == nil {
		s.Header = components.NewBrand(title)
	}

	s.Header.Title = title
	s.Header.WithThemeClient(s.theme)

	return s
}

func (s *CollapsibleSidebar) WithIcon(icon string) *CollapsibleSidebar {
	if s == nil {
		return s
	}

	s.Icon = strings.TrimSpace(icon)

	if s.Header == nil {
		s.Header = components.NewBrand(s.Title)
	}

	s.Header.Icon = s.Icon
	s.Header.WithThemeClient(s.theme)

	return s
}

func (s *CollapsibleSidebar) WithImage(img image.Image) *CollapsibleSidebar {
	if s == nil {
		return s
	}

	s.Image = img

	if s.Header == nil {
		s.Header = components.NewBrand(s.Title)
	}

	s.Header.Image = img
	s.Header.WithThemeClient(s.theme)

	return s
}

func (s *CollapsibleSidebar) expansionProgress(now time.Time) (float64, bool) {
	if s == nil || s.Flip == nil {
		return 1, false
	}

	return s.Flip.Value(now)
}

func (s *CollapsibleSidebar) IsCollapsed() bool {
	if s == nil || s.Flip == nil {
		return false
	}

	progress, _ := s.Flip.Value(time.Now())
	return progress <= s.CollapsedHideTextAt
}

func (s *CollapsibleSidebar) WithThemeClient(tc *theme.Client) *CollapsibleSidebar {
	if s == nil {
		return s
	}

	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	s.theme = tc

	tokens := tc.GetCurrentColorToken()
	if s.BG != nil {
		s.BG.JumpTo(tokens.SurfaceNRGBA())
	}

	if s.Header != nil {
		s.Header.WithThemeClient(tc)
	}

	return s
}

func (s *CollapsibleSidebar) LayoutHeader(gtx layout.Context) layout.Dimensions {
	if s == nil {
		return layout.Dimensions{}
	}

	if s.Header == nil {
		s.Header = components.NewBrand(s.Title).
			WithIcon(s.Icon).
			WithImage(s.Image).
			WithThemeClient(s.theme)
	}

	collapsed := s.IsCollapsed()

	s.Header.Title = s.Title
	s.Header.Icon = s.Icon
	s.Header.Image = s.Image
	s.Header.ShowTitle = !collapsed
	s.Header.IconSize = s.CollapsedButtonSize
	s.Header.Radius = s.ButtonRad

	inset := s.HeaderInset
	if collapsed {
		inset = s.CollapsedInset
	}

	return layout.UniformInset(inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if collapsed {
			size := gtx.Dp(s.CollapsedButtonSize)
			if size <= 0 {
				size = 40
			}

			headerGtx := gtx
			headerGtx.Constraints.Min.X = size
			headerGtx.Constraints.Max.X = size
			headerGtx.Constraints.Min.Y = size
			headerGtx.Constraints.Max.Y = size

			return layout.Center.Layout(headerGtx, s.Header.Layout)
		}

		return s.Header.Layout(gtx)
	})
}

func (s *CollapsibleSidebar) WithTabs(tabLayout *tabs.Layout) *CollapsibleSidebar {
	if s == nil {
		return s
	}

	s.Tabs = tabLayout
	return s
}

func (s *CollapsibleSidebar) Toggle(now time.Time) {
	if s == nil {
		return
	}

	if s.Flip != nil && s.Flip.Expanded() {
		s.Collapse(now)
		return
	}

	s.Expand(now)
}

func (s *CollapsibleSidebar) Expand(now time.Time) {
	if s == nil {
		return
	}

	if s.Flip != nil {
		s.Flip.Expand()
	}

	style := s.style()
	if s.BG != nil {
		s.BG.AnimateToAt(now, style.OpenBG)
	}

	if s.Rotation != nil {
		s.Rotation.AnimateToDegrees(180)
	}
}

func (s *CollapsibleSidebar) Collapse(now time.Time) {
	if s == nil {
		return
	}

	if s.Flip != nil {
		s.Flip.Collapse()
	}

	style := s.style()
	if s.BG != nil {
		s.BG.AnimateToAt(now, style.ClosedBG)
	}

	if s.Rotation != nil {
		s.Rotation.AnimateToDegrees(0)
	}
}

func (s *CollapsibleSidebar) CurrentTab() (tabs.Tab, bool) {
	if s == nil || s.Tabs == nil {
		return tabs.Tab{}, false
	}

	return s.Tabs.CurrentTab()
}

func (s *CollapsibleSidebar) Buttons() []tabs.Button {
	if s == nil || s.Tabs == nil {
		return nil
	}

	return s.Tabs.Buttons()
}

// LayoutWithTabs renders the built-in tab buttons in the sidebar and the current
// tab page in the content area.
func (s *CollapsibleSidebar) LayoutWithTabs(gtx layout.Context) layout.Dimensions {
	if s == nil {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	return s.Layout(
		gtx,
		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return s.LayoutHeader(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Spacer{Height: s.HeaderGap}.Layout(gtx)
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return s.LayoutTabButtons(gtx)
				}),
			)
		},
		func(gtx layout.Context) layout.Dimensions {
			if s.Tabs == nil {
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}

			return s.Tabs.Layout(gtx)
		},
	)
}

// Layout keeps the old generic API: custom sidebar widget + custom content widget.
func (s *CollapsibleSidebar) Layout(
	gtx layout.Context,
	sidebar layout.Widget,
	content layout.Widget,
) layout.Dimensions {
	if s == nil {
		if content != nil {
			return content(gtx)
		}
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}

	now := time.Now()
	style := s.style()

	s.syncThemeTweens(now, style)

	progress := 1.0
	flipRunning := false
	if s.Flip != nil {
		progress, flipRunning = s.Flip.Value(now)
	}

	bg := style.ClosedBG
	bgRunning := false
	if s.BG != nil {
		bg, bgRunning = s.BG.Value(now)
	}

	rotRunning := false
	if s.Rotation != nil {
		_, rotRunning = s.Rotation.Value(now)
	}

	if flipRunning || bgRunning || rotRunning {
		gtx.Execute(op.InvalidateCmd{})
	}

	if s.theme != nil && s.theme.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	bar := gtx.Dp(s.Bar)
	if bar <= 1 {
		bar = gtx.Dp(defaultBarWidth)
	}

	collapsedPx := gtx.Dp(s.CollapsedWidth)
	expandedPx := gtx.Dp(s.ExpandedWidth)
	sidebarWidth := tween.MapInt(progress, collapsedPx, expandedPx)

	if sidebarWidth < 0 {
		sidebarWidth = 0
	}

	if sidebarWidth > gtx.Constraints.Max.X {
		sidebarWidth = gtx.Constraints.Max.X
	}

	rightOffset := sidebarWidth + bar
	contentWidth := gtx.Constraints.Max.X - rightOffset
	if contentWidth < 0 {
		contentWidth = 0
	}

	s.handleDrag(gtx, sidebarWidth, bar)

	// Sidebar.
	if sidebar != nil && sidebarWidth > 0 {
		sidebarGtx := gtx
		sidebarGtx.Constraints = layout.Exact(image.Pt(sidebarWidth, gtx.Constraints.Max.Y))

		utils.SurfaceOutlined(
			sidebarGtx,
			bg,
			0,
			utils.SurfaceBorder{
				Color: style.Border,
				Width: unit.Dp(1),
			},
			func(gtx layout.Context) layout.Dimensions {
				return sidebar(gtx)
			},
		)
	}

	// Resize bar.
	s.layoutBar(gtx, style, sidebarWidth, bar)

	// Content.
	if content != nil {
		off := op.Offset(image.Pt(rightOffset, 0)).Push(gtx.Ops)

		contentGtx := gtx
		contentGtx.Constraints = layout.Exact(image.Pt(contentWidth, gtx.Constraints.Max.Y))

		content(contentGtx)

		off.Pop()
	}

	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (s *CollapsibleSidebar) LayoutTabButtons(gtx layout.Context) layout.Dimensions {
	if s == nil {
		return layout.Dimensions{}
	}

	inset := s.Inset
	if s.IsCollapsed() {
		inset = s.CollapsedInset
	}

	return layout.UniformInset(inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return s.LayoutTabButtonGroupsOnly(gtx)
	})
}

func (s *CollapsibleSidebar) LayoutTabButtonGroupsOnly(gtx layout.Context) layout.Dimensions {
	if s == nil || s.Tabs == nil {
		return layout.Dimensions{}
	}

	now := time.Now()
	progress, running := s.expansionProgress(now)
	if running {
		gtx.Execute(op.InvalidateCmd{})
	}

	collapsed := progress <= s.CollapsedHideTextAt

	style := s.style()
	buttons := s.Tabs.Buttons()

	normal := make([]tabs.Button, 0, len(buttons))
	pinned := make([]tabs.Button, 0, len(buttons))

	for _, btn := range buttons {
		if btn.Pinned {
			pinned = append(pinned, btn)
		} else {
			normal = append(normal, btn)
		}
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutTabButtonGroup(gtx, style, normal, collapsed)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(0, gtx.Constraints.Max.Y)}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutTabButtonGroup(gtx, style, pinned, collapsed)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !s.ShowExitButton {
				return layout.Dimensions{}
			}
			if len(pinned) > 0 {
				return layout.Spacer{Height: s.ButtonGap}.Layout(gtx)
			}
			return layout.Dimensions{}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutExitButton(gtx, style, collapsed)
		}),
	)
}

func (s *CollapsibleSidebar) layoutTabButtonGroup(
	gtx layout.Context,
	style sidebarStyle,
	buttons []tabs.Button,
	collapsed bool,
) layout.Dimensions {
	if len(buttons) == 0 {
		return layout.Dimensions{}
	}

	children := make([]layout.FlexChild, 0, len(buttons)*2)

	for i := range buttons {
		btn := buttons[i]

		if i > 0 {
			children = append(children, layout.Rigid(layout.Spacer{Height: s.ButtonGap}.Layout))
		}

		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if btn.Clickable != nil {
				for btn.Clickable.Clicked(gtx) {
					if s.Tabs != nil && s.Tabs.SwitchToID(btn.ID) {
						gtx.Execute(op.InvalidateCmd{})
					}
				}
			}

			return s.layoutTabButton(gtx, style, btn, collapsed)
		}))
	}

	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx, children...)
}

func (s *CollapsibleSidebar) layoutTabButton(
	gtx layout.Context,
	style sidebarStyle,
	btn tabs.Button,
	collapsed bool,
) layout.Dimensions {
	if btn.Clickable == nil {
		return layout.Dimensions{}
	}

	bg := color.NRGBA{}
	border := utils.SurfaceBorder{
		Color: color.NRGBA{},
		Width: unit.Dp(0),
	}

	textColor := style.TextMuted

	if btn.Active {
		bg = style.ActiveBG
		textColor = style.Text
		border = utils.SurfaceBorder{
			Color: style.ActiveBorder,
			Width: unit.Dp(1),
		}
	} else if btn.Clickable.Hovered() {
		bg = style.HoverBG
		textColor = style.Text
		border = utils.SurfaceBorder{
			Color: style.Border,
			Width: unit.Dp(1),
		}
	}

	return btn.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if collapsed {
			size := gtx.Dp(s.CollapsedButtonSize)
			if size <= 0 {
				size = 40
			}

			buttonGtx := gtx
			buttonGtx.Constraints.Min.X = size
			buttonGtx.Constraints.Max.X = size
			buttonGtx.Constraints.Min.Y = size
			buttonGtx.Constraints.Max.Y = size

			return utils.SurfaceOutlined(
				buttonGtx,
				bg,
				s.ButtonRad,
				border,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return s.layoutTabIcon(gtx, btn.Icon, textColor)
					})
				},
			)
		}

		return utils.SurfaceOutlined(
			gtx,
			bg,
			s.ButtonRad,
			border,
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(s.ButtonInset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Middle,
					}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.layoutTabIcon(gtx, btn.Icon, textColor)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if btn.Icon == "" || btn.Name == "" {
								return layout.Dimensions{}
							}

							return layout.Spacer{Width: unit.Dp(8)}.Layout(gtx)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if btn.Name == "" {
								return layout.Dimensions{}
							}

							lbl := material.Body1(material.NewTheme(), btn.Name)
							lbl.Color = textColor
							lbl.Alignment = text.Middle

							theme.ApplyTypography(&lbl, style.Typo, theme.TextRoleLabel)

							return lbl.Layout(gtx)
						}),
					)
				})
			},
		)
	})
}

func (s *CollapsibleSidebar) layoutTabIcon(
	gtx layout.Context,
	name string,
	col color.NRGBA,
) layout.Dimensions {
	if name == "" {
		return layout.Dimensions{}
	}

	size := s.IconSize
	if size <= 0 {
		size = unit.Dp(20)
	}

	px := gtx.Dp(size)
	if px <= 0 {
		return layout.Dimensions{}
	}

	ic, err := iconify.DefaultIconify.Icon(context.Background(), name)
	if err != nil || ic == nil {
		return layout.Dimensions{Size: image.Pt(px, px)}
	}

	iconGtx := gtx
	iconGtx.Constraints.Min.X = px
	iconGtx.Constraints.Max.X = px
	iconGtx.Constraints.Min.Y = px
	iconGtx.Constraints.Max.Y = px

	return ic.Layout(iconGtx, size, col)
}

func (s *CollapsibleSidebar) syncThemeTweens(now time.Time, style sidebarStyle) {
	if s == nil || s.BG == nil {
		return
	}

	target := style.ClosedBG
	if s.Flip != nil && s.Flip.Expanded() {
		target = style.OpenBG
	}

	s.BG.AnimateToAt(now, target)
}

func (s *CollapsibleSidebar) handleDrag(gtx layout.Context, sidebarWidth, bar int) {
	if bar <= 0 {
		return
	}

	barRect := image.Rect(
		sidebarWidth,
		0,
		sidebarWidth+bar,
		gtx.Constraints.Max.Y,
	)

	area := clip.Rect(barRect).Push(gtx.Ops)
	defer area.Pop()

	event.Op(gtx.Ops, s)
	pointer.CursorColResize.Add(gtx.Ops)

	for {
		ev, ok := gtx.Event(pointer.Filter{
			Target: s,
			Kinds:  pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel,
		})
		if !ok {
			break
		}

		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}

		switch e.Kind {
		case pointer.Press:
			if s.drag {
				break
			}

			s.dragID = e.PointerID
			s.dragX = e.Position.X
			s.drag = true

			if s.Flip != nil && !s.Flip.Expanded() {
				s.Expand(time.Now())
				gtx.Execute(op.InvalidateCmd{})
			}

		case pointer.Drag:
			if s.dragID != e.PointerID {
				break
			}

			deltaX := e.Position.X - s.dragX
			s.dragX = e.Position.X

			s.resizeBy(gtx, deltaX)
			gtx.Execute(op.InvalidateCmd{})

			if e.Priority < pointer.Grabbed {
				gtx.Execute(pointer.GrabCmd{
					Tag: s,
					ID:  s.dragID,
				})
			}

		case pointer.Release, pointer.Cancel:
			s.drag = false
		}
	}
}

func (s *CollapsibleSidebar) resizeBy(gtx layout.Context, deltaX float32) {
	current := gtx.Dp(s.ExpandedWidth)
	next := current + int(math.Round(float64(deltaX)))

	minW := gtx.Dp(s.MinWidth)
	maxW := gtx.Dp(s.MaxWidth)

	if minW > 0 && next < minW {
		next = minW
	}

	if maxW > 0 && next > maxW {
		next = maxW
	}

	if next < gtx.Dp(s.CollapsedWidth) {
		next = gtx.Dp(s.CollapsedWidth)
	}

	s.ExpandedWidth = utils.PxToDp(gtx, next)
}

func (s *CollapsibleSidebar) layoutBar(gtx layout.Context, style sidebarStyle, x, width int) layout.Dimensions {
	if width <= 0 {
		return layout.Dimensions{}
	}

	off := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
	defer off.Pop()

	rect := image.Rect(0, 0, width, gtx.Constraints.Max.Y)

	paint.FillShape(
		gtx.Ops,
		style.Bar,
		clip.Rect(rect).Op(),
	)

	return layout.Dimensions{Size: rect.Size()}
}

type sidebarStyle struct {
	Tokens *theme.ColorTokens
	Typo   theme.TypographyTokens

	ClosedBG color.NRGBA
	OpenBG   color.NRGBA
	Border   color.NRGBA
	Bar      color.NRGBA

	ActiveBG     color.NRGBA
	ActiveBorder color.NRGBA
	HoverBG      color.NRGBA

	Text      color.NRGBA
	TextMuted color.NRGBA

	TextInverse color.NRGBA
	Danger      color.NRGBA
	DangerMuted color.NRGBA
}

func (s *CollapsibleSidebar) style() sidebarStyle {
	tc := s.theme
	if tc == nil {
		tc = theme.DefaultThemeClient
		s.theme = tc
	}

	tokens := tc.GetCurrentColorToken()
	typo := tc.GetCurrentTypography()

	return sidebarStyle{
		Tokens: tokens,
		Typo:   typo,

		ClosedBG: tokens.SurfaceNRGBA(),
		OpenBG:   tokens.SurfaceAltNRGBA(),
		Border:   tokens.BorderNRGBA(),
		Bar:      tokens.DividerNRGBA(),

		ActiveBG:     tokens.SelectionNRGBA(),
		ActiveBorder: tokens.PrimaryNRGBA(),
		HoverBG:      tokens.SurfaceAltNRGBA(),

		Text:        tokens.TextPrimaryNRGBA(),
		TextMuted:   tokens.TextMutedNRGBA(),
		Danger:      tokens.DangerNRGBA(),
		DangerMuted: withAlpha(tokens.DangerNRGBA(), 36),
	}
}
func withAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}

func (s *CollapsibleSidebar) layoutExitButton(
	gtx layout.Context,
	style sidebarStyle,
	collapsed bool,
) layout.Dimensions {
	if s == nil || !s.ShowExitButton {
		return layout.Dimensions{}
	}
	if s.ExitButton == nil {
		s.ExitButton = &widget.Clickable{}
	}

	for s.ExitButton.Clicked(gtx) {
		s.exitFunc(gtx)
	}

	bg := color.NRGBA{}
	border := utils.SurfaceBorder{}
	textColor := style.Danger
	if s.ExitButton.Hovered() {
		bg = style.DangerMuted
		border = utils.SurfaceBorder{Color: style.Danger, Width: unit.Dp(1)}
	}
	if s.ExitButton.Pressed() {
		bg = style.Danger
		textColor = style.TextInverse
	}

	name := s.ExitButtonText
	if name == "" {
		name = "Exit"
	}
	icon := s.ExitButtonIcon
	if icon == "" {
		icon = "lucide:log-out"
	}

	btn := tabs.Button{
		Name:      name,
		Icon:      icon,
		Clickable: s.ExitButton,
	}

	return s.ExitButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if collapsed {
			size := gtx.Dp(s.CollapsedButtonSize)
			if size <= 0 {
				size = 40
			}

			buttonGtx := gtx
			buttonGtx.Constraints.Min.X = size
			buttonGtx.Constraints.Max.X = size
			buttonGtx.Constraints.Min.Y = size
			buttonGtx.Constraints.Max.Y = size

			return utils.SurfaceOutlined(
				buttonGtx,
				bg,
				s.ButtonRad,
				border,
				func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return s.layoutTabIcon(gtx, btn.Icon, textColor)
					})
				},
			)
		}

		return utils.SurfaceOutlined(
			gtx,
			bg,
			s.ButtonRad,
			border,
			func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(s.ButtonInset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.layoutTabIcon(gtx, btn.Icon, textColor)
						}),
						layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Body1(material.NewTheme(), btn.Name)
							lbl.Color = textColor
							lbl.Alignment = text.Middle
							theme.ApplyTypography(&lbl, style.Typo, theme.TextRoleLabel)
							return lbl.Layout(gtx)
						}),
					)
				})
			},
		)
	})
}

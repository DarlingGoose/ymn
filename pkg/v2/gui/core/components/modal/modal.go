package modal

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

type XPosition int

const (
	Left XPosition = iota
	Center
	Right
)

type YPosition int

const (
	Top YPosition = iota
	Middle
	Bottom
)

type Position struct {
	X XPosition
	Y YPosition
}

var (
	TopLeft      = Position{X: Left, Y: Top}
	TopCenter    = Position{X: Center, Y: Top}
	TopRight     = Position{X: Right, Y: Top}
	MiddleLeft   = Position{X: Left, Y: Middle}
	MiddleCenter = Position{X: Center, Y: Middle}
	MiddleRight  = Position{X: Right, Y: Middle}
	BottomLeft   = Position{X: Left, Y: Bottom}
	BottomCenter = Position{X: Center, Y: Bottom}
	BottomRight  = Position{X: Right, Y: Bottom}
)

var (
	LeftTop      = TopLeft
	LeftMiddle   = MiddleLeft
	LeftBottom   = BottomLeft
	CenterTop    = TopCenter
	CenterMiddle = MiddleCenter
	CenterBottom = BottomCenter
	RightTop     = TopRight
	RightMiddle  = MiddleRight
	RightBottom  = BottomRight
)

type Modal struct {
	id string

	Title       string
	Description string

	Visible bool

	Controller *overlay.Controller

	Theme *material.Theme
	tc    *theme.Client

	Position Position

	Width     unit.Dp
	MaxWidth  unit.Dp
	MaxHeight unit.Dp
	MinHeight unit.Dp

	Margin  unit.Dp
	Padding unit.Dp
	Gap     unit.Dp
	Radius  unit.Dp

	ShowScrim    bool
	CloseOnScrim bool
	ShowClose    bool
	ShowHeader   bool
	ShowBorder   bool

	Scrim color.NRGBA

	TitleRole       theme.TextRole
	DescriptionRole theme.TextRole

	Content layout.Widget
	Footer  layout.Widget

	scrimClick widget.Clickable
	closeClick widget.Clickable
}

func New(id, title string, content layout.Widget) *Modal {
	if id == "" {
		id = fmt.Sprintf("modal:%s", title)
	}

	return &Modal{
		id: id,

		Title:   title,
		Content: content,

		Controller: overlay.DefaultController,

		Theme: material.NewTheme(),
		tc:    theme.DefaultThemeClient,

		Position: MiddleCenter,

		Width:     unit.Dp(520),
		MaxWidth:  unit.Dp(720),
		MaxHeight: unit.Dp(640),
		MinHeight: unit.Dp(0),

		Margin:  unit.Dp(24),
		Padding: unit.Dp(18),
		Gap:     unit.Dp(12),
		Radius:  unit.Dp(18),

		ShowScrim:    true,
		CloseOnScrim: true,
		ShowClose:    true,
		ShowHeader:   true,
		ShowBorder:   true,

		Scrim: color.NRGBA{A: 150},

		TitleRole:       theme.TextRoleH4,
		DescriptionRole: theme.TextRoleBodySmall,
	}
}

func (m *Modal) ID() string {
	if m == nil {
		return ""
	}
	return m.id
}

func (m *Modal) Open() {
	if m == nil {
		return
	}
	m.Visible = true
}

func (m *Modal) Close() {
	if m == nil {
		return
	}

	// Do not call Controller.Clear here.
	// overlay.Controller may call Close while holding its lock.
	m.Visible = false
}

func (m *Modal) Dismiss() {
	if m == nil {
		return
	}

	if m.Controller != nil {
		m.Controller.Clear(m)
	}

	m.Close()
}

func (m *Modal) Toggle() {
	if m == nil {
		return
	}

	if m.Visible {
		m.Dismiss()
		return
	}

	if m.Controller != nil {
		m.Controller.SetActive(m)
		return
	}

	m.Open()
}

func (m *Modal) WithThemeClient(tc *theme.Client) *Modal {
	if m == nil {
		return m
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	m.tc = tc
	return m
}

func (m *Modal) WithMaterialTheme(th *material.Theme) *Modal {
	if m == nil {
		return m
	}
	if th != nil {
		m.Theme = th
	}
	return m
}

func (m *Modal) WithController(controller *overlay.Controller) *Modal {
	if m == nil {
		return m
	}
	if controller == nil {
		controller = overlay.DefaultController
	}
	m.Controller = controller
	return m
}

func (m *Modal) WithPosition(pos Position) *Modal {
	if m == nil {
		return m
	}
	m.Position = pos
	return m
}

func (m *Modal) WithSize(width, maxHeight unit.Dp) *Modal {
	if m == nil {
		return m
	}
	m.Width = width
	m.MaxHeight = maxHeight
	return m
}

func (m *Modal) WithFooter(footer layout.Widget) *Modal {
	if m == nil {
		return m
	}
	m.Footer = footer
	return m
}

func (m *Modal) WithDescription(description string) *Modal {
	if m == nil {
		return m
	}
	m.Description = description
	return m
}

func (m *Modal) OverlayLayout(gtx layout.Context) {
	if m == nil || !m.Visible {
		return
	}
	if m.Theme == nil {
		m.Theme = material.NewTheme()
	}
	if m.tc == nil {
		m.tc = theme.DefaultThemeClient
	}
	if m.tc.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	screen := gtx.Constraints.Max
	if screen.X <= 0 || screen.Y <= 0 {
		return
	}

	if m.ShowScrim {
		if m.layoutScrim(gtx, screen) {
			return
		}
	}

	m.layoutModal(gtx, screen)
}

func (m *Modal) layoutScrim(gtx layout.Context, screen image.Point) bool {
	gtx.Constraints.Min = screen
	gtx.Constraints.Max = screen

	for m.scrimClick.Clicked(gtx) {
		if m.CloseOnScrim {
			m.Dismiss()
			gtx.Execute(op.InvalidateCmd{})
			return true
		}
	}

	m.scrimClick.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(
			gtx.Ops,
			m.Scrim,
			clip.Rect{Max: screen}.Op(),
		)

		return layout.Dimensions{Size: screen}
	})

	if m.CloseOnScrim && m.scrimClick.Pressed() {
		m.Dismiss()
		gtx.Execute(op.InvalidateCmd{})
		return true
	}

	return false
}

func (m *Modal) layoutModal(gtx layout.Context, screen image.Point) {
	margin := gtx.Dp(m.Margin)
	if margin < 0 {
		margin = 0
	}

	max := image.Point{
		X: screen.X - margin*2,
		Y: screen.Y - margin*2,
	}
	if max.X < 1 {
		max.X = 1
	}
	if max.Y < 1 {
		max.Y = 1
	}

	width := gtx.Dp(m.Width)
	if width <= 0 {
		width = max.X
	}
	if m.MaxWidth > 0 {
		width = minInt(width, gtx.Dp(m.MaxWidth))
	}
	width = minInt(width, max.X)

	maxHeight := max.Y
	if m.MaxHeight > 0 {
		maxHeight = minInt(maxHeight, gtx.Dp(m.MaxHeight))
	}

	modalGtx := gtx
	modalGtx.Constraints.Min.X = width
	modalGtx.Constraints.Max.X = width
	modalGtx.Constraints.Max.Y = maxHeight

	if m.MinHeight > 0 {
		modalGtx.Constraints.Min.Y = minInt(gtx.Dp(m.MinHeight), maxHeight)
	}

	macro := op.Record(gtx.Ops)
	dims := m.layoutCard(modalGtx)
	call := macro.Stop()

	offset := m.offset(screen, dims.Size, margin)
	stack := op.Offset(offset).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()
}

func (m *Modal) layoutCard(gtx layout.Context) layout.Dimensions {
	tokens := m.tc.GetCurrentColorToken()

	border := utils.SurfaceBorder{}
	if m.ShowBorder {
		border = utils.SurfaceBorder{
			Color: tokens.BorderNRGBA(),
			Width: unit.Dp(1),
		}
	}

	return utils.SurfaceOutlined(
		gtx,
		tokens.SurfaceNRGBA(),
		m.Radius,
		border,
		func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(m.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := make([]layout.FlexChild, 0, 5)

				if m.ShowHeader && (m.Title != "" || m.Description != "" || m.ShowClose) {
					children = append(children, layout.Rigid(m.layoutHeader))
				}

				if m.Content != nil {
					if len(children) > 0 {
						children = append(children, layout.Rigid(layout.Spacer{Height: m.Gap}.Layout))
					}
					children = append(children, layout.Flexed(1, m.Content))
				}

				if m.Footer != nil {
					children = append(children,
						layout.Rigid(layout.Spacer{Height: m.Gap}.Layout),
						layout.Rigid(m.Footer),
					)
				}

				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx, children...)
			})
		},
	)
}

func (m *Modal) layoutHeader(gtx layout.Context) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{
				Axis: layout.Vertical,
			}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if m.Title == "" {
						return layout.Dimensions{}
					}

					return theme.ThemedLabel(
						gtx,
						m.Theme,
						m.tc,
						m.TitleRole,
						theme.ThemeColorTextPrimary,
						m.Title,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if m.Description == "" {
						return layout.Dimensions{}
					}

					return layout.Inset{Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return theme.ThemedLabel(
							gtx,
							m.Theme,
							m.tc,
							m.DescriptionRole,
							theme.ThemeColorTextMuted,
							m.Description,
						)
					})
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !m.ShowClose {
				return layout.Dimensions{}
			}
			return m.layoutCloseButton(gtx)
		}),
	)
}

func (m *Modal) layoutCloseButton(gtx layout.Context) layout.Dimensions {
	for m.closeClick.Clicked(gtx) {
		m.Dismiss()
		gtx.Execute(op.InvalidateCmd{})
	}

	tokens := m.tc.GetCurrentColorToken()

	size := gtx.Dp(unit.Dp(34))
	gtx.Constraints.Min = image.Pt(size, size)
	gtx.Constraints.Max = image.Pt(size, size)

	dims := m.closeClick.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return iconify.DefaultIconify.Layout(
				gtx,
				"lucide:x",
				unit.Dp(18),
				tokens.TextSecondaryNRGBA(),
			)
		})
	})

	if m.closeClick.Pressed() {
		m.Dismiss()
		gtx.Execute(op.InvalidateCmd{})
	}

	return dims
}

func (m *Modal) offset(screen image.Point, size image.Point, margin int) image.Point {
	x := margin
	y := margin

	switch m.Position.X {
	case Left:
		x = margin
	case Center:
		x = (screen.X - size.X) / 2
	case Right:
		x = screen.X - size.X - margin
	}

	switch m.Position.Y {
	case Top:
		y = margin
	case Middle:
		y = (screen.Y - size.Y) / 2
	case Bottom:
		y = screen.Y - size.Y - margin
	}

	if x < margin {
		x = margin
	}
	if y < margin {
		y = margin
	}

	return image.Pt(x, y)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

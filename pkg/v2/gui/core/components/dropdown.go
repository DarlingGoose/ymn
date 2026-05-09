package components

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/utils"
)

type DropdownItem struct {
	Label string
	Value string
}

type Dropdown struct {
	Flip *tween.Flip
	BG   *tween.ColorTween

	Items    []DropdownItem
	Selected int

	Button         widget.Clickable
	ItemsClickable []widget.Clickable

	List layout.List

	Width         unit.Dp
	ItemHeight    unit.Dp
	MaxMenuHeight unit.Dp
	Radius        unit.Dp
	Inset         unit.Dp

	ClosedBG color.NRGBA
	OpenBG   color.NRGBA
	MenuBG   color.NRGBA
	Text     color.NRGBA
	Muted    color.NRGBA
}

func NewDropdown(items []DropdownItem) *Dropdown {
	closedBG := color.NRGBA{R: 38, G: 40, B: 48, A: 255}
	openBG := color.NRGBA{R: 55, G: 68, B: 115, A: 255}

	d := &Dropdown{
		Flip: tween.NewFlip(
			180*time.Millisecond,
			tween.EaseOutCubic,
		),
		BG: tween.NewColorTween(
			140*time.Millisecond,
			tween.EaseOutCubic,
			closedBG,
		),

		Items:          items,
		ItemsClickable: make([]widget.Clickable, len(items)),

		List: layout.List{
			Axis: layout.Vertical,
		},

		Width:         unit.Dp(240),
		ItemHeight:    unit.Dp(42),
		MaxMenuHeight: unit.Dp(260),
		Radius:        unit.Dp(12),
		Inset:         unit.Dp(12),

		ClosedBG: closedBG,
		OpenBG:   openBG,
		MenuBG:   color.NRGBA{R: 28, G: 30, B: 38, A: 255},
		Text:     color.NRGBA{R: 245, G: 247, B: 255, A: 255},
		Muted:    color.NRGBA{R: 180, G: 185, B: 205, A: 255},
	}

	if len(items) == 0 {
		d.Selected = -1
	}

	return d
}

func (d *Dropdown) Toggle(now time.Time) {
	if d == nil || d.Flip == nil {
		return
	}

	if d.Flip.Expanded() {
		d.Close(now)
		return
	}

	d.Open(now)
}

func (d *Dropdown) Open(now time.Time) {
	if d == nil {
		return
	}

	if d.Flip != nil {
		d.Flip.Expand()
	}

	if d.BG != nil {
		d.BG.AnimateTo(d.OpenBG)
	}
}

func (d *Dropdown) Close(now time.Time) {
	if d == nil {
		return
	}

	if d.Flip != nil {
		d.Flip.Collapse()
	}

	if d.BG != nil {
		d.BG.AnimateTo(d.ClosedBG)
	}
}

func (d *Dropdown) SelectedItem() (DropdownItem, bool) {
	if d == nil || d.Selected < 0 || d.Selected >= len(d.Items) {
		return DropdownItem{}, false
	}

	return d.Items[d.Selected], true
}

func (d *Dropdown) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if d == nil {
		return layout.Dimensions{}
	}

	now := time.Now()

	for d.Button.Clicked(gtx) {
		d.Toggle(now)
		gtx.Execute(op.InvalidateCmd{})
	}

	progress := 0.0
	positionRunning := false
	if d.Flip != nil {
		progress, positionRunning = d.Flip.Value(now)
	}

	bg := d.ClosedBG
	colorRunning := false
	if d.BG != nil {
		bg, colorRunning = d.BG.Value(now)
	}

	if positionRunning || colorRunning {
		gtx.Execute(op.InvalidateCmd{})
	}

	width := gtx.Dp(d.Width)
	buttonHeight := gtx.Dp(unit.Dp(44))
	itemHeight := gtx.Dp(d.ItemHeight)

	fullMenuHeight := itemHeight * len(d.Items)
	maxMenuHeight := gtx.Dp(d.MaxMenuHeight)
	if maxMenuHeight <= 0 {
		maxMenuHeight = fullMenuHeight
	}

	targetMenuHeight := tween.MinInt(fullMenuHeight, maxMenuHeight)
	menuHeight := tween.MapInt(progress, 0, targetMenuHeight)

	buttonGtx := gtx
	buttonGtx.Constraints.Min.X = width
	buttonGtx.Constraints.Max.X = width
	buttonGtx.Constraints.Min.Y = buttonHeight
	buttonGtx.Constraints.Max.Y = buttonHeight

	buttonDims := d.layoutButton(buttonGtx, th, bg, buttonHeight)

	if menuHeight > 0 {
		stack := op.Offset(image.Pt(0, buttonDims.Size.Y+gtx.Dp(unit.Dp(6)))).Push(gtx.Ops)

		menuGtx := gtx
		menuGtx.Constraints.Min.X = width
		menuGtx.Constraints.Max.X = width
		menuGtx.Constraints.Min.Y = menuHeight
		menuGtx.Constraints.Max.Y = menuHeight

		d.layoutMenu(menuGtx, th, menuHeight)

		stack.Pop()
	}

	// Return only the button dimensions so the dropdown is absolute.
	return buttonDims
}

func (d *Dropdown) layoutButton(
	gtx layout.Context,
	th *material.Theme,
	bg color.NRGBA,
	height int,
) layout.Dimensions {
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height

	return utils.ClickableSurface(gtx, &d.Button, bg, d.Radius, func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(d.Inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := "Select..."
			if item, ok := d.SelectedItem(); ok {
				label = item.Label
			}

			return layout.Flex{
				Axis:      layout.Horizontal,
				Alignment: layout.Middle,
			}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Body1(th, label)
					lbl.Color = d.Text
					return lbl.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					arrow := "⌄"
					if d.Flip != nil && d.Flip.Expanded() {
						arrow = "⌃"
					}

					lbl := material.Body1(th, arrow)
					lbl.Color = d.Muted
					return lbl.Layout(gtx)
				}),
			)
		})
	})
}

func (d *Dropdown) layoutMenu(
	gtx layout.Context,
	th *material.Theme,
	menuHeight int,
) layout.Dimensions {
	gtx.Constraints.Min.Y = menuHeight
	gtx.Constraints.Max.Y = menuHeight

	return utils.Surface(gtx, d.MenuBG, d.Radius, func(gtx layout.Context) layout.Dimensions {
		// Clip the menu to the animated/opened height.
		clipStack := clip.Rect{
			Max: image.Pt(gtx.Constraints.Max.X, menuHeight),
		}.Push(gtx.Ops)
		defer clipStack.Pop()

		listGtx := gtx
		listGtx.Constraints.Min.Y = menuHeight
		listGtx.Constraints.Max.Y = menuHeight

		return d.List.Layout(listGtx, len(d.Items), func(gtx layout.Context, index int) layout.Dimensions {
			return d.layoutItem(gtx, th, index)
		})
	})
}
func (d *Dropdown) layoutItem(gtx layout.Context, th *material.Theme, index int) layout.Dimensions {
	if index < 0 || index >= len(d.Items) {
		return layout.Dimensions{}
	}

	now := time.Now()

	for d.ItemsClickable[index].Clicked(gtx) {
		d.Selected = index
		d.Close(now)
		gtx.Execute(op.InvalidateCmd{})
	}

	item := d.Items[index]
	itemHeight := gtx.Dp(d.ItemHeight)

	gtx.Constraints.Min.Y = itemHeight
	gtx.Constraints.Max.Y = itemHeight

	bg := color.NRGBA{A: 0}
	if index == d.Selected {
		bg = color.NRGBA{R: 65, G: 82, B: 135, A: 255}
	} else if d.ItemsClickable[index].Hovered() {
		bg = color.NRGBA{R: 42, G: 45, B: 58, A: 255}
	}

	return utils.ClickableSurface(gtx, &d.ItemsClickable[index], bg, unit.Dp(0), func(gtx layout.Context) layout.Dimensions {
		return layout.UniformInset(d.Inset).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body1(th, item.Label)
			lbl.Color = d.Text
			return lbl.Layout(gtx)
		})
	})
}

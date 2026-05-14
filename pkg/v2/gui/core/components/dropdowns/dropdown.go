package dropdowns

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"strings"
	"sync/atomic"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/overlay"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

var dropdownIDCounter uint64

type DropdownItem struct {
	Label string
	Value string
}

type Dropdown struct {
	id string

	Flip          *tween.Flip
	BG            *tween.ColorTween
	ArrowRotation *tween.RotationTween

	Items        []DropdownItem
	Selected     int
	selectedFunc func(item DropdownItem, valid bool)

	Button         widget.Clickable
	ItemsClickable []widget.Clickable

	List          layout.List
	Height        unit.Dp
	Width         unit.Dp
	ItemHeight    unit.Dp
	MaxMenuHeight unit.Dp
	Radius        unit.Dp
	Inset         unit.Dp
	Role          theme.TextRole

	ArrowIconName string
	ArrowIconSize unit.Dp

	Controller *overlay.Controller

	theme *theme.Client
}

func NewDropdown(items []DropdownItem) *Dropdown {
	tc := theme.DefaultThemeClient
	tokens := tc.GetCurrentColorToken()

	id := atomic.AddUint64(&dropdownIDCounter, 1)

	d := &Dropdown{
		id:         fmt.Sprintf("dropdown-%d", id),
		theme:      tc,
		Controller: overlay.DefaultController,
		Height:     unit.Dp(44),
		Role:       theme.TextRoleH3,

		ArrowIconName: "lucide:chevron-down",
		ArrowIconSize: unit.Dp(18),
		ArrowRotation: tween.NewRotationTweenDeg(
			180*time.Millisecond,
			tween.EaseOutCubic,
			0,
		),
		Flip: tween.NewFlip(
			180*time.Millisecond,
			tween.EaseOutCubic,
		),

		BG: tween.NewColorTween(
			140*time.Millisecond,
			tween.EaseOutCubic,
			tokens.SurfaceNRGBA(),
		),

		Items:          items,
		ItemsClickable: make([]widget.Clickable, len(items)),

		List: layout.List{
			Axis: layout.Vertical,
		},

		selectedFunc: func(item DropdownItem, valid bool) {},

		Width:         unit.Dp(240),
		ItemHeight:    unit.Dp(42),
		MaxMenuHeight: unit.Dp(260),
		Radius:        unit.Dp(12),
		Inset:         unit.Dp(12),
	}

	if len(items) == 0 {
		d.Selected = -1
	}

	return d
}
func (d *Dropdown) WithCompact() *Dropdown {
	d.Width = unit.Dp(120)
	d.Height = unit.Dp(24)
	d.ItemHeight = unit.Dp(22)
	d.MaxMenuHeight = unit.Dp(190)
	d.Inset = unit.Dp(8)
	d.ArrowIconSize = unit.Dp(14)
	d.WithRole(theme.TextRoleCaption)
	return d
}
func (d *Dropdown) WithArrowIcon(name string) *Dropdown {
	if d == nil {
		return d
	}

	if strings.TrimSpace(name) != "" {
		d.ArrowIconName = name
	}

	return d
}
func (d *Dropdown) ID() string {
	if d == nil {
		return ""
	}

	return d.id
}

func (d *Dropdown) WithID(id string) *Dropdown {
	if d == nil {
		return d
	}

	if strings.TrimSpace(id) != "" {
		d.id = id
	}

	return d
}

func (d *Dropdown) WithRole(role theme.TextRole) *Dropdown {
	if d == nil {
		return d
	}

	d.Role = role
	return d
}

func (d *Dropdown) WithThemeClient(tc *theme.Client) *Dropdown {
	if d == nil {
		return d
	}

	if tc == nil {
		tc = theme.DefaultThemeClient
	}

	d.theme = tc

	tokens := tc.GetCurrentColorToken()
	if d.BG != nil {
		d.BG.JumpTo(tokens.SurfaceNRGBA())
	}

	return d
}

func (d *Dropdown) WithController(controller *overlay.Controller) *Dropdown {
	if d == nil {
		return d
	}

	d.Controller = controller
	return d
}

func (d *Dropdown) Toggle(now time.Time) {
	if d == nil || d.Flip == nil {
		return
	}

	if d.Flip.Expanded() {
		d.Close(now)
		if d.Controller != nil {
			d.Controller.Clear(dropdownControlComponent{dropdown: d})
		}
		return
	}

	if d.Controller != nil {
		d.Controller.SetActive(dropdownControlComponent{dropdown: d})
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

	if d.ArrowRotation != nil {
		d.ArrowRotation.AnimateToDegrees(180)
	}

	style := d.style()
	if d.BG != nil {
		d.BG.AnimateToAt(now, style.ButtonOpenBG)
	}
}

func (d *Dropdown) Close(now time.Time) {
	if d == nil {
		return
	}

	if d.Flip != nil {
		d.Flip.Collapse()
	}

	if d.ArrowRotation != nil {
		d.ArrowRotation.AnimateToDegrees(0)
	}

	style := d.style()
	if d.BG != nil {
		d.BG.AnimateToAt(now, style.ButtonBG)
	}
}

func (d *Dropdown) SetItems(items []DropdownItem) {
	if d == nil {
		return
	}

	d.Items = items
	d.ItemsClickable = make([]widget.Clickable, len(items))

	if len(items) == 0 {
		d.Selected = -1
		return
	}

	if d.Selected < 0 || d.Selected >= len(items) {
		d.Selected = 0
	}
}

func (d *Dropdown) SelectItemEvent(v func(item DropdownItem, valid bool)) {
	if d == nil {
		return
	}

	if v == nil {
		d.selectedFunc = func(item DropdownItem, valid bool) {}
		return
	}

	d.selectedFunc = v
}

func (d *Dropdown) SelectItem(name string) bool {
	if d == nil {
		return false
	}

	for index, item := range d.Items {
		if strings.EqualFold(name, item.Value) || strings.EqualFold(name, item.Label) {
			d.Selected = index
			return true
		}
	}

	return false
}

func (d *Dropdown) SelectedItem() (DropdownItem, bool) {
	if d == nil || d.Selected < 0 || d.Selected >= len(d.Items) {
		return DropdownItem{}, false
	}

	return d.Items[d.Selected], true
}

func (d *Dropdown) Layout(gtx layout.Context, layer *overlay.Overlay) layout.Dimensions {
	if d == nil {
		return layout.Dimensions{}
	}

	now := time.Now()
	style := d.style()

	d.syncThemeTweens(now, style)

	for d.Button.Clicked(gtx) {
		d.Toggle(now)
		gtx.Execute(op.InvalidateCmd{})
	}

	progress := 0.0
	positionRunning := false
	if d.Flip != nil {
		progress, positionRunning = d.Flip.Value(now)
	}

	bg := style.Outline
	colorRunning := false
	if d.BG != nil {
		bg, colorRunning = d.BG.Value(now)
	}
	arrowRunning := false
	if d.ArrowRotation != nil {
		_, arrowRunning = d.ArrowRotation.Value(now)
	}

	if positionRunning || colorRunning || arrowRunning {
		gtx.Execute(op.InvalidateCmd{})
	}

	if d.theme != nil && d.theme.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	width := gtx.Dp(d.Width)
	buttonHeight := gtx.Dp(d.Height)
	itemHeight := gtx.Dp(d.ItemHeight)

	fullMenuHeight := itemHeight * len(d.Items)

	maxMenuHeight := gtx.Dp(d.MaxMenuHeight)
	if maxMenuHeight <= 0 {
		maxMenuHeight = fullMenuHeight
	}

	targetMenuHeight := tween.MinInt(fullMenuHeight, maxMenuHeight)
	menuHeight := tween.MapInt(progress, 0, targetMenuHeight)
	gap := gtx.Dp(unit.Dp(6))

	buttonGtx := gtx
	buttonGtx.Constraints.Min.X = width
	buttonGtx.Constraints.Max.X = width
	buttonGtx.Constraints.Min.Y = buttonHeight
	buttonGtx.Constraints.Max.Y = buttonHeight

	buttonDims := d.layoutButton(buttonGtx, style, bg, buttonHeight)

	if menuHeight > 0 {
		menu := dropdownMenuComponent{
			id:           d.id + "/menu",
			dropdown:     d,
			style:        style,
			width:        width,
			buttonHeight: buttonHeight,
			menuHeight:   menuHeight,
			gap:          gap,
		}

		if layer != nil {
			layer.Add(gtx, menu)
		} else {
			menu.Layout(gtx)
		}
	}

	return buttonDims
}

func (d *Dropdown) layoutButton(
	gtx layout.Context,
	style dropdownStyle,
	bg color.NRGBA,
	height int,
) layout.Dimensions {
	gtx.Constraints.Min.Y = height
	gtx.Constraints.Max.Y = height

	return utils.ClickableSurfaceOutlined(
		gtx,
		&d.Button,
		bg,
		d.Radius,
		utils.SurfaceBorder{Color: style.Outline, Width: unit.Dp(2)},
		func(gtx layout.Context) layout.Dimensions {
			inset := d.insetForHeight(gtx, height)

			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				label := "Select..."
				if item, ok := d.SelectedItem(); ok {
					label = item.Label
				}

				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis:      layout.Horizontal,
						Alignment: layout.Middle,
					}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return d.layoutTextHeight(gtx, style, label, style.Text, height)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return d.layoutArrowIcon(gtx, style)
						}),
					)
				})
			})
		},
	)
}

func (d *Dropdown) layoutArrowIcon(
	gtx layout.Context,
	style dropdownStyle,
) layout.Dimensions {
	if d == nil {
		return layout.Dimensions{}
	}

	iconName := d.ArrowIconName
	if strings.TrimSpace(iconName) == "" {
		iconName = "lucide:chevron-down"
	}

	size := d.ArrowIconSize
	if size <= 0 {
		size = unit.Dp(18)
	}

	angle := 0.0
	if d.ArrowRotation != nil {
		angle, _ = d.ArrowRotation.Value(time.Now())
	}

	ic, err := iconify.DefaultIconify.Icon(context.Background(), iconName)
	if err != nil || ic == nil {
		return layout.Dimensions{
			Size: image.Pt(gtx.Dp(size), gtx.Dp(size)),
		}
	}

	return iconify.LayoutRotatedIcon(gtx, ic, size, style.Muted, float32(angle))
}

func (d *Dropdown) WithArrowIconSize(size unit.Dp) *Dropdown {
	if d == nil {
		return d
	}

	if size > 0 {
		d.ArrowIconSize = size
	}

	return d
}

func (d *Dropdown) insetForHeight(gtx layout.Context, height int) layout.Inset {
	inset := d.Inset
	if inset <= 0 {
		inset = unit.Dp(8)
	}

	// Compact mode: avoid clipping text by removing vertical padding.
	if height <= gtx.Dp(unit.Dp(28)) {
		return layout.Inset{
			Top:    unit.Dp(0),
			Bottom: unit.Dp(0),
			Left:   unit.Dp(8),
			Right:  unit.Dp(8),
		}
	}

	if height <= gtx.Dp(unit.Dp(34)) {
		return layout.Inset{
			Top:    unit.Dp(2),
			Bottom: unit.Dp(2),
			Left:   inset,
			Right:  inset,
		}
	}

	return layout.UniformInset(inset)
}

func (d *Dropdown) compactText(gtx layout.Context, height int) bool {
	return height <= gtx.Dp(unit.Dp(30))
}

func (d *Dropdown) layoutMenu(
	gtx layout.Context,
	style dropdownStyle,
	menuHeight int,
) layout.Dimensions {
	gtx.Constraints.Min.Y = menuHeight
	gtx.Constraints.Max.Y = menuHeight

	return utils.SurfaceOutlined(gtx, style.MenuBG, d.Radius, utils.SurfaceBorder{
		Color: style.Outline,
		Width: unit.Dp(1),
	}, func(gtx layout.Context) layout.Dimensions {
		clipStack := clip.UniformRRect(
			image.Rectangle{Max: image.Pt(gtx.Constraints.Max.X, menuHeight)},
			gtx.Dp(d.Radius),
		).Push(gtx.Ops)
		defer clipStack.Pop()

		listGtx := gtx
		listGtx.Constraints.Min.Y = menuHeight
		listGtx.Constraints.Max.Y = menuHeight

		dims := d.List.Layout(listGtx, len(d.Items), func(gtx layout.Context, index int) layout.Dimensions {
			return d.layoutItem(gtx, style, index)
		})
		d.layoutScrollIndicator(gtx, style, menuHeight)
		return dims
	})
}

func (d *Dropdown) layoutScrollIndicator(gtx layout.Context, style dropdownStyle, menuHeight int) {
	if d == nil || len(d.Items) == 0 || menuHeight <= 0 {
		return
	}

	itemHeight := gtx.Dp(d.ItemHeight)
	if itemHeight <= 0 {
		return
	}

	fullHeight := itemHeight * len(d.Items)
	if fullHeight <= menuHeight {
		return
	}

	trackHeight := menuHeight - gtx.Dp(unit.Dp(16))
	if trackHeight <= 0 {
		return
	}

	thumbHeight := menuHeight * trackHeight / fullHeight
	minThumbHeight := gtx.Dp(unit.Dp(24))
	if thumbHeight < minThumbHeight {
		thumbHeight = minThumbHeight
	}
	if thumbHeight > trackHeight {
		thumbHeight = trackHeight
	}

	maxScroll := fullHeight - menuHeight
	scrollTop := d.List.Position.First*itemHeight - d.List.Position.Offset
	if scrollTop < 0 {
		scrollTop = 0
	}
	if scrollTop > maxScroll {
		scrollTop = maxScroll
	}

	thumbTop := gtx.Dp(unit.Dp(8))
	if maxScroll > 0 {
		thumbTop += scrollTop * (trackHeight - thumbHeight) / maxScroll
	}

	width := gtx.Dp(unit.Dp(3))
	x := gtx.Constraints.Max.X - gtx.Dp(unit.Dp(7))
	if x < 0 {
		x = 0
	}

	trackColor := style.Muted
	trackColor.A = 48
	thumbColor := style.Text
	thumbColor.A = 150

	track := image.Rect(x, gtx.Dp(unit.Dp(8)), x+width, gtx.Dp(unit.Dp(8))+trackHeight)
	thumb := image.Rect(x, thumbTop, x+width, thumbTop+thumbHeight)

	paint.FillShape(gtx.Ops, trackColor, clip.UniformRRect(track, width/2).Op(gtx.Ops))
	paint.FillShape(gtx.Ops, thumbColor, clip.UniformRRect(thumb, width/2).Op(gtx.Ops))
}

func (d *Dropdown) layoutItem(
	gtx layout.Context,
	style dropdownStyle,
	index int,
) layout.Dimensions {
	if d == nil || index < 0 || index >= len(d.Items) {
		return layout.Dimensions{}
	}

	now := time.Now()

	for d.ItemsClickable[index].Clicked(gtx) {
		d.Selected = index

		item, valid := d.SelectedItem()
		d.selectedFunc(item, valid)

		d.Close(now)
		if d.Controller != nil {
			d.Controller.Clear(dropdownControlComponent{dropdown: d})
		}

		gtx.Execute(op.InvalidateCmd{})
	}

	item := d.Items[index]
	itemHeight := gtx.Dp(d.ItemHeight)

	gtx.Constraints.Min.Y = itemHeight
	gtx.Constraints.Max.Y = itemHeight

	bg := color.NRGBA{A: 0}
	if index == d.Selected {
		bg = style.ItemSelectedBG
	} else if d.ItemsClickable[index].Hovered() {
		bg = style.ItemHoverBG
	}

	corners := utils.CornerRadius{}

	last := len(d.Items) - 1
	if index == 0 {
		corners.TopLeft = d.Radius
		corners.TopRight = d.Radius
	}
	if index == last {
		corners.BottomLeft = d.Radius
		corners.BottomRight = d.Radius
	}

	return utils.ClickableSurfaceCorners(
		gtx,
		&d.ItemsClickable[index],
		bg,
		corners,
		func(gtx layout.Context) layout.Dimensions {
			inset := d.insetForHeight(gtx, itemHeight)

			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return d.layoutTextHeight(gtx, style, item.Label, style.Text, itemHeight)
				})
			})
		},
	)
}

func (d *Dropdown) layoutTextHeight(
	gtx layout.Context,
	style dropdownStyle,
	value string,
	col color.NRGBA,
	height int,
) layout.Dimensions {
	th := material.NewTheme()

	lbl := material.Body1(th, value)
	lbl.Color = col
	lbl.Alignment = text.Middle

	theme.ApplyTypography(&lbl, style.Typo, d.Role)

	if d.compactText(gtx, height) {
		lbl.TextSize = unit.Sp(12)
		lbl.LineHeight = unit.Sp(14)
	}

	return lbl.Layout(gtx)
}

func (d *Dropdown) syncThemeTweens(now time.Time, style dropdownStyle) {
	if d == nil || d.BG == nil {
		return
	}

	target := style.ButtonBG
	if d.Flip != nil && d.Flip.Expanded() {
		target = style.ButtonOpenBG
	}

	d.BG.AnimateToAt(now, target)
}

func (d *Dropdown) layoutText(
	gtx layout.Context,
	style dropdownStyle,
	value string,
	col color.NRGBA,
) layout.Dimensions {
	th := material.NewTheme()

	lbl := material.Body1(th, value)
	lbl.Color = col
	lbl.Alignment = text.Middle

	theme.ApplyTypography(&lbl, style.Typo, d.Role)

	return lbl.Layout(gtx)
}

type dropdownControlComponent struct {
	dropdown *Dropdown
}

func (c dropdownControlComponent) ID() string {
	if c.dropdown == nil {
		return ""
	}

	return c.dropdown.ID()
}

func (c dropdownControlComponent) Open() {
	if c.dropdown == nil {
		return
	}

	c.dropdown.Open(time.Now())
}

func (c dropdownControlComponent) Close() {
	if c.dropdown == nil {
		return
	}

	c.dropdown.Close(time.Now())
}

func (c dropdownControlComponent) Layout(gtx layout.Context) {
	// Required by overlay.Component, but controller-only components do not draw.
}

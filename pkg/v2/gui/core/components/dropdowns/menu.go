package dropdowns

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
)

type dropdownMenuComponent struct {
	id string

	dropdown *Dropdown
	style    dropdownStyle

	width        int
	buttonHeight int
	menuHeight   int
	gap          int
}

func (m dropdownMenuComponent) OverlayLayout(gtx layout.Context) {
	if m.dropdown == nil || m.menuHeight <= 0 {
		return
	}

	stack := op.Offset(image.Pt(0, m.buttonHeight+m.gap)).Push(gtx.Ops)
	defer stack.Pop()

	menuGtx := gtx
	menuGtx.Constraints.Min.X = m.width
	menuGtx.Constraints.Max.X = m.width
	menuGtx.Constraints.Min.Y = m.menuHeight
	menuGtx.Constraints.Max.Y = m.menuHeight

	m.dropdown.layoutMenu(menuGtx, m.style, m.menuHeight)
}

func (m dropdownMenuComponent) Layout(gtx layout.Context) {
	m.OverlayLayout(gtx)
}

func (m dropdownMenuComponent) ID() string {
	return m.id
}

func (m dropdownMenuComponent) Open() {}

func (m dropdownMenuComponent) Close() {}

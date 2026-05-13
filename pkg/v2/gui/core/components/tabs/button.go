package tabs

import "gioui.org/widget"

type Button struct {
	ID     string
	Name   string
	Icon   string
	Pinned bool
	Active bool

	Clickable *widget.Clickable
}

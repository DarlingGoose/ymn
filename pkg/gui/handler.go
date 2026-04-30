package gui

import (
	"context"

	"gioui.org/app"
	"gioui.org/layout"
)

type EvenHandler interface {
	HandleEvents(gtx layout.Context, ctx context.Context, w *app.Window)
}

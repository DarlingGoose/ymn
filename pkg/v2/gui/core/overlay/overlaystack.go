package overlay

import (
	"gioui.org/layout"
	"gioui.org/op"
)

type Renderer interface {
	OverlayLayout(gtx layout.Context)
}

type RendererFunc func(gtx layout.Context)

func (fn RendererFunc) OverlayLayout(gtx layout.Context) {
	if fn != nil {
		fn(gtx)
	}
}

type Overlay struct {
	items []op.CallOp
}

func (o *Overlay) Add(gtx layout.Context, renderer Renderer) {
	if o == nil || renderer == nil {
		return
	}

	macro := op.Record(gtx.Ops)
	renderer.OverlayLayout(gtx)
	call := macro.Stop()

	// Draw later, but keep the current coordinate context.
	op.Defer(gtx.Ops, call)
}

func (o *Overlay) AddFunc(gtx layout.Context, fn func(gtx layout.Context)) {
	o.Add(gtx, RendererFunc(fn))
}

func (o *Overlay) Layout(gtx layout.Context, content layout.Widget) layout.Dimensions {
	if o == nil {
		return content(gtx)
	}

	return content(gtx)
}
func (o *Overlay) Reset() {
	if o == nil {
		return
	}

	o.items = o.items[:0]
}

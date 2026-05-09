package keypress

import (
	"context"

	"gioui.org/io/pointer"
)

type KeyHandler func(ctx context.Context, kind pointer.Kind)
type KeyPress struct {
}

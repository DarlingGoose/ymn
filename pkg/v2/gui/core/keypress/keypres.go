package keypress

import (
	"context"

	"gioui.org/io/pointer"
)

var DefaultKeyPress = &KeyPress{}

// todo add logic to get default key binds, and over write it and save it to the user config
// these should be able to do a number of of different actions
// want to support press and hold - handle event duration as well, and press and release
// want to be able to ignore keypresses when in text boxes to not trigger any unwanted affevts
type KeyPress struct {
}

type KeyAction interface {
	Name() string
	Description() string
	HandleEvent(ctx context.Context, kind pointer.Kind) error
}

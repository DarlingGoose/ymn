package toast

import (
	"context"
	"image"
	"image/color"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	bareui "github.com/Seann-Moser/bare/pkg/ui"
	"github.com/Seann-Moser/bare/pkg/ui/icons"
	barethemes "github.com/Seann-Moser/bare/pkg/ui/themes"
	bareutils "github.com/Seann-Moser/bare/pkg/ui/utils"
	"github.com/Seann-Moser/wgl/pkg/gui"
)

const (
	defaultQueueSize    = 32
	defaultToastTimeout = 4 * time.Second
	defaultMaxVisible   = 4
)

var _ gui.EvenHandler = &Toast{}

type Toast struct {
	Notifications chan Notification
	items         []item
	nextID        uint64
	MaxVisible    int
	DefaultTTL    time.Duration
}

func New() *Toast {
	return &Toast{
		Notifications: make(chan Notification, defaultQueueSize),
		MaxVisible:    defaultMaxVisible,
		DefaultTTL:    defaultToastTimeout,
	}
}

func (t *Toast) Shutdown() {
	if t.Notifications != nil {
		close(t.Notifications)
		t.Notifications = nil
	}
}

func (t *Toast) Queue(notification ...Notification) {
	for _, n := range notification {
		if t.Notifications == nil {
			return
		}
		if n.CloseAfter <= 0 {
			n.CloseAfter = t.DefaultTTL
			if n.CloseAfter <= 0 {
				n.CloseAfter = defaultToastTimeout
			}
		}
		select {
		case t.Notifications <- n:
		default:
			// Keep the caller non-blocking; drop the oldest visible toast.
			if len(t.items) > 0 {
				t.items = t.items[1:]
			}
			select {
			case t.Notifications <- n:
			default:
			}
		}
	}
}

func (t *Toast) HandleEvents(gtx layout.Context, _ context.Context, w *app.Window) {
	t.drainQueue()
	if len(t.items) == 0 {
		return
	}

	now := time.Now()
	filtered := t.items[:0]
	for _, toast := range t.items {
		dismissed := false
		for toast.close.Clicked(gtx) {
			dismissed = true
		}
		if dismissed || now.After(toast.expiresAt) {
			continue
		}
		filtered = append(filtered, toast)
	}
	t.items = filtered
	if len(t.items) > 0 {
		w.Invalidate()
	}
}

func (t *Toast) Layout(gtx layout.Context, theme barethemes.Theme, iconify *icons.Iconify) layout.Dimensions {
	if len(t.items) == 0 {
		return layout.Dimensions{}
	}
	visible := t.items
	if t.MaxVisible > 0 && len(visible) > t.MaxVisible {
		visible = visible[len(visible)-t.MaxVisible:]
	}
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: unit.Dp(18), Bottom: unit.Dp(18), Left: unit.Dp(18), Right: unit.Dp(18)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			maxWidth := gtx.Dp(unit.Dp(420))
			if maxWidth > gtx.Constraints.Max.X {
				maxWidth = gtx.Constraints.Max.X
			}
			gtx.Constraints.Min.X = 0
			gtx.Constraints.Max.X = maxWidth
			gtx.Constraints.Max.Y = gtx.Dp(unit.Dp(120))
			children := make([]layout.FlexChild, 0, len(visible))
			for i := range visible {
				toast := &visible[i]
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return t.layoutToast(gtx, theme, iconify, toast)
					})
				}))
			}
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx, children...)
		})
	})
}

func (t *Toast) layoutToast(gtx layout.Context, theme barethemes.Theme, iconify *icons.Iconify, toast *item) layout.Dimensions {
	gtx.Constraints.Min.X = 0

	closeBtn := bareui.Button{
		Clickable: &toast.close,
		Text:      "mdi:close",
		Icon:      true,
		Prefix:    "mdi:close",
		Variant:   bareui.ButtonGhost,
	}
	accent := t.accentColor(theme, toast.note.Type)
	bg := barethemes.Mix(accent, theme.Color.Surface, 0.10)
	borderColor := barethemes.Mix(accent, theme.Color.Border, 0.35)
	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return bareutils.Panel(gtx, bg, unit.Dp(theme.Radius.MD), func(gtx layout.Context) layout.Dimensions {
				return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return bareutils.Panel(gtx, barethemes.Mix(accent, bg, 0.55), unit.Dp(theme.Radius.SM), func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Min = image.Point{X: gtx.Dp(unit.Dp(4)), Y: gtx.Dp(unit.Dp(56))}
								gtx.Constraints.Max = gtx.Constraints.Min
								return layout.Dimensions{Size: gtx.Constraints.Min}
							})
						}),
						layout.Rigid(bareutils.SpacerW(unit.Dp(12))),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											if iconify != nil && t.iconName(toast.note) != "" {
												return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
													return iconify.Layout(gtx, t.iconName(toast.note), unit.Dp(18), accent)
												})
											}
											return layout.Dimensions{}
										}),
										layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
											lbl := material.Body1(theme.Gio(), t.titleText(toast.note))
											lbl.Color = theme.Color.Text
											return lbl.Layout(gtx)
										}),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											return closeBtn.Layout(gtx, theme, iconify)
										}),
									)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									message := strings.TrimSpace(toast.note.Message)
									if message == "" {
										return layout.Dimensions{}
									}
									return layout.Inset{Top: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										lbl := material.Body1(theme.Gio(), message)
										lbl.Color = theme.Color.TextMuted
										return lbl.Layout(gtx)
									})
								}),
							)
						}),
					)
				})
			})
		}),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			border := clip.Stroke{
				Path: clip.RRect{
					Rect: image.Rectangle{Max: gtx.Constraints.Max},
					NW:   gtx.Dp(unit.Dp(theme.Radius.MD)),
					NE:   gtx.Dp(unit.Dp(theme.Radius.MD)),
					SW:   gtx.Dp(unit.Dp(theme.Radius.MD)),
					SE:   gtx.Dp(unit.Dp(theme.Radius.MD)),
				}.Path(gtx.Ops),
				Width: float32(gtx.Dp(unit.Dp(1))),
			}.Op()
			paint.FillShape(gtx.Ops, borderColor, border)
			return layout.Dimensions{}
		}),
	)
}

func (t *Toast) drainQueue() {
	for {
		select {
		case n, ok := <-t.Notifications:
			if !ok {
				return
			}
			t.nextID++
			now := time.Now()
			t.items = append(t.items, item{
				id:        t.nextID,
				createdAt: now,
				expiresAt: now.Add(n.CloseAfter),
				note:      n,
			})
		default:
			return
		}
	}
}

func (t *Toast) titleText(n Notification) string {
	title := strings.TrimSpace(n.Title)
	if title != "" {
		return title
	}
	switch n.Type {
	case NotificationTypeSuccess:
		return "Success"
	case NotificationTypeWarning:
		return "Warning"
	case NotificationTypeError:
		return "Error"
	default:
		return "Info"
	}
}

func (t *Toast) iconName(n Notification) string {
	if strings.TrimSpace(n.Icon) != "" {
		return n.Icon
	}
	switch n.Type {
	case NotificationTypeSuccess:
		return "mdi:check-circle-outline"
	case NotificationTypeWarning:
		return "mdi:alert-outline"
	case NotificationTypeError:
		return "mdi:alert-circle-outline"
	default:
		return "mdi:information-outline"
	}
}

func (t *Toast) accentColor(theme barethemes.Theme, kind NotificationType) color.NRGBA {
	switch kind {
	//case NotificationTypeSuccess:
	//	return theme.Color.Success
	//case NotificationTypeWarning:
	//	return theme.Color.Warning
	//case NotificationTypeError:
	//	return theme.Color.Error
	default:
		return theme.Color.Primary
	}
}

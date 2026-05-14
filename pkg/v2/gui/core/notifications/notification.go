package notifications

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/iconify"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/utils"
)

const (
	defaultQueueSize    = 32
	defaultShowDuration = 4 * time.Second
	defaultMaxVisible   = 4
)

var DefaultNotificationClient = NewClient()

type NotificationType int

const (
	NotificationTypeDebug NotificationType = iota
	NotificationTypeInfo
	NotificationTypeSuccess
	NotificationTypeWarning
	NotificationTypeError
	NotificationTypeOff
)

type Position struct {
	X XPosition
	Y YPosition
}

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

type Notification struct {
	Type     NotificationType
	Title    string
	Message  string
	Messages string
	IconName string        // optional
	Duration time.Duration // if 0 default to client defaults
}

type Client struct {
	notifications chan Notification

	mu         sync.RWMutex
	invalidate func()

	nextID uint64
	items  []item

	defaultShowDuration time.Duration
	MaxVisible          int
	Position            Position

	Theme             *material.Theme
	ThemeClient       *theme.Client
	Animation         *tween.Client
	Iconify           *iconify.Iconify
	Width             unit.Dp
	Margin            unit.Dp
	Gap               unit.Dp
	Radius            unit.Dp
	Padding           unit.Dp
	NotificationLevel NotificationType // only show this or above
}

type item struct {
	id        uint64
	createdAt time.Time
	expiresAt time.Time
	note      Notification

	close widget.Clickable
	anim  *tween.Animation

	entered       bool
	closing       bool
	closeStarted  bool
	lastMeasuredW int
}

func NewClient() *Client {
	return &Client{
		notifications:       make(chan Notification, defaultQueueSize),
		defaultShowDuration: defaultShowDuration,
		MaxVisible:          defaultMaxVisible,
		Position:            TopRight,
		Theme:               material.NewTheme(),
		ThemeClient:         theme.DefaultThemeClient,
		Animation:           tween.NewClient(),
		Iconify:             iconify.DefaultIconify,
		Width:               unit.Dp(390),
		Margin:              unit.Dp(18),
		Gap:                 unit.Dp(10),
		Radius:              unit.Dp(14),
		Padding:             unit.Dp(14),
		NotificationLevel:   NotificationTypeInfo,
	}
}

func (c *Client) WithThemeClient(tc *theme.Client) *Client {
	if c == nil {
		return c
	}
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	c.ThemeClient = tc
	return c
}

func (c *Client) WithMaterialTheme(th *material.Theme) *Client {
	if c == nil {
		return c
	}
	if th != nil {
		c.Theme = th
	}
	return c
}

func (c *Client) WithPosition(pos Position) *Client {
	if c == nil {
		return c
	}
	c.Position = pos
	return c
}

func (c *Client) WithInvalidate(invalidate func()) *Client {
	if c == nil {
		return c
	}
	c.mu.Lock()
	c.invalidate = invalidate
	c.mu.Unlock()
	return c
}

func (c *Client) Send(notification Notification) {
	if c == nil || notification.Type < c.NotificationLevel {
		return
	}

	ch := c.queue()
	select {
	case ch <- notification:
	default:
	}
	c.invalidateWindow()
}

func (c *Client) Layout(gtx layout.Context) layout.Dimensions {
	c.OverlayLayout(gtx)
	return layout.Dimensions{Size: gtx.Constraints.Max}
}

func (c *Client) OverlayLayout(gtx layout.Context) {
	if c == nil {
		return
	}

	c.ensureDefaults()
	c.drainQueue()
	c.updateExpired(time.Now())

	if len(c.items) == 0 {
		return
	}

	if c.ThemeClient.ColorTweenRunning() {
		gtx.Execute(op.InvalidateCmd{})
	}

	screen := gtx.Constraints.Max
	if screen.X <= 0 || screen.Y <= 0 {
		return
	}

	visible := c.visibleItems()
	if len(visible) == 0 {
		return
	}

	margin := gtx.Dp(c.Margin)
	gap := gtx.Dp(c.Gap)
	width := c.notificationWidth(gtx, screen, margin)

	children := make([]layout.FlexChild, 0, len(visible))
	for _, idx := range visible {
		idx := idx
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: utils.PxToDp(gtx, gap)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return c.layoutItem(gtx, &c.items[idx], width)
			})
		}))
	}

	listGtx := gtx
	listGtx.Constraints.Min.X = width
	listGtx.Constraints.Max.X = width
	listGtx.Constraints.Min.Y = 0
	listGtx.Constraints.Max.Y = screen.Y - margin*2

	macro := op.Record(gtx.Ops)
	dims := layout.Flex{Axis: layout.Vertical}.Layout(listGtx, children...)
	call := macro.Stop()

	offset := c.offset(screen, dims.Size, margin)
	stack := op.Offset(offset).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()

	if c.hasAnimatingItems() {
		gtx.Execute(op.InvalidateCmd{})
	}
	c.pruneClosed()
}

func (c *Client) queue() chan Notification {
	if c.notifications != nil {
		return c.notifications
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.notifications == nil {
		c.notifications = make(chan Notification, defaultQueueSize)
	}
	return c.notifications
}

func (c *Client) ensureDefaults() {
	if c.Theme == nil {
		c.Theme = material.NewTheme()
	}
	if c.ThemeClient == nil {
		c.ThemeClient = theme.DefaultThemeClient
	}
	if c.Animation == nil {
		c.Animation = tween.NewClient()
	}
	if c.Iconify == nil {
		c.Iconify = iconify.DefaultIconify
	}
	if c.defaultShowDuration <= 0 {
		c.defaultShowDuration = defaultShowDuration
	}
	if c.MaxVisible <= 0 {
		c.MaxVisible = defaultMaxVisible
	}
	if c.Width <= 0 {
		c.Width = unit.Dp(390)
	}
	if c.Margin <= 0 {
		c.Margin = unit.Dp(18)
	}
	if c.Gap <= 0 {
		c.Gap = unit.Dp(10)
	}
	if c.Radius <= 0 {
		c.Radius = unit.Dp(14)
	}
	if c.Padding <= 0 {
		c.Padding = unit.Dp(14)
	}
}

func (c *Client) drainQueue() {
	for {
		select {
		case n, ok := <-c.queue():
			if !ok {
				return
			}
			if n.Type < c.NotificationLevel {
				continue
			}
			c.nextID++
			now := time.Now()
			duration := n.Duration
			if duration <= 0 {
				duration = c.defaultShowDuration
			}
			anim := c.Animation.NewAnimation(tween.New(190*time.Millisecond, tween.EaseOutCubic))
			c.items = append(c.items, item{
				id:        c.nextID,
				createdAt: now,
				expiresAt: now.Add(duration),
				note:      n,
				anim:      anim,
			})
		default:
			return
		}
	}
}

func (c *Client) updateExpired(now time.Time) {
	for i := range c.items {
		if !c.items[i].closing && now.After(c.items[i].expiresAt) {
			c.items[i].closing = true
		}
	}
}

func (c *Client) visibleItems() []int {
	if c.MaxVisible <= 0 || len(c.items) <= c.MaxVisible {
		out := make([]int, len(c.items))
		for i := range c.items {
			out[i] = i
		}
		return out
	}

	start := len(c.items) - c.MaxVisible
	out := make([]int, 0, c.MaxVisible)
	for i := start; i < len(c.items); i++ {
		out = append(out, i)
	}
	return out
}

func (c *Client) layoutItem(gtx layout.Context, it *item, width int) layout.Dimensions {
	now := time.Now()
	it.lastMeasuredW = width

	for it.close.Clicked(gtx) {
		it.closing = true
	}

	if it.anim != nil && !it.entered {
		it.anim.JumpTo(c.hiddenOffset(width), 0)
		it.anim.MoveTo(0, 0)
		it.entered = true
	}

	if it.anim != nil && it.closing && !it.closeStarted {
		it.anim.MoveTo(c.hiddenOffset(width), 0)
		it.closeStarted = true
	}

	itemGtx := gtx
	itemGtx.Constraints.Min.X = width
	itemGtx.Constraints.Max.X = width

	macro := op.Record(gtx.Ops)
	dims := c.layoutCard(itemGtx, it)
	call := macro.Stop()

	if it.anim == nil {
		call.Add(gtx.Ops)
		return dims
	}

	pt, running := it.anim.Tick(now)
	stack := op.Offset(pt).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()

	if running {
		gtx.Execute(op.InvalidateCmd{})
	}

	return dims
}

func (c *Client) layoutCard(gtx layout.Context, it *item) layout.Dimensions {
	tokens := c.ThemeClient.GetCurrentColorToken()
	accent := c.accentColor(tokens, it.note.Type)
	bg := theme.Mix(accent, tokens.SurfaceNRGBA(), 0.10)
	border := theme.Mix(accent, tokens.BorderNRGBA(), 0.42)

	return utils.ClickableSurfaceOutlined(
		gtx,
		nil,
		bg,
		c.Radius,
		utils.SurfaceBorder{Color: border, Width: unit.Dp(1)},
		func(gtx layout.Context) layout.Dimensions {
			return layout.UniformInset(c.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return c.layoutAccent(gtx, accent)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return c.layoutText(gtx, it, tokens, accent)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return c.layoutClose(gtx, it, tokens)
					}),
				)
			})
		},
	)
}

func (c *Client) layoutAccent(gtx layout.Context, accent color.NRGBA) layout.Dimensions {
	size := image.Pt(gtx.Dp(unit.Dp(4)), gtx.Dp(unit.Dp(56)))
	gtx.Constraints.Min = size
	gtx.Constraints.Max = size
	return utils.Surface(gtx, accent, unit.Dp(2), func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}

func (c *Client) layoutText(gtx layout.Context, it *item, tokens *theme.ColorTokens, accent color.NRGBA) layout.Dimensions {
	iconName := c.iconName(it.note)
	title := c.titleText(it.note)
	message := c.messageText(it.note)

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if iconName == "" || c.Iconify == nil {
						return layout.Dimensions{}
					}
					return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return c.Iconify.Layout(gtx, iconName, unit.Dp(18), accent)
					})
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return theme.ThemedLabel(gtx, c.Theme, c.ThemeClient, theme.TextRoleLabel, theme.ThemeColorTextPrimary, title)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if message == "" {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(5)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body1(c.Theme, message)
				theme.ApplyTypography(&lbl, c.ThemeClient.GetCurrentTypography(), theme.TextRoleBodySmall)
				lbl.Color = tokens.TextSecondaryNRGBA()
				return lbl.Layout(gtx)
			})
		}),
	)
}

func (c *Client) layoutClose(gtx layout.Context, it *item, tokens *theme.ColorTokens) layout.Dimensions {
	size := gtx.Dp(unit.Dp(28))
	closeGtx := gtx
	closeGtx.Constraints.Min = image.Pt(size, size)
	closeGtx.Constraints.Max = image.Pt(size, size)

	return it.close.Layout(closeGtx, func(gtx layout.Context) layout.Dimensions {
		col := tokens.TextMutedNRGBA()
		if it.close.Hovered() || it.close.Pressed() {
			col = tokens.TextPrimaryNRGBA()
		}
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if c.Iconify == nil {
				return layout.Dimensions{Size: image.Pt(size, size)}
			}
			dims := c.Iconify.LayoutWithSize(gtx, "lucide:x", unit.Dp(16), col)
			if dims.Size.X == 0 || dims.Size.Y == 0 {
				return layout.Dimensions{Size: image.Pt(size, size)}
			}
			return dims
		})
	})
}

func (c *Client) notificationWidth(gtx layout.Context, screen image.Point, margin int) int {
	width := gtx.Dp(c.Width)
	maxWidth := screen.X - margin*2
	if maxWidth < 1 {
		maxWidth = 1
	}
	if width <= 0 || width > maxWidth {
		width = maxWidth
	}
	return width
}

func (c *Client) offset(screen, size image.Point, margin int) image.Point {
	x := margin
	switch c.Position.X {
	case Center:
		x = (screen.X - size.X) / 2
	case Right:
		x = screen.X - size.X - margin
	}
	if x < margin {
		x = margin
	}

	y := margin
	switch c.Position.Y {
	case Middle:
		y = (screen.Y - size.Y) / 2
	case Bottom:
		y = screen.Y - size.Y - margin
	}
	if y < margin {
		y = margin
	}

	return image.Pt(x, y)
}

func (c *Client) hiddenOffset(width int) int {
	switch c.Position.X {
	case Left:
		return -width - 24
	default:
		return width + 24
	}
}

func (c *Client) hasAnimatingItems() bool {
	if c.Animation == nil {
		return false
	}
	return c.Animation.Active()
}

func (c *Client) pruneClosed() {
	filtered := c.items[:0]
	for i := range c.items {
		it := &c.items[i]
		if it.closing && it.closeStarted && (it.anim == nil || !it.anim.Active()) {
			if c.Animation != nil && it.anim != nil {
				c.Animation.Remove(it.anim)
			}
			continue
		}
		filtered = append(filtered, *it)
	}
	c.items = filtered
}

func (c *Client) titleText(n Notification) string {
	title := strings.TrimSpace(n.Title)
	if title != "" {
		return title
	}
	switch n.Type {
	case NotificationTypeDebug:
		return "Debug"
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

func ParseLevel(value string) (NotificationType, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug", "all":
		return NotificationTypeDebug, true
	case "info":
		return NotificationTypeInfo, true
	case "success":
		return NotificationTypeSuccess, true
	case "warning", "warnings":
		return NotificationTypeWarning, true
	case "error", "errors":
		return NotificationTypeError, true
	case "off", "none", "disabled":
		return NotificationTypeOff, true
	default:
		return NotificationTypeInfo, false
	}
}

func LevelValue(kind NotificationType) string {
	switch kind {
	case NotificationTypeDebug:
		return "debug"
	case NotificationTypeSuccess:
		return "success"
	case NotificationTypeWarning:
		return "warning"
	case NotificationTypeError:
		return "error"
	case NotificationTypeOff:
		return "off"
	default:
		return "info"
	}
}

func LevelLabel(kind NotificationType) string {
	switch kind {
	case NotificationTypeDebug:
		return "All"
	case NotificationTypeSuccess:
		return "Success and above"
	case NotificationTypeWarning:
		return "Warnings and errors"
	case NotificationTypeError:
		return "Errors only"
	case NotificationTypeOff:
		return "Off"
	default:
		return "Info and above"
	}
}

func (c *Client) messageText(n Notification) string {
	if msg := strings.TrimSpace(n.Message); msg != "" {
		return msg
	}
	return strings.TrimSpace(n.Messages)
}

func (c *Client) iconName(n Notification) string {
	if name := strings.TrimSpace(n.IconName); name != "" {
		return name
	}
	switch n.Type {
	case NotificationTypeDebug:
		return "lucide:bug"
	case NotificationTypeSuccess:
		return "lucide:circle-check"
	case NotificationTypeWarning:
		return "lucide:triangle-alert"
	case NotificationTypeError:
		return "lucide:circle-alert"
	default:
		return "lucide:info"
	}
}

func (c *Client) accentColor(tokens *theme.ColorTokens, kind NotificationType) color.NRGBA {
	if tokens == nil {
		return color.NRGBA{A: 255}
	}
	switch kind {
	case NotificationTypeDebug:
		return tokens.TextMutedNRGBA()
	case NotificationTypeSuccess:
		return tokens.SuccessNRGBA()
	case NotificationTypeWarning:
		return tokens.WarningNRGBA()
	case NotificationTypeError:
		return tokens.DangerNRGBA()
	default:
		return tokens.InfoNRGBA()
	}
}

func (c *Client) invalidateWindow() {
	c.mu.RLock()
	invalidate := c.invalidate
	c.mu.RUnlock()
	if invalidate != nil {
		invalidate()
	}
}

var notificationIDCounter uint64

func NewID(prefix string) string {
	if strings.TrimSpace(prefix) == "" {
		prefix = "notification"
	}
	return fmt.Sprintf("%s-%d", prefix, atomic.AddUint64(&notificationIDCounter, 1))
}

func Send(n Notification) {
	DefaultNotificationClient.Send(n)
}

func Info(message string) {
	Send(Notification{Type: NotificationTypeInfo, Message: message})
}

func Success(message string) {
	Send(Notification{Type: NotificationTypeSuccess, Message: message})
}

func Warning(message string) {
	Send(Notification{Type: NotificationTypeWarning, Message: message})
}

func Error(message string) {
	Send(Notification{Type: NotificationTypeError, Message: message})
}

func Debug(message string) {
	Send(Notification{Type: NotificationTypeDebug, Message: message})
}

func PreloadIcons(ctx context.Context) {
	if DefaultNotificationClient == nil || DefaultNotificationClient.Iconify == nil {
		return
	}
	for _, name := range []string{"lucide:info", "lucide:circle-check", "lucide:triangle-alert", "lucide:circle-alert", "lucide:bug", "lucide:x"} {
		_, _ = DefaultNotificationClient.Iconify.Icon(ctx, name)
	}
}

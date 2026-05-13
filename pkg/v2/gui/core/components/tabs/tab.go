package tabs

import (
	"fmt"
	"image"
	"math"
	"strings"
	"sync/atomic"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	"github.com/DarlingGoose/wgl/pkg/v2/gui/core/animations/tween"
)

type Tab struct {
	ID     string
	Name   string
	Icon   string
	Pinned bool

	Widget Widget
}

type Layout struct {
	id string

	Tabs []Tab

	clickables []widget.Clickable

	current  int
	previous int

	direction int

	Transition *tween.Flip

	Duration time.Duration
	Curve    tween.Curve
}

func New(tabs ...Tab) *Layout {
	id := atomic.AddUint64(&tabLayoutIDCounter, 1)

	l := &Layout{
		id: fmt.Sprintf("tabs-%d", id),

		current:   0,
		previous:  -1,
		direction: 1,

		Duration: 180 * time.Millisecond,
		Curve:    tween.EaseOutCubic,
	}

	l.Transition = tween.NewFlip(l.Duration, l.Curve)
	l.SetTabs(tabs...)

	return l
}

func NewTab(id, name, icon string, w Widget) Tab {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)

	if id == "" {
		id = safeTabID(name)
	}
	if id == "" {
		id = fmt.Sprintf("tab-%d", time.Now().UnixNano())
	}

	return Tab{
		ID:     id,
		Name:   name,
		Icon:   icon,
		Widget: w,
	}
}

func NewTabFunc(id, name, icon string, fn layout.Widget) Tab {
	return NewTab(id, name, icon, WidgetFunc(fn))
}

func (l *Layout) ID() string {
	if l == nil {
		return ""
	}

	return l.id
}

func (l *Layout) SetTabs(tabs ...Tab) {
	if l == nil {
		return
	}

	l.Tabs = normalizeTabs(tabs)
	l.clickables = make([]widget.Clickable, len(l.Tabs))

	if len(l.Tabs) == 0 {
		l.current = -1
		l.previous = -1
		return
	}

	if l.current < 0 || l.current >= len(l.Tabs) {
		l.current = 0
	}

	l.previous = -1
}

func (l *Layout) AddTab(tab Tab) {
	if l == nil {
		return
	}

	l.Tabs = append(l.Tabs, normalizeTab(tab, len(l.Tabs)))
	l.clickables = append(l.clickables, widget.Clickable{})

	if l.current < 0 {
		l.current = 0
	}
}

func (l *Layout) CurrentTab() (Tab, bool) {
	if l == nil || l.current < 0 || l.current >= len(l.Tabs) {
		return Tab{}, false
	}

	return l.Tabs[l.current], true
}

func (t Tab) WithPinned(pinned bool) Tab {
	t.Pinned = pinned
	return t
}

func (l *Layout) CurrentIndex() int {
	if l == nil {
		return -1
	}

	return l.current
}

func (l *Layout) CurrentID() string {
	tab, ok := l.CurrentTab()
	if !ok {
		return ""
	}

	return tab.ID
}

func (l *Layout) Buttons() []Button {
	if l == nil {
		return nil
	}

	buttons := make([]Button, 0, len(l.Tabs))

	for i := range l.Tabs {
		buttons = append(buttons, Button{
			ID:        l.Tabs[i].ID,
			Name:      l.Tabs[i].Name,
			Icon:      l.Tabs[i].Icon,
			Pinned:    l.Tabs[i].Pinned,
			Active:    i == l.current,
			Clickable: &l.clickables[i],
		})
	}

	return buttons
}

// Update checks the persistent button clickables and switches tabs.
// Call this from the parent layout after/before rendering the tab buttons.
func (l *Layout) Update(gtx layout.Context) bool {
	if l == nil {
		return false
	}

	changed := false

	for i := range l.clickables {
		for l.clickables[i].Clicked(gtx) {
			if l.SwitchToIndex(i) {
				changed = true
				gtx.Execute(op.InvalidateCmd{})
			}
		}
	}

	return changed
}

func (l *Layout) SwitchToID(id string) bool {
	if l == nil {
		return false
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return false
	}

	for i := range l.Tabs {
		if strings.EqualFold(l.Tabs[i].ID, id) {
			return l.SwitchToIndex(i)
		}
	}

	return false
}

func (l *Layout) SwitchToName(name string) bool {
	if l == nil {
		return false
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	for i := range l.Tabs {
		if strings.EqualFold(l.Tabs[i].Name, name) {
			return l.SwitchToIndex(i)
		}
	}

	return false
}

func (l *Layout) SwitchToIndex(index int) bool {
	if l == nil || index < 0 || index >= len(l.Tabs) {
		return false
	}

	if index == l.current {
		return false
	}

	l.previous = l.current
	l.direction = 1
	if index < l.current {
		l.direction = -1
	}

	l.current = index

	if l.Transition != nil {
		l.Transition.JumpExpanded(false)
		l.Transition.Expand()
	}

	return true
}

func (l *Layout) Next() bool {
	if l == nil || len(l.Tabs) == 0 {
		return false
	}

	next := l.current + 1
	if next >= len(l.Tabs) {
		next = 0
	}

	return l.SwitchToIndex(next)
}

func (l *Layout) Previous() bool {
	if l == nil || len(l.Tabs) == 0 {
		return false
	}

	prev := l.current - 1
	if prev < 0 {
		prev = len(l.Tabs) - 1
	}

	return l.SwitchToIndex(prev)
}

func (l *Layout) Layout(gtx layout.Context) layout.Dimensions {
	if l == nil || l.current < 0 || l.current >= len(l.Tabs) {
		return layout.Dimensions{}
	}

	now := time.Now()

	progress := 1.0
	running := false
	if l.Transition != nil {
		progress, running = l.Transition.Value(now)
	}

	if running {
		gtx.Execute(op.InvalidateCmd{})
	}

	current := l.Tabs[l.current]
	if current.Widget == nil {
		return layout.Dimensions{}
	}

	if !running || l.previous < 0 || l.previous >= len(l.Tabs) || l.previous == l.current {
		return current.Widget.Layout(gtx)
	}

	previous := l.Tabs[l.previous]
	if previous.Widget == nil {
		return current.Widget.Layout(gtx)
	}

	width := gtx.Constraints.Max.X
	if width <= 0 {
		width = 1
	}

	currentX := mapInt(progress, l.direction*width, 0)
	previousX := mapInt(progress, 0, -l.direction*width)

	// Clip the sliding pages to the available tab content bounds so they
	// do not paint outside the tab panel while animating.
	maxY := gtx.Constraints.Max.Y
	if maxY <= 0 {
		maxY = 1 << 20
	}

	clipStack := clip.Rect{
		Max: image.Pt(width, maxY),
	}.Push(gtx.Ops)
	defer clipStack.Pop()

	return layout.Stack{}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			stack := op.Offset(image.Pt(previousX, 0)).Push(gtx.Ops)
			defer stack.Pop()

			return previous.Widget.Layout(gtx)
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			stack := op.Offset(image.Pt(currentX, 0)).Push(gtx.Ops)
			defer stack.Pop()

			return current.Widget.Layout(gtx)
		}),
	)
}

func normalizeTabs(tabs []Tab) []Tab {
	out := make([]Tab, 0, len(tabs))

	seen := map[string]int{}

	for i, tab := range tabs {
		tab = normalizeTab(tab, i)

		key := strings.ToLower(tab.ID)
		if count := seen[key]; count > 0 {
			tab.ID = fmt.Sprintf("%s-%d", tab.ID, count+1)
			key = strings.ToLower(tab.ID)
		}

		seen[key]++
		out = append(out, tab)
	}

	return out
}

func normalizeTab(tab Tab, index int) Tab {
	tab.ID = strings.TrimSpace(tab.ID)
	tab.Name = strings.TrimSpace(tab.Name)

	if tab.ID == "" {
		tab.ID = safeTabID(tab.Name)
	}
	if tab.ID == "" {
		tab.ID = fmt.Sprintf("tab-%d", index)
	}
	if tab.Name == "" {
		tab.Name = tab.ID
	}

	return tab
}

func safeTabID(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(name))

	lastDash := false

	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == ' ':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
}

func mapInt(progress float64, from, to int) int {
	progress = clamp01(progress)
	return int(math.Round(float64(from) + float64(to-from)*progress))
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}

	if v > 1 {
		return 1
	}

	return v
}

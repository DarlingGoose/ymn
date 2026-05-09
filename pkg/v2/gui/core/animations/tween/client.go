package tween

import (
	"sync"

	"gioui.org/io/input"
	"gioui.org/op"
)

type An interface {
	Animate(op op.Ops, source input.Source)
}

// central client to manage animations
// should keep track and clean them up as they happen
type Client struct {
	mu sync.RWMutex

	animations map[*Animation]struct{}
}

func NewClient() *Client {
	return &Client{
		animations: make(map[*Animation]struct{}),
	}
}

func (c *Client) NewAnimation(tween *Tween) *Animation {
	a := NewAnimation(tween)

	c.mu.Lock()
	c.animations[a] = struct{}{}
	c.mu.Unlock()

	return a
}

func (c *Client) Add(a *Animation) {
	if c == nil || a == nil {
		return
	}

	c.mu.Lock()
	c.animations[a] = struct{}{}
	c.mu.Unlock()
}

func (c *Client) Remove(a *Animation) {
	if c == nil || a == nil {
		return
	}

	c.mu.Lock()
	delete(c.animations, a)
	c.mu.Unlock()
}

func (c *Client) Active() bool {
	if c == nil {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	for a := range c.animations {
		if a.Active() {
			return true
		}
	}

	return false
}

func (c *Client) Len() int {
	if c == nil {
		return 0
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.animations)
}

package overlay

import "sync"

var DefaultController = &Controller{}

type Component interface {
	ID() string
	Open()
	Close()
}

type Controller struct {
	mu     sync.Mutex
	active Component
}

func (c *Controller) SetActive(component Component) {
	if c == nil || component == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.active != nil && c.active.ID() != component.ID() {
		c.active.Close()
	}

	c.active = component
	component.Open()
}

func (c *Controller) Clear(component Component) {
	if c == nil || component == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.active != nil && c.active.ID() == component.ID() {
		c.active = nil
	}
}

func (c *Controller) ActiveID() string {
	if c == nil {
		return ""
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.active == nil {
		return ""
	}

	return c.active.ID()
}

func (c *Controller) CloseActive() {
	if c == nil {
		return
	}

	c.mu.Lock()
	active := c.active
	c.active = nil
	c.mu.Unlock()

	if active != nil {
		active.Close()
	}
}

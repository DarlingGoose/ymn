package toast

import (
	"time"

	"gioui.org/widget"
)

type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
)

type Notification struct {
	Title      string
	Message    string
	Icon       string
	Type       NotificationType
	CloseAfter time.Duration
}

type item struct {
	id        uint64
	createdAt time.Time
	expiresAt time.Time
	close     widget.Clickable
	note      Notification
}

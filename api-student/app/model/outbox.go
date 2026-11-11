package model

import "time"

type OutboxEvent struct {
	ID          int64      `json:"id"`
	EventType   string     `json:"event_type"`
	Payload     string     `json:"payload"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

type Notification struct {
	ID       int64      `json:"id"`
	OutboxID int64      `json:"outbox_id"`
	UserID   int        `json:"user_id"`
	Subject  string     `json:"subject"`
	Body     string     `json:"body"`
	SentAt   *time.Time `json:"sent_at,omitempty"`
}

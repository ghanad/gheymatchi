package notification

import "time"

const (
	ChannelInternal = "internal"
	StatusPending   = "pending"
)

type Notification struct {
	ID        string     `json:"id"`
	AlertID   *string    `json:"alert_id,omitempty"`
	Channel   string     `json:"channel"`
	Recipient string     `json:"recipient"`
	Status    string     `json:"status"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

package notification

import "time"

const (
	ChannelEmail  = "email"
	ChannelSMS    = "sms"
	ChannelDryRun = "dry_run"

	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
)

type Notification struct {
	ID           string     `json:"id"`
	AlertID      *string    `json:"alert_id,omitempty"`
	Channel      string     `json:"channel"`
	Recipient    string     `json:"recipient"`
	Status       string     `json:"status"`
	AttemptCount int        `json:"attempt_count"`
	LastError    *string    `json:"last_error,omitempty"`
	SentAt       *time.Time `json:"sent_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

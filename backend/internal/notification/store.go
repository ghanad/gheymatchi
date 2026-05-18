package notification

import (
	"context"
	"time"
)

type Store interface {
	List(ctx context.Context) ([]Notification, error)
	ListPending(ctx context.Context, limit int) ([]Notification, error)
	MarkSent(ctx context.Context, id string, sentAt time.Time) error
	RecordFailedAttempt(ctx context.Context, id string, message string, maxAttempts int) error
}

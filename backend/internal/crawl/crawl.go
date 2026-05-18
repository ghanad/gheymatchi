package crawl

import "time"

const (
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
)

type Run struct {
	ID           string
	SourceID     string
	Status       string
	StartedAt    time.Time
	FinishedAt   *time.Time
	ErrorMessage *string
	CreatedAt    time.Time
}

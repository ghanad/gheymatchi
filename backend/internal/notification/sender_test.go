package notification

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestProcessorMarksDryRunNotificationsSent(t *testing.T) {
	store := &fakeStore{
		pending: []Notification{{
			ID:        "notification-1",
			Channel:   ChannelDryRun,
			Recipient: "product-1",
			Status:    StatusPending,
		}},
	}
	processor := NewProcessor(store, DryRunSender{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}, 3)
	sentAt := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	processor.now = func() time.Time { return sentAt }

	if err := processor.ProcessPending(context.Background()); err != nil {
		t.Fatalf("ProcessPending() error = %v", err)
	}
	if store.sentID != "notification-1" {
		t.Fatalf("sentID = %q, want notification-1", store.sentID)
	}
	if !store.sentAt.Equal(sentAt) {
		t.Fatalf("sentAt = %v, want %v", store.sentAt, sentAt)
	}
}

func TestProcessorMarksUnsupportedChannelFailed(t *testing.T) {
	store := &fakeStore{
		pending: []Notification{{
			ID:        "notification-1",
			Channel:   ChannelEmail,
			Recipient: "user@example.com",
			Status:    StatusPending,
		}},
	}
	processor := NewProcessor(store, DryRunSender{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}, 3)

	if err := processor.ProcessPending(context.Background()); err != nil {
		t.Fatalf("ProcessPending() error = %v", err)
	}
	if store.failedID != "notification-1" {
		t.Fatalf("failedID = %q, want notification-1", store.failedID)
	}
	if store.maxAttempts != 3 {
		t.Fatalf("maxAttempts = %d, want 3", store.maxAttempts)
	}
	if store.sentID != "" {
		t.Fatalf("sentID = %q, want empty", store.sentID)
	}
}

func TestRoutingSenderSendsEmailNotifications(t *testing.T) {
	email := &fakeEmailSender{}
	sender := NewRoutingSender(
		NewDryRunSender(slog.New(slog.NewTextHandler(io.Discard, nil))),
		email,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	alertID := "alert-1"

	err := sender.Send(context.Background(), Notification{
		ID:        "notification-1",
		AlertID:   &alertID,
		Channel:   ChannelEmail,
		Recipient: "user@example.com",
		Status:    StatusPending,
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if email.message.To != "user@example.com" {
		t.Fatalf("email recipient = %q, want user@example.com", email.message.To)
	}
	if email.message.Subject == "" {
		t.Fatal("email subject is empty")
	}
	if email.message.Body == "" {
		t.Fatal("email body is empty")
	}
}

type fakeStore struct {
	pending       []Notification
	listErr       error
	sentID        string
	sentAt        time.Time
	failedID      string
	failedMessage string
	maxAttempts   int
}

func (f *fakeStore) List(ctx context.Context) ([]Notification, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeStore) ListPending(ctx context.Context, limit int) ([]Notification, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.pending, nil
}

func (f *fakeStore) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	f.sentID = id
	f.sentAt = sentAt
	return nil
}

func (f *fakeStore) RecordFailedAttempt(ctx context.Context, id string, message string, maxAttempts int) error {
	f.failedID = id
	f.failedMessage = message
	f.maxAttempts = maxAttempts
	return nil
}

type fakeEmailSender struct {
	message EmailMessage
}

func (f *fakeEmailSender) SendEmail(ctx context.Context, message EmailMessage) error {
	f.message = message
	return nil
}

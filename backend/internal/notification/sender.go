package notification

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Sender interface {
	Send(ctx context.Context, notification Notification) error
}

type DryRunSender struct {
	logger *slog.Logger
}

func NewDryRunSender(logger *slog.Logger) DryRunSender {
	return DryRunSender{logger: logger}
}

func (s DryRunSender) Send(ctx context.Context, notification Notification) error {
	if notification.Channel != ChannelDryRun {
		return fmt.Errorf("dry-run sender cannot send channel %q", notification.Channel)
	}
	s.logger.Info(
		"dry-run notification sent",
		slog.String("notification_id", notification.ID),
		slog.String("recipient", notification.Recipient),
		slog.String("status", notification.Status),
	)
	return nil
}

type Processor struct {
	store  Store
	sender Sender
	now    func() time.Time
}

func NewProcessor(store Store, sender Sender) Processor {
	return Processor{
		store:  store,
		sender: sender,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (p Processor) ProcessPending(ctx context.Context) error {
	notifications, err := p.store.ListPending(ctx, 50)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if err := p.sender.Send(ctx, notification); err != nil {
			if markErr := p.store.MarkFailed(ctx, notification.ID); markErr != nil {
				return fmt.Errorf("mark notification failed after send error %w: %w", err, markErr)
			}
			continue
		}
		if err := p.store.MarkSent(ctx, notification.ID, p.now()); err != nil {
			return err
		}
	}

	return nil
}

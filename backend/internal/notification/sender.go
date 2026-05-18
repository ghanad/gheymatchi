package notification

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Sender interface {
	Send(ctx context.Context, notification Notification) error
}

type EmailMessage struct {
	To      string
	Subject string
	Body    string
}

type EmailSender interface {
	SendEmail(ctx context.Context, message EmailMessage) error
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

type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPSender(host string, port string, username string, password string, from string) (SMTPSender, error) {
	if strings.TrimSpace(host) == "" {
		return SMTPSender{}, fmt.Errorf("SMTP_HOST is required")
	}
	if strings.TrimSpace(port) == "" {
		return SMTPSender{}, fmt.Errorf("SMTP_PORT is required")
	}
	if strings.TrimSpace(from) == "" {
		return SMTPSender{}, fmt.Errorf("SMTP_FROM is required")
	}
	if hasHeaderBreak(from) {
		return SMTPSender{}, fmt.Errorf("SMTP_FROM contains invalid characters")
	}
	return SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}, nil
}

func (s SMTPSender) SendEmail(ctx context.Context, message EmailMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.TrimSpace(message.To) == "" {
		return fmt.Errorf("email recipient is required")
	}
	if hasHeaderBreak(message.To) || hasHeaderBreak(message.Subject) {
		return fmt.Errorf("email message contains invalid header characters")
	}

	addr := net.JoinHostPort(s.host, s.port)
	var auth smtp.Auth
	if s.username != "" || s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	raw := strings.Join([]string{
		"From: " + s.from,
		"To: " + message.To,
		"Subject: " + message.Subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		message.Body,
	}, "\r\n")

	if err := smtp.SendMail(addr, auth, s.from, []string{message.To}, []byte(raw)); err != nil {
		return fmt.Errorf("send smtp email: %w", err)
	}
	return nil
}

func hasHeaderBreak(value string) bool {
	return strings.ContainsAny(value, "\r\n")
}

type RoutingSender struct {
	dryRun DryRunSender
	email  EmailSender
	logger *slog.Logger
}

func NewRoutingSender(dryRun DryRunSender, email EmailSender, logger *slog.Logger) RoutingSender {
	return RoutingSender{dryRun: dryRun, email: email, logger: logger}
}

func (s RoutingSender) Send(ctx context.Context, notification Notification) error {
	switch notification.Channel {
	case ChannelDryRun:
		return s.dryRun.Send(ctx, notification)
	case ChannelEmail:
		if s.email == nil {
			return fmt.Errorf("email sender is not configured")
		}
		message := EmailMessage{
			To:      notification.Recipient,
			Subject: "GheymatChi price alert",
			Body:    emailBody(notification),
		}
		if err := s.email.SendEmail(ctx, message); err != nil {
			return err
		}
		s.logger.Info(
			"email notification sent",
			slog.String("notification_id", notification.ID),
			slog.String("recipient", notification.Recipient),
		)
		return nil
	default:
		return fmt.Errorf("unsupported notification channel %q", notification.Channel)
	}
}

func emailBody(notification Notification) string {
	alertID := "unknown"
	if notification.AlertID != nil {
		alertID = *notification.AlertID
	}
	return fmt.Sprintf("A GheymatChi price alert was triggered.\n\nAlert ID: %s\nNotification ID: %s\n", alertID, notification.ID)
}

type Processor struct {
	store       Store
	sender      Sender
	maxAttempts int
	now         func() time.Time
}

func NewProcessor(store Store, sender Sender, maxAttempts int) Processor {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	return Processor{
		store:       store,
		sender:      sender,
		maxAttempts: maxAttempts,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func (p Processor) ProcessPending(ctx context.Context) error {
	notifications, err := p.store.ListPending(ctx, 50)
	if err != nil {
		return err
	}

	for _, notification := range notifications {
		if err := p.sender.Send(ctx, notification); err != nil {
			if markErr := p.store.RecordFailedAttempt(ctx, notification.ID, err.Error(), p.maxAttempts); markErr != nil {
				return fmt.Errorf("record notification failure after send error %w: %w", err, markErr)
			}
			continue
		}
		if err := p.store.MarkSent(ctx, notification.ID, p.now()); err != nil {
			return err
		}
	}

	return nil
}

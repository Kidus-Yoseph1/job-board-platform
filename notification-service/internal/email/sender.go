package email

import (
	"context"
	"net/smtp"

	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/logger"
)

// EmailSender defines the interface for sending emails
type EmailSender interface {
	Send(ctx context.Context, to string, subject string, body string) error
}

// SMTPSender implements EmailSender using standard SMTP protocol
type SMTPSender struct {
	host      string
	port      string
	username  string
	password  string
	fromEmail string
	log       *logger.Logger
}

// NewSMTPSender creates a new instance of SMTPSender
func NewSMTPSender(host, port, username, password, fromEmail string, log *logger.Logger) *SMTPSender {
	return &SMTPSender{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromEmail: fromEmail,
		log:       log,
	}
}

// Send transmits the email over SMTP. If host is empty or localhost, it logs the content and skips real delivery for local development.
func (s *SMTPSender) Send(ctx context.Context, to string, subject string, body string) error {
	s.log.Infow("initiating email delivery", "to", to, "subject", subject)

	// If host is empty or set to localhost, we print it to logs and skip real transmission
	if s.host == "" || s.host == "localhost" {
		s.log.Infow("[LOCAL MODE] skipping actual SMTP sending. Rendered email:", "to", to, "subject", subject, "body", body)
		return nil
	}

	// Draft standard MIME HTML message headers
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subjectHeader := "Subject: " + subject + "\n"
	fromHeader := "From: " + s.fromEmail + "\n"
	toHeader := "To: " + to + "\n"
	msg := []byte(fromHeader + toHeader + subjectHeader + mime + body)

	addr := s.host + ":" + s.port

	// Authenticate only if username is configured
	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, msg)
	if err != nil {
		s.log.Errorw("failed to deliver email over SMTP", "error", err, "to", to, "subject", subject)
		return err
	}

	s.log.Infow("email successfully delivered over SMTP", "to", to, "subject", subject)
	return nil
}

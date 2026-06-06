package worker

import (
	"context"
	"fmt"
	"net/smtp"
	"time"

	db "github.com/kidus-yoseph1/job-board-platform/notification-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/repository"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/logger"
)

type EmailWorker struct {
	repo         *repository.EmailJobRepo
	host         string
	port         string
	username     string
	password     string
	fromEmail    string
	pollInterval time.Duration
	log          *logger.Logger
}

func NewEmailWorker(
	repo *repository.EmailJobRepo,
	host, port, username, password, fromEmail string,
) *EmailWorker {
	return &EmailWorker{
		repo:         repo,
		host:         host,
		port:         port,
		username:     username,
		password:     password,
		fromEmail:    fromEmail,
		pollInterval: 5 * time.Second,
		log:          logger.Get(),
	}
}

func (w *EmailWorker) Start(ctx context.Context) {
	w.log.Infow("starting background email worker", "pollInterval", w.pollInterval)
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("stopping background email worker gracefully...")
			return
		case <-ticker.C:
			w.processPendingJobs(ctx)
		}
	}
}

func (w *EmailWorker) processPendingJobs(ctx context.Context) {
	// Fetch up to 10 pending email jobs
	jobs, err := w.repo.GetPendingEmailJobs(ctx, 10)
	if err != nil {
		w.log.Errorw("failed to fetch pending email jobs", "error", err)
		return
	}

	if len(jobs) > 0 {
		w.log.Infow("found pending email jobs", "count", len(jobs))
	}

	for _, job := range jobs {
		// Attempt to send email
		err := w.sendEmail(job.ToEmail, job.Subject, job.Body)

		var newStatus string
		if err != nil {
			w.log.Errorw("failed to send email", "error", err, "jobID", job.ID, "toEmail", job.ToEmail)
			newStatus = "failed"
		} else {
			w.log.Infow("email sent successfully", "jobID", job.ID, "toEmail", job.ToEmail)
			newStatus = "sent"
		}

		// Update job status in database
		updateParams := db.UpdateEmailJobStatusParams{
			Status: newStatus,
			ID:     job.ID,
		}
		if updateErr := w.repo.UpdateEmailJobStatus(ctx, updateParams); updateErr != nil {
			w.log.Errorw("failed to update email job status", "error", updateErr, "jobID", job.ID)
		}
	}
}

func (w *EmailWorker) sendEmail(to, subject, body string) error {
	// Skip sending if host is "localhost" and we don't have a real SMTP server configured (development mode mock)
	// We just simulate success.
	if w.host == "localhost" && w.username == "" {
		w.log.Infow("development mode: simulating email send", "to", to, "subject", subject)
		time.Sleep(100 * time.Millisecond) // Simulate network delay
		return nil
	}

	auth := smtp.PlainAuth("", w.username, w.password, w.host)
	addr := fmt.Sprintf("%s:%s", w.host, w.port)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, w.fromEmail, subject, body))

	return smtp.SendMail(addr, auth, w.fromEmail, []string{to}, msg)
}

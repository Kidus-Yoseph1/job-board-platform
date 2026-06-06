package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/kidus-yoseph1/job-board-platform/notification-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/repository"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/logger"
)

type EmailService struct {
	emailRepo *repository.EmailJobRepo
	log       *logger.Logger
}

func NewEmailService(emailRepo *repository.EmailJobRepo, log *logger.Logger) *EmailService {
	return &EmailService{emailRepo: emailRepo, log: log}
}

func (s *EmailService) CreateEmailJob(ctx context.Context, toEmail string, subject string, body string) (*db.EmailJob, error) {
	s.log.Infow("attempting to create email job", "toEmail", toEmail, "subject", subject)

	emailParams := db.CreateEmailJobParams{
		ToEmail: toEmail,
		Subject: subject,
		Body:    body,
	}
	email, err := s.emailRepo.CreateEmailJob(ctx, emailParams)
	if err != nil {
		s.log.Errorw("database error creating email job", "error", err, "toEmail", toEmail)
		return nil, domain.ErrInternal("creating email failed")
	}
	
	s.log.Infow("email job created successfully", "emailJobID", email.ID)
	return email, nil
}

func (s *EmailService) UpdateEmailJobStatus(ctx context.Context, status string, id uuid.UUID) error {
	s.log.Infow("attempting to update email job status", "emailJobID", id, "status", status)

	updatedEmailParams := db.UpdateEmailJobStatusParams{
		Status: status,
		ID:     id,
	}

	err := s.emailRepo.UpdateEmailJobStatus(ctx, updatedEmailParams)
	if err != nil {
		s.log.Errorw("database error updating email job status", "error", err, "emailJobID", id)
		return domain.ErrInternal("email status update failed ")
	}
	
	s.log.Infow("email job status updated successfully", "emailJobID", id)
	return nil
}

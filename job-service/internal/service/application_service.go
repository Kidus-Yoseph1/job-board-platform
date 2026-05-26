package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/repository"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/logger"
)

type ApplicationService struct {
	ApplicationRepo *repository.ApplicationRepo
	log             *logger.Logger
}

func NewApplicationService(ApplicationRepo *repository.ApplicationRepo) *ApplicationService {
	return &ApplicationService{
		ApplicationRepo: ApplicationRepo,
		log:             logger.Get(),
	}
}

func (s *ApplicationService) CreateApplication(ctx context.Context, jobID uuid.UUID, userID uuid.UUID, coverLetter string) (*db.Application, error) {
	s.log.Infow("attempting to create application", "jobID", jobID, "userID", userID)

	applicationParams := db.CreateApplicationParams{
		JobID:  jobID,
		UserID: userID,
		CoverLetter: sql.NullString{
			String: coverLetter,
			Valid:  coverLetter != "",
		},
	}

	application, err := s.ApplicationRepo.CreateApplication(ctx, applicationParams)
	if err != nil {
		s.log.Errorw("database error creating application", "error", err, "jobID", jobID, "userID", userID)
		return nil, domain.ErrInternal("something went wrong")
	}

	s.log.Infow("application created successfully", "applicationID", application.ID)
	return application, nil
}

func (s *ApplicationService) GetApplicationByID(ctx context.Context, id uuid.UUID) (*db.Application, error) {
	s.log.Infow("attempting to get application by id", "applicationID", id)

	application, err := s.ApplicationRepo.GetApplicationByID(ctx, id)
	if err != nil {
		s.log.Errorw("database error fetching application", "error", err, "applicationID", id)
		return nil, domain.ErrInternal("something went wrong")
	}
	if application == nil {
		s.log.Warnw("application not found", "applicationID", id)
		return nil, domain.ErrNotFound("application not found")
	}

	return application, nil
}

func (s *ApplicationService) ListApplicationsByJob(ctx context.Context, jobID uuid.UUID) ([]db.Application, error) {
	s.log.Infow("attempting to list applications by job", "jobID", jobID)

	applications, err := s.ApplicationRepo.ListApplicationsByJob(ctx, jobID)
	if err != nil {
		s.log.Errorw("database error listing applications by job", "error", err, "jobID", jobID)
		return nil, domain.ErrInternal("something went wrong")
	}
	if applications == nil {
		s.log.Warnw("no applications found for job", "jobID", jobID)
		return nil, domain.ErrNotFound("applications not found")
	}

	return applications, nil
}

func (s *ApplicationService) ListApplicationsByUser(ctx context.Context, userID uuid.UUID) ([]db.Application, error) {
	s.log.Infow("attempting to list applications by user", "userID", userID)

	applications, err := s.ApplicationRepo.ListApplicationsByUser(ctx, userID)
	if err != nil {
		s.log.Errorw("database error listing applications by user", "error", err, "userID", userID)
		return nil, domain.ErrInternal("something went wrong")
	}
	if applications == nil {
		s.log.Warnw("no applications found for user", "userID", userID)
		return nil, domain.ErrNotFound("applications not found")
	}

	return applications, nil
}

func (s *ApplicationService) UpdateApplicationStatus(ctx context.Context, id uuid.UUID, status string) (*db.Application, error) {
	s.log.Infow("attempting to update application status", "applicationID", id, "status", status)

	updatedApplicationParams := db.UpdateApplicationStatusParams{
		Status: status,
		ID:     id,
	}

	updatedApplication, err := s.ApplicationRepo.UpdateApplicationStatus(ctx, updatedApplicationParams)
	if err != nil {
		s.log.Errorw("database error updating application status", "error", err, "applicationID", id)
		return nil, domain.ErrInternal("something went wrong")
	}

	s.log.Infow("application status updated successfully", "applicationID", id)
	return updatedApplication, nil
}

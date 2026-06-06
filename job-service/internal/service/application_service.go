package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/cache"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/repository"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/logger"
)

type ApplicationService struct {
	ApplicationRepo *repository.ApplicationRepo
	JobRepo         *repository.JobRepo
	CompanyRepo     *repository.CompanyRepo
	redisCache      *cache.RedisCache
	log             *logger.Logger
}

func NewApplicationService(
	ApplicationRepo *repository.ApplicationRepo,
	JobRepo *repository.JobRepo,
	CompanyRepo *repository.CompanyRepo,
	redisCache *cache.RedisCache,
) *ApplicationService {
	return &ApplicationService{
		ApplicationRepo: ApplicationRepo,
		JobRepo:         JobRepo,
		CompanyRepo:     CompanyRepo,
		redisCache:      redisCache,
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

	// Publish job_applied event to Redis
	go func() {
		// Fetch job and company owner
		job, err := s.JobRepo.GetJobByID(context.Background(), jobID)
		if err != nil || job == nil {
			s.log.Errorw("failed to fetch job for pubsub (not found or error)", "error", err, "jobID", jobID)
			return
		}
		company, err := s.CompanyRepo.GetCompanyByID(context.Background(), job.CompanyID)
		if err != nil || company == nil {
			s.log.Errorw("failed to fetch company for pubsub (not found or error)", "error", err, "companyID", job.CompanyID)
			return
		}

		event := map[string]interface{}{
			"event_type": "job_applied",
			"payload": map[string]interface{}{
				"target_user_id": company.UserID.String(),
				"title":          "New Application Received",
				"message":        fmt.Sprintf("A new candidate has applied for your job posting: %s", job.Title),
			},
		}

		if err := s.redisCache.Publish(context.Background(), "job_events", event); err != nil {
			s.log.Errorw("failed to publish job_applied event to redis", "error", err)
		} else {
			s.log.Infow("job_applied event published to redis", "jobID", jobID, "targetUserID", company.UserID)
		}
	}()

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

	// Publish application_status_changed event to Redis
	go func() {
		job, err := s.JobRepo.GetJobByID(context.Background(), updatedApplication.JobID)
		if err != nil || job == nil {
			s.log.Errorw("failed to fetch job for pubsub status change (not found or error)", "error", err, "jobID", updatedApplication.JobID)
			return
		}

		event := map[string]interface{}{
			"event_type": "application_status_changed",
			"payload": map[string]interface{}{
				"target_user_id": updatedApplication.UserID.String(),
				"title":          "Application Status Updated",
				"message":        fmt.Sprintf("Your application for '%s' has been updated to: %s", job.Title, status),
			},
		}

		if err := s.redisCache.Publish(context.Background(), "job_events", event); err != nil {
			s.log.Errorw("failed to publish application_status_changed event to redis", "error", err)
		} else {
			s.log.Infow("application_status_changed event published to redis", "applicationID", id, "targetUserID", updatedApplication.UserID)
		}
	}()

	return updatedApplication, nil
}

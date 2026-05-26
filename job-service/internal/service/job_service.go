package service

import (
	"context"

	"github.com/google/uuid"
	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/repository"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/logger"
)

type JobService struct {
	JobRepo *repository.JobRepo
	log     *logger.Logger
}

func NewJobService(JobRepo *repository.JobRepo) *JobService {
	return &JobService{
		JobRepo: JobRepo,
		log:     logger.Get(),
	}
}

func (s *JobService) CreateJob(ctx context.Context, title string, description string,
	category string, location string, jobType string) (*db.Job, error) {
	s.log.Infow("attempting to create new job", "title", title, "category", category)

	jobParams := db.CreateJobParams{
		Title:        title,
		Description:  description,
		Category:     category,
		Location:     location,
		Type:         jobType,
		IsNegotiable: true,
	}
	job, err := s.JobRepo.CreateJob(ctx, jobParams)
	if err != nil {
		s.log.Errorw("failed to create job in database", "error", err, "title", title)
		return nil, domain.ErrInternal("something went wrong")
	}

	s.log.Infow("job created successfully", "jobID", job.ID)
	return job, nil
}

func (s *JobService) GetJobByID(ctx context.Context, id uuid.UUID) (*db.Job, error) {
	s.log.Infow("attempting to get job by id", "jobID", id)

	job, err := s.JobRepo.GetJobByID(ctx, id)
	if err != nil {
		s.log.Errorw("database error fetching job", "error", err, "jobID", id)
		return nil, domain.ErrInternal("something went wrong")
	}
	if job == nil {
		s.log.Warnw("job not found", "jobID", id)
		return nil, domain.ErrNotFound("job not found")
	}
	return job, nil
}

func (s *JobService) ListJobs(ctx context.Context, requestedLimit int32, offset int32) ([]db.Job, error) {
	limit := requestedLimit
	if limit < 0 || limit > 50 {
		limit = 10
	}
	s.log.Infow("attempting to list jobs", "limit", limit, "offset", offset)

	params := db.ListJobsParams{
		Limit:  limit,
		Offset: offset,
	}
	jobs, err := s.JobRepo.ListJobs(ctx, params)
	if err != nil {
		s.log.Errorw("database error listing jobs", "error", err)
		return nil, domain.ErrInternal("something went wrong")
	}
	return jobs, nil
}

func (s *JobService) ListJobsByCompany(ctx context.Context, companyID uuid.UUID) ([]db.Job, error) {
	s.log.Infow("attempting to list jobs by company", "companyID", companyID)

	jobs, err := s.JobRepo.ListJobsByCompany(ctx, companyID)
	if err != nil {
		s.log.Errorw("database error listing jobs by company", "error", err, "companyID", companyID)
		return nil, domain.ErrInternal("something went wrong")
	}
	if jobs == nil {
		s.log.Warnw("no jobs found for company", "companyID", companyID)
		return nil, domain.ErrNotFound("job not found")
	}
	return jobs, nil
}

func (s *JobService) UpdateJobStatus(ctx context.Context, id uuid.UUID, status string) (*db.Job, error) {
	s.log.Infow("attempting to update job status", "jobID", id, "status", status)

	updatedJobParams := db.UpdateJobStatusParams{
		ID:     id,
		Status: status,
	}

	updatedJob, err := s.JobRepo.UpdateJobStatus(ctx, updatedJobParams)
	if err != nil {
		s.log.Errorw("database error updating job status", "error", err, "jobID", id)
		return nil, domain.ErrInternal("failed to update job status")
	}

	s.log.Infow("job status updated successfully", "jobID", id)
	return updatedJob, nil
}

func (s *JobService) DeleteJob(ctx context.Context, id uuid.UUID) error {
	s.log.Infow("attempting to delete job", "jobID", id)

	err := s.JobRepo.DeleteJob(ctx, id)
	if err != nil {
		s.log.Errorw("database error deleting job", "error", err, "jobID", id)
		return domain.ErrInternal("failed to delete job")
	}

	s.log.Infow("job deleted successfully", "jobID", id)
	return nil
}

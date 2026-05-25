package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
)

type JobRepo struct {
	queries db.Querier
}

func NewJobRepo(queries db.Querier) *JobRepo {
	return &JobRepo{queries: queries}
}

func (r *JobRepo) CreateJob(ctx context.Context, arg db.CreateJobParams) (*db.Job, error) {
	job, err := r.queries.CreateJob(ctx, arg)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *JobRepo) DeleteJob(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteJob(ctx, id)
	return err
}

func (r *JobRepo) GetJobByID(ctx context.Context, id uuid.UUID) (*db.Job, error) {
	job, err := r.queries.GetJobByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *JobRepo) ListJobs(ctx context.Context, arg db.ListJobsParams) ([]db.Job, error) {
	jobs, err := r.queries.ListJobs(ctx, arg)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *JobRepo) ListJobsByCompany(ctx context.Context, companyID uuid.UUID) ([]db.Job, error) {
	jobs, err := r.queries.ListJobsByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *JobRepo) UpdateJobStatus(ctx context.Context, arg db.UpdateJobStatusParams) (*db.Job, error) {
	job, err := r.queries.UpdateJobStatus(ctx, arg)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

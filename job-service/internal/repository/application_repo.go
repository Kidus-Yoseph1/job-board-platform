package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
)

type ApplicationRepo struct {
	queries db.Querier
}

func NewApplicationRepo(queries db.Querier) *ApplicationRepo {
	return &ApplicationRepo{queries: queries}
}

func (r *ApplicationRepo) CreateApplication(ctx context.Context, arg db.CreateApplicationParams) (*db.Application, error) {
	app, err := r.queries.CreateApplication(ctx, arg)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *ApplicationRepo) GetApplicationByID(ctx context.Context, id uuid.UUID) (*db.Application, error) {
	app, err := r.queries.GetApplicationByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *ApplicationRepo) ListApplicationsByJob(ctx context.Context, jobID uuid.UUID) ([]db.Application, error) {
	apps, err := r.queries.ListApplicationsByJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *ApplicationRepo) ListApplicationsByUser(ctx context.Context, userID uuid.UUID) ([]db.Application, error) {
	apps, err := r.queries.ListApplicationsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *ApplicationRepo) UpdateApplicationStatus(ctx context.Context, arg db.UpdateApplicationStatusParams) (*db.Application, error) {
	app, err := r.queries.UpdateApplicationStatus(ctx, arg)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &app, nil
}

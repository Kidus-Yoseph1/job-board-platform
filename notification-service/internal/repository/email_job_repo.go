package repository

import (
	"context"
	"database/sql"

	db "github.com/kidus-yoseph1/job-board-platform/notification-service/db/generated"
)

type EmailJobRepo struct {
	queries db.Querier
}

func NewEmailJobRepo(queries db.Querier) *EmailJobRepo {
	return &EmailJobRepo{queries: queries}
}

func (r *EmailJobRepo) CreateEmailJob(ctx context.Context, arg db.CreateEmailJobParams) (*db.EmailJob, error) {
	emailJob, err := r.queries.CreateEmailJob(ctx, arg)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &emailJob, nil
}

func (r *EmailJobRepo) UpdateEmailJobStatus(ctx context.Context, arg db.UpdateEmailJobStatusParams) error {
	err := r.queries.UpdateEmailJobStatus(ctx, arg)
	return err
}

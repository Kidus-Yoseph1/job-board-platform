package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
)

type CompanyRepo struct {
	queries db.Querier
}

func NewCompanyRepo(queries db.Querier) *CompanyRepo {
	return &CompanyRepo{queries: queries}
}

func (r *CompanyRepo) CreateCompany(ctx context.Context, arg db.CreateCompanyParams) (*db.Company, error) {
	company, err := r.queries.CreateCompany(ctx, arg)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepo) GetCompanyByID(ctx context.Context, id uuid.UUID) (*db.Company, error) {
	company, err := r.queries.GetCompanyByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepo) GetCompanyByUserID(ctx context.Context, userID uuid.UUID) (*db.Company, error) {
	company, err := r.queries.GetCompanyByUserID(ctx, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &company, nil
}

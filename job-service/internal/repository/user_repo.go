package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/kidus-yoseph1/job-board-platform/job-service/db/generated"
)

type UserRepo struct {
	queries db.Querier
}

func NewUserRepo(queries db.Querier) *UserRepo {
	return &UserRepo{queries: queries}
}

func (r *UserRepo) CreateUser(ctx context.Context, arg db.CreateUserParams) (*db.User, error) {
	user, err := r.queries.CreateUser(ctx, arg)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*db.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*db.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

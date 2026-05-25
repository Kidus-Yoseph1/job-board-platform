package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	db "github.com/kidus-yoseph1/job-board-platform/notification-service/db/generated"
)

type NotificationRepo struct {
	queries db.Querier
}

func NewNotificationRepo(queries db.Querier) *NotificationRepo {
	return &NotificationRepo{queries: queries}
}

func (r *NotificationRepo) CreateNotification(ctx context.Context, arg db.CreateNotificationParams) (*db.Notification, error) {
	notification, err := r.queries.CreateNotification(ctx, arg)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepo) ListNotificationsByUser(ctx context.Context, arg db.ListNotificationsByUserParams) ([]db.Notification, error) {
	notifications, err := r.queries.ListNotificationsByUser(ctx, arg)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *NotificationRepo) MarkAllNotificationsRead(ctx context.Context, userID string) error {
	err := r.queries.MarkAllNotificationsRead(ctx, userID)
	return err
}

func (r *NotificationRepo) MarkNotificationRead(ctx context.Context, id uuid.UUID) error {
	err := r.queries.MarkNotificationRead(ctx, id)
	return err
}

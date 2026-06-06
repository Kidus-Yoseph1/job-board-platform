package service

import (
	"context"

	"github.com/google/uuid"

	db "github.com/kidus-yoseph1/job-board-platform/notification-service/db/generated"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/repository"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/logger"
)

type NotificationService struct {
	NotificationRepo *repository.NotificationRepo
	log              *logger.Logger
}

func NewNotificationService(NotificationRepo *repository.NotificationRepo, log *logger.Logger) *NotificationService {
	return &NotificationService{NotificationRepo: NotificationRepo, log: log}
}

func (s *NotificationService) CreateNotification(ctx context.Context, userID string,
	notificationType string, title string, body string) (*db.Notification, error) {
	s.log.Infow("creating new notification", "userID", userID, "type", notificationType)

	notificationParams := db.CreateNotificationParams{
		UserID: userID,
		Type:   notificationType,
		Title:  title,
		Body:   body,
	}

	notification, err := s.NotificationRepo.CreateNotification(ctx, notificationParams)
	if err != nil {
		s.log.Errorw("failed to create notification", "error", err, "userID", userID)
		return nil, domain.ErrInternal("something went wrong")
	}
	
	s.log.Infow("successfully created notification", "notificationID", notification.ID)
	return notification, nil
}

func (s *NotificationService) ListNotificationsByUser(ctx context.Context, userID string, requestedLimit int32, offset int32) ([]db.Notification, error) {
	s.log.Infow("listing notifications for user", "userID", userID, "limit", requestedLimit, "offset", offset)
	
	limit := requestedLimit
	if requestedLimit < 0 || requestedLimit > 10 {
		limit = 5
	}
	notificationParams := db.ListNotificationsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}
	notifications, err := s.NotificationRepo.ListNotificationsByUser(ctx, notificationParams)
	if err != nil {
		s.log.Errorw("failed to list notifications", "error", err, "userID", userID)
		return nil, domain.ErrInternal("something went wrong")
	}
	return notifications, nil
}

func (s *NotificationService) MarkAllNotificationsRead(ctx context.Context, userID string) error {
	s.log.Infow("marking all notifications as read", "userID", userID)
	
	err := s.NotificationRepo.MarkAllNotificationsRead(ctx, userID)
	if err != nil {
		s.log.Errorw("failed to mark all notifications as read", "error", err, "userID", userID)
		return domain.ErrInternal("something went wrong")
	}
	return nil
}

func (s *NotificationService) MarkNotificationRead(ctx context.Context, id uuid.UUID) error {
	s.log.Infow("marking notification as read", "notificationID", id)
	
	err := s.NotificationRepo.MarkNotificationRead(ctx, id)
	if err != nil {
		s.log.Errorw("failed to mark notification as read", "error", err, "notificationID", id)
		return domain.ErrInternal("something went wrong")
	}
	return nil
}

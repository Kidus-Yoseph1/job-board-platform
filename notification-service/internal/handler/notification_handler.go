package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/internal/service"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/logger"
	"github.com/kidus-yoseph1/job-board-platform/notification-service/pkg/response"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
	log                 *logger.Logger
}

func NewNotificationHandler(notificationService *service.NotificationService, log *logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		log:                 log,
	}
}

func (h *NotificationHandler) ListNotificationsByUserHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract user_id from the JWT token context
	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, 401, "unauthorized")
		return
	}

	limitStr := c.DefaultQuery("limit", "5")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		response.Error(c, 400, "invalid limit parameter")
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		response.Error(c, 400, "invalid offset parameter")
		return
	}

	h.log.Infow("handling ListNotificationsByUser request", "userID", userID, "limit", limit, "offset", offset)

	notifications, err := h.notificationService.ListNotificationsByUser(ctx, userID, int32(limit), int32(offset))
	if err != nil {
		h.log.Errorw("failed to list notifications by user", "error", err, "userID", userID)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 500, "something went wrong")
		return
	}

	response.Success(c, 200, gin.H{"notifications": notifications})
}

func (h *NotificationHandler) MarkAllNotificationsReadHandler(c *gin.Context) {
	ctx := c.Request.Context()

	userID := c.GetString("user_id")
	if userID == "" {
		response.Error(c, 401, "unauthorized")
		return
	}

	h.log.Infow("handling MarkAllNotificationsRead request", "userID", userID)

	err := h.notificationService.MarkAllNotificationsRead(ctx, userID)
	if err != nil {
		h.log.Errorw("failed to mark all notifications read", "error", err, "userID", userID)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 500, "something went wrong")
		return
	}

	response.Success(c, 200, gin.H{"message": "all notifications marked as read"})
}

func (h *NotificationHandler) MarkNotificationReadHandler(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "invalid notification id")
		return
	}

	h.log.Infow("handling MarkNotificationRead request", "notificationID", id)

	err = h.notificationService.MarkNotificationRead(ctx, id)
	if err != nil {
		h.log.Errorw("failed to mark notification read", "error", err, "notificationID", id)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 500, "something went wrong")
		return
	}

	response.Success(c, 200, gin.H{"message": "notification marked as read"})
}

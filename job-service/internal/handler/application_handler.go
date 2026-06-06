package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/service"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/logger"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/response"
)

type ApplicationHandler struct {
	applicationService *service.ApplicationService
	log                *logger.Logger
}

func NewApplicationHandler(applicationService *service.ApplicationService, log *logger.Logger) *ApplicationHandler {
	return &ApplicationHandler{
		applicationService: applicationService,
		log:                log,
	}
}

func (h *ApplicationHandler) CreateApplicationHandler(c *gin.Context) {
	ctx := c.Request.Context()
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.Error(c, 401, "unauthorized")
		return
	}

	var input struct {
		CoverLetter string `json:"cover_letter"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, "invalid request body")
		return
	}
	h.log.Infow("handling CreateApplication request", "jobID", jobID, "userID", userID)
	application, err := h.applicationService.CreateApplication(ctx, jobID, userID, input.CoverLetter)
	if err != nil {
		h.log.Errorw("failed to create application", "error", err, "jobID", jobID)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"application": application})
}

func (h *ApplicationHandler) GetApplicationByIDHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}
	h.log.Infow("handling GetApplicationByID request", "applicationID", id)
	application, err := h.applicationService.GetApplicationByID(ctx, id)
	if err != nil {
		h.log.Errorw("failed to fetch application", "error", err, "applicationID", id)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"application": application})
}

func (h *ApplicationHandler) ListApplicationsByJobHandler(c *gin.Context) {
	ctx := c.Request.Context()
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	h.log.Infow("handling ListApplicationsByJob request", "jobID", jobID)
	applications, err := h.applicationService.ListApplicationsByJob(ctx, jobID)
	if err != nil {
		h.log.Errorw("failed to list applications by job", "error", err, "jobID", jobID)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"applications": applications})
}

func (h *ApplicationHandler) ListApplicationsByUserHandler(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.Error(c, 401, "unauthorized")
		return
	}

	h.log.Infow("handling ListApplicationsByUser request", "userID", userID)
	applications, err := h.applicationService.ListApplicationsByUser(ctx, userID)
	if err != nil {
		h.log.Errorw("failed to list applications by user", "error", err, "userID", userID)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"applications": applications})
}

func (h *ApplicationHandler) UpdateApplicationStatusHandler(c *gin.Context) {
	ctx := c.Request.Context()

	role := c.GetString("role")
	if role == "candidate" {
		response.Error(c, 403, "candidates are not allowed to update application statuses")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "invalid application id")
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, "invalid request body")
		return
	}

	h.log.Infow("handling UpdateApplicationStatus request", "applicationID", id, "status", input.Status, "role", role)
	application, err := h.applicationService.UpdateApplicationStatus(ctx, id, input.Status)
	if err != nil {
		h.log.Errorw("failed to update application status", "error", err, "applicationID", id)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"application": application})
}

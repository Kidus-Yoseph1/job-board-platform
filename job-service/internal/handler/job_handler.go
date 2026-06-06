package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/service"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/logger"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/response"
)

type JobHandler struct {
	jobService *service.JobService
	log        *logger.Logger
}

func NewJobHandler(jobService *service.JobService, log *logger.Logger) *JobHandler {
	return &JobHandler{jobService: jobService, log: log}
}

func (h *JobHandler) CreateJobHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var input struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		Category     string `json:"category"`
		Location     string `json:"location"`
		Type         string `json:"type"`
		IsNegotiable bool   `json:"is_negotiable"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, "invalid request body")
		return
	}
	if input.Title == "" || input.Description == "" || input.Category == "" || input.Location == "" || input.Type == "" {
		response.Error(c, 400, "all fields are required")
		return
	}
	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.Error(c, 401, "unauthorized")
		return
	}

	h.log.Infow("handling CreateJob request", "title", input.Title, "userID", userID)
	job, err := h.jobService.CreateJob(ctx, userID, input.Title, input.Description, input.Category, input.Location, input.Type)
	if err != nil {
		h.log.Errorw("failed to create job", "error", err, "title", input.Title, "userID", userID)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 201, gin.H{"job": job})
}

func (h *JobHandler) GetJobByIDHandler(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}

	h.log.Infow("handling GetJobByID request", "jobID", id)
	job, err := h.jobService.GetJobByID(ctx, id)
	if err != nil {
		h.log.Errorw("failed to fetch job", "error", err, "jobID", id)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"job": job})
}

func (h *JobHandler) ListJobsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	limitStr := c.DefaultQuery("limit", "10")
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

	h.log.Infow("handling ListJobs request", "limit", limit, "offset", offset)
	jobs, err := h.jobService.ListJobs(ctx, int32(limit), int32(offset))
	if err != nil {
		h.log.Errorw("failed to list jobs", "error", err)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"jobs": jobs})
}

func (h *JobHandler) ListJobsByCompanyHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}
	h.log.Infow("handling ListJobsByCompany request", "companyID", id)
	jobs, err := h.jobService.ListJobsByCompany(ctx, id)
	if err != nil {
		h.log.Errorw("failed to list jobs by company", "error", err, "companyID", id)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"jobs": jobs})
}

func (h *JobHandler) UpdateJobStatusHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}
	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, "invalid request body")
		return
	}

	h.log.Infow("handling UpdateJobStatus request", "jobID", id, "status", input.Status)
	updatedJobStatus, err := h.jobService.UpdateJobStatus(ctx, id, input.Status)
	if err != nil {
		h.log.Errorw("failed to update job status", "error", err, "jobID", id)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"job status updated": updatedJobStatus})
}

func (h *JobHandler) DeleteJobHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, 400, "invalid id")
		return
	}
	h.log.Infow("handling DeleteJob request", "jobID", id)
	err = h.jobService.DeleteJob(ctx, id)
	if err != nil {
		h.log.Errorw("failed to delete job", "error", err, "jobID", id)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"message": "job deleted"})
}

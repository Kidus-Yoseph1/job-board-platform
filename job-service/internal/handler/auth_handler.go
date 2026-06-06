package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/domain"
	"github.com/kidus-yoseph1/job-board-platform/job-service/internal/service"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/logger"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
	log         *logger.Logger
}

func NewAuthHandler(authService *service.AuthService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, log: log}
}

func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var input struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, "invalid request body")
		return
	}

	if input.FullName == "" || input.Email == "" || input.Password == "" {
		response.Error(c, 400, "all fields are required")
		return
	}

	h.log.Infow("handling Register request", "email", input.Email)

	user, err := h.authService.Register(ctx, input.FullName, input.Email, input.Password, input.Role)
	if err != nil {
		h.log.Errorw("failed to register user", "error", err, "email", input.Email)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"user registered": user})
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 400, "invalid request body")
		return
	}
	if input.Email == "" || input.Password == "" {
		response.Error(c, 400, "all fields are required")
		return
	}

	h.log.Infow("handling Login request", "email", input.Email)

	token, err := h.authService.Login(ctx, input.Email, input.Password)
	if err != nil {
		h.log.Errorw("failed to login user", "error", err, "email", input.Email)
		if appErr, ok := err.(*domain.AppError); ok {
			response.Error(c, appErr.Code, appErr.Message)
			return
		}
		response.Error(c, 400, "something went wrong")
		return
	}
	response.Success(c, 200, gin.H{"logged in": token})
}

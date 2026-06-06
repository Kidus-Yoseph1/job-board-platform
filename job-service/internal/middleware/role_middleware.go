package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kidus-yoseph1/job-board-platform/job-service/pkg/response"
)

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")
		if userRole != role {
			response.Error(c, 403, "forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}

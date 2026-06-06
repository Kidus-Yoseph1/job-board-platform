package response

import "github.com/gin-gonic/gin"

type Response struct {
	Data  any    `json:"data"`
	Error string `json:"error"`
}

func Success(c *gin.Context, code int, data any) {
	c.JSON(code, Response{Data: data})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{Error: message})
}

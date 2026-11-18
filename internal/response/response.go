package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, Body{Status: "success", Message: message, Data: data})
}

func Fail(c *gin.Context, code int, message string) {
	c.JSON(code, Body{Status: "fail", Message: message})
}

func Error(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Body{Status: "error", Message: message})
}

func AbortFail(c *gin.Context, code int, message string) {
	Fail(c, code, message)
	c.Abort()
}

func AbortError(c *gin.Context, message string) {
	Error(c, message)
	c.Abort()
}

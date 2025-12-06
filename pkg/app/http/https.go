package http

import (
	"github.com/gin-gonic/gin"
	"strings"
)

func IsJsonResponse(c *gin.Context) bool {
	accepts := strings.Split(c.GetHeader("accept"), ",")
	for _, value := range accepts {
		//*/* 时默认返回json
		if value == "application/json" || value == "*/*" {
			return true
		}
	}
	return c.Request.Header.Get("x-requested-with") == "XMLHttpRequest"
	//|| c.Request.Header.Get("accept") == "application/json"
}

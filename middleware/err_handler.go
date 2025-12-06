package middleware

import (
	"chain-love/pkg/app"
	"chain-love/pkg/e"
	"chain-love/pkg/logging"

	"github.com/gin-gonic/gin"
)

func ErrHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var Err *e.Error
				if _e, ok := err.(*e.Error); ok {
					Err = _e
				} else if _e, ok := err.(error); ok {
					Err = e.OtherError(_e.Error())
				} else {
					Err = e.ServerError
					logging.Error(err)
					println(err.(string))
				}
				logging.ErrorLocate(Err.Code, Err.Msg, Err.Locate)
				app.Response(c, Err)
				return
			}
		}()
		c.Next()
	}
}

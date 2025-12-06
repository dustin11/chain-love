package app

import (
	"chain-love/pkg/app/http"
	"chain-love/pkg/e"

	"github.com/gin-gonic/gin"
)

// type ginc struct {
// 	C    *gin.Context
// 	User *JwtUser
// }

// var G *ginc
//
// func Setup(c *gin.Context, u *JwtUser) {
// 	G = &ginc{c, u}
// }

//func (g *ginc) Response(e *e.Error) {
//	g.C.Abort()
//	if http.IsJsonResponse(g.C) {
//		g.C.JSON(e.StatusCode, e)
//	} else {
//		g.C.HTML(e.StatusCode, "500.tmpl", gin.H{"message": e.Msg})
//	}
//}

func Response(c *gin.Context, e *e.Error) {
	c.Abort()
	if http.IsJsonResponse(c) {
		c.JSON(e.StatusCode, e)
	} else {
		c.HTML(e.StatusCode, "500.tmpl", gin.H{"message": e.Msg})
	}
}

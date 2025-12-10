package sys_api

import (
	"chain-love/pkg/app"
	context "chain-love/pkg/app/contextx"
	"chain-love/pkg/e"

	// "chain-love/pkg/logging"
	// "chain-love/pkg/setting/consts"
	// "encoding/json"
	// "io"
	// "strconv"

	"github.com/gin-gonic/gin"
)

// @Summary	用户注册
// @Tags		登录注册
// @Param		name	body		string	true	"如：{'nickname':'jack','username':'aj@qq.com','accountType':1,'password':'1','rePassword':'1'}"
// @Success	200		object		e.Error
// @Failure	500		{object}	e.Error
// @Security	ApiKeyAuth
// @Router		/api/v1/home/register [post]
// func Register(ctx *gin.Context) {
// 	var m sys.User
// 	e.PanicIfErrTipMsg(ctx.ShouldBind(&m), "输入的数据不合法")
// 	sys_service.UserAdd(&m)
// 	app.Response(ctx, e.Success)
// }

// @Summary	获取token
// @Tags		登录注册
// @version	1.0
// @Accept		application/x-json-stream
// @Param		para	body		string	true	"{'username':'a1@qq.com','password':'123456'}"
// @Success	200		object		e.Error
// @Failure	500		{object}	e.Error
// @Router		/api/v1/home/login [post]
func Login(ctx *gin.Context) {
	// logging.Info("get token start ....")
	// str, _ := io.ReadAll(ctx.Request.Body)
	// var para map[string]interface{}
	// json.Unmarshal(str, &para)
	// username := para["username"].(string)
	// password := para["password"].(string)
	// token, err := user_service.Login(username, password)
	// e.PanicIfErr(err)
	// logging.Info(token)
	// res := map[string]string{"token": consts.TOKEN_PREFIX + token,
	// 	"expire":   strconv.Itoa(3600 * 12 * 1000), //12小时后过期
	// 	"username": username,
	// }
	// app.Response(ctx, e.SuccessData(res))
}

// @Summary	退出
// @Tags		登录注册
// @version	1.0
// @Accept		application/x-json-stream
// @Success	200	object		e.Error
// @Failure	500	{object}	e.Error
// @Router		/api/v1/home/logout [post]
func Logout(ctx *gin.Context) {
	app.Response(ctx, e.SuccessData(nil))
}

// @Summary	当前用户
// @Tags		登录注册
// @version	1.0
// @Accept		application/x-json-stream
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/home/info [get]
func UserInfo(ctx *context.AppContext) {
	ctx.Response(e.SuccessData(ctx.User))
}

package middleware

////测试输出并发问题 已证实 http: wrote more than the declared Content-Length 错误
//// 由于app.G内共用context 又没加锁，并发时被另个请求覆盖导致。 之前担心的问题终于出现了，不过比想像中来的早

//todo 加锁也没解决上面的错误，只能暂时弃用请求中拦截设置context的写法，  在 app.G.Response(e.SuccessData(page)) 中加个context参数传递吧
//两个中间件都调用app.Setup导致的？？？

//func Context() gin.HandlerFunc {
//	var mutex sync.Mutex
//	return func(ctx *gin.Context) {
//		mutex.Lock()
//		app.Setup(ctx, nil)
//		mutex.Unlock()
//		ctx.Next()
//	}
//}

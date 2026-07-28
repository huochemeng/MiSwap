package router

import (
	v1 "MiSwap/api/v1"
	"MiSwap/service/svc"
	"github.com/gin-gonic/gin"
)

// v1版本所有的接口地址
func loadV1(r *gin.Engine, svcCtx *svc.ServerCtx) {
	apiV1 := r.Group("/api/v1")

	//	user区域
	user := apiV1.Group("/user")
	//	生成登录的签名信息
	user.GET("/:address/login-message", v1.GetLoginMessageHandler(svcCtx))

}

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
	{
		//	生成登录的签名信息
		user.GET("/:address/login-message", v1.GetLoginMessageHandler(svcCtx))
		//  使用签名登录信息，获取token
		user.POST("/login", v1.UserLoginHandler(svcCtx))
		//获取用户签名状态
		user.GET("/:address/sig-status", v1.GetSigStatusHandler(svcCtx))
	}

	//	collection区域
	collection := apiV1.Group("/collection")
	{
		//	指定集合的address查询collection详情
		collection.GET("/:address", v1.CollectionDetailHandler(svcCtx))
	}
}

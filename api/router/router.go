package router

import (
	"MiSwap/service/svc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

func NewRouter(svcCtx *svc.ServerCtx) *gin.Engine {
	r := gin.New()

	r.Use(cors.New(cors.Config{ // 使用cors中间件
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "X-CSRF-Token", "Authorization", "AccessToken", "Token"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "X-GW-Error-Code", "X-GW-Error-Message"},
		AllowCredentials: true,
		MaxAge:           1 * time.Hour,
	}))
	//	TODO 加载v1路由器
	loadV1(r, svcCtx)
	return r
}

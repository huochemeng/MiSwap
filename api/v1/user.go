package v1

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/pkg/xhttp"
	"MiSwap/service/svc"
	"MiSwap/service/v1"
	"github.com/gin-gonic/gin"
)

func GetLoginMessageHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Params.ByName("address")
		if address == "" {
			//c.JSON(http.StatusBadRequest, gin.H{
			//	"code":    400,
			//	"message": "address is required",
			//})
			xhttp.Error(c, errcode.NewCustomErr("address is null"))
			return
		}
		//	todo service层调用生成登录信息的方法
		res, err := service.GetLoginMessage(c.Request.Context(), svcCtx, address)
		if err != nil {
			//c.JSON(http.StatusInternalServerError, gin.H{
			//	"code":    500,
			//	"message": "generate login message failed",
			//})
			xhttp.Error(c, errcode.NewCustomErr(err.Error()))
			return
		}
		//	完整返回
		//	c.JSON(http.StatusOK, gin.H{
		//		"code":    200,
		//		"message": "generate login message success",
		//		"Data":    res,
		//	})
		xhttp.OkJson(c, res)
	}
}

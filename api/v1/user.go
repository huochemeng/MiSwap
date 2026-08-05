package v1

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/pkg/xhttp"
	"MiSwap/dto"
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

func UserLoginHandler(ctx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		//bodyBytes, _ := io.ReadAll(c.Request.Body)
		//log.Printf("[DEBUG] Raw Body Length: %d", len(bodyBytes))
		//log.Printf("[DEBUG] Raw Body Content: %s", string(bodyBytes))
		//log.Printf("[DEBUG] Content-Type: %s", c.ContentType())
		req := dto.LoginReq{}
		if err := c.BindJSON(&req); err != nil {
			xhttp.Error(c, err)
			return
		}
		//todo	验证参数的格式是否有误

		res, err := service.UserLogin(c.Request.Context(), ctx, req)
		if err != nil {
			xhttp.Error(c, err)
			return
		}

		xhttp.OkJson(c, dto.UserLoginResp{Result: res})
	}
}

func GetSigStatusHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		userAddr := c.Params.ByName("address")
		if userAddr == "" {
			xhttp.Error(c, errcode.NewCustomErr("user address is null"))
			return
		}
		//	数据库查询
		res, err := service.GetSigStatusMsg(c.Request.Context(), svcCtx, userAddr)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr(err.Error()))
			return
		}
		xhttp.OkJson(c, res)
	}
}

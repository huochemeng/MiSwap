package v1

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/pkg/xhttp"
	"MiSwap/service/svc"
	"MiSwap/service/v1"
	"github.com/gin-gonic/gin"
	"strconv"
)

func CollectionDetailHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		//URL查询字符串，解析chain_id
		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 32)
		if err != nil {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}
		//	chainID数字对应链名
		chain, ok := chainIDToChain[int(chainID)]
		if !ok {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}
		//	检验collection的address参数信息
		collectionAddr := c.Params.ByName("address")
		if collectionAddr == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}
		//业务层查询
		res, err := service.GetCollectionDetail(c.Request.Context(), svcCtx, chain, collectionAddr)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("handler: failed to get collection information"))
			return
		}
		xhttp.OkJson(c, res)
	}
}

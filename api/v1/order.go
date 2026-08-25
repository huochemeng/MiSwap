package v1

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/pkg/xhttp"
	"MiSwap/dto"
	"MiSwap/service/svc"
	"MiSwap/service/v1"
	"encoding/json"
	"github.com/gin-gonic/gin"
)

func OrderInfoHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		filterParam := c.Query("filters")
		if filterParam == "" {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		var filter dto.OrderInfosParam
		err := json.Unmarshal([]byte(filterParam), &filter)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		chain, ok := chainIDToChain[filter.ChainID]
		if !ok {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		res, err := service.GetOrderInfos(c.Request.Context(), svcCtx, filter.ChainID, chain, filter.UserAddress, filter.CollectionAddress, filter.TokenIds)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr(err.Error()))
			return
		}
		xhttp.OkJson(c, struct {
			Result interface{} `json:"result"`
		}{Result: res})
	}
}

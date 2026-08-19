package v1

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/pkg/xhttp"
	"MiSwap/dto"
	"MiSwap/service/svc"
	"MiSwap/service/v1"
	"encoding/json"
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

func CollectionBidsHandler(svcCtx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		filterParam := c.Query("filters")
		if filterParam == "" {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		var filter dto.CollectionBidFilterParams
		err := json.Unmarshal([]byte(filterParam), &filter)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		collectionAddr := c.Params.ByName("address")
		if collectionAddr == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		chain, ok := chainIDToChain[int(filter.ChainID)]
		if !ok {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		res, err := service.GetBids(c.Request.Context(), svcCtx, chain, collectionAddr, filter.Page, filter.PageSize)
		if err != nil {
			xhttp.Error(c, errcode.ErrUnexpected)
			return
		}
		xhttp.OkJson(c, res)
	}
}

func CollectionItemBidsHandler(ctx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		filterParam := c.Query("filters")
		if filterParam == "" {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		var filter dto.CollectionBidFilterParams
		err := json.Unmarshal([]byte(filterParam), &filter)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		collectionAddr := c.Params.ByName("address")
		if collectionAddr == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		tokenID := c.Params.ByName("token_id")
		if tokenID == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		chain, ok := chainIDToChain[int(filter.ChainID)]
		if !ok {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		res, err := service.GetItemBidsInfo(c.Request.Context(), ctx, chain, collectionAddr, tokenID, filter.Page, filter.PageSize)
		if err != nil {
			xhttp.Error(c, errcode.ErrUnexpected)
			return
		}
		xhttp.OkJson(c, res)
	}
}

func CollectionItemsHandler(ctx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		filterParam := c.Query("filters")
		if filterParam == "" {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		var filter dto.CollectionItemFilterParams
		err := json.Unmarshal([]byte(filterParam), &filter)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		collectionAddr := c.Params.ByName("address")
		if collectionAddr == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		chain, ok := chainIDToChain[filter.ChainID]
		if !ok {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}
		res, err := service.GetItems(c.Request.Context(), ctx, chain, filter, collectionAddr)
		if err != nil {
			xhttp.Error(c, errcode.ErrUnexpected)
			return
		}
		xhttp.OkJson(c, res)
	}
}

func ItemDetailHandler(ctx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		collectionAddr := c.Params.ByName("address")
		if collectionAddr == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		tokenID := c.Params.ByName("token_id")
		if tokenID == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 64)
		if err != nil {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		chain, ok := chainIDToChain[int(chainID)]
		if !ok {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		res, err := service.GetItem(c.Request.Context(), ctx, chain, int(chainID), collectionAddr, tokenID)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("get item error"))
			return

		}
		xhttp.OkJson(c, res)
	}
}

func ItemTraitsHandler(ctx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		collectionAddr := c.Params.ByName("address")
		if collectionAddr == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		tokenID := c.Params.ByName("token_id")
		if tokenID == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 64)
		if err != nil {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		chain, ok := chainIDToChain[int(chainID)]
		if !ok {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		itemTraits, err := service.GetItemTraits(c.Request.Context(), ctx, chain, collectionAddr, tokenID)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("get item traits error"))
			return
		}

		xhttp.OkJson(c, dto.ItemTraitsResp{Result: itemTraits})
	}
}

func ItemTopTraitPriceHandler(ctx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		collectionAddr := c.Params.ByName("address")
		if collectionAddr == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		filterParam := c.Query("filters")
		if filterParam == "" {
			xhttp.Error(c, errcode.NewCustomErr("Filter param is nil."))
			return
		}

		var filter dto.TopTraitFilterParams
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

		res, err := service.GetItemTopTraitPrice(c.Request.Context(), ctx, chain, collectionAddr, filter.TokenIds)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("get item error"))
			return
		}
		xhttp.OkJson(c, res)
	}
}

func GetItemImageHandler(ctx *svc.ServerCtx) gin.HandlerFunc {
	return func(c *gin.Context) {
		collectionAddr := c.Params.ByName("address")
		if collectionAddr == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		tokenID := c.Params.ByName("token_id")
		if tokenID == "" {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		chainID, err := strconv.ParseInt(c.Query("chain_id"), 10, 64)
		if err != nil {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		chain, ok := chainIDToChain[int(chainID)]
		if !ok {
			xhttp.Error(c, errcode.ErrInvalidParams)
			return
		}

		result, err := service.GetItemImage(c.Request.Context(), ctx, chain, collectionAddr, tokenID)
		if err != nil {
			xhttp.Error(c, errcode.NewCustomErr("failed on get item image"))
			return
		}

		xhttp.OkJson(c, struct {
			Result interface{} `json:"result"`
		}{Result: result})
	}
}

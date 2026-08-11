package service

import (
	"MiSwap/base/ordermanager"
	"MiSwap/base/pkg/errcode"
	"MiSwap/dto"
	"MiSwap/service/svc"
	"context"
	"github.com/shopspring/decimal"
	"log"
)

func GetCollectionDetail(ctx context.Context, svcCtx *svc.ServerCtx, chain string, addr string) (*dto.CollectionDetailResp, error) {
	//查询集合基本信息
	collection, err := svcCtx.Dao.QueryCollectionInfo(ctx, chain, addr)
	if err != nil {
		return nil, errcode.NewCustomErr("service: failed to get collection information")
	}
	tradeInfo, err := svcCtx.Dao.GetTradeInfoByCollection(ctx, chain, addr, "1d")
	if err != nil {
		log.Printf("[warn] service: failed to get collection trade information, err=%v", err)
	}
	// 针对结果判断，赋值24h的volume和turnover
	var volume24h int64
	var turnover24h decimal.Decimal
	if tradeInfo != nil {
		volume24h = tradeInfo.Volume
		turnover24h = tradeInfo.Turnover
	}
	listed, err := svcCtx.Dao.QueryListedAmount(ctx, chain, addr)
	if err != nil {
		log.Printf("[warn] service: failed to get listed amount, err=%v", err)
	} else {
		if err := svcCtx.Dao.CacheCollectionListed(ctx, chain, addr, int(listed)); err != nil {
			log.Printf("[warn] service: failed to cache collection listed, err=%v", err)
		}
	}
	floorPrice, err := svcCtx.Dao.QueryFloorPrice(ctx, chain, addr)
	if err != nil {
		log.Printf("[warn] service: failed to get floorPrice, err=%v", err)
	}
	if !floorPrice.Equal(collection.FloorPrice) {
		if err := ordermanager.AddUpdatePriceEvent(svcCtx.KvStore, &ordermanager.TradeEvent{
			EventType:      ordermanager.UpdateCollection,
			CollectionAddr: addr,
			Price:          floorPrice,
		}, chain); err != nil {
			log.Printf("[warn] service: ")
		}
	}
	var sellPrice string
	collectionSell, err := svcCtx.Dao.QueryCollectionSellPrice(ctx, chain, addr)
	if err != nil {
		log.Printf("[warn] service: failed to get highest selling price")
		sellPrice = "0"
	} else {
		sellPrice = collectionSell.SalePrice.String()
	}

	// 获取总的交易额
	var turnoverTotal decimal.Decimal
	collectionTurnover, err := svcCtx.Dao.GetCollectionTurnover(chain, addr)
	if err != nil {
		log.Printf("[warn] service: failed to get collection tatal turnover, err=%v", err)
	} else {
		turnoverTotal = collectionTurnover
	}

	//构建返回结果
	detail := dto.CollectionDetail{
		ImageUri:      collection.ImageUri,
		Name:          collection.Name,
		Address:       collection.Address,
		ChainId:       collection.ChainId,
		Volume24h:     volume24h,
		Turnover24h:   turnover24h,
		ListAmount:    listed,
		FloorPrice:    floorPrice,
		SellPrice:     sellPrice,
		TotalSupply:   collection.ItemAmount,
		OwnerAmount:   collection.OwnerAmount,
		TurnoverTotal: turnoverTotal,
	}
	return &dto.CollectionDetailResp{
		Result: detail,
	}, nil
}

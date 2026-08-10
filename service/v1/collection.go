package service

import (
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
	//todo 查询上架数量
	listed, err := svcCtx.Dao.QueryListedAmount(ctx, chain, addr)
	if err != nil {
		log.Printf("[warn] service: failed to get listed amount, err=%v", err)
	} else {
		//	todo 缓存上架数量

	}
	_ = listed
	//todo 查询地板价
	//todo 查询卖单价格
	//todo 如果地板价发生变化，更新价格时间
	//todo 获取24小时交易量和销售量
	//todo 查询总交易量

	//构建返回结果
	detail := dto.CollectionDetail{
		ImageUri:    collection.ImageUri,
		Name:        collection.Name,
		Address:     collection.Address,
		ChainId:     collection.ChainId,
		Volume24h:   volume24h,
		Turnover24h: turnover24h,
		//	其他数据，需要查询数据库
	}
	return &dto.CollectionDetailResp{
		Result: detail,
	}, nil
}

package service

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"MiSwap/dao"
	"MiSwap/dto"
	"MiSwap/service/svc"
	"context"
	"github.com/shopspring/decimal"
	"log/slog"
	"strconv"
	"strings"
	"sync"
)

const MinuteSeconds = 60
const HourSeconds = 60 * 60
const DaySeconds = 3600 * 24

func GetTopRanking(ctx context.Context, svcCtx *svc.ServerCtx, chain string, period string, limit int64) ([]*dto.CollectionRankingInfo, error) {
	// 获取集合交易信息
	tradeInfos, err := svcCtx.Dao.GetCollectionRankingByActivity(chain, period)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get collection trade info", "error", err)
	}

	// 构建交易信息map
	collectionTradeMap := make(map[string]dao.CollectionTrade)
	for _, tradeInfo := range tradeInfos {
		collectionTradeMap[strings.ToLower(tradeInfo.ContractAddress)] = *tradeInfo
	}

	// 时间范围映射表
	periodTime := map[string]int64{
		"15m": MinuteSeconds * 15,
		"1h":  HourSeconds,
		"6h":  HourSeconds * 6,
		"1d":  DaySeconds,
		"7d":  DaySeconds * 7,
		"30d": DaySeconds * 30,
	}
	// 获取地板价变化信息
	collectionFloorChange, err := svcCtx.Dao.QueryCollectionFloorChange(chain, periodTime[period])
	if err != nil {
		slog.ErrorContext(ctx, "failed to get collection floor change ", "error", err)
	}

	var wg sync.WaitGroup
	var queryErr error

	// 并发获取集合销售价格信息
	collectionSells := make(map[string]model.Collection)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sellInfos, err := svcCtx.Dao.QueryCollectionsSellPrice(ctx, chain)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get all collections info", "error", err)
			queryErr = errcode.NewCustomErr("failed to get all collections info")
			return
		}
		for _, sell := range sellInfos {
			collectionSells[strings.ToLower(sell.Address)] = sell
		}
	}()

	// 并发获取所有集合基本信息
	var allCollections []model.Collection
	wg.Add(1)
	go func() {
		defer wg.Done()
		allCollections, err = svcCtx.Dao.QueryAllCollectionInfo(ctx, chain)
		if err != nil {
			slog.ErrorContext(ctx, "failed to get all collections info", "error", err)
			queryErr = errcode.NewCustomErr("failed on get all collections info")
			return
		}
	}()

	wg.Wait()

	if queryErr != nil {
		return nil, queryErr
	}

	// 构建返回结果
	var respInfos []*dto.CollectionRankingInfo
	for _, collection := range allCollections {
		var priceChange float64
		var turnover decimal.Decimal
		var sellPrice decimal.Decimal
		var sales int64

		// 获取交易相关信息
		tradeInfo, ok := collectionTradeMap[strings.ToLower(collection.Address)] // 统一小写
		if ok {
			priceChange = collectionFloorChange[strings.ToLower(collection.Address)]
			turnover = tradeInfo.Turnover
			sales = tradeInfo.Volume
		}
		// 获取销售价格信息
		sellInfo, ok := collectionSells[strings.ToLower(collection.Address)]
		if ok {
			sellPrice = sellInfo.SalePrice
		}

		// 获取上架数量
		var listAmount int
		listed, err := svcCtx.Dao.QueryCollectionsListed(ctx, chain, []string{collection.Address})
		if err != nil {
			slog.ErrorContext(ctx, "failed to query collection listed", "error", err)
		} else {
			listAmount = listed[0].Count
		}

		// 构建单个集合的排名信息
		respInfos = append(respInfos, &dto.CollectionRankingInfo{
			Name:        collection.Name,
			Address:     collection.Address,
			ImageUri:    collection.ImageUri,
			FloorPrice:  collection.FloorPrice.String(),
			FloorChange: strconv.FormatFloat(priceChange, 'f', 4, 32),
			SellPrice:   sellPrice.String(),
			Turnover:    turnover,
			ItemSold:    sales,
			Volume:      collection.ItemAmount,
			ItemOwner:   collection.OwnerAmount,
			ListAmount:  listAmount,
			ChainID:     collection.ChainId,
		})
	}

	// 限制返回数量
	if limit < int64(len(respInfos)) {
		respInfos = respInfos[:limit]
	}

	return respInfos, nil
}

package service

import (
	"MiSwap/base/ordermanager"
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"MiSwap/dto"
	"MiSwap/service/svc"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"log"
	"strings"
	"sync"
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

func GetBids(ctx context.Context, svcCtx *svc.ServerCtx, chain string, addr string, page int, size int) (*dto.CollectionBidsResp, error) {
	bids, count, err := svcCtx.Dao.QueryCollectionBids(ctx, chain, addr, page, size)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get item info")
	}

	return &dto.CollectionBidsResp{
		Result: bids,
		Count:  count,
	}, nil
}

func GetItemBidsInfo(ctx context.Context, ctx2 *svc.ServerCtx, chain string, addr string, tokenID string, page int, size int) (*dto.CollectionBidsResp, error) {
	bids, count, err := ctx2.Dao.QueryItemBids(ctx, chain, addr, tokenID, page, size)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get item information")
	}

	for i := 0; i < len(bids); i++ {
		bids[i].OrderType = getBidType(bids[i].OrderType)
	}
	return &dto.CollectionBidsResp{
		Result: bids,
		Count:  count,
	}, nil
}

func GetItems(ctx context.Context, svcCtx *svc.ServerCtx, chain string, filter dto.CollectionItemFilterParams, addr string) (*dto.NFTListingInfoResp, error) {
	// 1. 查询基础Item信息和订单信息
	items, count, err := svcCtx.Dao.QueryCollectionItemOrder(ctx, chain, filter, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get item info")
	}

	// 整理提取需要查询的itemID和所有者地址
	var ItemIds []string
	var ItemOwners []string
	var itemPrice []dto.ItemPriceInfo
	for _, item := range items {
		if item.TokenId != "" {
			ItemIds = append(ItemIds, item.TokenId)
		}
		if item.Owner != "" {
			ItemOwners = append(ItemOwners, item.Owner)
		}
		// 记录已上架Item的价格信息
		if item.Listing {
			itemPrice = append(itemPrice, dto.ItemPriceInfo{
				CollectionAddress: item.CollectionAddress,
				TokenID:           item.TokenId,
				Maker:             item.Owner,
				Price:             item.ListPrice,
				OrderStatus:       model.OrderStatusActive,
			})
		}
	}

	// 3. 并发查询各类扩展信息
	var queryErr error
	var wg sync.WaitGroup

	// 3.1 查询订单详情
	ordersInfo := make(map[string]model.Order)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if len(itemPrice) > 0 {
			orders, err := svcCtx.Dao.QueryListingInfo(ctx, chain, itemPrice)
			if err != nil {
				queryErr = errors.Wrap(err, "failed to get orders time info")
				return
			}
			for _, order := range orders {
				ordersInfo[strings.ToLower(order.CollectionAddress+order.TokenId)] = order
			}
		}
	}()

	// 3.2 查询Item图片信息
	ItemsExternal := make(map[string]model.ItemExternal)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if len(ItemIds) != 0 {
			items, err := svcCtx.Dao.QueryCollectionItemsImage(ctx, chain, addr, ItemIds)
			if err != nil {
				queryErr = errors.Wrap(err, "failed to get items image info")
				return
			}
			for _, item := range items {
				ItemsExternal[strings.ToLower(item.TokenId)] = item
			}
		}
	}()

	// 3.3 查询用户持有数量
	userItemCount := make(map[string]int64)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if len(ItemIds) != 0 {
			itemCount, err := svcCtx.Dao.QueryUsersItemCount(ctx, chain, addr, ItemOwners)
			if err != nil {
				queryErr = errors.Wrap(err, "failed to get items image info")
				return
			}
			for _, v := range itemCount {
				userItemCount[strings.ToLower(v.Owner)] = v.Counts
			}
		}
	}()

	// 3.4 查询最近成交价格
	lastSales := make(map[string]decimal.Decimal)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if len(ItemIds) != 0 {
			lastSale, err := svcCtx.Dao.QueryLastSalePrice(ctx, chain, addr, ItemIds)
			if err != nil {
				queryErr = errors.Wrap(err, "failed to get items last sale info")
				return
			}
			for _, v := range lastSale {
				lastSales[strings.ToLower(v.TokenId)] = v.Price
			}
		}
	}()

	// 3.5 查询Item级别最高出价
	bestBids := make(map[string]model.Order)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if len(ItemIds) != 0 {
			bids, err := svcCtx.Dao.QueryBestBids(ctx, chain, filter.UserAddress, addr, ItemIds)
			if err != nil {
				queryErr = errors.Wrap(err, "failed to get items last sale info")
				return
			}
			for _, bid := range bids {
				order, ok := bestBids[strings.ToLower(bid.TokenId)]
				if !ok {
					bestBids[strings.ToLower(bid.TokenId)] = bid
					continue
				}
				if bid.Price.GreaterThan(order.Price) {
					bestBids[strings.ToLower(bid.TokenId)] = bid
				}
			}
		}
	}()

	// 3.6 查询集合级别最高出价
	var collectionBestBid model.Order
	wg.Add(1)
	go func() {
		defer wg.Done()
		collectionBestBid, err = svcCtx.Dao.QueryCollectionBestBid(ctx, chain, filter.UserAddress, addr)
		if err != nil {
			queryErr = errors.Wrap(err, "failed to get items last sale info")
			return
		}
	}()

	// 4. 等待所有查询完成
	wg.Wait()
	if queryErr != nil {
		return nil, errors.Wrap(queryErr, "failed to get items info")
	}

	// 5. 整合所有信息
	var respItems []*dto.NFTListingInfo
	for _, item := range items {
		// 设置Item名称
		nameStr := item.Name
		if nameStr == "" {
			nameStr = fmt.Sprintf("#%s", item.TokenId)
		}

		// 构建返回结构
		respItem := &dto.NFTListingInfo{
			Name:              nameStr,
			CollectionAddress: item.CollectionAddress,
			TokenID:           item.TokenId,
			OwnerAddress:      item.Owner,
			ListPrice:         item.ListPrice,
			MarketID:          item.MarketID,
			BidOrderID:        collectionBestBid.OrderID,
			BidExpireTime:     collectionBestBid.ExpireTime,
			BidPrice:          collectionBestBid.Price,
			BidTime:           collectionBestBid.EventTime,
			BidSalt:           collectionBestBid.Salt,
			BidMaker:          collectionBestBid.Maker,
			BidType:           getBidType(collectionBestBid.OrderType),
			BidSize:           collectionBestBid.Size,
			BidUnfilled:       collectionBestBid.QuantityRemaining,
		}

		// 添加订单信息
		listOrder, ok := ordersInfo[strings.ToLower(item.CollectionAddress+item.TokenId)]
		if ok {
			respItem.ListTime = listOrder.EventTime
			respItem.ListOrderID = listOrder.OrderID
			respItem.ListExpireTime = listOrder.ExpireTime
			respItem.ListSalt = listOrder.Salt
		}

		// 添加最高出价信息
		bidOrder, ok := bestBids[strings.ToLower(item.TokenId)]
		if ok {
			if bidOrder.Price.GreaterThan(collectionBestBid.Price) {
				respItem.BidOrderID = bidOrder.OrderID
				respItem.BidExpireTime = bidOrder.ExpireTime
				respItem.BidPrice = bidOrder.Price
				respItem.BidTime = bidOrder.EventTime
				respItem.BidSalt = bidOrder.Salt
				respItem.BidMaker = bidOrder.Maker
				respItem.BidType = getBidType(bidOrder.OrderType)
				respItem.BidSize = bidOrder.Size
				respItem.BidUnfilled = bidOrder.QuantityRemaining
			}
		}

		// 添加图片和视频信息
		itemExternal, ok := ItemsExternal[strings.ToLower(item.TokenId)]
		if ok {
			if itemExternal.IsUploadedOss {
				respItem.ImageURI = itemExternal.OssUri
			} else {
				respItem.ImageURI = itemExternal.ImageUri
			}
			if len(itemExternal.VideoUri) > 0 {
				respItem.VideoType = itemExternal.VideoType
				if itemExternal.IsVideoUploaded {
					respItem.VideoURI = itemExternal.VideoOssUri
				} else {
					respItem.VideoURI = itemExternal.VideoUri
				}
			}
		}

		// 添加用户持有数量
		count, ok := userItemCount[strings.ToLower(item.Owner)]
		if ok {
			respItem.OwnerOwnedAmount = count
		}

		// 添加最近成交价格
		price, ok := lastSales[strings.ToLower(item.TokenId)]
		if ok {
			respItem.LastSellPrice = price
		}

		respItems = append(respItems, respItem)
	}

	return &dto.NFTListingInfoResp{
		Result: respItems,
		Count:  count,
	}, nil
}

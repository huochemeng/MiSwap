package dao

import (
	"MiSwap/base/kit/time"
	"MiSwap/base/ordermanager"
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var collectionDetailFields = []string{"id", "chain_id", "token_standard", "name", "address", "image_uri", "floor_price", "sale_price", "item_amount", "owner_amount"}

func (d *Dao) QueryCollectionInfo(ctx context.Context, chain string, addr string) (*model.Collection, error) {
	var collection model.Collection
	if err := d.DB.WithContext(ctx).Table(model.CollectionTableName(chain)).
		Select(collectionDetailFields).Where("address = ?", addr).
		First(&collection).Error; err != nil {
		return nil, errcode.NewCustomErr("dao: failed to get collection information")
	}
	return &collection, nil
}

// CacheCollectionListed 缓存集合的上架数量
func (d *Dao) CacheCollectionListed(ctx context.Context, chain string, addr string, listedCount int) error {
	err := d.KvStore.SetInt(ordermanager.GenCollectionListedKey(chain, addr), listedCount)
	if err != nil {
		return errcode.NewCustomErr("failed to set collection listed count")
	}
	return nil
}

// QueryFloorPrice 查询NFT集合的地板价
func (d *Dao) QueryFloorPrice(ctx context.Context, chain string, addr string) (decimal.Decimal, error) {
	var order model.Order
	// SQL解释:
	// 1. 从Item表(ci)和订单表(co)联表查询
	// 2. 选择字段:co.price作为地板价
	// 3. 关联条件:集合地址和tokenID都相同
	// 4. WHERE条件:
	//    - 指定集合地址
	//    - 订单类型为listing(OrderType=1)
	//    - 订单状态为active(OrderStatus=0)
	//    - 卖家是NFT当前所有者
	//    - 排除marketplace_id=1的订单
	// 5. 按价格升序排序,取第一条记录(即最低价)
	sql := fmt.Sprintf(`SELECT co.price as price
		FROM %s as ci
				left join %s co on co.collection_address = ci.collection_address and co.token_id = ci.token_id
		WHERE (co.collection_address= ? and co.order_type = ? and
			co.order_status = ? and co.maker = ci.owner and co.marketplace_id != ?)
		order by co.price asc limit 1`, model.ItemTableName(chain), model.OrderTableName(chain))
	// 执行sql查询
	if err := d.DB.WithContext(ctx).Raw(
		sql,
		addr,
		model.ListingType,
		model.OrderStatusActive,
		1).Scan(&order).Error; err != nil {
		return decimal.Zero, errcode.NewCustomErr("failed to get collection floorPrice")
	}
	return order.Price, nil
}

// QueryCollectionSellPrice 查询指定NFT集合的最高卖单价格
func (d *Dao) QueryCollectionSellPrice(ctx context.Context, chain, collectionAddr string) (*model.Collection, error) {
	var collection model.Collection
	sql := fmt.Sprintf(`SELECT collection_address as address, co.price as sale_price
					FROM %s as co 
					where 
					    collection_address = ? and order_status = ? and order_type = ? 
					    and quantity_remaining > 0 and expire_time > ? order by price desc limit 1`,
		model.OrderTableName(chain))
	if err := d.DB.WithContext(ctx).Raw(
		sql,
		collectionAddr,
		model.OrderStatusActive,
		model.CollectionBidOrder,
		time.Now().Unix()).Scan(&collection).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get collection sell price")
	}

	return &collection, nil
}

// QueryHistorySalesPriceInfo 查询指定时间段内的NFT销售历史价格信息
func (d *Dao) QueryHistorySalesPriceInfo(ctx context.Context, chain string, addr string, stamp int64) ([]model.Activity, error) {
	var historySalesInfo []model.Activity
	now := time.Now().Unix()

	// SQL语句解释:
	// 1. 从activity表中查询指定字段(price,token_id,event_time)
	// 2. 条件:
	//   - 活动类型为Sale(销售)
	//   - 集合地址匹配
	//   - 事件时间在指定范围内(now-duration到now)
	if err := d.DB.WithContext(ctx).
		Table(model.ActivityTableName(chain)).
		Select("price", "token_id", "event_time").
		Where("activity_type = ? and collection_address = ? and event_time >= ? and event_time <= ?",
			model.Sale,
			addr,
			now-stamp,
			now).
		Find(&historySalesInfo).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get history sales info")
	}

	return historySalesInfo, nil
}

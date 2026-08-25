package dao

import (
	"MiSwap/base/ordermanager"
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"MiSwap/dto"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"time"
)

const MaxBatchReadCollections = 500
const MaxRetries = 3
const QueryTimeout = time.Second * 30

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

// QueryCollectionFloorChange 查询集合地板价变化情况
// @param chain string 链名称
// @param timeDiff int64 时间差(秒)
// @return map[string]float64 返回集合地址到地板价变化率的映射
// @return error 错误信息
func (d *Dao) QueryCollectionFloorChange(chain string, timeDiff int64) (map[string]float64, error) {
	collectionFloorChange := make(map[string]float64)

	var collectionPrices []model.CollectionFloorPrice
	// 这个SQL语句用于查询NFT集合的地板价变化情况:
	// 1. 从集合地板价表中选择collection_address(集合地址)、price(价格)和event_time(事件时间)
	// 2. WHERE子句包含两个条件:
	//    a) 查询每个集合的最新地板价记录(通过GROUP BY和MAX(event_time)获取)
	//    b) 查询每个集合在指定时间段之前的最新地板价记录(通过WHERE event_time <= UNIX_TIMESTAMP() - ? 筛选)
	// 3. 最后按集合地址和时间降序排序,这样可以方便计算价格变化率
	rawSql := fmt.Sprintf(`SELECT collection_address, price, event_time 
		FROM %s 
		WHERE (collection_address, event_time) IN (
			SELECT collection_address, MAX(event_time)
			FROM %s
			GROUP BY collection_address
		) OR (collection_address, event_time) IN (
			SELECT collection_address, MAX(event_time)
			FROM %s 
			WHERE event_time <= UNIX_TIMESTAMP() - ? 
			GROUP BY collection_address
		) 
		ORDER BY collection_address,event_time DESC`,
		model.CollectionFloorPriceTableName(chain),
		model.CollectionFloorPriceTableName(chain),
		model.CollectionFloorPriceTableName(chain))
	if err := d.DB.Raw(rawSql, timeDiff).Scan(&collectionPrices).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get collection floor change")
	}

	// 这个循环用于计算每个NFT集合的地板价变化率:
	// 1. 遍历collectionPrices数组,每个元素包含集合地址和对应时间点的地板价
	// 2. 对于每个集合:
	//    - 如果当前元素和下一个元素是同一个集合的记录(CollectionAddress相同)
	//    - 且下一个元素的价格大于0
	//    则:
	//    - 计算价格变化率 = (当前价格 - 历史价格) / 历史价格
	//    - 使用Price.Sub()计算价格差
	//    - 使用Div()计算变化率
	//    - 使用InexactFloat64()转换为float64类型
	//    - i++跳过下一个元素(因为已经使用过了)
	// 3. 如果不满足条件,则将该集合的变化率设为0
	// 4. 最终得到一个从集合地址到其地板价变化率的映射
	for i := 0; i < len(collectionPrices); i++ {
		if i < len(collectionPrices)-1 &&
			collectionPrices[i].CollectionAddress == collectionPrices[i+1].CollectionAddress &&
			collectionPrices[i+1].Price.GreaterThan(decimal.Zero) {
			collectionFloorChange[collectionPrices[i].CollectionAddress] = collectionPrices[i].Price.
				Sub(collectionPrices[i+1].Price).Div(collectionPrices[i+1].Price).InexactFloat64()
			i++
		} else {
			collectionFloorChange[collectionPrices[i].CollectionAddress] = 0.0
		}
	}

	return collectionFloorChange, nil
}

// QueryCollectionsSellPrice 查询所有集合的最高卖单价格
// @param ctx context.Context 上下文
// @param chain string 链名称
// @return []model.Collection 返回集合列表,每个集合包含地址和最高卖单价格
// @return error 错误信息
func (d *Dao) QueryCollectionsSellPrice(ctx context.Context, chain string) ([]model.Collection, error) {
	var collections []model.Collection
	// 这条SQL语句用于查询每个NFT集合的最高卖单价格:
	// 1. 从订单表中选择数据
	// 2. 选择字段:
	//    - collection_address 作为 address - NFT集合地址
	//    - max(co.price) 作为 sale_price - 该集合最高的卖单价格
	// 3. 查询条件:
	//    - order_status = ? - 订单状态(传入参数,筛选：有效订单)
	//    - order_type = ? - 订单类型(传入参数,筛选：卖订单)
	//    - expire_time > ? - 过期时间大于当前时间(筛选：未过期订单)
	// 4. group by collection_address - 按集合地址分组,获取每个集合的最高价
	sql := fmt.Sprintf(`SELECT collection_address as address, max(co.price) as sale_price
FROM %s as co where order_status = ? and order_type = ? and expire_time > ? group by collection_address`, model.OrderTableName(chain))
	if err := d.DB.WithContext(ctx).Raw(
		sql,
		model.OrderStatusActive,
		model.CollectionBidOrder,
		time.Now().Unix()).Scan(&collections).Error; err != nil {
		return nil, errors.Wrap(err, "failed on get collection sell price")
	}

	return collections, nil
}

// QueryAllCollectionInfo 查询指定链上的所有NFT集合信息
func (d *Dao) QueryAllCollectionInfo(ctx context.Context, chain string) ([]model.Collection, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	tx := d.DB.WithContext(ctx).Begin() // 开启事务
	defer func() {                      // 捕获异常
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
			panic(r)
		}
	}()

	cursor := int64(0) // 游标
	var allCollections []model.Collection
	// 循环分页查询所有集合信息
	for {
		var collections []model.Collection
		// 最多重试MaxRetries次查询
		for i := 0; i < MaxRetries; i++ {
			// 查询大于当前cursor的MaxBatchReadCollections条记录
			err := tx.Table(model.CollectionTableName(chain)).
				Select(collectionDetailFields).
				Where("id > ?", cursor).
				Limit(MaxBatchReadCollections).
				Order("id asc").
				Scan(&collections).Error

			// 查询成功则跳出重试循环
			if err == nil {
				break
			}
			// 达到最大重试次数仍失败,则回滚事务并返回错误
			if i == MaxRetries-1 {
				tx.Rollback()
				return nil, errors.Wrap(err, "failed on get collections info")
			}
			// 重试间隔时间递增
			time.Sleep(time.Duration(i+1) * time.Second)
		}

		// 将本次查询结果追加到总结果中
		allCollections = append(allCollections, collections...)
		// 如果本次查询结果数小于批次大小,说明已经查完所有记录
		if len(collections) < MaxBatchReadCollections {
			break
		}

		// 更新游标为最后一条记录的ID
		cursor = collections[len(collections)-1].Id
	}

	if err := tx.Commit().Error; err != nil { // 提交事务
		return nil, errors.Wrap(err, "failed to commit transaction")
	}
	return allCollections, nil
}

// QueryCollectionsListed 查询多个集合的上架数量
func (d *Dao) QueryCollectionsListed(ctx context.Context, chain string, collectionAddrs []string) ([]dto.CollectionListed, error) {
	var collectionsListed []dto.CollectionListed
	if len(collectionAddrs) == 0 {
		return collectionsListed, nil
	}

	for _, address := range collectionAddrs {
		count, err := d.KvStore.GetInt(ordermanager.GenCollectionListedKey(chain, address))
		if err != nil {
			return nil, errors.Wrap(err, "failed on set collection listed count")
		}
		collectionsListed = append(collectionsListed, dto.CollectionListed{
			CollectionAddr: address,
			Count:          count,
		})
	}

	return collectionsListed, nil
}

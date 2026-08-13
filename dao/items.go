package dao

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"MiSwap/dto"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"time"
)

const (
	BuyNow   = 1
	HasOffer = 2
	All      = 3
)

const (
	listTime      = 0
	listPriceAsc  = 1
	listPriceDesc = 2
	salePriceDesc = 3
	salePriceAsc  = 4
)

type CollectionItem struct {
	model.Item
	MarketID       int    `json:"market_id"`
	Listing        bool   `json:"listing"`
	OrderID        string `json:"order_id"`
	OrderStatus    int    `json:"order_status"`
	ListMaker      string `json:"list_maker"`
	ListTime       int64  `json:"list_time"`
	ListExpireTime int64  `json:"list_expire_time"`
	ListSalt       int64  `json:"list_salt"`
}

func (d *Dao) QueryListedAmount(ctx context.Context, chain string, addr string) (int64, error) {
	// 1. 从Item表(ci)和订单表(co)联表查询
	// 2. 关联条件:集合地址和tokenID都相同
	// 3. 使用distinct去重统计不同的tokenID数量
	// 4. WHERE条件:
	//   - 指定集合地址
	//   - 订单类型为listing(OrderType=1)
	//   - 订单状态为active(OrderStatus=0)
	//   - 卖家是NFT当前所有者
	//   - 排除marketplace_id=1的订单
	sql := fmt.Sprintf(`select count(distinct(co.token_id)) as counts
						from %s as ci
						join %s co on co.collection_address = ci.collection_address and co.token_id = ci.token_id
						where (co.collection_address=? and co.order_type = ? and
				co.order_status = ? and co.maker = ci.owner and co.marketplace_id != ?)
		`, model.ItemTableName(chain), model.OrderTableName(chain))
	var counts int64
	if err := d.DB.WithContext(ctx).Raw(
		sql,
		addr,
		model.ListingOrder,
		model.OrderStatusActive,
		1,
	).Scan(&counts).Error; err != nil {
		return 0, errcode.NewCustomErr("item dao: failed to get listed item amount")
	}

	return counts, nil
}

func (d *Dao) QueryCollectionBids(ctx context.Context, chain string, addr string, page int, size int) ([]dto.CollectionBids, int64, error) {
	var count int64
	// 统计总记录数
	// SQL解释:统计订单表中符合条件的记录数
	// 条件:1.指定集合地址 2.订单类型为出价单 3.订单状态为活跃 4.未过期
	// 按价格分组统计不同价格的出价数量   【这个统计的是一共有多少个不重复的价格档位】
	if err := d.DB.WithContext(ctx).
		Table(model.OrderTableName(chain)).
		Where("collection_address = ? and order_type = ? and order_status = ? and expire_time > ?",
			addr, model.CollectionBidOrder, model.OrderStatusActive, time.Now().Unix()).
		Group("price").
		Count(&count).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to count user items")
	}

	var bids []dto.CollectionBids
	db := d.DB.WithContext(ctx).Table(model.OrderTableName(chain))

	// 查询出价详情
	// SQL解释:查询订单表获取出价信息
	// 1. 统计每个价格的剩余数量总和(size)
	// 2. 获取价格(price)
	// 3. 计算总价值(total = size * price)
	// 4. 统计不同出价人数(bidders)
	// 条件与上面相同,增加quantity_remaining > 0确保有剩余数量
	// 按价格分组并降序排序,使用分页参数
	if err := db.Select(`
			sum(quantity_remaining) AS size, 
			price,
			sum(quantity_remaining)*price as total,
			COUNT(DISTINCT maker) AS bidders`).
		Where(`collection_address = ? and order_type = ? and order_status = ? 
			   and expire_time > ? and quantity_remaining > 0`,
			addr, model.CollectionBidOrder, model.OrderStatusActive, time.Now().Unix()).
		Group("price").
		Order("price desc").
		Limit(size).
		Offset(size * (page - 1)).
		Scan(&bids).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to query collection bids")
	}

	return bids, count, nil
}

func (d *Dao) QueryItemBids(ctx context.Context, chain string, addr string, tokenID string, page, size int) ([]dto.ItemBid, int64, error) {
	// 构建SQL查询
	// 查询字段包括:市场ID、集合地址、代币ID、订单ID、盐值、事件时间、过期时间
	// 价格、出价人、订单类型、未成交数量、出价总量
	db := d.DB.WithContext(ctx).Table(model.OrderTableName(chain)).
		Select("marketplace_id, collection_address, token_id, order_id, salt, "+
			"event_time, expire_time, price, maker as bidder, order_type, "+
			"quantity_remaining as bid_unfilled, size as bid_size").
		// 查询条件1:集合级别的出价 - 匹配集合地址,订单类型为集合出价,状态为活跃,未过期且有剩余数量
		Where("collection_address = ? and order_type = ? and order_status = ? "+
			"and expire_time > ? and quantity_remaining > 0",
			addr, model.CollectionBidOrder, model.OrderStatusActive, time.Now().Unix()).
		// 查询条件2:Item级别的出价 - 匹配集合地址和代币ID,订单类型为Item出价,其他条件同上
		Or("collection_address = ? and token_id=? and order_type = ? and order_status = ? "+
			"and expire_time > ? and quantity_remaining > 0",
			addr, tokenID, model.ItemBidOrder, model.OrderStatusActive, time.Now().Unix())

	// 查询总记录数
	var count int64
	countTx := db.Session(&gorm.Session{})
	if err := countTx.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrap(db.Error, "failed to count user items")
	}

	// 如果没有记录直接返回
	var itemBids []dto.ItemBid
	if count == 0 {
		return itemBids, count, nil
	}

	// 分页查询出价记录,按价格降序排列
	if err := db.Order("price desc").
		Offset((page - 1) * size).
		Limit(size).
		Scan(&itemBids).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to get user items")
	}

	return itemBids, count, nil
}

func (d *Dao) QueryCollectionItemOrder(ctx context.Context, chain string, filter dto.CollectionItemFilterParams, addr string) ([]*CollectionItem, int64, error) {
	// 如果未指定市场,默认使用OrderBookDex
	if len(filter.Markets) == 0 {
		filter.Markets = []int{model.OrderBookDex}
	}

	// 初始化数据库查询
	db := d.DB.WithContext(ctx).Table(fmt.Sprintf("%s as ci", model.ItemTableName(chain)))
	coTableName := model.OrderTableName(chain)

	// 根据状态过滤查询
	// status: 1-buy now(立即购买), 2-has offer(有报价), 3-all(所有)
	if len(filter.Status) == 1 {
		// 构建基础SELECT语句
		db.Select(
			"ci.id as id, ci.chain_id as chain_id, " +
				"ci.collection_address as collection_address,ci.token_id as token_id, " +
				"ci.name as name, ci.owner as owner, " +
				"min(co.price) as list_price, " +
				"SUBSTRING_INDEX(GROUP_CONCAT(co.marketplace_id ORDER BY co.price,co.marketplace_id),',', 1) AS market_id, " +
				"min(co.price) != 0 as listing")

		// 处理立即购买状态
		if filter.Status[0] == BuyNow {
			// SQL解释:
			// 1. 关联订单表和Item表
			// 2. 条件:集合地址匹配、订单类型为listing、订单状态active、卖家是Item所有者
			db.Joins(fmt.Sprintf(
				"join %s co on co.collection_address=ci.collection_address and co.token_id=ci.token_id",
				coTableName)).
				Where(
					"co.collection_address = ? and co.order_type = ? and co.order_status=? "+
						"and co.maker = ci.owner",
					addr, model.ListingOrder, model.OrderStatusActive)

			// 根据市场ID过滤
			if len(filter.Markets) == 1 {
				db.Where("co.marketplace_id = ?", filter.Markets[0])
			} else if len(filter.Markets) != 5 {
				db.Where("co.marketplace_id in (?)", filter.Markets)
			}

			// 根据tokenID和用户地址过滤
			if filter.TokenID != "" {
				db.Where("co.token_id =?", filter.TokenID)
			}
			if filter.UserAddress != "" {
				db.Where("ci.owner =?", filter.UserAddress)
			}

			db.Group("co.token_id")
		}

		// 处理有报价状态
		if filter.Status[0] == HasOffer {
			// SQL解释:
			// 1. 关联订单表和Item表
			// 2. 条件:集合地址匹配、订单类型为offer、订单状态active
			db.Joins(fmt.Sprintf(
				"join %s co on co.collection_address=ci.collection_address and co.token_id=ci.token_id",
				coTableName)).
				Where(
					"co.collection_address = ? and co.order_type = ? and co.order_status = ?",
					addr, model.OfferOrder, model.OrderStatusActive)

			// 根据市场ID过滤
			if len(filter.Markets) == 1 {
				db.Where("co.marketplace_id = ?", filter.Markets[0])
			} else if len(filter.Markets) != 5 {
				db.Where("co.marketplace_id in (?)", filter.Markets)
			}

			// 根据tokenID和用户地址过滤
			if filter.TokenID != "" {
				db.Where("co.token_id =?", filter.TokenID)
			}
			if filter.UserAddress != "" {
				db.Where("ci.owner =?", filter.UserAddress)
			}

			db.Group("co.token_id")
		}
	} else if len(filter.Status) == 2 {
		// 处理同时有买卖订单的情况
		// SQL解释:
		// 1. 关联订单表和Item表
		// 2. 条件:订单状态active、卖家是Item所有者
		// 3. 分组后需同时存在listing和offer订单
		// 选择字段:
		// 1. 基本信息:id、chain_id、collection_address、token_id、name、owner
		// 2. list_price: 取最低挂单价格(min(co.price))
		// 3. market_id: 使用SUBSTRING_INDEX和GROUP_CONCAT组合取最低价格对应的市场ID
		//    - GROUP_CONCAT按价格和市场ID排序,将marketplace_id连接成字符串
		//    - SUBSTRING_INDEX取第一个值,即最低价格对应的市场ID
		db.Select(
			"ci.id as id, ci.chain_id as chain_id," +
				"ci.collection_address as collection_address,ci.token_id as token_id, " +
				"ci.name as name, ci.owner as owner, " +
				"min(co.price) as list_price, " +
				"SUBSTRING_INDEX(GROUP_CONCAT(co.marketplace_id ORDER BY co.price,co.marketplace_id),',', 1) AS market_id")

		db.Joins(fmt.Sprintf(
			"join %s co on co.collection_address=ci.collection_address and co.token_id=ci.token_id",
			coTableName)).
			Where(
				"co.collection_address = ? and co.order_status=? and co.maker = ci.owner",
				addr, model.OrderStatusActive)

		// 根据市场ID过滤
		if len(filter.Markets) == 1 {
			db.Where("co.marketplace_id = ?", filter.Markets[0])
		} else if len(filter.Markets) != 5 {
			db.Where("co.marketplace_id in (?)", filter.Markets)
		}

		// 根据tokenID和用户地址过滤
		if filter.TokenID != "" {
			db.Where("co.token_id =?", filter.TokenID)
		}
		if filter.UserAddress != "" {
			db.Where("ci.owner =?", filter.UserAddress)
		}

		db.Group("co.token_id").Having(
			"min(co.type)=? and max(co.type)=?",
			model.ListingOrder, model.OfferOrder)

	} else {
		// 处理所有状态
		// SQL解释:
		// 1. 子查询获取每个token的最低listing价格
		// 2. 左连接子查询结果到Item表
		// 3. 根据条件过滤
		subQuery := d.DB.WithContext(ctx).Table(
			fmt.Sprintf("%s as cis", model.ItemTableName(chain))).
			Select(
				"cis.id as item_id,cis.collection_address as collection_address,"+
					"cis.token_id as token_id, cis.owner as owner, cos.order_id as order_id, "+
					"min(cos.price) as list_price, "+
					"SUBSTRING_INDEX(GROUP_CONCAT(cos.marketplace_id ORDER BY cos.price,cos.marketplace_id),',', 1) AS market_id, "+
					"min(cos.price) != 0 as listing").
			Joins(fmt.Sprintf(
				"join %s cos on cos.collection_address=cis.collection_address and cos.token_id=cis.token_id",
				coTableName)).
			Where(
				"cos.collection_address = ? and cos.order_type = ? and cos.order_status=? "+
					"and cos.maker = cis.owner",
				addr, model.ListingOrder, model.OrderStatusActive)

		if len(filter.Markets) == 1 {
			subQuery.Where("cos.marketplace_id = ?", filter.Markets[0])
		} else if len(filter.Markets) != 5 {
			subQuery.Where("cos.marketplace_id in (?)", filter.Markets)
		}
		subQuery.Group("cos.token_id")

		db.Joins("left join (?) co on co.collection_address=ci.collection_address and co.token_id=ci.token_id",
			subQuery).
			Select(
				"ci.id as id, ci.chain_id as chain_id," +
					"ci.collection_address as collection_address, ci.token_id as token_id, " +
					"ci.name as name, ci.owner as owner, " +
					"co.list_price as list_price, co.market_id as market_id, co.listing as listing").
			Where(fmt.Sprintf("ci.collection_address = '%s'", addr))

		if filter.TokenID != "" {
			db.Where(fmt.Sprintf("ci.token_id = '%s'", filter.TokenID))
		}
		if filter.UserAddress != "" {
			db.Where(fmt.Sprintf("ci.owner = '%s'", filter.UserAddress))
		}
	}

	// 统计总记录数
	var count int64
	countTx := db.Session(&gorm.Session{})
	if err := countTx.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrap(db.Error, "failed to count items")
	}

	// 处理排序
	if len(filter.Status) == 0 {
		db.Order("listing desc")
	}

	if filter.Sort == 0 {
		filter.Sort = listPriceAsc
	}

	// 根据不同排序条件设置ORDER BY
	switch filter.Sort {
	case listTime:
		db.Order("list_time desc,ci.id asc")
	case listPriceAsc:
		db.Order("list_price asc, ci.id asc")
	case listPriceDesc:
		db.Order("list_price desc,ci.id asc")
	case salePriceDesc:
		db.Order("sale_price desc,ci.id asc")
	case salePriceAsc:
		db.Order("sale_price = 0,sale_price asc,ci.id asc")
	}

	// 执行分页查询
	var items []*CollectionItem
	db.Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Scan(&items)

	if db.Error != nil {
		return nil, 0, errors.Wrap(db.Error, "failed to get query items info")
	}

	return items, count, nil
}

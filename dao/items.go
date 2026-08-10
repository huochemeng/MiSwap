package dao

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"context"
	"fmt"
)

const OrderType = 1
const OrderStatus = 0

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
		OrderType,
		OrderStatus,
		1,
	).Scan(&counts).Error; err != nil {
		return 0, errcode.NewCustomErr("item dao: failed to get listed item amount")
	}

	return counts, nil
}

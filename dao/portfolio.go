package dao

import (
	"MiSwap/base/stores/gdb/model"
	"MiSwap/dto"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"time"
)

// QueryMultiChainUserCollectionInfos 查询用户在多条链上的Collection信息
func (d *Dao) QueryMultiChainUserCollectionInfos(ctx context.Context, chainID []int,
	chainNames []string, userAddrs []string) ([]dto.UserCollections, error) {
	var userCollections []dto.UserCollections

	// 构建用户地址参数字符串,格式: 'addr1','addr2',...
	var userAddrsParam string
	for i, addr := range userAddrs {
		userAddrsParam += fmt.Sprintf(`'%s'`, addr)
		if i < len(userAddrs)-1 {
			userAddrsParam += ","
		}
	}

	// SQL语句组成部分
	sqlHead := "SELECT * FROM ("
	// 按照地板价*持有数量降序排序
	sqlTail := ") as combined ORDER BY combined.floor_price * " +
		"CAST(combined.item_count AS DECIMAL) DESC"
	var sqlMids []string

	// 遍历每条链,构建子查询
	for _, chainName := range chainNames {
		sqlMid := "("
		// 查询Collection基本信息和用户持有数量
		sqlMid += "select " +
			"gc.address as address, " +
			"gc.name as name, " +
			"gc.floor_price as floor_price, " +
			"gc.chain_id as chain_id, " +
			"gc.item_amount as item_amount, " +
			"gc.symbol as symbol, " +
			"gc.image_uri as image_uri, " +
			"count(*) as item_count "
		// 从Collection表和Item表联表查询
		sqlMid += fmt.Sprintf("from %s as gc ", model.CollectionTableName(chainName))
		sqlMid += fmt.Sprintf("join %s as gi ", model.ItemTableName(chainName))
		sqlMid += "on gc.address = gi.collection_address "
		// 过滤指定用户持有的Item
		sqlMid += fmt.Sprintf("where gi.owner in (%s) ", userAddrsParam)
		sqlMid += "group by gc.address"
		sqlMid += ")"

		sqlMids = append(sqlMids, sqlMid)
	}

	// 组装完整SQL,使用UNION ALL合并多链结果
	sql := sqlHead
	for i := 0; i < len(sqlMids); i++ {
		if i != 0 {
			sql += " UNION ALL "
		}
		sql += sqlMids[i]
	}
	sql += sqlTail

	// 执行SQL查询
	if err := d.DB.WithContext(ctx).Raw(sql).Scan(&userCollections).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get user multi chain collection infos")
	}

	return userCollections, nil
}

// QueryListedAmountEachCollection 查询多个集合中已上架NFT的数量
func (d *Dao) QueryListedAmountEachCollection(ctx context.Context, chain string, collectionAddrs []string, userAddrs []string) ([]dto.CollectionInfo, error) {
	var counts []dto.CollectionInfo

	// SQL解释:
	// 1. 从Item表(ci)和订单表(co)联表查询
	// 2. 选择字段:
	//    - ci.collection_address 作为 address
	//    - count(distinct co.token_id) 作为 list_amount,统计每个集合中不重复的tokenID数量
	// 3. 关联条件:集合地址和tokenID都相同
	// 4. WHERE条件:
	//    - 集合地址在给定列表中
	//    - NFT所有者在给定用户列表中
	//    - 订单类型为listing(OrderType=1)
	//    - 订单状态为active(OrderStatus=0)
	//    - 卖家是NFT当前所有者
	//    - 排除marketplace_id=1的订单
	// 5. 按集合地址分组,获取每个集合的统计结果
	sql := fmt.Sprintf(`SELECT  ci.collection_address as address, count(distinct (co.token_id)) as list_amount
			FROM %s as ci
					join %s co on co.collection_address = ci.collection_address and co.token_id = ci.token_id
			WHERE (co.collection_address in (?) and ci.owner in (?) and co.order_type = ? and
				co.order_status = ? and co.maker = ci.owner and co.marketplace_id != ?) group by ci.collection_address`,
		model.ItemTableName(chain), model.OrderTableName(chain))
	if err := d.DB.WithContext(ctx).Raw(
		sql,
		collectionAddrs,
		userAddrs,
		model.ListingOrder,
		model.OrderStatusActive,
		1,
	).Scan(&counts).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get listed item amount")
	}

	return counts, nil
}

// QueryMultiChainUserItemInfos 查询用户拥有nft的Item基本信息，list信息和bid信息，从Item表和Activity表中查询
// 参数:
// - chain: 链名称列表
// - userAddrs: 用户地址列表
// - contractAddrs: 合约地址列表
// - page: 页码
// - pageSize: 每页大小
func (d *Dao) QueryMultiChainUserItemInfos(ctx context.Context, chain []string, userAddrs []string,
	contractAddrs []string, page, pageSize int) ([]dto.PortfolioItemInfo, int64, error) {

	if len(chain) == 0 || len(userAddrs) == 0 {
		return nil, 0, nil
	}

	// 构建参数占位符与参数切片，避免 SQL 注入
	userPlaceholders := make([]string, len(userAddrs))
	args := make([]interface{}, 0, len(userAddrs)+len(contractAddrs)*2+2)
	for i, addr := range userAddrs {
		userPlaceholders[i] = "?"
		args = append(args, addr)
	}
	userInClause := strings.Join(userPlaceholders, ",")

	var contractInClause string
	if len(contractAddrs) > 0 {
		contractPlaceholders := make([]string, len(contractAddrs))
		for i, addr := range contractAddrs {
			contractPlaceholders[i] = "?"
			args = append(args, addr)
		}
		contractInClause = strings.Join(contractPlaceholders, ",")
	}

	// 为每条链构建子查询
	subQueries := make([]string, 0, len(chain))
	for _, chainName := range chain {
		itemTable := model.ItemTableName(chainName)
		activityTable := model.ActivityTableName(chainName)

		sub := fmt.Sprintf(`(
			SELECT
				gi.chain_id,
				gi.collection_address,
				gi.token_id,
				gi.name,
				gi.owner,
				sub.last_event_time AS owned_time
			FROM %s gi
			LEFT JOIN (
				SELECT
					sgi.collection_address,
					sgi.token_id,
					MAX(sga.event_time) AS last_event_time
				FROM %s sgi
				JOIN %s sga
					ON sgi.collection_address = sga.collection_address
					AND sgi.token_id = sga.token_id
				WHERE sgi.owner IN (%s)
				  AND sga.activity_type = ?`,
			itemTable, itemTable, activityTable, userInClause)

		subArgs := make([]interface{}, len(userAddrs))
		copy(subArgs, args[:len(userAddrs)])
		subArgs = append(subArgs, model.Sale)

		if len(contractAddrs) > 0 {
			sub += fmt.Sprintf(" AND sgi.collection_address IN (%s)", contractInClause)
			subArgs = append(subArgs, args[len(userAddrs):len(userAddrs)+len(contractAddrs)]...)
		}

		sub += fmt.Sprintf(`
				GROUP BY sgi.collection_address, sgi.token_id
			) sub
				ON gi.collection_address = sub.collection_address
				AND gi.token_id = sub.token_id
			WHERE gi.owner IN (%s)`, userInClause)

		if len(contractAddrs) > 0 {
			sub += fmt.Sprintf(" AND gi.collection_address IN (%s)", contractInClause)
		}

		sub += ")"

		subQueries = append(subQueries, sub)
		_ = subArgs // subArgs 仅用于说明参数顺序，实际统一由外层 args 管理
	}

	unionBody := strings.Join(subQueries, " UNION ALL ")

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS combined", unionBody)
	dataSQL := fmt.Sprintf("SELECT * FROM (%s) AS combined ORDER BY combined.owned_time DESC LIMIT ? OFFSET ?", unionBody)

	offset := (page - 1) * pageSize
	dataArgs := append(args, pageSize, offset)

	var count int64
	if err := d.DB.WithContext(ctx).Raw(countSQL, args...).Scan(&count).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to count user multi chain items")
	}

	var items []dto.PortfolioItemInfo
	if err := d.DB.WithContext(ctx).Raw(dataSQL, dataArgs...).Scan(&items).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to get user multi chain items")
	}

	return items, count, nil
}

// QueryCollectionsBestBid 查询多个集合的最高出价信息
// 该函数主要功能:
// 1. 根据链名称、用户地址和集合地址列表查询每个集合的最高出价订单
// 2. 如果指定了用户地址,则排除该用户的出价
// 3. 返回每个集合中价格最高的有效订单(未过期且有剩余数量)
func (d *Dao) QueryCollectionsBestBid(ctx context.Context, chain string, userAddr string, collectionAddrs []string) ([]*model.Order, error) {
	var bestBid []*model.Order

	// SQL解释:
	// 1. 主查询:从订单表中查询订单详细信息
	sql := fmt.Sprintf(`
		SELECT collection_address, order_id, price,event_time, expire_time, salt, maker, order_type, quantity_remaining, size  
		FROM %s `, model.OrderTableName(chain))

	// 2. 子查询:获取每个集合的最高出价
	sql += fmt.Sprintf(`where (collection_address,price) in (SELECT collection_address, max(price) as price FROM %s `, model.OrderTableName(chain))

	// 3. 子查询条件:
	//   - 集合地址在给定列表中
	//   - 订单类型为集合出价单
	//   - 订单状态为活跃
	//   - 剩余数量大于0
	//   - 未过期
	//   - 如果指定用户地址,则排除该用户
	sql += `where collection_address in (?) and order_type = ? and order_status = ? and quantity_remaining > 0 and expire_time > ? `
	if userAddr != "" {
		sql += fmt.Sprintf(" and maker != '%s'", userAddr)
	}
	sql += `group by collection_address ) `

	// 4. 主查询条件:与子查询条件相同
	sql += `and order_type = ? and order_status = ? and quantity_remaining > 0 and expire_time > ? `
	if userAddr != "" {
		sql += fmt.Sprintf(" and maker != '%s'", userAddr)
	}

	// 5. 执行查询
	if err := d.DB.Raw(sql, collectionAddrs, model.CollectionBidOrder, model.OrderStatusActive, time.Now().Unix(), model.CollectionBidOrder, model.OrderStatusActive, time.Now().Unix()).Scan(&bestBid).Error; err != nil {
		return bestBid, errors.Wrap(err, "failed to get item best bids")
	}

	return bestBid, nil
}

// QueryMultiChainCollectionsInfo 批量查询多条链上的NFT集合信息
// 参数collectionAddrs是一个二维数组,每个元素包含[合约地址,链名称]
// 返回多条链上的NFT集合信息列表
func (d *Dao) QueryMultiChainCollectionsInfo(ctx context.Context, collectionAddrs [][]string) ([]model.Collection, error) {
	addrs := removeRepeatedElementArr(collectionAddrs)
	var collections []model.Collection
	var collection model.Collection
	for _, collectionAddr := range addrs {
		if err := d.DB.WithContext(ctx).Table(model.CollectionTableName(collectionAddr[1])).
			Select(collectionDetailFields).Where("address = ?", collectionAddr[0]).
			Scan(&collection).Error; err != nil {
			return nil, errors.Wrap(err, "failed to get collection info")
		}
		collections = append(collections, collection)
	}

	return collections, nil
}

// QueryMultiChainUserItemsListInfo 查询多条链上用户NFT Item的挂单信息
// 主要功能:
// 1. 根据用户地址列表和Item信息列表查询每个Item的挂单状态
// 2. 支持跨链查询,按链名称分组处理
// 3. 返回每个Item的挂单价格、市场ID等信息
func (d *Dao) QueryMultiChainUserItemsListInfo(ctx context.Context, userAddrs []string,
	itemInfos []MultiChainItemInfo) ([]*CollectionItem, error) {
	var collectionItems []*CollectionItem

	// 构建用户地址参数字符串: 'addr1','addr2',...
	var userAddrsParam string
	for i, addr := range userAddrs {
		userAddrsParam += fmt.Sprintf(`'%s'`, addr)
		if i < len(userAddrs)-1 {
			userAddrsParam += ","
		}
	}

	// 按链名称对Item信息分组
	chainItems := make(map[string][]MultiChainItemInfo)
	for _, itemInfo := range itemInfos {
		items, ok := chainItems[strings.ToLower(itemInfo.ChainName)]
		if ok {
			items = append(items, itemInfo)
			chainItems[strings.ToLower(itemInfo.ChainName)] = items
		} else {
			chainItems[strings.ToLower(itemInfo.ChainName)] = []MultiChainItemInfo{itemInfo}
		}
	}

	// SQL语句组成部分
	sqlHead := "SELECT * FROM (" // 外层查询开始
	sqlTail := ") as combined"   // 外层查询结束
	var sqlMids []string         // 存储每条链的子查询

	// 遍历每条链构建子查询
	for chainName, items := range chainItems {
		// 构建IN查询条件: (('addr1','id1'),('addr2','id2'),...)
		tmpStat := fmt.Sprintf("(('%s','%s')", items[0].CollectionAddress, items[0].TokenID)
		for i := 1; i < len(items); i++ {
			tmpStat += fmt.Sprintf(",('%s','%s')", items[i].CollectionAddress, items[i].TokenID)
		}
		tmpStat += ") "

		// 构建子查询SQL
		sqlMid := "("
		// 选择字段:Item基本信息、最低挂单价格、市场ID等
		sqlMid += "select ci.id as id, ci.chain_id as chain_id,"
		sqlMid += "ci.collection_address as collection_address,ci.token_id as token_id, ci.name as name, ci.owner as owner,"
		sqlMid += "min(co.price) as list_price, " +
			"SUBSTRING_INDEX(GROUP_CONCAT(co.marketplace_id ORDER BY co.price,co.marketplace_id),',', 1) " +
			"AS market_id, min(co.price) != 0 as listing "
		// 关联Item表和订单表
		sqlMid += fmt.Sprintf("from %s as ci ", model.ItemTableName(chainName))
		sqlMid += fmt.Sprintf("join %s co ", model.OrderTableName(chainName))
		sqlMid += "on co.collection_address=ci.collection_address and co.token_id=ci.token_id "
		// 查询条件:匹配集合地址和tokenID、订单类型为listing、状态为active、卖家是Item所有者
		sqlMid += "where (co.collection_address,co.token_id) in "
		sqlMid += tmpStat
		sqlMid += fmt.Sprintf("and co.order_type = %d and co.order_status=%d "+
			"and co.maker = ci.owner and co.maker in (%s) ",
			model.ListingOrder, model.OrderStatusActive, userAddrsParam)
		sqlMid += "group by co.collection_address,co.token_id"
		sqlMid += ")"

		sqlMids = append(sqlMids, sqlMid)
	}

	// 使用UNION ALL组合所有子查询
	sql := sqlHead
	for i := 0; i < len(sqlMids); i++ {
		if i != 0 {
			sql += " UNION ALL " // 使用UNION ALL合并结果集
		}
		sql += sqlMids[i]
	}
	sql += sqlTail

	// 执行SQL查询
	if err := d.DB.WithContext(ctx).Raw(sql).Scan(&collectionItems).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query user multi chain items list info")
	}

	return collectionItems, nil
}

// QueryMultiChainListingInfo 查询多条链上的NFT挂单信息
func (d *Dao) QueryMultiChainListingInfo(ctx context.Context, priceInfos []MultiChainItemPriceInfo) ([]model.Order, error) {
	var orders []model.Order

	// 按链名称对价格信息分组
	chainItemPrices := make(map[string][]MultiChainItemPriceInfo)
	for _, priceInfo := range priceInfos {
		items, ok := chainItemPrices[strings.ToLower(priceInfo.ChainName)]
		if ok {
			items = append(items, priceInfo)
			chainItemPrices[strings.ToLower(priceInfo.ChainName)] = items
		} else {
			chainItemPrices[strings.ToLower(priceInfo.ChainName)] = []MultiChainItemPriceInfo{priceInfo}
		}
	}

	// SQL语句组成部分
	sqlHead := "SELECT * FROM (" // 外层查询开始
	sqlTail := ") as combined"   // 外层查询结束
	var sqlMids []string         // 存储每条链的子查询

	// 遍历每条链构建子查询
	for chainName, priceInfos := range chainItemPrices {
		// 构建IN查询条件: (('addr1','id1','maker1',status1,price1),...)
		tmpStat := fmt.Sprintf("(('%s','%s','%s',%d, %s)", priceInfos[0].CollectionAddress, priceInfos[0].TokenID, priceInfos[0].Maker, priceInfos[0].OrderStatus, priceInfos[0].Price.String())
		for i := 1; i < len(priceInfos); i++ {
			tmpStat += fmt.Sprintf(",('%s','%s','%s',%d, %s)", priceInfos[i].CollectionAddress, priceInfos[i].TokenID, priceInfos[i].Maker, priceInfos[i].OrderStatus, priceInfos[i].Price.String())
		}
		tmpStat += ") "

		// 构建子查询SQL:
		// 1. 选择订单的基本字段
		// 2. 从对应链的订单表查询
		// 3. 匹配集合地址、代币ID、创建者、状态和价格
		sqlMid := "("
		sqlMid += "select collection_address,token_id,order_id,salt,event_time,expire_time,maker "
		sqlMid += fmt.Sprintf("from %s ", model.OrderTableName(chainName))
		sqlMid += "where (collection_address,token_id,maker,order_status,price) in "
		sqlMid += tmpStat
		sqlMid += ")"

		sqlMids = append(sqlMids, sqlMid)
	}

	// 使用UNION ALL组合所有子查询
	sql := sqlHead
	for i := 0; i < len(sqlMids); i++ {
		if i != 0 {
			sql += " UNION ALL " // 使用UNION ALL合并结果集
		}
		sql += sqlMids[i]
	}
	sql += sqlTail

	// 执行SQL查询
	if err := d.DB.WithContext(ctx).Raw(sql).Scan(&orders).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query user multi chain order list info")
	}

	return orders, nil
}

// QueryMultiChainCollectionsItemsImage 查询多条链上NFT Item的图片信息
// 主要功能:
// 1. 按链名称对输入的Item信息进行分组
// 2. 构建多条链的联合查询SQL
// 3. 返回所有链上Item的图片信息
func (d *Dao) QueryMultiChainCollectionsItemsImage(ctx context.Context, itemInfos []MultiChainItemInfo) ([]model.ItemExternal, error) {
	var itemsExternal []model.ItemExternal

	// SQL语句组成部分
	sqlHead := "SELECT * FROM (" // 外层查询开始
	sqlTail := ") as combined"   // 外层查询结束
	var sqlMids []string         // 存储每条链的子查询

	// 按链名称对Item信息分组
	chainItems := make(map[string][]MultiChainItemInfo)
	for _, itemInfo := range itemInfos {
		items, ok := chainItems[strings.ToLower(itemInfo.ChainName)]
		if ok {
			items = append(items, itemInfo)
			chainItems[strings.ToLower(itemInfo.ChainName)] = items
		} else {
			chainItems[strings.ToLower(itemInfo.ChainName)] = []MultiChainItemInfo{itemInfo}
		}
	}

	// 遍历每条链构建子查询
	for chainName, items := range chainItems {
		// 构建IN查询条件: (('addr1','id1'),('addr2','id2'),...)
		tmpStat := fmt.Sprintf("(('%s','%s')", items[0].CollectionAddress, items[0].TokenID)
		for i := 1; i < len(items); i++ {
			tmpStat += fmt.Sprintf(",('%s','%s')", items[i].CollectionAddress, items[i].TokenID)
		}
		tmpStat += ") "

		// 构建子查询SQL:
		// 1. 选择Item的图片相关字段
		// 2. 从对应链的external表查询
		// 3. 匹配集合地址和tokenID
		sqlMid := "("
		sqlMid += "select collection_address, token_id, is_uploaded_oss, image_uri, oss_uri "
		sqlMid += fmt.Sprintf("from %s ", model.ItemExternalTableName(chainName))
		sqlMid += "where (collection_address,token_id) in "
		sqlMid += tmpStat
		sqlMid += ")"

		sqlMids = append(sqlMids, sqlMid)
	}

	// 使用UNION ALL组合所有子查询
	sql := sqlHead
	for i := 0; i < len(sqlMids); i++ {
		if i != 0 {
			sql += " UNION ALL " // 使用UNION ALL合并结果集
		}
		sql += sqlMids[i]
	}
	sql += sqlTail

	// 执行SQL查询
	if err := d.DB.WithContext(ctx).Raw(sql).Scan(&itemsExternal).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query multi chain items external info")
	}

	return itemsExternal, nil
}

// QueryMultiChainUserListingItemInfos 查询多链上用户挂单Item信息
func (d *Dao) QueryMultiChainUserListingItemInfos(ctx context.Context, chain []string, userAddrs []string,
	contractAddrs []string, page, pageSize int) ([]dto.PortfolioItemInfo, int64, error) {
	var count int64
	var items []dto.PortfolioItemInfo

	// 构建用户地址参数字符串
	var userAddrsParam string
	for i, addr := range userAddrs {
		userAddrsParam += fmt.Sprintf(`'%s'`, addr)
		if i < len(userAddrs)-1 {
			userAddrsParam += ","
		}
	}

	// SQL语句头部
	sqlCntHead := "SELECT COUNT(*) FROM ("
	sqlHead := "SELECT * FROM ("
	// 分页SQL
	sqlTail := fmt.Sprintf(") as combined ORDER BY combined.owned_time DESC LIMIT %d OFFSET %d",
		pageSize, page-1)
	var sqlMids []string

	// 遍历每条链构建SQL
	for _, chainName := range chain {
		sqlMid := "("
		// 查询Item基本信息和最后交易时间
		sqlMid += "select gi.chain_id as chain_id, gi.collection_address as collection_address, " +
			"gi.token_id as token_id, gi.name as name, gi.owner as owner, " +
			"sub.last_event_time as owned_time "
		sqlMid += fmt.Sprintf("from %s gi ", model.ItemTableName(chainName))
		sqlMid += "left join "
		// 子查询获取每个Item最后的交易时间
		sqlMid += "(select sgi.collection_address, sgi.token_id, " +
			"max(sga.event_time) as last_event_time "
		sqlMid += fmt.Sprintf("from %s sgi join %s sga ",
			model.ItemTableName(chainName), model.ActivityTableName(chainName))
		sqlMid += "on sgi.collection_address = sga.collection_address " +
			"and sgi.token_id = sga.token_id "
		// 过滤条件:指定用户和Sale类型活动
		sqlMid += fmt.Sprintf("where sgi.owner in (%s) and sga.activity_type = %d ",
			userAddrsParam, model.Sale)

		// 添加合约地址过滤
		if len(contractAddrs) > 0 {
			sqlMid += fmt.Sprintf("and sgi.collection_address in ('%s'", contractAddrs[0])
			for i := 1; i < len(contractAddrs); i++ {
				sqlMid += fmt.Sprintf(",'%s'", contractAddrs[i])
			}
			sqlMid += ") "
		}
		sqlMid += "group by sgi.collection_address, sgi.token_id) sub "
		sqlMid += "on gi.collection_address = sub.collection_address " +
			"and gi.token_id = sub.token_id "

		// 主查询过滤条件
		sqlMid += fmt.Sprintf("where gi.owner in (%s) ", userAddrsParam)
		if len(contractAddrs) > 0 {
			sqlMid += fmt.Sprintf("and gi.collection_address in ('%s'", contractAddrs[0])
			for i := 1; i < len(contractAddrs); i++ {
				sqlMid += fmt.Sprintf(",'%s'", contractAddrs[i])
			}
			sqlMid += ")"
		}
		sqlMid += ")"

		sqlMids = append(sqlMids, sqlMid)
	}

	// 使用UNION ALL合并多链结果
	sqlCnt := sqlCntHead
	sql := sqlHead
	for i := 0; i < len(sqlMids); i++ {
		if i != 0 {
			sql += " UNION ALL "
			sqlCnt += " UNION ALL "
		}
		sql += sqlMids[i]
		sqlCnt += sqlMids[i]
	}
	sql += sqlTail
	sqlCnt += ") as combined"

	// 执行SQL查询
	if err := d.DB.WithContext(ctx).Raw(sqlCnt).Scan(&count).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to count user multi chain items")
	}
	if err := d.DB.WithContext(ctx).Raw(sql).Scan(&items).Error; err != nil {
		return nil, 0, errors.Wrap(err, "failed to get user multi chain items")
	}

	return items, count, nil
}

// QueryMultiChainUserItemsExpireListInfo 查询多条链上用户Item的过期挂单信息
// 主要功能:
// 1. 根据用户地址列表和Item信息列表查询每个Item的挂单状态
// 2. 支持查询多条链上的Item信息
// 3. 返回Item的基本信息和挂单信息(价格、市场等)
func (d *Dao) QueryMultiChainUserItemsExpireListInfo(ctx context.Context, userAddrs []string,
	itemInfos []MultiChainItemInfo) ([]*CollectionItem, error) {
	var collectionItems []*CollectionItem

	// 构建用户地址参数字符串: 'addr1','addr2',...
	var userAddrsParam string
	for i, addr := range userAddrs {
		userAddrsParam += fmt.Sprintf(`'%s'`, addr)
		if i < len(userAddrs)-1 {
			userAddrsParam += ","
		}
	}

	// SQL语句组成部分
	sqlHead := "SELECT * FROM (" // 外层查询开始
	sqlTail := ") as combined"   // 外层查询结束
	var sqlMids []string         // 存储每个Item的子查询

	// 构建IN查询条件: (('addr1','id1'),('addr2','id2'),...)
	tmpStat := fmt.Sprintf("(('%s','%s')", itemInfos[0].CollectionAddress, itemInfos[0].TokenID)
	for i := 1; i < len(itemInfos); i++ {
		tmpStat += fmt.Sprintf(",('%s','%s')", itemInfos[i].CollectionAddress, itemInfos[i].TokenID)
	}
	tmpStat += ") "

	// 遍历每个Item构建子查询
	for _, info := range itemInfos {
		sqlMid := "("
		// 选择字段:Item基本信息、最低挂单价格、市场ID等
		sqlMid += "select ci.id as id, ci.chain_id as chain_id,"
		sqlMid += "ci.collection_address as collection_address,ci.token_id as token_id, " +
			"ci.name as name, ci.owner as owner,"
		sqlMid += "min(co.price) as list_price, " +
			"SUBSTRING_INDEX(GROUP_CONCAT(co.marketplace_id ORDER BY co.price,co.marketplace_id),',', 1) " +
			"AS market_id, min(co.price) != 0 as listing "

		// 关联Item表和订单表
		sqlMid += fmt.Sprintf("from %s as ci ", model.ItemTableName(info.ChainName))
		sqlMid += fmt.Sprintf("join %s co ", model.OrderTableName(info.ChainName))
		sqlMid += "on co.collection_address=ci.collection_address and co.token_id=ci.token_id "

		// 查询条件:
		// 1. 匹配集合地址和tokenID
		// 2. 订单类型为listing
		// 3. 订单状态为active或expired
		// 4. 卖家是Item所有者且在用户列表中
		sqlMid += "where (co.collection_address,co.token_id) in "
		sqlMid += tmpStat
		sqlMid += fmt.Sprintf("and co.order_type = %d and (co.order_status=%d or co.order_status=%d) "+
			"and co.maker = ci.owner and co.maker in (%s) ",
			model.ListingOrder, model.OrderStatusActive, model.OrderStatusExpired, userAddrsParam)
		sqlMid += "group by co.collection_address,co.token_id"
		sqlMid += ")"

		sqlMids = append(sqlMids, sqlMid)
	}

	// 使用UNION ALL组合所有子查询
	sql := sqlHead
	for i := 0; i < len(sqlMids); i++ {
		if i != 0 {
			sql += " UNION ALL " // 使用UNION ALL合并结果集
		}
		sql += sqlMids[i]
	}
	sql += sqlTail

	// 执行SQL查询
	if err := d.DB.WithContext(ctx).Raw(sql).Scan(&collectionItems).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query user multi chain items list info")
	}

	return collectionItems, nil
}

func (d *Dao) QueryUserBids(ctx context.Context, chain string, userAddrs []string, contractAddrs []string) ([]model.Order, error) {
	var userBids []model.Order

	// SQL解释:
	// 1. 从订单表中查询订单详细信息
	// 2. 选择字段包括:集合地址、代币ID、订单ID、订单类型、剩余数量等
	// 3. WHERE条件:
	//    - maker在给定用户地址列表中
	//    - 订单类型为Item出价或集合出价
	//    - 订单状态为活跃
	//    - 剩余数量大于0
	db := d.DB.WithContext(ctx).
		Table(model.OrderTableName(chain)).
		Select("collection_address, token_id, order_id, token_id,order_type,"+
			"quantity_remaining, size, event_time, price, salt, expire_time").
		Where("maker in (?) and order_type in (?,?) and order_status = ? and quantity_remaining > 0",
			userAddrs, model.ItemBidOrder, model.CollectionBidOrder, model.OrderStatusActive)

	// 如果指定了合约地址列表,添加集合地址过滤条件
	if len(contractAddrs) != 0 {
		db.Where("collection_address in (?)", contractAddrs)
	}

	if err := db.Scan(&userBids).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get user bids")
	}

	return userBids, nil
}

// QueryCollectionsInfo 批量查询指定链上的NFT集合信息
func (d *Dao) QueryCollectionsInfo(ctx context.Context, chain string, collectionAddrs []string) ([]model.Collection, error) {
	addrs := removeRepeatedElement(collectionAddrs)
	var collections []model.Collection
	if err := d.DB.WithContext(ctx).Table(model.CollectionTableName(chain)).
		Select(collectionDetailFields).Where("address in (?)", addrs).
		Scan(&collections).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get collection info")
	}

	return collections, nil
}

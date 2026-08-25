package dao

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"context"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"time"
)

type CollectionTrade struct {
	ContractAddress string          `json:"contract_address"`
	Volume          int64           `json:"volume"`
	Turnover        decimal.Decimal `json:"turnover"`
	TurnoverChange  int             `json:"turnover_change"`
	PreFloorPrice   decimal.Decimal `json:"pre_floor_price"`
	FloorChange     int             `json:"floor_change"`
}

var periodToDuration = map[string]time.Duration{
	"15m": 15 * time.Minute,
	"1h":  1 * time.Hour,
	"6h":  6 * time.Hour,
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
}

// GetTradeInfoByCollection 获取指定时间段内集合的交易统计信息
func (d *Dao) GetTradeInfoByCollection(ctx context.Context, chain string, addr string, period string) (*CollectionTrade, error) {
	// 查询当前时间段的交易信息
	//成交量
	var volume int64
	// 成交额
	var turnover decimal.Decimal

	var floorPrice decimal.Decimal

	//获取间隔时间段对应的时间
	duration, ok := periodToDuration[period]
	if !ok {
		return nil, errcode.ErrInvalidParams.WithMsg("invalid period: " + period)
	}
	//	计算开始和结束时间，now只运行一次，避免多次调用，存在时间差
	now := time.Now()
	startTime := now.Add(-duration)
	endTime := now
	// 统计当前时间段内的交易数量和总交易额,以及地板价
	err := d.DB.WithContext(ctx).Table(model.ActivityTableName(chain)).
		Where("collection_address = ? and activity_type = ? and event_time >= ? and event_time <= ?",
			addr, model.Sale, startTime, endTime).
		Select("count(*) as volume, "+
			"COALESCE(SUM(price), 0) as turnover,"+
			"COALESCE(MIN(price), 0) as floor_price"). //coalesce作用：合并为一个值。sum计算没有值的时候为null，通过该函数返回0
		Row().Scan(&volume, &turnover, &floorPrice)
	if err != nil {
		return nil, errcode.NewCustomErr("dao： failed to get trade volume, turnover and floorPrice")
	}
	//计算上一个时间段的开始和结束时间
	prevStartTime := startTime.Add(-duration)
	prevEndTime := startTime

	//上一个时间段的成交额和地板价
	var prevTurnover decimal.Decimal
	var prevFloorPrice decimal.Decimal

	// 获取上一个时间段的成交量，成交额以及地板价
	err = d.DB.WithContext(ctx).Table(model.ActivityTableName(chain)).
		Where("collection_address = ? and activity_type = ? and event_time >= ? and event_time <= ?",
			addr, model.Sale, prevStartTime, prevEndTime).
		Select("COALESCE(SUM(price), 0) as turnover,"+
			"COALESCE(MIN(price), 0) as floor_price").
		Row().Scan(&prevTurnover, &prevFloorPrice)
	if err != nil {
		return nil, errcode.NewCustomErr("dao： failed to get previous time trade volume, turnover and floorPrice")
	}
	//计算成交额和地板价的变化百分比
	turnoverChange := 0
	floorPriceChange := 0
	// 如果上一个时间段成交额不为0，计算成交额变化百分比
	if !prevTurnover.IsZero() {
		tcd := turnover.Sub(prevTurnover).Div(prevTurnover).Mul(decimal.NewFromInt(100))
		turnoverChange = int(tcd.IntPart())
	}
	// 如果上一时段地板价不为0,计算地板价变化百分比
	if !prevFloorPrice.IsZero() {
		floorChangeDecimal := floorPrice.Sub(prevFloorPrice).Div(prevFloorPrice).Mul(decimal.NewFromInt(100))
		floorPriceChange = int(floorChangeDecimal.IntPart())
	}
	// 返回集合交易统计信息
	return &CollectionTrade{
		ContractAddress: addr,
		Volume:          volume,
		Turnover:        turnover,
		TurnoverChange:  turnoverChange,
		PreFloorPrice:   prevFloorPrice,
		FloorChange:     floorPriceChange,
	}, nil

}

// GetCollectionTurnover 获取指定collection的交易总额
func (d *Dao) GetCollectionTurnover(chain string, addr string) (decimal.Decimal, error) {
	var turnover decimal.Decimal
	err := d.DB.WithContext(d.ctx).Table(model.ActivityTableName(chain)).
		Where("collection_address = ? AND activity_type = ?", addr, model.Sale).
		Select("COALESCE(SUM(price), 0)").
		Row().Scan(&turnover)
	if err != nil {
		return decimal.Zero, errcode.NewCustomErr("failed to get collection total turnover")
	}

	return turnover, nil
}

// GetCollectionRankingByActivity 根据Activity获取集合排行榜信息
func (d *Dao) GetCollectionRankingByActivity(chain, period string) ([]*CollectionTrade, error) {
	// 解析时间范围
	// 获取时间段对应的epoch值
	epoch, ok := periodToDuration[period]
	if !ok {
		return nil, errors.Errorf("invalid period: %s", period)
	}
	// 计算查询的时间范围
	startTime := time.Now().Add(-time.Duration(epoch) * time.Minute)
	endTime := time.Now()

	// 计算上一个时间段
	prevEndTime := startTime
	prevStartTime := startTime.Add(-time.Duration(epoch) * time.Minute)

	// 获取当前时间段的交易统计
	type TradeStats struct {
		CollectionAddress string
		Volume            int64
		Turnover          decimal.Decimal
		FloorPrice        decimal.Decimal
	}

	var currentStats []TradeStats
	err := d.DB.WithContext(d.ctx).Table(model.ActivityTableName(chain)).
		Select("collection_address, COUNT(*) as item_count, COALESCE(SUM(price), 0) as volume, COALESCE(MIN(price), 0) as floor_price").
		Where("activity_type = ? AND event_time >= ? AND event_time <= ?", model.Sale, startTime, endTime).
		Group("collection_address").
		Find(&currentStats).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current stats")
	}

	// 获取上一时间段的交易统计
	var prevStats []TradeStats
	err = d.DB.WithContext(d.ctx).Table(model.ActivityTableName(chain)).
		Select("collection_address, COUNT(*) as item_count, COALESCE(SUM(price), 0) as volume, COALESCE(MIN(price), 0) as floor_price").
		Where("activity_type = ? AND event_time >= ? AND event_time <= ?", model.Sale, prevStartTime, prevEndTime).
		Group("collection_address").
		Find(&prevStats).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to get previous stats")
	}

	// 构建上一时间段数据的map
	prevStatsMap := make(map[string]TradeStats)
	for _, stat := range prevStats {
		prevStatsMap[stat.CollectionAddress] = stat
	}

	// 构建结果
	var result []*CollectionTrade
	for _, curr := range currentStats {
		trade := &CollectionTrade{
			ContractAddress: curr.CollectionAddress,
			Volume:          curr.Volume,
			Turnover:        curr.Turnover,
			TurnoverChange:  0,
			PreFloorPrice:   decimal.Zero,
			FloorChange:     0,
		}

		// 计算变化率
		if prev, ok := prevStatsMap[curr.CollectionAddress]; ok {
			trade.PreFloorPrice = prev.FloorPrice

			if !prev.Turnover.IsZero() {
				volumeChangeDecimal := curr.Turnover.Sub(prev.Turnover).Div(prev.Turnover).Mul(decimal.NewFromInt(100))
				trade.TurnoverChange = int(volumeChangeDecimal.IntPart())
			}

			if !prev.FloorPrice.IsZero() {
				floorChangeDecimal := curr.FloorPrice.Sub(prev.FloorPrice).Div(prev.FloorPrice).Mul(decimal.NewFromInt(100))
				trade.FloorChange = int(floorChangeDecimal.IntPart())
			}
		}

		result = append(result, trade)
	}

	return result, nil
}

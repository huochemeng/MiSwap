package dao

import (
	"MiSwap/base/pkg/errcode"
	"MiSwap/base/stores/gdb/model"
	"context"
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

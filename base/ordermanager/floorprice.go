package ordermanager

import (
	"MiSwap/base/stores/xkv"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

type EventType int

const (
	Buy EventType = iota + 1
	Mint
	Listing
	Cancel
	Transfer         EventType = 8
	Expired          EventType = 9
	ImportCollection EventType = 10
	UpdateCollection EventType = 11
)

const (
	MaxBatchReqNum = 100
	maxQueueLength = 100
)

const CacheTradeEventsQueuePre = "cache:trade:events:%s"

type TradeEvent struct {
	EventType      EventType       `json:"event_type"`
	CollectionAddr string          `json:"collection_addr"`
	TokenID        string          `json:"token_id"`
	OrderId        string          `json:"order_id"`
	OrderHash      string          `json:"order_hash"`
	Price          decimal.Decimal `json:"price"`
	From           string          `json:"from"`
	To             string          `json:"to"`
	TxHash         string          `json:"txHash"`
}

func genTradeEventsCacheKey(chain string) string {
	return fmt.Sprintf(CacheTradeEventsQueuePre, chain)
}

// AddUpdatePriceEvent 函数用于添加更新价格的事件
// 主要功能:
// 1. 验证事件的合法性:
//   - 对于非Transfer/ImportCollection/UpdateCollection事件,必须有OrderId
//   - 对于Listing事件,价格不能为0且必须有TokenID
//
// 2. 将事件序列化并添加到Redis队列中
// 参数说明:
// - kv: Redis存储实例
// - event: 交易事件,包含事件类型、订单ID、价格等信息
// - chain: 链标识
// 返回值:
// - error: 错误信息
func AddUpdatePriceEvent(kv *xkv.Store, event *TradeEvent, chain string) error {
	// 验证事件合法性
	// 如果是非Transfer/ImportCollection/UpdateCollection事件,必须有OrderId
	if event.EventType != Transfer && event.EventType != ImportCollection && event.EventType != UpdateCollection && event.OrderId == "" {
		return errors.New("invalid update collection floor price. event order id is null")
	}
	// 如果是Listing事件,价格不能为0
	if event.EventType == Listing && event.Price.IsZero() {
		return errors.New("invalid update collection floor price. price is 0")
	}
	// 如果是Listing事件,必须有TokenID
	if event.EventType == Listing && event.TokenID == "" {
		return errors.New("invalid update collection floor price. token_id is null")
	}

	// 序列化事件
	rawEvent, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed on marshal event")
	}

	// 获取Redis队列key
	key := genTradeEventsCacheKey(chain)
	// 将事件添加到Redis队列
	if _, err := kv.Rpush(key, string(rawEvent)); err != nil {
		return errors.Wrap(err, "failed on push trade event to queue")
	}
	return nil
}

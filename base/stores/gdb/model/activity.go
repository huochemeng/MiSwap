package model

import (
	"fmt"
	"github.com/shopspring/decimal"
)

// 活动类型
/*
| 常量 | 值 | 含义 | 说明 |
| :--- | :--- | :--- | :--- |
| `Buy` | 1 | 购买 | 用户直接以固定价格买入 NFT |
| `Mint` | 2 | 铸造 | 首次从合约中铸造生成 NFT |
| `Listing` | 3 | 上架 | 卖家将 NFT 挂出固定售价 |
| `CancelListing` | 4 | 取消上架 | 卖家撤回固定价卖单 |
| `CancelOffer` | 5 | 取消报价 | 买家撤回针对单个 NFT 的出价 |
| `MakeOffer` | 6 | 发起报价 | 买家对单个 NFT 提交出价 |
| `Sale` | 7 | 成交 | 订单被撮合完成（买卖达成） |
| `Transfer` | 8 | 转移 | NFT 所有权变更（非交易，如空投、钱包间转账） |
| `CollectionBid` | 9 | 集合报价 | 买家对整个 Collection 下任意 NFT 提交的批量出价 |
| `ItemBid` | 10 | 单品报价 | 等同于 MakeOffer，语义上更强调针对特定 TokenId |
| `CancelCollectionBid` | 16 | 取消集合报价 | 撤回整个 Collection 级别的出价 |
| `CancelItemBid` | 17 | 取消单品报价 | 撤回针对特定 TokenId 的出价 |
编号说明：值从 10 直接跳到 16，中间 11-15 是预留位。这通常是因为 Bid 相关功能是后期迭代加入的，为避免与已有枚举冲突而选择了新的号段。
切勿使用 iota 自动重排，否则会导致历史数据解析错误

优化建议：两组常量分别定义自定义类型，防止传参混淆。如： type ActivityType int    Buy ActivityType = 1; type MarketplaceID int  Hub MarketplaceID = iota
*/
const (
	Buy                 = 1
	Mint                = 2
	Listing             = 3
	CancelListing       = 4
	CancelOffer         = 5
	MakeOffer           = 6
	Sale                = 7
	Transfer            = 8
	CollectionBid       = 9
	ItemBid             = 10
	CancelCollectionBid = 16
	CancelItemBid       = 17
)

//市场来源
/*
| 常量 | 值 | 含义 |
| :--- | :--- | :--- |
| `Hub` | 0 | 自有平台（本项目自身的交易市场） |
| `Opensea` | 1 | OpenSea 外部市场同步数据 |
| `Looksrare` | 2 | LooksRare 外部市场同步数据 |
| `X2Y2` | 3 | X2Y2 外部市场同步数据 |
| `Blur` | 4 | Blur 外部市场同步数据 |
| `OrderBookDex` | 5 | 链上订单簿 DEX（如 Sudoswap 等） |


*/
const (
	Hub int = iota
	Opensea
	Looksrare
	X2Y2
	Blur
	OrderBookDex
)

type Activity struct {
	Id int64 `json:"id" gorm:"primaryKey;autoIncrement;column:id;not null"`
	//1:Buy,2:Mint,3:List,4:Cancel List,5:Cancel Offer,6.Make Offer,7.Sell,8.Transfer.
	ActivityType      int             `json:"activity_type" gorm:"column:activity_type;type:tinyint(1);not null"`
	Maker             string          `json:"maker" gorm:"column:maker;type:varchar(42);not null"`
	Taker             string          `json:"taker" gorm:"column:taker;type:varchar(42);not null"`
	MarketplaceID     int             `json:"marketplace_id" gorm:"column:marketplace_id;type:tinyint;not null;default:0"`
	CollectionAddress string          `json:"collection_address" gorm:"column:collection_address;type:varchar(64);not null;default:''"`
	TokenId           string          `json:"token_id" gorm:"column:token_id"`
	CurrencyAddress   string          `gorm:"column:currency_address" json:"currency_address"`
	Price             decimal.Decimal `gorm:"column:price" json:"price"`
	SellPrice         decimal.Decimal `json:"sell_price" gorm:"column:sell_price;type:decimal(30);not null;default:0"`
	BuyPrice          decimal.Decimal `json:"buy_price" gorm:"column:buy_price;type:decimal(30);not null;default:0"`
	BlockNumber       int64           `json:"block_number" gorm:"column:block_number;type:bigint(20);not null"`
	TxHash            string          `json:"tx_hash" gorm:"column:tx_hash;type:varchar(255);not null"`
	EventTime         int64           `json:"event_time" gorm:"column:event_time;type:bigint(20);default:0;comment:链上事件发生的时间"`
	CreateTime        int64           `json:"create_time" gorm:"column:create_time;type:bigint(20);autoCreateTime:milli;comment:创建时间"` // 创建时间
	UpdateTime        int64           `json:"update_time" gorm:"column:update_time;type:bigint(20);autoUpdateTime:milli;comment:更新时间"` // 更新时间
}

func ActivityTableName(chainName string) string {
	return fmt.Sprintf("activity_%s", chainName)
}

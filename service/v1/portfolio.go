package service

const BidTypeOffset = 3

// 向后兼容的协议升级。旧版只有0,1,2.新版本增加了新类型，为了不破坏已有数据的解析，遇到新的值读取时减3还原
func getBidType(origin int64) int64 {
	if origin >= BidTypeOffset {
		return origin - BidTypeOffset
	} else {
		return origin
	}
}

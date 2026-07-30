package svc

import (
	"MiSwap/base/stores/xkv"
	"MiSwap/dao"
	"gorm.io/gorm"
)

type CtxConfig struct {
	db *gorm.DB
	//	dao
	dao *dao.Dao
	//	redis
	KvStore *xkv.Store
}

// NewServerCtx 先使用结构体的方式传参。 TODO 可优化成函数选项模式
func NewServerCtx(cfg CtxConfig) *ServerCtx {
	return &ServerCtx{
		DB:      cfg.db,
		Dao:     cfg.dao,
		KvStore: cfg.KvStore,
	}
}

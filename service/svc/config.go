package svc

import "gorm.io/gorm"

type CtxConfig struct {
	db *gorm.DB
	//	dao
	//	redis
}

// NewServerCtx 先使用结构体的方式传参。 TODO 可优化成函数选项模式
func NewServerCtx(cfg CtxConfig) *ServerCtx {
	return &ServerCtx{
		DB: cfg.db,
	}
}

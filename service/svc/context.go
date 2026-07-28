package svc

import (
	"MiSwap/base/stores/gdb"
	"MiSwap/config"
	"gorm.io/gorm"
)

type ServerCtx struct {
	C  *config.Config
	DB *gorm.DB
	//	todo redis,dao等
}

func NewServiceContext(c *config.Config) (*ServerCtx, error) {
	//todo 初始化log
	//	db
	db, err := gdb.NewDB(&c.DB)
	if err != nil {
		return nil, err
	}
	_ = db
	//todo 初始化dao
	//todo 初始化redis

	//所有环境对象添加到serverCtx中（使用函数选项模式，将这些依赖以可选参数的方式添加）

	return nil, err
}

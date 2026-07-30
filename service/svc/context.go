package svc

import (
	"MiSwap/base/stores/gdb"
	"MiSwap/base/stores/xkv"
	"MiSwap/config"
	"MiSwap/dao"
	"context"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
)

type ServerCtx struct {
	DB *gorm.DB
	//	todo redis,dao等
	Dao     *dao.Dao
	KvStore *xkv.Store
}

func NewServiceContext(c *config.Config) (*ServerCtx, error) {
	//todo 初始化log

	//	db
	db, err := gdb.NewDB(&c.DB)
	if err != nil {
		return nil, err
	}
	_ = db

	// 解析配置文件中redis的相关信息，添加到kvconf
	var kvConf kv.KvConf
	for _, con := range c.Kv.Redis {
		kvConf = append(kvConf, cache.NodeConf{
			RedisConf: redis.RedisConf{
				Host: con.Host,
				Pass: con.Pass,
				Type: con.Type,
			},
			Weight: 1,
		})
	}
	//获取redis对象
	store := xkv.NewStore(kvConf)

	d := dao.New(context.Background(), db, store)
	//服务上下文所需的环境
	serverCtx := NewServerCtx(CtxConfig{
		db:      db,
		dao:     d,
		KvStore: store,
	})

	//TODO 所有环境对象添加到serverCtx中（使用函数选项模式，将这些依赖以可选参数的方式添加，后续可优化的点,目前只是抽离成独立的结构体，并未使用函数选项模式）

	return serverCtx, err
}

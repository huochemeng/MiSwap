package dao

import (
	"MiSwap/base/stores/xkv"
	"context"
	"gorm.io/gorm"
)

type Dao struct {
	ctx     context.Context
	DB      *gorm.DB
	KvStore *xkv.Store
}

func New(ctx context.Context, db *gorm.DB, kvStore *xkv.Store) *Dao {
	return &Dao{
		ctx:     ctx,
		DB:      db,
		KvStore: kvStore,
	}
}

package xkv

import (
	"MiSwap/base/kit/convert"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"log"
)

type Store struct {
	kv.Store
	Redis *redis.Redis
}

// NewStore 新建键值存储
func NewStore(c kv.KvConf) *Store {
	if len(c) == 0 || cache.TotalWeights(c) <= 0 {
		log.Fatal("no cache nodes")
	}
	rds := redis.MustNewRedis(c[0].RedisConf)
	return &Store{
		Store: kv.NewStore(c),
		Redis: rds,
	}
}

func (s *Store) GetInt(key string) (int, error) {
	value, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	return convert.ToInt(value), nil
}

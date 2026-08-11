package xkv

import (
	"MiSwap/base/kit/convert"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/kv"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"log"
	"reflect"
)

const (
	// getAndDelScript 获取并删除key所关联的值lua脚本
	getAndDelScript = `local current = redis.call('GET', KEYS[1]);
if (current) then
    redis.call('DEL', KEYS[1]);
end
return current;`
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

// SetString 将 string value 关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) SetString(key, value string, seconds ...int) error {
	if len(seconds) > 0 {
		if seconds[0] <= 0 {
			return errors.New("setex ttl must be positive")
		}
		return errors.Wrapf(s.Setex(key, value, seconds[0]), "setex by seconds=%d err", seconds[0])
	}
	return errors.Wrap(s.Set(key, value), "set err")
}

// SetInt 将 int value 关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) SetInt(key string, value int, seconds ...int) error {
	return s.SetString(key, convert.ToString(value), seconds...)
}

// GetInt 返回给定 key 所关联的 int 值
func (s *Store) GetInt(key string) (int, error) {
	value, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	return convert.ToInt(value), nil
}

// SetInt64 将 int64 value 关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) SetInt64(key string, value int64, seconds ...int) error {
	return s.SetString(key, convert.ToString(value), seconds...)
}

// GetInt64 返回给定 key 所关联的 int64 值
func (s *Store) GetInt64(key string) (int64, error) {
	value, err := s.Get(key)
	if err != nil {
		return 0, err
	}
	return convert.ToInt64(value), nil
}

// GetBytes 返回给定 key 所关联的 []byte 值
func (s *Store) GetBytes(key string) ([]byte, error) {
	value, err := s.Get(key)
	if err != nil {
		return nil, err
	}
	return []byte(value), nil
}

// GetDel 原子地返回并删除给定 key 所关联的 string 值
func (s *Store) GetDel(key string) (string, error) {
	resp, err := s.Eval(getAndDelScript, key)
	if err != nil {
		return "", errors.Wrap(err, "eval getAndDel script err")
	}
	return convert.ToString(resp), nil
}

// Read 将给定 key 所关联的值反序列化到 obj 对象
// 返回 false 时代表给定 key 不存在
func (s *Store) Read(key string, obj interface{}) (bool, error) {
	if !isValidPtr(obj) {
		return false, errors.New("obj must be a non-nil pointer")
	}

	value, err := s.GetBytes(key)
	if err != nil {
		return false, errors.Wrap(err, "get bytes err")
	}
	if len(value) == 0 {
		return false, nil
	}

	if err = json.Unmarshal(value, obj); err != nil {
		return false, errors.Wrap(err, "json unmarshal value to obj err")
	}
	return true, nil
}

// Write 将对象 obj 序列化后关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) Write(key string, obj interface{}, seconds ...int) error {
	value, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrap(err, "json marshal obj err")
	}
	return s.SetString(key, string(value), seconds...)
}

// GetFunc 给定 key 不存在时调用的数据获取函数
type GetFunc func() (interface{}, error)

// ReadOrGet 将给定 key 所关联的值反序列化到 obj 对象
// 若给定 key 不存在则调用数据获取函数，调用成功时赋值至 obj 对象
// 并将其序列化后关联到给定 key，seconds 为 key 的过期时间（秒）
func (s *Store) ReadOrGet(key string, obj interface{}, gf GetFunc, seconds ...int) error {
	isExist, err := s.Read(key, obj)
	if err != nil {
		return errors.Wrap(err, "read obj err")
	}
	if isExist {
		return nil
	}

	data, err := gf()
	if err != nil {
		return errors.Wrap(err, "get data from source err")
	}
	if !isValidPtr(data) {
		return errors.New("get data must be a non-nil pointer")
	}

	ov := reflect.ValueOf(obj).Elem()
	dv := reflect.ValueOf(data).Elem()
	if !dv.Type().AssignableTo(ov.Type()) {
		return errors.Errorf("obj type %s and get data type %s are not assignable", ov.Type(), dv.Type())
	}
	ov.Set(dv)

	// 回填缓存失败不应阻塞主流程，仅记录日志（此处用 _ 忽略，实际项目建议 log.Warn）
	_ = s.Write(key, data, seconds...)
	return nil
}

// isValidPtr 判断对象是否为合法的非空指针
func isValidPtr(obj interface{}) bool {
	if obj == nil {
		return false
	}
	return reflect.ValueOf(obj).Kind() == reflect.Ptr
}

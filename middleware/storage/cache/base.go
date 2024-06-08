package cache

import (
	"github.com/lazygophers/utils/anyx"
	"github.com/lazygophers/utils/json"
	"time"
)

type baseCache struct {
	BaseCache
}

func (p *baseCache) GetBool(key string) (bool, error) {
	buf, err := p.Get(key)
	if err != nil {
		return false, err
	}

	return anyx.ToBool(buf), nil
}

func (p *baseCache) GetInt(key string) (int, error) {
	buf, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	return anyx.ToInt(buf), nil
}

func (p *baseCache) GetUint(key string) (uint, error) {
	buf, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	return anyx.ToUint(buf), nil
}

func (p *baseCache) GetInt32(key string) (int32, error) {
	buf, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	return anyx.ToInt32(buf), nil
}

func (p *baseCache) GetUint32(key string) (uint32, error) {
	buf, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	return anyx.ToUint32(buf), nil
}

func (p *baseCache) GetInt64(key string) (int64, error) {
	buf, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	return anyx.ToInt64(buf), nil
}

func (p *baseCache) GetUint64(key string) (uint64, error) {
	buf, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	return anyx.ToUint64(buf), nil
}

func (p *baseCache) GetFloat32(key string) (float32, error) {
	buf, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	return anyx.ToFloat32(buf), nil
}

func (p *baseCache) GetFloat64(key string) (float64, error) {
	buf, err := p.Get(key)
	if err != nil {
		return 0, err
	}

	return anyx.ToFloat64(buf), nil
}

func (p *baseCache) GetBoolSlice(key string) ([]bool, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		return nil, err
	}

	res := make([]bool, 0, len(list))
	for _, v := range list {
		res = append(res, anyx.ToBool(v))
	}

	return res, nil
}

func (p *baseCache) GetIntSlice(key string) ([]int, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		return nil, err
	}

	res := make([]int, 0, len(list))
	for _, v := range list {
		res = append(res, anyx.ToInt(v))
	}

	return res, nil
}

func (p *baseCache) GetUintSlice(key string) ([]uint, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		return nil, err
	}

	res := make([]uint, 0, len(list))
	for _, v := range list {
		res = append(res, anyx.ToUint(v))
	}

	return res, nil
}

func (p *baseCache) GetInt32Slice(key string) ([]int32, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		return nil, err
	}

	res := make([]int32, 0, len(list))
	for _, v := range list {
		res = append(res, anyx.ToInt32(v))
	}

	return res, nil
}

func (p *baseCache) GetUint32Slice(key string) ([]uint32, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		return nil, err
	}

	res := make([]uint32, 0, len(list))
	for _, v := range list {
		res = append(res, anyx.ToUint32(v))
	}

	return res, nil
}

func (p *baseCache) GetInt64Slice(key string) ([]int64, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		return nil, err
	}

	res := make([]int64, 0, len(list))
	for _, v := range list {
		res = append(res, anyx.ToInt64(v))
	}

	return res, nil
}

func (p *baseCache) GetUint64Slice(key string) ([]uint64, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		return nil, err
	}

	res := make([]uint64, 0, len(list))
	for _, v := range list {
		res = append(res, anyx.ToUint64(v))
	}

	return res, nil
}

func (p *baseCache) GetFloat32Slice(key string) ([]float32, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		return nil, err
	}

	res := make([]float32, 0, len(list))
	for _, v := range list {
		res = append(res, anyx.ToFloat32(v))
	}

	return res, nil
}

func (p *baseCache) GetFloat64Slice(key string) ([]float64, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		return nil, err
	}

	res := make([]float64, 0, len(list))
	for _, v := range list {
		res = append(res, anyx.ToFloat64(v))
	}

	return res, nil
}

func (p *baseCache) Limit(key string, limit int64, timeout time.Duration) (bool, error) {
	cnt, err := p.Incr(key)
	if err != nil {
		return false, err
	}

	if cnt == 1 {
		_, err = p.Expire(key, timeout)
		if err != nil {
			return false, err
		}
	}

	if cnt > limit {
		return false, nil
	}

	return true, nil
}

func (p *baseCache) GetSlice(key string) ([]string, error) {
	buf, err := p.Get(key)
	if err != nil {
		return nil, err
	}

	if buf == "" {
		return nil, nil
	}

	var list []string
	err = json.UnmarshalString(buf, &list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (p *baseCache) GetJson(key string, j interface{}) error {
	value, err := p.Get(key)
	if err != nil {
		return err
	}

	return json.UnmarshalString(value, j)
}

func (p *baseCache) HGetJson(key, field string, j interface{}) error {
	value, err := p.HGet(key, field)
	if err != nil {
		return err
	}

	return json.UnmarshalString(value, j)
}

func newBaseCache(c BaseCache) Cache {
	return &baseCache{
		BaseCache: c,
	}
}

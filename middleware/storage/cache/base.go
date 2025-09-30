package cache

import (
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/candy"
	"github.com/lazygophers/utils/json"
	"github.com/lazygophers/utils/stringx"
	"google.golang.org/protobuf/proto"
)

type baseCache struct {
	BaseCache
}

func (p *baseCache) SetPb(key string, j proto.Message) error {
	buffer, err := proto.Marshal(j)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = p.Set(key, stringx.ToString(buffer))
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *baseCache) SetPbEx(key string, j proto.Message, timeout time.Duration) error {
	buffer, err := proto.Marshal(j)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = p.SetEx(key, stringx.ToString(buffer), timeout)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *baseCache) GetPb(key string, j proto.Message) error {
	buffer, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = proto.Unmarshal(stringx.ToBytes(buffer), j)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *baseCache) GetBool(key string) (bool, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	return candy.ToBool(buf), nil
}

func (p *baseCache) GetInt(key string) (int, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return candy.ToInt(buf), nil
}

func (p *baseCache) GetUint(key string) (uint, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return candy.ToUint(buf), nil
}

func (p *baseCache) GetInt32(key string) (int32, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return candy.ToInt32(buf), nil
}

func (p *baseCache) GetUint32(key string) (uint32, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return candy.ToUint32(buf), nil
}

func (p *baseCache) GetInt64(key string) (int64, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return candy.ToInt64(buf), nil
}

func (p *baseCache) GetUint64(key string) (uint64, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return candy.ToUint64(buf), nil
}

func (p *baseCache) GetFloat32(key string) (float32, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return candy.ToFloat32(buf), nil
}

func (p *baseCache) GetFloat64(key string) (float64, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return candy.ToFloat64(buf), nil
}

func (p *baseCache) GetBoolSlice(key string) ([]bool, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	res := make([]bool, 0, len(list))
	for _, v := range list {
		res = append(res, candy.ToBool(v))
	}

	return res, nil
}

func (p *baseCache) GetIntSlice(key string) ([]int, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	res := make([]int, 0, len(list))
	for _, v := range list {
		res = append(res, candy.ToInt(v))
	}

	return res, nil
}

func (p *baseCache) GetUintSlice(key string) ([]uint, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	res := make([]uint, 0, len(list))
	for _, v := range list {
		res = append(res, candy.ToUint(v))
	}

	return res, nil
}

func (p *baseCache) GetInt32Slice(key string) ([]int32, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	res := make([]int32, 0, len(list))
	for _, v := range list {
		res = append(res, candy.ToInt32(v))
	}

	return res, nil
}

func (p *baseCache) GetUint32Slice(key string) ([]uint32, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	res := make([]uint32, 0, len(list))
	for _, v := range list {
		res = append(res, candy.ToUint32(v))
	}

	return res, nil
}

func (p *baseCache) GetInt64Slice(key string) ([]int64, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	res := make([]int64, 0, len(list))
	for _, v := range list {
		res = append(res, candy.ToInt64(v))
	}

	return res, nil
}

func (p *baseCache) GetUint64Slice(key string) ([]uint64, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	res := make([]uint64, 0, len(list))
	for _, v := range list {
		res = append(res, candy.ToUint64(v))
	}

	return res, nil
}

func (p *baseCache) GetFloat32Slice(key string) ([]float32, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	res := make([]float32, 0, len(list))
	for _, v := range list {
		res = append(res, candy.ToFloat32(v))
	}

	return res, nil
}

func (p *baseCache) GetFloat64Slice(key string) ([]float64, error) {
	var list []interface{}
	err := p.GetJson(key, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	res := make([]float64, 0, len(list))
	for _, v := range list {
		res = append(res, candy.ToFloat64(v))
	}

	return res, nil
}

func (p *baseCache) Limit(key string, limit int64, timeout time.Duration) (bool, error) {
	cnt, err := p.Incr(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	if cnt == 1 {
		_, err = p.Expire(key, timeout)
		if err != nil {
			log.Errorf("err:%v", err)
			return false, err
		}
	}

	if cnt > limit {
		return false, nil
	}

	return true, nil
}

func (p *baseCache) LimitUpdateOnCheck(key string, limit int64, timeout time.Duration) (bool, error) {
	cnt, err := p.Incr(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	_, err = p.Expire(key, timeout)
	if err != nil {
		log.Errorf("err:%v", err)
		return false, err
	}

	if cnt > limit {
		return false, nil
	}

	return true, nil
}

func (p *baseCache) GetSlice(key string) ([]string, error) {
	buf, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	if buf == "" {
		return nil, nil
	}

	var list []string
	err = json.UnmarshalString(buf, &list)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return list, nil
}

func (p *baseCache) GetJson(key string, j interface{}) error {
	value, err := p.Get(key)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = json.UnmarshalString(value, j)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *baseCache) HGetJson(key, field string, j interface{}) error {
	value, err := p.HGet(key, field)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = json.UnmarshalString(value, j)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func newBaseCache(c BaseCache) Cache {
	return &baseCache{
		BaseCache: c,
	}
}

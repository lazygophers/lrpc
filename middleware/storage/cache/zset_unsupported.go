package cache

import "errors"

var ErrZSetNotSupported = errors.New("zset operations not supported for this cache type")

// CacheBbolt ZSet methods (unsupported)

func (p *CacheBbolt) ZAdd(key string, members ...interface{}) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheBbolt) ZScore(key, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRangeByScore(key, min, max string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRem(key string, members ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheBbolt) ZCard(key string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheBbolt) ZCount(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheBbolt) ZIncrBy(key string, increment float64, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRevRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRevRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRangeWithScores(key string, start, stop int64) ([]Z, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRevRangeByScore(key, max, min string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheBbolt) ZRemRangeByScore(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheBbolt) ZUnionStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheBbolt) ZInterStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

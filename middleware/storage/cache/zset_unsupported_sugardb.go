//go:build sugardb

package cache

// CacheSugarDB ZSet methods (unsupported)
func (p *CacheSugarDB) ZAdd(key string, members ...interface{}) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZScore(key, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRangeByScore(key, min, max string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRem(key string, members ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZCard(key string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZCount(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZIncrBy(key string, increment float64, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRevRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRevRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRangeWithScores(key string, start, stop int64) ([]Z, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRevRangeByScore(key, max, min string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZRemRangeByScore(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZUnionStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheSugarDB) ZInterStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

//go:build database

package cache

// Database ZSet methods (unsupported)
func (p *CacheDatabase) ZAdd(key string, members ...interface{}) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheDatabase) ZScore(key, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRangeByScore(key, min, max string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRem(key string, members ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheDatabase) ZCard(key string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheDatabase) ZCount(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheDatabase) ZIncrBy(key string, increment float64, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRevRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRevRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRangeWithScores(key string, start, stop int64) ([]Z, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRevRangeByScore(key, max, min string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheDatabase) ZRemRangeByScore(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheDatabase) ZUnionStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheDatabase) ZInterStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

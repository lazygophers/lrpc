//go:build leveldb

package cache

// CacheLevelDB ZSet methods (unsupported)
func (p *CacheLevelDB) ZAdd(key string, members ...interface{}) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZScore(key, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRangeByScore(key, min, max string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRem(key string, members ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZCard(key string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZCount(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZIncrBy(key string, increment float64, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRevRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRevRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRangeWithScores(key string, start, stop int64) ([]Z, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRevRangeByScore(key, max, min string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZRemRangeByScore(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZUnionStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *CacheLevelDB) ZInterStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

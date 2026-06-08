//go:build database

package cache

// Database ZSet methods (unsupported)
func (p *Database) ZAdd(key string, members ...interface{}) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *Database) ZScore(key, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *Database) ZRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *Database) ZRangeByScore(key, min, max string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *Database) ZRem(key string, members ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *Database) ZCard(key string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *Database) ZCount(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *Database) ZIncrBy(key string, increment float64, member string) (float64, error) {
	return 0, ErrZSetNotSupported
}

func (p *Database) ZRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *Database) ZRevRange(key string, start, stop int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *Database) ZRevRank(key, member string) (int64, error) {
	return -1, ErrZSetNotSupported
}

func (p *Database) ZRangeWithScores(key string, start, stop int64) ([]Z, error) {
	return nil, ErrZSetNotSupported
}

func (p *Database) ZRevRangeByScore(key, max, min string, offset, count int64) ([]string, error) {
	return nil, ErrZSetNotSupported
}

func (p *Database) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *Database) ZRemRangeByScore(key, min, max string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *Database) ZUnionStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

func (p *Database) ZInterStore(destination string, keys ...string) (int64, error) {
	return 0, ErrZSetNotSupported
}

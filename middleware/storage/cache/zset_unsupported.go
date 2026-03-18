package cache

import "errors"

var ErrZSetNotSupported = errors.New("zset operations not supported for this cache type")

// ZSet方法的不支持实现（用于bbolt、database、echo、leveldb）

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

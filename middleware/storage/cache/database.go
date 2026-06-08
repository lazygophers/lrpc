//go:build database

package cache

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils"
	"github.com/lazygophers/utils/app"
	"github.com/lazygophers/utils/candy"
	_ "modernc.org/sqlite"
)

type CacheDatabase struct {
	prefix    string
	tableName string

	db *sql.DB
}

func (p *CacheDatabase) Clean() error {
	_, err := p.db.Exec(fmt.Sprintf("delete from %s where e > 0 and e < ?", p.tableName), time.Now().Unix())
	if err != nil {
		return p.coverError(err)
	}
	return nil
}

func (p *CacheDatabase) coverError(err error) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return ErrNotFound
	}

	return err
}

func (p *CacheDatabase) Get(key string) (string, error) {
	var value string
	err := p.db.QueryRow(fmt.Sprintf("select v from %s where k = ? and (e = 0 or e > ?)", p.tableName), key, time.Now().Unix()).Scan(&value)
	if err != nil {
		return "", p.coverError(err)
	}

	return value, nil
}

func (p *CacheDatabase) Set(key string, value any) error {
	_, err := p.db.Exec(fmt.Sprintf("insert or replace into %s (k, v, e) values (?,?,?)", p.tableName), key, candy.ToString(value), 0)
	if err != nil {
		return p.coverError(err)
	}

	return nil
}

func (p *CacheDatabase) SetEx(key string, value any, timeout time.Duration) error {
	_, err := p.db.Exec(fmt.Sprintf("insert or replace into %s (k, v, e) values (?,?,?)", p.tableName), key, candy.ToString(value), time.Now().Add(timeout).Unix())
	if err != nil {
		return p.coverError(err)
	}
	return nil
}

func (p *CacheDatabase) SetNx(key string, value interface{}) (bool, error) {
	_, err := p.db.Exec(fmt.Sprintf("insert into %s (k, v, e) values (?,?,?)", p.tableName), key, candy.ToString(value), 0)
	if err != nil {
		if strings.Contains(err.Error(), "(1555)") && strings.Contains(err.Error(), "UNIQUE") {
			return false, nil
		}

		if strings.Contains(err.Error(), "Duplicate") {
			return false, nil
		}

		return false, p.coverError(err)
	}
	return true, nil
}

func (p *CacheDatabase) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
	ok, err := p.SetNx(key, value)
	if err != nil {
		return ok, err
	}
	if !ok {
		return ok, nil
	}

	_, err = p.Expire(key, timeout)
	if err != nil {
		return ok, err
	}

	return ok, nil
}

func (p *CacheDatabase) Ttl(key string) (time.Duration, error) {
	var value int64
	err := p.db.QueryRow(fmt.Sprintf("select e from %s where k = ? and (e = 0 or e > ?)", p.tableName), key, time.Now().Unix()).Scan(&value)
	if err != nil {
		return -1, p.coverError(err)
	}

	return time.Unix(value, 0).Sub(time.Now()), nil
}

func (p *CacheDatabase) Expire(key string, timeout time.Duration) (bool, error) {
	res, err := p.db.Exec(fmt.Sprintf("update %s set e = ? where k = ?", p.tableName), time.Now().Add(timeout).Unix(), key)
	if err != nil {
		return false, err
	}
	return utils.Ignore(res.RowsAffected()) > 0, nil
}

func (p *CacheDatabase) Incr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) Decr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) IncrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) DecrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) Exists(keys ...string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HSet(key string, field string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HGet(key, field string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HDel(key string, fields ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HKeys(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HGetAll(key string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HExists(key string, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HIncr(key string, subKey string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HIncrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HDecr(key string, field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) HDecrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) SAdd(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) SMembers(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) SRem(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) SRandMember(key string, count ...int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) SPop(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) SisMember(key, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *CacheDatabase) Del(key ...string) error {
	if len(key) == 0 {
		return nil
	}

	_, err := p.db.Exec(fmt.Sprintf("delete from %s where k in (?)", p.tableName), key)
	if err != nil {
		return p.coverError(err)
	}
	return nil
}

func (p *CacheDatabase) Close() error {
	return p.db.Close()
}

func (p *CacheDatabase) Client() any {
	return p.db
}

func (p *CacheDatabase) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *CacheDatabase) Ping() error {
	err := p.db.Ping()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *CacheDatabase) Publish(channel string, message interface{}) (int64, error) {
	return 0, errors.New("database cache does not support pub/sub")
}

func (p *CacheDatabase) Subscribe(handler func(channel string, message []byte) error, channels ...string) error {
	return errors.New("database cache does not support pub/sub")
}

func (p *CacheDatabase) XAdd(stream string, values map[string]interface{}) (string, error) {
	return "", errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XLen(stream string) (int64, error) {
	return 0, errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	return nil, errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XRevRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	return nil, errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XDel(stream string, ids ...string) (int64, error) {
	return 0, errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XTrim(stream string, maxLen int64) (int64, error) {
	return 0, errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XGroupCreate(stream, group, start string) error {
	return errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XGroupDestroy(stream, group string) error {
	return errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XGroupSetID(stream, group, id string) error {
	return errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XReadGroup(handler func(stream string, id string, body []byte) error, group, consumer, stream string) error {
	return errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XAck(stream, group string, ids ...string) (int64, error) {
	return 0, errors.New("database cache does not support stream")
}

func (p *CacheDatabase) XPending(stream, group string) (int64, error) {
	return 0, errors.New("database cache does not support stream")
}

func NewDatabase(db *sql.DB, tableName string) (Cache, error) {
	p := &CacheDatabase{
		db:        db,
		tableName: tableName,
	}

	_, err := db.Exec(fmt.Sprintf("create table if not exists %s(k varchar(255) primary key,v blob,e bigint default 0)", tableName))
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return newBaseCache(p), nil
}

func newDatabaseFromConfig(c *Config) (Cache, error) {
	tableName := app.Name + "_cache"
	db, err := sql.Open("sqlite", c.Address)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return NewDatabase(db, tableName)
}

func init() {
	RegisterBuilder(Database, func(c *Config) (Cache, error) {
		return newDatabaseFromConfig(c)
	})
}

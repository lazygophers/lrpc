package cache

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils"
	"github.com/lazygophers/utils/candy"
)

type Database struct {
	prefix    string
	tableName string

	db *sql.DB
}

func (p *Database) Clean() error {
	_, err := p.db.Exec(fmt.Sprintf("delete from %s where e > 0 and e < ?", p.tableName), time.Now().Unix())
	if err != nil {
		return p.coverError(err)
	}
	return nil
}

func (p *Database) coverError(err error) error {
	if err == nil {
		return nil
	}

	if err == sql.ErrNoRows {
		return ErrNotFound
	}

	return err
}

func (p *Database) Get(key string) (string, error) {
	var value string
	err := p.db.QueryRow(fmt.Sprintf("select v from %s where k = ? and (e = 0 or e > ?)", p.tableName), key, time.Now().Unix()).Scan(&value)
	if err != nil {
		return "", p.coverError(err)
	}

	return value, nil
}

func (p *Database) Set(key string, value any) error {
	_, err := p.db.Exec(fmt.Sprintf("insert or replace into %s (k, v, e) values (?,?,?)", p.tableName), key, candy.ToString(value), 0)
	if err != nil {
		return p.coverError(err)
	}

	return nil
}

func (p *Database) SetEx(key string, value any, timeout time.Duration) error {
	_, err := p.db.Exec(fmt.Sprintf("insert or replace into %s (k, v, e) values (?,?,?)", p.tableName), key, candy.ToString(value), time.Now().Add(timeout).Unix())
	if err != nil {
		return p.coverError(err)
	}
	return nil
}

func (p *Database) SetNx(key string, value interface{}) (bool, error) {
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

func (p *Database) SetNxWithTimeout(key string, value interface{}, timeout time.Duration) (bool, error) {
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

func (p *Database) Ttl(key string) (time.Duration, error) {
	var value int64
	err := p.db.QueryRow(fmt.Sprintf("select e from %s where k = ? and (e = 0 or e > ?)", p.tableName), key, time.Now().Unix()).Scan(&value)
	if err != nil {
		return -1, p.coverError(err)
	}

	return time.Unix(value, 0).Sub(time.Now()), nil
}

func (p *Database) Expire(key string, timeout time.Duration) (bool, error) {
	res, err := p.db.Exec(fmt.Sprintf("update %s set e = ? where k = ?", p.tableName), time.Now().Add(timeout).Unix(), key)
	if err != nil {
		return false, err
	}
	return utils.Ignore(res.RowsAffected()) > 0, nil
}

func (p *Database) Incr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) Decr(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) IncrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) DecrBy(key string, value int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) Exists(keys ...string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HSet(key string, field string, value interface{}) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HGet(key, field string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HDel(key string, fields ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HKeys(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HGetAll(key string) (map[string]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HExists(key string, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HIncr(key string, subKey string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HIncrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HDecr(key string, field string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) HDecrBy(key string, field string, increment int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) SAdd(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) SMembers(key string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) SRem(key string, members ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) SRandMember(key string, count ...int64) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) SPop(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) SisMember(key, field string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (p *Database) Del(key ...string) error {
	if len(key) == 0 {
		return nil
	}

	_, err := p.db.Exec(fmt.Sprintf("delete from %s where k in (?)", p.tableName), key)
	if err != nil {
		return p.coverError(err)
	}
	return nil
}

func (p *Database) Close() error {
	return p.db.Close()
}

func (p *Database) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *Database) Ping() error {
	err := p.db.Ping()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

func (p *Database) Publish(channel string, message interface{}) (int64, error) {
	return 0, errors.New("database cache does not support pub/sub")
}

func (p *Database) Subscribe(channels ...string) (chan []byte, chan error, error) {
	return nil, nil, errors.New("database cache does not support pub/sub")
}

func (p *Database) XAdd(stream string, values map[string]interface{}) (string, error) {
	return "", errors.New("database cache does not support stream")
}

func (p *Database) XLen(stream string) (int64, error) {
	return 0, errors.New("database cache does not support stream")
}

func (p *Database) XRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	return nil, errors.New("database cache does not support stream")
}

func (p *Database) XRevRange(stream string, start, stop string, count ...int64) ([]map[string]interface{}, error) {
	return nil, errors.New("database cache does not support stream")
}

func (p *Database) XDel(stream string, ids ...string) (int64, error) {
	return 0, errors.New("database cache does not support stream")
}

func (p *Database) XTrim(stream string, maxLen int64) (int64, error) {
	return 0, errors.New("database cache does not support stream")
}

func NewDatabase(db *sql.DB, tableName string) (Cache, error) {
	p := &Database{
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

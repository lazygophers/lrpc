package clickhouse

import (
	"database/sql"
	"fmt"
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	_ "gorm.io/driver/clickhouse"
	"time"
)

type Client struct {
	db *sql.DB
}

func NewClient(c *db.Config) (*Client, error) {
	p := &Client{}

	if c.Logger == nil {
		c.Logger = db.GetDefaultLogger()
	}

	dsn := "clickhouse://"
	if c.Username != "" && c.Password != "" {
		dsn += c.Username + ":" + c.Password + "@"
	} else if c.Username != "" && c.Password == "" {
		dsn += c.Username + "@"
	} else if c.Username == "" && c.Password != "" {
		dsn += c.Password + "@"
	}

	dsn += fmt.Sprintf("%s:%d", c.Address, c.Port)
	if c.Name != "" {
		dsn += "/" + c.Name
	}
	dsn += "?dial_timeout=10s&read_timeout=300s"

	dbConn, err := sql.Open("clickhouse", dsn)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, nil
	}
	// TODO 其他的sql操作
	dbConn.SetMaxOpenConns(50)
	dbConn.SetMaxIdleConns(50)
	dbConn.SetConnMaxIdleTime(time.Second * 90)
	dbConn.SetConnMaxLifetime(time.Second * 90)

	p.db = dbConn
	return p, nil
}

func (p *Client) Database() *sql.DB {
	return p.db
}

func (p *Client) NewScoop() *Scoop {
	return NewScoop(p.db)
}

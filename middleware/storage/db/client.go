package db

import (
	"database/sql"

	"github.com/lazygophers/log"
	"gorm.io/gorm"
)

type Client struct {
	db *gorm.DB

	clientType string
}

func (p *Client) Database() *gorm.DB {
	if p.db == nil {
		log.Errorf("database connection is nil")
		return nil
	}
	return p.db.Session(&gorm.Session{
		Initialized: true,
	})
}

func (p *Client) SqlDB() (*sql.DB, error) {
	db, err := p.db.DB()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}
	return db, nil
}

func (p *Client) DriverType() string {
	return p.clientType
}

func (p *Client) NewScoop() *Scoop {
	if p.db == nil {
		log.Errorf("database connection is nil")
		return nil
	}
	return NewScoop(p.db, p.clientType)
}

func (p *Client) Ping() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *Client) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = sqlDB.Close()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

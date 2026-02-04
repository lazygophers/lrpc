package db

import (
	"database/sql"

	"github.com/lazygophers/log"
	"gorm.io/gorm"
)

type Client struct {
	db         *gorm.DB
	mockDB     *MockDB
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

// MockDB returns the underlying MockDB for mock configuration
// This provides direct access to the MockDB for advanced mock configurations
func (p *Client) MockDB() *MockDB {
	return p.mockDB
}

// ExpectQuery sets up an expectation for a query operation
// Returns the ExpectedQuery for further configuration (WillReturnRows, etc.)
func (p *Client) ExpectQuery(sql string) *ExpectedQuery {
	if p.mockDB == nil {
		log.Errorf("mockDB is nil, cannot set expectation")
		return nil
	}
	return &ExpectedQuery{expected: p.mockDB.Mock.ExpectQuery(sql)}
}

// ExpectExec sets up an expectation for an exec operation
// Returns the ExpectedExec for further configuration (WillReturnResult, etc.)
func (p *Client) ExpectExec(sql string) *ExpectedExec {
	if p.mockDB == nil {
		log.Errorf("mockDB is nil, cannot set expectation")
		return nil
	}
	return &ExpectedExec{expected: p.mockDB.Mock.ExpectExec(sql)}
}

// ExpectBegin sets up an expectation for a transaction begin operation
func (p *Client) ExpectBegin() *ExpectedBegin {
	if p.mockDB == nil {
		log.Errorf("mockDB is nil, cannot set expectation")
		return nil
	}
	return &ExpectedBegin{expected: p.mockDB.Mock.ExpectBegin()}
}

// ExpectCommit sets up an expectation for a transaction commit operation
func (p *Client) ExpectCommit() *ExpectedCommit {
	if p.mockDB == nil {
		log.Errorf("mockDB is nil, cannot set expectation")
		return nil
	}
	return &ExpectedCommit{expected: p.mockDB.Mock.ExpectCommit()}
}

// ExpectRollback sets up an expectation for a transaction rollback operation
func (p *Client) ExpectRollback() *ExpectedRollback {
	if p.mockDB == nil {
		log.Errorf("mockDB is nil, cannot set expectation")
		return nil
	}
	return &ExpectedRollback{expected: p.mockDB.Mock.ExpectRollback()}
}

// ExpectClose sets up an expectation for a database close operation
func (p *Client) ExpectClose() *ExpectedClose {
	if p.mockDB == nil {
		log.Errorf("mockDB is nil, cannot set expectation")
		return nil
	}
	return &ExpectedClose{expected: p.mockDB.Mock.ExpectClose()}
}

// ExpectationsWereMet checks if all mock expectations were met
// This should be called at the end of each test to ensure all expectations were satisfied
func (p *Client) ExpectationsWereMet() error {
	if p.mockDB == nil {
		return nil
	}
	return p.mockDB.ExpectationsWereMet()
}

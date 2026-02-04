package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/log"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// MockDB represents a mock database connection with sqlmock
type MockDB struct {
	DB    *gorm.DB
	Mock  sqlmock.Sqlmock
	SqlDB *sql.DB
}

// ExpectedQuery wraps sqlmock.ExpectedQuery
type ExpectedQuery struct {
	expected *sqlmock.ExpectedQuery
}

func (e *ExpectedQuery) WillReturnRows(rows *sqlmock.Rows) *ExpectedQuery {
	e.expected.WillReturnRows(rows)
	return e
}

func (e *ExpectedQuery) WillReturnError(err error) *ExpectedQuery {
	e.expected.WillReturnError(err)
	return e
}

func (e *ExpectedQuery) WithArgs(args ...interface{}) *ExpectedQuery {
	driverArgs := make([]driver.Value, len(args))
	for i, arg := range args {
		driverArgs[i] = arg
	}
	e.expected.WithArgs(driverArgs...)
	return e
}

// ExpectedExec wraps sqlmock.ExpectedExec
type ExpectedExec struct {
	expected *sqlmock.ExpectedExec
}

func (e *ExpectedExec) WillReturnResult(result sql.Result) *ExpectedExec {
	e.expected.WillReturnResult(result)
	return e
}

func (e *ExpectedExec) WillReturnError(err error) *ExpectedExec {
	e.expected.WillReturnError(err)
	return e
}

func (e *ExpectedExec) WithArgs(args ...interface{}) *ExpectedExec {
	driverArgs := make([]driver.Value, len(args))
	for i, arg := range args {
		driverArgs[i] = arg
	}
	e.expected.WithArgs(driverArgs...)
	return e
}

// ExpectedBegin wraps sqlmock.ExpectedBegin
type ExpectedBegin struct {
	expected *sqlmock.ExpectedBegin
}

func (e *ExpectedBegin) WillReturnError(err error) *ExpectedBegin {
	e.expected.WillReturnError(err)
	return e
}

// ExpectedCommit wraps sqlmock.ExpectedCommit
type ExpectedCommit struct {
	expected *sqlmock.ExpectedCommit
}

func (e *ExpectedCommit) WillReturnError(err error) *ExpectedCommit {
	e.expected.WillReturnError(err)
	return e
}

// ExpectedRollback wraps sqlmock.ExpectedRollback
type ExpectedRollback struct {
	expected *sqlmock.ExpectedRollback
}

func (e *ExpectedRollback) WillReturnError(err error) *ExpectedRollback {
	e.expected.WillReturnError(err)
	return e
}

// ExpectedClose wraps sqlmock.ExpectedClose
type ExpectedClose struct {
	expected *sqlmock.ExpectedClose
}

func (e *ExpectedClose) WillReturnError(err error) *ExpectedClose {
	e.expected.WillReturnError(err)
	return e
}

// newMock creates a new mock database connection using sqlmock
// This is an internal function for unit testing without requiring a real database
func newMock(c *Config, tables ...interface{}) (*Client, error) {
	log.Infof("creating mock database connection for type: %s", c.Type)

	// Create sqlmock instance
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Select dialector based on database type
	var dialector gorm.Dialector
	switch c.Type {
	case Sqlite:
		// For SQLite mock, we use postgres dialector as sqlite driver doesn't support custom connections
		log.Infof("using postgres dialector for sqlite mock (sqlite driver limitation)")
		dialector = postgres.New(postgres.Config{
			Conn: sqlDB,
		})

	case MySQL, TiDB:
		// MySQL and TiDB (MySQL-compatible) use mysql dialector
		dialector = mysql.New(mysql.Config{
			Conn:                      sqlDB,
			SkipInitializeWithVersion: true, // Skip version check in mock mode
		})

	case Postgres, GaussDB:
		// Postgres and GaussDB (PostgreSQL-compatible) use postgres dialector
		dialector = postgres.New(postgres.Config{
			Conn: sqlDB,
		})

	case ClickHouse:
		// ClickHouse uses clickhouse dialector
		// Note: ClickHouse driver may have limitations with sqlmock
		dialector = clickhouse.New(clickhouse.Config{
			Conn: sqlDB,
		})

	default:
		sqlDB.Close()
		err := fmt.Errorf("unsupported database type for mock: %s", c.Type)
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Create GORM instance with mock connection
	gormDB, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: &schema.NamingStrategy{
			SingularTable: true,
		},
		FullSaveAssociations: true,
		PrepareStmt:          false, // Disable prepared statements for mock

		DisableForeignKeyConstraintWhenMigrating: true,
		IgnoreRelationshipsWhenMigrating:         true,

		AllowGlobalUpdate: true,
		CreateBatchSize:   100,

		TranslateError: true,

		PropagateUnscoped: true,
	})
	if err != nil {
		log.Errorf("err:%v", err)
		sqlDB.Close()
		return nil, err
	}

	mockDB := &MockDB{
		DB:    gormDB,
		Mock:  mock,
		SqlDB: sqlDB,
	}

	client := &Client{
		db:         gormDB,
		mockDB:     mockDB,
		clientType: c.Type, // Use the actual database type
	}

	// Note: AutoMigrate is skipped in mock mode
	// Users should set up their own expectations for table creation
	if len(tables) > 0 {
		log.Warnf("AutoMigrate is skipped in mock mode. Please set up table expectations manually.")
	}

	return client, nil
}

// Close closes the mock database connection
func (m *MockDB) Close() error {
	if m.SqlDB != nil {
		err := m.SqlDB.Close()
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}
	return nil
}

// ExpectationsWereMet checks if all expectations were met
func (m *MockDB) ExpectationsWereMet() error {
	err := m.Mock.ExpectationsWereMet()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	return nil
}

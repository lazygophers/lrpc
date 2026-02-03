package db

import (
	"database/sql"
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

// NewMock creates a new mock database connection using sqlmock
// This is useful for unit testing without requiring a real database
func NewMock(c *Config, tables ...interface{}) (*Client, *MockDB, error) {
	log.Infof("creating mock database connection for type: %s", c.Type)

	// Create sqlmock instance
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, nil, err
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
		return nil, nil, err
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
		return nil, nil, err
	}

	client := &Client{
		db:         gormDB,
		clientType: c.Type, // Use the actual database type
	}

	mockDB := &MockDB{
		DB:    gormDB,
		Mock:  mock,
		SqlDB: sqlDB,
	}

	// Note: AutoMigrate is skipped in mock mode
	// Users should set up their own expectations for table creation
	if len(tables) > 0 {
		log.Warnf("AutoMigrate is skipped in mock mode. Please set up table expectations manually.")
	}

	return client, mockDB, nil
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

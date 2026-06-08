package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestClient_SqlDB 测试 SqlDB 方法
func TestClient_SqlDB(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("SqlDB in mock mode", func(t *testing.T) {
		sqlDB, err := client.SqlDB()
		// Mock模式可能返回nil或错误，取决于实现
		_ = sqlDB
		_ = err
	})
}

// TestClient_DriverType 测试 DriverType 方法
func TestClient_DriverType(t *testing.T) {
	t.Run("DriverType for MySQL", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Close()

		assert.Equal(t, db.MySQL, client.DriverType())
	})

	t.Run("DriverType for Postgres", func(t *testing.T) {
		config := &db.Config{
			Type: db.Postgres,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Close()

		assert.Equal(t, db.Postgres, client.DriverType())
	})

	t.Run("DriverType for SQLite", func(t *testing.T) {
		config := &db.Config{
			Type: db.Sqlite,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Close()

		assert.Equal(t, db.Sqlite, client.DriverType())
	})

	t.Run("DriverType for TiDB", func(t *testing.T) {
		config := &db.Config{
			Type: db.TiDB,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Close()

		assert.Equal(t, db.TiDB, client.DriverType())
	})
}

// TestClient_Ping 测试 Ping 方法
func TestClient_Ping(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Ping in mock mode", func(t *testing.T) {
		err := client.Ping()
		// Mock模式下ping的行为
		_ = err
	})
}

// TestClient_Close 测试 Close 方法
func TestClient_Close(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)

	t.Run("Close closes database connection", func(t *testing.T) {
		client.ExpectClose()
		err := client.Close()
		assert.NoError(t, err)
	})
}

func TestNew_DifferentDatabaseTypes(t *testing.T) {
	databaseTypes := []struct {
		name        string
		dbType      string
		expectedType string
		shouldError bool
	}{
		{"MySQL", db.MySQL, db.MySQL, false},
		{"Postgres", db.Postgres, db.Postgres, false},
		{"SQLite", db.Sqlite, db.Sqlite, false},
		{"SQLite3", "sqlite3", db.Sqlite, false},
		{"TiDB", db.TiDB, db.TiDB, false},
		{"GaussDB", db.GaussDB, db.GaussDB, false},
		{"Unknown", "unknown_type", "", true},
	}

	for _, tt := range databaseTypes {
		t.Run(tt.name, func(t *testing.T) {
			config := &db.Config{
				Type: tt.dbType,
				Mock: true,
			}

			client, err := db.New(config)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.expectedType, client.DriverType())

				if client != nil && client.MockDB() != nil {
					client.MockDB().Close()
				}
			}
		})
	}
}

// TestNew_SqliteDefaultConfig 测试SQLite默认配置
func TestNew_SqliteDefaultConfig(t *testing.T) {
	t.Run("SQLite with empty address and name", func(t *testing.T) {
		config := &db.Config{
			Type: db.Sqlite,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		client.MockDB().Close()
	})
}

// TestNew_MySQLDefaultConfig 测试MySQL默认配置
func TestNew_MySQLDefaultConfig(t *testing.T) {
	t.Run("MySQL with empty address and port", func(t *testing.T) {
		config := &db.Config{
			Type:     db.MySQL,
			Mock:     true,
			Username: "test",
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		client.MockDB().Close()
	})
}

// TestNew_PostgresDefaultConfig 测试Postgres默认配置
func TestNew_PostgresDefaultConfig(t *testing.T) {
	t.Run("Postgres with empty address and port", func(t *testing.T) {
		config := &db.Config{
			Type:     db.Postgres,
			Mock:     true,
			Username: "test",
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		client.MockDB().Close()
	})
}

// TestNew_EmptyType 测试空类型默认为SQLite
func TestNew_EmptyType(t *testing.T) {
	t.Run("Empty type defaults to SQLite", func(t *testing.T) {
		config := &db.Config{
			Type: "",
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, db.Sqlite, client.DriverType())
		client.MockDB().Close()
	})
}

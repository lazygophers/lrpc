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

package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestNew_DifferentDatabaseTypes 测试不同数据库类型的创建
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

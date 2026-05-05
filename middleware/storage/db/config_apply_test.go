package db_test

import (
	"testing"
	"time"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestConfig_Apply_PostgresAliases 测试 Postgres 别名
func TestConfig_Apply_PostgresAliases(t *testing.T) {
	postgresAliases := []string{"postgres", "pg", "postgresql", "pgsql"}

	for _, alias := range postgresAliases {
		t.Run("Postgres alias: "+alias, func(t *testing.T) {
			config := &db.Config{
				Type: alias,
				Mock: true,
			}

			client, err := db.New(config)
			assert.NoError(t, err)
			assert.NotNil(t, client)
			assert.Equal(t, db.Postgres, client.DriverType())
			client.MockDB().Close()
		})
	}
}

// TestConfig_Apply_DefaultValues 测试默认值设置
func TestConfig_Apply_DefaultValues(t *testing.T) {
	t.Run("MySQL with default values", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		client.MockDB().Close()
	})

	t.Run("Postgres with default values", func(t *testing.T) {
		config := &db.Config{
			Type: db.Postgres,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		client.MockDB().Close()
	})
}

// TestConfig_Apply_Timeouts 测试超时配置
func TestConfig_Apply_Timeouts(t *testing.T) {
	t.Run("MySQL with custom timeouts", func(t *testing.T) {
		config := &db.Config{
			Type:           db.MySQL,
			Mock:           true,
			ConnectTimeout: time.Second * 10,
			ReadTimeout:    time.Second * 60,
			WriteTimeout:   time.Second * 60,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		client.MockDB().Close()
	})

	t.Run("MySQL with zero timeouts (uses defaults)", func(t *testing.T) {
		config := &db.Config{
			Type:           db.MySQL,
			Mock:           true,
			ConnectTimeout: 0,
			ReadTimeout:    0,
			WriteTimeout:   0,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		client.MockDB().Close()
	})
}

// TestConfig_DSN 测试 DSN 生成
func TestConfig_DSN(t *testing.T) {
	t.Run("DSN for MySQL", func(t *testing.T) {
		config := &db.Config{
			Type:     db.MySQL,
			Address:  "localhost",
			Port:     3306,
			Username: "test",
			Password: "pass",
			Name:     "testdb",
		}

		dsn := config.DSN()
		assert.Contains(t, dsn, "test:pass@tcp(localhost:3306)")
	})

	t.Run("DSN for Postgres", func(t *testing.T) {
		config := &db.Config{
			Type:     db.Postgres,
			Address:  "localhost",
			Port:     5432,
			Username: "test",
			Password: "pass",
			Name:     "testdb",
		}

		dsn := config.DSN()
		assert.Contains(t, dsn, "host=localhost")
	})
}

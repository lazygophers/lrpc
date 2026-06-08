package db_test

import (
"github.com/lazygophers/lrpc/middleware/storage/db"
"gotest.tools/v3/assert"
"strings"
"testing"
"time"
)

// TestConfig_DSN_Sqlite 测试 SQLite DSN 生成
func TestConfig_DSN_Sqlite(t *testing.T) {
	t.Run("sqlite with file address and name", func(t *testing.T) {
		config := &db.Config{
			Type:    db.Sqlite,
			Address: "file:/tmp/test",
			Name:    "test",
		}
		dsn := config.DSN()
		assert.Assert(t, dsn != "", "DSN should not be empty")
		assert.Assert(t, strings.Contains(dsn, "test.db?"), "should contain .db extension")
	})

	t.Run("sqlite with custom name", func(t *testing.T) {
		config := &db.Config{
			Type:    db.Sqlite,
			Address: "file:/tmp",
			Name:    "custom",
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "custom.db"), "should contain custom name")
	})

	t.Run("sqlite with password", func(t *testing.T) {
		config := &db.Config{
			Type:     db.Sqlite,
			Address:  "file:/tmp",
			Name:     "test",
			Password: "secret",
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "_key=secret"), "should contain encryption key")
		assert.Assert(t, strings.Contains(dsn, "_cipher=sqlcipher"), "should contain cipher parameter")
	})

	t.Run("sqlite with extras", func(t *testing.T) {
		config := &db.Config{
			Type:    db.Sqlite,
			Address: "file:/tmp",
			Name:    "test",
			Extras: map[string]string{
				"custom_param": "value123",
			},
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "custom_param=value123"), "should contain extra parameters")
	})
}

// TestConfig_DSN_MySQL 测试 MySQL DSN 生成
func TestConfig_DSN_MySQL(t *testing.T) {
	t.Run("mysql with all fields", func(t *testing.T) {
		config := &db.Config{
			Type:     db.MySQL,
			Address:  "127.0.0.1",
			Port:     3306,
			Name:     "test",
			Username: "root",
			Password: "pass",
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "@tcp(127.0.0.1:3306)/"), "should contain address")
		assert.Assert(t, strings.Contains(dsn, "root:pass"), "should contain credentials")
		assert.Assert(t, strings.Contains(dsn, "/test?"), "should contain database name")
	})

	t.Run("mysql with custom values", func(t *testing.T) {
		config := &db.Config{
			Type:     db.MySQL,
			Address:  "192.168.1.100",
			Port:     3307,
			Name:     "mydb",
			Username: "root",
			Password: "password",
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "root:password"), "should contain credentials")
		assert.Assert(t, strings.Contains(dsn, "192.168.1.100:3307"), "should contain custom address")
	})
}

// TestConfig_DSN_Postgres 测试 PostgreSQL DSN 生成
func TestConfig_DSN_Postgres(t *testing.T) {
	t.Run("postgres standard type", func(t *testing.T) {
		config := &db.Config{
			Type:     db.Postgres,
			Address:  "localhost",
			Port:     5432,
			Name:     "testdb",
			Username: "pguser",
			Password: "pgpass",
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "host=localhost"), "should contain host")
		assert.Assert(t, strings.Contains(dsn, "port=5432"), "should contain port")
		assert.Assert(t, strings.Contains(dsn, "user=pguser"), "should contain user")
		assert.Assert(t, strings.Contains(dsn, "password=pgpass"), "should contain password")
		assert.Assert(t, strings.Contains(dsn, "dbname=testdb"), "should contain dbname")
	})

	t.Run("postgres with extras", func(t *testing.T) {
		config := &db.Config{
			Type:     db.Postgres,
			Address:  "localhost",
			Port:     5432,
			Name:     "testdb",
			Username: "pguser",
			Password: "pgpass",
			Extras: map[string]string{
				"connect_timeout": "10",
			},
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "connect_timeout=10"), "should contain extras")
	})
}

// TestConfig_DSN_ClickHouse 测试 ClickHouse DSN 生成
func TestConfig_DSN_ClickHouse(t *testing.T) {
	t.Run("clickhouse standard type", func(t *testing.T) {
		config := &db.Config{
			Type:     db.ClickHouse,
			Address:  "localhost",
			Port:     9000,
			Name:     "default",
			Username: "default",
			Password: "pass",
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "clickhouse://"), "should contain protocol")
		assert.Assert(t, strings.Contains(dsn, "default:pass@"), "should contain credentials")
		assert.Assert(t, strings.Contains(dsn, "localhost:9000"), "should contain address")
	})

	t.Run("clickhouse without credentials", func(t *testing.T) {
		config := &db.Config{
			Type:    "clickhouse",
			Address: "localhost",
			Port:    9000,
			Name:    "default",
		}
		dsn := config.DSN()
		assert.Assert(t, !strings.Contains(dsn, "@"), "should not contain @ without credentials")
	})
}

// TestConfig_DSN_TiDB 测试 TiDB DSN 生成
func TestConfig_DSN_TiDB(t *testing.T) {
	config := &db.Config{
		Type:     db.TiDB,
		Address:  "localhost",
		Port:     4000,
		Name:     "test",
		Username: "root",
		Password: "pass",
	}
	dsn := config.DSN()
	assert.Assert(t, strings.Contains(dsn, "root:pass"), "should contain credentials")
	assert.Assert(t, strings.Contains(dsn, "@tcp(localhost:4000)/"), "should contain address")
	assert.Assert(t, strings.Contains(dsn, "/test?"), "should contain database name")
}

// TestConfig_DSN_GaussDB 测试 GaussDB DSN 生成
func TestConfig_DSN_GaussDB(t *testing.T) {
	config := &db.Config{
		Type:     db.GaussDB,
		Address:  "localhost",
		Port:     5432,
		Name:     "testdb",
		Username: "gaussuser",
		Password: "gausspass",
	}
	dsn := config.DSN()
	assert.Assert(t, strings.Contains(dsn, "host=localhost"), "should contain host")
	assert.Assert(t, strings.Contains(dsn, "port=5432"), "should contain port")
	assert.Assert(t, strings.Contains(dsn, "user=gaussuser"), "should contain user")
	assert.Assert(t, strings.Contains(dsn, "password=gausspass"), "should contain password")
	assert.Assert(t, strings.Contains(dsn, "dbname=testdb"), "should contain dbname")
}

// TestConfig_DSN_UnknownType 测试未知类型
func TestConfig_DSN_UnknownType(t *testing.T) {
	config := &db.Config{
		Type: "unknown",
	}
	dsn := config.DSN()
	assert.Equal(t, "", dsn, "unknown type should return empty DSN")
}

func TestConfig_Apply_PostgresAliases(t *testing.T) {
	postgresAliases := []string{"postgres", "pg", "postgresql", "pgsql"}

	for _, alias := range postgresAliases {
		t.Run("Postgres alias: "+alias, func(t *testing.T) {
			config := &db.Config{
				Type: alias,
				Mock: true,
			}

			client, err := db.New(config)
			assert.NilError(t, err)
			assert.Assert(t, client != nil)
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
		assert.NilError(t, err)
		assert.Assert(t, client != nil)
		client.MockDB().Close()
	})

	t.Run("Postgres with default values", func(t *testing.T) {
		config := &db.Config{
			Type: db.Postgres,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NilError(t, err)
		assert.Assert(t, client != nil)
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
		assert.NilError(t, err)
		assert.Assert(t, client != nil)
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
		assert.NilError(t, err)
		assert.Assert(t, client != nil)
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
		assert.Assert(t, strings.Contains(dsn, "test:pass@tcp(localhost:3306)"), "test:pass@tcp(localhost:3306)")
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
		assert.Assert(t, strings.Contains(dsn, "host=localhost"), "host=localhost")
	})
}

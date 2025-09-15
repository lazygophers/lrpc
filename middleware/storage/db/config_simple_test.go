package db_test

import (
	"strings"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

func TestConfigDSN(t *testing.T) {
	t.Run("sqlite with defaults", func(t *testing.T) {
		config := &db.Config{
			Type: db.Sqlite,
		}
		dsn := config.DSN()
		// Should contain basic sqlite elements
		assert.Assert(t, strings.Contains(dsn, ".db"))
	})
	
	t.Run("sqlite with custom values", func(t *testing.T) {
		config := &db.Config{
			Type:    db.Sqlite,
			Address: "file:/custom/path",
			Name:    "mydb",
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "/custom/path"))
		assert.Assert(t, strings.Contains(dsn, "mydb.db"))
	})
	
	t.Run("sqlite with auth", func(t *testing.T) {
		config := &db.Config{
			Type:     db.Sqlite,
			Username: "user",
			Password: "pass",
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "_auth=1"))
		assert.Assert(t, strings.Contains(dsn, "_auth_user=user"))
		assert.Assert(t, strings.Contains(dsn, "_auth_pass=pass"))
	})
	
	t.Run("sqlite with extras", func(t *testing.T) {
		config := &db.Config{
			Type: db.Sqlite,
			Extras: map[string]string{
				"custom": "value",
				"cache":  "shared",
			},
		}
		dsn := config.DSN()
		assert.Assert(t, strings.Contains(dsn, "custom=value"))
		assert.Assert(t, strings.Contains(dsn, "cache=shared"))
	})
	
	t.Run("mysql returns empty", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
		}
		dsn := config.DSN()
		assert.Equal(t, "", dsn)
	})
	
	t.Run("postgres returns empty", func(t *testing.T) {
		config := &db.Config{
			Type: "postgres",
		}
		dsn := config.DSN()
		assert.Equal(t, "", dsn)
	})
	
	t.Run("empty type returns empty DSN", func(t *testing.T) {
		config := &db.Config{}
		dsn := config.DSN()
		// Empty config returns empty DSN until apply() is called
		assert.Equal(t, "", dsn)
	})
	
	t.Run("explicit sqlite type works", func(t *testing.T) {
		config := &db.Config{Type: db.Sqlite}
		dsn := config.DSN()
		// Explicit sqlite should generate DSN
		assert.Assert(t, dsn != "")
		assert.Assert(t, strings.Contains(dsn, ".db"))
	})
}
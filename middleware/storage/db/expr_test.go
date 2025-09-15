package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

func TestExpr(t *testing.T) {
	t.Run("simple expression", func(t *testing.T) {
		expr := db.Expr("COUNT(*)")
		assert.Equal(t, "COUNT(*)", expr.SQL)
		assert.Equal(t, 0, len(expr.Vars))
	})
	
	t.Run("expression with args", func(t *testing.T) {
		expr := db.Expr("name = ? AND age > ?", "John", 18)
		assert.Equal(t, "name = ? AND age > ?", expr.SQL)
		assert.Equal(t, 2, len(expr.Vars))
		assert.Equal(t, "John", expr.Vars[0])
		assert.Equal(t, 18, expr.Vars[1])
	})
	
	t.Run("expression without args", func(t *testing.T) {
		expr := db.Expr("NOW()")
		assert.Equal(t, "NOW()", expr.SQL)
		assert.Equal(t, 0, len(expr.Vars))
	})
}

func TestExprInc(t *testing.T) {
	t.Run("increment by 1", func(t *testing.T) {
		expr := db.ExprInc("counter")
		assert.Equal(t, "counter + 1", expr.SQL)
		assert.Equal(t, 0, len(expr.Vars))
	})
	
	t.Run("increment different field", func(t *testing.T) {
		expr := db.ExprInc("view_count")
		assert.Equal(t, "view_count + 1", expr.SQL)
		assert.Equal(t, 0, len(expr.Vars))
	})
}

func TestExprIncBy(t *testing.T) {
	t.Run("increment by specific amount", func(t *testing.T) {
		expr := db.ExprIncBy("counter", 5)
		assert.Equal(t, "counter + 5", expr.SQL)
		assert.Equal(t, 0, len(expr.Vars))
	})
	
	t.Run("increment by negative amount", func(t *testing.T) {
		expr := db.ExprIncBy("counter", -3)
		assert.Equal(t, "counter + -3", expr.SQL)
		assert.Equal(t, 0, len(expr.Vars))
	})
	
	t.Run("increment by zero", func(t *testing.T) {
		expr := db.ExprIncBy("counter", 0)
		assert.Equal(t, "counter + 0", expr.SQL)
		assert.Equal(t, 0, len(expr.Vars))
	})
}

func TestExprIf(t *testing.T) {
	// Save original DefaultDriver
	originalDriver := db.DefaultDriver
	defer func() {
		db.DefaultDriver = originalDriver
	}()
	
	t.Run("mysql driver", func(t *testing.T) {
		db.DefaultDriver = "mysql"
		expr := db.ExprIf("status = 1", "active", "inactive")
		assert.Equal(t, "IF(status = 1, ?, ?)", expr.SQL)
		assert.Equal(t, 2, len(expr.Vars))
		assert.Equal(t, "active", expr.Vars[0])
		assert.Equal(t, "inactive", expr.Vars[1])
	})
	
	t.Run("sqlite driver", func(t *testing.T) {
		db.DefaultDriver = "sqlite"
		expr := db.ExprIf("status = 1", "active", "inactive")
		assert.Equal(t, "IIF(status = 1, ?, ?)", expr.SQL)
		assert.Equal(t, 2, len(expr.Vars))
		assert.Equal(t, "active", expr.Vars[0])
		assert.Equal(t, "inactive", expr.Vars[1])
	})
	
	t.Run("sqlite3 driver", func(t *testing.T) {
		db.DefaultDriver = "sqlite3"
		expr := db.ExprIf("status = 1", "active", "inactive")
		assert.Equal(t, "IIF(status = 1, ?, ?)", expr.SQL)
		assert.Equal(t, 2, len(expr.Vars))
		assert.Equal(t, "active", expr.Vars[0])
		assert.Equal(t, "inactive", expr.Vars[1])
	})
	
	t.Run("default driver (postgres)", func(t *testing.T) {
		db.DefaultDriver = "postgres"
		expr := db.ExprIf("status = 1", "active", "inactive")
		assert.Equal(t, "IF(status = 1, ?, ?)", expr.SQL)
		assert.Equal(t, 2, len(expr.Vars))
		assert.Equal(t, "active", expr.Vars[0])
		assert.Equal(t, "inactive", expr.Vars[1])
	})
	
	t.Run("with numeric values", func(t *testing.T) {
		db.DefaultDriver = "mysql"
		expr := db.ExprIf("count > 0", 1, 0)
		assert.Equal(t, "IF(count > 0, ?, ?)", expr.SQL)
		assert.Equal(t, 2, len(expr.Vars))
		assert.Equal(t, 1, expr.Vars[0])
		assert.Equal(t, 0, expr.Vars[1])
	})
}

func TestDefaultDriver(t *testing.T) {
	// Test that DefaultDriver is properly initialized
	assert.Equal(t, "mysql", db.DefaultDriver)
}
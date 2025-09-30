package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
	"gorm.io/gorm/clause"
)

func TestExpr(t *testing.T) {
	t.Run("basic expression", func(t *testing.T) {
		expr := db.Expr("field = ?", 123)
		assert.Assert(t, expr.SQL != "")
		assert.Equal(t, "field = ?", expr.SQL)
		assert.Equal(t, 1, len(expr.Vars))
		assert.Equal(t, 123, expr.Vars[0])
	})

	t.Run("expression without args", func(t *testing.T) {
		expr := db.Expr("NOW()")
		assert.Equal(t, "NOW()", expr.SQL)
		assert.Equal(t, 0, len(expr.Vars))
	})

	t.Run("expression with multiple args", func(t *testing.T) {
		expr := db.Expr("field1 = ? AND field2 = ?", "value1", "value2")
		assert.Equal(t, "field1 = ? AND field2 = ?", expr.SQL)
		assert.Equal(t, 2, len(expr.Vars))
	})
}

func TestExprInc(t *testing.T) {
	testCases := []struct {
		name       string
		clientType string
		field      string
		expected   string
	}{
		{
			name:       "MySQL increment",
			clientType: db.MySQL,
			field:      "count",
			expected:   "`count` + 1",
		},
		{
			name:       "SQLite increment",
			clientType: db.Sqlite,
			field:      "count",
			expected:   "\"count\" + 1",
		},
		{
			name:       "PostgreSQL increment",
			clientType: db.Postgres,
			field:      "count",
			expected:   "\"count\" + 1",
		},
		{
			name:       "TiDB increment",
			clientType: db.TiDB,
			field:      "score",
			expected:   "`score` + 1",
		},
		{
			name:       "ClickHouse increment",
			clientType: db.ClickHouse,
			field:      "total",
			expected:   "`total` + 1",
		},
		{
			name:       "GaussDB increment",
			clientType: db.GaussDB,
			field:      "amount",
			expected:   "\"amount\" + 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expr := db.ExprInc(tc.clientType, tc.field)
			assert.Equal(t, tc.expected, expr.SQL)
			assert.Equal(t, 0, len(expr.Vars))
		})
	}
}

func TestExprIncBy(t *testing.T) {
	testCases := []struct {
		name       string
		clientType string
		field      string
		count      int64
		expected   string
	}{
		{
			name:       "MySQL increment by 5",
			clientType: db.MySQL,
			field:      "count",
			count:      5,
			expected:   "`count` + 5",
		},
		{
			name:       "SQLite increment by 10",
			clientType: db.Sqlite,
			field:      "count",
			count:      10,
			expected:   "\"count\" + 10",
		},
		{
			name:       "PostgreSQL decrement by 3",
			clientType: db.Postgres,
			field:      "balance",
			count:      -3,
			expected:   "\"balance\" + -3",
		},
		{
			name:       "TiDB increment by 100",
			clientType: db.TiDB,
			field:      "points",
			count:      100,
			expected:   "`points` + 100",
		},
		{
			name:       "ClickHouse increment by 50",
			clientType: db.ClickHouse,
			field:      "views",
			count:      50,
			expected:   "`views` + 50",
		},
		{
			name:       "GaussDB decrement by 2",
			clientType: db.GaussDB,
			field:      "stock",
			count:      -2,
			expected:   "\"stock\" + -2",
		},
		{
			name:       "increment by 0",
			clientType: db.MySQL,
			field:      "field",
			count:      0,
			expected:   "`field` + 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expr := db.ExprIncBy(tc.clientType, tc.field, tc.count)
			assert.Equal(t, tc.expected, expr.SQL)
			assert.Equal(t, 0, len(expr.Vars))
		})
	}
}

func TestExprIf(t *testing.T) {
	testCases := []struct {
		name       string
		clientType string
		expr       string
		ok         interface{}
		nok        interface{}
		expectedIf string
	}{
		{
			name:       "MySQL IF",
			clientType: db.MySQL,
			expr:       "status = 1",
			ok:         "active",
			nok:        "inactive",
			expectedIf: "IF(status = 1, ?, ?)",
		},
		{
			name:       "SQLite IIF",
			clientType: db.Sqlite,
			expr:       "age > 18",
			ok:         "adult",
			nok:        "minor",
			expectedIf: "IIF(age > 18, ?, ?)",
		},
		{
			name:       "PostgreSQL IF (defaults to IF)",
			clientType: db.Postgres,
			expr:       "price > 100",
			ok:         "expensive",
			nok:        "cheap",
			expectedIf: "IF(price > 100, ?, ?)",
		},
		{
			name:       "TiDB IF",
			clientType: db.TiDB,
			expr:       "score >= 60",
			ok:         "pass",
			nok:        "fail",
			expectedIf: "IF(score >= 60, ?, ?)",
		},
		{
			name:       "ClickHouse IF",
			clientType: db.ClickHouse,
			expr:       "count > 0",
			ok:         1,
			nok:        0,
			expectedIf: "IF(count > 0, ?, ?)",
		},
		{
			name:       "GaussDB IF (defaults to IF)",
			clientType: db.GaussDB,
			expr:       "enabled = true",
			ok:         "yes",
			nok:        "no",
			expectedIf: "IF(enabled = true, ?, ?)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expr := db.ExprIf(tc.clientType, tc.expr, tc.ok, tc.nok)
			assert.Equal(t, tc.expectedIf, expr.SQL)
			assert.Equal(t, 2, len(expr.Vars))
			assert.Equal(t, tc.ok, expr.Vars[0])
			assert.Equal(t, tc.nok, expr.Vars[1])
		})
	}
}

func TestExprFieldQuoting(t *testing.T) {
	t.Run("field with table prefix", func(t *testing.T) {
		expr := db.ExprInc(db.MySQL, "users.count")
		assert.Equal(t, "`users`.`count` + 1", expr.SQL)
	})

	t.Run("field already quoted", func(t *testing.T) {
		expr := db.ExprInc(db.MySQL, "`count`")
		assert.Equal(t, "`count` + 1", expr.SQL)
	})

	t.Run("SQLite field with table prefix", func(t *testing.T) {
		expr := db.ExprIncBy(db.Sqlite, "orders.total", 100)
		assert.Equal(t, "\"orders\".\"total\" + 100", expr.SQL)
	})
}

func TestExprReturnType(t *testing.T) {
	t.Run("return type is clause.Expr", func(t *testing.T) {
		var expr clause.Expr

		expr = db.Expr("test")
		assert.Assert(t, expr.SQL != "")

		expr = db.ExprInc(db.MySQL, "field")
		assert.Assert(t, expr.SQL != "")

		expr = db.ExprIncBy(db.MySQL, "field", 5)
		assert.Assert(t, expr.SQL != "")

		expr = db.ExprIf(db.MySQL, "condition", "ok", "nok")
		assert.Assert(t, expr.SQL != "")
	})
}

func TestExprEdgeCases(t *testing.T) {
	t.Run("empty field name", func(t *testing.T) {
		expr := db.ExprInc(db.MySQL, "")
		assert.Equal(t, "`` + 1", expr.SQL)
	})

	t.Run("field with special characters", func(t *testing.T) {
		expr := db.ExprIncBy(db.Sqlite, "field-name", 1)
		assert.Equal(t, "\"field-name\" + 1", expr.SQL)
	})

	t.Run("large increment value", func(t *testing.T) {
		expr := db.ExprIncBy(db.MySQL, "bignum", 9223372036854775807)
		assert.Equal(t, "`bignum` + 9223372036854775807", expr.SQL)
	})

	t.Run("negative increment value", func(t *testing.T) {
		expr := db.ExprIncBy(db.MySQL, "balance", -1000)
		assert.Equal(t, "`balance` + -1000", expr.SQL)
	})

	t.Run("ExprIf with nil values", func(t *testing.T) {
		expr := db.ExprIf(db.MySQL, "field IS NULL", nil, "default")
		assert.Equal(t, "IF(field IS NULL, ?, ?)", expr.SQL)
		assert.Equal(t, 2, len(expr.Vars))
	})

	t.Run("ExprIf with complex expression", func(t *testing.T) {
		expr := db.ExprIf(db.Sqlite, "status IN (1, 2, 3) AND deleted_at IS NULL", "valid", "invalid")
		assert.Equal(t, "IIF(status IN (1, 2, 3) AND deleted_at IS NULL, ?, ?)", expr.SQL)
	})
}

// Benchmark tests

func BenchmarkExpr(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = db.Expr("field = ?", 123)
	}
}

func BenchmarkExprInc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = db.ExprInc(db.MySQL, "count")
	}
}

func BenchmarkExprIncBy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = db.ExprIncBy(db.MySQL, "count", 5)
	}
}

func BenchmarkExprIf(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = db.ExprIf(db.MySQL, "status = 1", "active", "inactive")
	}
}

func BenchmarkExprIfSqlite(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = db.ExprIf(db.Sqlite, "status = 1", "active", "inactive")
	}
}

// Concurrent benchmark tests

func BenchmarkConcurrentExprInc(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = db.ExprInc(db.MySQL, "count")
		}
	})
}

func BenchmarkConcurrentExprIf(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = db.ExprIf(db.MySQL, "status = 1", "active", "inactive")
		}
	})
}

// Memory allocation tests

func BenchmarkExprMemory(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := db.ExprInc(db.MySQL, "field")
		_ = expr.SQL
	}
}
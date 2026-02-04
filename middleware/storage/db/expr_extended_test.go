package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestExprIfExtended 测试 ExprIf 的扩展场景
func TestExprIfExtended(t *testing.T) {
	t.Run("ExprIf with SQLite", func(t *testing.T) {
		// 测试 SQLite 类型的 ExprIf - 使用 IIF 函数
		expr := db.ExprIf(db.Sqlite, "is_active", 1, 0)
		assert.NotNil(t, expr)
		// 应该生成 "IIF(is_active, ?, ?)" 格式
	})

	t.Run("ExprIf with MySQL", func(t *testing.T) {
		// 测试 MySQL 类型的 ExprIf - 使用 IF 函数
		expr := db.ExprIf(db.MySQL, "status", "active", "inactive")
		assert.NotNil(t, expr)
		// 应该生成 "IF(status, ?, ?)" 格式
	})

	t.Run("ExprIf with PostgreSQL", func(t *testing.T) {
		// 测试 PostgreSQL 类型 - 也使用 IF 函数
		expr := db.ExprIf(db.Postgres, "enabled", true, false)
		assert.NotNil(t, expr)
	})

	t.Run("ExprIf with different types", func(t *testing.T) {
		// 测试不同的值类型
		expr1 := db.ExprIf(db.MySQL, "count", 100, 0)
		assert.NotNil(t, expr1)

		expr2 := db.ExprIf(db.MySQL, "name", "default", nil)
		assert.NotNil(t, expr2)

		expr3 := db.ExprIf(db.Sqlite, "price", 19.99, 0.0)
		assert.NotNil(t, expr3)
	})

	t.Run("ExprIf in actual query with MySQL", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		// 使用 ExprIf 更新字段
		client.ExpectExec("UPDATE test_users SET .*").
			WillReturnResult(sqlmock.NewResult(1, 1))

		result := client.NewScoop().Model(TestUser{}).
			Where("id", 1).
			Updates("status", db.ExprIf(db.MySQL, "age > 18", "adult", "minor"))
		assert.NoError(t, result.Error)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("ExprIf in actual query with SQLite", func(t *testing.T) {
		config := &db.Config{
			Type: db.Sqlite,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		// 使用 ExprIf 更新字段
		client.ExpectExec("UPDATE test_users SET .*").
			WillReturnResult(sqlmock.NewResult(1, 1))

		result := client.NewScoop().Model(TestUser{}).
			Where("id", 1).
			Updates("category", db.ExprIf(db.Sqlite, "score > 60", "pass", "fail"))
		assert.NoError(t, result.Error)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

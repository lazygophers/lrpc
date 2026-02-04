package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestExprFunctions 测试 Expr 相关函数
func TestExprFunctions(t *testing.T) {
	t.Run("ExprInc with simple field", func(t *testing.T) {
		// 测试 ExprInc 函数
		expr := db.ExprInc(db.MySQL, "counter")
		assert.NotNil(t, expr)
		// Expr 生成的 SQL 格式为 "`counter` + 1"
	})

	t.Run("ExprIncBy with custom increment", func(t *testing.T) {
			// 测试 ExprIncBy 函数
		expr := db.ExprIncBy(db.MySQL, "views", 5)
		assert.NotNil(t, expr)
		// 生成 "`views` + 5"
	})

	t.Run("ExprIncBy with table prefix", func(t *testing.T) {
		// 测试带表前缀的字段
		expr := db.ExprIncBy(db.MySQL, "users.count", 1)
		assert.NotNil(t, expr)
		// 应该生成 `users`.`count` + 1
	})

	t.Run("ExprIncBy with PostgreSQL", func(t *testing.T) {
		// 测试 PostgreSQL 类型
		expr := db.ExprIncBy(db.Postgres, "id", 10)
		assert.NotNil(t, expr)
		// PostgreSQL 使用双引号
	})

	t.Run("ExprIf with true condition", func(t *testing.T) {
		// 测试 ExprIf 函数 - true 分支
		expr := db.ExprIf(db.MySQL, "status", "active", "inactive")
		assert.NotNil(t, expr)
		// 生成 "IF(status, ?, ?)" 格式
	})

	t.Run("ExprIf with false condition", func(t *testing.T) {
		// 测试 ExprIf 函数 - false 分支（使用 ExprIf 的逻辑）
		expr := db.ExprIf(db.MySQL, "is_active", true, false)
		assert.NotNil(t, expr)
	})

	t.Run("ExprInc in actual query", func(t *testing.T) {
		// 测试在实际查询中使用 ExprInc
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		// 使用 ExprInc 增加计数器
		client.ExpectExec("UPDATE test_users SET .*").
			WillReturnResult(sqlmock.NewResult(1, 1))

		result := client.NewScoop().Model(TestUser{}).
			Where("id", 1).
			Updates("counter", db.ExprInc(db.MySQL, "views"))
		assert.NoError(t, result.Error)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("ExprIncBy in actual query", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		// 使用 ExprIncBy
		client.ExpectExec("UPDATE test_users SET .*").
			WillReturnResult(sqlmock.NewResult(1, 1))

		result := client.NewScoop().Model(TestUser{}).
			Where("id", 1).
			Updates("age", db.ExprIncBy(db.MySQL, "age", 5))
		assert.NoError(t, result.Error)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("quoteFieldWithTable edge cases", func(t *testing.T) {
		// 间接测试 quoteFieldWithTable 函数
		// 测试已经带引号的字段名
		expr1 := db.ExprIncBy(db.MySQL, "`field`", 1)
		assert.NotNil(t, expr1)

		// 测试带表前缀的字段
		expr2 := db.ExprIncBy(db.MySQL, "table.field", 1)
		assert.NotNil(t, expr2)

		// 测试简单的字段名
		expr3 := db.ExprIncBy(db.MySQL, "simple", 1)
		assert.NotNil(t, expr3)

		// 测试 PostgreSQL 的双引号
		expr4 := db.ExprIncBy(db.Postgres, "field", 1)
		assert.NotNil(t, expr4)
	})
}

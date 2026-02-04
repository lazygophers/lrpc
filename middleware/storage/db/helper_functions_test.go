package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestHelperFunctions 测试辅助函数
func TestHelperFunctions(t *testing.T) {
	t.Run("QuoteFieldName", func(t *testing.T) {
		// 测试字段名引用
		result := db.QuoteFieldName("test_field", "`")
		assert.NotEmpty(t, result)
	})

	t.Run("ExprInc", func(t *testing.T) {
		// 测试表达式增量
		result := db.ExprInc(db.MySQL, "age")
		assert.NotEmpty(t, result)
	})

	t.Run("ExprIncBy", func(t *testing.T) {
		result := db.ExprIncBy(db.MySQL, "count", int64(10))
		assert.NotEmpty(t, result)
	})

	t.Run("ExprIf", func(t *testing.T) {
		// 测试条件表达式
		result1 := db.ExprIf(db.MySQL, "status = ?", "1", "active")
		assert.NotEmpty(t, result1)

		result2 := db.ExprIf(db.MySQL, "status = ?", "0", "inactive")
		assert.NotEmpty(t, result2)
	})
}

// TestSQLGeneration 测试 SQL 生成
func TestSQLGeneration(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)

	t.Run("ToSQL for select", func(t *testing.T) {
		sql := client.NewScoop().Model(TestUser{}).ToSQL(db.SQLOperationSelect)
		assert.NotEmpty(t, sql)
	})

	t.Run("ToSQL for insert", func(t *testing.T) {
		user := &TestUser{
			Name:  "Test",
			Email: "test@example.com",
			Age:   25,
		}
		sql := client.NewScoop().Model(TestUser{}).ToSQL(db.SQLOperationInsert, user)
		assert.NotEmpty(t, sql)
	})

	t.Run("ToSQL for update", func(t *testing.T) {
		updates := map[string]interface{}{
			"name": "Updated",
			"age":  30,
		}
		sql := client.NewScoop().Model(TestUser{}).ToSQL(db.SQLOperationUpdate, updates)
		assert.NotEmpty(t, sql)
	})
}

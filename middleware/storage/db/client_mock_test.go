package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestClientMockMethods 测试 Client 的 Mock 方法
// 这个测试演示了两种配置 mock 的方式：
// 1. 使用 client.MockDB() 访问 MockDB
// 2. 使用 client.ExpectQuery/ExpectExec 等便捷方法
func TestClientMockMethods(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("use client.MockDB() method", func(t *testing.T) {
		// 方式1: 使用 client.MockDB() 访问 MockDB
		mockDB := client.MockDB()
		assert.NotNil(t, mockDB)

		// 设置查询期望 - 不要使用反引号
		mockDB.Mock.ExpectQuery("SELECT \\* FROM test_users").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "John"))

		// 执行查询
		var user TestUser
		result := client.NewScoop().Model(TestUser{}).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, 1, user.Id)
		assert.Equal(t, "John", user.Name)

		// 验证期望
		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("use client.ExpectQuery() method", func(t *testing.T) {
		// 方式2: 使用 client.ExpectQuery() 便捷方法
		client.ExpectQuery("SELECT \\* FROM test_users").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(2), "Alice"))

		// 执行查询
		var user TestUser
		result := client.NewScoop().Model(TestUser{}).First(&user)
		assert.NoError(t, result.Error)
		assert.Equal(t, 2, user.Id)
		assert.Equal(t, "Alice", user.Name)

		// 验证期望
		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("use client.ExpectExec() method", func(t *testing.T) {
		// 使用 client.ExpectExec() 便捷方法
		client.ExpectExec("UPDATE test_users SET .* WHERE .*").
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 执行更新
		result := client.NewScoop().Model(TestUser{}).Where("id", 1).Updates("name", "Updated")
		assert.NoError(t, result.Error)

		// 验证期望
		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("use client.ExpectBegin/Commit/Close() methods", func(t *testing.T) {
		// 测试事务方法
		client.ExpectBegin()
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))
		client.ExpectCommit()

		// 执行事务
		tx := client.NewScoop().Begin()
		user := &TestUser{Name: "Test", Email: "test@example.com", Age: 25}
		result := tx.Model(TestUser{}).Create(user)
		assert.NoError(t, result.Error)
		tx.Commit()

		// 验证期望
		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

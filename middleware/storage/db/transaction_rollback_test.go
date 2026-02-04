package db_test

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestTransaction_RollbackOnError 测试事务在错误时回滚
func TestTransaction_RollbackOnError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("rollback on create error", func(t *testing.T) {
		// 开始事务
		client.ExpectBegin()

		// 创建失败
		client.ExpectExec("INSERT INTO test_users").
			WillReturnError(errors.New("duplicate key error"))

		// 回滚事务
		client.ExpectRollback()

		// 执行事务
		tx := client.NewScoop().Begin()
		user := &TestUser{Name: "Test", Email: "test@example.com", Age: 25}
		result := tx.Model(TestUser{}).Create(user)
		assert.Error(t, result.Error)

		// 手动回滚
		tx.Rollback()

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("rollback on update error", func(t *testing.T) {
		client.ExpectBegin()

		// 更新失败
		client.ExpectExec("UPDATE test_users.*").
			WillReturnError(errors.New("connection lost"))

		client.ExpectRollback()

		tx := client.NewScoop().Begin()
		result := tx.Model(TestUser{}).Where("id", 1).Updates("name", "Updated")
		assert.Error(t, result.Error)

		tx.Rollback()

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("rollback on delete error", func(t *testing.T) {
		client.ExpectBegin()

		// 删除失败
		client.ExpectExec("UPDATE test_users.*").
			WillReturnError(errors.New("constraint violation"))

		client.ExpectRollback()

		tx := client.NewScoop().Begin()
		result := tx.Model(TestUser{}).Where("id", 1).Delete()
		assert.Error(t, result.Error)

		tx.Rollback()

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestTransaction_MultiOperationWithRollback 测试多操作事务中的回滚
func TestTransaction_MultiOperationWithRollback(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("rollback after partial success", func(t *testing.T) {
		client.ExpectBegin()

		// 第一个操作成功
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 第二个操作失败
		client.ExpectExec("UPDATE test_users.*").
			WillReturnError(errors.New("foreign key constraint"))

		client.ExpectRollback()

		tx := client.NewScoop().Begin()

		// 创建用户（成功）
		user1 := &TestUser{Name: "User1", Email: "user1@example.com", Age: 25}
		result1 := tx.Model(TestUser{}).Create(user1)
		assert.NoError(t, result1.Error)

		// 更新用户（失败）
		result2 := tx.Model(TestUser{}).Where("id", 2).Updates("name", "Updated")
		assert.Error(t, result2.Error)

		// 回滚整个事务
		tx.Rollback()

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("rollback on query error in transaction", func(t *testing.T) {
		client.ExpectBegin()

		// 查询失败
		client.ExpectQuery("SELECT \\* FROM test_users.*").
			WillReturnError(errors.New("table not found"))

		client.ExpectRollback()

		tx := client.NewScoop().Begin()

		var users []TestUser
		result := tx.Model(TestUser{}).Find(&users)
		assert.Error(t, result.Error)

		tx.Rollback()

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestTransaction_CommitSuccess 测试成功提交事务
func TestTransaction_CommitSuccess(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("commit after successful operations", func(t *testing.T) {
		client.ExpectBegin()

		// 创建操作
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 更新操作
		client.ExpectExec("UPDATE test_users.*").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 提交事务
		client.ExpectCommit()

		tx := client.NewScoop().Begin()

		// 创建用户
		user := &TestUser{Name: "User", Email: "user@example.com", Age: 30}
		result1 := tx.Model(TestUser{}).Create(user)
		assert.NoError(t, result1.Error)

		// 更新用户
		result2 := tx.Model(TestUser{}).Where("id", 1).Updates("age", 31)
		assert.NoError(t, result2.Error)

		// 提交事务
		tx.Commit()

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestTransaction_NestedOperations 测试事务中的嵌套操作
func TestTransaction_NestedOperations(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("multiple creates in transaction", func(t *testing.T) {
		client.ExpectBegin()

		// 第一次创建
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 第二次创建
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(2, 1))

		// 第三次创建
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(3, 1))

		client.ExpectCommit()

		tx := client.NewScoop().Begin()

		// 创建多个用户
		users := []*TestUser{
			{Name: "User1", Email: "user1@example.com", Age: 25},
			{Name: "User2", Email: "user2@example.com", Age: 26},
			{Name: "User3", Email: "user3@example.com", Age: 27},
		}

		for _, user := range users {
			result := tx.Model(TestUser{}).Create(user)
			assert.NoError(t, result.Error)
		}

		tx.Commit()

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("create and query in transaction", func(t *testing.T) {
		client.ExpectBegin()

		// 创建
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 查询 - Where 条件会包裹在括号中
		client.ExpectQuery("SELECT \\* FROM test_users WHERE \\(id = 1\\) AND deleted_at = 0.*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}).
				AddRow(int64(1), "User1", "user1@example.com", int64(25)))

		client.ExpectCommit()

		tx := client.NewScoop().Begin()

		// 创建用户
		user := &TestUser{Name: "User1", Email: "user1@example.com", Age: 25}
		result1 := tx.Model(TestUser{}).Create(user)
		assert.NoError(t, result1.Error)

		// 查询用户
		var found TestUser
		result2 := tx.Model(TestUser{}).Where("id", 1).First(&found)
		assert.NoError(t, result2.Error)
		assert.Equal(t, 1, found.Id)

		tx.Commit()

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestTransaction_BeginError 测试 Begin 失败的情况
func TestTransaction_BeginError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("begin fails due to connection error", func(t *testing.T) {
		// Begin 失败
		client.ExpectBegin().WillReturnError(errors.New("connection timeout"))

		// Begin 返回的 Scoop 仍然可用，但底层的 transaction 会是 nil
		tx := client.NewScoop().Begin()
		assert.NotNil(t, tx) // Begin 总是返回一个 Scoop，不会返回 nil

		// 注意：由于 transaction 是 nil，实际操作会 panic
		// 这个测试主要验证 ExpectBegin.WillReturnError 的设置
		// 在实际使用中，应该检查 Begin 的返回

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

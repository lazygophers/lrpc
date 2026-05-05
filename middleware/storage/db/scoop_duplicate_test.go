package db_test

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/stretchr/testify/assert"
)

// TestScoop_Create_DuplicateKeyError 测试 Create 方法处理重复键错误
func TestScoop_Create_DuplicateKeyError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("Create with duplicate key error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnError(xerror.New(xerror.ErrConflict, "duplicate key error"))

		scoop := client.NewScoop().Table("test_users")
		user := &TestUser{Name: "Test", Email: "test@example.com", Age: 25}
		result := scoop.Create(user)

		// 应该有错误
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "duplicate key error")

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})
}

// TestScoop_IsDuplicatedKeyError 测试 IsDuplicatedKeyError 方法
func TestScoop_IsDuplicatedKeyError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("IsDuplicatedKeyError with ErrConflict", func(t *testing.T) {
		scoop := client.NewScoop()
		err := xerror.New(xerror.ErrConflict, "duplicate key")
		assert.True(t, scoop.IsDuplicatedKeyError(err))
	})

	t.Run("IsDuplicatedKeyError with other error", func(t *testing.T) {
		scoop := client.NewScoop()
		err := errors.New("some other error")
		assert.False(t, scoop.IsDuplicatedKeyError(err))
	})

	t.Run("IsDuplicatedKeyError with nil", func(t *testing.T) {
		scoop := client.NewScoop()
		assert.False(t, scoop.IsDuplicatedKeyError(nil))
	})
}

// TestScoop_Update_DuplicateKeyError 测试 Update 方法处理重复键错误
func TestScoop_Update_DuplicateKeyError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("Updates with duplicate key error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("UPDATE test_users.*").
			WillReturnError(xerror.New(xerror.ErrConflict, "duplicate key error"))

		scoop := client.NewScoop().Table("test_users")
		result := scoop.Where("id", 1).Updates("name", "Updated")

		// 应该有错误
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "duplicate key error")

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})
}

// TestScoop_CreateInBatches_DuplicateKeyError 测试批量创建时的重复键错误
func TestScoop_CreateInBatches_DuplicateKeyError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("CreateInBatches with duplicate key error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnError(xerror.New(xerror.ErrConflict, "duplicate key error"))

		scoop := client.NewScoop().Table("test_users")
		users := []*TestUser{
			{Name: "User1", Email: "user1@example.com", Age: 25},
			{Name: "User2", Email: "user2@example.com", Age: 26},
		}
		result := scoop.CreateInBatches(users, 50)

		// 应该有错误
		assert.Error(t, result.Error)

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})
}

// TestScoop_Create_TableName 测试使用 Table 方法的 Create
func TestScoop_Create_TableName(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("Create with Table method", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO custom_table").
			WillReturnResult(sqlmock.NewResult(1, 1))

		scoop := client.NewScoop().Table("custom_table")
		user := &TestUser{Name: "Test", Email: "test@example.com", Age: 25}
		result := scoop.Create(user)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})
}

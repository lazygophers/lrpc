package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_Create 测试 Create 方法
func TestScoop_Create(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer func() {
		mockDB.Mock.ExpectClose()
		mockDB.Close()
	}()

	t.Run("create with pointer to struct", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		result := client.NewScoop().Model(TestUser{}).Create(&TestUser{
			Name:  "Test User",
			Email: "test@example.com",
			Age:   25,
		})
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)
	})

	t.Run("create with error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnError(assert.AnError)

		result := client.NewScoop().Model(TestUser{}).Create(&TestUser{
			Name:  "Error User",
			Email: "error@example.com",
		})
		assert.Error(t, result.Error)
	})
}

// TestScoop_CreateInBatches 测试批量创建
func TestScoop_CreateInBatches(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer func() {
		mockDB.Mock.ExpectClose()
		mockDB.Close()
	}()

	t.Run("create in batches with small batch", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(0, 5))

		users := make([]*TestUser, 5)
		for i := 0; i < 5; i++ {
			users[i] = &TestUser{
				Name:  "User",
				Email: "user@example.com",
				Age:   i,
			}
		}

		result := client.NewScoop().Model(TestUser{}).CreateInBatches(users, 50)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(5), result.RowsAffected)
	})

	t.Run("create in batches with empty slice", func(t *testing.T) {
		users := []*TestUser{}
		result := client.NewScoop().Model(TestUser{}).CreateInBatches(users, 50)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(0), result.RowsAffected)
	})

	t.Run("create in batches with error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnError(assert.AnError)

		users := []*TestUser{
			{Name: "User", Email: "user@example.com", Age: 1},
		}
		result := client.NewScoop().Model(TestUser{}).CreateInBatches(users, 50)
		assert.Error(t, result.Error)
	})
}

// TestScoop_Create_AutoMigrate 测试 Scoop.AutoMigrate 方法
func TestScoop_Create_AutoMigrate(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("AutoMigrate with Scoop", func(t *testing.T) {
		// 测试 Scoop.AutoMigrate 方法
		scoop := client.NewScoop()
		err := scoop.AutoMigrate(TestUser{})
		// Mock 模式下可能失败，但不应该 panic
		_ = err
	})
}

// TestScoop_Create_DuplicateKey 测试创建时的重复键错误处理
func TestScoop_Create_DuplicateKey(t *testing.T) {
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
			WillReturnError(&mysql.MySQLError{
				Number: 1062, // ER_DUP_ENTRY
				Message: "Duplicate entry",
			})

		user := &TestUser{Name: "Test", Email: "test@example.com", Age: 25}
		result := client.NewScoop().Create(user)
		// 应该有错误但不应该 panic
		assert.NotNil(t, result.Error)

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})
}

// TestScoop_CreateWithModel 测试使用 Model 的 Create
func TestScoop_CreateWithModel(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("Create with Model", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		user := &TestUser{Name: "Test", Email: "test@example.com", Age: 25}
		scoop := client.NewScoop().Model(user)
		result := scoop.Create(user)
		assert.NoError(t, result.Error)

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})

	t.Run("Create with nil", func(t *testing.T) {
		result := client.NewScoop().Create(nil)
		// 应该有错误但不应该 panic
		assert.Error(t, result.Error)
	})
}

package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

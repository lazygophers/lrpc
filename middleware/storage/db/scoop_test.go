package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_MockBasicOperations 测试 Scoop 的基本操作（使用 Mock）
func TestScoop_MockBasicOperations(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, mockDB, err := db.NewMock(config)
	assert.NoError(t, err)

	t.Run("test Count operation", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM test_users").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		count, err := client.NewScoop().Model(TestUser{}).Count()
		assert.NoError(t, err)
		assert.Equal(t, uint64(10), count)
	})

	t.Run("test Exist operation", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT id FROM test_users WHERE deleted_at = 0 LIMIT 1 OFFSET 0").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		exist, err := client.NewScoop().Model(TestUser{}).Exist()
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	mockDB.Mock.ExpectClose()
	err = mockDB.Close()
	assert.NoError(t, err)

	err = mockDB.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestClient_PingAndClose 测试 Client 的 Ping 和 Close 方法
func TestClient_PingAndClose(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, mockDB, err := db.NewMock(config)
	assert.NoError(t, err)

	t.Run("test Ping", func(t *testing.T) {
		mockDB.Mock.ExpectPing()
		err := client.Ping()
		assert.NoError(t, err)
	})

	t.Run("test Close", func(t *testing.T) {
		mockDB.Mock.ExpectClose()
		err := client.Close()
		assert.NoError(t, err)
	})

	err = mockDB.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestScoop_Conditions 测试各种查询条件
func TestScoop_Conditions(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, mockDB, err := db.NewMock(config)
	assert.NoError(t, err)

	t.Run("test NotEqual condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).NotEqual("age", 25)
		assert.NotNil(t, scoop)
	})

	t.Run("test NotIn condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).NotIn("id", []int{1, 2, 3})
		assert.NotNil(t, scoop)
	})

	t.Run("test LeftLike condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).LeftLike("name", "Alice")
		assert.NotNil(t, scoop)
	})

	t.Run("test RightLike condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).RightLike("name", "Alice")
		assert.NotNil(t, scoop)
	})

	t.Run("test NotLike condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).NotLike("name", "Alice")
		assert.NotNil(t, scoop)
	})

	t.Run("test NotBetween condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).NotBetween("age", 20, 30)
		assert.NotNil(t, scoop)
	})

	t.Run("test Unscoped", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Unscoped()
		assert.NotNil(t, scoop)
	})

	t.Run("test Limit and Offset", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Limit(10).Offset(5)
		assert.NotNil(t, scoop)
	})

	t.Run("test Group", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Group("age")
		assert.NotNil(t, scoop)
	})

	t.Run("test Order", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Order("age DESC")
		assert.NotNil(t, scoop)
	})

	t.Run("test Select", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Select("id", "name")
		assert.NotNil(t, scoop)
	})

	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestScoop_Joins 测试 JOIN 操作
func TestScoop_Joins(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, mockDB, err := db.NewMock(config)
	assert.NoError(t, err)

	t.Run("test InnerJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			InnerJoin("orders", "users.id = orders.user_id")
		assert.NotNil(t, scoop)
	})

	t.Run("test LeftJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			LeftJoin("orders", "users.id = orders.user_id")
		assert.NotNil(t, scoop)
	})

	t.Run("test RightJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			RightJoin("orders", "users.id = orders.user_id")
		assert.NotNil(t, scoop)
	})

	t.Run("test FullJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			FullJoin("orders", "users.id = orders.user_id")
		assert.NotNil(t, scoop)
	})

	t.Run("test CrossJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			CrossJoin("orders")
		assert.NotNil(t, scoop)
	})

	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestScoop_Having 测试 HAVING 子句
func TestScoop_Having(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, mockDB, err := db.NewMock(config)
	assert.NoError(t, err)

	scoop := client.NewScoop().Model(TestUser{}).
		Group("age").
		Having("COUNT(*) > ?", 5)
	assert.NotNil(t, scoop)

	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestScoop_Ignore 测试 Ignore 方法
func TestScoop_Ignore(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, mockDB, err := db.NewMock(config)
	assert.NoError(t, err)

	t.Run("test Ignore with default", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Ignore()
		assert.NotNil(t, scoop)
	})

	t.Run("test Ignore with true", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Ignore(true)
		assert.NotNil(t, scoop)
	})

	t.Run("test Ignore with false", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Ignore(false)
		assert.NotNil(t, scoop)
	})

	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

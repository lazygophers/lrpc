package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_Updates 测试 Updates 方法
func TestScoop_Updates(t *testing.T) {
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

	t.Run("updates with single field", func(t *testing.T) {
		mockDB.Mock.ExpectExec("UPDATE test_users SET .* WHERE .* AND deleted_at = 0").
			WillReturnResult(sqlmock.NewResult(0, 1))

		result := client.NewScoop().Model(TestUser{}).Where("id", 1).Updates("name", "Updated Name")
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)
	})

	t.Run("updates with multiple fields", func(t *testing.T) {
		mockDB.Mock.ExpectExec("UPDATE test_users SET .* WHERE .* AND deleted_at = 0").
			WillReturnResult(sqlmock.NewResult(0, 1))

		result := client.NewScoop().Model(TestUser{}).Where("id", 1).
			Updates("name", "Updated Name", "age", 30)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)
	})

	t.Run("updates with error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("UPDATE test_users SET .* WHERE .* AND deleted_at = 0").
			WillReturnError(assert.AnError)

		result := client.NewScoop().Model(TestUser{}).Where("id", 999).
			Updates("name", "Error Name")
		assert.Error(t, result.Error)
	})

	t.Run("updates with map", func(t *testing.T) {
		mockDB.Mock.ExpectExec("UPDATE test_users SET .* WHERE .* AND deleted_at = 0").
			WillReturnResult(sqlmock.NewResult(0, 1))

		result := client.NewScoop().Model(TestUser{}).Where("id", 1).
			Updates(map[string]interface{}{
				"name": "Map Updated",
				"age":  25,
			})
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)
	})
}

func TestScoop_Updates_Variadic(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Updates with single key-value pair", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		result := scoop.Where("id", 1).Updates("name", "updated")
		_ = result
	})

	t.Run("Updates with multiple key-value pairs", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		result := scoop.Where("id", 1).Updates("name", "updated", "age", 30)
		_ = result
	})

	t.Run("Updates with odd number of args", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		result := scoop.Where("id", 1).Updates("name", "updated", "age")
		assert.Error(t, result.Error)
	})

	t.Run("Updates with non-string key", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		result := scoop.Where("id", 1).Updates(123, "value")
		assert.Error(t, result.Error)
	})
}

// TestScoop_Updates_Map 测试 Updates 使用 map
func TestScoop_Updates_Map(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Updates with map", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		result := scoop.Where("id", 1).Updates(map[string]interface{}{
			"name": "updated",
			"age":  30,
		})
		_ = result
	})

	t.Run("Updates with empty map", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		result := scoop.Where("id", 1).Updates(map[string]interface{}{})
		assert.Error(t, result.Error)
	})
}

// TestScoop_Updates_Struct 测试 Updates 使用 struct
func TestScoop_Updates_Struct(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Updates with struct pointer", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		user := &TestUser{Name: "updated", Age: 30}
		result := scoop.Where("id", 1).Updates(user)
		_ = result
	})

	t.Run("Updates with struct value", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		user := TestUser{Name: "updated", Age: 30}
		result := scoop.Where("id", 1).Updates(user)
		_ = result
	})

	t.Run("Updates with non-struct", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		result := scoop.Where("id", 1).Updates("string value")
		assert.Error(t, result.Error)
	})
}

package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_Updates_Variadic 测试 Updates 方法的不同调用方式
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

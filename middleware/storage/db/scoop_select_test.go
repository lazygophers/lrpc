package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_Select 测试 Select 方法
func TestScoop_Select(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Select with multiple fields", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Select("id", "name", "email")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Select with single field", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Select("id")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Chaining Select calls", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Select("id").Select("name")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Limit 测试 Limit 方法
func TestScoop_Limit(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Limit with positive value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Limit(10)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Limit with zero value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Limit(0)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Offset 测试 Offset 方法
func TestScoop_Offset(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Offset with positive value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Offset(5)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Offset with zero value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Offset(0)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

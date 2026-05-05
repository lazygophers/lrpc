package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_Equal 测试 Equal 方法
func TestScoop_Equal(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Equal with int value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Equal("id", 1)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Equal with string value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Equal("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Equal with nil value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Equal("deleted_at", nil)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Chaining Equal calls", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Equal("id", 1).Equal("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotEqual 测试 NotEqual 方法
func TestScoop_NotEqual(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotEqual with int value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotEqual("status", 0)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotEqual with string value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotEqual("name", "deleted")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

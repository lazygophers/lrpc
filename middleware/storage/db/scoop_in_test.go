package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_In 测试 In 方法
func TestScoop_In(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("In with slice", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.In("id", []int{1, 2, 3})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("In with empty slice", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.In("id", []int{})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotIn 测试 NotIn 方法
func TestScoop_NotIn(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotIn with slice", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotIn("id", []int{1, 2, 3})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotIn with empty slice", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotIn("id", []int{})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Between 测试 Between 方法
func TestScoop_Between(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Between with int values", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Between("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Between with string values", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Between("name", "A", "Z")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_NotBetween 测试 NotBetween 方法
func TestScoop_NotBetween(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotBetween with int values", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotBetween("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotBetween with string values", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotBetween("name", "A", "Z")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotLeftLike 测试 NotLeftLike 方法
func TestScoop_NotLeftLike(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotLeftLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotLeftLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotRightLike 测试 NotRightLike 方法
func TestScoop_NotRightLike(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotRightLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotRightLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

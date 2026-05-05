package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelScoop_LimitOffset 测试 Limit 和 Offset 方法
func TestModelScoop_LimitOffset(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Limit with value", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Limit(10)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Offset with value", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Offset(5)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Chaining Limit and Offset", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Limit(10).Offset(5)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_Select 测试 Select 方法
func TestModelScoop_Select(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Select with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Select("id", "name", "email")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Chaining Select calls", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Select("id").Select("name")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_Equal 测试 Equal 方法
func TestModelScoop_Equal(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Equal with value", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Equal("id", 1)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_GroupOrder 测试 Group 和 Order 方法
func TestModelScoop_GroupOrder(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Group with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Group("status")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Order with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Order("id", "created_at")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Desc with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Desc("id", "created_at")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Asc with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Asc("id", "name")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_Unscoped 测试 Unscoped 方法
func TestModelScoop_Unscoped(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Unscoped without argument", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Unscoped()
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Unscoped with true", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Unscoped(true)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Unscoped with false", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Unscoped(false)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_Ignore 测试 Ignore 方法
func TestModelScoop_Ignore(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Ignore without argument", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Ignore()
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Ignore with true", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Ignore(true)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Ignore with false", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Ignore(false)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

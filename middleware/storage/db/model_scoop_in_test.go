package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelScoop_InNotIn 测试 In 和 NotIn 方法
func TestModelScoop_InNotIn(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("In with slice", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.In("id", []int{1, 2, 3})
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("In with empty slice", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.In("id", []int{})
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("NotIn with slice", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.NotIn("id", []int{1, 2, 3})
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("NotIn with empty slice", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.NotIn("id", []int{})
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_LikeMethods 测试 Like 相关方法
func TestModelScoop_LikeMethods(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Like", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Like("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("LeftLike", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.LeftLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("RightLike", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.RightLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("NotLike", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.NotLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("NotLeftLike", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.NotLeftLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("NotRightLike", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.NotRightLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_Between 测试 Between 和 NotBetween 方法
func TestModelScoop_Between(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Between", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Between("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("NotBetween", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.NotBetween("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_NotEqual 测试 NotEqual 方法
func TestModelScoop_NotEqual(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotEqual with value", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.NotEqual("status", 0)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_WhereOr 测试 Where 和 Or 方法
func TestModelScoop_WhereOr(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Where with args", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Where("id", 1)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Or with args", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Or("status", 1)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

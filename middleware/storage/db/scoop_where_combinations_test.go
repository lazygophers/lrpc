package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_Where_Combinations 测试 Where 方法组合
func TestScoop_Where_Combinations(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Where with map", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Where(map[string]interface{}{
			"name": "test",
			"age":  25,
		})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Where with multiple args", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Where("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotEqual_NotIn 测试 NotEqual 和 NotIn 组合
func TestScoop_NotEqual_NotIn(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotEqual with different types", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotEqual("status", 0).NotEqual("deleted", 1)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotIn with multiple slices", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotIn("id", []int{1, 2}).NotIn("status", []int{0, 1})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Like_Between 测试 Like 和 Between 组合
func TestScoop_Like_Between(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Like and Between combination", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Like("name", "test").Between("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotLike and NotBetween combination", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotLike("name", "test").NotBetween("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_In_Like_Mixed 测试混合条件
func TestScoop_In_Like_Mixed(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("In and Like combination", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.In("id", []int{1, 2, 3}).Like("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotIn and NotLike combination", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotIn("id", []int{1, 2, 3}).NotLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Where_SubConditions 测试 Where 子条件
func TestScoop_Where_SubConditions(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Where with operator", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Where("age", ">", 18)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Where with multiple operators", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Where("age", ">", 18).Where("age", "<", 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

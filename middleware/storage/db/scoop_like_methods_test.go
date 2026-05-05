package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_LikeMethods 测试 Like 相关方法
func TestScoop_LikeMethods(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Like with normal value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Like("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("LeftLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.LeftLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("RightLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.RightLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

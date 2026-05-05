package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_Where 测试 Where 方法
func TestScoop_Where(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("Where with single condition", func(t *testing.T) {
		scoop := client.NewScoop().Where("id", 1)
		assert.NotNil(t, scoop)
	})

	t.Run("Where with multiple conditions", func(t *testing.T) {
		scoop := client.NewScoop().Where("name", "test").Where("age", ">", 18)
		assert.NotNil(t, scoop)
	})

	t.Run("Where with map", func(t *testing.T) {
		scoop := client.NewScoop().Where(map[string]interface{}{
			"name": "test",
			"age":  25,
		})
		assert.NotNil(t, scoop)
	})
}

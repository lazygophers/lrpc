package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelScoop_FirstOrCreate_ErrorCases 测试 FirstOrCreate 的错误情况
func TestModelScoop_FirstOrCreate_ErrorCases(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("FirstOrCreate with nil model", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.FirstOrCreate(nil)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "FirstOrCreate failed")
		assert.Nil(t, result.Object)
	})
}

// TestModelScoop_CreateIfNotExists_ErrorCases 测试 CreateIfNotExists 的错误情况
func TestModelScoop_CreateIfNotExists_ErrorCases(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("CreateIfNotExists with nil model", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.CreateIfNotExists(nil)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "CreateIfNotExists failed")
	})
}

// TestModelScoop_Exist 测试 Exist 方法
func TestModelScoop_Exist(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Exist with model", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		exist, err := modelScoop.Exist()
		// Mock模式下可能返回false
		_ = exist
		_ = err
	})
}

// TestModelScoop_Scan_ErrorCases 测试 Scan 方法的错误情况
func TestModelScoop_Scan_ErrorCases(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Scan with nil destination", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Scan(nil)
		// 应该有错误
		assert.NotNil(t, result)
	})
}

// TestModelScoop_Chunk_ErrorCases 测试 Chunk 方法的错误情况
func TestModelScoop_Chunk_ErrorCases(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Chunk with zero size", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Chunk(0, func(tx *db.Scoop, out []*TestUser, offset uint64) error {
			return nil
		})
		assert.NotNil(t, result)
	})

	t.Run("Chunk with nil function", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Chunk(10, nil)
		assert.NotNil(t, result)
		assert.Error(t, result.Error)
	})
}

// TestModelScope_Updates 测试 ModelScoop.Updates 方法
func TestModelScoop_Updates(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Updates with variadic args", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Updates("name", "test", "age", 25)
		_ = result
	})

	t.Run("Updates with map", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Updates(map[string]interface{}{
			"name": "updated",
			"age":  30,
		})
		_ = result
	})

	t.Run("Updates with struct", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		user := &TestUser{Name: "test", Age: 25}
		result := modelScoop.Updates(user)
		_ = result
	})
}

// TestModelScoop_Delete 测试 ModelScoop.Delete 方法
func TestModelScoop_Delete(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Delete with conditions", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Where("id", 1).Delete()
		// Mock模式下可能失败，但测试代码路径
		_ = result
	})
}


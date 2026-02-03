package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestModel_SetNotFound 测试设置 NotFound 错误
func TestModel_SetNotFound(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	customErr := gorm.ErrInvalidData

	result := model.SetNotFound(customErr)
	assert.Equal(t, model, result)
	assert.Equal(t, customErr, model.notFoundError)
}

// TestModel_SetDuplicatedKeyError 测试设置重复键错误
func TestModel_SetDuplicatedKeyError(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	customErr := gorm.ErrInvalidValue

	result := model.SetDuplicatedKeyError(customErr)
	assert.Equal(t, model, result)
	assert.Equal(t, customErr, model.duplicatedKeyError)
}

// TestModel_IsNotFound 测试判断是否为 NotFound 错误
func TestModel_IsNotFound(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "gorm.ErrRecordNotFound",
			err:      gorm.ErrRecordNotFound,
			expected: true,
		},
		{
			name:     "custom not found error",
			err:      model.notFoundError,
			expected: true,
		},
		{
			name:     "other error",
			err:      gorm.ErrInvalidData,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.IsNotFound(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestModel_IsDuplicatedKeyError 测试判断是否为重复键错误
func TestModel_IsDuplicatedKeyError(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "gorm.ErrDuplicatedKey",
			err:      gorm.ErrDuplicatedKey,
			expected: true,
		},
		{
			name:     "custom duplicated key error",
			err:      model.duplicatedKeyError,
			expected: true,
		},
		{
			name:     "other error",
			err:      gorm.ErrInvalidData,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.IsDuplicatedKeyError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestModel_TableName 测试获取表名
func TestModel_TableName(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)
	tableName := model.TableName()
	assert.Equal(t, "test_users", tableName)
}

// TestModel_NewScoop 测试创建 ModelScoop
func TestModel_NewScoop(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("create scoop without transaction", func(t *testing.T) {
		scoop := model.NewScoop()
		assert.NotNil(t, scoop)
		assert.Equal(t, model.table, scoop.table)
		assert.Equal(t, model.hasId, scoop.hasId)
		assert.Equal(t, model.hasCreatedAt, scoop.hasCreatedAt)
		assert.Equal(t, model.hasUpdatedAt, scoop.hasUpdatedAt)
		assert.Equal(t, model.hasDeletedAt, scoop.hasDeletedAt)
		assert.Equal(t, model.notFoundError, scoop.notFoundError)
		assert.Equal(t, model.duplicatedKeyError, scoop.duplicatedKeyError)
	})

	t.Run("create scoop with transaction", func(t *testing.T) {
		mockDB.Mock.ExpectBegin()
		tx := client.NewScoop().Begin()
		assert.NotNil(t, tx)

		scoop := model.NewScoop(tx)
		assert.NotNil(t, scoop)
		assert.Equal(t, model.table, scoop.table)
	})
}

// TestNewModel 测试创建 Model
func TestNewModel(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("create model with struct", func(t *testing.T) {
		model := NewModel[TestUser](client)
		assert.NotNil(t, model)
		assert.Equal(t, client, model.db)
		assert.Equal(t, gorm.ErrRecordNotFound, model.notFoundError)
		assert.Equal(t, gorm.ErrDuplicatedKey, model.duplicatedKeyError)
		assert.Equal(t, "test_users", model.table)
	})

	t.Run("create model with pointer", func(t *testing.T) {
		type TestModel struct {
			ID   int64  `gorm:"column:id;primaryKey"`
			Name string `gorm:"column:name"`
		}

		model := NewModel[*TestModel](client)
		assert.NotNil(t, model)
	})
}

// TestModel_ErrorHandling 测试 Model 的错误处理
func TestModel_ErrorHandling(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	model := NewModel[TestUser](client)

	t.Run("chain error setters", func(t *testing.T) {
		customNotFound := gorm.ErrInvalidData
		customDuplicated := gorm.ErrInvalidValue

		result := model.
			SetNotFound(customNotFound).
			SetDuplicatedKeyError(customDuplicated)

		assert.Equal(t, model, result)
		assert.True(t, model.IsNotFound(customNotFound))
		assert.True(t, model.IsDuplicatedKeyError(customDuplicated))
	})
}

// TestModel_WithDifferentTypes 测试不同类型的 Model
func TestModel_WithDifferentTypes(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("model without timestamps", func(t *testing.T) {
		type SimpleModel struct {
			ID   int64  `gorm:"column:id;primaryKey"`
			Name string `gorm:"column:name"`
		}

		model := NewModel[SimpleModel](client)
		assert.NotNil(t, model)
	})

	t.Run("model without id", func(t *testing.T) {
		type NoIDModel struct {
			Name string `gorm:"column:name"`
		}

		model := NewModel[NoIDModel](client)
		assert.NotNil(t, model)
		assert.False(t, model.hasId)
	})
}

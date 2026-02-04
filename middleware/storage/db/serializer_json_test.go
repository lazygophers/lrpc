package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelWithJSON 测试使用 JsonSerializer 的模型
type TestModelWithJSON struct {
	Id      int            `gorm:"primaryKey;autoIncrement"`
	Name    string         `gorm:"size:100;not null"`
	Config  map[string]any `gorm:"column:config;type:text;serializer:json;not null"`
	Created int64          `gorm:"autoCreateTime:milli"`
	Updated int64          `gorm:"autoUpdateTime:milli"`
}

func (TestModelWithJSON) TableName() string {
	return "test_json_models"
}

// TestModelWithJSON2 测试 JSON 序列化器的 Scan 方法
type TestModelWithJSON2 struct {
	Id     int      `gorm:"primaryKey;autoIncrement"`
	Name   string   `gorm:"size:100;not null"`
	Tags   []string `gorm:"column:tags;type:text;serializer:json;not null"`
	Created int64    `gorm:"autoCreateTime:milli"`
	Updated int64    `gorm:"autoUpdateTime:milli"`
}

func (TestModelWithJSON2) TableName() string {
	return "test_json_models2"
}

// TestJsonSerializer 测试 JsonSerializer 的 Scan 和 Value 方法
func TestJsonSerializer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("json serializer with map", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_serializer",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithJSON{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		model := &TestModelWithJSON{
			Name: "Test JSON",
			Config: map[string]any{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
		}

		// 测试 Create（使用 Value 方法）
		err = client.NewScoop().Model(TestModelWithJSON{}).Create(model).Error
		assert.NoError(t, err)
		assert.Greater(t, model.Id, 0)

		// 测试 First（使用 Scan 方法）
		var found TestModelWithJSON
		result := client.NewScoop().Model(TestModelWithJSON{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, model.Name, found.Name)
		assert.NotNil(t, found.Config)
		assert.Equal(t, "value1", found.Config["key1"])
		assert.Equal(t, float64(123), found.Config["key2"]) // JSON 数字解析为 float64
		assert.Equal(t, true, found.Config["key3"])
	})

	t.Run("json serializer with slice", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_tags_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_serializer2",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithJSON2{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		model := &TestModelWithJSON2{
			Name: "Test JSON Tags",
			Tags: []string{"tag1", "tag2", "tag3"},
		}

		// 测试 Create
		err = client.NewScoop().Model(TestModelWithJSON2{}).Create(model).Error
		assert.NoError(t, err)
		assert.Greater(t, model.Id, 0)

		// 测试 First
		var found TestModelWithJSON2
		result := client.NewScoop().Model(TestModelWithJSON2{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, model.Name, found.Name)
		assert.NotNil(t, found.Tags)
		assert.Len(t, found.Tags, 3)
		assert.Equal(t, "tag1", found.Tags[0])
		assert.Equal(t, "tag2", found.Tags[1])
		assert.Equal(t, "tag3", found.Tags[2])
	})

	t.Run("json serializer with nil value", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_nil_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_serializer3",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithJSON{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据，Config 为 nil
		model := &TestModelWithJSON{
			Name:   "Test JSON Nil",
			Config: nil,
		}

		// 测试 Create
		err = client.NewScoop().Model(TestModelWithJSON{}).Create(model).Error
		assert.NoError(t, err)
		assert.Greater(t, model.Id, 0)

		// 测试 First
		var found TestModelWithJSON
		result := client.NewScoop().Model(TestModelWithJSON{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, model.Name, found.Name)
	})

	t.Run("json serializer with empty map", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_empty_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_serializer4",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithJSON{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据，Config 为空 map
		model := &TestModelWithJSON{
			Name:   "Test JSON Empty",
			Config: map[string]any{},
		}

		// 测试 Create
		err = client.NewScoop().Model(TestModelWithJSON{}).Create(model).Error
		assert.NoError(t, err)
		assert.Greater(t, model.Id, 0)

		// 测试 First
		var found TestModelWithJSON
		result := client.NewScoop().Model(TestModelWithJSON{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, model.Name, found.Name)
		assert.NotNil(t, found.Config)
		assert.Empty(t, found.Config)
	})
}

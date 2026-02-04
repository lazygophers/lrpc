package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelWithBSONAdvanced 测试使用 BSON 序列化器的高级场景
type TestModelWithBSONAdvanced struct {
	Id       int                 `gorm:"primaryKey;autoIncrement"`
	Name     string              `gorm:"size:100;not null"`
	Metadata map[string]any      `gorm:"column:metadata;type:blob;serializer:bson;not null"`
	Config   map[string]any      `gorm:"column:config;type:blob;serializer:bson;not null"`
	Settings map[string]string  `gorm:"column:settings;type:blob;serializer:bson"`
	Counters  map[string]int64   `gorm:"column:counters;type:blob;serializer:bson"`
	Created  int64               `gorm:"autoCreateTime:milli"`
	Updated  int64               `gorm:"autoUpdateTime:milli"`
}

func (TestModelWithBSONAdvanced) TableName() string {
	return "test_bson_advanced_models"
}

// TestBsonSerializerAdvanced 测试 BSON 序列化器的高级场景
func TestBsonSerializerAdvanced(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("bson with nested map", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_bson_nested_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_bson_nested",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithBSONAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithBSONAdvanced{
			Name: "Test Nested BSON",
			Metadata: map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": int32(8080),
				},
				"features": []any{"feature1", "feature2"},
			},
		}

		err = client.NewScoop().Model(TestModelWithBSONAdvanced{}).Create(model).Error
		assert.NoError(t, err)
		assert.Greater(t, model.Id, 0)

		var found TestModelWithBSONAdvanced
		result := client.NewScoop().Model(TestModelWithBSONAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.Metadata)

		// 验证嵌套结构
		server, ok := found.Metadata["server"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "localhost", server["host"])
		assert.Equal(t, int32(8080), server["port"])
	})

	t.Run("bson with various data types", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_bson_types_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_bson_types",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithBSONAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithBSONAdvanced{
			Name: "Test Types BSON",
			Config: map[string]any{
				"string_val": "test",
				"int_val":    int32(42),
				"bool_val":   true,
				"float_val":  3.14,
				"null_val":   nil,
				"array_val":  []any{"a", "b", "c"},
			},
		}

		err = client.NewScoop().Model(TestModelWithBSONAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithBSONAdvanced
		result := client.NewScoop().Model(TestModelWithBSONAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)

		assert.Equal(t, "test", found.Config["string_val"])
		assert.Equal(t, int32(42), found.Config["int_val"])
		assert.Equal(t, true, found.Config["bool_val"])
		assert.NotNil(t, found.Config["float_val"])
		assert.Nil(t, found.Config["null_val"])
		assert.NotNil(t, found.Config["array_val"])
	})

	t.Run("bson with string map", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_bson_string_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_bson_string",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithBSONAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithBSONAdvanced{
			Name: "Test String BSON",
			Settings: map[string]string{
				"env":     "production",
				"region":  "us-west-2",
				"version": "2.0.0",
			},
		}

		err = client.NewScoop().Model(TestModelWithBSONAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithBSONAdvanced
		result := client.NewScoop().Model(TestModelWithBSONAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)

		assert.Equal(t, "production", found.Settings["env"])
		assert.Equal(t, "us-west-2", found.Settings["region"])
		assert.Equal(t, "2.0.0", found.Settings["version"])
	})

	t.Run("bson with int64 map", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_bson_int64_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_bson_int64",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithBSONAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithBSONAdvanced{
			Name: "Test Int64 BSON",
			Counters: map[string]int64{
				"requests":   10000,
				"errors":     50,
				"timeout_ms":  5000,
			},
		}

		err = client.NewScoop().Model(TestModelWithBSONAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithBSONAdvanced
		result := client.NewScoop().Model(TestModelWithBSONAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)

		assert.Equal(t, int64(10000), found.Counters["requests"])
		assert.Equal(t, int64(50), found.Counters["errors"])
		assert.Equal(t, int64(5000), found.Counters["timeout_ms"])
	})

	t.Run("bson update operations", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_bson_update_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_bson_update",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithBSONAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithBSONAdvanced{
			Name: "Test BSON Update",
			Metadata: map[string]any{
				"key1": "value1",
			},
		}

		err = client.NewScoop().Model(TestModelWithBSONAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		// 更新 Metadata
		model.Metadata = map[string]any{
			"key1": "updated_value1",
			"key2": "value2",
		}

		updateResult := client.NewScoop().Model(TestModelWithBSONAdvanced{}).Where("id", model.Id).Updates(model)
		assert.NoError(t, updateResult.Error)

		var found TestModelWithBSONAdvanced
		result := client.NewScoop().Model(TestModelWithBSONAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)

		assert.Equal(t, "updated_value1", found.Metadata["key1"])
		assert.Equal(t, "value2", found.Metadata["key2"])
	})

	t.Run("bson with empty and nil values", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_bson_empty_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_bson_empty",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithBSONAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithBSONAdvanced{
			Name:     "Test Empty BSON",
			Metadata: map[string]any{},
			Settings: map[string]string{},
			Counters: map[string]int64{},
		}

		err = client.NewScoop().Model(TestModelWithBSONAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithBSONAdvanced
		result := client.NewScoop().Model(TestModelWithBSONAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)

		assert.Empty(t, found.Metadata)
		assert.Empty(t, found.Settings)
		assert.Empty(t, found.Counters)
	})

	t.Run("bson with special characters", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_bson_special_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_bson_special",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithBSONAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithBSONAdvanced{
			Name: "Test Special BSON",
			Settings: map[string]string{
				"unicode": "测试中文字符",
				"quotes": "\"quoted\"",
				"mixed":  "value\nwith\tnewlines",
			},
		}

		err = client.NewScoop().Model(TestModelWithBSONAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithBSONAdvanced
		result := client.NewScoop().Model(TestModelWithBSONAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)

		assert.Equal(t, "测试中文字符", found.Settings["unicode"])
		assert.Equal(t, "\"quoted\"", found.Settings["quotes"])
		assert.Equal(t, "value\nwith\tnewlines", found.Settings["mixed"])
	})
}

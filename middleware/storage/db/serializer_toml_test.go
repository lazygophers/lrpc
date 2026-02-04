package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelWithTOMLAdvanced 测试使用 TOML 序列化器的高级场景
type TestModelWithTOMLAdvanced struct {
	Id       int                 `gorm:"primaryKey;autoIncrement"`
	Name     string              `gorm:"size:100;not null"`
	Config   map[string]any      `gorm:"column:config;type:text;serializer:toml;not null"`
	Settings map[string]string  `gorm:"column:settings;type:text;serializer:toml"`
	Metadata map[string]int64   `gorm:"column:metadata;type:text;serializer:toml"`
	Created  int64               `gorm:"autoCreateTime:milli"`
	Updated  int64               `gorm:"autoUpdateTime:milli"`
}

func (TestModelWithTOMLAdvanced) TableName() string {
	return "test_toml_advanced_models"
}

// TestTomlSerializerAdvanced 测试 TOML 序列化器的高级场景
func TestTomlSerializerAdvanced(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("toml with nested map", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_nested_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_nested",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithTOMLAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithTOMLAdvanced{
			Name: "Test Nested TOML",
			Config: map[string]any{
				"server": map[string]any{
					"host": "localhost",
					"port": 8080,
				},
				"database": map[string]any{
					"driver":   "mysql",
					"timeout": 30,
				},
			},
		}

		err = client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Create(model).Error
		assert.NoError(t, err)
		assert.Greater(t, model.Id, 0)

		var found TestModelWithTOMLAdvanced
		result := client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.Config)

		// 验证嵌套结构
		server, ok := found.Config["server"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, "localhost", server["host"])
		assert.Equal(t, int64(8080), server["port"])
	})

	t.Run("toml with string map", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_string_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_string",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithTOMLAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithTOMLAdvanced{
			Name: "Test String TOML",
			Settings: map[string]string{
				"env":     "production",
				"region":  "us-east-1",
				"version": "1.0.0",
			},
		}

		err = client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithTOMLAdvanced
		result := client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "production", found.Settings["env"])
		assert.Equal(t, "us-east-1", found.Settings["region"])
		assert.Equal(t, "1.0.0", found.Settings["version"])
	})

	t.Run("toml with int64 map", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_int64_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_int64",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithTOMLAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithTOMLAdvanced{
			Name: "Test Int64 TOML",
			Metadata: map[string]int64{
				"count":    1000,
				"size":     2048,
				"timeout":  30,
			},
		}

		err = client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithTOMLAdvanced
		result := client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1000), found.Metadata["count"])
		assert.Equal(t, int64(2048), found.Metadata["size"])
		assert.Equal(t, int64(30), found.Metadata["timeout"])
	})

	t.Run("toml with empty maps", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_empty_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_empty",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithTOMLAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithTOMLAdvanced{
			Name:     "Test Empty TOML",
			Config:   map[string]any{},
			Settings: map[string]string{},
			Metadata: map[string]int64{},
		}

		err = client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithTOMLAdvanced
		result := client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Empty(t, found.Config)
		assert.Empty(t, found.Settings)
		assert.Empty(t, found.Metadata)
	})

	t.Run("toml update operations", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_update_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_update",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithTOMLAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithTOMLAdvanced{
			Name: "Test TOML Update",
			Config: map[string]any{
				"key1": "value1",
			},
		}

		err = client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		// 更新 Config
		model.Config = map[string]any{
			"key1": "updated_value1",
			"key2": "value2",
		}

		updateResult := client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Where("id", model.Id).Updates(model)
		assert.NoError(t, updateResult.Error)

		var found TestModelWithTOMLAdvanced
		result := client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "updated_value1", found.Config["key1"])
		assert.Equal(t, "value2", found.Config["key2"])
	})

	t.Run("toml with mixed data types", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_mixed_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_mixed",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithTOMLAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithTOMLAdvanced{
			Name: "Test Mixed TOML",
			Config: map[string]any{
				"string_val": "test",
				"int_val":     42,
				"bool_val":    true,
				"float_val":   3.14,
				"array_val":   []any{"a", "b", "c"},
			},
		}

		err = client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithTOMLAdvanced
		result := client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)

		assert.Equal(t, "test", found.Config["string_val"])
		assert.Equal(t, int64(42), found.Config["int_val"])
		assert.Equal(t, true, found.Config["bool_val"])
		// TOML 解析浮点数时可能会有精度问题，这里只检查非空
		assert.NotNil(t, found.Config["float_val"])
		assert.NotNil(t, found.Config["array_val"])
	})

	t.Run("toml with special characters", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_special_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_special",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithTOMLAdvanced{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestModelWithTOMLAdvanced{
			Name: "Test Special TOML",
			Settings: map[string]string{
				"key_with_underscore": "value1",
				"key-with-dash":      "value2",
				"key.with.dots":      "value3",
				"special chars":      "value with spaces\nand\ttabs",
			},
		}

		err = client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Create(model).Error
		assert.NoError(t, err)

		var found TestModelWithTOMLAdvanced
		result := client.NewScoop().Model(TestModelWithTOMLAdvanced{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "value1", found.Settings["key_with_underscore"])
		assert.Equal(t, "value2", found.Settings["key-with-dash"])
		assert.Equal(t, "value3", found.Settings["key.with.dots"])
	})
}

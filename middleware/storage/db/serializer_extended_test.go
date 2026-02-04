package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestSerializerJsonWithArrays 测试 JSON 序列化器处理数组
type TestJsonArrayModel struct {
	Id    int      `gorm:"primaryKey;autoIncrement"`
	Name  string   `gorm:"size:100;not null"`
	Tags  []string `gorm:"column:tags;type:text;serializer:json;not null"`
	Ids   []int    `gorm:"column:ids;type:text;serializer:json;not null"`
}

func (TestJsonArrayModel) TableName() string {
	return "test_json_array_models"
}

// TestJsonSerializerArrays 测试 JSON 序列化器处理数组类型
func TestJsonSerializerArrays(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("json with string array", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_array_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_array",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonArrayModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestJsonArrayModel{
			Name: "Array Test",
			Tags: []string{"tag1", "tag2", "tag3"},
		}

		err = client.NewScoop().Model(TestJsonArrayModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestJsonArrayModel
		result := client.NewScoop().Model(TestJsonArrayModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Len(t, found.Tags, 3)
		assert.Equal(t, "tag1", found.Tags[0])
	})

	t.Run("json with int array", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_int_array_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_int_array",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonArrayModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestJsonArrayModel{
			Name: "Int Array Test",
			Ids:  []int{1, 2, 3, 4, 5},
		}

		err = client.NewScoop().Model(TestJsonArrayModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestJsonArrayModel
		result := client.NewScoop().Model(TestJsonArrayModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Len(t, found.Ids, 5)
		assert.Equal(t, 1, found.Ids[0])
	})

	t.Run("json with nil arrays", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_nil_array_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_nil_array",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonArrayModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestJsonArrayModel{
			Name: "Nil Array Test",
			Tags: nil,
			Ids:  nil,
		}

		err = client.NewScoop().Model(TestJsonArrayModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestJsonArrayModel
		result := client.NewScoop().Model(TestJsonArrayModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Nil(t, found.Tags)
		assert.Nil(t, found.Ids)
	})

	t.Run("json with empty arrays", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_empty_array_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_empty_array",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonArrayModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestJsonArrayModel{
			Name: "Empty Array Test",
			Tags: []string{},
			Ids:  []int{},
		}

		err = client.NewScoop().Model(TestJsonArrayModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestJsonArrayModel
		result := client.NewScoop().Model(TestJsonArrayModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Empty(t, found.Tags)
		assert.Empty(t, found.Ids)
	})
}

// TestSerializerYamlComplex 测试 YAML 序列化器处理复杂数据
type TestYamlComplexModel struct {
	Id       int            `gorm:"primaryKey;autoIncrement"`
	Name     string         `gorm:"size:100;not null"`
	Metadata map[string]any `gorm:"column:metadata;type:text;serializer:yaml;not null"`
}

func (TestYamlComplexModel) TableName() string {
	return "test_yaml_complex_models"
}

func TestYamlSerializerComplex(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("yaml with nested structures", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_yaml_nested_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_yaml_nested",
			Debug:   true,
		}

		client, err := db.New(config, TestYamlComplexModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestYamlComplexModel{
			Name: "Nested YAML Test",
			Metadata: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"value": "deep",
					},
				},
			},
		}

		err = client.NewScoop().Model(TestYamlComplexModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestYamlComplexModel
		result := client.NewScoop().Model(TestYamlComplexModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.Metadata)
	})

	t.Run("yaml with list values", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_yaml_list_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_yaml_list",
			Debug:   true,
		}

		client, err := db.New(config, TestYamlComplexModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestYamlComplexModel{
			Name: "List YAML Test",
			Metadata: map[string]any{
				"items": []any{"item1", "item2", "item3"},
			},
		}

		err = client.NewScoop().Model(TestYamlComplexModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestYamlComplexModel
		result := client.NewScoop().Model(TestYamlComplexModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.Metadata)
	})
}


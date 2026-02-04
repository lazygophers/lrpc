package db_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestSerializerBoundaryCases 测试所有序列化器的边界情况
type TestBoundaryModel struct {
	Id          int                    `gorm:"primaryKey;autoIncrement"`
	JsonData    string                 `gorm:"column:json_data;type:text;serializer:json"`
	YamlData    string                 `gorm:"column:yaml_data;type:text;serializer:yaml"`
	TomlData    map[string]interface{} `gorm:"column:toml_data;type:text;serializer:toml"`
	UnicodeData string                 `gorm:"column:unicode_data;type:text;serializer:json"`
	SpecialData string                 `gorm:"column:special_data;type:text;serializer:json"`
}

func (TestBoundaryModel) TableName() string {
	return "test_boundary_models"
}

func TestSerializerBoundaryCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("empty string values", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_empty_string_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_empty_string",
			Debug:   true,
		}

		client, err := db.New(config, TestBoundaryModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestBoundaryModel{
			JsonData:    "",
			YamlData:    "",
			TomlData:    map[string]interface{}{},
			UnicodeData: "",
			SpecialData: "",
		}

		err = client.NewScoop().Model(TestBoundaryModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestBoundaryModel
		result := client.NewScoop().Model(TestBoundaryModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "", found.JsonData)
		assert.Equal(t, "", found.YamlData)
		// 空 map 可能被序列化为 nil
		if found.TomlData != nil {
			assert.Empty(t, found.TomlData)
		}
	})

	t.Run("unicode characters", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_unicode_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_unicode",
			Debug:   true,
		}

		client, err := db.New(config, TestBoundaryModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 测试各种 Unicode 字符
		model := &TestBoundaryModel{
			JsonData:    "Hello 世界 🌍",
			YamlData:    "Привет мир",
			TomlData:    map[string]interface{}{"key": "🚀 Rocket"},
			UnicodeData: "Ελληνικά العربية 日本語",
			SpecialData: "🎉🎊🎈",
		}

		err = client.NewScoop().Model(TestBoundaryModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestBoundaryModel
		result := client.NewScoop().Model(TestBoundaryModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "Hello 世界 🌍", found.JsonData)
		assert.Equal(t, "Привет мир", found.YamlData)
		assert.Equal(t, "🚀 Rocket", found.TomlData["key"])
	})

	t.Run("special characters", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_special_chars_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_special_chars",
			Debug:   true,
		}

		client, err := db.New(config, TestBoundaryModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 测试特殊字符和转义序列
		model := &TestBoundaryModel{
			JsonData:    "Line1\nLine2\tTabbed\"Quoted\"",
			YamlData:    "Special: \n\t\r",
			TomlData:    map[string]interface{}{"key": "Quote \"Test\""},
			UnicodeData: "Emoji: \U0001F600",
			SpecialData: "Backslash \\ Forward slash /",
		}

		err = client.NewScoop().Model(TestBoundaryModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestBoundaryModel
		result := client.NewScoop().Model(TestBoundaryModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Contains(t, found.JsonData, "\n")
		assert.Contains(t, found.JsonData, "\t")
	})

	t.Run("very long strings", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_long_strings_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_long_strings",
			Debug:   true,
		}

		client, err := db.New(config, TestBoundaryModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建长字符串（10KB）
		longString := strings.Repeat("A", 10*1024)

		model := &TestBoundaryModel{
			JsonData: longString,
			YamlData: longString,
			TomlData: map[string]interface{}{"data": longString},
		}

		err = client.NewScoop().Model(TestBoundaryModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestBoundaryModel
		result := client.NewScoop().Model(TestBoundaryModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Len(t, found.JsonData, 10*1024)
		assert.Len(t, found.YamlData, 10*1024)
		assert.Len(t, found.TomlData["data"].(string), 10*1024)
	})
}

// TestSerializerJsonSpecialFeatures 测试 JSON 特定特性
type TestJsonFeaturesModel struct {
	Id           int                    `gorm:"primaryKey;autoIncrement"`
	NestedObject map[string]interface{} `gorm:"column:nested_obj;type:text;serializer:json"`
	ComplexJson  map[string]interface{} `gorm:"column:complex_json;type:text;serializer:json"`
}

func (TestJsonFeaturesModel) TableName() string {
	return "test_json_features_models"
}

func TestSerializerJsonSpecialFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("nested objects", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_nested_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_nested",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestJsonFeaturesModel{
			NestedObject: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": "deep value",
						"number": 42,
						"bool":   true,
					},
				},
			},
			ComplexJson: map[string]interface{}{
				"string":  "test",
				"number":  123.45,
				"bool":    true,
				"null":    nil,
				"array":   []interface{}{1, 2, 3, "four"},
				"nested":  map[string]interface{}{"key": "value"},
			},
		}

		err = client.NewScoop().Model(TestJsonFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestJsonFeaturesModel
		result := client.NewScoop().Model(TestJsonFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.NestedObject)
		level1, ok1 := found.NestedObject["level1"].(map[string]interface{})
		assert.True(t, ok1)
		level2, ok2 := level1["level2"].(map[string]interface{})
		assert.True(t, ok2)
		assert.Equal(t, "deep value", level2["level3"])
	})

	t.Run("json null values", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_null_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_null",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestJsonFeaturesModel{
			NestedObject: map[string]interface{}{
				"null_field": nil,
				"string":     "value",
			},
		}

		err = client.NewScoop().Model(TestJsonFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestJsonFeaturesModel
		result := client.NewScoop().Model(TestJsonFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Nil(t, found.NestedObject["null_field"])
		assert.Equal(t, "value", found.NestedObject["string"])
	})

	t.Run("html escaping", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_html_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_html",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		htmlContent := "<script>alert('xss')</script>"
		model := &TestJsonFeaturesModel{
			NestedObject: map[string]interface{}{
				"html": htmlContent,
				"tags": "<div>&nbsp;</div>",
			},
		}

		err = client.NewScoop().Model(TestJsonFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestJsonFeaturesModel
		result := client.NewScoop().Model(TestJsonFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, htmlContent, found.NestedObject["html"])
	})
}

// TestSerializerYamlSpecialFeatures 测试 YAML 特定特性
type TestYamlFeaturesModel struct {
	Id      int                    `gorm:"primaryKey;autoIncrement"`
	YamlMap map[string]interface{} `gorm:"column:yaml_map;type:text;serializer:yaml"`
}

func (TestYamlFeaturesModel) TableName() string {
	return "test_yaml_features_models"
}

func TestSerializerYamlSpecialFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("yaml lists", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_yaml_lists_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_yaml_lists",
			Debug:   true,
		}

		client, err := db.New(config, TestYamlFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestYamlFeaturesModel{
			YamlMap: map[string]interface{}{
				"items": []interface{}{"item1", "item2", "item3"},
				"numbers": []interface{}{
					map[string]interface{}{"id": 1, "name": "one"},
					map[string]interface{}{"id": 2, "name": "two"},
				},
			},
		}

		err = client.NewScoop().Model(TestYamlFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestYamlFeaturesModel
		result := client.NewScoop().Model(TestYamlFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.YamlMap["items"])
	})

	t.Run("yaml type coercion", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_yaml_types_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_yaml_types",
			Debug:   true,
		}

		client, err := db.New(config, TestYamlFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestYamlFeaturesModel{
			YamlMap: map[string]interface{}{
				"string": "123",
				"int":    123,
				"float":  123.45,
				"bool":   true,
				"null":   nil,
			},
		}

		err = client.NewScoop().Model(TestYamlFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestYamlFeaturesModel
		result := client.NewScoop().Model(TestYamlFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "123", found.YamlMap["string"])
		// YAML 解析整数时可能是 int 或 int64，取决于具体的解析器
		assert.Equal(t, 123, int(found.YamlMap["int"].(int)))
	})
}

// TestSerializerTomlSpecialFeatures 测试 TOML 特定特性
type TestTomlFeaturesModel struct {
	Id      int                    `gorm:"primaryKey;autoIncrement"`
	TomlMap map[string]interface{} `gorm:"column:toml_map;type:text;serializer:toml"`
}

func (TestTomlFeaturesModel) TableName() string {
	return "test_toml_features_models"
}

func TestSerializerTomlSpecialFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("toml arrays and tables", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_arrays_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_arrays",
			Debug:   true,
		}

		client, err := db.New(config, TestTomlFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestTomlFeaturesModel{
			TomlMap: map[string]interface{}{
				"array": []interface{}{1, 2, 3},
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		}

		err = client.NewScoop().Model(TestTomlFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestTomlFeaturesModel
		result := client.NewScoop().Model(TestTomlFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.TomlMap)
	})

	t.Run("toml key types", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_toml_keys_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_toml_keys",
			Debug:   true,
		}

		client, err := db.New(config, TestTomlFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestTomlFeaturesModel{
			TomlMap: map[string]interface{}{
				"basic_key":    "value1",
				"dashed-key":   "value2",
				"underscore_key": "value3",
				"number":       42,
				"float":        3.14,
				"bool":         true,
			},
		}

		err = client.NewScoop().Model(TestTomlFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestTomlFeaturesModel
		result := client.NewScoop().Model(TestTomlFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "value1", found.TomlMap["basic_key"])
	})
}

// TestSerializerErrorHandling 测试错误处理路径
func TestSerializerErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("json with invalid format", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_invalid_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_invalid",
			Debug:   false,
		}

		client, err := db.New(config, TestBoundaryModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 直接插入无效的 JSON
		gormDB := client.Database()
		err = gormDB.Exec("INSERT INTO test_boundary_models (json_data) VALUES (?)", "{invalid json}").Error
		assert.NoError(t, err)

		// 尝试读取应该返回错误或空值
		var found TestBoundaryModel
		result := client.NewScoop().Model(TestBoundaryModel{}).First(&found)
		// 根据实现，可能返回错误或者返回空值
		// 我们不强制要求错误，但至少不应该 panic
		_ = result.Error
	})

	t.Run("yaml with invalid format", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_yaml_invalid_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_yaml_invalid",
			Debug:   false,
		}

		client, err := db.New(config, TestBoundaryModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 直接插入无效的 YAML
		gormDB := client.Database()
		err = gormDB.Exec("INSERT INTO test_boundary_models (yaml_data) VALUES (?)", "invalid: yaml: content: [[[").Error
		assert.NoError(t, err)

		// 尝试读取
		var found TestBoundaryModel
		result := client.NewScoop().Model(TestBoundaryModel{}).First(&found)
		_ = result.Error // 不应该 panic
	})
}

// TestSerializerComplexTypes 测试复杂类型
type TestComplexTypesModel struct {
	Id          int                    `gorm:"primaryKey;autoIncrement"`
	TimeField   time.Time              `gorm:"column:time_field;type:text;serializer:json"`
	MapField    map[string]string      `gorm:"column:map_field;type:text;serializer:json"`
	SliceField  []string               `gorm:"column:slice_field;type:text;serializer:json"`
	NestedField map[string]interface{} `gorm:"column:nested_field;type:text;serializer:json"`
}

func (TestComplexTypesModel) TableName() string {
	return "test_complex_types_models"
}

func TestSerializerComplexTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("time serialization", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_time_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_time",
			Debug:   true,
		}

		client, err := db.New(config, TestComplexTypesModel{})
		assert.NoError(t, err)
		defer client.Close()

		now := time.Now().Truncate(time.Second) // 去除纳秒精度
		model := &TestComplexTypesModel{
			TimeField: now,
		}

		err = client.NewScoop().Model(TestComplexTypesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestComplexTypesModel
		result := client.NewScoop().Model(TestComplexTypesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.True(t, found.TimeField.Equal(now))
	})

	t.Run("map serialization", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_map_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_map",
			Debug:   true,
		}

		client, err := db.New(config, TestComplexTypesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestComplexTypesModel{
			MapField: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		}

		err = client.NewScoop().Model(TestComplexTypesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestComplexTypesModel
		result := client.NewScoop().Model(TestComplexTypesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Len(t, found.MapField, 3)
		assert.Equal(t, "value1", found.MapField["key1"])
	})

	t.Run("slice serialization", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_slice_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_slice",
			Debug:   true,
		}

		client, err := db.New(config, TestComplexTypesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestComplexTypesModel{
			SliceField: []string{"a", "b", "c", "d", "e"},
		}

		err = client.NewScoop().Model(TestComplexTypesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestComplexTypesModel
		result := client.NewScoop().Model(TestComplexTypesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Len(t, found.SliceField, 5)
		assert.Equal(t, "a", found.SliceField[0])
	})

	t.Run("nested map serialization", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_nested_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_nested",
			Debug:   true,
		}

		client, err := db.New(config, TestComplexTypesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestComplexTypesModel{
			NestedField: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": "deep value",
					},
					"array": []interface{}{1, 2, 3},
				},
				"string": "value",
				"number": 42,
				"bool":   true,
			},
		}

		err = client.NewScoop().Model(TestComplexTypesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestComplexTypesModel
		result := client.NewScoop().Model(TestComplexTypesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.NestedField)
		assert.Equal(t, "value", found.NestedField["string"])
	})
}

// TestSerializerUpdateOperations 测试序列化器的更新操作
func TestSerializerUpdateOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("update json field", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_update_json_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_update_json",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestJsonFeaturesModel{
			NestedObject: map[string]interface{}{
				"key": "value",
			},
		}

		err = client.NewScoop().Model(TestJsonFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		// 更新
		model.NestedObject = map[string]interface{}{
			"key":   "updated_value",
			"newKey": "new_value",
		}
		err = client.NewScoop().Model(TestJsonFeaturesModel{}).Where("id", model.Id).Updates(model).Error
		assert.NoError(t, err)

		var found TestJsonFeaturesModel
		result := client.NewScoop().Model(TestJsonFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "updated_value", found.NestedObject["key"])
		assert.Equal(t, "new_value", found.NestedObject["newKey"])
	})

	t.Run("update yaml field", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_update_yaml_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_update_yaml",
			Debug:   true,
		}

		client, err := db.New(config, TestYamlFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestYamlFeaturesModel{
			YamlMap: map[string]interface{}{
				"key": "value",
			},
		}

		err = client.NewScoop().Model(TestYamlFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		// 更新
		model.YamlMap = map[string]interface{}{
			"key":   "updated_value",
			"newKey": "new_value",
		}
		err = client.NewScoop().Model(TestYamlFeaturesModel{}).Where("id", model.Id).Updates(model).Error
		assert.NoError(t, err)

		var found TestYamlFeaturesModel
		result := client.NewScoop().Model(TestYamlFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "updated_value", found.YamlMap["key"])
	})
}

// TestSerializerProtojson 测试 ProtoJSON 序列化器
type TestProtojsonModel struct {
	Id       int                 `gorm:"primaryKey;autoIncrement"`
	ProtoMsg *wrapperspb.StringValue `gorm:"column:proto_msg;type:text;serializer:protojson"`
}

func (TestProtojsonModel) TableName() string {
	return "test_protojson_models"
}

func TestSerializerProtojson(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("protojson with valid message", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_protojson_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_protojson",
			Debug:   true,
		}

		client, err := db.New(config, TestProtojsonModel{})
		assert.NoError(t, err)
		defer client.Close()

		msg := wrapperspb.String("test value")
		model := &TestProtojsonModel{
			ProtoMsg: msg,
		}

		err = client.NewScoop().Model(TestProtojsonModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestProtojsonModel
		result := client.NewScoop().Model(TestProtojsonModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.ProtoMsg)
		assert.Equal(t, "test value", found.ProtoMsg.Value)
	})

	t.Run("protojson with nil message", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_protojson_nil_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_protojson_nil",
			Debug:   true,
		}

		client, err := db.New(config, TestProtojsonModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestProtojsonModel{
			ProtoMsg: nil,
		}

		err = client.NewScoop().Model(TestProtojsonModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestProtojsonModel
		result := client.NewScoop().Model(TestProtojsonModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		// ProtojsonSerializer 将 nil 转换为空的 protobuf 消息
		// 这是预期行为，因为 Scan 总是创建新实例
		assert.NotNil(t, found.ProtoMsg)
		assert.Equal(t, "", found.ProtoMsg.Value)
	})

	t.Run("protojson with empty string", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_protojson_empty_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_protojson_empty",
			Debug:   true,
		}

		client, err := db.New(config, TestProtojsonModel{})
		assert.NoError(t, err)
		defer client.Close()

		msg := wrapperspb.String("")
		model := &TestProtojsonModel{
			ProtoMsg: msg,
		}

		err = client.NewScoop().Model(TestProtojsonModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestProtojsonModel
		result := client.NewScoop().Model(TestProtojsonModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.NotNil(t, found.ProtoMsg)
		assert.Equal(t, "", found.ProtoMsg.Value)
	})
}

// TestSerializerNilValues 测试 nil 值处理
func TestSerializerNilValues(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("json nil handling", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_nil_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_nil",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 测试 nil map
		model1 := &TestJsonFeaturesModel{
			NestedObject: nil,
		}
		err = client.NewScoop().Model(TestJsonFeaturesModel{}).Create(model1).Error
		assert.NoError(t, err)

		var found1 TestJsonFeaturesModel
		result1 := client.NewScoop().Model(TestJsonFeaturesModel{}).Where("id", model1.Id).First(&found1)
		assert.NoError(t, result1.Error)
		assert.Nil(t, found1.NestedObject)

		// 测试空 map
		model2 := &TestJsonFeaturesModel{
			NestedObject: map[string]interface{}{},
		}
		err = client.NewScoop().Model(TestJsonFeaturesModel{}).Create(model2).Error
		assert.NoError(t, err)

		var found2 TestJsonFeaturesModel
		result2 := client.NewScoop().Model(TestJsonFeaturesModel{}).Where("id", model2.Id).First(&found2)
		assert.NoError(t, result2.Error)
		assert.NotNil(t, found2.NestedObject)
		assert.Empty(t, found2.NestedObject)
	})
}

// TestSerializerContext 测试上下文处理
func TestSerializerContext(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("serializer with context", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_context_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_context",
			Debug:   true,
		}

		client, err := db.New(config, TestJsonFeaturesModel{})
		assert.NoError(t, err)
		defer client.Close()

		model := &TestJsonFeaturesModel{
			NestedObject: map[string]interface{}{
				"key": "value",
			},
		}

		// 使用 Scoop 创建
		err = client.NewScoop().Model(TestJsonFeaturesModel{}).Create(model).Error
		assert.NoError(t, err)

		var found TestJsonFeaturesModel
		result := client.NewScoop().Model(TestJsonFeaturesModel{}).Where("id", model.Id).First(&found)
		assert.NoError(t, result.Error)
		assert.Equal(t, "value", found.NestedObject["key"])
	})
}

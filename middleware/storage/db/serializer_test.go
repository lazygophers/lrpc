package db_test

import (
"context"
"github.com/lazygophers/lrpc/middleware/storage/db"
"github.com/stretchr/testify/assert"
"reflect"
"testing"
)

// TestUserWithJSON 测试包含JSON字段的用户模型
type TestUserWithJSON struct {
	Id       int64  `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"size:100;not null"`
	Metadata string `gorm:"type:json;serializer:json"` // 使用JsonSerializer
}

func (TestUserWithJSON) TableName() string {
	return "test_users_with_json"
}

// TestScoop_CreateWithJSON 测试创建包含JSON字段的记录
func TestScoop_CreateWithJSON(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Create with JSON metadata", func(t *testing.T) {
		model := db.NewModel[TestUserWithJSON](client)
		assert.NotNil(t, model)

		// 测试模型的创建
		assert.Equal(t, "test_users_with_json", model.TableName())
	})
}

// TestModel_WithSerializer 测试带序列化器的模型
func TestModel_WithSerializer(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Model with JSON serializer", func(t *testing.T) {
		model := db.NewModel[TestUserWithJSON](client)
		modelScoop := model.NewScoop()

		// 测试序列化器注册
		assert.NotNil(t, modelScoop)
	})
}

// TestSerializerRegistration 测试序列化器注册
func TestSerializerRegistration(t *testing.T) {
	t.Run("All serializers registered", func(t *testing.T) {
		// 这个测试验证所有序列化器都已正确注册
		// 序列化器在init()函数中注册
		// 通过创建各种类型的模型来间接测试

		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Close()

		// JsonSerializer
		jsonModel := db.NewModel[TestUserWithJSON](client)
		assert.NotNil(t, jsonModel)
	})
}

type TestModelWithYaml struct {
	Id       int64       `gorm:"primaryKey;autoIncrement"`
	Name     string      `gorm:"size:100;not null"`
	Metadata string      `gorm:"type:yaml;serializer:yaml"` // 使用YamlSerializer
}

func (TestModelWithYaml) TableName() string {
	return "test_model_with_yaml"
}

// TestModelWithIni 测试包含INI字段的用户模型
type TestModelWithIni struct {
	Id       int64  `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"size:100;not null"`
	Metadata string `gorm:"type:ini;serializer:ini"` // 使用IniSerializer
}

func (TestModelWithIni) TableName() string {
	return "test_model_with_ini"
}

// TestModelWithBson 测试包含BSON字段的用户模型
type TestModelWithBson struct {
	Id       int64  `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"size:100;not null"`
	Metadata string `gorm:"type:bson;serializer:bson"` // 使用BsonSerializer
}

func (TestModelWithBson) TableName() string {
	return "test_model_with_bson"
}

// TestModelWithToml 测试包含TOML字段的用户模型
type TestModelWithToml struct {
	Id       int64  `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"size:100;not null"`
	Metadata string `gorm:"type:toml;serializer:toml"` // 使用TomlSerializer
}

func (TestModelWithToml) TableName() string {
	return "test_model_with_toml"
}

// TestModelWithProtojson 测试包含Protojson字段的用户模型
type TestModelWithProtojson struct {
	Id       int64  `gorm:"primaryKey;autoIncrement"`
	Name     string `gorm:"size:100;not null"`
	Metadata string `gorm:"type:protojson;serializer:protojson"` // 使用ProtojsonSerializer
}

func (TestModelWithProtojson) TableName() string {
	return "test_model_with_protojson"
}

// TestSerializer_Yaml 测试YAML序列化器
func TestSerializer_Yaml(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Model with YAML serializer", func(t *testing.T) {
		model := db.NewModel[TestModelWithYaml](client)
		modelScoop := model.NewScoop()

		assert.NotNil(t, modelScoop)
		assert.Equal(t, "test_model_with_yaml", model.TableName())
	})
}

// TestSerializer_Ini 测试INI序列化器
func TestSerializer_Ini(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Model with INI serializer", func(t *testing.T) {
		model := db.NewModel[TestModelWithIni](client)
		modelScoop := model.NewScoop()

		assert.NotNil(t, modelScoop)
		assert.Equal(t, "test_model_with_ini", model.TableName())
	})
}

// TestSerializer_Bson 测试BSON序列化器
func TestSerializer_Bson(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Model with BSON serializer", func(t *testing.T) {
		model := db.NewModel[TestModelWithBson](client)
		modelScoop := model.NewScoop()

		assert.NotNil(t, modelScoop)
		assert.Equal(t, "test_model_with_bson", model.TableName())
	})
}

// TestSerializer_Toml 测试TOML序列化器
func TestSerializer_Toml(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Model with TOML serializer", func(t *testing.T) {
		model := db.NewModel[TestModelWithToml](client)
		modelScoop := model.NewScoop()

		assert.NotNil(t, modelScoop)
		assert.Equal(t, "test_model_with_toml", model.TableName())
	})
}

// TestSerializer_Protojson 测试Protojson序列化器
func TestSerializer_Protojson(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Model with Protojson serializer", func(t *testing.T) {
		model := db.NewModel[TestModelWithProtojson](client)
		modelScoop := model.NewScoop()

		assert.NotNil(t, modelScoop)
		assert.Equal(t, "test_model_with_protojson", model.TableName())
	})
}

// TestSerializer_AllSerializers 测试所有序列化器注册
func TestSerializer_AllSerializers(t *testing.T) {
	t.Run("All serializers models", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Close()

		// JsonSerializer
		jsonModel := db.NewModel[TestUserWithJSON](client)
		assert.NotNil(t, jsonModel)

		// YamlSerializer
		yamlModel := db.NewModel[TestModelWithYaml](client)
		assert.NotNil(t, yamlModel)

		// IniSerializer
		iniModel := db.NewModel[TestModelWithIni](client)
		assert.NotNil(t, iniModel)

		// BsonSerializer
		bsonModel := db.NewModel[TestModelWithBson](client)
		assert.NotNil(t, bsonModel)

		// TomlSerializer
		tomlModel := db.NewModel[TestModelWithToml](client)
		assert.NotNil(t, tomlModel)

		// ProtojsonSerializer
		protojsonModel := db.NewModel[TestModelWithProtojson](client)
		assert.NotNil(t, protojsonModel)
	})
}

func TestSerializerMethods_Json(t *testing.T) {
	serializer := &db.JsonSerializer{}
	ctx := context.Background()

	t.Run("Value with valid data", func(t *testing.T) {
		data := map[string]interface{}{
			"key": "value",
		}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Value with nil", func(t *testing.T) {
		var data interface{}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, nil)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})

	t.Run("Value with struct", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		data := TestStruct{Name: "test", Age: 25}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestSerializerMethods_Yaml 测试 YamlSerializer Value 方法
func TestSerializerMethods_Yaml(t *testing.T) {
	serializer := &db.YamlSerializer{}
	ctx := context.Background()

	t.Run("Value with valid data", func(t *testing.T) {
		data := map[string]interface{}{
			"key": "value",
		}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Value with nil", func(t *testing.T) {
		var data interface{}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, nil)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

// TestSerializerMethods_Bson 测试 BsonSerializer Value 方法
func TestSerializerMethods_Bson(t *testing.T) {
	serializer := &db.BsonSerializer{}
	ctx := context.Background()

	t.Run("Value with valid data", func(t *testing.T) {
		data := map[string]interface{}{
			"key": "value",
		}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Value with nil", func(t *testing.T) {
		var data interface{}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, nil)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

// TestSerializerMethods_Toml 测试 TomlSerializer Value 方法
func TestSerializerMethods_Toml(t *testing.T) {
	serializer := &db.TomlSerializer{}
	ctx := context.Background()

	t.Run("Value with valid data", func(t *testing.T) {
		data := map[string]interface{}{
			"key": "value",
		}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, data)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestSerializerMethods_Protojson 测试 ProtojsonSerializer Value 方法
func TestSerializerMethods_Protojson(t *testing.T) {
	serializer := &db.ProtojsonSerializer{}
	ctx := context.Background()

	t.Run("Value with nil", func(t *testing.T) {
		var data interface{}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, nil)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

// TestSerializerMethods_Ini 测试 IniSerializer Value 方法
func TestSerializerMethods_Ini(t *testing.T) {
	serializer := &db.IniSerializer{}
	ctx := context.Background()

	t.Run("Value with nil", func(t *testing.T) {
		var data interface{}
		dst := reflect.ValueOf(&data).Elem()
		result, err := serializer.Value(ctx, nil, dst, nil)
		assert.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

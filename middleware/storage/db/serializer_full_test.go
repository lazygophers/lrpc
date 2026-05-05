package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelWithYaml 测试包含YAML字段的用户模型
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

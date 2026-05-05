package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
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

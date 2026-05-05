package db_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestSerializerMethods_Json 测试 JsonSerializer Value 方法
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

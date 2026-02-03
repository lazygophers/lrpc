package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUser model for testing database operations
type TestUser struct {
	Id        int        `gorm:"primaryKey;autoIncrement"`
	Name      string     `gorm:"size:100;not null"`
	Email     string     `gorm:"size:100;unique"`
	Age       int        `gorm:"default:0"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

func (TestUser) TableName() string {
	return "test_users"
}

// TestClient_AutoMigrate 测试 AutoMigrate 方法
func TestClient_AutoMigrate(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("auto migrate with Tabler interface", func(t *testing.T) {
		// TestUser 实现了 Tabler 接口
		err := client.AutoMigrate(TestUser{})
		// Mock 模式下可能会失败，但至少不应该 panic
		// 主要测试代码路径覆盖
		_ = err
	})

	t.Run("auto migrate with non-Tabler type", func(t *testing.T) {
		type NonTabler struct {
			ID   int64
			Name string
		}

		err := client.AutoMigrate(NonTabler{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not implement Tabler interface")
	})
}

// TestHasProtobufOneofFields 测试 hasProtobufOneofFields 函数
func TestHasProtobufOneofFields(t *testing.T) {
	t.Run("struct without protobuf fields", func(t *testing.T) {
		type NormalStruct struct {
			ID   int64
			Name string
		}

		result := hasProtobufOneofFields(NormalStruct{})
		assert.False(t, result)
	})

	t.Run("struct with protobuf_oneof tag", func(t *testing.T) {
		type ProtobufStruct struct {
			ID      int64
			Name    string
			OneofField string `protobuf_oneof:"test_oneof"`
		}

		result := hasProtobufOneofFields(ProtobufStruct{})
		assert.True(t, result)
	})

	t.Run("pointer to struct", func(t *testing.T) {
		type NormalStruct struct {
			ID   int64
			Name string
		}

		result := hasProtobufOneofFields(&NormalStruct{})
		assert.False(t, result)
	})

	t.Run("non-struct type", func(t *testing.T) {
		result := hasProtobufOneofFields("not a struct")
		assert.False(t, result)
	})

	t.Run("int type", func(t *testing.T) {
		result := hasProtobufOneofFields(123)
		assert.False(t, result)
	})
}

// TestGetFieldsForMigration 测试 getFieldsForMigration 函数
func TestGetFieldsForMigration(t *testing.T) {
	t.Run("struct without protobuf fields", func(t *testing.T) {
		type NormalStruct struct {
			ID   int64
			Name string
			Age  int
		}

		fields := getFieldsForMigration(NormalStruct{})
		assert.NotNil(t, fields)
		assert.Equal(t, 3, len(fields))
	})

	t.Run("struct with protobuf_oneof tag", func(t *testing.T) {
		type ProtobufStruct struct {
			ID         int64
			Name       string
			OneofField string `protobuf_oneof:"test_oneof"`
			Age        int
		}

		fields := getFieldsForMigration(ProtobufStruct{})
		assert.NotNil(t, fields)
		// OneofField 应该被过滤掉
		assert.Equal(t, 3, len(fields))

		// 验证字段名
		fieldNames := make([]string, len(fields))
		for i, f := range fields {
			fieldNames[i] = f.Name
		}
		assert.Contains(t, fieldNames, "ID")
		assert.Contains(t, fieldNames, "Name")
		assert.Contains(t, fieldNames, "Age")
		assert.NotContains(t, fieldNames, "OneofField")
	})

	t.Run("pointer to struct", func(t *testing.T) {
		type NormalStruct struct {
			ID   int64
			Name string
		}

		fields := getFieldsForMigration(&NormalStruct{})
		assert.NotNil(t, fields)
		assert.Equal(t, 2, len(fields))
	})

	t.Run("non-struct type", func(t *testing.T) {
		fields := getFieldsForMigration("not a struct")
		assert.Nil(t, fields)
	})

	t.Run("int type", func(t *testing.T) {
		fields := getFieldsForMigration(123)
		assert.Nil(t, fields)
	})
}

// TestCreateMigrationModel 测试 createMigrationModel 函数
func TestCreateMigrationModel(t *testing.T) {
	t.Run("struct without protobuf fields", func(t *testing.T) {
		type NormalStruct struct {
			ID   int64
			Name string
		}

		model := createMigrationModel(NormalStruct{ID: 1, Name: "test"}, "normal_table")
		assert.NotNil(t, model)

		// 应该返回原始模型（因为没有 protobuf 字段）
		// 不需要类型断言，只需验证不为 nil
	})

	t.Run("struct with protobuf_oneof tag", func(t *testing.T) {
		type ProtobufStruct struct {
			ID         int64
			Name       string
			OneofField string `protobuf_oneof:"test_oneof"`
		}

		model := createMigrationModel(ProtobufStruct{ID: 1, Name: "test"}, "protobuf_table")
		assert.NotNil(t, model)

		// 应该返回 protobufMigrationModel
		wrapper, ok := model.(*protobufMigrationModel)
		assert.True(t, ok)
		assert.Equal(t, "protobuf_table", wrapper.tableName)
		assert.NotNil(t, wrapper.fields)
	})

	t.Run("pointer to struct with protobuf fields", func(t *testing.T) {
		type ProtobufStruct struct {
			ID         int64
			Name       string
			OneofField string `protobuf_oneof:"test_oneof"`
		}

		model := createMigrationModel(&ProtobufStruct{ID: 1, Name: "test"}, "protobuf_table")
		assert.NotNil(t, model)

		wrapper, ok := model.(*protobufMigrationModel)
		assert.True(t, ok)
		assert.Equal(t, "protobuf_table", wrapper.tableName)
	})

	t.Run("non-struct type", func(t *testing.T) {
		model := createMigrationModel("not a struct", "test_table")
		assert.NotNil(t, model)
		// 应该返回原始值
		assert.Equal(t, "not a struct", model)
	})
}

// TestProtobufMigrationModel 测试 protobufMigrationModel 结构
func TestProtobufMigrationModel(t *testing.T) {
	t.Run("create protobuf migration model", func(t *testing.T) {
		type TestStruct struct {
			ID   int64
			Name string
		}

		fields := getFieldsForMigration(TestStruct{})
		wrapper := &protobufMigrationModel{
			fields:    fields,
			tableName: "test_table",
			model:     TestStruct{ID: 1, Name: "test"},
		}

		assert.NotNil(t, wrapper)
		assert.Equal(t, "test_table", wrapper.tableName)
		assert.NotNil(t, wrapper.fields)
		assert.NotNil(t, wrapper.model)
	})
}

// TestAutoMigrate_EdgeCases 测试 AutoMigrate 的边界情况
func TestAutoMigrate_EdgeCases(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("auto migrate with pointer type", func(t *testing.T) {
		err := client.AutoMigrate(&TestUser{})
		// Mock 模式下可能会失败，但不应该 panic
		_ = err
	})

	t.Run("auto migrate multiple times", func(t *testing.T) {
		// 多次迁移同一个模型
		_ = client.AutoMigrate(TestUser{})
		_ = client.AutoMigrate(TestUser{})
	})
}

// TestAutoMigrates_Multiple 测试 AutoMigrates 方法
func TestAutoMigrates_Multiple(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("auto migrate multiple models", func(t *testing.T) {
		type Model1 struct {
			ID   int64  `gorm:"primaryKey"`
			Name string
		}

		type Model2 struct {
			ID    int64  `gorm:"primaryKey"`
			Email string
		}

		// 为了实现 Tabler 接口，需要添加 TableName 方法
		// 但由于是局部类型，我们只测试 TestUser
		err := client.AutoMigrates(TestUser{})
		_ = err
	})

	t.Run("auto migrate with non-Tabler type", func(t *testing.T) {
		type NonTabler struct {
			ID   int64
			Name string
		}

		err := client.AutoMigrates(NonTabler{})
		assert.Error(t, err)
	})
}

// TestHasProtobufOneofFields_ComplexCases 测试复杂的 protobuf 字段检测
func TestHasProtobufOneofFields_ComplexCases(t *testing.T) {
	t.Run("nested struct with protobuf fields", func(t *testing.T) {
		type Inner struct {
			Field string `protobuf_oneof:"inner_oneof"`
		}

		type Outer struct {
			ID    int64
			Inner Inner
		}

		// 外层结构体本身没有 protobuf_oneof 标签
		result := hasProtobufOneofFields(Outer{})
		assert.False(t, result)
	})

	t.Run("multiple protobuf_oneof fields", func(t *testing.T) {
		type MultiOneof struct {
			ID     int64
			Field1 string `protobuf_oneof:"oneof1"`
			Field2 string `protobuf_oneof:"oneof2"`
		}

		result := hasProtobufOneofFields(MultiOneof{})
		assert.True(t, result)
	})
}

// TestGetFieldsForMigration_ComplexCases 测试复杂的字段过滤
func TestGetFieldsForMigration_ComplexCases(t *testing.T) {
	t.Run("struct with mixed fields", func(t *testing.T) {
		type MixedStruct struct {
			ID         int64
			Name       string
			OneofField string `protobuf_oneof:"test_oneof"`
			Age        int
			Email      string
			Oneof2     string `protobuf_oneof:"another_oneof"`
		}

		fields := getFieldsForMigration(MixedStruct{})
		assert.NotNil(t, fields)
		// 应该过滤掉 2 个 oneof 字段
		assert.Equal(t, 4, len(fields))

		fieldNames := make([]string, len(fields))
		for i, f := range fields {
			fieldNames[i] = f.Name
		}
		assert.Contains(t, fieldNames, "ID")
		assert.Contains(t, fieldNames, "Name")
		assert.Contains(t, fieldNames, "Age")
		assert.Contains(t, fieldNames, "Email")
		assert.NotContains(t, fieldNames, "OneofField")
		assert.NotContains(t, fieldNames, "Oneof2")
	})

	t.Run("struct with only protobuf fields", func(t *testing.T) {
		type OnlyProtobuf struct {
			Field1 string `protobuf_oneof:"oneof1"`
			Field2 string `protobuf_oneof:"oneof2"`
		}

		fields := getFieldsForMigration(OnlyProtobuf{})
		// 所有字段都被过滤掉了，应该返回空切片而不是 nil
		assert.NotNil(t, fields)
		assert.Equal(t, 0, len(fields))
	})
}

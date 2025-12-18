package db

import (
	"testing"

	"gorm.io/gorm"
	"gotest.tools/v3/assert"
)

// 测试结构体，包含 gorm:"-" 标记的字段
type TestModelWithIgnoreFields struct {
	Id            uint64 `gorm:"type:uint;column:id;primaryKey;autoIncrement"`
	Name          string `gorm:"type:varchar(255);column:name"`
	IgnoredField1 string `gorm:"-"` // 应该被忽略
	IgnoredField2 int    `gorm:"-"` // 应该被忽略
	NormalField   string `gorm:"type:varchar(255);column:normal_field"`
}

func (TestModelWithIgnoreFields) TableName() string {
	return "test_ignore_fields"
}

// 测试 GORM schema 是否正确处理 gorm:"-" 标签
func TestGormIgnoreFieldsSchema(t *testing.T) {
	// 创建测试数据库配置
	cfg := &Config{
		Type:     MySQL,
		Address:  "127.0.0.1",
		Port:     3306,
		Name:     "test",
		Username: "root",
		Password: "HNEzz4fang.",
	}

	// 创建客户端
	client, err := New(cfg)
	assert.NilError(t, err)
	assert.Assert(t, client != nil)

	// 解析 schema
	stmt := &gorm.Statement{DB: client.db}
	err = stmt.Parse(&TestModelWithIgnoreFields{})
	assert.NilError(t, err)

	// 检查字段是否正确设置 Creatable/Updatable/Readable 属性
	for _, field := range stmt.Schema.Fields {
		t.Logf("Field: %s, DBName: %s, Creatable: %v, Updatable: %v, Readable: %v",
			field.Name, field.DBName, field.Creatable, field.Updatable, field.Readable)

		// 被标记为 gorm:"-" 的字段应该不可创建、不可更新、不可读取
		if field.Name == "IgnoredField1" || field.Name == "IgnoredField2" {
			assert.Assert(t, !field.Creatable, "Field %s should not be creatable", field.Name)
			assert.Assert(t, !field.Updatable, "Field %s should not be updatable", field.Name)
			assert.Assert(t, !field.Readable, "Field %s should not be readable", field.Name)
			// DBName 应该为空
			assert.Equal(t, field.DBName, "", "Field %s should have empty DBName", field.Name)
		} else {
			// 正常字段应该可创建、可更新、可读取
			if field.Name != "ID" { // ID 是自增主键，特殊处理
				assert.Assert(t, field.Creatable, "Field %s should be creatable", field.Name)
				assert.Assert(t, field.Updatable, "Field %s should be updatable", field.Name)
			}
			assert.Assert(t, field.Readable, "Field %s should be readable", field.Name)
		}
	}
}

// 测试 Create 方法是否正确跳过 gorm:"-" 标记的字段
func TestCreateWithIgnoreFields(t *testing.T) {
	// 创建测试数据库配置
	cfg := &Config{
		Type:     MySQL,
		Address:  "127.0.0.1",
		Port:     3306,
		Name:     "test",
		Username: "root",
		Password: "HNEzz4fang.",
	}

	// 创建客户端
	client, err := New(cfg, &TestModelWithIgnoreFields{})
	assert.NilError(t, err)
	assert.Assert(t, client != nil)

	// 清理测试数据
	t.Cleanup(func() {
		_ = client.NewScoop().Table("test_ignore_fields").Delete()
	})

	// 创建测试数据
	testModel := &TestModelWithIgnoreFields{
		Name:          "test",
		IgnoredField1: "should_be_ignored",
		IgnoredField2: 999,
		NormalField:   "normal",
	}

	result := client.NewScoop().Model(testModel).Create(testModel)
	assert.NilError(t, result.Error)
	assert.Assert(t, result.RowsAffected > 0)

	// 验证插入成功且 Id 被设置
	assert.Assert(t, testModel.Id > 0)

	// 从数据库读取并验证
	var retrieved TestModelWithIgnoreFields
	findResult := client.NewScoop().Model(&TestModelWithIgnoreFields{}).Equal("id", testModel.Id).First(&retrieved)
	assert.NilError(t, findResult.Error)
	assert.Equal(t, retrieved.Name, "test")
	assert.Equal(t, retrieved.NormalField, "normal")
	// IgnoredField1 和 IgnoredField2 不应该被保存到数据库，应该是零值
	assert.Equal(t, retrieved.IgnoredField1, "")
	assert.Equal(t, retrieved.IgnoredField2, 0)
}

// 测试 CreateInBatches 方法是否正确跳过 gorm:"-" 标记的字段
func TestCreateInBatchesWithIgnoreFields(t *testing.T) {
	// 创建测试数据库配置
	cfg := &Config{
		Type:     MySQL,
		Address:  "127.0.0.1",
		Port:     3306,
		Name:     "test",
		Username: "root",
		Password: "HNEzz4fang.",
	}

	// 创建客户端
	client, err := New(cfg, &TestModelWithIgnoreFields{})
	assert.NilError(t, err)
	assert.Assert(t, client != nil)

	// 清理测试数据
	t.Cleanup(func() {
		_ = client.NewScoop().Table("test_ignore_fields").Delete()
	})

	// 创建批量测试数据
	testModels := []*TestModelWithIgnoreFields{
		{
			Name:          "test1",
			IgnoredField1: "should_be_ignored1",
			IgnoredField2: 111,
			NormalField:   "normal1",
		},
		{
			Name:          "test2",
			IgnoredField1: "should_be_ignored2",
			IgnoredField2: 222,
			NormalField:   "normal2",
		},
		{
			Name:          "test3",
			IgnoredField1: "should_be_ignored3",
			IgnoredField2: 333,
			NormalField:   "normal3",
		},
	}

	result := client.NewScoop().Model(&TestModelWithIgnoreFields{}).CreateInBatches(testModels, 10)
	assert.NilError(t, result.Error)
	assert.Equal(t, result.RowsAffected, int64(3))

	// 从数据库读取并验证
	var retrieved []*TestModelWithIgnoreFields
	findResult := client.NewScoop().Model(&TestModelWithIgnoreFields{}).Find(&retrieved)
	assert.NilError(t, findResult.Error)
	assert.Equal(t, len(retrieved), 3)

	// 验证每条记录
	for i, record := range retrieved {
		assert.Equal(t, record.Name, testModels[i].Name)
		assert.Equal(t, record.NormalField, testModels[i].NormalField)
		// IgnoredField1 和 IgnoredField2 不应该被保存到数据库
		assert.Equal(t, record.IgnoredField1, "")
		assert.Equal(t, record.IgnoredField2, 0)
	}
}

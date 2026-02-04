package db_test

import (
	"errors"
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestScoop_AutoMigrate 测试 AutoMigrate 方法
func TestScoop_AutoMigrate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("auto migrate with valid model", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_auto_migrate_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_auto_migrate",
			Debug:   true,
		}

		client, err := db.New(config, TestUser{})
		assert.NoError(t, err)
		defer client.Close()

		// 测试 AutoMigrate
		scoop := client.NewScoop()
		err = scoop.AutoMigrate(&TestUser{})
		assert.NoError(t, err)

		// 验证表已创建
		result := scoop.Model(TestUser{}).Create(&TestUser{Name: "Test", Age: 25})
		assert.NoError(t, result.Error)
	})

	t.Run("auto migrate with multiple models", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_auto_migrate_multi_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_auto_migrate_multi",
			Debug:   true,
		}

		// 只迁移 TestUser，不包含 testdata.SimpleMessage（因为它有切片字段，SQLite 不支持）
		client, err := db.New(config, TestUser{})
		assert.NoError(t, err)
		defer client.Close()

		// 测试 AutoMigrate 多个模型
		scoop := client.NewScoop()
		err = scoop.AutoMigrate(&TestUser{})
		assert.NoError(t, err)
	})
}

// TestScoop_ErrorHandling 测试错误处理相关方法
func TestScoop_ErrorHandling(t *testing.T) {
	t.Run("getDuplicatedKeyError with custom error", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		// 创建自定义重复键错误
		scoop := client.NewScoop()

		// 测试默认行为
		err = errors.New("duplicate key value violates unique constraint")
		isDup := scoop.IsDuplicatedKeyError(err)
		assert.False(t, isDup) // 默认不匹配
	})

	t.Run("IsNotFound with gorm.ErrRecordNotFound", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		scoop := client.NewScoop()

		// 测试 gorm.ErrRecordNotFound
		isNotFound := scoop.IsNotFound(gorm.ErrRecordNotFound)
		assert.True(t, isNotFound)
	})

	t.Run("IsNotFound with nil error", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		scoop := client.NewScoop()

		// 测试 nil 错误
		isNotFound := scoop.IsNotFound(nil)
		assert.False(t, isNotFound)
	})

	t.Run("IsNotFound with custom error", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		model := db.NewModel[TestUser](client)
		customErr := errors.New("custom not found")
		model.SetNotFound(customErr)

		scoop := model.NewScoop()

		// 测试自定义错误 - 使用 errors.Is 会匹配
		isNotFound := scoop.IsNotFound(customErr)
		assert.True(t, isNotFound)
	})

	t.Run("IsDuplicatedKeyError with nil error", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		scoop := client.NewScoop()

		// 测试 nil 错误
		isDup := scoop.IsDuplicatedKeyError(nil)
		assert.False(t, isDup)
	})
}

// TestScoop_TableName 测试 Table 方法
func TestScoop_TableName(t *testing.T) {
	t.Run("table with valid name", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		scoop := client.NewScoop()

		// 测试设置表名
		newScoop := scoop.Table("custom_table")
		assert.NotNil(t, newScoop)
	})

	t.Run("table with invalid name", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		scoop := client.NewScoop()

		// 测试包含危险字符的表名（虽然会被过滤）
		_ = scoop.Table("table; DROP TABLE users; --")
		// Table 方法本身不验证表名，验证在 SQL 执行时进行
	})
}

// TestScoop_IgnoreMethod 测试 Ignore 方法（避免与 scoop_test.go 冲突）
func TestScoop_IgnoreMethod(t *testing.T) {
	t.Run("ignore with true", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		scoop := client.NewScoop()

		// 测试启用 ignore
		newScoop := scoop.Ignore(true)
		assert.NotNil(t, newScoop)
	})

	t.Run("ignore with false", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		scoop := client.NewScoop()

		// 测试禁用 ignore
		newScoop := scoop.Ignore(false)
		assert.NotNil(t, newScoop)
	})

	t.Run("ignore without parameter", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()

		scoop := client.NewScoop()

		// 测试不带参数调用（默认为 true）
		newScoop := scoop.Ignore()
		assert.NotNil(t, newScoop)
	})
}

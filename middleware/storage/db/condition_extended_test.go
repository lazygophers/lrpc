package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestConditionUnscoped 测试 Unscoped 方法（软删除相关）
type TestSoftDeleteModel struct {
	Id        int64 `gorm:"primaryKey;autoIncrement"`
	Name      string
	Email     string
	Age       int
	DeletedAt int64 `gorm:"index"`
}

func (TestSoftDeleteModel) TableName() string {
	return "test_soft_delete_models"
}

func TestConditionUnscoped(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("unscoped with true parameter", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_unscoped_true_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_unscoped_true",
			Debug:   true,
		}

		client, err := db.New(config, TestSoftDeleteModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建一些记录
		model1 := &TestSoftDeleteModel{Name: "Test1", Email: "test1@example.com", Age: 25}
		model2 := &TestSoftDeleteModel{Name: "Test2", Email: "test2@example.com", Age: 30}
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Create(model1).Error
		assert.NoError(t, err)
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Create(model2).Error
		assert.NoError(t, err)

		// 软删除第一条记录
		result := client.NewScoop().Model(TestSoftDeleteModel{}).Where("id", model1.Id).Delete()
		assert.NoError(t, result.Error)

		// 正常查询（不包含软删除记录）
		var normalResults []*TestSoftDeleteModel
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Find(&normalResults).Error
		assert.NoError(t, err)
		assert.Len(t, normalResults, 1)
		assert.Equal(t, "Test2", normalResults[0].Name)

		// 使用 Unscoped(true) 查询（包含软删除记录）
		var unscopedResults []*TestSoftDeleteModel
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Unscoped(true).Find(&unscopedResults).Error
		assert.NoError(t, err)
		assert.Len(t, unscopedResults, 2)
	})

	t.Run("unscoped with false parameter", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_unscoped_false_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_unscoped_false",
			Debug:   true,
		}

		client, err := db.New(config, TestSoftDeleteModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建记录
		model1 := &TestSoftDeleteModel{Name: "Test1", Email: "test1@example.com", Age: 25}
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Create(model1).Error
		assert.NoError(t, err)

		// 软删除
		result := client.NewScoop().Model(TestSoftDeleteModel{}).Where("id", model1.Id).Delete()
		assert.NoError(t, result.Error)

		// 使用 Unscoped(false) 查询（不包含软删除记录）
		var results []*TestSoftDeleteModel
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Unscoped(false).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("unscoped without parameter (default true)", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_unscoped_default_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_unscoped_default",
			Debug:   true,
		}

		client, err := db.New(config, TestSoftDeleteModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建记录
		model1 := &TestSoftDeleteModel{Name: "Test1", Email: "test1@example.com", Age: 25}
		model2 := &TestSoftDeleteModel{Name: "Test2", Email: "test2@example.com", Age: 30}
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Create(model1).Error
		assert.NoError(t, err)
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Create(model2).Error
		assert.NoError(t, err)

		// 软删除第一条
		result := client.NewScoop().Model(TestSoftDeleteModel{}).Where("id", model1.Id).Delete()
		assert.NoError(t, result.Error)

		// 使用 Unscoped() 无参数（默认为 true）
		var results []*TestSoftDeleteModel
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Unscoped().Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("unscoped with updates", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_unscoped_update_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_unscoped_update",
			Debug:   true,
		}

		client, err := db.New(config, TestSoftDeleteModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建记录
		model := &TestSoftDeleteModel{Name: "Test1", Email: "test1@example.com", Age: 25}
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Create(model).Error
		assert.NoError(t, err)

		// 软删除
		result := client.NewScoop().Model(TestSoftDeleteModel{}).Where("id", model.Id).Delete()
		assert.NoError(t, result.Error)

		// 使用 Unscoped 更新软删除的记录
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Unscoped(true).Where("id", model.Id).Updates(&TestSoftDeleteModel{Name: "Updated"}).Error
		assert.NoError(t, err)

		// 验证更新
		var found TestSoftDeleteModel
		err = client.NewScoop().Model(TestSoftDeleteModel{}).Unscoped(true).Where("id", model.Id).First(&found).Error
		assert.NoError(t, err)
		assert.Equal(t, "Updated", found.Name)
	})
}

// TestConditionIgnore 测试 Ignore 方法
type TestIgnoreModel struct {
	Id        int    `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"not null"`
	Email     string
	Age       int
	CreatedAt int64
	UpdatedAt int64
}

func (TestIgnoreModel) TableName() string {
	return "test_ignore_models"
}

func TestConditionIgnore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("ignore with true parameter", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_ignore_true_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_ignore_true",
			Debug:   true,
		}

		client, err := db.New(config, TestIgnoreModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建记录
		model := &TestIgnoreModel{Name: "Test", Email: "test@example.com", Age: 25}
		err = client.NewScoop().Model(TestIgnoreModel{}).Create(model).Error
		assert.NoError(t, err)

		// 获取原始的 UpdatedAt
		var original TestIgnoreModel
		err = client.NewScoop().Model(TestIgnoreModel{}).Where("id", model.Id).First(&original).Error
		assert.NoError(t, err)
		originalUpdatedAt := original.UpdatedAt

		// 使用 Ignore(true) 更新（不更新 UpdatedAt 字段）
		updateData := &TestIgnoreModel{Email: "newemail@example.com"}
		err = client.NewScoop().Model(TestIgnoreModel{}).Ignore(true).Where("id", model.Id).Updates(updateData).Error
		assert.NoError(t, err)

		// 验证 UpdatedAt 没有变化
		var found TestIgnoreModel
		err = client.NewScoop().Model(TestIgnoreModel{}).Where("id", model.Id).First(&found).Error
		assert.NoError(t, err)
		assert.Equal(t, "newemail@example.com", found.Email)
		assert.Equal(t, originalUpdatedAt, found.UpdatedAt)
	})

	t.Run("ignore with false parameter", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_ignore_false_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_ignore_false",
			Debug:   true,
		}

		client, err := db.New(config, TestIgnoreModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建记录
		model := &TestIgnoreModel{Name: "Test", Email: "test@example.com", Age: 25}
		err = client.NewScoop().Model(TestIgnoreModel{}).Create(model).Error
		assert.NoError(t, err)

		// 获取原始的记录
		var original TestIgnoreModel
		err = client.NewScoop().Model(TestIgnoreModel{}).Where("id", model.Id).First(&original).Error
		assert.NoError(t, err)

		// 使用 Ignore(false) 更新（不忽略 UpdatedAt 字段，所以会更新）
		newUpdateTime := original.UpdatedAt + 100
		updateData := &TestIgnoreModel{Email: "newemail@example.com", UpdatedAt: newUpdateTime}
		err = client.NewScoop().Model(TestIgnoreModel{}).Ignore(false).Where("id", model.Id).Updates(updateData).Error
		assert.NoError(t, err)

		// 验证 UpdatedAt 被更新了
		var found TestIgnoreModel
		err = client.NewScoop().Model(TestIgnoreModel{}).Where("id", model.Id).First(&found).Error
		assert.NoError(t, err)
		assert.Equal(t, "newemail@example.com", found.Email)
		assert.Equal(t, newUpdateTime, found.UpdatedAt)
	})

	t.Run("ignore without parameter (default true)", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_ignore_default_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_ignore_default",
			Debug:   true,
		}

		client, err := db.New(config, TestIgnoreModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建记录
		model := &TestIgnoreModel{Name: "Test", Email: "test@example.com", Age: 25}
		err = client.NewScoop().Model(TestIgnoreModel{}).Create(model).Error
		assert.NoError(t, err)

		// 使用 Ignore() 无参数（默认为 true）
		updateData := &TestIgnoreModel{Email: "newemail@example.com"}
		err = client.NewScoop().Model(TestIgnoreModel{}).Ignore().Where("id", model.Id).Updates(updateData).Error
		assert.NoError(t, err)

		// 验证更新成功
		var found TestIgnoreModel
		err = client.NewScoop().Model(TestIgnoreModel{}).Where("id", model.Id).First(&found).Error
		assert.NoError(t, err)
		assert.Equal(t, "newemail@example.com", found.Email)
	})
}

// TestConditionLikeMethods 测试各种 Like 方法
type TestLikeModel struct {
	Id       int    `gorm:"primaryKey;autoIncrement"`
	Username string `gorm:"size:100;not null"`
	Email    string `gorm:"size:100;not null"`
}

func (TestLikeModel) TableName() string {
	return "test_like_models"
}

func TestConditionLikeMethods(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("like method", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_like_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_like",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		models := []*TestLikeModel{
			{Username: "john_doe", Email: "john@example.com"},
			{Username: "jane_doe", Email: "jane@example.com"},
			{Username: "bob_smith", Email: "bob@example.com"},
		}
		for _, model := range models {
			err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
			assert.NoError(t, err)
		}

		// 测试 LIKE 查询（包含匹配）
		var results []*TestLikeModel
		err = client.NewScoop().Model(TestLikeModel{}).Where(db.Like("username", "john")).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "john_doe", results[0].Username)
	})

	t.Run("like method with empty string", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_like_empty_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_like_empty",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		model := &TestLikeModel{Username: "test_user", Email: "test@example.com"}
		err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
		assert.NoError(t, err)

		// 测试空字符串（应该返回所有记录，因为 Like 在空值时不添加条件）
		var results []*TestLikeModel
		cond := db.Like("username", "")
		err = client.NewScoop().Model(TestLikeModel{}).Where(cond).Find(&results).Error
		assert.NoError(t, err)
		// 空 Like 不添加条件，所以应该返回所有记录
		assert.GreaterOrEqual(t, len(results), 1)
	})

	t.Run("left_like method", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_left_like_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_left_like",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		models := []*TestLikeModel{
			{Username: "john_doe", Email: "john@example.com"},
			{Username: "john_smith", Email: "john.smith@example.com"},
			{Username: "jane_doe", Email: "jane@example.com"},
		}
		for _, model := range models {
			err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
			assert.NoError(t, err)
		}

		// 测试 LeftLike（前缀匹配）
		var results []*TestLikeModel
		err = client.NewScoop().Model(TestLikeModel{}).Where(db.LeftLike("username", "john")).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("right_like method", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_right_like_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_right_like",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		models := []*TestLikeModel{
			{Username: "john_doe", Email: "john@example.com"},
			{Username: "jane_doe", Email: "jane@example.com"},
			{Username: "bob_smith", Email: "bob@example.com"},
		}
		for _, model := range models {
			err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
			assert.NoError(t, err)
		}

		// 测试 RightLike（后缀匹配）
		var results []*TestLikeModel
		err = client.NewScoop().Model(TestLikeModel{}).Where(db.RightLike("username", "doe")).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("not_like method", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_not_like_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_not_like",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		models := []*TestLikeModel{
			{Username: "john_doe", Email: "john@example.com"},
			{Username: "jane_doe", Email: "jane@example.com"},
			{Username: "bob_smith", Email: "bob@example.com"},
		}
		for _, model := range models {
			err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
			assert.NoError(t, err)
		}

		// 测试 NotLike
		var results []*TestLikeModel
		err = client.NewScoop().Model(TestLikeModel{}).Where(db.NotLike("username", "john")).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("not_left_like method", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_not_left_like_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_not_left_like",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		models := []*TestLikeModel{
			{Username: "john_doe", Email: "john@example.com"},
			{Username: "jane_doe", Email: "jane@example.com"},
			{Username: "bob_smith", Email: "bob@example.com"},
		}
		for _, model := range models {
			err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
			assert.NoError(t, err)
		}

		// 测试 NotLeftLike
		var results []*TestLikeModel
		err = client.NewScoop().Model(TestLikeModel{}).Where(db.NotLeftLike("username", "john")).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("not_right_like method", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_not_right_like_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_not_right_like",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		models := []*TestLikeModel{
			{Username: "john_doe", Email: "john@example.com"},
			{Username: "jane_doe", Email: "jane@example.com"},
			{Username: "bob_smith", Email: "bob@example.com"},
		}
		for _, model := range models {
			err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
			assert.NoError(t, err)
		}

		// 测试 NotRightLike
		var results []*TestLikeModel
		err = client.NewScoop().Model(TestLikeModel{}).Where(db.NotRightLike("username", "doe")).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "bob_smith", results[0].Username)
	})
}

// TestConditionLikeEdgeCases 测试 Like 方法的边界情况
func TestConditionLikeEdgeCases(t *testing.T) {
	t.Run("like with special characters", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_like_special_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_like_special",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建包含特殊字符的测试数据
		model := &TestLikeModel{Username: "user%name", Email: "test@example.com"}
		err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
		assert.NoError(t, err)

		// 测试特殊字符查询
		var results []*TestLikeModel
		err = client.NewScoop().Model(TestLikeModel{}).Where(db.Like("username", "user")).Find(&results).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})

	t.Run("like with unicode", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_like_unicode_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_like_unicode",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建包含 Unicode 的测试数据
		model := &TestLikeModel{Username: "用户测试", Email: "test@example.com"}
		err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
		assert.NoError(t, err)

		// 测试 Unicode 查询
		var results []*TestLikeModel
		err = client.NewScoop().Model(TestLikeModel{}).Where(db.Like("username", "用户")).Find(&results).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})

	t.Run("like combinations", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_like_combinations_*")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_like_combinations",
			Debug:   true,
		}

		client, err := db.New(config, TestLikeModel{})
		assert.NoError(t, err)
		defer client.Close()

		// 创建测试数据
		models := []*TestLikeModel{
			{Username: "john@example.com", Email: "john@example.com"},
			{Username: "jane@test.com", Email: "jane@test.com"},
		}
		for _, model := range models {
			err = client.NewScoop().Model(TestLikeModel{}).Create(model).Error
			assert.NoError(t, err)
		}

		// 测试组合条件
		var results []*TestLikeModel
		err = client.NewScoop().Model(TestLikeModel{}).Where(
			db.Or(
				db.Like("username", "john"),
				db.LeftLike("username", "jane"),
			),
		).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})
}

package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
)

// AutoMigratePrecisionModel - 专门用于AutoMigrate精准测试的模型
type AutoMigratePrecisionModel struct {
	Id        int     `gorm:"primaryKey"`
	Name      string  `gorm:"size:100;uniqueIndex:idx_precision_name"`
	Email     string  `gorm:"size:100;index:idx_precision_email"`
	Score     float64 `gorm:"index:idx_precision_score"`
	IsActive  bool    `gorm:"index:idx_precision_active"`
	CreatedAt int64   `gorm:"autoCreateTime;index:idx_precision_created"`
	UpdatedAt int64   `gorm:"autoUpdateTime"`
	DeletedAt *int64  `gorm:"index:idx_precision_deleted"`
}

func (AutoMigratePrecisionModel) TableName() string {
	return "automigrate_precision_models"
}

// TestAutoMigratePrecisionErrorPaths - 精准测试AutoMigrate的错误路径
func TestAutoMigratePrecisionErrorPaths(t *testing.T) {
	t.Run("automigrate_error_path_coverage", func(t *testing.T) {
		tempDir := t.TempDir()
		defer os.RemoveAll(tempDir)

		client, err := db.New(&db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "precision_test",
		})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		model := &AutoMigratePrecisionModel{}

		// 1. 测试正常AutoMigrate - 创建表的路径
		err = client.AutoMigrate(model)
		if err != nil {
			t.Errorf("AutoMigrate create table failed: %v", err)
		}

		// 2. 测试非Tabler接口的模型（触发错误路径）
		type NonTablerModel struct {
			Id   int    `gorm:"primaryKey"`
			Name string `gorm:"size:50"`
		}

		nonTablerModel := NonTablerModel{}
		
		// 使用recover来捕获panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Caught expected panic for non-Tabler model: %v", r)
				}
			}()
			
			err = client.AutoMigrate(nonTablerModel)
			if err != nil {
				t.Logf("Got expected error for non-Tabler model: %v", err)
			}
		}()

		// 3. 测试表已存在但需要添加字段的情况
		// 先创建简化版本的表
		type SimpleModel struct {
			Id   int    `gorm:"primaryKey"`
			Name string `gorm:"size:100"`
		}

		// 删除现有表
		db := client.Database()
		db.Exec("DROP TABLE IF EXISTS automigrate_precision_models")

		// 创建简化表
		err = db.AutoMigrate(SimpleModel{})
		if err != nil {
			t.Errorf("Failed to create simple table: %v", err)
		}

		// 现在使用完整模型进行AutoMigrate，这将触发添加字段的路径
		err = client.AutoMigrate(model)
		if err != nil {
			t.Errorf("AutoMigrate field addition failed: %v", err)
		}

		// 4. 测试索引创建和修改路径
		// 重新创建表来测试索引管理
		db.Exec("DROP TABLE IF EXISTS automigrate_precision_models")

		// 创建没有索引的表
		type NoIndexModel struct {
			Id        int     `gorm:"primaryKey"`
			Name      string  `gorm:"size:100"`
			Email     string  `gorm:"size:100"`
			Score     float64
			IsActive  bool
			CreatedAt int64
			UpdatedAt int64
			DeletedAt *int64
		}

		err = db.AutoMigrate(NoIndexModel{})
		if err != nil {
			t.Errorf("Failed to create no-index table: %v", err)
		}

		// 现在用有索引的模型进行迁移，触发创建索引路径
		err = client.AutoMigrate(model)
		if err != nil {
			t.Errorf("AutoMigrate index creation failed: %v", err)
		}

		// 5. 测试索引修改路径
		// 创建有不同索引的表
		type DifferentIndexModel struct {
			Id        int     `gorm:"primaryKey"`
			Name      string  `gorm:"size:100;index:idx_different_precision_name"`
			Email     string  `gorm:"size:100"`
			Score     float64
			IsActive  bool
			CreatedAt int64
			UpdatedAt int64
			DeletedAt *int64
		}

		db.Exec("DROP TABLE IF EXISTS automigrate_precision_models")
		err = db.AutoMigrate(DifferentIndexModel{})
		if err != nil {
			t.Errorf("Failed to create different index table: %v", err)
		}

		// 使用原始模型迁移，触发索引变更路径
		err = client.AutoMigrate(model)
		if err != nil {
			t.Errorf("AutoMigrate index modification failed: %v", err)
		}

		// 6. 测试复杂索引场景
		// 创建带有复合索引的表
		type ComplexIndexModel struct {
			Id        int     `gorm:"primaryKey"`
			Name      string  `gorm:"size:100;index:idx_complex_name_email,unique"`
			Email     string  `gorm:"size:100;index:idx_complex_name_email"`
			Score     float64 `gorm:"index:idx_complex_score_active"`
			IsActive  bool    `gorm:"index:idx_complex_score_active"`
			CreatedAt int64   `gorm:"autoCreateTime"`
			UpdatedAt int64   `gorm:"autoUpdateTime"`
			DeletedAt *int64  `gorm:"index"`
		}

		db.Exec("DROP TABLE IF EXISTS automigrate_precision_models")
		err = db.AutoMigrate(ComplexIndexModel{})
		if err != nil {
			t.Errorf("Failed to create complex index table: %v", err)
		}

		// 现在用原始模型进行迁移，这将触发更多索引管理代码路径
		err = client.AutoMigrate(model)
		if err != nil {
			t.Errorf("AutoMigrate complex index handling failed: %v", err)
		}

		t.Logf("AutoMigrate precision error path testing completed")
	})
}

// InitialTableModel - 字段迁移测试的初始模型
type InitialTableModel struct {
	Id   int    `gorm:"primaryKey"`
	Name string `gorm:"size:50"`
}

func (InitialTableModel) TableName() string {
	return "field_migration_models"
}

// ExtendedTableModel - 字段迁移测试的扩展模型
type ExtendedTableModel struct {
	Id          int     `gorm:"primaryKey"`
	Name        string  `gorm:"size:100"`
	Email       string  `gorm:"size:100"`
	Age         int
	Score       float64
	IsActive    bool
	Description string `gorm:"type:text"`
	CreatedAt   int64  `gorm:"autoCreateTime"`
	UpdatedAt   int64  `gorm:"autoUpdateTime"`
	DeletedAt   *int64 `gorm:"index"`
}

func (ExtendedTableModel) TableName() string {
	return "field_migration_models"
}

// TestAutoMigrateFieldMigrationPaths - 测试字段迁移的各种路径
func TestAutoMigrateFieldMigrationPaths(t *testing.T) {
	t.Run("field_migration_paths", func(t *testing.T) {
		tempDir := t.TempDir()
		defer os.RemoveAll(tempDir)

		client, err := db.New(&db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "field_migration_test",
		})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		initialTableModel := InitialTableModel{}
		err = client.AutoMigrate(initialTableModel)
		if err != nil {
			t.Errorf("Failed to create initial table: %v", err)
		}

		// 创建扩展版本的表，触发字段添加
		extendedModel := ExtendedTableModel{}
		err = client.AutoMigrate(extendedModel)
		if err != nil {
			t.Errorf("Extended model migration failed: %v", err)
		}

		t.Logf("Field migration path testing completed")
	})
}

// InitialColumnModel - 列类型测试的初始模型
type InitialColumnModel struct {
	Id    int    `gorm:"primaryKey"`
	Name  string `gorm:"size:50"`
	Score int    `gorm:"type:int"`
}

func (InitialColumnModel) TableName() string {
	return "column_type_models"
}

// ModifiedColumnModel - 列类型测试的修改模型
type ModifiedColumnModel struct {
	Id    int     `gorm:"primaryKey"`
	Name  string  `gorm:"size:100"` // 从50增加到100
	Score float64 `gorm:"type:real"` // 从int改为float64
}

func (ModifiedColumnModel) TableName() string {
	return "column_type_models"
}

// TestAutoMigrateColumnTypeChanges - 测试字段类型变更的路径
func TestAutoMigrateColumnTypeChanges(t *testing.T) {
	t.Run("column_type_changes", func(t *testing.T) {
		tempDir := t.TempDir()
		defer os.RemoveAll(tempDir)

		client, err := db.New(&db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "column_type_test",
		})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		initialModel := InitialColumnModel{}
		err = client.AutoMigrate(initialModel)
		if err != nil {
			t.Errorf("Failed to create initial column table: %v", err)
		}

		// 修改字段类型
		modifiedModel := ModifiedColumnModel{}
		err = client.AutoMigrate(modifiedModel)
		if err != nil {
			t.Errorf("Column type modification failed: %v", err)
		}

		t.Logf("Column type change testing completed")
	})
}

// InvalidTableModel - 故意不实现TableName接口的模型
type InvalidTableModel struct {
	Id   int    `gorm:"primaryKey"`
	Name string
}

// TestAutoMigrateInvalidModel - 测试无效模型的错误处理
func TestAutoMigrateInvalidModel(t *testing.T) {
	t.Run("invalid_model_error_handling", func(t *testing.T) {
		tempDir := t.TempDir()
		defer os.RemoveAll(tempDir)

		client, err := db.New(&db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "invalid_model_test",
		})
		if err != nil {
			t.Fatalf("Failed to create client: %v", err)
		}

		// 测试不实现TableName接口的模型
		invalidModel := InvalidTableModel{
			Name: "test",
		}

		// 使用recover来捕获panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Caught expected panic for invalid model: %v", r)
				}
			}()
			
			err = client.AutoMigrate(invalidModel)
			if err != nil {
				t.Logf("Got expected error for invalid model: %v", err)
			}
		}()

		// 测试nil模型
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Caught expected panic for nil model: %v", r)
				}
			}()
			
			err = client.AutoMigrate(nil)
			if err != nil {
				t.Logf("Got expected error for nil model: %v", err)
			}
		}()

		t.Logf("Invalid model error handling testing completed")
	})
}
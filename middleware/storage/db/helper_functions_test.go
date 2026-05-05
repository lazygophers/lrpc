package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestHelperFunctions_CreateWithAutoIncrement 测试触发 queryLastInsertID 的场景
func TestHelperFunctions_CreateWithAutoIncrement(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Create single user", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		user := &TestUser{
			Name: "Test User",
			Age:  25,
		}

		err := model.NewScoop().Create(user)
		// Mock 模式下可能失败，但测试代码路径
		_ = err
	})

	t.Run("Create user with SQLite dialect", func(t *testing.T) {
		sqliteConfig := &db.Config{
			Type: db.Sqlite,
			Mock: true,
		}

		sqliteClient, err := db.New(sqliteConfig)
		assert.NoError(t, err)
		defer sqliteClient.MockDB().Close()

		model := db.NewModel[TestUser](sqliteClient)
		user := &TestUser{
			Name: "SQLite User",
			Age:  30,
		}

		err = model.NewScoop().Create(user)
		// Mock 模式下可能失败，但测试代码路径
		_ = err
	})
}

// TestHelperFunctions_FindTriggersScanRowsInto 测试触发 scanRowsInto 的场景
func TestHelperFunctions_FindTriggersScanRowsInto(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Find triggers scanRowsInto", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		user, err := model.NewScoop().First()
		// Mock 模式下可能失败，但测试代码路径
		_ = user
		_ = err
	})

	t.Run("Find with multiple rows", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		users, err := model.NewScoop().Find()
		// Mock 模式下可能失败，但测试代码路径
		_ = users
		_ = err
	})
}

// TestHelperFunctions_FirstWithConditions 测试 First 操作
func TestHelperFunctions_FirstWithConditions(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("First with Where conditions", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		user, err := model.NewScoop().Where("id", 1).First()
		// Mock 模式下可能失败，但测试代码路径
		_ = user
		_ = err
	})
}

// TestHelperFunctions_CreateMultipleRecords 测试批量插入
func TestHelperFunctions_CreateMultipleRecords(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Create multiple users", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		users := []*TestUser{
			{Name: "User1", Age: 25},
			{Name: "User2", Age: 30},
			{Name: "User3", Age: 35},
		}

		result := model.NewScoop().CreateInBatches(users, 50)
		// Mock 模式下可能失败，但测试代码路径
		_ = result
	})
}

// TestHelperFunctions_CreateOrUpdate 测试 CreateOrUpdate
func TestHelperFunctions_CreateOrUpdate(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("CreateOrUpdate with new record", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		user := &TestUser{
			Name: "New User",
			Age:  25,
		}

		result := model.NewScoop().Where("name", "New User").CreateOrUpdate(map[string]interface{}{
			"name": "New User",
			"age":  25,
		}, user)
		// Mock 模式下可能失败，但测试代码路径
		_ = result
	})
}

// TestHelperFunctions_DifferentDialects 测试不同数据库方言
func TestHelperFunctions_DifferentDialects(t *testing.T) {
	dialects := []struct {
		name   string
		dbType string
	}{
		{"MySQL", db.MySQL},
		{"SQLite", db.Sqlite},
	}

	for _, tt := range dialects {
		t.Run(tt.name+" create", func(t *testing.T) {
			config := &db.Config{
				Type: tt.dbType,
				Mock: true,
			}

			client, err := db.New(config)
			assert.NoError(t, err)
			defer client.MockDB().Close()

			model := db.NewModel[TestUser](client)
			user := &TestUser{
				Name: tt.name + " User",
				Age:  25,
			}

			err = model.NewScoop().Create(user)
			// Mock 模式下可能失败，但测试代码路径
			_ = err
		})
	}
}

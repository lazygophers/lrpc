package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelScoop_FindByPage 测试 FindByPage 方法
func TestModelScoop_FindByPage(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("find by page successfully", func(t *testing.T) {
		// 先执行 Find 查询获取数据 (offset=0 时不会包含 OFFSET 0)
		client.ExpectQuery("SELECT \\* FROM test_users WHERE deleted_at = 0.*LIMIT 10").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "User1").
				AddRow(2, "User2").
				AddRow(3, "User3"))

		// 然后执行 Count 查询 (因为 ShowTotal=true，deleted_at 条件可能重复)
		client.ExpectQuery("SELECT COUNT\\(\\*\\) FROM test_users WHERE deleted_at = 0.*").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(25)))

		opt := &core.ListOption{}
		opt.SetOffset(0)
		opt.SetLimit(10)
		opt.ShowTotal = true

		page, users, err := model.NewScoop().FindByPage(opt)
		assert.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, uint64(25), page.Total)
		assert.Len(t, users, 3)
		assert.Equal(t, int64(1), users[0].Id)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("find by page with empty result", func(t *testing.T) {
		// Find 查询返回空结果 (offset=0 时不包含 OFFSET 0)
		client.ExpectQuery("SELECT \\* FROM test_users WHERE deleted_at = 0.*LIMIT 10").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

		// Count 查询
		client.ExpectQuery("SELECT COUNT\\(\\*\\) FROM test_users WHERE deleted_at = 0.*").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

		opt := &core.ListOption{}
		opt.SetOffset(0)
		opt.SetLimit(10)
		opt.ShowTotal = true

		page, users, err := model.NewScoop().FindByPage(opt)
		assert.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, uint64(0), page.Total)
		assert.Len(t, users, 0)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("find second page", func(t *testing.T) {
		// Find 查询第二页数据
		client.ExpectQuery("SELECT \\* FROM test_users WHERE deleted_at = 0.*LIMIT 10 OFFSET 10").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(11, "User11").
				AddRow(12, "User12"))

		// Count 查询
		client.ExpectQuery("SELECT COUNT\\(\\*\\) FROM test_users WHERE deleted_at = 0.*").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(25)))

		opt := &core.ListOption{}
		opt.SetOffset(10)
		opt.SetLimit(10)
		opt.ShowTotal = true

		page, users, err := model.NewScoop().FindByPage(opt)
		assert.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, uint64(25), page.Total)
		assert.Len(t, users, 2)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("find by page without ShowTotal", func(t *testing.T) {
		// 不设置 ShowTotal，所以不会有 Count 查询
		client.ExpectQuery("SELECT \\* FROM test_users WHERE deleted_at = 0.*LIMIT 5").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "User1").
				AddRow(2, "User2"))

		opt := &core.ListOption{}
		opt.SetOffset(0)
		opt.SetLimit(5)
		// ShowTotal 默认为 false

		page, users, err := model.NewScoop().FindByPage(opt)
		assert.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, uint64(0), page.Total) // ShowTotal=false 时 Total 为 0
		assert.Len(t, users, 2)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestModelScoop_LimitOffset(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Limit with value", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Limit(10)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Offset with value", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Offset(5)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Chaining Limit and Offset", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Limit(10).Offset(5)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_Select 测试 Select 方法
func TestModelScoop_Select(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Select with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Select("id", "name", "email")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Chaining Select calls", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Select("id").Select("name")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_Equal 测试 Equal 方法
func TestModelScoop_Equal(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Equal with value", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Equal("id", 1)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_GroupOrder 测试 Group 和 Order 方法
func TestModelScoop_GroupOrder(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Group with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Group("status")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Order with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Order("id", "created_at")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Desc with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Desc("id", "created_at")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Asc with fields", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Asc("id", "name")
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_Unscoped 测试 Unscoped 方法
func TestModelScoop_Unscoped(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Unscoped without argument", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Unscoped()
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Unscoped with true", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Unscoped(true)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Unscoped with false", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Unscoped(false)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

// TestModelScoop_Ignore 测试 Ignore 方法
func TestModelScoop_Ignore(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Ignore without argument", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Ignore()
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Ignore with true", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Ignore(true)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})

	t.Run("Ignore with false", func(t *testing.T) {
		model := db.NewModel[TestUser](client)
		modelScoop := model.NewScoop()

		result := modelScoop.Ignore(false)
		assert.NotNil(t, result)
		assert.Equal(t, modelScoop, result)
	})
}

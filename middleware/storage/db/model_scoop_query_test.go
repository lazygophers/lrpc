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
		assert.Equal(t, 1, users[0].Id)

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

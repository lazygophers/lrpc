package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestScoop_NotLikeVariants 测试 NotLike 变体方法
func TestScoop_NotLikeVariants(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("NotLeftLike method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "Johnson"))

		var users []*TestUser
		result := client.NewScoop().Model(TestUser{}).NotLeftLike("name", "John").Find(&users)
		assert.NoError(t, result.Error)
		assert.Len(t, users, 1)

		err := client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("NotRightLike method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "John Doe"))

		var users []*TestUser
		result := client.NewScoop().Model(TestUser{}).NotRightLike("name", "Doe").Find(&users)
		assert.NoError(t, result.Error)
		assert.Len(t, users, 1)

		err := client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("NotLeftLike with ModelScoop", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "Test"))

		model := db.NewModel[TestUser](client)
		users, err := model.NewScoop().NotLeftLike("name", "Test").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("NotRightLike with ModelScoop", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "Test"))

		model := db.NewModel[TestUser](client)
		users, err := model.NewScoop().NotRightLike("name", "Test").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestScoop_WhereWithSlice 测试使用切片参数的 Where 方法
func TestScoop_WhereWithSlice(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("where with interface slice", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(1, "User1"))

		args := []interface{}{"id", 1}
		var users []*TestUser
		result := client.NewScoop().Model(TestUser{}).Where(args...).Find(&users)
		assert.NoError(t, result.Error)
		assert.Len(t, users, 1)

		err := client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestScoop_ToSQL 测试 ToSQL 方法
func TestScoop_ToSQL(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("ToSQL with select query", func(t *testing.T) {
		sql := client.NewScoop().Model(TestUser{}).Where("id", 1).ToSQL(db.SQLOperationSelect)
		assert.NotEmpty(t, sql)
		assert.Contains(t, sql, "SELECT")
	})

	t.Run("ToSQL with update", func(t *testing.T) {
		sql := client.NewScoop().Model(TestUser{}).Where("id", 1).ToSQL(db.SQLOperationUpdate)
		assert.NotEmpty(t, sql)
		assert.Contains(t, sql, "UPDATE")
	})

	t.Run("ToSQL with delete", func(t *testing.T) {
		sql := client.NewScoop().Model(TestUser{}).Where("id", 1).ToSQL(db.SQLOperationDelete)
		assert.NotEmpty(t, sql)
	})
}

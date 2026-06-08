package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelScoop_CreateAdvanced 测试高级创建方法
func TestModelScoop_CreateAdvanced(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("CreateInBatches with valid data", func(t *testing.T) {
		// 为批量插入设置期望 - 3个项目，batch size为2，所以会有2次调用
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 2))
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(3, 1))

		users := []*TestUser{
			{Name: "User1", Email: "user1@example.com", Age: 25},
			{Name: "User2", Email: "user2@example.com", Age: 30},
			{Name: "User3", Email: "user3@example.com", Age: 35},
		}

		result := model.NewScoop().CreateInBatches(users, 2)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(3), result.RowsAffected)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("CreateInBatches with empty slice", func(t *testing.T) {
		users := []*TestUser{}

		result := model.NewScoop().CreateInBatches(users, 10)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(0), result.RowsAffected)
	})

	t.Run("CreateInBatches with single item", func(t *testing.T) {
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		users := []*TestUser{
			{Name: "Single User", Email: "single@example.com", Age: 28},
		}

		result := model.NewScoop().CreateInBatches(users, 10)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("CreateIfNotExists when record exists", func(t *testing.T) {
		// 查询返回存在的记录 - 注意：CreateIfNotExists 使用 SELECT id FROM
		client.ExpectQuery("SELECT id FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))

		user := &TestUser{
			Name:  "Existing User",
			Email: "existing@example.com",
			Age:   30,
		}

		result := model.NewScoop().Where("email", "existing@example.com").CreateIfNotExists(user)
		assert.NoError(t, result.Error)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("CreateIfNotExists when record not exists", func(t *testing.T) {
		// 查询返回空
		client.ExpectQuery("SELECT id FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(10, 1))

		user := &TestUser{
			Name:  "New User",
			Email: "newuser@example.com",
			Age:   25,
		}

		result := model.NewScoop().Where("email", "newuser@example.com").CreateIfNotExists(user)
		assert.NoError(t, result.Error)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("CreateNotExist when record not exists", func(t *testing.T) {
		// 查询返回空 - CreateNotExist 使用 FirstOrCreate，会使用 SELECT * FROM
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(20, 1))

		user := &TestUser{
			Name:  "Unique User",
			Email: "unique@example.com",
			Age:   27,
		}

		result := model.NewScoop().Where("email", "unique@example.com").CreateNotExist(user)
		assert.NoError(t, result.Error)
		assert.True(t, result.IsCreated)
		assert.NotNil(t, result.Object)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("CreateNotExist when record exists", func(t *testing.T) {
		// 查询返回存在的记录
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}).
				AddRow(5, "Existing", "exists@example.com", 40))

		user := &TestUser{
			Name:  "Existing",
			Email: "exists@example.com",
			Age:   40,
		}

		result := model.NewScoop().Where("email", "exists@example.com").CreateNotExist(user)
		assert.NoError(t, result.Error)
		assert.False(t, result.IsCreated)
		assert.NotNil(t, result.Object)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("CreateInBatches with error", func(t *testing.T) {
		client.ExpectExec("INSERT INTO test_users").
			WillReturnError(assert.AnError)

		users := []*TestUser{
			{Name: "Error User", Email: "error@example.com", Age: 25},
		}

		result := model.NewScoop().CreateInBatches(users, 10)
		assert.Error(t, result.Error)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

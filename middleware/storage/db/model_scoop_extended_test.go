package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelScoop_FirstOrCreateExtended 测试 FirstOrCreate 的扩展场景
func TestModelScoop_FirstOrCreateExtended(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("first or create with empty result", func(t *testing.T) {
		// 查询返回空
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(10, 1))

		user := &TestUser{
			Name:  "New User",
			Email: "new@example.com",
			Age:   25,
		}

		result := model.NewScoop().Where("email", "new@example.com").FirstOrCreate(user)
		assert.NoError(t, result.Error)
		assert.True(t, result.IsCreated)
		assert.NotNil(t, result.Object)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("first or create with multiple where conditions", func(t *testing.T) {
		// 查询返回空
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(11, 1))

		user := &TestUser{
			Name:  "Multi Condition",
			Email: "multi@example.com",
			Age:   30,
		}

		result := model.NewScoop().Where("name", "Multi Condition").
			Where("age", 30).
			FirstOrCreate(user)
		assert.NoError(t, result.Error)
		assert.True(t, result.IsCreated)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("first or create with like condition", func(t *testing.T) {
		// 查询返回空
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(12, 1))

		user := &TestUser{
			Name:  "Like Test",
			Email: "like@example.com",
			Age:   28,
		}

		result := model.NewScoop().Like("name", "Like").FirstOrCreate(user)
		assert.NoError(t, result.Error)
		assert.True(t, result.IsCreated)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestModelScoop_CreateOrUpdateExtended 测试 CreateOrUpdate 的扩展场景
func TestModelScoop_CreateOrUpdateExtended(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("create or update with in condition", func(t *testing.T) {
		// 查询返回空，需要创建
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(20, 1))

		user := &TestUser{
			Name:  "In Test",
			Email: "intest@example.com",
			Age:   35,
		}

		values := map[string]interface{}{
			"name":  "In Test",
			"email": "intest@example.com",
			"age":   35,
		}

		result := model.NewScoop().In("id", []int{1, 2, 3}).CreateOrUpdate(values, user)
		assert.NoError(t, result.Error)
		assert.True(t, result.Created)
		assert.False(t, result.Updated)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("create or update with between condition", func(t *testing.T) {
		// 查询返回空
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(21, 1))

		user := &TestUser{
			Name:  "Between Test",
			Email: "between@example.com",
			Age:   40,
		}

		values := map[string]interface{}{
			"name": "Between Test",
			"age":  40,
		}

		result := model.NewScoop().Between("age", 30, 50).CreateOrUpdate(values, user)
		assert.NoError(t, result.Error)
		assert.True(t, result.Created)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("create or update with complex values", func(t *testing.T) {
		// 查询返回空
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(22, 1))

		user := &TestUser{
			Name:  "Complex Values",
			Email: "complex@example.com",
			Age:   45,
		}

		values := map[string]interface{}{
			"name":       "Complex Values",
			"email":      "complex@example.com",
			"age":        45,
			"created_at": "1234567890",
			"updated_at": "1234567890",
		}

		result := model.NewScoop().Where("name", "Complex Values").CreateOrUpdate(values, user)
		assert.NoError(t, result.Error)
		assert.True(t, result.Created)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

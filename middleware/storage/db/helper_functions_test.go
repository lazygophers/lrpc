package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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
	defer client.MockDB().Mock.ExpectClose()

	t.Run("Create single user", func(t *testing.T) {
		// Expect INSERT query
		client.MockDB().Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1)) // lastInsertID=1, rowsAffected=1

		model := db.NewModel[TestUser](client)
		user := &TestUser{
			Name: "Test User",
			Age:  25,
		}

		err := model.NewScoop().Create(user)
		assert.NoError(t, err)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Create user with SQLite dialect", func(t *testing.T) {
		sqliteConfig := &db.Config{
			Type: db.Sqlite,
			Mock: true,
		}

		sqliteClient, err := db.New(sqliteConfig)
		assert.NoError(t, err)
		defer sqliteClient.MockDB().Mock.ExpectClose()

		// Expect INSERT query
		sqliteClient.MockDB().Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(2, 1)) // lastInsertID=2, rowsAffected=1

		model := db.NewModel[TestUser](sqliteClient)
		user := &TestUser{
			Name: "SQLite User",
			Age:  30,
		}

		err = model.NewScoop().Create(user)
		assert.NoError(t, err)

		err = sqliteClient.ExpectationsWereMet()
		assert.NoError(t, err)
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
	defer client.MockDB().Mock.ExpectClose()

	t.Run("Find with snake_case columns triggers scanRowsInto", func(t *testing.T) {
		// 测试带 snake_case 字段名的查询（触发 getCachedFieldName）
		client.MockDB().Mock.ExpectQuery("SELECT \\* FROM `test_users`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age", "created_at", "updated_at", "deleted_at"}).
				AddRow(int64(1), "User1", "user1@example.com", 25, int64(1000), int64(2000), int64(0)))

		model := db.NewModel[TestUser](client)
		user, err := model.NewScoop().First()
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(1), user.Id)
		assert.Equal(t, "User1", user.Name)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Find with multiple rows", func(t *testing.T) {
		client.MockDB().Mock.ExpectQuery("SELECT \\* FROM `test_users`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age", "created_at", "updated_at", "deleted_at"}).
				AddRow(int64(1), "User1", "user1@example.com", 25, int64(1000), int64(2000), int64(0)).
				AddRow(int64(2), "User2", "user2@example.com", 30, int64(1001), int64(2001), int64(0)).
				AddRow(int64(3), "User3", "user3@example.com", 35, int64(1002), int64(2002), int64(0)))

		model := db.NewModel[TestUser](client)
		users, err := model.NewScoop().Find()
		assert.NoError(t, err)
		assert.Len(t, users, 3)
		assert.Equal(t, int64(1), users[0].Id)
		assert.Equal(t, int64(2), users[1].Id)
		assert.Equal(t, int64(3), users[2].Id)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestHelperFunctions_FirstWithConditions 测试 First 操作触发 scanRowsInto
func TestHelperFunctions_FirstWithConditions(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("First with Where conditions", func(t *testing.T) {
		client.MockDB().Mock.ExpectQuery("SELECT \\* FROM `test_users` WHERE.*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age", "created_at", "updated_at", "deleted_at"}).
				AddRow(int64(1), "User1", "user1@example.com", 25, int64(1000), int64(2000), int64(0)))

		model := db.NewModel[TestUser](client)
		user, err := model.NewScoop().Where("id", 1).First()
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(1), user.Id)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
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
	defer client.MockDB().Mock.ExpectClose()

	t.Run("Create multiple users", func(t *testing.T) {
		// 批量插入
		client.MockDB().Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 3)) // lastInsertID=1, rowsAffected=3

		model := db.NewModel[TestUser](client)
		users := []*TestUser{
			{Name: "User1", Age: 25},
			{Name: "User2", Age: 30},
			{Name: "User3", Age: 35},
		}

		result := model.NewScoop().CreateInBatches(users, 50)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(3), result.RowsAffected)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestHelperFunctions_CreateOrUpdate 测试 CreateOrUpdate 触发相关逻辑
func TestHelperFunctions_CreateOrUpdate(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	t.Run("CreateOrUpdate with new record", func(t *testing.T) {
		// 首先查询记录是否存在
		client.MockDB().Mock.ExpectQuery("SELECT.*FROM `test_users` WHERE.*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"})) // 空结果

		// 插入新记录
		client.MockDB().Mock.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		model := db.NewModel[TestUser](client)
		user := &TestUser{
			Name: "New User",
			Age:  25,
		}

		result := model.NewScoop().Where("name", "New User").CreateOrUpdate(map[string]interface{}{
			"name": "New User",
			"age":  25,
		}, user)
		assert.NoError(t, result.Error)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestHelperFunctions_DifferentDialects 测试不同数据库方言
func TestHelperFunctions_DifferentDialects(t *testing.T) {
	dialects := []struct {
		name        string
		dbType      string
		lastIDQuery string
	}{
		{"MySQL", db.MySQL, "SELECT LAST_INSERT_ID"},
		{"SQLite", db.Sqlite, "SELECT last_insert_rowid"},
	}

	for _, tt := range dialects {
		t.Run(tt.name+" last insert ID", func(t *testing.T) {
			config := &db.Config{
				Type: tt.dbType,
				Mock: true,
			}

			client, err := db.New(config)
			assert.NoError(t, err)
			defer client.MockDB().Mock.ExpectClose()

			// Expect INSERT query
			client.MockDB().Mock.ExpectExec("INSERT INTO test_users").
				WillReturnResult(sqlmock.NewResult(1, 1))

			model := db.NewModel[TestUser](client)
			user := &TestUser{
				Name: tt.name + " User",
				Age:  25,
			}

			err = model.NewScoop().Create(user)
			assert.NoError(t, err)

			err = client.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}

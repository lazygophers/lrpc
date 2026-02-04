package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestModelScoop_ConditionMethods 测试 ModelScoop 的条件方法
func TestModelScoop_ConditionMethods(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("Or method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1"))

		users, err := model.NewScoop().Where("id", 1).Or("email", "test@example.com").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("NotEqual method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(2), "User2"))

		users, err := model.NewScoop().NotEqual("name", "Admin").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("In method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1").
				AddRow(int64(2), "User2"))

		users, err := model.NewScoop().In("id", []int{1, 2, 3}).Find()
		assert.NoError(t, err)
		assert.Len(t, users, 2)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("NotIn method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(4), "User4"))

		users, err := model.NewScoop().NotIn("id", []int{1, 2, 3}).Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Like method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "John Doe"))

		users, err := model.NewScoop().Like("name", "John").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("LeftLike method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "Johnson"))

		users, err := model.NewScoop().LeftLike("name", "John").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("RightLike method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "John Doe"))

		users, err := model.NewScoop().RightLike("name", "Doe").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("NotLike method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "Jane"))

		users, err := model.NewScoop().NotLike("name", "John").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Between method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1").
				AddRow(int64(2), "User2"))

		users, err := model.NewScoop().Between("age", 20, 30).Find()
		assert.NoError(t, err)
		assert.Len(t, users, 2)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("NotBetween method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1"))

		users, err := model.NewScoop().NotBetween("age", 20, 30).Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Group method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users .* GROUP BY .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1"))

		users, err := model.NewScoop().Group("age").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Order method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users .* ORDER BY .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1"))

		users, err := model.NewScoop().Order("age DESC").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Desc method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users .* ORDER BY .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1"))

		users, err := model.NewScoop().Desc("id").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Asc method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users .* ORDER BY .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1"))

		users, err := model.NewScoop().Asc("id").Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Unscoped method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1"))

		users, err := model.NewScoop().Unscoped().Where("id", 1).Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("Ignore method", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1"))

		users, err := model.NewScoop().Ignore(true).Where("id", 1).Find()
		assert.NoError(t, err)
		assert.Len(t, users, 1)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestModelScoop_FirstOrCreate 测试 FirstOrCreate 方法
func TestModelScoop_FirstOrCreate(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("create when not exists", func(t *testing.T) {
		// 第一次查询未找到记录
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		user := &TestUser{Name: "New User", Email: "new@example.com", Age: 25}
		result := model.NewScoop().Where("email", "new@example.com").FirstOrCreate(user)
		assert.NoError(t, result.Error)
		assert.True(t, result.IsCreated)
		assert.NotNil(t, result.Object)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("return existing record", func(t *testing.T) {
		// 查询到已存在的记录
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}).
				AddRow(int64(1), "Existing User", "existing@example.com", int64(30)))

		user := &TestUser{Name: "Existing User", Email: "existing@example.com", Age: 30}
		result := model.NewScoop().Where("email", "existing@example.com").FirstOrCreate(user)
		assert.NoError(t, result.Error)
		assert.False(t, result.IsCreated)
		assert.NotNil(t, result.Object)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestModelScoop_Chunk 测试 Chunk 方法
func TestModelScoop_Chunk(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("chunk successfully", func(t *testing.T) {
		// 第一次分块查询 (offset 0, limit 3)
		client.ExpectQuery("SELECT \\* FROM test_users .* LIMIT 3").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(1), "User1").
				AddRow(int64(2), "User2").
				AddRow(int64(3), "User3"))

		// 第二次分块查询 (offset 3, limit 3)
		client.ExpectQuery("SELECT \\* FROM test_users .* LIMIT 3 OFFSET 3").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(4), "User4").
				AddRow(int64(5), "User5").
				AddRow(int64(6), "User6"))

		// 第三次分块查询 (offset 6, limit 3)
		client.ExpectQuery("SELECT \\* FROM test_users .* LIMIT 3 OFFSET 6").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(7), "User7").
				AddRow(int64(8), "User8").
				AddRow(int64(9), "User9"))

		// 第四次分块查询 (offset 9, limit 3)
		client.ExpectQuery("SELECT \\* FROM test_users .* LIMIT 3 OFFSET 9").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
				AddRow(int64(10), "User10"))

		// 第五次查询返回空，结束循环
		client.ExpectQuery("SELECT \\* FROM test_users .* LIMIT 3 OFFSET 12").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

		chunkCount := 0
		totalUsers := 0

		result := model.NewScoop().Chunk(3, func(tx *db.Scoop, users []*TestUser, offset uint64) error {
			chunkCount++
			totalUsers += len(users)
			return nil
		})

		assert.NoError(t, result.Error)
		assert.Equal(t, 4, chunkCount)
		assert.Equal(t, 10, totalUsers)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("chunk with empty result", func(t *testing.T) {
		// 第一次查询就返回空
		client.ExpectQuery("SELECT \\* FROM test_users .* LIMIT 10").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

		chunkCount := 0

		result := model.NewScoop().Chunk(10, func(tx *db.Scoop, users []*TestUser, offset uint64) error {
			chunkCount++
			return nil
		})

		assert.NoError(t, result.Error)
		assert.Equal(t, 0, chunkCount)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestModelScoop_Scan 测试 Scan 方法
func TestModelScoop_Scan(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("scan into slice successfully", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}).
				AddRow(int64(1), "User1", "user1@example.com", int64(25)).
				AddRow(int64(2), "User2", "user2@example.com", int64(30)))

		var users []TestUser
		result := model.NewScoop().Where("age >=", 25).Scan(&users)
		assert.NoError(t, result.Error)
		assert.Len(t, users, 2)
		assert.Equal(t, 1, users[0].Id)
		assert.Equal(t, "User1", users[0].Name)
		assert.Equal(t, 2, users[1].Id)
		assert.Equal(t, "User2", users[1].Name)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("scan with empty result", func(t *testing.T) {
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		var users []TestUser
		result := model.NewScoop().Where("id", 999).Scan(&users)
		assert.NoError(t, result.Error)
		assert.Len(t, users, 0)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestModelScoop_UpdateOrCreate 测试 UpdateOrCreate 方法
func TestModelScoop_UpdateOrCreate(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("update existing record", func(t *testing.T) {
		// 查询到已存在的记录
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}).
				AddRow(int64(1), "Old Name", "user@example.com", int64(25)))

		// 更新记录
		client.ExpectExec("UPDATE test_users SET .* WHERE .*").
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 查询更新后的记录
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}).
				AddRow(int64(1), "Updated Name", "user@example.com", int64(30)))

		user := &TestUser{Id: 1, Name: "Updated Name", Email: "user@example.com", Age: 30}
		values := map[string]interface{}{
			"name": "Updated Name",
			"age":  30,
		}
		result := model.NewScoop().Where("id", 1).UpdateOrCreate(values, user)
		assert.NoError(t, result.Error)
		assert.False(t, result.IsCreated)
		assert.NotNil(t, result.Object)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("create new record", func(t *testing.T) {
		// 未查询到记录
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(2, 1))

		user := &TestUser{Name: "New User", Email: "newuser@example.com", Age: 28}
		values := map[string]interface{}{
			"name":  "New User",
			"email": "newuser@example.com",
			"age":   28,
		}
		result := model.NewScoop().Where("email", "newuser@example.com").UpdateOrCreate(values, user)
		assert.NoError(t, result.Error)
		assert.True(t, result.IsCreated)
		assert.NotNil(t, result.Object)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

// TestModelScoop_CreateOrUpdate 测试 CreateOrUpdate 方法
func TestModelScoop_CreateOrUpdate(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Mock.ExpectClose()

	model := db.NewModel[TestUser](client)

	t.Run("create if not exists", func(t *testing.T) {
		// 未查询到记录
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}))

		// 插入新记录
		client.ExpectExec("INSERT INTO test_users").
			WillReturnResult(sqlmock.NewResult(1, 1))

		user := &TestUser{Name: "User1", Email: "user1@example.com", Age: 25}
		values := map[string]interface{}{
			"name":  "User1",
			"email": "user1@example.com",
			"age":   25,
		}
		result := model.NewScoop().Where("email", "user1@example.com").CreateOrUpdate(values, user)
		assert.NoError(t, result.Error)
		assert.True(t, result.Created)
		assert.False(t, result.Updated)
		assert.NotNil(t, result.Object)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("update if exists", func(t *testing.T) {
		// 查询到已存在的记录
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}).
				AddRow(int64(5), "Old Name", "user5@example.com", int64(25)))

		// 更新记录
		client.ExpectExec("UPDATE test_users SET .* WHERE .*").
			WillReturnResult(sqlmock.NewResult(5, 1))

		// 查询更新后的记录
		client.ExpectQuery("SELECT \\* FROM test_users WHERE .* LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "age"}).
				AddRow(int64(5), "Updated User", "user5@example.com", int64(35)))

		user := &TestUser{Id: 5, Name: "Updated User", Email: "user5@example.com", Age: 35}
		values := map[string]interface{}{
			"name": "Updated User",
			"age":  35,
		}
		result := model.NewScoop().Where("id", 5).CreateOrUpdate(values, user)
		assert.NoError(t, result.Error)
		assert.False(t, result.Created)
		assert.True(t, result.Updated)
		assert.NotNil(t, result.Object)

		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

package db_test

import (
"errors"
"github.com/DATA-DOG/go-sqlmock"
"github.com/lazygophers/lrpc/middleware/storage/db"
"github.com/lazygophers/utils/xerror"
"github.com/stretchr/testify/assert"
"testing"
)

// TestScoop_MockBasicOperations 测试 Scoop 的基本操作（使用 Mock）
func TestScoop_MockBasicOperations(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)

	t.Run("test Count operation", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM test_users").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

		count, err := client.NewScoop().Model(TestUser{}).Count()
		assert.NoError(t, err)
		assert.Equal(t, uint64(10), count)
	})

	t.Run("test Exist operation", func(t *testing.T) {
		mockDB.Mock.ExpectQuery("SELECT id FROM test_users WHERE deleted_at = 0 LIMIT 1 OFFSET 0").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		exist, err := client.NewScoop().Model(TestUser{}).Exist()
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	mockDB.Mock.ExpectClose()
	err = mockDB.Close()
	assert.NoError(t, err)

	err = client.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestClient_PingAndClose 测试 Client 的 Ping 和 Close 方法
func TestClient_PingAndClose(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)

	t.Run("test Ping", func(t *testing.T) {
		mockDB.Mock.ExpectPing()
		err := client.Ping()
		assert.NoError(t, err)
	})

	t.Run("test Close", func(t *testing.T) {
		mockDB.Mock.ExpectClose()
		err := client.Close()
		assert.NoError(t, err)
	})

	err = client.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestScoop_Conditions 测试各种查询条件
func TestScoop_Conditions(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)

	t.Run("test NotEqual condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).NotEqual("age", 25)
		assert.NotNil(t, scoop)
	})

	t.Run("test NotIn condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).NotIn("id", []int{1, 2, 3})
		assert.NotNil(t, scoop)
	})

	t.Run("test LeftLike condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).LeftLike("name", "Alice")
		assert.NotNil(t, scoop)
	})

	t.Run("test RightLike condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).RightLike("name", "Alice")
		assert.NotNil(t, scoop)
	})

	t.Run("test NotLike condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).NotLike("name", "Alice")
		assert.NotNil(t, scoop)
	})

	t.Run("test NotBetween condition", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).NotBetween("age", 20, 30)
		assert.NotNil(t, scoop)
	})

	t.Run("test Unscoped", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Unscoped()
		assert.NotNil(t, scoop)
	})

	t.Run("test Limit and Offset", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Limit(10).Offset(5)
		assert.NotNil(t, scoop)
	})

	t.Run("test Group", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Group("age")
		assert.NotNil(t, scoop)
	})

	t.Run("test Order", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Order("age DESC")
		assert.NotNil(t, scoop)
	})

	t.Run("test Select", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Select("id", "name")
		assert.NotNil(t, scoop)
	})

	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestScoop_Joins 测试 JOIN 操作
func TestScoop_Joins(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)

	t.Run("test InnerJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			InnerJoin("orders", "users.id = orders.user_id")
		assert.NotNil(t, scoop)
	})

	t.Run("test LeftJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			LeftJoin("orders", "users.id = orders.user_id")
		assert.NotNil(t, scoop)
	})

	t.Run("test RightJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			RightJoin("orders", "users.id = orders.user_id")
		assert.NotNil(t, scoop)
	})

	t.Run("test FullJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			FullJoin("orders", "users.id = orders.user_id")
		assert.NotNil(t, scoop)
	})

	t.Run("test CrossJoin", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).
			CrossJoin("orders")
		assert.NotNil(t, scoop)
	})

	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestScoop_Having 测试 HAVING 子句
func TestScoop_Having(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)

	scoop := client.NewScoop().Model(TestUser{}).
		Group("age").
		Having("COUNT(*) > ?", 5)
	assert.NotNil(t, scoop)

	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestScoop_Ignore 测试 Ignore 方法
func TestScoop_Ignore(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)

	t.Run("test Ignore with default", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Ignore()
		assert.NotNil(t, scoop)
	})

	t.Run("test Ignore with true", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Ignore(true)
		assert.NotNil(t, scoop)
	})

	t.Run("test Ignore with false", func(t *testing.T) {
		scoop := client.NewScoop().Model(TestUser{}).Ignore(false)
		assert.NotNil(t, scoop)
	})

	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestScoop_TableMethod 测试 Table 方法的覆盖率
func TestScoop_TableMethod(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("Table with valid name", func(t *testing.T) {
		scoop := client.NewScoop().Table("test_users")
		assert.NotNil(t, scoop)
	})

	t.Run("Table with invalid name containing spaces", func(t *testing.T) {
		scoop := client.NewScoop().Table("invalid table name")
		assert.NotNil(t, scoop)
		// 应该记录错误但不返回 nil
	})

	t.Run("Table with empty name", func(t *testing.T) {
		scoop := client.NewScoop().Table("")
		assert.NotNil(t, scoop)
	})

	mockDB.Mock.ExpectClose()
	err = mockDB.Close()
	assert.NoError(t, err)

	err = client.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestScoop_Create_DuplicateKeyError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("Create with duplicate key error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnError(xerror.New(xerror.CodeConflict, "duplicate key error"))

		scoop := client.NewScoop().Table("test_users")
		user := &TestUser{Name: "Test", Email: "test@example.com", Age: 25}
		result := scoop.Create(user)

		// 应该有错误
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "duplicate key error")

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})
}

// TestScoop_IsDuplicatedKeyError 测试 IsDuplicatedKeyError 方法
func TestScoop_IsDuplicatedKeyError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("IsDuplicatedKeyError with ErrConflict", func(t *testing.T) {
		scoop := client.NewScoop()
		err := xerror.New(xerror.CodeConflict, "duplicate key")
		assert.True(t, scoop.IsDuplicatedKeyError(err))
	})

	t.Run("IsDuplicatedKeyError with other error", func(t *testing.T) {
		scoop := client.NewScoop()
		err := errors.New("some other error")
		assert.False(t, scoop.IsDuplicatedKeyError(err))
	})

	t.Run("IsDuplicatedKeyError with nil", func(t *testing.T) {
		scoop := client.NewScoop()
		assert.False(t, scoop.IsDuplicatedKeyError(nil))
	})
}

// TestScoop_Update_DuplicateKeyError 测试 Update 方法处理重复键错误
func TestScoop_Update_DuplicateKeyError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("Updates with duplicate key error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("UPDATE test_users.*").
			WillReturnError(xerror.New(xerror.CodeConflict, "duplicate key error"))

		scoop := client.NewScoop().Table("test_users")
		result := scoop.Where("id", 1).Updates("name", "Updated")

		// 应该有错误
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "duplicate key error")

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})
}

// TestScoop_CreateInBatches_DuplicateKeyError 测试批量创建时的重复键错误
func TestScoop_CreateInBatches_DuplicateKeyError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("CreateInBatches with duplicate key error", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO test_users").
			WillReturnError(xerror.New(xerror.CodeConflict, "duplicate key error"))

		scoop := client.NewScoop().Table("test_users")
		users := []*TestUser{
			{Name: "User1", Email: "user1@example.com", Age: 25},
			{Name: "User2", Email: "user2@example.com", Age: 26},
		}
		result := scoop.CreateInBatches(users, 50)

		// 应该有错误
		assert.Error(t, result.Error)

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})
}

// TestScoop_Create_TableName 测试使用 Table 方法的 Create
func TestScoop_Create_TableName(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	mockDB := client.MockDB()
	assert.NoError(t, err)
	defer mockDB.Close()

	t.Run("Create with Table method", func(t *testing.T) {
		mockDB.Mock.ExpectExec("INSERT INTO custom_table").
			WillReturnResult(sqlmock.NewResult(1, 1))

		scoop := client.NewScoop().Table("custom_table")
		user := &TestUser{Name: "Test", Email: "test@example.com", Age: 25}
		result := scoop.Create(user)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), result.RowsAffected)

		mockDB.Mock.ExpectClose()
		err = mockDB.Close()
		assert.NoError(t, err)
	})
}

func TestScoop_Equal(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Equal with int value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Equal("id", 1)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Equal with string value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Equal("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Equal with nil value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Equal("deleted_at", nil)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Chaining Equal calls", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Equal("id", 1).Equal("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotEqual 测试 NotEqual 方法
func TestScoop_NotEqual(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotEqual with int value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotEqual("status", 0)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotEqual with string value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotEqual("name", "deleted")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

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

func TestScoop_In(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("In with slice", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.In("id", []int{1, 2, 3})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("In with empty slice", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.In("id", []int{})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotIn 测试 NotIn 方法
func TestScoop_NotIn(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotIn with slice", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotIn("id", []int{1, 2, 3})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotIn with empty slice", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotIn("id", []int{})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Between 测试 Between 方法
func TestScoop_Between(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Between with int values", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Between("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Between with string values", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Between("name", "A", "Z")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

func TestScoop_LikeMethods(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Like with normal value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Like("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("LeftLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.LeftLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("RightLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.RightLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

func TestScoop_NotBetween(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotBetween with int values", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotBetween("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotBetween with string values", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotBetween("name", "A", "Z")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotLeftLike 测试 NotLeftLike 方法
func TestScoop_NotLeftLike(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotLeftLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotLeftLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotRightLike 测试 NotRightLike 方法
func TestScoop_NotRightLike(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotRightLike", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotRightLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

func TestScoop_Select(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Select with multiple fields", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Select("id", "name", "email")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Select with single field", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Select("id")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Chaining Select calls", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Select("id").Select("name")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Limit 测试 Limit 方法
func TestScoop_Limit(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Limit with positive value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Limit(10)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Limit with zero value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Limit(0)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Offset 测试 Offset 方法
func TestScoop_Offset(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Offset with positive value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Offset(5)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Offset with zero value", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Offset(0)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

func TestScoop_Where_Combinations(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Where with map", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Where(map[string]interface{}{
			"name": "test",
			"age":  25,
		})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Where with multiple args", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Where("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_NotEqual_NotIn 测试 NotEqual 和 NotIn 组合
func TestScoop_NotEqual_NotIn(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NotEqual with different types", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotEqual("status", 0).NotEqual("deleted", 1)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotIn with multiple slices", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotIn("id", []int{1, 2}).NotIn("status", []int{0, 1})
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Like_Between 测试 Like 和 Between 组合
func TestScoop_Like_Between(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Like and Between combination", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Like("name", "test").Between("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotLike and NotBetween combination", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotLike("name", "test").NotBetween("age", 18, 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_In_Like_Mixed 测试混合条件
func TestScoop_In_Like_Mixed(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("In and Like combination", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.In("id", []int{1, 2, 3}).Like("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("NotIn and NotLike combination", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.NotIn("id", []int{1, 2, 3}).NotLike("name", "test")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// TestScoop_Where_SubConditions 测试 Where 子条件
func TestScoop_Where_SubConditions(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Where with operator", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Where("age", ">", 18)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})

	t.Run("Where with multiple operators", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Where("age", ">", 18).Where("age", "<", 65)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

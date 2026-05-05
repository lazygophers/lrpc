package db_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"github.com/stretchr/testify/assert"
)

// TestNewMock_MySQL 测试 MySQL mock
func TestNewMock_MySQL(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, db.MySQL, client.DriverType())

	// 方式1: 使用 client.MockDB() 访问 MockDB
	mockDB := client.MockDB()
	assert.NotNil(t, mockDB)

	// 设置查询期望
	mockDB.Mock.ExpectQuery("SELECT (.+) FROM `test_users`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(1, "Alice", 25))

	// 执行查询
	var user TestUser
	err = client.Database().Table("test_users").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.Id)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, 25, user.Age)

	// 验证所有期望都被满足
	err = client.ExpectationsWereMet()
	assert.NoError(t, err)

	// 设置 Close 期望
	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestNewMock_Postgres 测试 Postgres mock
func TestNewMock_Postgres(t *testing.T) {
	config := &db.Config{
		Type: db.Postgres,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, db.Postgres, client.DriverType())

	mockDB := client.MockDB()

	// 设置插入期望 - GORM 使用 RETURNING 子句返回插入的 ID
	// 使用 AnyArg() 匹配动态生成的时间戳值
	mockDB.Mock.ExpectQuery("INSERT INTO \"test_users\"").
		WithArgs("Bob", "bob@example.com", 30, sqlmock.AnyArg(), sqlmock.AnyArg(), 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	// 执行插入
	user := TestUser{Name: "Bob", Age: 30, Email: "bob@example.com"}
	err = client.Database().Table("test_users").Create(&user).Error
	assert.NoError(t, err)

	// 验证所有期望都被满足
	err = client.ExpectationsWereMet()
	assert.NoError(t, err)

	// 设置 Close 期望
	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestNewMock_SQLite 测试 SQLite mock
func TestNewMock_SQLite(t *testing.T) {
	config := &db.Config{
		Type: db.Sqlite,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, db.Sqlite, client.DriverType())

	mockDB := client.MockDB()

	// SQLite mock 使用 postgres dialector，所以 SQL 语法类似 postgres
	mockDB.Mock.ExpectQuery("SELECT (.+) FROM \"test_users\"").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(3, "Charlie", 35))

	// 执行查询
	var user TestUser
	err = client.Database().Table("test_users").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(3), user.Id)
	assert.Equal(t, "Charlie", user.Name)
	assert.Equal(t, 35, user.Age)

	// 验证所有期望都被满足
	err = client.ExpectationsWereMet()
	assert.NoError(t, err)

	// 设置 Close 期望
	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestNewMock_TiDB 测试 TiDB mock (MySQL 兼容)
func TestNewMock_TiDB(t *testing.T) {
	config := &db.Config{
		Type: db.TiDB,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, db.TiDB, client.DriverType())

	mockDB := client.MockDB()

	// TiDB 使用 MySQL dialector
	mockDB.Mock.ExpectQuery("SELECT (.+) FROM `test_users`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(4, "David", 40))

	// 执行查询
	var user TestUser
	err = client.Database().Table("test_users").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(4), user.Id)
	assert.Equal(t, "David", user.Name)

	// 验证所有期望都被满足
	err = client.ExpectationsWereMet()
	assert.NoError(t, err)

	// 设置 Close 期望
	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestNewMock_UnsupportedType 测试不支持的数据库类型
func TestNewMock_UnsupportedType(t *testing.T) {
	config := &db.Config{
		Type: "unsupported",
		Mock: true,
	}

	client, err := db.New(config)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "unsupported database type")
}

// TestNew_WithMockConfig 测试通过 New 函数使用 mock 配置
func TestNew_WithMockConfig(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, db.MySQL, client.DriverType())

	// 注意：通过 New 函数创建的 mock 客户端无法直接访问 mockDB
	// 如果需要设置期望，应该使用 NewMock 函数
}

// TestClient_ExpectClose 测试 ExpectClose 方法
func TestClient_ExpectClose(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// 使用 client.ExpectClose() 而不是 client.MockDB().Mock.ExpectClose()
	expectedClose := client.ExpectClose()
	assert.NotNil(t, expectedClose)

	// 关闭 mock 连接以满足期望
	client.MockDB().Close()

	// 验证期望被满足
	err = client.ExpectationsWereMet()
	assert.NoError(t, err)
}

// TestClient_ExpectCloseWithNilMockDB 测试 mockDB 为 nil 时的情况
func TestClient_ExpectCloseWithNilMockDB(t *testing.T) {
	// 创建一个非 mock 的客户端
	config := &db.Config{
		Type: db.MySQL,
		Mock: false,
	}

	client, err := db.New(config)
	// 由于没有真实的数据库连接，这会失败
	// 但我们可以测试 ExpectClose 在 nil mockDB 的情况
	if err != nil {
		// 如果创建失败，说明没有真实数据库，这是预期的
		t.Skip("No real database available for testing")
	}

	if client != nil && client.MockDB() == nil {
		// 测试 nil mockDB 情况
		expectedClose := client.ExpectClose()
		assert.Nil(t, expectedClose)
	}
}

// TestExpectedQuery_WithArgs 测试 ExpectedQuery.WithArgs 方法
func TestExpectedQuery_WithArgs(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("WithArgs on ExpectQuery", func(t *testing.T) {
		// 测试 WithArgs 方法
		expectedQuery := client.ExpectQuery("SELECT \\* FROM users WHERE id = ?")
		expectedQuery.WithArgs(1)
		assert.NotNil(t, expectedQuery)
	})
}

// TestExpectedQuery_WillReturnError 测试 ExpectedQuery.WillReturnError 方法
func TestExpectedQuery_WillReturnError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("WillReturnError on ExpectQuery", func(t *testing.T) {
		expectedQuery := client.ExpectQuery("SELECT \\* FROM users")
		expectedQuery.WillReturnError(assert.AnError)
		assert.NotNil(t, expectedQuery)
	})
}

// TestExpectedExec_WithArgs 测试 ExpectedExec.WithArgs 方法
func TestExpectedExec_WithArgs(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("WithArgs on ExpectExec", func(t *testing.T) {
		expectedExec := client.ExpectExec("INSERT INTO users")
		expectedExec.WithArgs("John", 25)
		assert.NotNil(t, expectedExec)
	})
}

// TestExpectedExec_WillReturnError 测试 ExpectedExec.WillReturnError 方法
func TestExpectedExec_WillReturnError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("WillReturnError on ExpectExec", func(t *testing.T) {
		expectedExec := client.ExpectExec("INSERT INTO users")
		expectedExec.WillReturnError(assert.AnError)
		assert.NotNil(t, expectedExec)
	})
}

// TestClient_Database 测试 Database 方法
func TestClient_Database(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("Database returns gorm.DB", func(t *testing.T) {
		db := client.Database()
		assert.NotNil(t, db)
	})
}

// TestExpectedBegin_WillReturnError 测试 ExpectedBegin.WillReturnError
func TestExpectedBegin_WillReturnError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("WillReturnError on ExpectBegin", func(t *testing.T) {
		expectedBegin := client.ExpectBegin()
		expectedBegin.WillReturnError(assert.AnError)
		assert.NotNil(t, expectedBegin)
	})
}

// TestExpectedCommit_WillReturnError 测试 ExpectedCommit.WillReturnError
func TestExpectedCommit_WillReturnError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("WillReturnError on ExpectCommit", func(t *testing.T) {
		expectedCommit := client.ExpectCommit()
		expectedCommit.WillReturnError(assert.AnError)
		assert.NotNil(t, expectedCommit)
	})
}

// TestExpectedRollback_WillReturnError 测试 ExpectedRollback.WillReturnError
func TestExpectedRollback_WillReturnError(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("WillReturnError on ExpectRollback", func(t *testing.T) {
		expectedRollback := client.ExpectRollback()
		expectedRollback.WillReturnError(assert.AnError)
		assert.NotNil(t, expectedRollback)
	})
}


// TestClient_NewScoop 测试 NewScoop 方法
func TestClient_NewScoop(t *testing.T) {
	config := &db.Config{
		Type: db.MySQL,
		Mock: true,
	}

	client, err := db.New(config)
	assert.NoError(t, err)
	defer client.MockDB().Close()

	t.Run("NewScoop returns Scoop", func(t *testing.T) {
		scoop := client.NewScoop()
		assert.NotNil(t, scoop)
	})
}

// TestMockDB_ExpectationsWereMet 测试 ExpectationsWereMet 方法
func TestMockDB_ExpectationsWereMet(t *testing.T) {
	t.Run("ExpectationsWereMet with unmet expectations", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()
		defer client.MockDB().Close()

		// 设置一个期望但不满足它
		client.ExpectQuery("SELECT * FROM users")
		// 不执行任何查询，期望不会被满足
		err = client.ExpectationsWereMet()
		assert.Error(t, err) // 应该有错误因为期望未满足
	})

	t.Run("ExpectationsWereMet with all expectations met", func(t *testing.T) {
		config := &db.Config{
			Type: db.MySQL,
			Mock: true,
		}

		client, err := db.New(config)
		assert.NoError(t, err)
		defer client.MockDB().Mock.ExpectClose()
		defer client.MockDB().Close()

		// 不设置任何期望，应该通过
		err = client.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

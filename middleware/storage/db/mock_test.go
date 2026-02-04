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
	assert.Equal(t, 1, user.Id)
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
	assert.Equal(t, 3, user.Id)
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
	assert.Equal(t, 4, user.Id)
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

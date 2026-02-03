package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestUser 测试用的用户模型
type TestUser struct {
	ID   int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name string `gorm:"column:name;type:varchar(100)"`
	Age  int    `gorm:"column:age"`
}

func (TestUser) TableName() string {
	return "test_users"
}

// TestNewMock_MySQL 测试 MySQL mock
func TestNewMock_MySQL(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, mockDB)
	assert.Equal(t, MySQL, client.DriverType())

	// 设置查询期望
	mockDB.Mock.ExpectQuery("SELECT (.+) FROM `test_users`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(1, "Alice", 25))

	// 执行查询
	var user TestUser
	err = client.Database().Table("test_users").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, 25, user.Age)

	// 验证所有期望都被满足
	err = mockDB.ExpectationsWereMet()
	assert.NoError(t, err)

	// 设置 Close 期望
	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestNewMock_Postgres 测试 Postgres mock
func TestNewMock_Postgres(t *testing.T) {
	config := &Config{
		Type: Postgres,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, mockDB)
	assert.Equal(t, Postgres, client.DriverType())

	// 设置插入期望（GORM 配置了 SkipDefaultTransaction，所以不会自动开启事务）
	mockDB.Mock.ExpectQuery("INSERT INTO \"test_users\"").
		WithArgs("Bob", 30).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	// 执行插入
	user := TestUser{Name: "Bob", Age: 30}
	err = client.Database().Table("test_users").Create(&user).Error
	assert.NoError(t, err)

	// 验证所有期望都被满足
	err = mockDB.ExpectationsWereMet()
	assert.NoError(t, err)

	// 设置 Close 期望
	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestNewMock_SQLite 测试 SQLite mock
func TestNewMock_SQLite(t *testing.T) {
	config := &Config{
		Type: Sqlite,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, mockDB)
	assert.Equal(t, Sqlite, client.DriverType())

	// SQLite mock 使用 postgres dialector，所以 SQL 语法类似 postgres
	mockDB.Mock.ExpectQuery("SELECT (.+) FROM \"test_users\"").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(3, "Charlie", 35))

	// 执行查询
	var user TestUser
	err = client.Database().Table("test_users").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(3), user.ID)
	assert.Equal(t, "Charlie", user.Name)
	assert.Equal(t, 35, user.Age)

	// 验证所有期望都被满足
	err = mockDB.ExpectationsWereMet()
	assert.NoError(t, err)

	// 设置 Close 期望
	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestNewMock_TiDB 测试 TiDB mock (MySQL 兼容)
func TestNewMock_TiDB(t *testing.T) {
	config := &Config{
		Type: TiDB,
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, mockDB)
	assert.Equal(t, TiDB, client.DriverType())

	// TiDB 使用 MySQL dialector
	mockDB.Mock.ExpectQuery("SELECT (.+) FROM `test_users`").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).
			AddRow(4, "David", 40))

	// 执行查询
	var user TestUser
	err = client.Database().Table("test_users").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(4), user.ID)
	assert.Equal(t, "David", user.Name)

	// 验证所有期望都被满足
	err = mockDB.ExpectationsWereMet()
	assert.NoError(t, err)

	// 设置 Close 期望
	mockDB.Mock.ExpectClose()
	mockDB.Close()
}

// TestNewMock_UnsupportedType 测试不支持的数据库类型
func TestNewMock_UnsupportedType(t *testing.T) {
	config := &Config{
		Type: "unsupported",
		Mock: true,
	}

	client, mockDB, err := NewMock(config)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Nil(t, mockDB)
	assert.Contains(t, err.Error(), "unsupported database type")
}

// TestNew_WithMockConfig 测试通过 New 函数使用 mock 配置
func TestNew_WithMockConfig(t *testing.T) {
	config := &Config{
		Type: MySQL,
		Mock: true,
	}

	client, err := New(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, MySQL, client.DriverType())

	// 注意：通过 New 函数创建的 mock 客户端无法直接访问 mockDB
	// 如果需要设置期望，应该使用 NewMock 函数
}

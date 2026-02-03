# Database Mock 使用指南

本文档介绍如何使用 LRPC 数据库中间件的 Mock 功能进行单元测试。

## 概述

Mock 功能基于 [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) 实现，支持在不需要真实数据库的情况下进行单元测试。

## 支持的数据库类型

- **MySQL**: 使用 MySQL dialector
- **PostgreSQL**: 使用 PostgreSQL dialector
- **SQLite**: 使用 PostgreSQL dialector（SQLite 驱动限制）
- **TiDB**: 使用 MySQL dialector（TiDB 兼容 MySQL）
- **GaussDB**: 使用 PostgreSQL dialector（GaussDB 兼容 PostgreSQL）
- **ClickHouse**: 使用 ClickHouse dialector（可能有限制）

## 基本用法

### 方式一：使用 NewMock 函数（推荐）

这种方式可以直接访问 `sqlmock.Sqlmock` 实例，方便设置期望。

```go
package mypackage

import (
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/lazygophers/lrpc/middleware/storage/db"
    "github.com/stretchr/testify/assert"
)

func TestUserRepository(t *testing.T) {
    // 创建 Mock 数据库
    config := &db.Config{
        Type: db.MySQL,
        Mock: true,
    }

    client, mockDB, err := db.NewMock(config)
    assert.NoError(t, err)
    defer mockDB.Close()

    // 设置查询期望
    mockDB.Mock.ExpectQuery("SELECT (.+) FROM `users`").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
            AddRow(1, "Alice", "alice@example.com"))

    // 执行查询
    var user User
    err = client.Database().Table("users").Where("id = ?", 1).First(&user).Error
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)

    // 验证所有期望都被满足
    err = mockDB.ExpectationsWereMet()
    assert.NoError(t, err)
}
```

### 方式二：使用 New 函数

这种方式更简洁，但无法直接访问 `sqlmock.Sqlmock` 实例。

```go
func TestWithNewFunction(t *testing.T) {
    config := &db.Config{
        Type: db.MySQL,
        Mock: true,
    }

    client, err := db.New(config)
    assert.NoError(t, err)

    // 注意：这种方式无法设置 sqlmock 期望
    // 适合不需要精确控制 SQL 行为的场景
}
```

## 常见测试场景

### 1. 测试查询操作

```go
func TestQuery(t *testing.T) {
    config := &db.Config{Type: db.MySQL, Mock: true}
    client, mockDB, err := db.NewMock(config)
    assert.NoError(t, err)

    // MySQL 使用反引号
    mockDB.Mock.ExpectQuery("SELECT (.+) FROM `users`").
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
            AddRow(1, "Alice").
            AddRow(2, "Bob"))

    var users []User
    err = client.Database().Table("users").Find(&users).Error
    assert.NoError(t, err)
    assert.Len(t, users, 2)

    mockDB.Mock.ExpectClose()
    mockDB.Close()
}
```

### 2. 测试插入操作

```go
func TestInsert(t *testing.T) {
    config := &db.Config{Type: db.Postgres, Mock: true}
    client, mockDB, err := db.NewMock(config)
    assert.NoError(t, err)

    // PostgreSQL 使用双引号，并且 GORM 会返回 ID
    mockDB.Mock.ExpectQuery("INSERT INTO \"users\"").
        WithArgs("Charlie", "charlie@example.com").
        WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

    user := User{Name: "Charlie", Email: "charlie@example.com"}
    err = client.Database().Table("users").Create(&user).Error
    assert.NoError(t, err)

    mockDB.Mock.ExpectClose()
    mockDB.Close()
}
```

### 3. 测试更新操作

```go
func TestUpdate(t *testing.T) {
    config := &db.Config{Type: db.MySQL, Mock: true}
    client, mockDB, err := db.NewMock(config)
    assert.NoError(t, err)

    mockDB.Mock.ExpectExec("UPDATE `users` SET `name`=\\?").
        WithArgs("Alice Updated", 1).
        WillReturnResult(sqlmock.NewResult(0, 1))

    err = client.Database().Table("users").
        Where("id = ?", 1).
        Update("name", "Alice Updated").Error
    assert.NoError(t, err)

    mockDB.Mock.ExpectClose()
    mockDB.Close()
}
```

### 4. 测试删除操作

```go
func TestDelete(t *testing.T) {
    config := &db.Config{Type: db.MySQL, Mock: true}
    client, mockDB, err := db.NewMock(config)
    assert.NoError(t, err)

    mockDB.Mock.ExpectExec("DELETE FROM `users`").
        WithArgs(1).
        WillReturnResult(sqlmock.NewResult(0, 1))

    err = client.Database().Table("users").
        Where("id = ?", 1).
        Delete(&User{}).Error
    assert.NoError(t, err)

    mockDB.Mock.ExpectClose()
    mockDB.Close()
}
```

## 不同数据库类型的 SQL 语法差异

### MySQL / TiDB
- 表名和列名使用反引号：`` `users` ``
- 参数占位符：`?`

```go
mockDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE `id` = \\?")
```

### PostgreSQL / GaussDB
- 表名和列名使用双引号：`"users"`
- 参数占位符：`$1`, `$2`, `$3`...

```go
mockDB.Mock.ExpectQuery("SELECT (.+) FROM \"users\" WHERE \"id\" = \\$1")
```

### SQLite
- 由于驱动限制，使用 PostgreSQL dialector
- SQL 语法与 PostgreSQL 相同

## 注意事项

1. **PrepareStmt 已禁用**：Mock 模式下自动禁用预编译语句，以简化测试
2. **AutoMigrate 跳过**：Mock 模式下不会自动执行表迁移，需要手动设置表结构期望
3. **SkipDefaultTransaction**：GORM 配置了跳过默认事务，如需测试事务，需手动开启
4. **正则表达式匹配**：sqlmock 使用正则表达式匹配 SQL，注意转义特殊字符
5. **Close 期望**：如果需要测试 Close 行为，记得添加 `ExpectClose()`

## 完整示例

```go
package repository

import (
    "testing"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/lazygophers/lrpc/middleware/storage/db"
    "github.com/stretchr/testify/assert"
)

type User struct {
    ID    int64  `gorm:"column:id;primaryKey"`
    Name  string `gorm:"column:name"`
    Email string `gorm:"column:email"`
}

func (User) TableName() string {
    return "users"
}

type UserRepository struct {
    client *db.Client
}

func (r *UserRepository) GetByID(id int64) (*User, error) {
    var user User
    err := r.client.Database().Table("users").Where("id = ?", id).First(&user).Error
    return &user, err
}

func TestUserRepository_GetByID(t *testing.T) {
    // 创建 Mock 数据库
    config := &db.Config{
        Type: db.MySQL,
        Mock: true,
    }

    client, mockDB, err := db.NewMock(config)
    assert.NoError(t, err)

    // 创建 repository
    repo := &UserRepository{client: client}

    // 设置期望
    mockDB.Mock.ExpectQuery("SELECT (.+) FROM `users`").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
            AddRow(1, "Alice", "alice@example.com"))

    // 执行测试
    user, err := repo.GetByID(1)
    assert.NoError(t, err)
    assert.Equal(t, int64(1), user.ID)
    assert.Equal(t, "Alice", user.Name)
    assert.Equal(t, "alice@example.com", user.Email)

    // 验证期望
    err = mockDB.ExpectationsWereMet()
    assert.NoError(t, err)

    // 清理
    mockDB.Mock.ExpectClose()
    mockDB.Close()
}
```

## 参考资料

- [go-sqlmock 文档](https://github.com/DATA-DOG/go-sqlmock)
- [GORM 文档](https://gorm.io/docs/)

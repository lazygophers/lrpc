# MongoDB Mock 使用指南

本文档介绍如何使用 LRPC MongoDB 中间件的 Mock 功能进行单元测试。

## 概述

MongoDB Mock 功能提供了一个轻量级的 Mock 客户端实现，无需真实的 MongoDB 服务器即可进行单元测试。

## Mock 实现方式

与 GORM 的 sqlmock 不同，MongoDB Mock 采用**手动 Mock** 方式：
- 通过 `MockClient` 手动设置方法的返回值
- 支持调用跟踪和断言
- 适合测试业务逻辑而非数据库操作本身

## 基本用法

### 方式一：使用 NewMock 函数（推荐）

直接创建 Mock 客户端：

```go
package mypackage

import (
    "testing"

    "github.com/lazygophers/lrpc/middleware/storage/mongo"
    "github.com/stretchr/testify/require"
)

func TestWithNewMock(t *testing.T) {
    // 创建 Mock 客户端
    cfg := &mongo.Config{
        Address:  "localhost",
        Port:     27017,
        Database: "test_db",
        Mock:     true,
    }

    client, err := mongo.NewMock(cfg)
    require.NoError(t, err)
    require.NotNil(t, client)

    // 使用客户端
    db := client.GetDatabase()
    require.Equal(t, "test_db", db)
}
```

### 方式二：使用 New 函数

通过配置启用 Mock 模式：

```go
func TestWithNew(t *testing.T) {
    cfg := &mongo.Config{
        Database: "test_db",
        Mock:     true, // 启用 Mock 模式
    }

    client, err := mongo.New(cfg)
    require.NoError(t, err)
    require.NotNil(t, client)
}
```

### 方式三：直接使用 MockClient（高级用法）

完全手动控制 Mock 行为：

```go
func TestWithMockClient(t *testing.T) {
    // 创建 MockClient
    mockClient := mongo.NewMockClient()

    // 设置返回值
    mockClient.
        SetupPingSuccess().
        SetupDatabase("test_db").
        SetupHealthSuccess()

    // 执行操作
    err := mockClient.Ping()
    require.NoError(t, err)

    db := mockClient.GetDatabase()
    require.Equal(t, "test_db", db)

    // 验证调用
    require.True(t, mockClient.AssertCalled("Ping"))
    require.Equal(t, 1, mockClient.GetCallCount("Ping"))
}
```

## MockClient 功能

### 设置返回值

```go
mockClient := mongo.NewMockClient()

// 设置成功响应
mockClient.SetupPingSuccess()
mockClient.SetupCloseSuccess()
mockClient.SetupHealthSuccess()
mockClient.SetupAutoMigrateSuccess()

// 设置错误响应
mockClient.SetupPingError(errors.New("connection failed"))
mockClient.SetupHealthError(errors.New("health check failed"))

// 设置配置和数据
mockClient.SetupDatabase("my_database")
mockClient.SetupConfig(&mongo.Config{Database: "test"})
mockClient.SetupContext(context.Background())
```

### 调用跟踪

```go
mockClient := mongo.NewMockClient()

// 执行操作
mockClient.Ping()
mockClient.GetDatabase()
mockClient.Ping() // 再次调用

// 验证调用
require.True(t, mockClient.AssertCalled("Ping"))
require.True(t, mockClient.AssertCalled("GetDatabase"))
require.False(t, mockClient.AssertCalled("Close"))

// 获取调用次数
require.Equal(t, 2, mockClient.GetCallCount("Ping"))
require.Equal(t, 1, mockClient.GetCallCount("GetDatabase"))

// 获取所有调用记录
calls := mockClient.GetCalls()
for _, call := range calls {
    fmt.Printf("Method: %s, Time: %v\n", call.Method, call.Time)
}

// 重置调用记录
mockClient.ResetCalls()
```

## 常见测试场景

### 1. 测试连接和健康检查

```go
func TestConnectionHealth(t *testing.T) {
    mockClient := mongo.NewMockClient()
    mockClient.SetupPingSuccess().SetupHealthSuccess()

    // 测试 Ping
    err := mockClient.Ping()
    require.NoError(t, err)

    // 测试 Health
    err = mockClient.Health()
    require.NoError(t, err)

    // 验证调用
    require.Equal(t, 1, mockClient.GetCallCount("Ping"))
    require.Equal(t, 1, mockClient.GetCallCount("Health"))
}
```

### 2. 测试错误处理

```go
func TestErrorHandling(t *testing.T) {
    mockClient := mongo.NewMockClient()

    // 设置错误
    expectedErr := errors.New("connection timeout")
    mockClient.SetupPingError(expectedErr)

    // 执行操作
    err := mockClient.Ping()

    // 验证错误
    require.Error(t, err)
    require.Equal(t, expectedErr, err)
}
```

### 3. 测试配置获取

```go
func TestGetConfig(t *testing.T) {
    cfg := &mongo.Config{
        Address:  "localhost",
        Port:     27017,
        Database: "test_db",
    }

    mockClient := mongo.NewMockClient()
    mockClient.SetupConfig(cfg)

    // 获取配置
    retrievedCfg := mockClient.GetConfig()
    require.Equal(t, cfg, retrievedCfg)
    require.Equal(t, "localhost", retrievedCfg.Address)
    require.Equal(t, 27017, retrievedCfg.Port)
}
```

### 4. 测试 AutoMigrate

```go
func TestAutoMigrate(t *testing.T) {
    mockClient := mongo.NewMockClient()
    mockClient.SetupAutoMigrateSuccess()

    // 定义模型
    type User struct {
        ID   string `bson:"_id"`
        Name string `bson:"name"`
    }

    // 执行迁移
    err := mockClient.AutoMigrate(&User{})
    require.NoError(t, err)

    // 验证调用
    require.True(t, mockClient.AssertCalled("AutoMigrate"))
}
```

### 5. 测试批量迁移

```go
func TestAutoMigrates(t *testing.T) {
    mockClient := mongo.NewMockClient()
    mockClient.SetupAutoMigratesSuccess()

    type User struct {
        ID string `bson:"_id"`
    }
    type Post struct {
        ID string `bson:"_id"`
    }

    // 批量迁移
    err := mockClient.AutoMigrates(&User{}, &Post{})
    require.NoError(t, err)

    // 验证调用
    require.True(t, mockClient.AssertCalled("AutoMigrates"))
}
```

## MockModel 使用

MockModel 用于测试实现了 Model 接口的类型：

```go
func TestMockModel(t *testing.T) {
    // 创建 MockModel
    mockModel := mongo.NewMockModel[User]()

    // 设置返回值
    mockModel.
        SetupCollectionName("users").
        SetupIsNotFound(false)

    // 使用 MockModel
    collName := mockModel.CollectionName()
    require.Equal(t, "users", collName)

    // 验证调用
    require.True(t, mockModel.AssertCalled("CollectionName"))
}
```

## 完整示例

```go
package repository

import (
    "testing"

    "github.com/lazygophers/lrpc/middleware/storage/mongo"
    "github.com/stretchr/testify/require"
)

type User struct {
    ID   string `bson:"_id"`
    Name string `bson:"name"`
}

type UserRepository struct {
    client *mongo.Client
}

func (r *UserRepository) Ping() error {
    return r.client.Ping()
}

func TestUserRepository_Ping(t *testing.T) {
    // 创建 Mock 客户端
    cfg := &mongo.Config{
        Database: "test_db",
        Mock:     true,
    }

    client, err := mongo.New(cfg)
    require.NoError(t, err)

    // 创建 repository
    repo := &UserRepository{client: client}

    // 测试 Ping（Mock 模式下默认成功）
    err = repo.Ping()
    require.NoError(t, err)
}

func TestUserRepository_WithMockClient(t *testing.T) {
    // 使用 MockClient 进行更精细的控制
    mockClient := mongo.NewMockClient()
    mockClient.SetupPingSuccess()

    // 注意：这里需要将 MockClient 转换为 Client
    // 实际使用中，建议通过接口抽象来解决这个问题

    // 执行测试
    err := mockClient.Ping()
    require.NoError(t, err)

    // 验证调用
    require.Equal(t, 1, mockClient.GetCallCount("Ping"))
}
```

## 最佳实践

### 1. 使用接口抽象

为了更好地支持 Mock，建议定义接口：

```go
type MongoClient interface {
    Ping() error
    Close() error
    GetDatabase() string
    Health() error
    // ... 其他方法
}

// 确保 Client 和 MockClient 都实现该接口
var _ MongoClient = (*mongo.Client)(nil)
var _ MongoClient = (*mongo.MockClient)(nil)
```

### 2. 测试业务逻辑而非数据库操作

Mock 适合测试业务逻辑，而不是数据库操作本身：

```go
// ✅ 好的用法：测试业务逻辑
func TestUserService_CreateUser(t *testing.T) {
    mockClient := mongo.NewMockClient()
    mockClient.SetupPingSuccess()

    service := NewUserService(mockClient)
    err := service.ValidateConnection()
    require.NoError(t, err)
}

// ❌ 不推荐：测试数据库操作
// 对于真实的 CRUD 操作，建议使用集成测试
```

### 3. 清理调用记录

在测试多个场景时，记得重置调用记录：

```go
func TestMultipleScenarios(t *testing.T) {
    mockClient := mongo.NewMockClient()

    // 场景 1
    mockClient.Ping()
    require.Equal(t, 1, mockClient.GetCallCount("Ping"))

    // 重置
    mockClient.ResetCalls()

    // 场景 2
    mockClient.Ping()
    require.Equal(t, 1, mockClient.GetCallCount("Ping")) // 重新计数
}
```

## 与 GORM Mock 的区别

| 特性 | MongoDB Mock | GORM Mock (sqlmock) |
|------|-------------|---------------------|
| 实现方式 | 手动 Mock | SQL 驱动 Mock |
| 设置期望 | 手动设置返回值 | 设置 SQL 期望 |
| 适用场景 | 业务逻辑测试 | SQL 查询测试 |
| 灵活性 | 高（完全控制） | 中（需匹配 SQL） |
| 学习曲线 | 低 | 中 |

## 注意事项

1. **Mock 模式限制**：Mock 客户端不会执行真实的数据库操作
2. **集成测试**：对于复杂的数据库操作，建议使用真实的 MongoDB 实例进行集成测试
3. **接口设计**：通过接口抽象可以更好地支持 Mock 和真实实现的切换
4. **调用验证**：充分利用调用跟踪功能来验证方法是否被正确调用

## 参考资料

- [MongoDB Go Driver 文档](https://www.mongodb.com/docs/drivers/go/current/)
- [testify 断言库](https://github.com/stretchr/testify)

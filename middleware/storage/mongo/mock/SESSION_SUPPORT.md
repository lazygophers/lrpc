# Session Support in MongoDB Mock

## Summary

Session 和事务功能在 MongoDB Mock 中**不完全支持**。

## 实现状态

### ❌ 不支持的功能

- `StartSession()` - 返回 `ErrNotImplemented`
- `UseSession()` - 返回 `ErrNotImplemented`
- `UseSessionWithOptions()` - 返回 `ErrNotImplemented`

### 技术原因

MongoDB Driver 的 `mongo.Session` 接口包含**未导出的方法**(unexported methods),这意味着我们无法在 `mongo-driver` 包外部实现这个接口。

来自 MongoDB 官方文档:
```go
type Session interface {
    // ... 公开方法 ...

    // Has unexported methods.
}
```

官方注释明确指出:
> "Custom implementations of this interface should not be used in production."

这是 MongoDB 团队的有意设计,目的是:
1. 确保 Session 实现符合内部要求
2. 防止不正确的 Session 实现导致数据不一致
3. 维护与 MongoDB 服务器的协议一致性

## 替代方案

### 方案 1: 使用真实 MongoDB (推荐)

在测试中使用 Docker 运行真实的 MongoDB 实例:

```bash
# 启动 MongoDB
docker run -d --name mongo-test -p 27017:27017 mongo:latest

# 运行测试
go test ./...

# 清理
docker stop mongo-test && docker rm mongo-test
```

测试代码示例:
```go
func TestWithRealMongo(t *testing.T) {
    // 连接到本地 MongoDB
    client, err := mongo.Connect(context.Background(),
        options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        t.Skip("MongoDB not available, skipping test")
    }
    defer client.Disconnect(context.Background())

    // 测试 Session 相关功能
    session, err := client.StartSession()
    require.NoError(t, err)
    defer session.EndSession(context.Background())

    err = session.StartTransaction()
    require.NoError(t, err)

    // ... 执行事务操作 ...

    err = session.CommitTransaction(context.Background())
    require.NoError(t, err)
}
```

### 方案 2: 使用 Mockgen

使用 `mockgen` 生成 Session 的 Mock:

```bash
# 安装 mockgen
go install github.com/golang/mock/mockgen@latest

# 生成 Session Mock
mockgen -destination=mocks/mock_session.go \
    go.mongodb.org/mongo-driver/mongo Session,SessionContext
```

测试代码示例:
```go
func TestWithMockgen(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockSession := mocks.NewMockSession(ctrl)
    mockSession.EXPECT().StartTransaction(gomock.Any()).Return(nil)
    mockSession.EXPECT().CommitTransaction(gomock.Any()).Return(nil)

    // 测试使用 mockSession
}
```

### 方案 3: 重构代码使用依赖注入

在更高层面进行 Mock,而不是 Mock Session 本身:

```go
// 定义业务接口
type UserRepository interface {
    CreateUser(ctx context.Context, user *User) error
    UpdateUser(ctx context.Context, user *User) error
}

// 真实实现使用 MongoDB Session
type mongoUserRepository struct {
    client mongo.Client
}

func (r *mongoUserRepository) CreateUser(ctx context.Context, user *User) error {
    session, err := r.client.StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(ctx)

    return session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
        // ... 事务操作 ...
        return nil, nil
    })
}

// 测试时 Mock 整个 Repository
type mockUserRepository struct {
    users map[string]*User
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *User) error {
    m.users[user.ID] = user
    return nil
}
```

### 方案 4: 测试集成层而非单元测试

如果 Session 逻辑非常关键,考虑编写集成测试而不是单元测试:

```go
//go:build integration
// +build integration

func TestTransactionIntegration(t *testing.T) {
    // 使用真实 MongoDB 进行集成测试
}
```

## 推荐实践

1. **单元测试**: 使用 Mock 测试不涉及 Session 的业务逻辑
2. **集成测试**: 使用真实 MongoDB + Docker 测试 Session/事务逻辑
3. **代码设计**: 将事务逻辑封装在独立的服务层,便于测试

## 示例项目结构

```
yourproject/
├── repository/
│   ├── user_repository.go           # 定义接口
│   ├── user_repository_mongo.go     # MongoDB 实现(包含事务)
│   └── user_repository_mock.go      # Mock 实现(测试用)
├── service/
│   └── user_service.go               # 业务逻辑
└── tests/
    ├── unit/
    │   └── user_service_test.go      # 单元测试(使用 Mock)
    └── integration/
        └── user_transaction_test.go  # 集成测试(真实 MongoDB)
```

## 参考链接

- [MongoDB Sessions Documentation](https://www.mongodb.com/docs/manual/reference/server-sessions/)
- [MongoDB Go Driver Documentation](https://www.mongodb.com/docs/drivers/go/current/)
- [gomock Documentation](https://github.com/golang/mock)

## 结论

虽然无法在 Mock 中完全实现 Session 接口,但通过合理的测试策略(单元测试 + 集成测试)和代码设计(依赖注入),我们仍然可以充分测试包含 Session/事务逻辑的代码。

对于大多数测试场景,Mock 提供的基础 CRUD 功能已经足够。对于需要测试事务隔离、回滚等复杂场景,推荐使用 Docker 运行真实 MongoDB 实例。

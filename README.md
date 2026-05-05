# LRPC - Lightweight RPC Framework

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> 一个基于 fasthttp 的轻量级、高性能 RPC 框架

## ✨ 特性

- 🚀 **高性能**: 基于 [fasthttp](https://github.com/valyala/fasthttp) 构建，性能优于标准库 net/http
- 🎯 **灵活路由**: 支持静态路由、参数路由和通配符路由
- 🔌 **中间件生态**: 内置丰富的中间件支持（认证、缓存、压缩、限流等）
- 🛡️ **类型安全**: 使用 Go 泛型提供类型安全的路由和中间件
- 🎨 **反射处理**: 自动处理请求解析和响应序列化
- 🔄 **连接池**: 内置连接池和对象池优化
- 📊 **可观测性**: 内置指标收集和健康检查
- 🔧 **可扩展**: 插件系统支持自定义功能扩展

## 📦 安装

```bash
go get github.com/lazygophers/lrpc
```

## 🚀 快速开始

### 基础示例

```go
package main

import (
    "github.com/lazygophers/lrpc"
    "github.com/lazygophers/log"
)

type HelloRequest struct {
    Name string `json:"name"`
}

type HelloResponse struct {
    Message string `json:"message"`
}

func main() {
    app := lrpc.New(lrpc.Config{
        Name: "hello-service",
        Port: 8080,
    })

    // 简单处理器
    app.Get("/hello", func(ctx *lrpc.Ctx) error {
        return ctx.SendString("Hello, World!")
    })

    // 带请求/响应类型的处理器
    app.Post("/hello", func(ctx *lrpc.Ctx, req *HelloRequest) (*HelloResponse, error) {
        return &HelloResponse{
            Message: "Hello, " + req.Name,
        }, nil
    })

    log.Fatal(app.Listen())
}
```

## 📚 路由系统

### 支持的路由类型

```go
// 静态路由
app.Get("/api/users", handler)

// 参数路由
app.Get("/api/users/:id", handler)

// 通配符路由
app.Get("/static/*", handler)
```

### 路由分组

```go
api := app.Group("/api")
{
    // GET /api/users
    api.Get("/users", listUsers)

    // POST /api/users
    api.Post("/users", createUser)

    v1 := api.Group("/v1")
    {
        // GET /api/v1/info
        v1.Get("/info", getInfo)
    }
}
```

### HTTP 方法

```go
app.Get("/path", handler)      // GET
app.Post("/path", handler)     // POST
app.Put("/path", handler)      // PUT
app.Delete("/path", handler)   // DELETE
app.Patch("/path", handler)    // PATCH
app.Head("/path", handler)     // HEAD
app.Options("/path", handler)  // OPTIONS
app.Any("/path", handler)      // All methods
```

## 🔧 中间件

### 内置中间件

#### 认证

```go
import "github.com/lazygophers/lrpc/middleware/auth"

// JWT 认证
app.Use(auth.JWT(auth.JWTConfig{
    SigningKey:    "your-secret-key",
    SigningMethod: "HS256",
}))

// Basic 认证
app.Use(auth.BasicAuth(auth.BasicAuthConfig{
    Users: map[string]string{
        "admin": "password",
    },
}))
```

#### 安全

```go
import "github.com/lazygophers/lrpc/middleware/security"

// CORS
app.Use(security.CORS(security.CORSConfig{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
}))

// 安全头
app.Use(security.SecurityHeaders(security.DefaultSecurityHeadersConfig))

// 限流
app.Use(security.RateLimit(security.RateLimitMiddlewareConfig{
    Rate:   100,
    Window: time.Minute,
}))
```

#### 压缩

```go
import "github.com/lazygophers/lrpc/middleware/compress"

app.Use(compress.Compress(compress.Config{
    Level:     compress.LevelDefault,
    MinLength: 1024,
}))
```

#### 缓存

```go
import "github.com/lazygophers/lrpc/middleware/cache"

app.Use(cache.Cache(cache.CacheConfig{
    MaxAge: 3600,
    Public: true,
}))
```

#### 指标收集

```go
import "github.com/lazygophers/lrpc/middleware/metrics"

collector := metrics.NewCollector()
app.Use(metrics.Metrics(metrics.Config{
    Collector: collector,
    SlowRequestConfig: metrics.SlowRequestConfig{
        Threshold: time.Second,
    },
}))

// 获取指标
stats := collector.GetMetrics()
```

#### 健康检查

```go
import "github.com/lazygophers/lrpc/middleware/health"

checker := health.NewChecker()
checker.AddCheck("database", health.DatabaseCheck(db.Ping))
checker.AddCheck("cache", health.CacheCheck(cache.Ping))

app.Get("/health", func(ctx *lrpc.Ctx) error {
    return ctx.SendJson(checker.RunChecks())
})
```

### 自定义中间件

```go
func Logger() lrpc.HandlerFunc {
    return func(ctx *lrpc.Ctx) error {
        start := time.Now()

        // 处理请求
        err := ctx.Next()

        // 记录日志
        log.Infof("method=%s path=%s duration=%v",
            ctx.Method(), ctx.Path(), time.Since(start))

        return err
    }
}

app.Use(Logger())
```

## 🗃️ 数据存储

### 数据库 (GORM)

```go
import "github.com/lazygophers/lrpc/middleware/storage/db"

// 初始化数据库
dbClient, err := db.NewClient(&db.Config{
    Driver: "mysql",
    DSN:    "user:pass@tcp(localhost:3306)/dbname",
})

// 注入到 context
app.Use(db.Middleware(dbClient))

// 在处理器中使用
func handler(ctx *lrpc.Ctx) error {
    db := db.GetDB(ctx)

    var users []User
    db.Find(&users)

    return ctx.SendJson(users)
}
```

### 缓存 (Redis/Memory)

```go
import "github.com/lazygophers/lrpc/middleware/storage/cache/redis"

// Redis 缓存
cache, err := redis.NewClient(&redis.Config{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
})

// 使用
err = cache.Set("key", value, time.Hour)
val, err := cache.Get("key")
```

### etcd

```go
import "github.com/lazygophers/lrpc/middleware/storage/etcd"

client, err := etcd.NewClient(etcd.Config{
    Endpoints: []string{"localhost:2379"},
})

// 服务发现
discovery := etcd.NewServiceDiscovery(client)
```

## 🔌 gRPC 集成

### HTTP 到 gRPC 桥接

```go
import "github.com/lazygophers/lrpc/middleware/grpc"

bridge := grpc.DefaultHTTPtoGRPCBridge
adapter := grpc.NewGRPCServiceAdapter(bridge)

// 将 gRPC handler 适配为 HTTP handler
app.Post("/api/grpc", adapter.UnaryHandler(grpcHandler, reqType))
```

### gRPC 客户端

```go
import "github.com/lazygophers/lrpc/middleware/grpc"

config := grpc.DefaultClientConfig
config.Address = "localhost:9090"

conn, err := grpc.NewClient(config)
```

## 🔄 连接池

### 自定义连接池

```go
import "github.com/lazygophers/lrpc/middleware/pool"

pool, err := pool.NewPool(
    pool.PoolConfig{
        MaxConns:    100,
        MinConns:    10,
        MaxIdleTime: 5 * time.Minute,
        MaxLifetime: 1 * time.Hour,
    },
    func() (interface{}, error) {
        // 创建连接
        return createConnection()
    },
    func(conn interface{}) error {
        // 关闭连接
        return conn.Close()
    },
)

// 使用连接
conn, err := pool.Acquire()
defer pool.Release(conn)
```

### 服务器连接池配置

```go
import "github.com/lazygophers/lrpc/middleware/pool"

// 高性能配置
config := pool.HighPerformanceConfig()

// 低内存配置
config := pool.LowMemoryConfig()

// 应用到服务器
pool.ApplyServerPoolConfig(server, config)
```

## 🔧 插件系统

### 创建插件

```go
import "github.com/lazygophers/lrpc/middleware/plugin"

type MyPlugin struct {
    *plugin.BasePlugin
}

func NewMyPlugin() *MyPlugin {
    return &MyPlugin{
        BasePlugin: plugin.NewBasePlugin("my-plugin", "1.0.0"),
    }
}

func (p *MyPlugin) Init(config interface{}) error {
    // 初始化逻辑
    return p.BasePlugin.Init(config)
}

func (p *MyPlugin) Start() error {
    // 启动逻辑
    return p.BasePlugin.Start()
}
```

### 使用插件管理器

```go
manager := plugin.NewManager()

// 注册插件
myPlugin := NewMyPlugin()
manager.Register(myPlugin)

// 初始化所有插件
manager.InitAll(configs)

// 启动所有插件
manager.StartAll()

// 停止所有插件
defer manager.StopAll()
```

## 📊 性能优化

### 对象池

框架内部使用 `sync.Pool` 优化内存分配：

- Context 对象池
- Buffer 池
- 连接池

### 性能建议

1. **使用连接池**: 重用数据库和 gRPC 连接
2. **启用压缩**: 对大响应启用 gzip 压缩
3. **配置限流**: 防止服务过载
4. **监控指标**: 使用 metrics 中间件监控性能
5. **调整 fasthttp 配置**: 根据负载调整并发和缓冲区大小

## 🧪 测试

框架包含完整的测试套件：

```bash
# 运行所有测试（单元测试）
go test ./...

# 运行特定包的测试
go test ./middleware/auth/...

# 查看覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 运行集成测试（需要启动 Docker 服务）
make test-setup
make test-integration
make test-teardown

# 运行基准测试
go test ./... -bench=. -benchmem
```

**测试说明**：
- 单元测试使用 mock 模式，不需要外部依赖
- 集成测试需要 Docker 服务（Redis, MySQL, ClickHouse），通过 `make test` 自动管理
- db 包当前使用 mock 测试，覆盖率约 74%

## 📖 示例

查看 `examples/` 目录获取更多示例：

- [基础 HTTP 服务](examples/basic/)
- [REST API](examples/rest-api/)
- [微服务](examples/microservice/)
- [gRPC 集成](examples/grpc/)

## 🛠️ 配置选项

### 应用配置

```go
app := lrpc.New(lrpc.Config{
    Name:              "my-service",
    Port:              8080,
    ReadTimeout:       30 * time.Second,
    WriteTimeout:      30 * time.Second,
    MaxRequestBodySize: 4 * 1024 * 1024, // 4MB
    EnableMetrics:     true,
    EnableHealth:      true,
})
```

### 中间件配置

每个中间件都有可配置的选项，参见各个中间件包的文档。

## 🤝 贡献

欢迎贡献！请遵循以下步骤：

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

### 开发指南

- 遵循 Go 代码规范
- 添加单元测试
- 更新文档
- 运行 `golangci-lint run` 检查代码质量

## 📝 License

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [fasthttp](https://github.com/valyala/fasthttp) - 高性能 HTTP 框架
- [GORM](https://gorm.io/) - ORM 库
- [go-redis](https://github.com/redis/go-redis) - Redis 客户端

## 📧 联系方式

- Issues: [GitHub Issues](https://github.com/lazygophers/lrpc/issues)
- Discussions: [GitHub Discussions](https://github.com/lazygophers/lrpc/discussions)

---

⭐ 如果这个项目对你有帮助，请给个 Star！

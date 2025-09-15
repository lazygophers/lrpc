# Cache 缓存中间件

高性能、多后端支持的 Go 缓存中间件，提供统一的缓存接口和多种存储实现。

## 🚀 特性

- **多种后端支持**: Memory、Redis、BboltDB、SugarDB、Bitcask、Database
- **统一接口**: 提供一致的缓存操作 API
- **类型安全**: 支持多种数据类型的直接存取
- **丰富功能**: 支持哈希、集合、限流等高级功能
- **高性能**: 优化的内存管理和连接池
- **测试完善**: 72.6%+ 测试覆盖率，全面的功能和错误路径测试

## 📦 支持的缓存类型

| 缓存类型 | 描述 | 适用场景 |
|---------|------|----------|
| **Memory** | 内存缓存 | 单机应用，快速访问 |
| **Redis** | 分布式缓存 | 分布式系统，数据共享 |
| **BboltDB** | 嵌入式键值数据库 | 持久化存储，单机应用 |
| **SugarDB** | 高性能键值存储 | 大数据量，高并发 |
| **Bitcask** | 日志型存储引擎 | 写密集型应用 |
| **Database** | SQL 数据库缓存 | 与现有数据库集成 |

## 🛠 安装

```bash
go get github.com/lazygophers/lrpc/middleware/storage/cache
```

## 📖 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/lazygophers/lrpc/middleware/storage/cache"
)

func main() {
    // 创建内存缓存
    c := cache.NewMem()
    defer c.Close()
    
    // 基本操作
    c.Set("key", "value")
    value, err := c.Get("key")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Value: %s\n", value)
    
    // 带过期时间
    c.SetEx("temp_key", "temp_value", time.Minute*5)
    
    // 检查是否存在
    exists, err := c.Exists("key")
    fmt.Printf("Key exists: %v\n", exists)
}
```

### 类型化操作

```go
// 存储和获取不同类型的数据
c.Set("number", 42)
c.Set("flag", true)
c.Set("price", 99.99)

// 类型安全的获取
intVal, err := c.GetInt("number")
boolVal, err := c.GetBool("flag")
floatVal, err := c.GetFloat64("price")
```

### 集合操作

```go
// 哈希操作
c.HSet("user:1", "name", "Alice")
c.HSet("user:1", "age", "25")
name, err := c.HGet("user:1", "name")
allFields, err := c.HGetAll("user:1")

// 集合操作
c.SAdd("tags", "go", "cache", "redis")
members, err := c.SMembers("tags")
isMember, err := c.SisMember("tags", "go")
```

### 配置和高级用法

```go
// 使用配置创建缓存
config := &cache.Config{
    Type:    cache.Redis,
    Address: "localhost:6379",
    Password: "password",
    Db:      0,
}

c, err := cache.New(config)
if err != nil {
    panic(err)
}
defer c.Close()

// 原子操作
count, err := c.Incr("counter")
c.IncrBy("score", 10)

// 限流功能
allowed, err := c.Limit("api:user:123", 100, time.Hour)
if !allowed {
    fmt.Println("Rate limit exceeded")
}
```

## 🔧 配置选项

### 内存缓存

```go
c := cache.NewMem()
```

### Redis 缓存

```go
c, err := cache.NewRedis("localhost:6379", 
    redis.DialDatabase(0),
    redis.DialPassword("password"),
    redis.DialConnectTimeout(time.Second*3),
)
```

### BboltDB 缓存

```go
c, err := cache.NewBbolt("/path/to/cache.db", &bbolt.Options{
    Timeout:      time.Second * 5,
    ReadOnly:     false,
    FreelistType: bbolt.FreelistArrayType,
})
```

### SugarDB 缓存

```go
config := &cache.Config{
    Type:    cache.SugarDB,
    DataDir: "/path/to/data",
}
c, err := cache.NewSugarDB(config)
```

## 📊 API 参考

### 基础操作

| 方法 | 说明 |
|------|------|
| `Set(key, value)` | 设置键值 |
| `Get(key)` | 获取值 |
| `SetEx(key, value, timeout)` | 设置带过期时间的键值 |
| `Del(keys...)` | 删除键 |
| `Exists(keys...)` | 检查键是否存在 |
| `Ttl(key)` | 获取键的剩余生存时间 |
| `Expire(key, timeout)` | 设置键的过期时间 |

### 类型化获取

| 方法 | 返回类型 |
|------|----------|
| `GetBool(key)` | `bool` |
| `GetInt(key)` | `int` |
| `GetUint(key)` | `uint` |
| `GetInt32(key)` | `int32` |
| `GetUint32(key)` | `uint32` |
| `GetInt64(key)` | `int64` |
| `GetUint64(key)` | `uint64` |
| `GetFloat32(key)` | `float32` |
| `GetFloat64(key)` | `float64` |

### 切片操作

| 方法 | 返回类型 |
|------|----------|
| `GetSlice(key)` | `[]string` |
| `GetBoolSlice(key)` | `[]bool` |
| `GetIntSlice(key)` | `[]int` |
| `GetFloat64Slice(key)` | `[]float64` |

### 哈希操作

| 方法 | 说明 |
|------|------|
| `HSet(key, field, value)` | 设置哈希字段 |
| `HGet(key, field)` | 获取哈希字段值 |
| `HGetAll(key)` | 获取所有哈希字段 |
| `HDel(key, fields...)` | 删除哈希字段 |
| `HExists(key, field)` | 检查哈希字段是否存在 |
| `HKeys(key)` | 获取所有哈希字段名 |

### 集合操作

| 方法 | 说明 |
|------|------|
| `SAdd(key, members...)` | 添加集合成员 |
| `SMembers(key)` | 获取所有集合成员 |
| `SRem(key, members...)` | 删除集合成员 |
| `SisMember(key, member)` | 检查是否为集合成员 |
| `SPop(key)` | 随机弹出一个成员 |
| `SRandMember(key, count)` | 随机获取成员 |

### 原子操作

| 方法 | 说明 |
|------|------|
| `Incr(key)` | 递增 1 |
| `Decr(key)` | 递减 1 |
| `IncrBy(key, value)` | 递增指定值 |
| `DecrBy(key, value)` | 递减指定值 |

### 高级功能

| 方法 | 说明 |
|------|------|
| `SetNx(key, value)` | 仅当键不存在时设置 |
| `SetNxWithTimeout(key, value, timeout)` | 带过期时间的 SetNx |
| `Limit(key, limit, timeout)` | 限流功能 |
| `GetJson(key, obj)` | JSON 反序列化 |
| `SetPb(key, msg)` | Protocol Buffers 序列化 |
| `GetPb(key, msg)` | Protocol Buffers 反序列化 |

## 🧪 测试

本项目拥有完善的测试覆盖率：

```bash
# 运行所有测试
go test ./middleware/storage/cache

# 运行测试并查看覆盖率
go test -cover ./middleware/storage/cache

# 生成详细覆盖率报告
go test -coverprofile=coverage.out ./middleware/storage/cache
go tool cover -html=coverage.out
```

### 测试文件说明

- `comprehensive_coverage_test.go` - SugarDB 和基础缓存全面测试
- `missing_coverage_test.go` - 数据库缓存和错误路径测试  
- `bbolt_additional_coverage_test.go` - BboltDB 额外覆盖率测试
- `echo_test.go` - SugarDB 实现测试
- `mem_test.go` - 内存缓存测试
- `bbolt_test.go` - BboltDB 缓存测试

## ⚡ 性能

不同缓存后端的性能特点：

| 后端 | 读性能 | 写性能 | 内存使用 | 持久化 | 分布式 |
|------|--------|--------|----------|---------|---------|
| Memory | 🟢 极高 | 🟢 极高 | 🔴 高 | ❌ 否 | ❌ 否 |
| Redis | 🟢 高 | 🟢 高 | 🟡 中 | ✅ 是 | ✅ 是 |
| BboltDB | 🟡 中 | 🟡 中 | 🟢 低 | ✅ 是 | ❌ 否 |
| SugarDB | 🟢 高 | 🟢 高 | 🟡 中 | ✅ 是 | ❌ 否 |
| Bitcask | 🟡 中 | 🟢 高 | 🟡 中 | ✅ 是 | ❌ 否 |

## 🔒 错误处理

所有缓存操作都遵循 Go 的错误处理约定：

```go
value, err := c.Get("key")
if err != nil {
    if err == cache.ErrNotFound {
        // 键不存在
        fmt.Println("Key not found")
    } else {
        // 其他错误
        fmt.Printf("Cache error: %v", err)
    }
}
```

### 常见错误

- `cache.ErrNotFound` - 键不存在
- 连接错误 - 网络或数据库连接问题
- 序列化错误 - 数据格式转换失败
- 权限错误 - 访问权限不足

## 🛡 最佳实践

### 1. 错误处理

```go
// 推荐的错误处理方式
value, err := c.Get(key)
if err != nil {
    log.Errorf("err:%v", err)
    return nil, err
}
```

### 2. 资源清理

```go
// 始终关闭缓存连接
defer c.Close()
```

### 3. 超时设置

```go
// 为重要数据设置合适的过期时间
c.SetEx("session:"+sessionID, sessionData, time.Hour*24)
```

### 4. 键命名规范

```go
// 使用有意义的键名和命名空间
c.Set("user:profile:"+userID, profileData)
c.Set("api:rate_limit:"+apiKey, rateLimitData)
```

### 5. 批量操作

```go
// 对于多个相关操作，使用事务或批量接口
c.SAdd("user_tags:"+userID, "golang", "backend", "cache")
```

## 🔧 开发和贡献

### 代码质量要求

- 所有新代码必须通过 `golangci-lint` 检查
- 测试覆盖率应保持在 70% 以上
- 遵循项目的错误处理约定
- 添加适当的文档注释

### 运行开发测试

```bash
# 安装依赖
go mod tidy

# 运行 lint 检查
make lint

# 运行所有测试
go test ./middleware/storage/cache

# 运行特定测试
go test -run TestMemory ./middleware/storage/cache
```

## 📝 更新日志

### v1.2.0 (Latest)
- ✨ 大幅提升测试覆盖率至 72.6%+
- 🧪 新增 3 个综合测试文件，874 行测试代码
- 🐛 修复 BboltDB 切片越界问题
- 🔧 优化错误处理和边界情况处理
- 📖 完善文档和使用示例

### v1.1.0
- ✨ 完整实现 SugarDB 缓存中间件
- 🔧 优化 Bitcask 缓存性能
- 📊 增强监控和日志功能

## 📄 许可证

本项目采用 MIT 许可证，详见 [LICENSE](../../LICENSE) 文件。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

如有问题或建议，欢迎提交 [Issue](https://github.com/lazygophers/lrpc/issues)。
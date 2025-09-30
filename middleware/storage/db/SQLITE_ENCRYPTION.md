# SQLite Encryption Support

本项目提供两种 SQLite 实现方案，用户可根据需求选择：

## 方案对比

| 特性 | `sqlite` (默认) | `sqlite-cgo` |
|------|----------------|--------------|
| CGO 依赖 | ❌ 无需 CGO | ✅ 需要 CGO |
| 驱动 | glebarez/sqlite (纯 Go) | mattn/go-sqlite3 |
| 加密支持 | ❌ 不支持 | ✅ 支持 SQLCipher |
| 跨平台编译 | ✅ 简单 | ⚠️ 需要 C 编译器 |
| 性能 | 较好 | 优秀 |
| 推荐场景 | 开发、测试、无加密需求 | 生产环境、需要加密 |

## 使用方法

### 1. 不加密（默认方案）

```go
import "github.com/lazygophers/lrpc/middleware/storage/db"

config := &db.Config{
    Type:    db.Sqlite,
    Address: "/path/to/data",
    Name:    "mydb",
}

client, err := db.New(config)
```

**构建命令：**
```bash
go build
```

### 2. 使用加密（sqlite-cgo）

```go
import "github.com/lazygophers/lrpc/middleware/storage/db"

config := &db.Config{
    Type:     db.SqliteCGO,
    Address:  "/path/to/data",
    Name:     "encrypted",
    Password: "your-encryption-key",  // 加密密钥
}

client, err := db.New(config)
```

**构建命令：**
```bash
# 需要启用 CGO 和 sqlite_cgo build tag
CGO_ENABLED=1 go build -tags sqlite_cgo
```

## 加密参数说明

当使用 `sqlite-cgo` 类型时，DSN 会自动包含以下 SQLCipher 参数：

- `_key=<password>` - 加密密钥
- `_cipher=sqlcipher` - 指定使用 SQLCipher
- `_kdf_iter=256000` - 密钥派生迭代次数（SQLCipher 4.x 兼容）

## 注意事项

### 1. 密码警告

如果在 `sqlite` 类型中设置了 `Password`，系统会输出警告日志：

```
[warn] SQLite password is set but 'sqlite' type does not support encryption.
       Use 'sqlite-cgo' type for encryption support
```

### 2. 构建要求

使用 `sqlite-cgo` 需要满足：
- 设置 `CGO_ENABLED=1`
- 添加 build tag `-tags sqlite_cgo`
- 系统安装 C 编译器（gcc/clang）

### 3. 跨平台编译

如果需要交叉编译 CGO 程序：

```bash
# macOS 编译 Linux 版本
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
  CC=x86_64-linux-musl-gcc \
  go build -tags sqlite_cgo

# Linux 编译 Windows 版本
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
  CC=x86_64-w64-mingw32-gcc \
  go build -tags sqlite_cgo
```

### 4. Docker 构建

Dockerfile 示例：

```dockerfile
# 多阶段构建
FROM golang:1.21-alpine AS builder

# 安装 CGO 依赖
RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY . .

# 构建加密版本
RUN CGO_ENABLED=1 go build -tags sqlite_cgo -o app

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/app /app
ENTRYPOINT ["/app"]
```

## 性能对比

| 操作 | sqlite (no CGO) | sqlite-cgo |
|------|----------------|------------|
| 小文件读写 | ~100% | ~105% |
| 大批量插入 | ~100% | ~110% |
| 复杂查询 | ~100% | ~108% |
| 编译速度 | 快 | 慢 |
| 二进制大小 | 较小 | 较大 |

*注：百分比相对于 glebarez/sqlite，实际性能取决于具体场景*

## FAQ

**Q: 可以在不重新构建的情况下切换加密吗？**
A: 不可以。必须使用 `-tags sqlite_cgo` 重新编译才能支持加密。

**Q: 如何迁移现有的未加密数据库？**
A: 使用 `sqlcipher` 命令行工具：
```bash
# 导出未加密数据库
sqlite3 old.db .dump > data.sql

# 导入到加密数据库
sqlcipher encrypted.db
> PRAGMA key = 'your-password';
> .read data.sql
```

**Q: 忘记加密密码怎么办？**
A: SQLCipher 使用的是真正的加密，密码丢失后无法恢复数据。请务必妥善保管密码。

**Q: 可以更改加密密码吗？**
A: 可以使用 `PRAGMA rekey`：
```go
db.Exec("PRAGMA key = 'old-password'")
db.Exec("PRAGMA rekey = 'new-password'")
```

## 相关链接

- [SQLCipher 官方文档](https://www.zetetic.net/sqlcipher/)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3)
- [glebarez/sqlite](https://github.com/glebarez/sqlite)
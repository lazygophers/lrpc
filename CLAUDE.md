# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

LRPC is a simple RPC framework written in Go that provides HTTP-based routing and handler functionality. The framework is built on top of fasthttp for performance and includes middleware support for various features like database access, caching, and service discovery.

## Development Commands

This is a Go project using Go modules. Common development tasks:

- **Build**: `go build`
- **Test**: `go test ./...` 
- **Run tests for specific packages**: `go test ./middleware/storage/db/`
- **Install dependencies**: `go mod tidy`
- **Format code**: `go fmt ./...`

## Core Architecture

### Main Components

1. **App (`app.go`)**: The main application struct that manages:
   - Server configuration and initialization using fasthttp
   - Route management with custom search tree routing
   - Context pooling for performance
   - Hook system for extensibility

2. **Context (`ctx.go`)**: Request context wrapper around fasthttp.RequestCtx:
   - Parameter extraction and parsing
   - Request/response handling
   - Local storage for middleware data
   - Body parsing with JSON and protobuf support

3. **Routing System (`tree.go`, `route.go`)**: 
   - Custom search tree implementation for path matching
   - Supports parameter extraction (`:param`) and wildcard routes
   - Generic tree structure `SearchTree[M any]` for type safety

4. **Handler System (`handler.go`)**:
   - Reflection-based handler conversion from regular functions to `HandlerFunc`
   - Automatic request parsing and response serialization
   - Support for various handler signatures:
     - `func(ctx *Ctx) error` - simple handlers
     - `func(ctx *Ctx, req *RequestStruct) error` - with request parsing
     - `func(ctx *Ctx) (*ResponseStruct, error)` - with response data
     - `func(ctx *Ctx, req *RequestStruct) (*ResponseStruct, error)` - full handlers

5. **Server (`server.go`)**: Server lifecycle management:
   - Listen configurations (local, LAN, specific IP)
   - Graceful shutdown handling
   - Hook integration for server events

### Middleware Architecture

Located in `middleware/` directory:

- **Storage (`middleware/storage/`)**: Database and caching layers
  - `db/`: Database middleware with GORM integration
  - `cache/`: Caching solutions (memory, Redis)
  - `etcd/`: etcd integration for distributed storage
- **Core (`middleware/core/`)**: Core middleware functionality
- **Service Discovery (`middleware/service_discovery/`)**: Service registration and discovery
- **I18n (`middleware/i18n/`)**: Internationalization support
- **Error Handling (`middleware/xerror/`)**: Error handling middleware

### Key Patterns

1. **Generic Types**: Extensive use of Go generics for type-safe routing and middleware
2. **Reflection-Based Handlers**: Automatic handler conversion based on function signatures
3. **Hook System**: Extensible event system for server lifecycle and routing events
4. **Context Pooling**: Performance optimization through object pooling
5. **FastHTTP Integration**: Built on fasthttp for high-performance HTTP handling

## Route Definition

Routes support:
- Static paths: `/api/users`
- Parameter paths: `/api/users/:id` 
- Wildcard paths: `/api/files/*`

Handler functions are automatically converted from various signatures to the internal `HandlerFunc` type.

## Configuration

Configuration is managed through the `Config` struct with options for:
- Server settings (name, ports, timeouts)
- Error handling callbacks
- Handler lifecycle hooks
- Middleware configuration

The framework uses a builder pattern with functional options for configuration.

## Error Handling Style

This codebase follows a consistent error handling pattern throughout all Go files. **Always use this style for error handling:**

### Preferred Style (Standard Pattern)

```go
err := someFunction()
if err != nil {
    log.Errorf("err:%v", err)
    return nil, err
}
```

**Key characteristics:**
1. **Separate assignment and check**: Assign error to `err` variable first, then check on next line
2. **Always log errors**: Use `log.Errorf("err:%v", err)` format for consistent error logging
3. **Explicit error propagation**: Return the error (don't ignore it)
4. **Clear structure**: Easy to read and debug

### Examples from codebase:

**Database connections:**
```go
cli, err := clientv3.New(clientv3.Config{...})
if err != nil {
    log.Errorf("err:%v", err)
    return nil, err
}
```

**JSON operations:**
```go
buffer, err := proto.Marshal(v)
if err != nil {
    log.Errorf("err:%v", err)
    return err
}
```

**File operations:**
```go
p.cli, err = bitcask.Open(c.DataDir, bitcask.WithAutoRecovery(true), bitcask.WithSyncWrites(false))
if err != nil {
    log.Errorf("err:%v", err)
    return nil, err
}
```

### Avoid This Style

**❌ DO NOT use inline error checking:**
```go
if err := json.Unmarshal(v, &item); err != nil {
    // Treat invalid JSON as non-existent
    return nil
}
```

**Why avoid inline style:**
- Less readable and harder to debug
- Inconsistent with codebase patterns
- Missing error logging
- Comments in conditional blocks are harder to maintain

### Special Cases

**Graceful error handling** (when appropriate):
Some operations like JSON parsing in cache layers may handle errors gracefully without propagating them, but should still follow the separate assignment pattern:

```go
err := json.Unmarshal(v, &item)
if err != nil {
    // Treat invalid JSON as non-existent (graceful handling)
    allExist = false
    return nil
}
```
- 针对middleware的测试，如果需要第三方那个服务（如 redis、etcd 等），可以通过 make test 创建、清理 docker 用以创建相关的临时服务
- 若需求是“middleware 添加语言包”，优先在 `middleware` 下创建独立目录（如 `middleware/language`）实现标准语言类型与解析处理，不放在 `middleware/i18n` 内耦合实现
- `middleware/language` 中的语言类型命名和常量命名需与 `golang.org/x/text/language` 风格对齐（如 `type Language string`、`English`、`SimplifiedChinese`、`TraditionalChinese`）
- `middleware/language` 的语言常量与标准集合需补全常用语言，并将“常量定义”和“标准/别名注册表”拆为两个独立文件维护
- 用户要求时需移除 `standardLanguageMap`、`IsStandardCode`、`IsStandard`、`StandardLanguages` 相关实现，不保留标准集合判断接口
- 用户要求时需移除公开 `Normalize` API（仅允许包内私有标准化函数）
- `ParseLangCode` 需使用缓存（复用解析结果，降低重复解析内存分配）
- 用户要求时不得使用统一 `language_registry.go` 注册表，需按场景分离映射逻辑（如 Parse 场景、Accept-Language 场景）
- 中文地区码（`zh-hk`、`zh-tw`、`zh-mo`、`zh-sg`、`zh-cht`、`zh-hant`）必须作为独立 code 处理，不能统一折叠为同一个 code
- 用户明确要求时仅提供 `ParseForHeader` 场景接口，不额外扩展其他场景 API
- 用户要求时需移除 `normalize` 函数，不保留该命名的标准化实现
- 用户要求时 `Parse`/`ParseLangCode` 不能依赖 `xlanguage.Parse`，需使用自实现语言标签解析规则

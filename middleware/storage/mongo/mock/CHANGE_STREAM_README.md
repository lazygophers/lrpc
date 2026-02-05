# Change Stream Implementation

## 概述

本实现为 MongoDB mock 提供了变更流（Change Stream）功能，用于监听数据变更事件。这是一个**基础实现**，旨在满足测试场景的需求。

## 功能特性

### 已实现功能

✅ **四种操作类型支持**
- Insert：文档插入
- Update：文档更新
- Replace：文档替换
- Delete：文档删除

✅ **三级 Watch 支持**
- Collection 级别：监听特定集合的变更
- Database 级别：监听所有集合的变更（由于没有数据库隔离，等同于 Client 级别）
- Client 级别：监听所有集合的变更

✅ **事件过滤**
- 集合名称过滤
- 操作类型过滤（内部支持，未暴露到公共 API）

✅ **事件详情**
- Insert/Replace：完整文档内容
- Update：更新字段和删除字段信息
- Delete：文档 _id

✅ **并发安全**
- 线程安全的事件发布
- 并发流管理

✅ **Context 支持**
- Context 取消
- Context 超时

## 架构设计

### 核心组件

#### 1. MockChangeStream

模拟 MongoDB 的 ChangeStream，提供事件接收和解码功能。

```go
type MockChangeStream struct {
    events     chan ChangeEvent  // 事件通道
    current    *ChangeEvent      // 当前事件
    closed     bool              // 关闭状态
    mu         sync.RWMutex      // 并发保护
    ctx        context.Context   // 生命周期控制
    cancelFunc context.CancelFunc
    filter     ChangeStreamFilter // 事件过滤
}
```

**关键方法**：
- `Next(ctx)`: 获取下一个事件
- `Decode(val)`: 解码事件到目标结构
- `Close(ctx)`: 关闭流

#### 2. ChangeEvent

表示一个变更事件。

```go
type ChangeEvent struct {
    OperationType     ChangeEventType    // 操作类型
    CollectionName    string             // 集合名称
    DocumentKey       bson.M             // 文档 _id
    FullDocument      bson.M             // 完整文档（insert/replace）
    UpdateDescription *UpdateDescription // 更新详情（update）
}
```

#### 3. MemoryStorage 扩展

在 MemoryStorage 中添加了事件发布机制。

**新增字段**：
```go
changeStreams []*MockChangeStream // 注册的流
streamsMu     sync.RWMutex        // 流管理锁
```

**新增方法**：
- `registerChangeStream()`: 注册流
- `unregisterChangeStream()`: 注销流
- `publishChangeEvent()`: 发布事件
- `ReplaceOne()`: 替换文档（发布 replace 事件）

### 事件流程

```
[数据操作]
    ↓
[MemoryStorage 操作方法]
    ↓
[publishChangeEvent()]
    ↓
[遍历所有注册的 MockChangeStream]
    ↓
[应用过滤器]
    ↓
[发送到 stream.events channel]
    ↓
[stream.Next() 接收]
    ↓
[stream.Decode() 解码]
    ↓
[用户代码处理]
```

## 使用示例

### 基础用法

```go
// 创建 client 和 collection
client := mock.NewMockClient()
mockClient := client.(*mock.MockClient)
db := mockClient.Database("testdb")
coll := db.Collection("users")
mockColl := coll.(*mock.MockCollection)

// 创建 change stream
ctx := context.Background()
stream := mockColl.WatchMock(ctx)
defer stream.Close(ctx)

// 在另一个 goroutine 中执行操作
go func() {
    time.Sleep(100 * time.Millisecond)
    doc := bson.M{
        "_id":  primitive.NewObjectID(),
        "name": "Alice",
        "age":  30,
    }
    mockColl.InsertOne(ctx, doc)
}()

// 监听事件
if stream.Next(ctx) {
    var event bson.M
    if err := stream.Decode(&event); err != nil {
        log.Fatalf("decode error: %v", err)
    }

    fmt.Printf("Operation: %s\n", event["operationType"])
    if fullDoc, ok := event["fullDocument"].(bson.M); ok {
        fmt.Printf("Document: %v\n", fullDoc)
    }
}
```

### 监听多个操作

```go
stream := mockColl.WatchMock(ctx)
defer stream.Close(ctx)

// 执行多个操作
go func() {
    // Insert
    mockColl.InsertOne(ctx, bson.M{"name": "Alice"})

    // Update
    mockColl.UpdateOne(ctx,
        bson.M{"name": "Alice"},
        bson.M{"$set": bson.M{"age": 30}})

    // Delete
    mockColl.DeleteOne(ctx, bson.M{"name": "Alice"})
}()

// 接收三个事件
for i := 0; i < 3; i++ {
    if stream.Next(ctx) {
        var event bson.M
        stream.Decode(&event)
        fmt.Printf("Event %d: %s\n", i+1, event["operationType"])
    }
}
```

## 实现细节

### 1. 事件发布时机

| 操作 | 发布时机 | 事件类型 | 包含数据 |
|------|---------|----------|---------|
| Insert | 插入后 | insert | FullDocument |
| InsertMany | 每个文档插入后 | insert | FullDocument |
| UpdateOne | 更新后 | update | UpdateDescription |
| UpdateMany | 每个文档更新后 | update | UpdateDescription |
| ReplaceOne | 替换后 | replace | FullDocument |
| DeleteOne | 删除后 | delete | DocumentKey |
| DeleteMany | 每个文档删除后 | delete | DocumentKey |

### 2. 并发处理

- **MemoryStorage 锁策略**：
  - 数据操作使用 `mu` 锁
  - 流管理使用 `streamsMu` 锁
  - 避免锁竞争

- **事件发布**：
  - 非阻塞发送（使用 select + default）
  - 缓冲满时丢弃事件并记录警告

### 3. 内存管理

- **事件缓冲**：默认 100 个事件
- **Document 拷贝**：发布前深拷贝，避免外部修改
- **流清理**：Close 时关闭 channel 和 context

### 4. 过滤机制

```go
type ChangeStreamFilter struct {
    CollectionName string            // 集合名过滤
    OperationTypes []ChangeEventType // 操作类型过滤
}
```

- Collection.WatchMock(): 自动设置 CollectionName 过滤器
- Database.WatchMock(): 无过滤器（监听所有）
- Client.WatchMock(): 无过滤器（监听所有）

## 限制和已知问题

### 1. mongo.ChangeStream 类型限制

**问题**：`mongo.ChangeStream` 是具体类型，不是接口，无法直接 mock。

**解决方案**：提供 `WatchMock()` 方法返回 `*MockChangeStream`。

**影响**：
- 标准 `Watch()` 方法返回 `ErrNotImplemented`
- 测试代码需要使用 `WatchMock()` 而非 `Watch()`

### 2. Pipeline 不支持

**当前状态**：pipeline 参数被忽略。

**原因**：基础实现，复杂度考虑。

**替代方案**：在客户端代码中手动过滤事件。

### 3. Resume Token 不支持

**当前状态**：`ResumeToken()` 返回空值。

**原因**：需要持久化和状态管理，超出基础实现范围。

**影响**：无法从中断点恢复流。

### 4. 数据库隔离

**当前状态**：MemoryStorage 没有数据库命名空间。

**影响**：
- Database.WatchMock() 和 Client.WatchMock() 行为相同
- 无法区分不同数据库的集合

### 5. 事件顺序保证

**保证**：同一集合的事件保持操作顺序。

**不保证**：跨集合的事件顺序（由于并发）。

### 6. 缓冲区溢出

**行为**：缓冲满时丢弃新事件。

**日志**：记录警告日志。

**建议**：及时处理事件或增大缓冲区。

## 测试覆盖

### 测试用例

1. ✅ `TestMockChangeStream_Insert` - 插入事件
2. ✅ `TestMockChangeStream_Update` - 更新事件
3. ✅ `TestMockChangeStream_Delete` - 删除事件
4. ✅ `TestMockChangeStream_Replace` - 替换事件
5. ✅ `TestMockChangeStream_MultipleCollections` - 集合过滤
6. ✅ `TestMockChangeStream_ClientWatch` - 客户端级监听
7. ✅ `TestMockChangeStream_ContextCancellation` - Context 取消
8. ✅ `TestMockChangeStream_CloseStream` - 流关闭

### 测试覆盖率

运行测试：
```bash
go test -v -run TestMockChangeStream
```

所有测试都应该通过。

## 性能考虑

### 1. 内存使用

- 每个流：~1KB（结构体） + 缓冲区大小 × 事件大小
- 默认缓冲区：100 × ~1KB = ~100KB/流
- 建议：不要创建过多同时活跃的流

### 2. CPU 使用

- 事件发布：O(N)，N = 注册的流数量
- 过滤：O(1)
- 建议：使用集合级流而非客户端级流

### 3. Goroutine

- 每个流：1 个 context goroutine
- 建议：记得关闭不用的流

## 未来增强

可能的改进方向：

1. **Pipeline 支持**
   - 基本的 $match 过滤
   - 字段投影

2. **Resume Token**
   - 简单的序列号机制
   - 内存中的事件历史

3. **数据库隔离**
   - 在 MemoryStorage 中支持数据库命名空间
   - Database.WatchMock() 真正只监听该数据库

4. **可配置缓冲区**
   - 允许用户指定缓冲区大小
   - 提供缓冲区满时的策略选择

5. **更丰富的操作类型**
   - invalidate
   - drop
   - rename
   - dropDatabase

6. **Pre/Post Image**
   - 支持 fullDocumentBeforeChange
   - 支持 fullDocument 选项

7. **性能优化**
   - 减少锁竞争
   - 事件池复用

## 总结

本实现提供了一个功能完整的基础变更流实现，适用于大多数测试场景。虽然有一些限制（主要是 mongo.ChangeStream 类型限制），但通过 WatchMock() 方法可以很好地满足测试需求。

关键优点：
- ✅ 实现简单清晰
- ✅ 并发安全
- ✅ 测试覆盖完整
- ✅ 易于使用

关键限制：
- ❌ 无法直接替换 mongo.Watch()
- ❌ 不支持 pipeline
- ❌ 不支持 resume token

对于需要更高级功能的场景，建议使用真实的 MongoDB 实例进行集成测试。

# MongoDB SDK 设计文档

## 核心设计理念

MongoDB SDK 提供两层 API：
1. **Scoop** - 低级查询构建器（灵活）
2. **Model** - 高级类型安全包装器（便捷）

## 1. Scoop 设计

### 概念

Scoop 是一个**查询构建器**，提供类似 Laravel Eloquent、SQLAlchemy 的链式 API。

```go
// Scoop 的本质：链式构建 MongoDB 查询（不依赖模型）
scoop := client.NewScoop("users")
    .Where("age", bson.M{"$gte": 25})
    .Equal("status", "active")
    .Sort("createdAt", -1)
    .Limit(10)
```

### 核心结构

```go
type Scoop struct {
    client *Client              // MongoDB 客户端
    coll   *mongo.Collection    // 集合引用
    filter bson.M              // 查询过滤条件
    opts   *options.FindOptions // MongoDB 选项（分页、排序等）
}
```

**关键变化：**
- **移除了 `model` 字段**：Scoop 现在完全独立于任何模型，是纯粹的泛型查询构建器
- **集合名作为显式参数**：通过 `Client.NewScoop(collectionName)` 方法

### Scoop 的设计特点

#### ✅ 优点

| 特点 | 说明 |
|-----|------|
| **灵活** | 支持所有 MongoDB 操作符（$eq、$gte、$in 等） |
| **通用** | 不需要类型参数，任何对象都能用 |
| **链式** | 所有方法返回 `*Scoop`，支持流畅 API |
| **可复用** | `Clone()` 支持快速复制现有查询状态 |
| **原生** | 直接使用 `bson.M` 操作符，无封装 |

#### ❌ 缺点

```go
// 需要手动处理 bson.M，容易出错
scoop.Where("age", bson.M{"$gte": 25})  // 类型不安全
scoop.Where("email", "user@example.com") // 类型不安全，可能发送错误数据

// 返回 bson.M，需要手动转换
var results []bson.M
scoop.Find(&results)  // 不是 []User

// 无法支持 OR 条件和复杂条件组合
// 每个 Where() 都是 AND 操作
```

### Scoop API 概览

#### 查询条件

```go
// 相等
scoop.Where("status", "active")
scoop.Equal("status", "active")

// 不相等
scoop.Ne("status", "deleted")

// 范围
scoop.In("type", "admin", "user", "guest")
scoop.NotIn("type", "banned")
scoop.Between("age", 18, 65)

// 比较
scoop.Gt("age", 25)       // >
scoop.Lt("age", 60)       // <
scoop.Gte("score", 100)   // >=
scoop.Lte("score", 100)   // <=

// 文本
scoop.Like("name", "john")  // 正则匹配
```

#### 结果处理

```go
// 查询
var results []bson.M
scoop.Find(&results)

var single bson.M
scoop.First(&single)

// 统计
count, err := scoop.Count()

// 检查
exist, err := scoop.Exist()

// CRUD
err := scoop.Create(doc)
modified, err := scoop.Update(updates)
deleted, err := scoop.Delete()
```

#### 分页和排序

```go
scoop.Limit(10)           // 限制条数
scoop.Offset(20)          // 跳过条数
scoop.Skip(20)            // 同 Offset

scoop.Sort("createdAt")       // 升序（默认）
scoop.Sort("createdAt", 1)    // 升序（显式指定）
scoop.Sort("createdAt", -1)   // 降序

scoop.Select("name", "email")  // 选择字段
```

#### 聚合

```go
agg := scoop.Aggregate(
    bson.M{"$match": bson.M{"age": bson.M{"$gte": 25}}},
    bson.M{"$group": bson.M{"_id": "$type", "count": bson.M{"$sum": 1}}},
)
var results []bson.M
agg.Execute(&results)
```

### Scoop 使用场景

```go
// 场景 1：动态条件查询
func GetUsers(filters map[string]interface{}) ([]bson.M, error) {
    scoop := client.NewScoop("users")
    for key, val := range filters {
        scoop.Where(key, val)
    }
    var results []bson.M
    scoop.Find(&results)
    return results, nil
}

// 场景 2：复杂聚合
func GetUserStats() ([]bson.M, error) {
    scoop := client.NewScoop("users")
    agg := scoop.Aggregate(
        bson.M{"$match": bson.M{"status": "active"}},
        bson.M{"$group": bson.M{
            "_id": "$department",
            "count": bson.M{"$sum": 1},
            "avgAge": bson.M{"$avg": "$age"},
        }},
    )
    var results []bson.M
    agg.Execute(&results)
    return results, nil
}

// 场景 3：批量操作
func UpdateUsers(ids []ObjectID, updates bson.M) (int64, error) {
    scoop := client.NewScoop("users")
    scoop.In("_id", ids...)
    return scoop.Update(updates)
}
```

---

## 2. Model 设计

### 概念

Model 是一个**轻量级的元数据包装器**，专注于模型元信息和集合名称的管理。类型安全的查询操作由 ModelScoop 提供。

```go
// Model 管理模型元数据和客户端引用
type User struct {
    ID    ObjectID `bson:"_id"`
    Name  string   `bson:"name"`
    Age   int      `bson:"age"`
    Email string   `bson:"email"`
}

model := NewModel(client, User{})

// 获取类型安全的查询构建器
modelScoop := model.NewScoop()
users, err := modelScoop.Find(bson.M{"age": bson.M{"$gte": 25}})
```

### 核心结构

```go
type Model[M any] struct {
    client *Client    // MongoDB 客户端
    model  M          // 泛型模型（用于元数据）
}

// 主要方法
func (m *Model[M]) NewScoop() *ModelScoop[M]  // 获取类型安全查询构建器
func (m *Model[M]) CollectionName() string         // 获取集合名

func (m *Model[M]) GetModel() M                    // 获取模型
```

### Model 的设计特点

#### ✅ 优点

| 特点 | 说明 |
|-----|------|
| **轻量级** | 仅管理元数据和客户端引用，逻辑清晰 |
| **灵活性** | 通过 NewScoop() 获取类型安全查询构建器 |
| **易用** | 集合名称自动解析，无需手动指定 |
| **关注分离** | Model 专注元数据，查询逻辑交由 ModelScoop |

#### ❌ 缺点

```go
// 需要定义结构体
type User struct {
    ID    ObjectID `bson:"_id"`
    Name  string   `bson:"name"`
    Age   int      `bson:"age"`
}

// 仅支持预定义的类型（不能动态查询）
model := NewModel(client, User{})
// 无法查询 Post、Comment 等其他类型（使用 Scoop 代替）
```

### Model API 概览

Model 主要用于管理元数据和获取 ModelScoop：

```go
// 获取类型安全查询构建器
modelScoop := model.NewScoop()

// 获取集合名
collectionName := model.CollectionName()

// 获取模型
user := model.GetModel()

// 设置客户端
model.SetClient(newClient)

// 获取反射类型信息
reflectType := model.GetReflectType()
reflectValue := model.GetReflectValue()
```

---

## 3. ModelScoop 设计

### 概念

ModelScoop 是 Scoop 的**类型安全包装器**，通过 Go 泛型提供类型检查的查询接口。

```go
// ModelScoop 提供类型安全的 CRUD 操作
modelScoop := model.NewScoop()

// 所有返回值都是类型安全的
users, err := modelScoop.Find(bson.M{"age": bson.M{"$gte": 25}})  // []User
user, err := modelScoop.First(bson.M{"email": "john@example.com"}) // *User
count, err := modelScoop.Count(bson.M{"status": "active"})        // int64
```

### 核心结构

```go
type ModelScoop[M any] struct {
    scoop *Scoop    // 底层 Scoop 查询构建器
    model M         // 模型（用于类型转换）
}

// CRUD 方法返回类型安全的结果
func (ms *ModelScoop[M]) Find(filter bson.M) ([]M, error)
func (ms *ModelScoop[M]) First(filter bson.M) (*M, error)
func (ms *ModelScoop[M]) Count(filter bson.M) (int64, error)
func (ms *ModelScoop[M]) Create(doc M) (interface{}, error)
func (ms *ModelScoop[M]) Update(filter bson.M, update interface{}) (int64, error)
func (ms *ModelScoop[M]) Delete(filter bson.M) (int64, error)
```

### ModelScoop API 概览

#### CRUD 操作

```go
// Create
err := modelScoop.Create(User{
    Name:  "John",
    Email: "john@example.com",
})

// Read - 单条 (First)
user, err := modelScoop.First(bson.M{"email": "john@example.com"})

// Read - 多条
users, err := modelScoop.Find(bson.M{"age": bson.M{"$gte": 25}})

// Update
count, err := modelScoop.Update(
    bson.M{"_id": id},
    bson.M{"$set": bson.M{"status": "active"}},
)

// Delete
count, err := modelScoop.Delete(bson.M{"_id": id})
```

#### 检查和计数

```go
// 检查是否存在
exist, err := modelScoop.Exist(bson.M{"email": email})

// 计数
count, err := modelScoop.Count(bson.M{"status": "active"})
```

#### 链式方法

```go
// 支持链式调用
modelScoop.
    Where("age", bson.M{"$gte": 25}).
    Limit(10).
    Offset(5).
    Sort("name", 1).
    Select("name", "email").
    Find(&users)
```

#### 高级功能

```go
// 聚合
agg := modelScoop.Aggregate(
    bson.M{"$match": bson.M{"status": "active"}},
    bson.M{"$group": bson.M{"_id": "$type"}},
)
var result []bson.M
agg.Execute(&result)

// 变更流
stream, err := modelScoop.Watch()

// 获取底层 Scoop（需要特殊操作时）
scoop := modelScoop.GetScoop()
collection := modelScoop.GetCollection()
```

### ModelScoop 使用场景

```go
// 场景 1：标准 CRUD
func CreateUser(ctx context.Context, user User) (interface{}, error) {
    model := NewModel(client, User{})
    modelScoop := model.NewScoop()
    return modelScoop.Create(user)
}

// 场景 2：类型安全查询
func GetActiveUsers(ctx context.Context) ([]User, error) {
    model := NewModel(client, User{})
    modelScoop := model.NewScoop()
    return modelScoop.Find(bson.M{"status": "active"})
}

// 场景 3：业务逻辑层
type UserService struct {
    model *Model[User]
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
    return s.model.NewScoop().First(bson.M{"email": email})
}

func (s *UserService) UpdateName(ctx context.Context, id ObjectID, name string) error {
    _, err := s.model.NewScoop().Update(
        bson.M{"_id": id},
        bson.M{"$set": bson.M{"name": name}},
    )
    return err
}
```

---

## 4. Scoop vs Model vs ModelScoop 对比

| 维度 | Scoop | Model | ModelScoop |
|-----|-------|-------|-----------|
| **类型安全** | ❌ 无 | ⭐ 元数据 | ✅ 完全 |
| **灵活性** | ✅ 极高 | ⭐ 中等 | ⭐ 中等 |
| **动态查询** | ✅ 支持 | ❌ 不支持 | ❌ 不支持 |
| **返回类型** | `bson.M` | - | 泛型 `M` |
| **代码简洁** | ❌ 冗长 | ⭐ 简单 | ✅ 简洁 |
| **性能** | ✅ 轻量 | ✅ 轻量 | ✅ 轻量 |
| **学习成本** | ⭐ 中等 | ✅ 低 | ✅ 低 |
| **错误易捕捉** | ❌ 运行时 | ⭐ 编译时 | ✅ 编译时 |
| **关键职责** | 查询构建 | 元数据管理 | 类型安全查询 |

### 场景选择矩阵

```
┌──────────────────────┬──────────┬────────┬──────────────┐
│ 使用场景              │ Scoop    │ Model  │ ModelScoop   │
├──────────────────────┼──────────┼────────┼──────────────┤
│ 标准 CRUD             │ ⭐⭐    │ ✅    │ ⭐⭐⭐⭐⭐  │
│ 动态条件查询           │ ⭐⭐⭐⭐⭐ │ ❌    │ ⭐⭐        │
│ 复杂聚合              │ ⭐⭐⭐⭐⭐ │ ⭐⭐ │ ⭐⭐⭐      │
│ 数据转换/ETL           │ ⭐⭐⭐⭐⭐ │ ❌    │ ❌          │
│ 业务逻辑层             │ ⭐⭐⭐  │ ✅   │ ⭐⭐⭐⭐⭐  │
│ API 响应生成           │ ⭐⭐    │ ✅   │ ⭐⭐⭐⭐⭐  │
│ 分析和报表             │ ⭐⭐⭐⭐⭐ │ ⭐⭐ │ ⭐⭐⭐      │
│ 原型开发(快速迭代)     │ ⭐⭐⭐⭐⭐ │ ✅   │ ⭐⭐⭐⭐⭐  │
└──────────────────────┴──────────┴────────┴──────────────┘
```

**使用指南：**
- **Scoop**: 需要完全灵活性和动态查询的场景
- **Model**: 用于管理模型元数据和获取 ModelScoop 实例
- **ModelScoop**: 大多数业务逻辑场景，优先选择

---

## 5. 参考对比：@middleware/storage/db

### DB 的设计

```go
// DB 中的 Scoop
type Scoop struct {
    orm    *gorm.DB
    filter bson.M
}

// 使用示例
db.NewScoop(&User{})
    .Where("age", bson.M{"$gte": 25})
    .Find(&results)

// DB 中的 Model
type Model[M any] struct {
    db *gorm.DB
}
```

### MongoDB SDK 的改进

| 改进点 | 说明 |
|-------|------|
| **内聚性** | Model 和 Scoop 紧密协作（Scoop 作为 Model 内部实现） |
| **聚合支持** | 原生支持 MongoDB 聚合管道 |
| **事务支持** | 专门的 Transaction 组件 |
| **变更流** | ChangeStream 用于实时监听 |
| **副本集** | Docker 配置支持事务和变更流 |

---

## 5. 实际应用示例

### 场景 1：用户管理服务

```go
// 数据模型
type User struct {
    ID        ObjectID  `bson:"_id"`
    Email     string    `bson:"email"`
    Name      string    `bson:"name"`
    Age       int       `bson:"age"`
    Status    string    `bson:"status"`
    CreatedAt time.Time `bson:"createdAt"`
}

func (u User) Collection() string {
    return "users"
}

// 服务层
type UserService struct {
    model *Model[User]
}

// 使用 ModelScoop 处理业务逻辑
func (s *UserService) CreateUser(ctx context.Context, user User) error {
    _, err := s.model.NewScoop().Create(user)
    return err
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
    return s.model.NewScoop().First(bson.M{"email": email})
}

func (s *UserService) ListActive(ctx context.Context) ([]User, error) {
    return s.model.NewScoop().Find(bson.M{"status": "active"})
}
```

### 场景 2：数据分析和报表

```go
// 使用 Scoop 进行复杂聚合
func GetUserStats(ctx context.Context) ([]map[string]interface{}, error) {
    scoop := client.NewScoop("users")

    agg := scoop.Aggregate(
        // 过滤活跃用户
        bson.M{"$match": bson.M{"status": "active"}},

        // 按年龄段分组
        bson.M{"$group": bson.M{
            "_id": bson.M{
                "$cond": bson.A{
                    bson.M{"$gte": bson.A{"$age", 60}},
                    "senior",
                    bson.M{
                        "$cond": bson.A{
                            bson.M{"$gte": bson.A{"$age", 30}},
                            "adult",
                            "young",
                        },
                    },
                },
            },
            "count": bson.M{"$sum": 1},
            "avgAge": bson.M{"$avg": "$age"},
        }},

        // 排序
        bson.M{"$sort": bson.M{"count": -1}},
    )

    var results []map[string]interface{}
    agg.Execute(&results)
    return results, nil
}
```

### 场景 3：事务处理

```go
// 转账操作（需要事务）
func Transfer(fromID, toID ObjectID, amount int) error {
    txn := client.NewTransaction()

    return txn.WithTx(func(txCtx mongo.SessionContext) error {
        fromModel := NewModel(client, User{})
        toModel := NewModel(client, User{})

        // 扣款
        _, err := fromModel.NewScoop().Update(
            bson.M{"_id": fromID},
            bson.M{"$inc": bson.M{"balance": -amount}},
        )
        if err != nil {
            return err
        }

        // 加款
        _, err = toModel.NewScoop().Update(
            bson.M{"_id": toID},
            bson.M{"$inc": bson.M{"balance": amount}},
        )
        return err
    })
}
```

---

## 总结

- **Scoop**: 查询构建器，通过 `Client.NewScoop(collectionName)` 创建，专注于**灵活性**和**动态性**
- **Model**: 元数据管理器，通过 `Model.NewScoop()` 获取 ModelScoop，专注于**模型元信息**
- **ModelScoop**: 类型安全查询构建器，提供类型检查的 CRUD 操作
- **组合使用**: 大多数业务场景使用 `Model -> NewScoop() -> ModelScoop` 的组合，需要特殊操作时可直接使用 Scoop

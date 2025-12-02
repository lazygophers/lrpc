# SQLite 使用指南

## 自增主键的正确用法

SQLite 的自增主键有特殊要求，与 MySQL 等数据库不同。

### 问题说明

SQLite 的 `AUTOINCREMENT` 只能用于 `INTEGER PRIMARY KEY` 类型，不能用于 `BIGINT` 或其他整数类型。

### 错误示例

```go
type Model struct {
    // ❌ 错误：显式指定 type:bigint unsigned 会导致 SQLite 无法使用 AUTOINCREMENT
    Id uint64 `gorm:"type:bigint unsigned;not null;primaryKey;autoIncrement;column:id"`
}
```

使用上述定义时，SQLite 会创建如下表结构：
```sql
CREATE TABLE `model` (`id` bigint unsigned NOT NULL, ... PRIMARY KEY (`id`));
```

这会导致插入数据时失败，错误信息：
```
constraint failed: NOT NULL constraint failed: model.id
```

### 正确示例

```go
type Model struct {
    // ✅ 正确：不指定具体类型，让 GORM 根据数据库自动选择
    Id uint64 `gorm:"primaryKey;autoIncrement;column:id;comment:唯一标识"`
}
```

使用上述定义时：
- **MySQL/TiDB**: 会生成 `BIGINT UNSIGNED AUTO_INCREMENT`
- **SQLite**: 会生成 `INTEGER PRIMARY KEY AUTOINCREMENT`

### 推荐实践

1. **不要在自增主键字段上使用 `type` 标签**
   - 让 GORM 根据数据库类型自动选择合适的数据类型

2. **如果必须跨数据库兼容，使用条件编译**
   ```go
   type Model struct {
       // 使用 GORM 的默认类型推断
       Id uint64 `gorm:"primaryKey;autoIncrement"`
   }
   ```

3. **测试时验证表结构**
   ```bash
   # 查看 SQLite 表结构
   sqlite3 your_database.db ".schema table_name"

   # 正确的自增主键应该显示为：
   # `id` INTEGER PRIMARY KEY AUTOINCREMENT
   ```

### 相关链接

- [SQLite AUTOINCREMENT 文档](https://www.sqlite.org/autoinc.html)
- [GORM SQLite 驱动](https://github.com/glebarez/sqlite)
- [相关 Issue](https://github.com/lazygophers/lrpc/issues/xxx)

### 故障排查

如果遇到 `NOT NULL constraint failed` 错误：

1. 检查模型定义，确认自增主键字段没有使用 `type:bigint` 标签
2. 删除数据库文件，让 GORM 重新创建表结构
3. 使用 `sqlite3 your_database.db ".schema table_name"` 验证表结构

### 技术细节

SQLite 的 `INTEGER PRIMARY KEY` 有特殊含义：
- 它是 `ROWID` 的别名
- 自动递增，即使不指定 `AUTOINCREMENT`
- `AUTOINCREMENT` 关键字确保 ID 永不重复（即使删除记录）

因此，对于 SQLite：
- `INTEGER PRIMARY KEY` → 自动递增
- `INTEGER PRIMARY KEY AUTOINCREMENT` → 自动递增且永不重复
- `BIGINT PRIMARY KEY` → **不会**自动递增

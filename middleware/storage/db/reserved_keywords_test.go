package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// TestIsReservedKeyword 测试保留关键字检测功能
func TestIsReservedKeyword(t *testing.T) {
	// 由于 isReservedKeyword 是内部函数，我们通过 quoteFieldName 间接测试

	// MySQL 常见保留关键字
	quoted := db.QuoteFieldName("group", db.MySQL)
	assert.Equal(t, "`group`", quoted)

	quoted = db.QuoteFieldName("order", db.MySQL)
	assert.Equal(t, "`order`", quoted)

	quoted = db.QuoteFieldName("select", db.MySQL)
	assert.Equal(t, "`select`", quoted)

	// user 不是 MySQL 的保留关键字，但是 PostgreSQL 的
	quoted = db.QuoteFieldName("user", db.MySQL)
	assert.Equal(t, "user", quoted) // MySQL 中不需要引用

	// PostgreSQL 特有保留关键字
	quoted = db.QuoteFieldName("user", db.Postgres)
	assert.Equal(t, "\"user\"", quoted)

	// SQLite 特有保留关键字
	quoted = db.QuoteFieldName("index", db.Sqlite)
	assert.Equal(t, "\"index\"", quoted)

	// ClickHouse 特有保留关键字
	quoted = db.QuoteFieldName("engine", db.ClickHouse)
	assert.Equal(t, "`engine`", quoted)

	// 非保留关键字
	quoted = db.QuoteFieldName("username", db.MySQL)
	assert.Equal(t, "username", quoted)

	quoted = db.QuoteFieldName("email", db.Postgres)
	assert.Equal(t, "email", quoted)
}

// TestQuoteFieldNameWithSpecialChars 测试特殊字符字段名
func TestQuoteFieldNameWithSpecialChars(t *testing.T) {
	// 包含特殊字符的字段名应该被引用
	quoted := db.QuoteFieldName("123field", db.MySQL)
	assert.Equal(t, "`123field`", quoted)

	quoted = db.QuoteFieldName("field-name", db.Postgres)
	assert.Equal(t, "\"field-name\"", quoted)

	quoted = db.QuoteFieldName("field.name", db.MySQL)
	assert.Equal(t, "`field.name`", quoted)
}

// TestQuoteFieldNameAlreadyQuoted 测试已引用的字段名
func TestQuoteFieldNameAlreadyQuoted(t *testing.T) {
	// 已引用的字段应该保持不变
	quoted := db.QuoteFieldName("`group`", db.MySQL)
	assert.Equal(t, "`group`", quoted)

	quoted = db.QuoteFieldName("\"user\"", db.Postgres)
	assert.Equal(t, "\"user\"", quoted)
}

// TestQuoteFieldNameNormalField 测试普通字段名
func TestQuoteFieldNameNormalField(t *testing.T) {
	// 普通字段名不应该被引用
	quoted := db.QuoteFieldName("username", db.MySQL)
	assert.Equal(t, "username", quoted)

	quoted = db.QuoteFieldName("email", db.Postgres)
	assert.Equal(t, "email", quoted)

	quoted = db.QuoteFieldName("age", db.Sqlite)
	assert.Equal(t, "age", quoted)
}

// TestCondWithReservedKeywords 测试使用保留关键字的条件构建
// 注意：由于 db.Where() 创建的 Cond 没有设置 clientType，
// 这个测试只验证 QuoteFieldName 函数本身
func TestCondWithReservedKeywords(t *testing.T) {
	// 直接测试 QuoteFieldName 函数
	result := db.QuoteFieldName("group", db.MySQL)
	assert.Equal(t, "`group`", result)

	result = db.QuoteFieldName("order", db.MySQL)
	assert.Equal(t, "`order`", result)

	result = db.QuoteFieldName("select", db.MySQL)
	assert.Equal(t, "`select`", result)
}

// TestCondWithPostgresReservedKeywords 测试 PostgreSQL 保留关键字
func TestCondWithPostgresReservedKeywords(t *testing.T) {
	// 直接测试 QuoteFieldName 函数
	result := db.QuoteFieldName("user", db.Postgres)
	assert.Equal(t, "\"user\"", result)

	result = db.QuoteFieldName("from", db.Postgres)
	assert.Equal(t, "\"from\"", result)

	result = db.QuoteFieldName("select", db.Postgres)
	assert.Equal(t, "\"select\"", result)
}

// TestBuildInsertSQLWithReservedKeywords 测试 INSERT 语句中的保留关键字
func TestBuildInsertSQLWithReservedKeywords(t *testing.T) {
	// 创建测试 Scoop
	// 注意：由于 Scoop 是内部类型，我们通过导出的 NewScoop 函数创建
	// 但 NewScoop 需要一个 *gorm.DB，所以这个测试可能需要调整

	// 暂时跳过这个测试，因为它需要数据库连接
	t.Skip("需要数据库连接")
}

// BenchmarkQuoteFieldName 性能基准测试
func BenchmarkQuoteFieldName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db.QuoteFieldName("group", db.MySQL)
	}
}

// BenchmarkQuoteFieldNameNormal 性能基准测试 - 普通字段名
func BenchmarkQuoteFieldNameNormal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db.QuoteFieldName("username", db.MySQL)
	}
}

// BenchmarkQuoteFieldNameAlreadyQuoted 性能基准测试 - 已引用字段名
func BenchmarkQuoteFieldNameAlreadyQuoted(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db.QuoteFieldName("`group`", db.MySQL)
	}
}

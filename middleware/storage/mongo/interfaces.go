package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoCollectionOps 定义 MongoDB 集合操作的接口
type MongoCollectionOps interface {
	// Ping 检查连接
	Ping(ctx context.Context) error

	// CountDocuments 计数文档
	CountDocuments(ctx context.Context, filter interface{}) (int64, error)

	// FindOne 查询单个文档
	FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult

	// Find 查询多个文档
	Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error)

	// InsertOne 插入单个文档
	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)

	// InsertMany 插入多个文档
	InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error)

	// UpdateMany 更新多个文档
	UpdateMany(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error)

	// DeleteMany 删除多个文档
	DeleteMany(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error)

	// Aggregate 聚合操作
	Aggregate(ctx context.Context, pipeline interface{}) (*mongo.Cursor, error)
}

// SessionOps 定义 MongoDB Session 操作的接口
type SessionOps interface {
	// StartTransaction 开始事务
	StartTransaction() error

	// CommitTransaction 提交事务
	CommitTransaction(ctx context.Context) error

	// AbortTransaction 中止事务
	AbortTransaction(ctx context.Context) error

	// EndSession 结束会话
	EndSession(ctx context.Context)
}

// ClientOps 定义 MongoDB Client 操作的接口
type ClientOps interface {
	// Ping 检查连接
	Ping(ctx context.Context) error

	// StartSession 开始会话
	StartSession() (SessionOps, error)

	// Watch 监视数据库变化
	Watch(ctx context.Context, pipeline interface{}) (*mongo.ChangeStream, error)
}

// FailureInjector 故障注入接口，用于测试
type FailureInjector interface {
	// ShouldFailPing 检查是否应该模拟 Ping 失败
	ShouldFailPing() bool

	// GetPingError 获取模拟的 Ping 错误
	GetPingError() error

	// ShouldFailFind 检查是否应该模拟 Find 失败
	ShouldFailFind() bool

	// GetFindError 获取模拟的 Find 错误
	GetFindError() error

	// ShouldFailCount 检查是否应该模拟 Count 失败
	ShouldFailCount() bool

	// GetCountError 获取模拟的 Count 错误
	GetCountError() error

	// ShouldFailDelete 检查是否应该模拟 Delete 失败
	ShouldFailDelete() bool

	// GetDeleteError 获取模拟的 Delete 错误
	GetDeleteError() error

	// ShouldFailTransaction 检查是否应该模拟事务失败
	ShouldFailTransaction() bool

	// GetTransactionError 获取模拟的事务错误
	GetTransactionError() error

	// ShouldFailWatch 检查是否应该模拟 Watch 失败
	ShouldFailWatch() bool

	// GetWatchError 获取模拟的 Watch 错误
	GetWatchError() error

	// ShouldFailClose 检查是否应该模拟 Close 失败
	ShouldFailClose() bool

	// GetCloseError 获取模拟的 Close 错误
	GetCloseError() error
}

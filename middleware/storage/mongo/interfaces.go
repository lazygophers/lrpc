package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// MongoClient 包装 *mongo.Client，提供 MongoDB 客户端操作接口
type MongoClient interface {
	// Connect 初始化客户端连接池并启动后台监控 goroutine
	Connect(ctx context.Context) error

	// Database 返回指定名称的数据库实例
	Database(name string, opts ...*options.DatabaseOptions) MongoDatabase

	// Disconnect 关闭客户端的所有连接
	Disconnect(ctx context.Context) error

	// ListDatabaseNames 返回所有数据库的名称列表
	ListDatabaseNames(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) ([]string, error)

	// ListDatabases 返回所有数据库的详细信息
	ListDatabases(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) (mongo.ListDatabasesResult, error)

	// NumberSessionsInProgress 返回当前正在进行的 session 数量
	NumberSessionsInProgress() int

	// Ping 验证到 MongoDB 部署的连接
	Ping(ctx context.Context, rp *readpref.ReadPref) error

	// StartSession 启动一个新的 session
	StartSession(opts ...*options.SessionOptions) (mongo.Session, error)

	// Timeout 返回客户端的超时时间
	Timeout() *time.Duration

	// UseSession 在提供的 session 上下文中执行函数
	UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error

	// UseSessionWithOptions 在提供的 session 选项和上下文中执行函数
	UseSessionWithOptions(ctx context.Context, opts *options.SessionOptions, fn func(mongo.SessionContext) error) error

	// Watch 创建一个变更流来监听客户端级别的变更
	Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)
}

// MongoDatabase 包装 *mongo.Database，提供 MongoDB 数据库操作接口
type MongoDatabase interface {
	// Aggregate 在数据库级别执行聚合管道
	Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (MongoCursor, error)

	// Client 返回创建此数据库的客户端
	Client() MongoClient

	// Collection 返回指定名称的集合实例
	Collection(name string, opts ...*options.CollectionOptions) MongoCollection

	// CreateCollection 创建一个新集合
	CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error

	// CreateView 创建一个视图
	CreateView(ctx context.Context, viewName, viewOn string, pipeline interface{}, opts ...*options.CreateViewOptions) error

	// Drop 删除此数据库
	Drop(ctx context.Context) error

	// ListCollectionNames 返回此数据库中所有集合的名称
	ListCollectionNames(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) ([]string, error)

	// ListCollectionSpecifications 返回此数据库中所有集合的详细规范
	ListCollectionSpecifications(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) ([]*mongo.CollectionSpecification, error)

	// ListCollections 返回此数据库中所有集合的游标
	ListCollections(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) (MongoCursor, error)

	// Name 返回此数据库的名称
	Name() string

	// ReadConcern 返回此数据库的读关注
	ReadConcern() *readconcern.ReadConcern

	// ReadPreference 返回此数据库的读偏好
	ReadPreference() *readpref.ReadPref

	// RunCommand 在此数据库上执行命令
	RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) *mongo.SingleResult

	// RunCommandCursor 在此数据库上执行命令并返回游标
	RunCommandCursor(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) (MongoCursor, error)

	// Watch 创建一个变更流来监听数据库级别的变更
	Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)

	// WriteConcern 返回此数据库的写关注
	WriteConcern() *writeconcern.WriteConcern
}

// MongoCollection 包装 *mongo.Collection，提供 MongoDB 集合操作接口
type MongoCollection interface {
	// Aggregate 对集合执行聚合管道
	Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (MongoCursor, error)

	// BulkWrite 执行批量写操作
	BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error)

	// Clone 创建此集合的副本
	Clone(opts ...*options.CollectionOptions) (MongoCollection, error)

	// CountDocuments 返回与过滤器匹配的文档数量
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)

	// Database 返回此集合所属的数据库
	Database() MongoDatabase

	// DeleteMany 删除与过滤器匹配的所有文档
	DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)

	// DeleteOne 删除与过滤器匹配的单个文档
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)

	// Distinct 返回与过滤器匹配的文档中指定字段的所有不同值
	Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error)

	// Drop 删除此集合
	Drop(ctx context.Context) error

	// EstimatedDocumentCount 返回集合中文档的估计数量
	EstimatedDocumentCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (int64, error)

	// Find 查找与过滤器匹配的所有文档
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (MongoCursor, error)

	// FindOne 查找与过滤器匹配的单个文档
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult

	// FindOneAndDelete 查找并删除与过滤器匹配的单个文档
	FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult

	// FindOneAndReplace 查找并替换与过滤器匹配的单个文档
	FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult

	// FindOneAndUpdate 查找并更新与过滤器匹配的单个文档
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult

	// Indexes 返回集合的索引视图
	Indexes() mongo.IndexView

	// InsertMany 插入多个文档
	InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error)

	// InsertOne 插入单个文档
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)

	// Name 返回集合的名称
	Name() string

	// ReplaceOne 替换与过滤器匹配的单个文档
	ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error)

	// SearchIndexes 返回集合的搜索索引视图
	SearchIndexes() mongo.SearchIndexView

	// UpdateByID 通过 ID 更新单个文档
	UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// UpdateMany 更新与过滤器匹配的所有文档
	UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// UpdateOne 更新与过滤器匹配的单个文档
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)

	// Watch 创建一个变更流来监听集合级别的变更
	Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)
}

// MongoCursor 包装 *mongo.Cursor，提供 MongoDB 游标操作接口
type MongoCursor interface {
	// All 将游标中的所有文档解码到 results 中
	All(ctx context.Context, results interface{}) error

	// Close 关闭游标
	Close(ctx context.Context) error

	// Decode 将当前文档解码到 val 中
	Decode(val interface{}) error

	// Err 返回游标的错误
	Err() error

	// ID 返回游标的 ID
	ID() int64

	// Next 将游标移动到下一个文档
	Next(ctx context.Context) bool

	// RemainingBatchLength 返回当前批次中剩余的文档数量
	RemainingBatchLength() int

	// SetBatchSize 设置游标的批次大小
	SetBatchSize(batchSize int32)

	// SetComment 设置游标的注释
	SetComment(comment interface{})

	// SetMaxTime 设置游标的最大执行时间
	SetMaxTime(dur time.Duration)

	// TryNext 尝试将游标移动到下一个文档
	TryNext(ctx context.Context) bool

	// Current 返回当前文档的 BSON 数据
	Current() bson.Raw
}

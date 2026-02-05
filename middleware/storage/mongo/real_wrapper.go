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

// RealClient 包装 *mongo.Client，实现 MongoClient 接口
// 直接委托所有方法给官方 MongoDB 驱动
type RealClient struct {
	*mongo.Client
}

// NewRealClient 创建一个新的 RealClient 实例
func NewRealClient(client *mongo.Client) *RealClient {
	return &RealClient{Client: client}
}

// Connect 初始化客户端连接池并启动后台监控 goroutine
func (c *RealClient) Connect(ctx context.Context) error {
	return c.Client.Connect(ctx)
}

// Database 返回指定名称的数据库实例
func (c *RealClient) Database(name string, opts ...*options.DatabaseOptions) MongoDatabase {
	db := c.Client.Database(name, opts...)
	return &RealDatabase{
		Database: db,
		client:   c,
	}
}

// Disconnect 关闭客户端的所有连接
func (c *RealClient) Disconnect(ctx context.Context) error {
	return c.Client.Disconnect(ctx)
}

// ListDatabaseNames 返回所有数据库的名称列表
func (c *RealClient) ListDatabaseNames(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) ([]string, error) {
	return c.Client.ListDatabaseNames(ctx, filter, opts...)
}

// ListDatabases 返回所有数据库的详细信息
func (c *RealClient) ListDatabases(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) (mongo.ListDatabasesResult, error) {
	return c.Client.ListDatabases(ctx, filter, opts...)
}

// NumberSessionsInProgress 返回当前正在进行的 session 数量
func (c *RealClient) NumberSessionsInProgress() int {
	return c.Client.NumberSessionsInProgress()
}

// Ping 验证到 MongoDB 部署的连接
func (c *RealClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	return c.Client.Ping(ctx, rp)
}

// StartSession 启动一个新的 session
func (c *RealClient) StartSession(opts ...*options.SessionOptions) (mongo.Session, error) {
	return c.Client.StartSession(opts...)
}

// Timeout 返回客户端的超时时间
func (c *RealClient) Timeout() *time.Duration {
	return c.Client.Timeout()
}

// UseSession 在提供的 session 上下文中执行函数
func (c *RealClient) UseSession(ctx context.Context, fn func(mongo.SessionContext) error) error {
	return c.Client.UseSession(ctx, fn)
}

// UseSessionWithOptions 在提供的 session 选项和上下文中执行函数
func (c *RealClient) UseSessionWithOptions(ctx context.Context, opts *options.SessionOptions, fn func(mongo.SessionContext) error) error {
	return c.Client.UseSessionWithOptions(ctx, opts, fn)
}

// Watch 创建一个变更流来监听客户端级别的变更
func (c *RealClient) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return c.Client.Watch(ctx, pipeline, opts...)
}

// RealDatabase 包装 *mongo.Database，实现 MongoDatabase 接口
// 直接委托所有方法给官方 MongoDB 驱动
type RealDatabase struct {
	*mongo.Database
	client *RealClient
}

// Aggregate 在数据库级别执行聚合管道
func (d *RealDatabase) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (MongoCursor, error) {
	cursor, err := d.Database.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, err
	}
	return &RealCursor{Cursor: cursor}, nil
}

// Client 返回创建此数据库的客户端
func (d *RealDatabase) Client() MongoClient {
	return d.client
}

// Collection 返回指定名称的集合实例
func (d *RealDatabase) Collection(name string, opts ...*options.CollectionOptions) MongoCollection {
	coll := d.Database.Collection(name, opts...)
	return &RealCollection{
		Collection: coll,
		database:   d,
	}
}

// CreateCollection 创建一个新集合
func (d *RealDatabase) CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error {
	return d.Database.CreateCollection(ctx, name, opts...)
}

// CreateView 创建一个视图
func (d *RealDatabase) CreateView(ctx context.Context, viewName, viewOn string, pipeline interface{}, opts ...*options.CreateViewOptions) error {
	return d.Database.CreateView(ctx, viewName, viewOn, pipeline, opts...)
}

// Drop 删除此数据库
func (d *RealDatabase) Drop(ctx context.Context) error {
	return d.Database.Drop(ctx)
}

// ListCollectionNames 返回此数据库中所有集合的名称
func (d *RealDatabase) ListCollectionNames(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) ([]string, error) {
	return d.Database.ListCollectionNames(ctx, filter, opts...)
}

// ListCollectionSpecifications 返回此数据库中所有集合的详细规范
func (d *RealDatabase) ListCollectionSpecifications(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) ([]*mongo.CollectionSpecification, error) {
	return d.Database.ListCollectionSpecifications(ctx, filter, opts...)
}

// ListCollections 返回此数据库中所有集合的游标
func (d *RealDatabase) ListCollections(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) (MongoCursor, error) {
	cursor, err := d.Database.ListCollections(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return &RealCursor{Cursor: cursor}, nil
}

// Name 返回此数据库的名称
func (d *RealDatabase) Name() string {
	return d.Database.Name()
}

// ReadConcern 返回此数据库的读关注
func (d *RealDatabase) ReadConcern() *readconcern.ReadConcern {
	return d.Database.ReadConcern()
}

// ReadPreference 返回此数据库的读偏好
func (d *RealDatabase) ReadPreference() *readpref.ReadPref {
	return d.Database.ReadPreference()
}

// RunCommand 在此数据库上执行命令
func (d *RealDatabase) RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) *mongo.SingleResult {
	return d.Database.RunCommand(ctx, runCommand, opts...)
}

// RunCommandCursor 在此数据库上执行命令并返回游标
func (d *RealDatabase) RunCommandCursor(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) (MongoCursor, error) {
	cursor, err := d.Database.RunCommandCursor(ctx, runCommand, opts...)
	if err != nil {
		return nil, err
	}
	return &RealCursor{Cursor: cursor}, nil
}

// Watch 创建一个变更流来监听数据库级别的变更
func (d *RealDatabase) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return d.Database.Watch(ctx, pipeline, opts...)
}

// WriteConcern 返回此数据库的写关注
func (d *RealDatabase) WriteConcern() *writeconcern.WriteConcern {
	return d.Database.WriteConcern()
}

// RealCollection 包装 *mongo.Collection，实现 MongoCollection 接口
// 直接委托所有方法给官方 MongoDB 驱动
type RealCollection struct {
	*mongo.Collection
	database *RealDatabase
}

// Aggregate 对集合执行聚合管道
func (c *RealCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (MongoCursor, error) {
	cursor, err := c.Collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		return nil, err
	}
	return &RealCursor{Cursor: cursor}, nil
}

// BulkWrite 执行批量写操作
func (c *RealCollection) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	return c.Collection.BulkWrite(ctx, models, opts...)
}

// Clone 创建此集合的副本
func (c *RealCollection) Clone(opts ...*options.CollectionOptions) (MongoCollection, error) {
	coll, err := c.Collection.Clone(opts...)
	if err != nil {
		return nil, err
	}
	return &RealCollection{
		Collection: coll,
		database:   c.database,
	}, nil
}

// CountDocuments 返回与过滤器匹配的文档数量
func (c *RealCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return c.Collection.CountDocuments(ctx, filter, opts...)
}

// Database 返回此集合所属的数据库
func (c *RealCollection) Database() MongoDatabase {
	return c.database
}

// DeleteMany 删除与过滤器匹配的所有文档
func (c *RealCollection) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.Collection.DeleteMany(ctx, filter, opts...)
}

// DeleteOne 删除与过滤器匹配的单个文档
func (c *RealCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	return c.Collection.DeleteOne(ctx, filter, opts...)
}

// Distinct 返回与过滤器匹配的文档中指定字段的所有不同值
func (c *RealCollection) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	return c.Collection.Distinct(ctx, fieldName, filter, opts...)
}

// Drop 删除此集合
func (c *RealCollection) Drop(ctx context.Context) error {
	return c.Collection.Drop(ctx)
}

// EstimatedDocumentCount 返回集合中文档的估计数量
func (c *RealCollection) EstimatedDocumentCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (int64, error) {
	return c.Collection.EstimatedDocumentCount(ctx, opts...)
}

// Find 查找与过滤器匹配的所有文档
func (c *RealCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (MongoCursor, error) {
	cursor, err := c.Collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	return &RealCursor{Cursor: cursor}, nil
}

// FindOne 查找与过滤器匹配的单个文档
func (c *RealCollection) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	return c.Collection.FindOne(ctx, filter, opts...)
}

// FindOneAndDelete 查找并删除与过滤器匹配的单个文档
func (c *RealCollection) FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {
	return c.Collection.FindOneAndDelete(ctx, filter, opts...)
}

// FindOneAndReplace 查找并替换与过滤器匹配的单个文档
func (c *RealCollection) FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {
	return c.Collection.FindOneAndReplace(ctx, filter, replacement, opts...)
}

// FindOneAndUpdate 查找并更新与过滤器匹配的单个文档
func (c *RealCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	return c.Collection.FindOneAndUpdate(ctx, filter, update, opts...)
}

// Indexes 返回集合的索引视图
func (c *RealCollection) Indexes() mongo.IndexView {
	return c.Collection.Indexes()
}

// InsertMany 插入多个文档
func (c *RealCollection) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	return c.Collection.InsertMany(ctx, documents, opts...)
}

// InsertOne 插入单个文档
func (c *RealCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	return c.Collection.InsertOne(ctx, document, opts...)
}

// Name 返回集合的名称
func (c *RealCollection) Name() string {
	return c.Collection.Name()
}

// ReplaceOne 替换与过滤器匹配的单个文档
func (c *RealCollection) ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}, opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	return c.Collection.ReplaceOne(ctx, filter, replacement, opts...)
}

// SearchIndexes 返回集合的搜索索引视图
func (c *RealCollection) SearchIndexes() mongo.SearchIndexView {
	return c.Collection.SearchIndexes()
}

// UpdateByID 通过 ID 更新单个文档
func (c *RealCollection) UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.Collection.UpdateByID(ctx, id, update, opts...)
}

// UpdateMany 更新与过滤器匹配的所有文档
func (c *RealCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.Collection.UpdateMany(ctx, filter, update, opts...)
}

// UpdateOne 更新与过滤器匹配的单个文档
func (c *RealCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	return c.Collection.UpdateOne(ctx, filter, update, opts...)
}

// Watch 创建一个变更流来监听集合级别的变更
func (c *RealCollection) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return c.Collection.Watch(ctx, pipeline, opts...)
}

// RealCursor 包装 *mongo.Cursor，实现 MongoCursor 接口
// 直接委托所有方法给官方 MongoDB 驱动
type RealCursor struct {
	*mongo.Cursor
}

// All 将游标中的所有文档解码到 results 中
func (c *RealCursor) All(ctx context.Context, results interface{}) error {
	return c.Cursor.All(ctx, results)
}

// Close 关闭游标
func (c *RealCursor) Close(ctx context.Context) error {
	return c.Cursor.Close(ctx)
}

// Decode 将当前文档解码到 val 中
func (c *RealCursor) Decode(val interface{}) error {
	return c.Cursor.Decode(val)
}

// Err 返回游标的错误
func (c *RealCursor) Err() error {
	return c.Cursor.Err()
}

// ID 返回游标的 ID
func (c *RealCursor) ID() int64 {
	return c.Cursor.ID()
}

// Next 将游标移动到下一个文档
func (c *RealCursor) Next(ctx context.Context) bool {
	return c.Cursor.Next(ctx)
}

// RemainingBatchLength 返回当前批次中剩余的文档数量
func (c *RealCursor) RemainingBatchLength() int {
	return c.Cursor.RemainingBatchLength()
}

// SetBatchSize 设置游标的批次大小
func (c *RealCursor) SetBatchSize(batchSize int32) {
	c.Cursor.SetBatchSize(batchSize)
}

// SetComment 设置游标的注释
func (c *RealCursor) SetComment(comment interface{}) {
	c.Cursor.SetComment(comment)
}

// SetMaxTime 设置游标的最大执行时间
func (c *RealCursor) SetMaxTime(dur time.Duration) {
	c.Cursor.SetMaxTime(dur)
}

// TryNext 尝试将游标移动到下一个文档
func (c *RealCursor) TryNext(ctx context.Context) bool {
	return c.Cursor.TryNext(ctx)
}

// Current 返回当前文档的 BSON 数据
func (c *RealCursor) Current() bson.Raw {
	return c.Cursor.Current
}

package mongo

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// FindResult contains the result of a Find operation
type FindResult struct {
	DocsAffected int64
	Error        error
}

// FirstResult contains the result of a First operation
type FirstResult struct {
	Error error
}

// DeleteResult contains the result of a Delete operation
type DeleteResult struct {
	DocsAffected int64
	Error        error
}

// UpdateResult contains the result of an Update operation
type UpdateResult struct {
	DocsAffected int64
	Error        error
}

// Scoop is a generic MongoDB query builder for simplified operations
type Scoop struct {
	client        *Client
	coll          MongoCollection // 使用 MongoCollection interface 便于测试和扩展
	filter        *Cond
	limit         *int64
	offset        *int64
	sort          bson.M
	projection    bson.M
	session       mongo.Session
	notFoundError error
	depth         int
	logger        Logger
}

// NewScoop creates a new scoop instance, optionally accepting a transaction scoop
func (p *Client) NewScoop(tx ...*Scoop) *Scoop {
	scoop := &Scoop{
		client:        p,
		filter:        NewCond(),
		sort:          bson.M{},
		projection:    bson.M{},
		notFoundError: mongo.ErrNoDocuments,
		depth:         3,
		logger:        GetDefaultLogger(),
	}

	// If a transaction scoop is provided, inherit its session and notFoundError
	if len(tx) > 0 && tx[0] != nil {
		scoop.session = tx[0].session
		scoop.logger = tx[0].logger
	}

	return scoop
}

// getCollectionNameFromOut 从 out 参数的类型推导集合名
func (s *Scoop) getCollectionNameFromOut(out interface{}) string {
	if out == nil {
		return ""
	}

	// 反射获取类型
	outType := reflect.TypeOf(out)

	// 如果是指针，获取指向的类型
	for outType.Kind() == reflect.Ptr {
		outType = outType.Elem()
	}

	// 如果是 slice，获取元素类型
	if outType.Kind() == reflect.Slice || outType.Kind() == reflect.Array {
		outType = outType.Elem()
		for outType.Kind() == reflect.Ptr {
			outType = outType.Elem()
		}
	}

	// 尝试调用 Collection() 方法
	outValue := reflect.New(outType)
	if m := outValue.MethodByName("Collection"); m.IsValid() {
		result := m.Call(nil)
		if len(result) > 0 && result[0].Kind() == reflect.String {
			if collName := result[0].String(); collName != "" {
				return collName
			}
		}
	}

	// 默认使用类型名称
	return outType.Name()
}

// getCollection retrieves a MongoDB collection using MGM or client.db in Mock mode
// 返回 MongoCollection interface 便于测试和扩展
func (s *Scoop) getCollection(collName string) MongoCollection {
	// In Mock mode, use client.db.Collection() instead of MGM
	if s.client.cfg.Mock && s.client.db != nil {
		return s.client.db.Collection(collName)
	}

	// Real mode - use MGM
	mgmColl := mgm.CollectionByName(collName)
	// MGM Collection embeds *mongo.Collection，需要包装成 RealCollection
	return &RealCollection{
		Collection: mgmColl.Collection,
		database:   nil, // MGM 不提供 database 引用，这里可以为 nil
	}
}

// ensureCollection 确保 Scoop 有关联的集合
func (s *Scoop) ensureCollection(out interface{}) error {
	if s.coll != nil {
		return nil
	}

	collName := s.getCollectionNameFromOut(out)
	if collName == "" {
		return fmt.Errorf("cannot determine collection name from out parameter")
	}

	s.coll = s.getCollection(collName)
	return nil
}

// Collection 设置要操作的集合
// getContext returns the appropriate context for this scoop
// If a session exists, it wraps the session in a SessionContext
// Otherwise, returns a background context
func (s *Scoop) getContext() context.Context {
	if s.session != nil {
		return mongo.NewSessionContext(context.Background(), s.session)
	}
	return context.Background()
}

func (s *Scoop) Collection(model interface{}) *Scoop {
	collName := s.getCollectionNameFromOut(model)
	if collName != "" {
		s.coll = s.getCollection(collName)
	}
	return s
}

// CollectionName sets the collection by name string
func (s *Scoop) CollectionName(name string) *Scoop {
	if name != "" {
		s.coll = s.getCollection(name)
	}
	return s
}

// Where adds filter conditions to the scoop
// Supports multiple calling forms:
//   - Where(key string, value interface{}) - simple equality condition
//   - Where(key string, op string, value interface{}) - with operator
//   - Where(cond *Cond) - condition builder
func (s *Scoop) Where(args ...interface{}) *Scoop {
	if len(args) == 0 {
		return s
	}

	// Delegate to Cond.where() which handles all cases including *Cond
	s.filter.where(args...)
	return s
}

// Equal adds an equality condition
func (s *Scoop) Equal(key string, value interface{}) *Scoop {
	s.filter.Equal(key, value)
	return s
}

// Ne adds a != condition using $ne operator
func (s *Scoop) Ne(key string, value interface{}) *Scoop {
	s.filter.Ne(key, value)
	return s
}

// In adds an $in condition
func (s *Scoop) In(key string, values ...interface{}) *Scoop {
	s.filter.In(key, values...)
	return s
}

// NotIn adds a $nin condition
func (s *Scoop) NotIn(key string, values ...interface{}) *Scoop {
	s.filter.NotIn(key, values...)
	return s
}

// Like adds a regex pattern match (case-insensitive)
func (s *Scoop) Like(key string, pattern string) *Scoop {
	s.filter.Like(key, pattern)
	return s
}

// Gt adds a > condition using $gt operator
func (s *Scoop) Gt(key string, value interface{}) *Scoop {
	s.filter.Gt(key, value)
	return s
}

// Lt adds a < condition using $lt operator
func (s *Scoop) Lt(key string, value interface{}) *Scoop {
	s.filter.Lt(key, value)
	return s
}

// Gte adds a >= condition using $gte operator
func (s *Scoop) Gte(key string, value interface{}) *Scoop {
	s.filter.Gte(key, value)
	return s
}

// Lte adds a <= condition using $lte operator
func (s *Scoop) Lte(key string, value interface{}) *Scoop {
	s.filter.Lte(key, value)
	return s
}

// Between adds a $gte and $lte condition
func (s *Scoop) Between(key string, min interface{}, max interface{}) *Scoop {
	s.filter.Between(key, min, max)
	return s
}

// Limit sets the limit
func (s *Scoop) Limit(limit int64) *Scoop {
	s.limit = &limit
	return s
}

// Offset sets the offset/skip
func (s *Scoop) Offset(offset int64) *Scoop {
	s.offset = &offset
	return s
}

// Skip is an alias for Offset
func (s *Scoop) Skip(skip int64) *Scoop {
	return s.Offset(skip)
}

// Sort adds sorting
// direction: 1 for ascending (default), -1 for descending
func (s *Scoop) Sort(key string, direction ...int) *Scoop {
	dir := 1 // default ascending
	if len(direction) > 0 {
		dir = direction[0]
	}
	s.sort[key] = dir
	return s
}

// Select specifies which fields to return
func (s *Scoop) Select(fields ...string) *Scoop {
	for _, field := range fields {
		s.projection[field] = 1
	}
	return s
}

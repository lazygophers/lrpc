package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/kamva/mgm/v3"
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	coll          *mongo.Collection
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
func (c *Client) NewScoop(tx ...*Scoop) *Scoop {
	scoop := &Scoop{
		client:        c,
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

// getCollection retrieves a MongoDB collection using MGM
func (s *Scoop) getCollection(collName string) *mongo.Collection {
	mgmColl := mgm.CollectionByName(collName)
	// MGM Collection embeds *mongo.Collection
	return mgmColl.Collection
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

// Find finds documents matching the filter and returns a FindResult
func (s *Scoop) Find(result interface{}) *FindResult {
	// Check for injected failures (test only)
	injector := GetGlobalInjector()
	if injector.ShouldFailFind() {
		return &FindResult{
			Error: injector.GetFindError(),
		}
	}

	begin := time.Now()
	var docsCount int64

	err := s.ensureCollection(result)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.find(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FindResult{Error: err}
	}

	// Build FindOptions from scoop fields
	opts := options.Find()
	if s.limit != nil {
		opts.SetLimit(*s.limit)
	}
	if s.offset != nil {
		opts.SetSkip(*s.offset)
	}
	if len(s.sort) > 0 {
		opts.SetSort(s.sort)
	}
	if len(s.projection) > 0 {
		opts.SetProjection(s.projection)
	}

	ctx := s.getContext()
	cursor, err := s.coll.Find(ctx, s.filter.ToBson(), opts)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.find(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FindResult{Error: err}
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, result)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.find(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FindResult{Error: err}
	}

	// Count the documents returned
	if resultVal := reflect.ValueOf(result); resultVal.Kind() == reflect.Ptr && resultVal.Elem().Kind() == reflect.Slice {
		docsCount = int64(resultVal.Elem().Len())
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.find(%v)", s.coll.Name(), s.filter.ToBson()), docsCount
	}, nil)

	return &FindResult{
		DocsAffected: docsCount,
		Error:        nil,
	}
}

// First finds a single document and returns a FirstResult
func (s *Scoop) First(result interface{}) *FirstResult {
	begin := time.Now()

	err := s.ensureCollection(result)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.findOne(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FirstResult{Error: err}
	}

	ctx := s.getContext()
	opts := options.FindOne()
	if len(s.projection) > 0 {
		opts.SetProjection(s.projection)
	}
	sr := s.coll.FindOne(ctx, s.filter.ToBson(), opts)
	if sr.Err() != nil {
		log.Errorf("err:%v", sr.Err())
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.findOne(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, sr.Err())
		return &FirstResult{Error: sr.Err()}
	}

	err = sr.Decode(result)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.findOne(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &FirstResult{Error: err}
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.findOne(%v)", s.coll.Name(), s.filter.ToBson()), 1
	}, nil)

	return &FirstResult{Error: nil}
}

// Count counts documents matching the filter
func (s *Scoop) Count() (int64, error) {
	begin := time.Now()

	// Check for injected failures (test only)
	injector := GetGlobalInjector()
	if injector.ShouldFailCount() {
		return 0, injector.GetCountError()
	}

	count, err := s.coll.CountDocuments(s.getContext(), s.filter.ToBson())
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.countDocuments(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return 0, err
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.countDocuments(%v)", s.coll.Name(), s.filter.ToBson()), count
	}, nil)

	return count, nil
}

// Exist checks if documents matching the filter exist by fetching only _id field
func (s *Scoop) Exist() (bool, error) {
	// Clone scoop and select only _id field for efficiency
	scoop := s.Clone()
	scoop.Select("_id")

	ctx := scoop.getContext()
	opts := options.FindOne()
	opts.SetProjection(scoop.projection)

	sr := scoop.coll.FindOne(ctx, scoop.filter.ToBson(), opts)
	if sr.Err() != nil {
		if sr.Err() == mongo.ErrNoDocuments {
			return false, nil
		}
		log.Errorf("err:%v", sr.Err())
		return false, sr.Err()
	}

	return true, nil
}

// Create inserts a new document
func (s *Scoop) Create(doc interface{}) error {
	begin := time.Now()

	err := s.ensureCollection(doc)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	_, err = s.coll.InsertOne(s.getContext(), doc)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.insertOne(...)", s.coll.Name()), 0
		}, err)
		return err
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.insertOne(...)", s.coll.Name()), 1
	}, nil)

	return nil
}

// BatchCreate inserts multiple documents
func (s *Scoop) BatchCreate(docs ...interface{}) error {
	begin := time.Now()

	if len(docs) == 0 {
		err := fmt.Errorf("no documents to insert")
		log.Errorf("err:%v", err)
		return err
	}

	err := s.ensureCollection(docs[0])
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	result, err := s.coll.InsertMany(s.getContext(), docs)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.insertMany(...) [%d docs]", s.coll.Name(), len(docs)), 0
		}, err)
		return err
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.insertMany(...) [%d docs]", s.coll.Name(), len(docs)), int64(len(result.InsertedIDs))
	}, nil)

	return nil
}

// Update updates documents matching the filter and returns an UpdateResult
func (s *Scoop) Update(update interface{}) *UpdateResult {
	begin := time.Now()

	if s.coll == nil {
		err := fmt.Errorf("collection not set, call Collection(model) or use Find/Create first")
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db collection not set"), 0
		}, err)
		return &UpdateResult{Error: err}
	}

	updateDoc := bson.M{}

	// If update is a map, wrap it in $set
	switch v := update.(type) {
	case bson.M:
		if _, ok := v["$set"]; !ok && len(v) > 0 {
			updateDoc = bson.M{"$set": v}
		} else {
			updateDoc = v
		}
	case map[string]interface{}:
		// Check if it has update operators
		hasOperator := false
		for key := range v {
			if len(key) > 0 && key[0] == '$' {
				hasOperator = true
				break
			}
		}
		if !hasOperator {
			updateDoc = bson.M{"$set": v}
		} else {
			updateDoc = bson.M(v)
		}
	default:
		// Convert to map via JSON
		data, err := json.Marshal(v)
		if err != nil {
			log.Errorf("err:%v", err)
			s.logger.Log(s.depth, begin, func() (string, int64) {
				return fmt.Sprintf("db.%s marshal update", s.coll.Name()), 0
			}, err)
			return &UpdateResult{Error: err}
		}

		var m map[string]interface{}
		err = json.Unmarshal(data, &m)
		if err != nil {
			log.Errorf("err:%v", err)
			s.logger.Log(s.depth, begin, func() (string, int64) {
				return fmt.Sprintf("db.%s unmarshal update", s.coll.Name()), 0
			}, err)
			return &UpdateResult{Error: err}
		}

		updateDoc = bson.M{"$set": m}
	}

	result, err := s.coll.UpdateMany(s.getContext(), s.filter.ToBson(), updateDoc)
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.updateMany(%v, %v)", s.coll.Name(), s.filter.ToBson(), updateDoc), 0
		}, err)
		return &UpdateResult{Error: err}
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.updateMany(%v, %v)", s.coll.Name(), s.filter.ToBson(), updateDoc), result.ModifiedCount
	}, nil)

	return &UpdateResult{
		DocsAffected: result.ModifiedCount,
		Error:        nil,
	}
}

// Delete deletes documents matching the filter and returns a DeleteResult
func (s *Scoop) Delete() *DeleteResult {
	begin := time.Now()

	// Check for injected failures (test only)
	injector := GetGlobalInjector()
	if injector.ShouldFailDelete() {
		err := injector.GetDeleteError()
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.deleteMany(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &DeleteResult{Error: err}
	}

	result, err := s.coll.DeleteMany(s.getContext(), s.filter.ToBson())
	if err != nil {
		log.Errorf("err:%v", err)
		s.logger.Log(s.depth, begin, func() (string, int64) {
			return fmt.Sprintf("db.%s.deleteMany(%v)", s.coll.Name(), s.filter.ToBson()), 0
		}, err)
		return &DeleteResult{Error: err}
	}

	s.logger.Log(s.depth, begin, func() (string, int64) {
		return fmt.Sprintf("db.%s.deleteMany(%v)", s.coll.Name(), s.filter.ToBson()), result.DeletedCount
	}, nil)

	return &DeleteResult{
		DocsAffected: result.DeletedCount,
		Error:        nil,
	}
}

// Aggregate creates an aggregation pipeline
func (s *Scoop) Aggregate(pipeline ...bson.M) *Aggregation {
	return NewAggregation(s.coll, s.getContext(), pipeline...)
}

// Clone creates a copy of the scoop with current state
func (s *Scoop) Clone() *Scoop {
	newScoop := &Scoop{
		client:        s.client,
		coll:          s.coll,
		filter:        NewCond(),
		sort:          bson.M{},
		projection:    bson.M{},
		session:       s.session,
		notFoundError: s.notFoundError,
		depth:         s.depth,
		logger:        s.logger,
	}

	// Deep copy filter conditions
	if s.filter != nil && len(s.filter.conds) > 0 {
		newScoop.filter.conds = make([]bson.M, len(s.filter.conds))
		for i, cond := range s.filter.conds {
			// Deep copy each BSON condition
			newScoop.filter.conds[i] = make(bson.M)
			for k, v := range cond {
				newScoop.filter.conds[i][k] = v
			}
		}
		newScoop.filter.isOr = s.filter.isOr
	}

	// Copy limit and offset
	if s.limit != nil {
		newScoop.limit = s.limit
	}
	if s.offset != nil {
		newScoop.offset = s.offset
	}

	// Deep copy sort
	if len(s.sort) > 0 {
		newScoop.sort = make(bson.M)
		for k, v := range s.sort {
			newScoop.sort[k] = v
		}
	}

	// Deep copy projection
	if len(s.projection) > 0 {
		newScoop.projection = make(bson.M)
		for k, v := range s.projection {
			newScoop.projection[k] = v
		}
	}

	return newScoop
}

// Clear resets the scoop
func (s *Scoop) Clear() *Scoop {
	s.filter = NewCond()
	s.limit = nil
	s.offset = nil
	s.sort = bson.M{}
	s.projection = bson.M{}
	return s
}

// GetCollection returns the underlying MongoDB collection
func (s *Scoop) GetCollection() *mongo.Collection {
	return s.coll
}

// SetNotFound sets the not found error for this scoop
func (s *Scoop) SetNotFound(err error) *Scoop {
	s.notFoundError = err
	return s
}

// IsNotFound checks if the error is a not found error
func (s *Scoop) IsNotFound(err error) bool {
	return err == s.notFoundError || err == mongo.ErrNoDocuments
}

// Begin starts a transaction - creates session lazily if needed
func (s *Scoop) Begin() (*Scoop, error) {
	// Check for injected failures (test only)
	injector := GetGlobalInjector()
	if injector.ShouldFailTransaction() {
		return nil, injector.GetTransactionError()
	}

	// Lazy initialization: create session only when needed
	if s.session == nil {
		_, mongoClient, _, err := mgm.DefaultConfigs()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		session, err := mongoClient.StartSession()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}

		s.session = session
	}

	// Start transaction on the session
	err := s.session.StartTransaction()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Return a new Scoop with the same session (for transactional operations)
	newScoop := &Scoop{
		client:     s.client,
		coll:       s.coll,
		filter:     NewCond(),
		sort:       bson.M{},
		projection: bson.M{},
		session:    s.session,
	}

	return newScoop, nil
}

// Commit commits the transaction
func (s *Scoop) Commit() error {
	if s.session == nil {
		return fmt.Errorf("no active transaction")
	}

	err := s.session.CommitTransaction(context.Background())
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	s.session.EndSession(context.Background())
	return nil
}

// Rollback aborts/rolls back the transaction
func (s *Scoop) Rollback() error {
	if s.session == nil {
		return fmt.Errorf("no active transaction")
	}

	err := s.session.AbortTransaction(context.Background())
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	s.session.EndSession(context.Background())
	return nil
}

// inc increments the depth counter
func (s *Scoop) inc() {
	s.depth++
}

// dec decrements the depth counter
func (s *Scoop) dec() {
	s.depth--
}

// FindByPage finds documents matching the filter with pagination support
// Returns paginated results along with total count if ShowTotal is true
func (s *Scoop) FindByPage(opt *core.ListOption, values any) (*core.Paginate, error) {
	if opt == nil {
		opt = &core.ListOption{
			Offset: core.DefaultOffset,
			Limit:  core.DefaultLimit,
		}
	}

	s.Offset(int64(opt.Offset)).Limit(int64(opt.Limit))

	page := &core.Paginate{
		Offset: opt.Offset,
		Limit:  opt.Limit,
	}

	s.inc()
	defer s.dec()

	findResult := s.Find(values)
	if findResult.Error != nil {
		log.Errorf("err:%v", findResult.Error)
		return nil, findResult.Error
	}

	if opt.ShowTotal {
		// Create a new scoop for counting to avoid modifying the current one's state
		countScoop := s.Clone()
		count, err := countScoop.Count()
		if err != nil {
			log.Errorf("err:%v", err)
			return nil, err
		}
		page.Total = uint64(count)
	}

	return page, nil
}

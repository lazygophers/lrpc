package mock

import (
	"context"
	"fmt"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
	gomongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// MockDatabase implements MongoDatabase interface for testing
// It provides an in-memory database implementation backed by MemoryStorage
type MockDatabase struct {
	name    string
	storage *MemoryStorage
	client  mongo.MongoClient
}

// NewMockDatabase creates a new MockDatabase instance
// Parameters:
//   - name: database name
//   - storage: MemoryStorage backend (shared with MockClient)
//   - client: parent client reference
func NewMockDatabase(name string, storage *MemoryStorage, client mongo.MongoClient) *MockDatabase {
	return &MockDatabase{
		name:    name,
		storage: storage,
		client:  client,
	}
}

// Aggregate executes an aggregation pipeline at database level
// Database-level aggregation operates across all collections
// Returns ErrNotImplemented as database-level aggregation requires collection specification
func (d *MockDatabase) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (mongo.MongoCursor, error) {
	// Database-level aggregation typically requires collection specification in the pipeline
	// For simplicity, return not implemented error
	err := ErrNotImplemented
	log.Errorf("err:%v", err)
	return nil, err
}

// Client returns the client that created this database
func (d *MockDatabase) Client() mongo.MongoClient {
	return d.client
}

// Collection returns a collection instance with the specified name
// The collection shares the same storage as the database
func (d *MockDatabase) Collection(name string, opts ...*options.CollectionOptions) mongo.MongoCollection {
	log.Debugf("getting collection: %s from database: %s", name, d.name)
	return NewMockCollection(name, d.storage, d)
}

// CreateCollection creates a new collection in the database
// Returns error if collection already exists
func (d *MockDatabase) CreateCollection(ctx context.Context, name string, opts ...*options.CreateCollectionOptions) error {
	log.Infof("creating collection: %s in database: %s", name, d.name)

	err := d.storage.CreateCollection(name)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

// CreateView creates a view in the database
// Mock mode: views are not actually created, only logged
// Returns nil to indicate success (no actual storage needed)
func (d *MockDatabase) CreateView(ctx context.Context, viewName, viewOn string, pipeline interface{}, opts ...*options.CreateViewOptions) error {
	log.Debugf("creating view %s on collection %s (mock mode)", viewName, viewOn)
	log.Debugf("view pipeline: %+v", pipeline)

	// Mock mode下不实际创建视图，只记录日志
	// 视图功能在测试场景下通常不重要
	return nil
}

// Drop deletes this database and all its collections
// Note: This implementation clears all collections in storage
// In a real scenario with multiple databases, this should only clear collections in this database
func (d *MockDatabase) Drop(ctx context.Context) error {
	log.Infof("dropping database: %s", d.name)

	// Get all collection names
	collNames := d.storage.ListCollections()

	// Drop each collection
	for _, collName := range collNames {
		err := d.storage.DropCollection(collName)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	log.Infof("dropped database: %s with %d collections", d.name, len(collNames))
	return nil
}

// ListCollectionNames returns the names of all collections in this database
// Supports filter to match collection names (filter support is basic in mock)
func (d *MockDatabase) ListCollectionNames(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) ([]string, error) {
	log.Debugf("listing collection names in database: %s", d.name)

	// Get all collection names from storage
	names := d.storage.ListCollections()

	// TODO: Apply filter if needed
	// For now, we return all collections
	// In a production mock, we might want to filter based on the filter parameter

	log.Debugf("found %d collections in database: %s", len(names), d.name)
	return names, nil
}

// ListCollectionSpecifications returns detailed specifications for all collections
// Returns basic specifications with name and type information
func (d *MockDatabase) ListCollectionSpecifications(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) ([]*gomongo.CollectionSpecification, error) {
	log.Debugf("listing collection specifications in database: %s", d.name)

	names, err := d.ListCollectionNames(ctx, filter, opts...)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Create specifications for each collection
	specs := make([]*gomongo.CollectionSpecification, 0, len(names))
	for _, name := range names {
		specs = append(specs, &gomongo.CollectionSpecification{
			Name: name,
			Type: "collection",
		})
	}

	log.Debugf("created %d collection specifications", len(specs))
	return specs, nil
}

// ListCollections returns a cursor over all collections in this database
// The cursor contains collection information documents
func (d *MockDatabase) ListCollections(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) (mongo.MongoCursor, error) {
	log.Debugf("listing collections in database: %s", d.name)

	names, err := d.ListCollectionNames(ctx, filter, opts...)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Create documents for each collection
	docs := make([]bson.M, 0, len(names))
	for _, name := range names {
		docs = append(docs, bson.M{
			"name": name,
			"type": "collection",
		})
	}

	cursor := NewMockCursor(docs)
	log.Debugf("created cursor with %d collection documents", len(docs))
	return cursor, nil
}

// Name returns the name of this database
func (d *MockDatabase) Name() string {
	return d.name
}

// ReadConcern returns the read concern for this database
// Returns nil as mock doesn't support read concerns
func (d *MockDatabase) ReadConcern() *readconcern.ReadConcern {
	return nil
}

// ReadPreference returns the read preference for this database
// Returns nil as mock doesn't support read preferences
func (d *MockDatabase) ReadPreference() *readpref.ReadPref {
	return nil
}

// RunCommand executes a command on this database
// Returns an empty SingleResult as mock doesn't support commands
func (d *MockDatabase) RunCommand(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) *gomongo.SingleResult {
	log.Warnf("RunCommand is not fully implemented in mock, returning empty result")

	// Return an empty SingleResult
	// In a real implementation, this would execute the command and return results
	return gomongo.NewSingleResultFromDocument(bson.M{}, nil, nil)
}

// RunCommandCursor executes a command on this database and returns a cursor
// Mock mode: returns an empty cursor for most commands
// Supports basic commands like listCollections for testing purposes
func (d *MockDatabase) RunCommandCursor(ctx context.Context, runCommand interface{}, opts ...*options.RunCmdOptions) (mongo.MongoCursor, error) {
	log.Debugf("running command cursor: %+v (mock mode)", runCommand)

	// Mock 模式下返回空 cursor
	// 对于测试场景，大多数命令返回空结果即可
	cursor := NewMockCursor([]bson.M{})
	return cursor, nil
}

// Watch creates a change stream to monitor database-level changes
// Returns ErrNotImplemented as mongo.ChangeStream is a concrete type that cannot be mocked
// Use WatchMock() instead for testing with change streams
func (d *MockDatabase) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*gomongo.ChangeStream, error) {
	err := ErrNotImplemented
	log.Errorf("err:%v - use WatchMock() instead for testing", err)
	return nil, err
}

// WatchMock creates a mock change stream to monitor all changes in this database
// This is a testing-specific method that returns MockChangeStream directly
// Use this method in tests instead of Watch()
// Note: Since MemoryStorage doesn't track database namespaces, this watches all collections
func (d *MockDatabase) WatchMock(ctx context.Context) *MockChangeStream {
	log.Debugf("creating mock change stream for database: %s (all collections)", d.name)

	// No filter - watch all collections (database isolation not implemented)
	filter := ChangeStreamFilter{}

	// Create the change stream
	stream := NewMockChangeStream(ctx, filter, 100)

	// Register with storage
	d.storage.registerChangeStream(stream)

	return stream
}

// WriteConcern returns the write concern for this database
// Returns nil as mock doesn't support write concerns
func (d *MockDatabase) WriteConcern() *writeconcern.WriteConcern {
	return nil
}

// ensureDatabaseNamespace ensures database name is used as prefix for collections
// This is a helper method to support multi-database scenarios in the future
func (d *MockDatabase) ensureDatabaseNamespace() {
	// TODO: In a future implementation, we might want to prefix collection names
	// with the database name to support multiple databases in the same storage
	// For now, all collections are stored at the global level
}

// getCollectionFullName returns the full collection name including database prefix
// This helper prepares for future multi-database support
func (d *MockDatabase) getCollectionFullName(collectionName string) string {
	// For future enhancement: return fmt.Sprintf("%s.%s", d.name, collectionName)
	// For now, we use the collection name directly
	return collectionName
}

// validateCollectionName validates that a collection name is valid
// Returns error if name is empty or invalid
func (d *MockDatabase) validateCollectionName(name string) error {
	if name == "" {
		err := fmt.Errorf("collection name cannot be empty")
		log.Errorf("err:%v", err)
		return err
	}

	// Add more validation rules if needed
	// MongoDB has specific rules for collection names (no $, no null bytes, etc.)

	return nil
}

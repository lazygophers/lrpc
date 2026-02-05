package mock

import (
	"context"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
	gomongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MockCollection implements MongoCollection interface for testing
// It provides an in-memory collection implementation backed by MemoryStorage
type MockCollection struct {
	name     string
	storage  *MemoryStorage
	database mongo.MongoDatabase
}

// NewMockCollection creates a new MockCollection instance
// Parameters:
//   - name: collection name
//   - storage: MemoryStorage backend
//   - database: parent database reference
func NewMockCollection(name string, storage *MemoryStorage, database mongo.MongoDatabase) *MockCollection {
	return &MockCollection{
		name:     name,
		storage:  storage,
		database: database,
	}
}

// Name returns the name of the collection
func (m *MockCollection) Name() string {
	return m.name
}

// Database returns the database that this collection belongs to
func (m *MockCollection) Database() mongo.MongoDatabase {
	return m.database
}

// Drop drops the entire collection
func (m *MockCollection) Drop(ctx context.Context) error {
	err := m.storage.DropCollection(m.name)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

// EstimatedDocumentCount returns an estimate of the count of documents in the collection
// In mock implementation, returns the exact count
func (m *MockCollection) EstimatedDocumentCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (int64, error) {
	count, err := m.storage.Count(m.name, bson.M{})
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return count, nil
}

// CountDocuments returns the count of documents that match the filter
// Supports standard MongoDB filters
func (m *MockCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	count, err := m.storage.Count(m.name, filterDoc)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return count, nil
}

// Clone creates a copy of this collection
// Returns a new MockCollection instance sharing the same storage
func (m *MockCollection) Clone(opts ...*options.CollectionOptions) (mongo.MongoCollection, error) {
	return &MockCollection{
		name:     m.name,
		storage:  m.storage,
		database: m.database,
	}, nil
}

// Indexes returns the index view for this collection
// Mock mode does not support indexes, returns zero value
func (m *MockCollection) Indexes() gomongo.IndexView {
	log.Warnf("Mock mode does not support indexes, returning zero value")
	return gomongo.IndexView{}
}

// SearchIndexes returns the search index view for this collection
// Mock mode does not support search indexes, returns zero value
func (m *MockCollection) SearchIndexes() gomongo.SearchIndexView {
	log.Warnf("Mock mode does not support search indexes, returning zero value")
	return gomongo.SearchIndexView{}
}

// ptrInt64 is a helper function to create a pointer to int64
func ptrInt64(v int64) *int64 {
	return &v
}

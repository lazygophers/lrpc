package mock

import (
	"context"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
	gomongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MockClient implements MongoClient interface for testing
// It provides an in-memory MongoDB client implementation backed by MemoryStorage
type MockClient struct {
	storage *MemoryStorage
}

// NewMockClient creates a new MockClient instance with a fresh MemoryStorage
// Returns a MongoClient interface that can be used in tests
func NewMockClient() mongo.MongoClient {
	log.Debugf("creating new MockClient with fresh MemoryStorage")
	return &MockClient{
		storage: NewMemoryStorage(),
	}
}

// Connect initializes the client connection pool and starts background monitoring goroutines
// In mock implementation, this is a no-op that always succeeds
func (c *MockClient) Connect(ctx context.Context) error {
	log.Debugf("MockClient.Connect() called - simulating successful connection")
	return nil
}

// Database returns a database instance with the specified name
// The database shares the same MemoryStorage as the client
// All databases created from the same client share the same storage backend
func (c *MockClient) Database(name string, opts ...*options.DatabaseOptions) mongo.MongoDatabase {
	log.Debugf("MockClient.Database() called with name: %s", name)
	return NewMockDatabase(name, c.storage, c)
}

// Disconnect closes all connections to the MongoDB deployment
// In mock implementation, this is a no-op that always succeeds
func (c *MockClient) Disconnect(ctx context.Context) error {
	log.Debugf("MockClient.Disconnect() called - simulating successful disconnection")
	return nil
}

// ListDatabaseNames returns the names of all databases
// In mock implementation, returns a single database name since we don't track databases separately
// In a future enhancement, we could track databases in MemoryStorage
func (c *MockClient) ListDatabaseNames(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) ([]string, error) {
	log.Debugf("MockClient.ListDatabaseNames() called")

	// Since our MemoryStorage doesn't track databases separately,
	// we return a default database name if collections exist
	collections := c.storage.ListCollections()
	if len(collections) > 0 {
		// If we have collections, return a default database name
		return []string{"mock_database"}, nil
	}

	// No collections means no databases
	return []string{}, nil
}

// ListDatabases returns detailed information about all databases
// In mock implementation, returns basic information for databases that have collections
func (c *MockClient) ListDatabases(ctx context.Context, filter interface{}, opts ...*options.ListDatabasesOptions) (gomongo.ListDatabasesResult, error) {
	log.Debugf("MockClient.ListDatabases() called")

	names, err := c.ListDatabaseNames(ctx, filter, opts...)
	if err != nil {
		log.Errorf("err:%v", err)
		return gomongo.ListDatabasesResult{}, err
	}

	// Create database specifications
	databases := make([]gomongo.DatabaseSpecification, 0, len(names))
	for _, name := range names {
		databases = append(databases, gomongo.DatabaseSpecification{
			Name:       name,
			SizeOnDisk: 0, // Mock doesn't track disk size
			Empty:      false,
		})
	}

	result := gomongo.ListDatabasesResult{
		Databases: databases,
		TotalSize: 0, // Mock doesn't track total size
	}

	log.Debugf("returning %d databases", len(databases))
	return result, nil
}

// NumberSessionsInProgress returns the number of sessions currently in progress
// In mock implementation, always returns 0 since sessions are not fully supported
func (c *MockClient) NumberSessionsInProgress() int {
	log.Debugf("MockClient.NumberSessionsInProgress() called - returning 0")
	return 0
}

// Ping verifies the connection to the MongoDB deployment
// In mock implementation, this is a no-op that always succeeds
func (c *MockClient) Ping(ctx context.Context, rp *readpref.ReadPref) error {
	log.Debugf("MockClient.Ping() called - simulating successful ping")
	return nil
}

// StartSession starts a new session
// Returns ErrNotImplemented as full session support is not possible in mock
// due to mongo.Session interface having unexported methods
// See README.md for alternatives when testing session-based code
func (c *MockClient) StartSession(opts ...*options.SessionOptions) (gomongo.Session, error) {
	err := ErrNotImplemented
	log.Errorf("err:%v", err)
	return nil, err
}

// Timeout returns the client's timeout setting
// In mock implementation, returns nil as we don't support timeouts
func (c *MockClient) Timeout() *time.Duration {
	log.Debugf("MockClient.Timeout() called - returning nil")
	return nil
}

// UseSession executes a function within a session context
// Returns ErrNotImplemented as full session support is not possible in mock
// See README.md for alternatives when testing session-based code
func (c *MockClient) UseSession(ctx context.Context, fn func(gomongo.SessionContext) error) error {
	err := ErrNotImplemented
	log.Errorf("err:%v", err)
	return err
}

// UseSessionWithOptions executes a function within a session context with options
// Returns ErrNotImplemented as full session support is not possible in mock
// See README.md for alternatives when testing session-based code
func (c *MockClient) UseSessionWithOptions(ctx context.Context, opts *options.SessionOptions, fn func(gomongo.SessionContext) error) error {
	err := ErrNotImplemented
	log.Errorf("err:%v", err)
	return err
}

// Watch creates a change stream to monitor client-level changes
// Returns ErrNotImplemented as mongo.ChangeStream is a concrete type that cannot be mocked
// Use WatchMock() instead for testing with change streams
func (c *MockClient) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*gomongo.ChangeStream, error) {
	err := ErrNotImplemented
	log.Errorf("err:%v - use WatchMock() instead for testing", err)
	return nil, err
}

// WatchMock creates a mock change stream to monitor all changes across all collections
// This is a testing-specific method that returns MockChangeStream directly
// Use this method in tests instead of Watch()
func (c *MockClient) WatchMock(ctx context.Context) *MockChangeStream {
	log.Debugf("creating mock change stream for client (all collections)")

	// No filter - watch all collections
	filter := ChangeStreamFilter{}

	// Create the change stream
	stream := NewMockChangeStream(ctx, filter, 100)

	// Register with storage
	c.storage.registerChangeStream(stream)

	return stream
}

// getStorage returns the internal MemoryStorage
// This is a helper method for testing purposes
func (c *MockClient) getStorage() *MemoryStorage {
	return c.storage
}

// Clear removes all data from the client's storage
// Useful for test cleanup and resetting state between tests
func (c *MockClient) Clear() {
	log.Debugf("MockClient.Clear() called - clearing all data")
	c.storage.Clear()
}

// applyFilter applies a BSON filter to check if it matches
// This is a helper method for ListDatabaseNames filtering
// Currently not used but prepared for future enhancements
func (c *MockClient) applyFilter(name string, filter interface{}) bool {
	// If no filter, match everything
	if filter == nil {
		return true
	}

	// Try to parse filter as bson.M
	if filterMap, ok := filter.(bson.M); ok {
		// Check if name filter exists
		if nameFilter, exists := filterMap["name"]; exists {
			if nameStr, ok := nameFilter.(string); ok {
				return name == nameStr
			}
		}
		// If no name filter or filter is empty, match everything
		return len(filterMap) == 0
	}

	// Unknown filter type, match everything to be safe
	return true
}

// init 在包初始化时自动注册 Mock client 工厂函数
func init() {
	mongo.RegisterMockClientFactory(NewMockClient)
}

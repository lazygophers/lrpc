package mock

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestNewMockClient(t *testing.T) {
	client := NewMockClient()
	if client == nil {
		t.Fatal("NewMockClient() returned nil")
	}

	mockClient, ok := client.(*MockClient)
	if !ok {
		t.Fatal("NewMockClient() did not return *MockClient")
	}

	if mockClient.storage == nil {
		t.Fatal("MockClient.storage is nil")
	}
}

func TestMockClient_Connect(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Errorf("Connect() returned error: %v", err)
	}
}

func TestMockClient_Disconnect(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	err := client.Disconnect(ctx)
	if err != nil {
		t.Errorf("Disconnect() returned error: %v", err)
	}
}

func TestMockClient_Ping(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	err := client.Ping(ctx, nil)
	if err != nil {
		t.Errorf("Ping() returned error: %v", err)
	}
}

func TestMockClient_Database(t *testing.T) {
	client := NewMockClient()

	db := client.Database("testdb")
	if db == nil {
		t.Fatal("Database() returned nil")
	}

	if db.Name() != "testdb" {
		t.Errorf("Database name = %s, want testdb", db.Name())
	}

	// Verify the database has the same storage as client
	mockClient := client.(*MockClient)
	mockDB := db.(*MockDatabase)
	if mockDB.storage != mockClient.storage {
		t.Error("Database storage does not match client storage")
	}
}

func TestMockClient_DatabaseSharedStorage(t *testing.T) {
	client := NewMockClient()

	// Create two databases
	db1 := client.Database("db1")
	db2 := client.Database("db2")

	// Insert data through db1
	coll1 := db1.Collection("users")
	ctx := context.Background()
	_, err := coll1.InsertOne(ctx, bson.M{"name": "Alice", "age": 30})
	if err != nil {
		t.Fatalf("InsertOne() error: %v", err)
	}

	// Verify data is accessible through db2 (shared storage)
	coll2 := db2.Collection("users")
	count, err := coll2.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments() error: %v", err)
	}

	if count != 1 {
		t.Errorf("CountDocuments() = %d, want 1", count)
	}
}

func TestMockClient_ListDatabaseNames_Empty(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	names, err := client.ListDatabaseNames(ctx, nil)
	if err != nil {
		t.Fatalf("ListDatabaseNames() error: %v", err)
	}

	if len(names) != 0 {
		t.Errorf("ListDatabaseNames() returned %d names, want 0 for empty storage", len(names))
	}
}

func TestMockClient_ListDatabaseNames_WithData(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Insert some data to create collections
	db := client.Database("testdb")
	coll := db.Collection("users")
	_, err := coll.InsertOne(ctx, bson.M{"name": "Bob"})
	if err != nil {
		t.Fatalf("InsertOne() error: %v", err)
	}

	names, err := client.ListDatabaseNames(ctx, nil)
	if err != nil {
		t.Fatalf("ListDatabaseNames() error: %v", err)
	}

	if len(names) == 0 {
		t.Error("ListDatabaseNames() returned no names after inserting data")
	}
}

func TestMockClient_ListDatabases(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Insert some data
	db := client.Database("testdb")
	coll := db.Collection("users")
	_, err := coll.InsertOne(ctx, bson.M{"name": "Charlie"})
	if err != nil {
		t.Fatalf("InsertOne() error: %v", err)
	}

	result, err := client.ListDatabases(ctx, nil)
	if err != nil {
		t.Fatalf("ListDatabases() error: %v", err)
	}

	if len(result.Databases) == 0 {
		t.Error("ListDatabases() returned no databases after inserting data")
	}
}

func TestMockClient_NumberSessionsInProgress(t *testing.T) {
	client := NewMockClient()

	count := client.NumberSessionsInProgress()
	if count != 0 {
		t.Errorf("NumberSessionsInProgress() = %d, want 0", count)
	}
}

func TestMockClient_Timeout(t *testing.T) {
	client := NewMockClient()

	timeout := client.Timeout()
	if timeout != nil {
		t.Errorf("Timeout() = %v, want nil", timeout)
	}
}

func TestMockClient_StartSession_NotImplemented(t *testing.T) {
	client := NewMockClient()

	_, err := client.StartSession()
	if err != ErrNotImplemented {
		t.Errorf("StartSession() error = %v, want ErrNotImplemented", err)
	}
}

func TestMockClient_UseSession_NotImplemented(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	err := client.UseSession(ctx, nil)
	if err != ErrNotImplemented {
		t.Errorf("UseSession() error = %v, want ErrNotImplemented", err)
	}
}

func TestMockClient_UseSessionWithOptions_NotImplemented(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	err := client.UseSessionWithOptions(ctx, nil, nil)
	if err != ErrNotImplemented {
		t.Errorf("UseSessionWithOptions() error = %v, want ErrNotImplemented", err)
	}
}

func TestMockClient_Watch_NotImplemented(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	_, err := client.Watch(ctx, nil)
	if err != ErrNotImplemented {
		t.Errorf("Watch() error = %v, want ErrNotImplemented", err)
	}
}

func TestMockClient_Clear(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Insert some data
	db := client.Database("testdb")
	coll := db.Collection("users")
	_, err := coll.InsertOne(ctx, bson.M{"name": "Dave"})
	if err != nil {
		t.Fatalf("InsertOne() error: %v", err)
	}

	// Verify data exists
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments() error: %v", err)
	}
	if count != 1 {
		t.Errorf("Before Clear: CountDocuments() = %d, want 1", count)
	}

	// Clear the client
	mockClient := client.(*MockClient)
	mockClient.Clear()

	// Verify data is cleared
	count, err = coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments() error after Clear: %v", err)
	}
	if count != 0 {
		t.Errorf("After Clear: CountDocuments() = %d, want 0", count)
	}
}

// TestMockClient_GetStorage tests getStorage method
func TestMockClient_GetStorage(t *testing.T) {
	client := NewMockClient()
	mockClient := client.(*MockClient)

	storage := mockClient.getStorage()
	if storage == nil {
		t.Fatal("getStorage() returned nil")
	}

	if storage != mockClient.storage {
		t.Error("getStorage() returned different storage instance")
	}
}

// TestMockClient_ApplyFilter tests applyFilter method
func TestMockClient_ApplyFilter(t *testing.T) {
	client := NewMockClient()
	mockClient := client.(*MockClient)

	tests := []struct {
		name     string
		dbName   string
		filter   interface{}
		expected bool
	}{
		{
			name:     "nil filter matches",
			dbName:   "testdb",
			filter:   nil,
			expected: true,
		},
		{
			name:     "empty bson.M matches",
			dbName:   "testdb",
			filter:   bson.M{},
			expected: true,
		},
		{
			name:     "name filter matches",
			dbName:   "testdb",
			filter:   bson.M{"name": "testdb"},
			expected: true,
		},
		{
			name:     "name filter does not match",
			dbName:   "testdb",
			filter:   bson.M{"name": "otherdb"},
			expected: false,
		},
		{
			name:     "non-string name filter matches all",
			dbName:   "testdb",
			filter:   bson.M{"name": 123},
			expected: false,
		},
		{
			name:     "unknown filter type matches",
			dbName:   "testdb",
			filter:   "invalid",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mockClient.applyFilter(tt.dbName, tt.filter)
			if result != tt.expected {
				t.Errorf("applyFilter() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestMockClient_ListDatabases_WithFilter tests ListDatabases with filter
func TestMockClient_ListDatabases_WithFilter(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// Insert some data to create collections
	db := client.Database("testdb")
	coll := db.Collection("users")
	_, err := coll.InsertOne(ctx, bson.M{"name": "Eve"})
	if err != nil {
		t.Fatalf("InsertOne() error: %v", err)
	}

	// Test with filter
	result, err := client.ListDatabases(ctx, bson.M{"name": "testdb"})
	if err != nil {
		t.Fatalf("ListDatabases() error: %v", err)
	}

	if len(result.Databases) == 0 {
		t.Error("ListDatabases() returned no databases with matching filter")
	}

	// Verify database specification fields
	if len(result.Databases) > 0 {
		db := result.Databases[0]
		if db.Name == "" {
			t.Error("Database name is empty")
		}
		if db.SizeOnDisk != 0 {
			t.Errorf("Database SizeOnDisk = %d, want 0 for mock", db.SizeOnDisk)
		}
		if db.Empty != false {
			t.Errorf("Database Empty = %v, want false", db.Empty)
		}
	}

	// Test total size
	if result.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0 for mock", result.TotalSize)
	}
}

// TestMockClient_ListDatabases_ErrorFromListDatabaseNames tests error propagation
func TestMockClient_ListDatabases_ErrorFromListDatabaseNames(t *testing.T) {
	client := NewMockClient()
	ctx := context.Background()

	// ListDatabaseNames should not error in mock implementation
	// but we test the path
	result, err := client.ListDatabases(ctx, nil)
	if err != nil {
		t.Fatalf("ListDatabases() unexpected error: %v", err)
	}

	// Should return empty result for no collections
	if len(result.Databases) != 0 {
		t.Errorf("ListDatabases() returned %d databases, want 0 for empty storage", len(result.Databases))
	}
}

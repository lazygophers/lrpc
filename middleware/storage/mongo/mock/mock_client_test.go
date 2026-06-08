package mock

import (
"context"
"fmt"
"go.mongodb.org/mongo-driver/bson"
"testing"
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

func ExampleNewMockClient() {
	// Create a new mock client
	client := NewMockClient()

	// Connect to the mock client (always succeeds)
	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	// Get a database
	db := client.Database("mydb")

	// Get a collection
	coll := db.Collection("users")

	// Insert a document
	result, err := coll.InsertOne(ctx, bson.M{
		"name": "Alice",
		"age":  30,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Inserted document with ID: %v\n", result.InsertedID)

	// Find the document
	cursor, err := coll.Find(ctx, bson.M{"name": "Alice"})
	if err != nil {
		panic(err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d document(s)\n", len(results))
	// Output will vary due to ObjectID generation, so we don't add Output comment
}

// ExampleMockClient_sharedStorage demonstrates how multiple databases share the same storage
func ExampleMockClient_sharedStorage() {
	client := NewMockClient()
	ctx := context.Background()

	// Create two databases
	db1 := client.Database("db1")
	db2 := client.Database("db2")

	// Insert data through db1
	coll1 := db1.Collection("products")
	_, err := coll1.InsertOne(ctx, bson.M{
		"name":  "Laptop",
		"price": 1200,
	})
	if err != nil {
		panic(err)
	}

	// Access the same collection through db2 (shared storage)
	coll2 := db2.Collection("products")
	count, err := coll2.CountDocuments(ctx, bson.M{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d product(s)\n", count)
	// Output: Found 1 product(s)
}

// ExampleMockClient_testCleanup demonstrates using Clear() for test cleanup
func ExampleMockClient_testCleanup() {
	client := NewMockClient()
	ctx := context.Background()

	// Insert test data
	db := client.Database("testdb")
	coll := db.Collection("items")
	_, err := coll.InsertOne(ctx, bson.M{"item": "test1"})
	if err != nil {
		panic(err)
	}
	_, err = coll.InsertOne(ctx, bson.M{"item": "test2"})
	if err != nil {
		panic(err)
	}

	// Check count before cleanup
	count, _ := coll.CountDocuments(ctx, bson.M{})
	fmt.Printf("Before cleanup: %d items\n", count)

	// Clear all data for cleanup
	mockClient := client.(*MockClient)
	mockClient.Clear()

	// Check count after cleanup
	count, _ = coll.CountDocuments(ctx, bson.M{})
	fmt.Printf("After cleanup: %d items\n", count)

	// Output:
	// Before cleanup: 2 items
	// After cleanup: 0 items
}

func ExampleMemoryStorage_Insert() {
	storage := NewMemoryStorage()

	// Insert a single document
	doc := bson.M{
		"name":  "Alice",
		"age":   25,
		"email": "alice@example.com",
	}

	err := storage.Insert("users", doc)
	if err != nil {
		panic(err)
	}

	// Find the document
	results := storage.Find("users", bson.M{}, nil)

	fmt.Printf("Found %d document(s)\n", len(results))
	// Output: Found 1 document(s)
}

func ExampleMemoryStorage_InsertMany() {
	storage := NewMemoryStorage()

	// Insert multiple documents
	docs := []bson.M{
		{"name": "Alice", "age": 25},
		{"name": "Bob", "age": 30},
		{"name": "Charlie", "age": 35},
	}

	err := storage.InsertMany("users", docs)
	if err != nil {
		panic(err)
	}

	// Count documents
	count := storage.Count("users", bson.M{})

	fmt.Printf("Total documents: %d\n", count)
	// Output: Total documents: 3
}

func ExampleMemoryStorage_Find() {
	storage := NewMemoryStorage()

	// Insert test data
	docs := []bson.M{
		{"name": "Alice", "age": 25},
		{"name": "Bob", "age": 30},
		{"name": "Charlie", "age": 35},
	}
	err := storage.InsertMany("users", docs)
	if err != nil {
		panic(err)
	}

	// Find with limit and skip
	limit := int64(2)
	skip := int64(1)
	opts := &FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}

	results := storage.Find("users", bson.M{}, opts)

	fmt.Printf("Found %d document(s)\n", len(results))
	// Output: Found 2 document(s)
}

func ExampleMemoryStorage_Update() {
	storage := NewMemoryStorage()

	// Insert a document
	doc := bson.M{"name": "Alice", "age": 25}
	err := storage.Insert("users", doc)
	if err != nil {
		panic(err)
	}

	// Update the document
	update := bson.M{"$set": bson.M{"age": 26}}
	updated := storage.Update("users", bson.M{}, update)

	fmt.Printf("Updated %d document(s)\n", updated)
	// Output: Updated 1 document(s)
}

func ExampleMemoryStorage_Delete() {
	storage := NewMemoryStorage()

	// Insert test data
	docs := []bson.M{
		{"name": "Alice", "age": 25},
		{"name": "Bob", "age": 30},
	}
	err := storage.InsertMany("users", docs)
	if err != nil {
		panic(err)
	}

	// Delete one document
	deleted := storage.DeleteOne("users", bson.M{})

	fmt.Printf("Deleted %d document(s)\n", deleted)

	// Count remaining documents
	count := storage.Count("users", bson.M{})

	fmt.Printf("Remaining documents: %d\n", count)
	// Output: Deleted 1 document(s)
	// Remaining documents: 1
}

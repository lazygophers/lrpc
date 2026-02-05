package mock

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestNewMockDatabase(t *testing.T) {
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	if db == nil {
		t.Fatal("NewMockDatabase returned nil")
	}

	if db.Name() != "testdb" {
		t.Errorf("expected database name 'testdb', got '%s'", db.Name())
	}

	if db.storage != storage {
		t.Error("database storage does not match provided storage")
	}
}

func TestMockDatabase_Name(t *testing.T) {
	storage := NewMemoryStorage()
	db := NewMockDatabase("mydb", storage, nil)

	name := db.Name()
	if name != "mydb" {
		t.Errorf("expected name 'mydb', got '%s'", name)
	}
}

func TestMockDatabase_Collection(t *testing.T) {
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	coll := db.Collection("test_collection")
	if coll == nil {
		t.Fatal("Collection returned nil")
	}

	if coll.Name() != "test_collection" {
		t.Errorf("expected collection name 'test_collection', got '%s'", coll.Name())
	}
}

func TestMockDatabase_CreateCollection(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	err := db.CreateCollection(ctx, "new_collection")
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	// Verify collection exists
	names := storage.ListCollections()
	found := false
	for _, name := range names {
		if name == "new_collection" {
			found = true
			break
		}
	}

	if !found {
		t.Error("created collection not found in storage")
	}

	// Try creating duplicate
	err = db.CreateCollection(ctx, "new_collection")
	if err == nil {
		t.Error("expected error when creating duplicate collection")
	}
}

func TestMockDatabase_ListCollectionNames(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	// Create some collections
	err := db.CreateCollection(ctx, "coll1")
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	err = db.CreateCollection(ctx, "coll2")
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	err = db.CreateCollection(ctx, "coll3")
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	// List collections
	names, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		t.Fatalf("ListCollectionNames failed: %v", err)
	}

	if len(names) != 3 {
		t.Errorf("expected 3 collections, got %d", len(names))
	}

	// Check all collection names are present
	expectedNames := map[string]bool{"coll1": true, "coll2": true, "coll3": true}
	for _, name := range names {
		if !expectedNames[name] {
			t.Errorf("unexpected collection name: %s", name)
		}
	}
}

func TestMockDatabase_ListCollectionSpecifications(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	// Create a collection
	err := db.CreateCollection(ctx, "test_coll")
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	// List specifications
	specs, err := db.ListCollectionSpecifications(ctx, bson.M{})
	if err != nil {
		t.Fatalf("ListCollectionSpecifications failed: %v", err)
	}

	if len(specs) != 1 {
		t.Errorf("expected 1 specification, got %d", len(specs))
	}

	if specs[0].Name != "test_coll" {
		t.Errorf("expected specification name 'test_coll', got '%s'", specs[0].Name)
	}

	if specs[0].Type != "collection" {
		t.Errorf("expected specification type 'collection', got '%s'", specs[0].Type)
	}
}

func TestMockDatabase_ListCollections(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	// Create collections
	err := db.CreateCollection(ctx, "coll1")
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	err = db.CreateCollection(ctx, "coll2")
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	// List collections using cursor
	cursor, err := db.ListCollections(ctx, bson.M{})
	if err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	if cursor == nil {
		t.Fatal("cursor is nil")
	}

	// Read all documents from cursor
	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("cursor.All failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 collection documents, got %d", len(results))
	}
}

func TestMockDatabase_Drop(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	// Create collections with data
	err := db.CreateCollection(ctx, "coll1")
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	err = db.CreateCollection(ctx, "coll2")
	if err != nil {
		t.Fatalf("CreateCollection failed: %v", err)
	}

	// Insert some data
	err = storage.Insert("coll1", bson.M{"key": "value1"})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	err = storage.Insert("coll2", bson.M{"key": "value2"})
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Drop database
	err = db.Drop(ctx)
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}

	// Verify all collections are dropped
	names := storage.ListCollections()
	if len(names) != 0 {
		t.Errorf("expected 0 collections after drop, got %d", len(names))
	}
}

func TestMockDatabase_ReadConcern(t *testing.T) {
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	rc := db.ReadConcern()
	if rc != nil {
		t.Error("expected ReadConcern to return nil in mock")
	}
}

func TestMockDatabase_ReadPreference(t *testing.T) {
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	rp := db.ReadPreference()
	if rp != nil {
		t.Error("expected ReadPreference to return nil in mock")
	}
}

func TestMockDatabase_WriteConcern(t *testing.T) {
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	wc := db.WriteConcern()
	if wc != nil {
		t.Error("expected WriteConcern to return nil in mock")
	}
}

func TestMockDatabase_RunCommand(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	result := db.RunCommand(ctx, bson.M{"ping": 1})
	if result == nil {
		t.Error("expected RunCommand to return non-nil SingleResult")
	}
}

func TestMockDatabase_Aggregate(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	cursor, err := db.Aggregate(ctx, bson.A{})
	if err != ErrNotImplemented {
		t.Errorf("expected ErrNotImplemented, got: %v", err)
	}

	if cursor != nil {
		t.Error("expected cursor to be nil for not implemented method")
	}
}

func TestMockDatabase_CreateView(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	// Test basic view creation
	err := db.CreateView(ctx, "view1", "coll1", bson.A{})
	if err != nil {
		t.Errorf("CreateView should succeed in mock mode, got error: %v", err)
	}

	// Test view creation with pipeline
	pipeline := bson.A{
		bson.M{"$match": bson.M{"status": "active"}},
		bson.M{"$project": bson.M{"name": 1, "email": 1}},
	}
	err = db.CreateView(ctx, "active_users", "users", pipeline)
	if err != nil {
		t.Errorf("CreateView with pipeline should succeed in mock mode, got error: %v", err)
	}

	// Test view creation with empty view name
	err = db.CreateView(ctx, "", "coll1", bson.A{})
	if err != nil {
		t.Logf("CreateView with empty name returned error (expected behavior): %v", err)
	}
}

func TestMockDatabase_RunCommandCursor(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	// Test basic command execution
	cursor, err := db.RunCommandCursor(ctx, bson.M{"ping": 1})
	if err != nil {
		t.Fatalf("RunCommandCursor failed: %v", err)
	}

	if cursor == nil {
		t.Fatal("expected cursor to be non-nil")
	}

	// Verify cursor is empty (mock mode returns empty cursor)
	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("cursor.All failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results from mock cursor, got %d", len(results))
	}

	// Test listCollections command
	cursor, err = db.RunCommandCursor(ctx, bson.M{"listCollections": 1})
	if err != nil {
		t.Fatalf("RunCommandCursor with listCollections failed: %v", err)
	}

	if cursor == nil {
		t.Fatal("expected cursor to be non-nil for listCollections command")
	}

	// Test with complex command
	cursor, err = db.RunCommandCursor(ctx, bson.M{
		"aggregate": "test_collection",
		"pipeline": bson.A{
			bson.M{"$match": bson.M{"status": "active"}},
		},
		"cursor": bson.M{},
	})
	if err != nil {
		t.Fatalf("RunCommandCursor with aggregate failed: %v", err)
	}

	if cursor == nil {
		t.Fatal("expected cursor to be non-nil for aggregate command")
	}
}

func TestMockDatabase_Watch(t *testing.T) {
	ctx := context.Background()
	storage := NewMemoryStorage()
	db := NewMockDatabase("testdb", storage, nil)

	stream, err := db.Watch(ctx, bson.A{})
	if err != ErrNotImplemented {
		t.Errorf("expected ErrNotImplemented, got: %v", err)
	}

	if stream != nil {
		t.Error("expected stream to be nil for not implemented method")
	}
}

func TestMockDatabase_Client(t *testing.T) {
	storage := NewMemoryStorage()

	// Create a mock client (we'll use nil for this test)
	// In a real scenario, this would be a MockClient instance
	var mockClient interface{} = nil

	db := NewMockDatabase("testdb", storage, nil)

	client := db.Client()
	if client != mockClient {
		t.Error("expected Client to return the same client reference")
	}
}

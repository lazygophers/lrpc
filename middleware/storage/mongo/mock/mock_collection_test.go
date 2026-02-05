package mock_test

import (
	"context"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMockCollection_InsertOne(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()
	doc := bson.M{"name": "Alice", "age": 25}

	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	if result.InsertedID == nil {
		t.Error("InsertedID should not be nil")
	}
}

func TestMockCollection_InsertMany(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	}

	result, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	if len(result.InsertedIDs) != 3 {
		t.Errorf("Expected 3 inserted IDs, got %d", len(result.InsertedIDs))
	}
}

func TestMockCollection_Find(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Find all documents
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("cursor.All failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(results))
	}
}

func TestMockCollection_FindOne(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Find one document
	result := coll.FindOne(ctx, bson.M{})

	var found bson.M
	err = result.Decode(&found)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if found["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got '%v'", found["name"])
	}
}

func TestMockCollection_FindOne_NotFound(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Find in empty collection
	result := coll.FindOne(ctx, bson.M{})

	var found bson.M
	err := result.Decode(&found)
	if err != mongo.ErrNoDocuments {
		t.Errorf("Expected ErrNoDocuments, got %v", err)
	}
}

func TestMockCollection_UpdateOne(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Update document
	update := bson.M{"$set": bson.M{"age": 26}}
	result, err := coll.UpdateOne(ctx, bson.M{}, update)
	if err != nil {
		t.Fatalf("UpdateOne failed: %v", err)
	}

	if result.MatchedCount != 1 {
		t.Errorf("Expected MatchedCount 1, got %d", result.MatchedCount)
	}
	if result.ModifiedCount != 1 {
		t.Errorf("Expected ModifiedCount 1, got %d", result.ModifiedCount)
	}

	// Verify update
	findResult := coll.FindOne(ctx, bson.M{})
	var updated bson.M
	err = findResult.Decode(&updated)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	age, ok := updated["age"].(int32)
	if !ok || age != 26 {
		t.Errorf("Expected age 26, got %v", updated["age"])
	}
}

func TestMockCollection_UpdateOne_Upsert(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Update non-existent document with upsert
	update := bson.M{"$set": bson.M{"name": "Alice", "age": 25}}
	upsert := true
	opts := options.Update().SetUpsert(upsert)

	result, err := coll.UpdateOne(ctx, bson.M{}, update, opts)
	if err != nil {
		t.Fatalf("UpdateOne failed: %v", err)
	}

	if result.UpsertedCount != 1 {
		t.Errorf("Expected UpsertedCount 1, got %d", result.UpsertedCount)
	}
	if result.UpsertedID == nil {
		t.Error("UpsertedID should not be nil")
	}
}

func TestMockCollection_UpdateByID(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	insertResult, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Update by ID
	update := bson.M{"$set": bson.M{"age": 26}}
	result, err := coll.UpdateByID(ctx, insertResult.InsertedID, update)
	if err != nil {
		t.Fatalf("UpdateByID failed: %v", err)
	}

	if result.MatchedCount != 1 {
		t.Errorf("Expected MatchedCount 1, got %d", result.MatchedCount)
	}
}

func TestMockCollection_DeleteOne(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Delete one document
	result, err := coll.DeleteOne(ctx, bson.M{})
	if err != nil {
		t.Fatalf("DeleteOne failed: %v", err)
	}

	if result.DeletedCount != 1 {
		t.Errorf("Expected DeletedCount 1, got %d", result.DeletedCount)
	}

	// Verify count
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 document remaining, got %d", count)
	}
}

func TestMockCollection_DeleteMany(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Delete all documents
	result, err := coll.DeleteMany(ctx, bson.M{})
	if err != nil {
		t.Fatalf("DeleteMany failed: %v", err)
	}

	if result.DeletedCount != 3 {
		t.Errorf("Expected DeletedCount 3, got %d", result.DeletedCount)
	}

	// Verify count
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 documents remaining, got %d", count)
	}
}

func TestMockCollection_CountDocuments(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Count all documents
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestMockCollection_EstimatedDocumentCount(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Estimate count
	count, err := coll.EstimatedDocumentCount(ctx)
	if err != nil {
		t.Fatalf("EstimatedDocumentCount failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestMockCollection_ReplaceOne(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"_id": primitive.NewObjectID(), "name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Replace document
	replacement := bson.M{"name": "Alice Smith", "email": "alice@example.com"}
	result, err := coll.ReplaceOne(ctx, bson.M{"name": "Alice"}, replacement)
	if err != nil {
		t.Fatalf("ReplaceOne failed: %v", err)
	}

	if result.MatchedCount != 1 {
		t.Errorf("Expected MatchedCount 1, got %d", result.MatchedCount)
	}
	if result.ModifiedCount != 1 {
		t.Errorf("Expected ModifiedCount 1, got %d", result.ModifiedCount)
	}

	// Verify replacement
	findResult := coll.FindOne(ctx, bson.M{})
	var replaced bson.M
	err = findResult.Decode(&replaced)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if replaced["name"] != "Alice Smith" {
		t.Errorf("Expected name 'Alice Smith', got '%v'", replaced["name"])
	}
	if replaced["email"] != "alice@example.com" {
		t.Errorf("Expected email 'alice@example.com', got '%v'", replaced["email"])
	}
	// age should be removed
	if _, hasAge := replaced["age"]; hasAge {
		t.Error("Expected age field to be removed")
	}
}

func TestMockCollection_FindOneAndUpdate(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Find and update (return old document)
	update := bson.M{"$set": bson.M{"age": 26}}
	result := coll.FindOneAndUpdate(ctx, bson.M{}, update)

	var found bson.M
	err = result.Decode(&found)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Should return old document
	age, ok := found["age"].(int32)
	if !ok || age != 25 {
		t.Errorf("Expected age 25 (old value), got %v", found["age"])
	}
}

func TestMockCollection_FindOneAndDelete(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Find and delete
	result := coll.FindOneAndDelete(ctx, bson.M{})

	var found bson.M
	err = result.Decode(&found)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if found["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got '%v'", found["name"])
	}

	// Verify deletion
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 documents after delete, got %d", count)
	}
}

func TestMockCollection_Name(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	if coll.Name() != "users" {
		t.Errorf("Expected collection name 'users', got '%s'", coll.Name())
	}
}

func TestMockCollection_Drop(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Drop collection
	err = coll.Drop(ctx)
	if err != nil {
		t.Fatalf("Drop failed: %v", err)
	}

	// Verify drop
	collections := storage.ListCollections()
	for _, name := range collections {
		if name == "users" {
			t.Error("Collection 'users' should be dropped")
		}
	}
}

func TestMockCollection_NotImplementedMethods(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Test Aggregate - now implemented, should not return error for empty pipeline
	cursor, err := coll.Aggregate(ctx, bson.A{})
	if err != nil {
		t.Errorf("Aggregate should work with empty pipeline, got error: %v", err)
	}
	if cursor == nil {
		t.Errorf("Aggregate should return non-nil cursor")
	}

	// Test Watch - still not implemented
	_, err = coll.Watch(ctx, bson.A{})
	if err != mock.ErrNotImplemented {
		t.Errorf("Watch should return ErrNotImplemented, got %v", err)
	}
}

// invalidBsonType is a type that cannot be marshaled to BSON
type invalidBsonType struct {
	Channel chan int
}

// TestMockCollection_CountDocuments_InvalidFilter tests CountDocuments with invalid filter
func TestMockCollection_CountDocuments_InvalidFilter(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Try to count with invalid filter (channel cannot be marshaled to BSON)
	invalidFilter := invalidBsonType{Channel: make(chan int)}
	_, err := coll.CountDocuments(ctx, invalidFilter)
	if err == nil {
		t.Error("Expected error for invalid filter, got nil")
	}
}

// TestMockCollection_CountDocuments_EmptyCollection tests CountDocuments on empty collection
func TestMockCollection_CountDocuments_EmptyCollection(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Count documents in empty collection
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0 for empty collection, got %d", count)
	}
}

// TestMockCollection_CountDocuments_WithOptions tests CountDocuments with options
func TestMockCollection_CountDocuments_WithOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Count with filter
	count, err := coll.CountDocuments(ctx, bson.M{"age": bson.M{"$gt": 25}}, options.Count())
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

// TestMockCollection_Database tests Database method
func TestMockCollection_Database(t *testing.T) {
	storage := mock.NewMemoryStorage()
	mockDB := mock.NewMockDatabase("testdb", storage, nil)
	coll := mock.NewMockCollection("users", storage, mockDB)

	db := coll.Database()
	if db == nil {
		t.Error("Database should not be nil")
	}

	if db != mockDB {
		t.Error("Database should return the same database instance")
	}
}

// TestMockCollection_DeleteMany_InvalidFilter tests DeleteMany with invalid filter
func TestMockCollection_DeleteMany_InvalidFilter(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Try to delete with invalid filter
	invalidFilter := invalidBsonType{Channel: make(chan int)}
	_, err := coll.DeleteMany(ctx, invalidFilter)
	if err == nil {
		t.Error("Expected error for invalid filter, got nil")
	}
}

// TestMockCollection_DeleteMany_WithOptions tests DeleteMany with options
func TestMockCollection_DeleteMany_WithOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Delete with filter and options
	result, err := coll.DeleteMany(ctx, bson.M{"age": bson.M{"$gte": 30}}, options.Delete())
	if err != nil {
		t.Fatalf("DeleteMany failed: %v", err)
	}

	if result.DeletedCount != 2 {
		t.Errorf("Expected DeletedCount 2, got %d", result.DeletedCount)
	}
}

// TestMockCollection_DeleteOne_InvalidFilter tests DeleteOne with invalid filter
func TestMockCollection_DeleteOne_InvalidFilter(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Try to delete with invalid filter
	invalidFilter := invalidBsonType{Channel: make(chan int)}
	_, err := coll.DeleteOne(ctx, invalidFilter)
	if err == nil {
		t.Error("Expected error for invalid filter, got nil")
	}
}

// TestMockCollection_DeleteOne_WithOptions tests DeleteOne with options
func TestMockCollection_DeleteOne_WithOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Delete with options
	result, err := coll.DeleteOne(ctx, bson.M{"age": 30}, options.Delete())
	if err != nil {
		t.Fatalf("DeleteOne failed: %v", err)
	}

	if result.DeletedCount != 1 {
		t.Errorf("Expected DeletedCount 1, got %d", result.DeletedCount)
	}
}

// TestMockCollection_Drop_NonExistentCollection tests Drop on non-existent collection
func TestMockCollection_Drop_NonExistentCollection(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("nonexistent", storage, nil)

	ctx := context.Background()

	// Drop non-existent collection should fail
	err := coll.Drop(ctx)
	if err == nil {
		t.Error("Expected error when dropping non-existent collection, got nil")
	}
}

// TestMockCollection_EstimatedDocumentCount_EmptyCollection tests EstimatedDocumentCount on empty collection
func TestMockCollection_EstimatedDocumentCount_EmptyCollection(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Estimate count on empty collection
	count, err := coll.EstimatedDocumentCount(ctx)
	if err != nil {
		t.Fatalf("EstimatedDocumentCount failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected count 0 for empty collection, got %d", count)
	}
}

// TestMockCollection_EstimatedDocumentCount_WithOptions tests EstimatedDocumentCount with options
func TestMockCollection_EstimatedDocumentCount_WithOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Estimate count with options
	count, err := coll.EstimatedDocumentCount(ctx, options.EstimatedDocumentCount())
	if err != nil {
		t.Fatalf("EstimatedDocumentCount failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

// TestMockCollection_Find_InvalidFilter tests Find with invalid filter
func TestMockCollection_Find_InvalidFilter(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Try to find with invalid filter
	invalidFilter := invalidBsonType{Channel: make(chan int)}
	_, err := coll.Find(ctx, invalidFilter)
	if err == nil {
		t.Error("Expected error for invalid filter, got nil")
	}
}

// TestMockCollection_Find_WithLimit tests Find with limit option
func TestMockCollection_Find_WithLimit(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Find with limit
	opts := options.Find().SetLimit(2)
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("cursor.All failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 documents with limit, got %d", len(results))
	}
}

// TestMockCollection_Find_WithSkip tests Find with skip option
func TestMockCollection_Find_WithSkip(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Find with skip
	opts := options.Find().SetSkip(1)
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("cursor.All failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 documents with skip, got %d", len(results))
	}
}

// TestMockCollection_Find_WithSort tests Find with sort option
func TestMockCollection_Find_WithSort(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Charlie", "age": 35},
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Find with sort
	opts := options.Find().SetSort(bson.M{"age": 1})
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("cursor.All failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 documents, got %d", len(results))
	}

	// Verify sort order
	if results[0]["name"] != "Alice" {
		t.Errorf("Expected first document to be Alice, got %v", results[0]["name"])
	}
}

// TestMockCollection_Find_WithInvalidSort tests Find with invalid sort
func TestMockCollection_Find_WithInvalidSort(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Find with invalid sort (channel cannot be marshaled)
	opts := options.Find().SetSort(invalidBsonType{Channel: make(chan int)})
	_, err = coll.Find(ctx, bson.M{}, opts)
	if err == nil {
		t.Error("Expected error for invalid sort, got nil")
	}
}

// TestMockCollection_Find_WithProjection tests Find with projection option
func TestMockCollection_Find_WithProjection(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25, "email": "alice@example.com"},
		bson.M{"name": "Bob", "age": 30, "email": "bob@example.com"},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Find with projection
	opts := options.Find().SetProjection(bson.M{"name": 1, "age": 1})
	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("cursor.All failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(results))
	}

	// Verify projection (email should not be present)
	if _, hasEmail := results[0]["email"]; hasEmail {
		t.Error("Expected email field to be excluded by projection")
	}
}

// TestMockCollection_Find_WithInvalidProjection tests Find with invalid projection
func TestMockCollection_Find_WithInvalidProjection(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Find with invalid projection (channel cannot be marshaled)
	opts := options.Find().SetProjection(invalidBsonType{Channel: make(chan int)})
	_, err = coll.Find(ctx, bson.M{}, opts)
	if err == nil {
		t.Error("Expected error for invalid projection, got nil")
	}
}

// TestMockCollection_FindOne_InvalidFilter tests FindOne with invalid filter
func TestMockCollection_FindOne_InvalidFilter(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Try to find with invalid filter
	invalidFilter := invalidBsonType{Channel: make(chan int)}
	result := coll.FindOne(ctx, invalidFilter)

	var found bson.M
	err := result.Decode(&found)
	if err == nil {
		t.Error("Expected error for invalid filter, got nil")
	}
}

// TestMockCollection_FindOne_WithProjection tests FindOne with projection option
func TestMockCollection_FindOne_WithProjection(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25, "email": "alice@example.com"}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Find one with projection
	opts := options.FindOne().SetProjection(bson.M{"name": 1, "age": 1})
	result := coll.FindOne(ctx, bson.M{}, opts)

	var found bson.M
	err = result.Decode(&found)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify projection (email should not be present)
	if _, hasEmail := found["email"]; hasEmail {
		t.Error("Expected email field to be excluded by projection")
	}
}

// TestMockCollection_FindOne_WithInvalidProjection tests FindOne with invalid projection
func TestMockCollection_FindOne_WithInvalidProjection(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Find one with invalid projection
	opts := options.FindOne().SetProjection(invalidBsonType{Channel: make(chan int)})
	result := coll.FindOne(ctx, bson.M{}, opts)

	var found bson.M
	err = result.Decode(&found)
	if err == nil {
		t.Error("Expected error for invalid projection, got nil")
	}
}

// TestMockCollection_FindOneAndDelete_InvalidFilter tests FindOneAndDelete with invalid filter
func TestMockCollection_FindOneAndDelete_InvalidFilter(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Try to find and delete with invalid filter
	invalidFilter := invalidBsonType{Channel: make(chan int)}
	result := coll.FindOneAndDelete(ctx, invalidFilter)

	var found bson.M
	err := result.Decode(&found)
	if err == nil {
		t.Error("Expected error for invalid filter, got nil")
	}
}

// TestMockCollection_FindOneAndDelete_NotFound tests FindOneAndDelete when document not found
func TestMockCollection_FindOneAndDelete_NotFound(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Find and delete in empty collection
	result := coll.FindOneAndDelete(ctx, bson.M{})

	var found bson.M
	err := result.Decode(&found)
	if err != mongo.ErrNoDocuments {
		t.Errorf("Expected ErrNoDocuments, got %v", err)
	}
}

// TestMockCollection_InsertMany_EmptyArray tests InsertMany with empty array
func TestMockCollection_InsertMany_EmptyArray(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert empty array
	result, err := coll.InsertMany(ctx, []interface{}{})
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	if len(result.InsertedIDs) != 0 {
		t.Errorf("Expected 0 inserted IDs for empty array, got %d", len(result.InsertedIDs))
	}
}

// TestMockCollection_InsertMany_InvalidDocument tests InsertMany with invalid document
func TestMockCollection_InsertMany_InvalidDocument(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Try to insert invalid document (channel cannot be marshaled)
	docs := []interface{}{
		invalidBsonType{Channel: make(chan int)},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err == nil {
		t.Error("Expected error for invalid document, got nil")
	}
}

// TestMockCollection_InsertMany_WithID tests InsertMany with pre-assigned IDs
func TestMockCollection_InsertMany_WithID(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert documents with pre-assigned IDs
	id1 := primitive.NewObjectID()
	id2 := primitive.NewObjectID()
	docs := []interface{}{
		bson.M{"_id": id1, "name": "Alice", "age": 25},
		bson.M{"_id": id2, "name": "Bob", "age": 30},
	}

	result, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	if len(result.InsertedIDs) != 2 {
		t.Errorf("Expected 2 inserted IDs, got %d", len(result.InsertedIDs))
	}

	// Verify IDs match
	if result.InsertedIDs[0] != id1 {
		t.Errorf("Expected first ID to be %v, got %v", id1, result.InsertedIDs[0])
	}
}

// TestMockCollection_InsertOne_InvalidDocument tests InsertOne with invalid document
func TestMockCollection_InsertOne_InvalidDocument(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Try to insert invalid document (channel cannot be marshaled)
	invalidDoc := invalidBsonType{Channel: make(chan int)}
	_, err := coll.InsertOne(ctx, invalidDoc)
	if err == nil {
		t.Error("Expected error for invalid document, got nil")
	}
}

// TestMockCollection_InsertOne_WithID tests InsertOne with pre-assigned ID
func TestMockCollection_InsertOne_WithID(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert document with pre-assigned ID
	id := primitive.NewObjectID()
	doc := bson.M{"_id": id, "name": "Alice", "age": 25}

	result, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	if result.InsertedID != id {
		t.Errorf("Expected InsertedID to be %v, got %v", id, result.InsertedID)
	}
}

// TestMockCollection_InsertOne_WithOptions tests InsertOne with options
func TestMockCollection_InsertOne_WithOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert with options
	doc := bson.M{"name": "Alice", "age": 25}
	result, err := coll.InsertOne(ctx, doc, options.InsertOne())
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	if result.InsertedID == nil {
		t.Error("InsertedID should not be nil")
	}
}

// TestMockCollection_Find_NilOptions tests Find with nil options
func TestMockCollection_Find_NilOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Find with nil options
	cursor, err := coll.Find(ctx, bson.M{}, nil)
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("cursor.All failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 document, got %d", len(results))
	}
}

// TestMockCollection_FindOne_NilOptions tests FindOne with nil options
func TestMockCollection_FindOne_NilOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Find one with nil options
	result := coll.FindOne(ctx, bson.M{}, nil)

	var found bson.M
	err = result.Decode(&found)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if found["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got '%v'", found["name"])
	}
}

// TestMockCollection_InsertMany_WithOptions tests InsertMany with options
func TestMockCollection_InsertMany_WithOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert with options
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	result, err := coll.InsertMany(ctx, docs, options.InsertMany())
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	if len(result.InsertedIDs) != 2 {
		t.Errorf("Expected 2 inserted IDs, got %d", len(result.InsertedIDs))
	}
}

// TestMockCollection_CountDocuments_NilOptions tests CountDocuments with nil options
func TestMockCollection_CountDocuments_NilOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Count with nil options
	count, err := coll.CountDocuments(ctx, bson.M{}, nil)
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

// TestMockCollection_DeleteMany_NilOptions tests DeleteMany with nil options
func TestMockCollection_DeleteMany_NilOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Delete with nil options
	result, err := coll.DeleteMany(ctx, bson.M{}, nil)
	if err != nil {
		t.Fatalf("DeleteMany failed: %v", err)
	}

	if result.DeletedCount != 2 {
		t.Errorf("Expected DeletedCount 2, got %d", result.DeletedCount)
	}
}

// TestMockCollection_DeleteOne_NilOptions tests DeleteOne with nil options
func TestMockCollection_DeleteOne_NilOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Delete with nil options
	result, err := coll.DeleteOne(ctx, bson.M{}, nil)
	if err != nil {
		t.Fatalf("DeleteOne failed: %v", err)
	}

	if result.DeletedCount != 1 {
		t.Errorf("Expected DeletedCount 1, got %d", result.DeletedCount)
	}
}

// TestMockCollection_EstimatedDocumentCount_NilOptions tests EstimatedDocumentCount with nil options
func TestMockCollection_EstimatedDocumentCount_NilOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Estimate count with nil options
	count, err := coll.EstimatedDocumentCount(ctx, nil)
	if err != nil {
		t.Fatalf("EstimatedDocumentCount failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

// TestMockCollection_FindOneAndDelete_NilOptions tests FindOneAndDelete with nil options
func TestMockCollection_FindOneAndDelete_NilOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Find and delete with nil options
	result := coll.FindOneAndDelete(ctx, bson.M{}, nil)

	var found bson.M
	err = result.Decode(&found)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if found["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got '%v'", found["name"])
	}
}

// TestMockCollection_Database_NilDatabase tests Database method when database is nil
func TestMockCollection_Database_NilDatabase(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	db := coll.Database()
	if db != nil {
		t.Error("Database should be nil when not set")
	}
}

// TestMockCollection_Distinct_SimpleField tests Distinct with simple field name
func TestMockCollection_Distinct_SimpleField(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data with duplicate values
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Alice", "age": 26},
		bson.M{"name": "Charlie", "age": 30},
		bson.M{"name": "Bob", "age": 35},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Get distinct names
	result, err := coll.Distinct(ctx, "name", bson.M{})
	if err != nil {
		t.Fatalf("Distinct failed: %v", err)
	}

	// Should have 3 distinct names: Alice, Bob, Charlie
	if len(result) != 3 {
		t.Errorf("Expected 3 distinct names, got %d", len(result))
	}

	// Verify all distinct values are present
	nameMap := make(map[string]bool)
	for _, v := range result {
		if name, ok := v.(string); ok {
			nameMap[name] = true
		}
	}

	if !nameMap["Alice"] || !nameMap["Bob"] || !nameMap["Charlie"] {
		t.Errorf("Expected Alice, Bob, Charlie in distinct names, got %v", result)
	}
}

// TestMockCollection_Distinct_NestedField tests Distinct with nested field name
func TestMockCollection_Distinct_NestedField(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data with nested fields
	docs := []interface{}{
		bson.M{"name": "Alice", "profile": bson.M{"city": "NYC", "country": "USA"}},
		bson.M{"name": "Bob", "profile": bson.M{"city": "LA", "country": "USA"}},
		bson.M{"name": "Charlie", "profile": bson.M{"city": "NYC", "country": "USA"}},
		bson.M{"name": "David", "profile": bson.M{"city": "London", "country": "UK"}},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Get distinct cities
	result, err := coll.Distinct(ctx, "profile.city", bson.M{})
	if err != nil {
		t.Fatalf("Distinct failed: %v", err)
	}

	// Should have 3 distinct cities: NYC, LA, London
	if len(result) != 3 {
		t.Errorf("Expected 3 distinct cities, got %d", len(result))
	}

	// Verify all distinct values are present
	cityMap := make(map[string]bool)
	for _, v := range result {
		if city, ok := v.(string); ok {
			cityMap[city] = true
		}
	}

	if !cityMap["NYC"] || !cityMap["LA"] || !cityMap["London"] {
		t.Errorf("Expected NYC, LA, London in distinct cities, got %v", result)
	}
}

// TestMockCollection_Distinct_WithFilter tests Distinct with filter
func TestMockCollection_Distinct_WithFilter(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25, "country": "USA"},
		bson.M{"name": "Bob", "age": 30, "country": "USA"},
		bson.M{"name": "Charlie", "age": 35, "country": "UK"},
		bson.M{"name": "David", "age": 40, "country": "USA"},
		bson.M{"name": "Eve", "age": 45, "country": "UK"},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Get distinct names for users from USA
	result, err := coll.Distinct(ctx, "name", bson.M{"country": "USA"})
	if err != nil {
		t.Fatalf("Distinct failed: %v", err)
	}

	// Should have 3 distinct names from USA: Alice, Bob, David
	if len(result) != 3 {
		t.Errorf("Expected 3 distinct names from USA, got %d", len(result))
	}

	// Verify all distinct values are present
	nameMap := make(map[string]bool)
	for _, v := range result {
		if name, ok := v.(string); ok {
			nameMap[name] = true
		}
	}

	if !nameMap["Alice"] || !nameMap["Bob"] || !nameMap["David"] {
		t.Errorf("Expected Alice, Bob, David in distinct names, got %v", result)
	}
}

// TestMockCollection_Distinct_EmptyCollection tests Distinct on empty collection
func TestMockCollection_Distinct_EmptyCollection(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Get distinct values from empty collection
	result, err := coll.Distinct(ctx, "name", bson.M{})
	if err != nil {
		t.Fatalf("Distinct failed: %v", err)
	}

	// Should return empty array
	if len(result) != 0 {
		t.Errorf("Expected 0 distinct values from empty collection, got %d", len(result))
	}
}

// TestMockCollection_Distinct_FieldNotExists tests Distinct when field doesn't exist
func TestMockCollection_Distinct_FieldNotExists(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data without the field we'll query
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Get distinct values for non-existent field
	result, err := coll.Distinct(ctx, "email", bson.M{})
	if err != nil {
		t.Fatalf("Distinct failed: %v", err)
	}

	// Should return empty array since field doesn't exist
	if len(result) != 0 {
		t.Errorf("Expected 0 distinct values for non-existent field, got %d", len(result))
	}
}

// TestMockCollection_Distinct_InvalidFilter tests Distinct with invalid filter
func TestMockCollection_Distinct_InvalidFilter(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Try to get distinct with invalid filter
	invalidFilter := invalidBsonType{Channel: make(chan int)}
	_, err := coll.Distinct(ctx, "name", invalidFilter)
	if err == nil {
		t.Error("Expected error for invalid filter, got nil")
	}
}

// TestMockCollection_Distinct_NumericValues tests Distinct with numeric values
func TestMockCollection_Distinct_NumericValues(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data with duplicate numeric values
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 25},
		bson.M{"name": "David", "age": 30},
		bson.M{"name": "Eve", "age": 35},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Get distinct ages
	result, err := coll.Distinct(ctx, "age", bson.M{})
	if err != nil {
		t.Fatalf("Distinct failed: %v", err)
	}

	// Should have 3 distinct ages: 25, 30, 35
	if len(result) != 3 {
		t.Errorf("Expected 3 distinct ages, got %d", len(result))
	}
}

// TestMockCollection_Distinct_WithOptions tests Distinct with options
func TestMockCollection_Distinct_WithOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	docs := []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Alice", "age": 26},
	}
	_, err := coll.InsertMany(ctx, docs)
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Get distinct with options (options parameter exists but is currently unused in implementation)
	result, err := coll.Distinct(ctx, "name", bson.M{}, options.Distinct())
	if err != nil {
		t.Fatalf("Distinct failed: %v", err)
	}

	// Should have 2 distinct names
	if len(result) != 2 {
		t.Errorf("Expected 2 distinct names, got %d", len(result))
	}
}

// TestMockCollection_Clone tests Clone method
func TestMockCollection_Clone(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data in original collection
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Clone the collection
	clonedColl, err := coll.Clone()
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	// Verify cloned collection is not nil
	if clonedColl == nil {
		t.Fatal("Cloned collection should not be nil")
	}

	// Verify cloned collection has the same name
	if clonedColl.Name() != coll.Name() {
		t.Errorf("Expected cloned collection name '%s', got '%s'", coll.Name(), clonedColl.Name())
	}

	// Verify cloned collection shares the same storage (can access the same data)
	result := clonedColl.FindOne(ctx, bson.M{"name": "Alice"})
	var found bson.M
	err = result.Decode(&found)
	if err != nil {
		t.Fatalf("FindOne on cloned collection failed: %v", err)
	}

	if found["name"] != "Alice" {
		t.Errorf("Expected name 'Alice' from cloned collection, got '%v'", found["name"])
	}
}

// TestMockCollection_Clone_WithOptions tests Clone method with options
func TestMockCollection_Clone_WithOptions(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Clone with options (options are currently unused but should not cause error)
	opts := options.Collection()
	clonedColl, err := coll.Clone(opts)
	if err != nil {
		t.Fatalf("Clone with options failed: %v", err)
	}

	if clonedColl == nil {
		t.Fatal("Cloned collection should not be nil")
	}

	// Verify functionality
	count, err := clonedColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments on cloned collection failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected count 1 from cloned collection, got %d", count)
	}
}

// TestMockCollection_Clone_IndependentOperations tests that cloned collection can operate independently
func TestMockCollection_Clone_IndependentOperations(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert test data in original collection
	doc := bson.M{"name": "Alice", "age": 25}
	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Clone the collection
	clonedColl, err := coll.Clone()
	if err != nil {
		t.Fatalf("Clone failed: %v", err)
	}

	// Insert data using cloned collection
	doc2 := bson.M{"name": "Bob", "age": 30}
	_, err = clonedColl.InsertOne(ctx, doc2)
	if err != nil {
		t.Fatalf("InsertOne on cloned collection failed: %v", err)
	}

	// Verify both collections can see the new data (shared storage)
	count1, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments on original collection failed: %v", err)
	}

	count2, err := clonedColl.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments on cloned collection failed: %v", err)
	}

	if count1 != 2 {
		t.Errorf("Expected count 2 in original collection, got %d", count1)
	}

	if count2 != 2 {
		t.Errorf("Expected count 2 in cloned collection, got %d", count2)
	}
}

// TestMockCollection_Indexes tests Indexes method
func TestMockCollection_Indexes(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	// Call Indexes method
	indexView := coll.Indexes()

	// In mock mode, it returns zero value of IndexView
	// We just verify it doesn't panic and returns a value
	_ = indexView

	// Note: We cannot test much more since IndexView is a struct from mongo driver
	// and we're returning its zero value
}

// TestMockCollection_SearchIndexes tests SearchIndexes method
func TestMockCollection_SearchIndexes(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	// Call SearchIndexes method
	searchIndexView := coll.SearchIndexes()

	// In mock mode, it returns zero value of SearchIndexView
	// We just verify it doesn't panic and returns a value
	_ = searchIndexView

	// Note: We cannot test much more since SearchIndexView is a struct from mongo driver
	// and we're returning its zero value
}

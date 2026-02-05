package mock_test

import (
	"context"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// invalidBsonTypeForBulkWrite is a type that cannot be marshaled to BSON
type invalidBsonTypeForBulkWrite struct {
	Channel chan int
}

// TestMockCollection_BulkWrite_Empty tests BulkWrite with empty models
func TestMockCollection_BulkWrite_Empty(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Execute bulk write with empty models
	result, err := coll.BulkWrite(ctx, []mongo.WriteModel{})
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
	}

	if result.InsertedCount != 0 {
		t.Errorf("Expected InsertedCount 0, got %d", result.InsertedCount)
	}
	if result.MatchedCount != 0 {
		t.Errorf("Expected MatchedCount 0, got %d", result.MatchedCount)
	}
	if result.ModifiedCount != 0 {
		t.Errorf("Expected ModifiedCount 0, got %d", result.ModifiedCount)
	}
	if result.DeletedCount != 0 {
		t.Errorf("Expected DeletedCount 0, got %d", result.DeletedCount)
	}
	if result.UpsertedCount != 0 {
		t.Errorf("Expected UpsertedCount 0, got %d", result.UpsertedCount)
	}
}

// TestMockCollection_BulkWrite_InsertOneModel tests BulkWrite with InsertOneModel
func TestMockCollection_BulkWrite_InsertOneModel(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Prepare models
	models := []mongo.WriteModel{
		&mongo.InsertOneModel{
			Document: bson.M{"name": "Alice", "age": 25},
		},
		&mongo.InsertOneModel{
			Document: bson.M{"name": "Bob", "age": 30},
		},
	}

	// Execute bulk write
	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
	}

	if result.InsertedCount != 2 {
		t.Errorf("Expected InsertedCount 2, got %d", result.InsertedCount)
	}

	// Verify documents were inserted
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 documents, got %d", count)
	}
}

// TestMockCollection_BulkWrite_UpdateOneModel tests BulkWrite with UpdateOneModel
func TestMockCollection_BulkWrite_UpdateOneModel(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert initial data
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	})
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Prepare models
	models := []mongo.WriteModel{
		&mongo.UpdateOneModel{
			Filter: bson.M{"name": "Alice"},
			Update: bson.M{"$set": bson.M{"age": 26}},
		},
	}

	// Execute bulk write
	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
	}

	if result.MatchedCount != 1 {
		t.Errorf("Expected MatchedCount 1, got %d", result.MatchedCount)
	}
	if result.ModifiedCount != 1 {
		t.Errorf("Expected ModifiedCount 1, got %d", result.ModifiedCount)
	}

	// Verify update
	findResult := coll.FindOne(ctx, bson.M{"name": "Alice"})
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

// TestMockCollection_BulkWrite_UpdateOneModel_Upsert tests BulkWrite with UpdateOneModel and upsert
func TestMockCollection_BulkWrite_UpdateOneModel_Upsert(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Prepare models with upsert
	upsert := true
	models := []mongo.WriteModel{
		&mongo.UpdateOneModel{
			Filter: bson.M{"name": "Alice"},
			Update: bson.M{"$set": bson.M{"name": "Alice", "age": 25}},
			Upsert: &upsert,
		},
	}

	// Execute bulk write
	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
	}

	if result.UpsertedCount != 1 {
		t.Errorf("Expected UpsertedCount 1, got %d", result.UpsertedCount)
	}

	if result.UpsertedIDs[0] == nil {
		t.Error("Expected UpsertedIDs[0] to be set")
	}

	// Verify document was inserted
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 document, got %d", count)
	}
}

// TestMockCollection_BulkWrite_UpdateManyModel tests BulkWrite with UpdateManyModel
func TestMockCollection_BulkWrite_UpdateManyModel(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert initial data
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	})
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Prepare models
	models := []mongo.WriteModel{
		&mongo.UpdateManyModel{
			Filter: bson.M{"age": bson.M{"$gte": 30}},
			Update: bson.M{"$set": bson.M{"status": "senior"}},
		},
	}

	// Execute bulk write
	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
	}

	if result.MatchedCount != 2 {
		t.Errorf("Expected MatchedCount 2, got %d", result.MatchedCount)
	}
	if result.ModifiedCount != 2 {
		t.Errorf("Expected ModifiedCount 2, got %d", result.ModifiedCount)
	}

	// Verify updates
	count, err := coll.CountDocuments(ctx, bson.M{"status": "senior"})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 documents with status 'senior', got %d", count)
	}
}

// TestMockCollection_BulkWrite_ReplaceOneModel tests BulkWrite with ReplaceOneModel
func TestMockCollection_BulkWrite_ReplaceOneModel(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert initial data
	_, err := coll.InsertOne(ctx, bson.M{"name": "Alice", "age": 25})
	if err != nil {
		t.Fatalf("InsertOne failed: %v", err)
	}

	// Prepare models
	models := []mongo.WriteModel{
		&mongo.ReplaceOneModel{
			Filter:      bson.M{"name": "Alice"},
			Replacement: bson.M{"name": "Alice Smith", "email": "alice@example.com"},
		},
	}

	// Execute bulk write
	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
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
	// age should be removed
	if _, hasAge := replaced["age"]; hasAge {
		t.Error("Expected age field to be removed")
	}
}

// TestMockCollection_BulkWrite_DeleteOneModel tests BulkWrite with DeleteOneModel
func TestMockCollection_BulkWrite_DeleteOneModel(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert initial data
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	})
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Prepare models
	models := []mongo.WriteModel{
		&mongo.DeleteOneModel{
			Filter: bson.M{"name": "Alice"},
		},
	}

	// Execute bulk write
	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
	}

	if result.DeletedCount != 1 {
		t.Errorf("Expected DeletedCount 1, got %d", result.DeletedCount)
	}

	// Verify deletion
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 document remaining, got %d", count)
	}
}

// TestMockCollection_BulkWrite_DeleteManyModel tests BulkWrite with DeleteManyModel
func TestMockCollection_BulkWrite_DeleteManyModel(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert initial data
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
		bson.M{"name": "Charlie", "age": 35},
	})
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Prepare models
	models := []mongo.WriteModel{
		&mongo.DeleteManyModel{
			Filter: bson.M{"age": bson.M{"$gte": 30}},
		},
	}

	// Execute bulk write
	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
	}

	if result.DeletedCount != 2 {
		t.Errorf("Expected DeletedCount 2, got %d", result.DeletedCount)
	}

	// Verify deletion
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 document remaining, got %d", count)
	}
}

// TestMockCollection_BulkWrite_Mixed tests BulkWrite with mixed model types
func TestMockCollection_BulkWrite_Mixed(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Insert initial data
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice", "age": 25},
		bson.M{"name": "Bob", "age": 30},
	})
	if err != nil {
		t.Fatalf("InsertMany failed: %v", err)
	}

	// Prepare mixed models
	models := []mongo.WriteModel{
		// Insert
		&mongo.InsertOneModel{
			Document: bson.M{"name": "Charlie", "age": 35},
		},
		// Update
		&mongo.UpdateOneModel{
			Filter: bson.M{"name": "Alice"},
			Update: bson.M{"$set": bson.M{"age": 26}},
		},
		// Delete
		&mongo.DeleteOneModel{
			Filter: bson.M{"name": "Bob"},
		},
	}

	// Execute bulk write
	result, err := coll.BulkWrite(ctx, models)
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
	}

	if result.InsertedCount != 1 {
		t.Errorf("Expected InsertedCount 1, got %d", result.InsertedCount)
	}
	if result.MatchedCount != 1 {
		t.Errorf("Expected MatchedCount 1, got %d", result.MatchedCount)
	}
	if result.ModifiedCount != 1 {
		t.Errorf("Expected ModifiedCount 1, got %d", result.ModifiedCount)
	}
	if result.DeletedCount != 1 {
		t.Errorf("Expected DeletedCount 1, got %d", result.DeletedCount)
	}

	// Verify final state
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 documents, got %d", count)
	}

	// Verify Alice's age was updated
	findResult := coll.FindOne(ctx, bson.M{"name": "Alice"})
	var alice bson.M
	err = findResult.Decode(&alice)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	age, ok := alice["age"].(int32)
	if !ok || age != 26 {
		t.Errorf("Expected Alice's age to be 26, got %v", alice["age"])
	}

	// Verify Bob was deleted
	bobResult := coll.FindOne(ctx, bson.M{"name": "Bob"})
	var bob bson.M
	err = bobResult.Decode(&bob)
	if err != mongo.ErrNoDocuments {
		t.Error("Expected Bob to be deleted")
	}
}

// TestMockCollection_BulkWrite_Ordered tests BulkWrite with ordered mode (default)
func TestMockCollection_BulkWrite_Ordered(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Prepare models with invalid filter (will cause error)
	models := []mongo.WriteModel{
		&mongo.InsertOneModel{
			Document: bson.M{"name": "Alice", "age": 25},
		},
		&mongo.UpdateOneModel{
			Filter: invalidBsonTypeForBulkWrite{Channel: make(chan int)}, // Invalid filter
			Update: bson.M{"$set": bson.M{"age": 30}},
		},
		&mongo.InsertOneModel{
			Document: bson.M{"name": "Charlie", "age": 35},
		},
	}

	// Execute bulk write (ordered mode by default)
	result, err := coll.BulkWrite(ctx, models)
	if err == nil {
		t.Error("Expected error for invalid filter, got nil")
	}

	// First insert should succeed
	if result.InsertedCount != 1 {
		t.Errorf("Expected InsertedCount 1 (before error), got %d", result.InsertedCount)
	}

	// Verify only first document was inserted
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 document (stopped at error), got %d", count)
	}
}

// TestMockCollection_BulkWrite_Unordered tests BulkWrite with unordered mode
func TestMockCollection_BulkWrite_Unordered(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Prepare models with invalid filter (will cause error)
	models := []mongo.WriteModel{
		&mongo.InsertOneModel{
			Document: bson.M{"name": "Alice", "age": 25},
		},
		&mongo.UpdateOneModel{
			Filter: invalidBsonTypeForBulkWrite{Channel: make(chan int)}, // Invalid filter
			Update: bson.M{"$set": bson.M{"age": 30}},
		},
		&mongo.InsertOneModel{
			Document: bson.M{"name": "Charlie", "age": 35},
		},
	}

	// Execute bulk write with unordered mode
	ordered := false
	opts := options.BulkWrite().SetOrdered(ordered)
	result, err := coll.BulkWrite(ctx, models, opts)
	if err != nil {
		t.Fatalf("BulkWrite failed: %v", err)
	}

	// Both inserts should succeed (continue after error)
	if result.InsertedCount != 2 {
		t.Errorf("Expected InsertedCount 2 (continued after error), got %d", result.InsertedCount)
	}

	// Verify both documents were inserted
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("CountDocuments failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 documents (continued after error), got %d", count)
	}
}

// TestMockCollection_BulkWrite_InvalidModel tests BulkWrite with invalid model type
func TestMockCollection_BulkWrite_InvalidModel(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)

	ctx := context.Background()

	// Create invalid model (nil)
	models := []mongo.WriteModel{
		nil,
	}

	// Execute bulk write
	result, err := coll.BulkWrite(ctx, models)
	if err == nil {
		t.Error("Expected error for invalid model, got nil")
	}

	// No operations should succeed
	if result.InsertedCount != 0 {
		t.Errorf("Expected InsertedCount 0, got %d", result.InsertedCount)
	}
}

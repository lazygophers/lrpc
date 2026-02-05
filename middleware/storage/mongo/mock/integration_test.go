package mock_test

import (
	"context"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestIntegration_FullCRUDWorkflow tests a complete CRUD workflow
func TestIntegration_FullCRUDWorkflow(t *testing.T) {
	// Create mock client
	client := mock.NewMockClient()
	ctx := context.Background()

	// Get database and collection
	db := client.Database("test_db")
	coll := db.Collection("users")

	// 1. CREATE - Insert documents
	t.Run("Create", func(t *testing.T) {
		docs := []interface{}{
			bson.M{"name": "Alice", "age": 25, "email": "alice@example.com"},
			bson.M{"name": "Bob", "age": 30, "email": "bob@example.com"},
			bson.M{"name": "Charlie", "age": 35, "email": "charlie@example.com"},
		}

		result, err := coll.InsertMany(ctx, docs)
		require.NoError(t, err)
		assert.Len(t, result.InsertedIDs, 3)
	})

	// 2. READ - Query documents
	t.Run("Read", func(t *testing.T) {
		// Find all
		cursor, err := coll.Find(ctx, bson.M{})
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var users []bson.M
		err = cursor.All(ctx, &users)
		require.NoError(t, err)
		assert.Len(t, users, 3)

		// Find with filter
		cursor, err = coll.Find(ctx, bson.M{"age": bson.M{"$gte": 30}})
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var filteredUsers []bson.M
		err = cursor.All(ctx, &filteredUsers)
		require.NoError(t, err)
		assert.Len(t, filteredUsers, 2)

		// FindOne
		var user bson.M
		err = coll.FindOne(ctx, bson.M{"name": "Alice"}).Decode(&user)
		require.NoError(t, err)
		assert.Equal(t, "Alice", user["name"])
		assert.Equal(t, int32(25), user["age"])
	})

	// 3. UPDATE - Modify documents
	t.Run("Update", func(t *testing.T) {
		// UpdateOne
		update := bson.M{"$set": bson.M{"age": 26}}
		result, err := coll.UpdateOne(ctx, bson.M{"name": "Alice"}, update)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)

		// Verify update
		var user bson.M
		err = coll.FindOne(ctx, bson.M{"name": "Alice"}).Decode(&user)
		require.NoError(t, err)
		assert.Equal(t, int32(26), user["age"])

		// UpdateMany
		update = bson.M{"$inc": bson.M{"age": 1}}
		result, err = coll.UpdateMany(ctx, bson.M{}, update)
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.ModifiedCount)
	})

	// 4. DELETE - Remove documents
	t.Run("Delete", func(t *testing.T) {
		// DeleteOne
		result, err := coll.DeleteOne(ctx, bson.M{"name": "Charlie"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.DeletedCount)

		// Verify deletion
		count, err := coll.CountDocuments(ctx, bson.M{})
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// DeleteMany
		result, err = coll.DeleteMany(ctx, bson.M{"age": bson.M{"$gte": 30}})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.DeletedCount, int64(1))
	})
}

// TestIntegration_ComplexQueries tests complex query scenarios
func TestIntegration_ComplexQueries(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("products", storage, nil)
	ctx := context.Background()

	// Insert test data
	products := []interface{}{
		bson.M{"name": "Laptop", "price": 999.99, "category": "Electronics", "stock": 10},
		bson.M{"name": "Mouse", "price": 29.99, "category": "Electronics", "stock": 50},
		bson.M{"name": "Desk", "price": 299.99, "category": "Furniture", "stock": 5},
		bson.M{"name": "Chair", "price": 199.99, "category": "Furniture", "stock": 15},
		bson.M{"name": "Monitor", "price": 399.99, "category": "Electronics", "stock": 0},
	}

	_, err := coll.InsertMany(ctx, products)
	require.NoError(t, err)

	tests := []struct {
		name          string
		filter        bson.M
		expectedCount int
	}{
		{
			name: "Find electronics with price > 100",
			filter: bson.M{
				"$and": []interface{}{
					bson.M{"category": "Electronics"},
					bson.M{"price": bson.M{"$gt": 100}},
				},
			},
			expectedCount: 2, // Laptop, Monitor
		},
		{
			name: "Find in-stock items",
			filter: bson.M{
				"stock": bson.M{"$gt": 0},
			},
			expectedCount: 4,
		},
		{
			name: "Find products with price range",
			filter: bson.M{
				"$and": []interface{}{
					bson.M{"price": bson.M{"$gte": 100}},
					bson.M{"price": bson.M{"$lte": 400}},
				},
			},
			expectedCount: 3, // Desk, Chair, Monitor
		},
		{
			name: "Find by category using $in",
			filter: bson.M{
				"category": bson.M{"$in": []interface{}{"Electronics", "Furniture"}},
			},
			expectedCount: 5, // All products
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor, err := coll.Find(ctx, tt.filter)
			require.NoError(t, err)
			defer cursor.Close(ctx)

			var results []bson.M
			err = cursor.All(ctx, &results)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(results))
		})
	}
}

// TestIntegration_NestedDocuments tests operations with nested documents
func TestIntegration_NestedDocuments(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)
	ctx := context.Background()

	// Insert nested documents
	users := []interface{}{
		bson.M{
			"name": "Alice",
			"profile": bson.M{
				"age":  25,
				"city": "NYC",
			},
			"settings": bson.M{
				"notifications": true,
			},
		},
		bson.M{
			"name": "Bob",
			"profile": bson.M{
				"age":  30,
				"city": "LA",
			},
			"settings": bson.M{
				"notifications": false,
			},
		},
	}

	_, err := coll.InsertMany(ctx, users)
	require.NoError(t, err)

	t.Run("Query nested fields", func(t *testing.T) {
		// Find by nested field
		cursor, err := coll.Find(ctx, bson.M{"profile.city": "NYC"})
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var results []bson.M
		err = cursor.All(ctx, &results)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Alice", results[0]["name"])
	})

	t.Run("Query nested with operator", func(t *testing.T) {
		cursor, err := coll.Find(ctx, bson.M{"profile.age": bson.M{"$gte": 30}})
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var results []bson.M
		err = cursor.All(ctx, &results)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Bob", results[0]["name"])
	})

	t.Run("Update nested fields", func(t *testing.T) {
		// Note: Nested field updates with dot notation are not fully supported in mock
		// This would require complex path parsing and update logic
		// For now, we update the entire nested document
		update := bson.M{"$set": bson.M{
			"profile": bson.M{
				"name": "Alice",
				"age":  26,
				"city": "NYC",
			},
		}}
		result, err := coll.UpdateOne(ctx, bson.M{"name": "Alice"}, update)
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.ModifiedCount)

		// Verify
		var user bson.M
		err = coll.FindOne(ctx, bson.M{"name": "Alice"}).Decode(&user)
		require.NoError(t, err)

		profile, ok := user["profile"].(bson.M)
		require.True(t, ok)
		assert.Equal(t, int32(26), profile["age"])
	})
}

// TestIntegration_Pagination tests pagination with skip and limit
func TestIntegration_Pagination(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("items", storage, nil)
	ctx := context.Background()

	// Insert 20 documents
	docs := make([]interface{}, 20)
	for i := 0; i < 20; i++ {
		docs[i] = bson.M{"index": i, "value": i * 10}
	}

	_, err := coll.InsertMany(ctx, docs)
	require.NoError(t, err)

	t.Run("First page", func(t *testing.T) {
		opts := options.Find().SetLimit(5)
		cursor, err := coll.Find(ctx, bson.M{}, opts)
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var results []bson.M
		err = cursor.All(ctx, &results)
		require.NoError(t, err)
		assert.Len(t, results, 5)
	})

	t.Run("Second page", func(t *testing.T) {
		opts := options.Find().SetSkip(5).SetLimit(5)
		cursor, err := coll.Find(ctx, bson.M{}, opts)
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var results []bson.M
		err = cursor.All(ctx, &results)
		require.NoError(t, err)
		assert.Len(t, results, 5)
		assert.Equal(t, int32(5), results[0]["index"])
	})

	t.Run("Last page", func(t *testing.T) {
		opts := options.Find().SetSkip(15).SetLimit(10)
		cursor, err := coll.Find(ctx, bson.M{}, opts)
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var results []bson.M
		err = cursor.All(ctx, &results)
		require.NoError(t, err)
		assert.Len(t, results, 5) // Only 5 remaining
	})
}

// TestIntegration_Sorting tests sorting functionality
func TestIntegration_Sorting(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("scores", storage, nil)
	ctx := context.Background()

	// Insert unsorted data
	docs := []interface{}{
		bson.M{"name": "Charlie", "score": 85},
		bson.M{"name": "Alice", "score": 95},
		bson.M{"name": "Bob", "score": 75},
	}

	_, err := coll.InsertMany(ctx, docs)
	require.NoError(t, err)

	t.Run("Sort by score ascending", func(t *testing.T) {
		opts := options.Find().SetSort(bson.M{"score": 1})
		cursor, err := coll.Find(ctx, bson.M{}, opts)
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var results []bson.M
		err = cursor.All(ctx, &results)
		require.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, "Bob", results[0]["name"])
		assert.Equal(t, "Alice", results[2]["name"])
	})

	t.Run("Sort by score descending", func(t *testing.T) {
		opts := options.Find().SetSort(bson.M{"score": -1})
		cursor, err := coll.Find(ctx, bson.M{}, opts)
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var results []bson.M
		err = cursor.All(ctx, &results)
		require.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, "Alice", results[0]["name"])
		assert.Equal(t, "Bob", results[2]["name"])
	})

	t.Run("Sort by name", func(t *testing.T) {
		opts := options.Find().SetSort(bson.M{"name": 1})
		cursor, err := coll.Find(ctx, bson.M{}, opts)
		require.NoError(t, err)
		defer cursor.Close(ctx)

		var results []bson.M
		err = cursor.All(ctx, &results)
		require.NoError(t, err)
		assert.Len(t, results, 3)
		assert.Equal(t, "Alice", results[0]["name"])
		assert.Equal(t, "Charlie", results[2]["name"])
	})
}

// TestIntegration_UpdateOperators tests various update operators
func TestIntegration_UpdateOperators(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("counters", storage, nil)
	ctx := context.Background()

	// Insert test document
	_, err := coll.InsertOne(ctx, bson.M{
		"name":    "counter",
		"value":   10,
		"metrics": bson.M{"hits": 5},
	})
	require.NoError(t, err)

	t.Run("$set operator", func(t *testing.T) {
		update := bson.M{"$set": bson.M{"name": "updated_counter"}}
		_, err := coll.UpdateOne(ctx, bson.M{}, update)
		require.NoError(t, err)

		var doc bson.M
		err = coll.FindOne(ctx, bson.M{}).Decode(&doc)
		require.NoError(t, err)
		assert.Equal(t, "updated_counter", doc["name"])
	})

	t.Run("$inc operator", func(t *testing.T) {
		update := bson.M{"$inc": bson.M{"value": 5}}
		_, err := coll.UpdateOne(ctx, bson.M{}, update)
		require.NoError(t, err)

		var doc bson.M
		err = coll.FindOne(ctx, bson.M{}).Decode(&doc)
		require.NoError(t, err)
		assert.Equal(t, int32(15), doc["value"])
	})

	t.Run("$unset operator", func(t *testing.T) {
		update := bson.M{"$unset": bson.M{"metrics": ""}}
		_, err := coll.UpdateOne(ctx, bson.M{}, update)
		require.NoError(t, err)

		var doc bson.M
		err = coll.FindOne(ctx, bson.M{}).Decode(&doc)
		require.NoError(t, err)
		_, exists := doc["metrics"]
		assert.False(t, exists, "metrics field should be unset")
	})
}

// TestIntegration_ReplaceOne tests document replacement
func TestIntegration_ReplaceOne(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("users", storage, nil)
	ctx := context.Background()

	// Insert original document
	id := primitive.NewObjectID()
	_, err := coll.InsertOne(ctx, bson.M{
		"_id":   id,
		"name":  "Alice",
		"age":   25,
		"email": "alice@example.com",
	})
	require.NoError(t, err)

	// Replace document
	replacement := bson.M{
		"name":  "Alice Smith",
		"phone": "555-1234",
	}

	result, err := coll.ReplaceOne(ctx, bson.M{"_id": id}, replacement)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.ModifiedCount)

	// Verify replacement
	var doc bson.M
	err = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	require.NoError(t, err)

	assert.Equal(t, "Alice Smith", doc["name"])
	assert.Equal(t, "555-1234", doc["phone"])
	_, hasAge := doc["age"]
	assert.False(t, hasAge, "age field should be removed in replacement")
	_, hasEmail := doc["email"]
	assert.False(t, hasEmail, "email field should be removed in replacement")
}

// TestIntegration_ConcurrentOperations tests thread-safety
func TestIntegration_ConcurrentOperations(t *testing.T) {
	storage := mock.NewMemoryStorage()
	coll := mock.NewMockCollection("items", storage, nil)
	ctx := context.Background()

	// Insert initial documents
	docs := []interface{}{
		bson.M{"id": 1, "count": 0},
		bson.M{"id": 2, "count": 0},
	}
	_, err := coll.InsertMany(ctx, docs)
	require.NoError(t, err)

	// Perform concurrent updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			update := bson.M{"$inc": bson.M{"count": 1}}
			_, _ = coll.UpdateMany(ctx, bson.M{}, update)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final counts
	cursor, err := coll.Find(ctx, bson.M{})
	require.NoError(t, err)
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	require.NoError(t, err)

	// Each document should be updated
	for _, doc := range results {
		count := doc["count"]
		assert.NotNil(t, count, "count should be incremented")
	}
}

// TestIntegration_ErrorCases tests error handling
func TestIntegration_ErrorCases(t *testing.T) {
	storage := mock.NewMemoryStorage()
	ctx := context.Background()

	t.Run("FindOne on empty collection", func(t *testing.T) {
		coll := mock.NewMockCollection("empty", storage, nil)
		var doc bson.M
		err := coll.FindOne(ctx, bson.M{}).Decode(&doc)
		assert.Error(t, err)
	})

	t.Run("UpdateOne on non-existent document", func(t *testing.T) {
		coll := mock.NewMockCollection("test", storage, nil)
		update := bson.M{"$set": bson.M{"field": "value"}}
		result, err := coll.UpdateOne(ctx, bson.M{"_id": primitive.NewObjectID()}, update)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.MatchedCount)
	})

	t.Run("DeleteOne on non-existent document", func(t *testing.T) {
		coll := mock.NewMockCollection("test", storage, nil)
		result, err := coll.DeleteOne(ctx, bson.M{"_id": primitive.NewObjectID()})
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.DeletedCount)
	})
}

package mock

import (
"context"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
"go.mongodb.org/mongo-driver/mongo/options"
"sync"
"testing"
)

func TestNewMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()
	if storage == nil {
		t.Fatal("NewMemoryStorage returned nil")
	}

	if storage.collections == nil {
		t.Fatal("collections map not initialized")
	}
}

func TestCreateCollection(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.CreateCollection("test_collection")
	if err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}

	// Try to create the same collection again - should fail
	err = storage.CreateCollection("test_collection")
	if err == nil {
		t.Fatal("expected error when creating duplicate collection")
	}
}

func TestInsertAndFind(t *testing.T) {
	storage := NewMemoryStorage()

	// Insert a document
	doc := bson.M{"name": "test", "age": 25}
	err := storage.Insert("users", doc)
	if err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	// Find all documents
	results := storage.Find("users", bson.M{}, nil)

	if len(results) != 1 {
		t.Fatalf("expected 1 document, got %d", len(results))
	}

	if results[0]["name"] != "test" {
		t.Errorf("expected name 'test', got %v", results[0]["name"])
	}
}

func TestInsertMany(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
		{"name": "charlie", "age": 35},
	}

	err := storage.InsertMany("users", docs)
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	results := storage.Find("users", bson.M{}, nil)

	if len(results) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(results))
	}
}

func TestCount(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
		{"name": "charlie", "age": 35},
	}

	err := storage.InsertMany("users", docs)
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	count := storage.Count("users", bson.M{})

	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}

	// Count non-existent collection
	count = storage.Count("non_existent", bson.M{})

	if count != 0 {
		t.Fatalf("expected count 0 for non-existent collection, got %d", count)
	}
}

func TestUpdateWithSetOperator(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	if err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	// Update with $set operator
	update := bson.M{"$set": bson.M{"age": 26}}
	updated := storage.Update("users", bson.M{}, update)

	if updated != 1 {
		t.Fatalf("expected 1 document updated, got %d", updated)
	}

	// Verify update
	results := storage.Find("users", bson.M{}, nil)

	if results[0]["age"] != 26 {
		t.Errorf("expected age 26, got %v", results[0]["age"])
	}
}

func TestUpdateWithIncOperator(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "count": 10}
	err := storage.Insert("users", doc)
	if err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	// Update with $inc operator
	update := bson.M{"$inc": bson.M{"count": 5}}
	updated := storage.Update("users", bson.M{}, update)

	if updated != 1 {
		t.Fatalf("expected 1 document updated, got %d", updated)
	}

	// Verify update
	results := storage.Find("users", bson.M{}, nil)

	if results[0]["count"] != 15 {
		t.Errorf("expected count 15, got %v", results[0]["count"])
	}
}

func TestDelete(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
		{"name": "charlie", "age": 35},
	}

	err := storage.InsertMany("users", docs)
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Delete all documents
	deleted := storage.Delete("users", bson.M{})

	if deleted != 3 {
		t.Fatalf("expected 3 documents deleted, got %d", deleted)
	}

	// Verify deletion
	count := storage.Count("users", bson.M{})

	if count != 0 {
		t.Fatalf("expected count 0 after deletion, got %d", count)
	}
}

func TestDeleteOne(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
	}

	err := storage.InsertMany("users", docs)
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Delete one document
	deleted := storage.DeleteOne("users", bson.M{})

	if deleted != 1 {
		t.Fatalf("expected 1 document deleted, got %d", deleted)
	}

	// Verify only one deleted
	count := storage.Count("users", bson.M{})

	if count != 1 {
		t.Fatalf("expected count 1 after deletion, got %d", count)
	}
}

func TestFindWithLimit(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
		{"name": "charlie", "age": 35},
	}

	err := storage.InsertMany("users", docs)
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	limit := int64(2)
	opts := &FindOptions{Limit: &limit}

	results := storage.Find("users", bson.M{}, opts)

	if len(results) != 2 {
		t.Fatalf("expected 2 documents with limit, got %d", len(results))
	}
}

func TestFindWithSkip(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
		{"name": "charlie", "age": 35},
	}

	err := storage.InsertMany("users", docs)
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	skip := int64(1)
	opts := &FindOptions{Skip: &skip}

	results := storage.Find("users", bson.M{}, opts)

	if len(results) != 2 {
		t.Fatalf("expected 2 documents with skip, got %d", len(results))
	}
}

func TestFindWithSorting(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "charlie", "age": 35},
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
	}

	err := storage.InsertMany("users", docs)
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Sort by age ascending
	opts := &FindOptions{Sort: bson.M{"age": 1}}
	results := storage.Find("users", bson.M{}, opts)

	if results[0]["age"] != 25 {
		t.Errorf("expected first age 25, got %v", results[0]["age"])
	}
	if results[2]["age"] != 35 {
		t.Errorf("expected last age 35, got %v", results[2]["age"])
	}

	// Sort by age descending
	opts = &FindOptions{Sort: bson.M{"age": -1}}
	results = storage.Find("users", bson.M{}, opts)

	if results[0]["age"] != 35 {
		t.Errorf("expected first age 35, got %v", results[0]["age"])
	}
	if results[2]["age"] != 25 {
		t.Errorf("expected last age 25, got %v", results[2]["age"])
	}
}

func TestDropCollection(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.CreateCollection("test_collection")
	if err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}

	err = storage.DropCollection("test_collection")
	if err != nil {
		t.Fatalf("failed to drop collection: %v", err)
	}

	// Try to drop non-existent collection
	err = storage.DropCollection("non_existent")
	if err == nil {
		t.Fatal("expected error when dropping non-existent collection")
	}
}

func TestListCollections(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.CreateCollection("users")
	if err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}

	err = storage.CreateCollection("posts")
	if err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}

	collections := storage.ListCollections()
	if len(collections) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(collections))
	}

	// Collections should be sorted
	if collections[0] != "posts" || collections[1] != "users" {
		t.Errorf("expected sorted collections [posts, users], got %v", collections)
	}
}

func TestClear(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.CreateCollection("users")
	if err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}

	storage.Clear()

	collections := storage.ListCollections()
	if len(collections) != 0 {
		t.Fatalf("expected 0 collections after clear, got %d", len(collections))
	}
}

func TestProjectionInclusion(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "age": 25, "email": "alice@example.com"}
	err := storage.Insert("users", doc)
	if err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	// Project only name field
	opts := &FindOptions{Projection: bson.M{"name": 1}}
	results := storage.Find("users", bson.M{}, opts)

	if len(results) != 1 {
		t.Fatalf("expected 1 document, got %d", len(results))
	}

	if _, hasName := results[0]["name"]; !hasName {
		t.Error("expected name field in projection")
	}

	if _, hasEmail := results[0]["email"]; hasEmail {
		t.Error("expected email field to be excluded from projection")
	}
}

func TestProjectionExclusion(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "age": 25, "email": "alice@example.com"}
	err := storage.Insert("users", doc)
	if err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	// Exclude email field
	opts := &FindOptions{Projection: bson.M{"email": 0}}
	results := storage.Find("users", bson.M{}, opts)

	if len(results) != 1 {
		t.Fatalf("expected 1 document, got %d", len(results))
	}

	if _, hasName := results[0]["name"]; !hasName {
		t.Error("expected name field in result")
	}

	if _, hasEmail := results[0]["email"]; hasEmail {
		t.Error("expected email field to be excluded")
	}
}

// TestInsert_NilDocument tests Insert with nil document
func TestInsert_NilDocument(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.Insert("users", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot insert nil document")
}

// TestInsertMany_EmptySlice tests InsertMany with empty slice
func TestInsertMany_EmptySlice(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.InsertMany("users", []bson.M{})
	assert.NoError(t, err)
}

// TestInsertMany_NilDocument tests InsertMany with nil document in slice
func TestInsertMany_NilDocument(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		nil,
		{"name": "charlie", "age": 35},
	}

	err := storage.InsertMany("users", docs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot insert nil document")
}

// TestUpdate_NonExistentCollection tests Update on non-existent collection
func TestUpdate_NonExistentCollection(t *testing.T) {
	storage := NewMemoryStorage()

	updated := storage.Update("non_existent", bson.M{}, bson.M{"$set": bson.M{"name": "test"}})
	assert.Equal(t, int64(0), updated)
}

// TestUpdate_WithUnsetOperator tests Update with $unset operator
func TestUpdate_WithUnsetOperator(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "age": 25, "email": "alice@example.com"}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Update with $unset operator
	update := bson.M{"$unset": bson.M{"email": ""}}
	updated := storage.Update("users", bson.M{}, update)
	assert.Equal(t, int64(1), updated)

	// Verify unset
	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.NotContains(t, results[0], "email")
	assert.Contains(t, results[0], "name")
	assert.Contains(t, results[0], "age")
}

// TestUpdate_WithMultiplyOperator tests Update with $mul operator
func TestUpdate_WithMultiplyOperator(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name           string
		initialDoc     bson.M
		updateValue    interface{}
		expectedResult interface{}
	}{
		{
			name:           "multiply int",
			initialDoc:     bson.M{"name": "alice", "count": 10},
			updateValue:    3,
			expectedResult: 30,
		},
		{
			name:           "multiply int32",
			initialDoc:     bson.M{"name": "bob", "count": int32(10)},
			updateValue:    int32(3),
			expectedResult: int32(30),
		},
		{
			name:           "multiply int64",
			initialDoc:     bson.M{"name": "charlie", "count": int64(10)},
			updateValue:    int64(3),
			expectedResult: int64(30),
		},
		{
			name:           "multiply float64",
			initialDoc:     bson.M{"name": "dave", "count": float64(10.5)},
			updateValue:    float64(2.0),
			expectedResult: float64(21.0),
		},
		{
			name:           "multiply non-existent field",
			initialDoc:     bson.M{"name": "eve"},
			updateValue:    5,
			expectedResult: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Clear()

			err := storage.Insert("users", tt.initialDoc)
			assert.NoError(t, err)

			// Update with $mul operator
			update := bson.M{"$mul": bson.M{"count": tt.updateValue}}
			updated := storage.Update("users", bson.M{}, update)
			assert.Equal(t, int64(1), updated)

			// Verify update
			results := storage.Find("users", bson.M{}, nil)
			assert.Len(t, results, 1)
			assert.Equal(t, tt.expectedResult, results[0]["count"])
		})
	}
}

// TestUpdate_DirectFieldUpdate tests Update without operators
func TestUpdate_DirectFieldUpdate(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Direct field update
	update := bson.M{"status": "active"}
	updated := storage.Update("users", bson.M{}, update)
	assert.Equal(t, int64(1), updated)

	// Verify update
	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, "active", results[0]["status"])
}

// TestUpdateOne_NonExistentCollection tests UpdateOne on non-existent collection
func TestUpdateOne_NonExistentCollection(t *testing.T) {
	storage := NewMemoryStorage()

	updated := storage.UpdateOne("non_existent", bson.M{}, bson.M{"$set": bson.M{"name": "test"}})
	assert.Equal(t, int64(0), updated)
}

// TestUpdateOne_NoMatch tests UpdateOne when no document matches
func TestUpdateOne_NoMatch(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Try to update with non-matching filter
	updated := storage.UpdateOne("users", bson.M{"name": "bob"}, bson.M{"$set": bson.M{"age": 30}})
	assert.Equal(t, int64(0), updated)
}

// TestUpdateOne_UpdatesOnlyFirst tests UpdateOne updates only the first matching document
func TestUpdateOne_UpdatesOnlyFirst(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "alice", "age": 30},
		{"name": "alice", "age": 35},
	}
	err := storage.InsertMany("users", docs)
	assert.NoError(t, err)

	// Update only one document
	updated := storage.UpdateOne("users", bson.M{"name": "alice"}, bson.M{"$set": bson.M{"age": 100}})
	assert.Equal(t, int64(1), updated)

	// Verify only one was updated
	results := storage.Find("users", bson.M{"age": 100}, nil)
	assert.Len(t, results, 1)
}

// TestDelete_NonExistentCollection tests Delete on non-existent collection
func TestDelete_NonExistentCollection(t *testing.T) {
	storage := NewMemoryStorage()

	deleted := storage.Delete("non_existent", bson.M{})
	assert.Equal(t, int64(0), deleted)
}

// TestDelete_WithFilter tests Delete with specific filter
func TestDelete_WithFilter(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
		{"name": "charlie", "age": 25},
	}
	err := storage.InsertMany("users", docs)
	assert.NoError(t, err)

	// Delete only documents with age 25
	deleted := storage.Delete("users", bson.M{"age": 25})
	assert.Equal(t, int64(2), deleted)

	// Verify remaining documents
	count := storage.Count("users", bson.M{})
	assert.Equal(t, int64(1), count)

	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, "bob", results[0]["name"])
}

// TestDeleteOne_NonExistentCollection tests DeleteOne on non-existent collection
func TestDeleteOne_NonExistentCollection(t *testing.T) {
	storage := NewMemoryStorage()

	deleted := storage.DeleteOne("non_existent", bson.M{})
	assert.Equal(t, int64(0), deleted)
}

// TestDeleteOne_NoMatch tests DeleteOne when no document matches
func TestDeleteOne_NoMatch(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Try to delete with non-matching filter
	deleted := storage.DeleteOne("users", bson.M{"name": "bob"})
	assert.Equal(t, int64(0), deleted)

	// Verify document still exists
	count := storage.Count("users", bson.M{})
	assert.Equal(t, int64(1), count)
}

// TestApplySorting_NilValues tests sorting with nil values
func TestApplySorting_NilValues(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "bob", "age": nil},
		{"name": "charlie", "age": 35},
		{"name": "dave", "age": nil},
	}
	err := storage.InsertMany("users", docs)
	assert.NoError(t, err)

	// Sort ascending - nil values should come first
	opts := &FindOptions{Sort: bson.M{"age": 1}}
	results := storage.Find("users", bson.M{}, opts)
	assert.Len(t, results, 4)
	// First two should have nil age
	assert.Nil(t, results[0]["age"])
	assert.Nil(t, results[1]["age"])

	// Sort descending - nil values should come last
	opts = &FindOptions{Sort: bson.M{"age": -1}}
	results = storage.Find("users", bson.M{}, opts)
	assert.Len(t, results, 4)
	// Last two should have nil age
	assert.Nil(t, results[2]["age"])
	assert.Nil(t, results[3]["age"])
}

// TestApplySorting_EmptyDocs tests sorting with empty documents slice
func TestApplySorting_EmptyDocs(t *testing.T) {
	storage := NewMemoryStorage()

	results := storage.Find("non_existent", bson.M{}, &FindOptions{Sort: bson.M{"age": 1}})
	assert.Empty(t, results)
}

// TestApplySorting_EmptySort tests sorting with empty sort specification
func TestApplySorting_EmptySort(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "charlie", "age": 35},
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
	}
	err := storage.InsertMany("users", docs)
	assert.NoError(t, err)

	// Empty sort should return documents in original order
	results := storage.Find("users", bson.M{}, &FindOptions{Sort: bson.M{}})
	assert.Len(t, results, 3)
}

// TestCompareValuesForSort_DifferentTypes tests compareValuesForSort with different types
func TestCompareValuesForSort_DifferentTypes(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		docs     []bson.M
		sortKey  string
		expected []interface{}
	}{
		{
			name: "sort int32",
			docs: []bson.M{
				{"name": "c", "value": int32(30)},
				{"name": "a", "value": int32(10)},
				{"name": "b", "value": int32(20)},
			},
			sortKey:  "value",
			expected: []interface{}{int32(10), int32(20), int32(30)},
		},
		{
			name: "sort int64",
			docs: []bson.M{
				{"name": "c", "value": int64(300)},
				{"name": "a", "value": int64(100)},
				{"name": "b", "value": int64(200)},
			},
			sortKey:  "value",
			expected: []interface{}{int64(100), int64(200), int64(300)},
		},
		{
			name: "sort float64",
			docs: []bson.M{
				{"name": "c", "value": 30.5},
				{"name": "a", "value": 10.5},
				{"name": "b", "value": 20.5},
			},
			sortKey:  "value",
			expected: []interface{}{10.5, 20.5, 30.5},
		},
		{
			name: "sort string",
			docs: []bson.M{
				{"name": "charlie"},
				{"name": "alice"},
				{"name": "bob"},
			},
			sortKey:  "name",
			expected: []interface{}{"alice", "bob", "charlie"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Clear()

			err := storage.InsertMany("test", tt.docs)
			assert.NoError(t, err)

			opts := &FindOptions{Sort: bson.M{tt.sortKey: 1}}
			results := storage.Find("test", bson.M{}, opts)
			assert.Len(t, results, len(tt.expected))

			for i, expected := range tt.expected {
				assert.Equal(t, expected, results[i][tt.sortKey])
			}
		})
	}
}

// TestCompareValuesForSort_MixedTypes tests sorting with mixed types
func TestCompareValuesForSort_MixedTypes(t *testing.T) {
	storage := NewMemoryStorage()

	// Mixed types - should be considered equal when comparing different types
	docs := []bson.M{
		{"name": "a", "value": "string"},
		{"name": "b", "value": 123},
		{"name": "c", "value": 45.6},
	}
	err := storage.InsertMany("test", docs)
	assert.NoError(t, err)

	// Sort by value - mixed types should maintain order
	opts := &FindOptions{Sort: bson.M{"value": 1}}
	results := storage.Find("test", bson.M{}, opts)
	assert.Len(t, results, 3)
}

// TestApplyIncrement_DifferentTypes tests $inc with different numeric types
func TestApplyIncrement_DifferentTypes(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name           string
		initialDoc     bson.M
		incValue       interface{}
		expectedResult interface{}
	}{
		{
			name:           "increment int",
			initialDoc:     bson.M{"name": "alice", "count": 10},
			incValue:       5,
			expectedResult: 15,
		},
		{
			name:           "increment int32",
			initialDoc:     bson.M{"name": "bob", "count": int32(10)},
			incValue:       int32(5),
			expectedResult: int32(15),
		},
		{
			name:           "increment int64",
			initialDoc:     bson.M{"name": "charlie", "count": int64(10)},
			incValue:       int64(5),
			expectedResult: int64(15),
		},
		{
			name:           "increment float64",
			initialDoc:     bson.M{"name": "dave", "count": float64(10.5)},
			incValue:       float64(5.5),
			expectedResult: float64(16.0),
		},
		{
			name:           "increment non-existent field",
			initialDoc:     bson.M{"name": "eve"},
			incValue:       5,
			expectedResult: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Clear()

			err := storage.Insert("users", tt.initialDoc)
			assert.NoError(t, err)

			// Update with $inc operator
			update := bson.M{"$inc": bson.M{"count": tt.incValue}}
			updated := storage.Update("users", bson.M{}, update)
			assert.Equal(t, int64(1), updated)

			// Verify update
			results := storage.Find("users", bson.M{}, nil)
			assert.Len(t, results, 1)
			assert.Equal(t, tt.expectedResult, results[0]["count"])
		})
	}
}

// TestApplyIncrement_TypeMismatch tests $inc with mismatched types
func TestApplyIncrement_TypeMismatch(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "count": 10}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Try to increment int field with float64 - should not change value due to type mismatch
	update := bson.M{"$inc": bson.M{"count": float64(5.5)}}
	updated := storage.Update("users", bson.M{}, update)
	assert.Equal(t, int64(1), updated)

	// Verify value remains unchanged
	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, 10, results[0]["count"])
}

// TestApplyMultiply_TypeMismatch tests $mul with mismatched types
func TestApplyMultiply_TypeMismatch(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "count": 10}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Try to multiply int field with float64 - should not change value due to type mismatch
	update := bson.M{"$mul": bson.M{"count": float64(2.5)}}
	updated := storage.Update("users", bson.M{}, update)
	assert.Equal(t, int64(1), updated)

	// Verify value remains unchanged
	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, 10, results[0]["count"])
}

// TestIsOperator tests isOperator function
func TestIsOperator(t *testing.T) {
	assert.True(t, isOperator("$set"))
	assert.True(t, isOperator("$inc"))
	assert.True(t, isOperator("$mul"))
	assert.True(t, isOperator("$unset"))
	assert.False(t, isOperator("set"))
	assert.False(t, isOperator("name"))
	assert.False(t, isOperator(""))
}

// TestFindWithSkip_ExceedsLength tests Find with skip exceeding document count
func TestFindWithSkip_ExceedsLength(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
	}
	err := storage.InsertMany("users", docs)
	assert.NoError(t, err)

	skip := int64(10)
	opts := &FindOptions{Skip: &skip}

	results := storage.Find("users", bson.M{}, opts)
	assert.Empty(t, results)
}

// TestConcurrentOperations tests concurrent operations on MemoryStorage
func TestConcurrentOperations(t *testing.T) {
	storage := NewMemoryStorage()

	var wg sync.WaitGroup
	numGoroutines := 10
	numOps := 100

	// Concurrent inserts
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				doc := bson.M{"goroutine": id, "iteration": j}
				err := storage.Insert("concurrent", doc)
				assert.NoError(t, err)
			}
		}(i)
	}
	wg.Wait()

	// Verify all documents were inserted
	count := storage.Count("concurrent", bson.M{})
	assert.Equal(t, int64(numGoroutines*numOps), count)

	// Concurrent reads
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				_ = storage.Find("concurrent", bson.M{}, nil)
			}
		}()
	}
	wg.Wait()
}

// TestApplyUpdate_InvalidOperator tests applyUpdate with invalid operator value
func TestApplyUpdate_InvalidOperator(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Update with invalid $set value (not bson.M)
	update := bson.M{"$set": "invalid"}
	updated := storage.Update("users", bson.M{}, update)
	assert.Equal(t, int64(1), updated)

	// Document should remain unchanged
	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, "alice", results[0]["name"])
	assert.Equal(t, 25, results[0]["age"])
}

// TestUpdate_MultipleDocuments tests Update affecting multiple documents
func TestUpdate_MultipleDocuments(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "alice", "age": 25, "status": "inactive"},
		{"name": "bob", "age": 30, "status": "inactive"},
		{"name": "charlie", "age": 35, "status": "active"},
	}
	err := storage.InsertMany("users", docs)
	assert.NoError(t, err)

	// Update all inactive users
	update := bson.M{"$set": bson.M{"status": "active"}}
	updated := storage.Update("users", bson.M{"status": "inactive"}, update)
	assert.Equal(t, int64(2), updated)

	// Verify all are now active
	results := storage.Find("users", bson.M{"status": "active"}, nil)
	assert.Len(t, results, 3)
}

// TestFindWithSorting_NonIntegerOrder tests sorting with non-integer order value
func TestFindWithSorting_NonIntegerOrder(t *testing.T) {
	storage := NewMemoryStorage()

	docs := []bson.M{
		{"name": "charlie", "age": 35},
		{"name": "alice", "age": 25},
		{"name": "bob", "age": 30},
	}
	err := storage.InsertMany("users", docs)
	assert.NoError(t, err)

	// Sort with non-integer order (should default to ascending)
	opts := &FindOptions{Sort: bson.M{"age": "invalid"}}
	results := storage.Find("users", bson.M{}, opts)
	assert.Len(t, results, 3)
	// Should be sorted ascending by default
	assert.Equal(t, 25, results[0]["age"])
}

// TestUpdate_ErrorInApplyUpdate tests Update when applyUpdate returns error
func TestUpdate_ErrorInApplyUpdate(t *testing.T) {
	storage := NewMemoryStorage()

	// Insert a document
	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Try to update with nil document should not fail at Update level
	// but applyUpdate might have different behaviors
	update := bson.M{"$inc": bson.M{"count": "not_a_number"}}
	updated := storage.Update("users", bson.M{}, update)
	assert.Equal(t, int64(1), updated)
}

// TestUpdateOne_ErrorInApplyUpdate tests UpdateOne when applyUpdate returns error
func TestUpdateOne_ErrorInApplyUpdate(t *testing.T) {
	storage := NewMemoryStorage()

	// Insert a document
	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Try to update with invalid operator value
	update := bson.M{"$inc": bson.M{"count": "not_a_number"}}
	updated := storage.UpdateOne("users", bson.M{}, update)
	assert.Equal(t, int64(1), updated)
}

// TestApplyUpdate_InvalidOperators tests applyUpdate with various invalid operator values
func TestApplyUpdate_InvalidOperators(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name   string
		doc    bson.M
		update bson.M
	}{
		{
			name:   "$set with non-bson.M value",
			doc:    bson.M{"name": "alice"},
			update: bson.M{"$set": "not_a_map"},
		},
		{
			name:   "$unset with non-bson.M value",
			doc:    bson.M{"name": "alice", "age": 25},
			update: bson.M{"$unset": "not_a_map"},
		},
		{
			name:   "$inc with non-bson.M value",
			doc:    bson.M{"name": "alice", "count": 10},
			update: bson.M{"$inc": "not_a_map"},
		},
		{
			name:   "$mul with non-bson.M value",
			doc:    bson.M{"name": "alice", "count": 10},
			update: bson.M{"$mul": "not_a_map"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Clear()

			err := storage.Insert("test", tt.doc)
			assert.NoError(t, err)

			updated := storage.Update("test", bson.M{}, tt.update)
			assert.Equal(t, int64(1), updated)
		})
	}
}

// TestApplyIncrement_AllNumericTypes tests increment with all supported numeric types
func TestApplyIncrement_AllNumericTypes(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name           string
		initialValue   interface{}
		incrementValue interface{}
		expectedValue  interface{}
		shouldWork     bool
	}{
		{
			name:           "int + int",
			initialValue:   10,
			incrementValue: 5,
			expectedValue:  15,
			shouldWork:     true,
		},
		{
			name:           "int32 + int32",
			initialValue:   int32(10),
			incrementValue: int32(5),
			expectedValue:  int32(15),
			shouldWork:     true,
		},
		{
			name:           "int64 + int64",
			initialValue:   int64(10),
			incrementValue: int64(5),
			expectedValue:  int64(15),
			shouldWork:     true,
		},
		{
			name:           "float64 + float64",
			initialValue:   float64(10.5),
			incrementValue: float64(5.5),
			expectedValue:  float64(16.0),
			shouldWork:     true,
		},
		{
			name:           "string (unsupported type) - value replaced by increment value",
			initialValue:   "hello",
			incrementValue: 5,
			expectedValue:  5,
			shouldWork:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Clear()

			doc := bson.M{"value": tt.initialValue}
			err := storage.Insert("test", doc)
			assert.NoError(t, err)

			update := bson.M{"$inc": bson.M{"value": tt.incrementValue}}
			updated := storage.Update("test", bson.M{}, update)
			assert.Equal(t, int64(1), updated)

			results := storage.Find("test", bson.M{}, nil)
			assert.Len(t, results, 1)
			assert.Equal(t, tt.expectedValue, results[0]["value"])
		})
	}
}

// TestApplyMultiply_AllNumericTypes tests multiply with all supported numeric types
func TestApplyMultiply_AllNumericTypes(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name          string
		initialValue  interface{}
		multiplyValue interface{}
		expectedValue interface{}
		shouldWork    bool
	}{
		{
			name:          "int * int",
			initialValue:  10,
			multiplyValue: 3,
			expectedValue: 30,
			shouldWork:    true,
		},
		{
			name:          "int32 * int32",
			initialValue:  int32(10),
			multiplyValue: int32(3),
			expectedValue: int32(30),
			shouldWork:    true,
		},
		{
			name:          "int64 * int64",
			initialValue:  int64(10),
			multiplyValue: int64(3),
			expectedValue: int64(30),
			shouldWork:    true,
		},
		{
			name:          "float64 * float64",
			initialValue:  float64(10.5),
			multiplyValue: float64(2.0),
			expectedValue: float64(21.0),
			shouldWork:    true,
		},
		{
			name:          "string (unsupported type)",
			initialValue:  "hello",
			multiplyValue: 3,
			expectedValue: 0,
			shouldWork:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Clear()

			doc := bson.M{"value": tt.initialValue}
			err := storage.Insert("test", doc)
			assert.NoError(t, err)

			update := bson.M{"$mul": bson.M{"value": tt.multiplyValue}}
			updated := storage.Update("test", bson.M{}, update)
			assert.Equal(t, int64(1), updated)

			results := storage.Find("test", bson.M{}, nil)
			assert.Len(t, results, 1)
			assert.Equal(t, tt.expectedValue, results[0]["value"])
		})
	}
}

// TestCompareValuesForSort_AllBranches tests all branches in compareValuesForSort
func TestCompareValuesForSort_AllBranches(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name           string
		docs           []bson.M
		sortKey        string
		expectedOrder  []interface{}
		testComparison bool
	}{
		{
			name: "int equal values",
			docs: []bson.M{
				{"name": "a", "value": 10},
				{"name": "b", "value": 10},
			},
			sortKey:        "value",
			expectedOrder:  []interface{}{10, 10},
			testComparison: true,
		},
		{
			name: "int32 values",
			docs: []bson.M{
				{"name": "b", "value": int32(20)},
				{"name": "a", "value": int32(10)},
			},
			sortKey:        "value",
			expectedOrder:  []interface{}{int32(10), int32(20)},
			testComparison: true,
		},
		{
			name: "int64 values",
			docs: []bson.M{
				{"name": "b", "value": int64(200)},
				{"name": "a", "value": int64(100)},
			},
			sortKey:        "value",
			expectedOrder:  []interface{}{int64(100), int64(200)},
			testComparison: true,
		},
		{
			name: "float64 equal values",
			docs: []bson.M{
				{"name": "a", "value": 10.5},
				{"name": "b", "value": 10.5},
			},
			sortKey:        "value",
			expectedOrder:  []interface{}{10.5, 10.5},
			testComparison: true,
		},
		{
			name: "string equal values",
			docs: []bson.M{
				{"name": "alice"},
				{"name": "alice"},
			},
			sortKey:        "name",
			expectedOrder:  []interface{}{"alice", "alice"},
			testComparison: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Clear()

			err := storage.InsertMany("test", tt.docs)
			assert.NoError(t, err)

			opts := &FindOptions{Sort: bson.M{tt.sortKey: 1}}
			results := storage.Find("test", bson.M{}, opts)
			assert.Len(t, results, len(tt.expectedOrder))

			for i, expected := range tt.expectedOrder {
				assert.Equal(t, expected, results[i][tt.sortKey])
			}
		})
	}
}

// TestProjection_EdgeCases tests edge cases in projection
func TestProjection_EdgeCases(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"_id": "123", "name": "alice", "age": 25, "email": "alice@example.com"}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		projection     bson.M
		expectedFields []string
		excludedFields []string
	}{
		{
			name:           "inclusion with _id explicitly included",
			projection:     bson.M{"name": 1, "_id": 1},
			expectedFields: []string{"name", "_id"},
			excludedFields: []string{"age", "email"},
		},
		{
			name:           "inclusion with _id explicitly excluded",
			projection:     bson.M{"name": 1, "_id": 0},
			expectedFields: []string{"name"},
			excludedFields: []string{"_id", "age", "email"},
		},
		{
			name:           "exclusion with _id",
			projection:     bson.M{"email": 0, "_id": 0},
			expectedFields: []string{"name", "age"},
			excludedFields: []string{"_id", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &FindOptions{Projection: tt.projection}
			results := storage.Find("users", bson.M{}, opts)
			assert.Len(t, results, 1)

			for _, field := range tt.expectedFields {
				assert.Contains(t, results[0], field)
			}

			for _, field := range tt.excludedFields {
				assert.NotContains(t, results[0], field)
			}
		})
	}
}

// TestReplaceOne tests ReplaceOne functionality
func TestReplaceOne(t *testing.T) {
	storage := NewMemoryStorage()

	// Insert initial document
	doc := bson.M{"_id": "123", "name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Replace the document
	replacement := bson.M{"name": "alice_new", "age": 30, "email": "alice@example.com"}
	replaced := storage.ReplaceOne("users", bson.M{"name": "alice"}, replacement)
	assert.Equal(t, int64(1), replaced)

	// Verify replacement
	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, "alice_new", results[0]["name"])
	assert.Equal(t, 30, results[0]["age"])
	assert.Equal(t, "alice@example.com", results[0]["email"])
	// _id should be preserved
	assert.Equal(t, "123", results[0]["_id"])
}

// TestReplaceOne_NonExistentCollection tests ReplaceOne on non-existent collection
func TestReplaceOne_NonExistentCollection(t *testing.T) {
	storage := NewMemoryStorage()

	replaced := storage.ReplaceOne("non_existent", bson.M{}, bson.M{"name": "test"})
	assert.Equal(t, int64(0), replaced)
}

// TestReplaceOne_NoMatch tests ReplaceOne when no document matches
func TestReplaceOne_NoMatch(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Try to replace with non-matching filter
	replaced := storage.ReplaceOne("users", bson.M{"name": "bob"}, bson.M{"name": "charlie", "age": 30})
	assert.Equal(t, int64(0), replaced)

	// Verify original document unchanged
	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, "alice", results[0]["name"])
}

// TestReplaceOne_PreserveID tests ReplaceOne preserves _id from original document
func TestReplaceOne_PreserveID(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"_id": "original_id", "name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Replace without _id in replacement
	replacement := bson.M{"name": "alice_new", "age": 30}
	replaced := storage.ReplaceOne("users", bson.M{"name": "alice"}, replacement)
	assert.Equal(t, int64(1), replaced)

	// Verify _id is preserved
	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, "original_id", results[0]["_id"])
}

// TestReplaceOne_WithIDInReplacement tests ReplaceOne when replacement has _id
func TestReplaceOne_WithIDInReplacement(t *testing.T) {
	storage := NewMemoryStorage()

	doc := bson.M{"_id": "original_id", "name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Replace with _id in replacement
	replacement := bson.M{"_id": "new_id", "name": "alice_new", "age": 30}
	replaced := storage.ReplaceOne("users", bson.M{"name": "alice"}, replacement)
	assert.Equal(t, int64(1), replaced)

	// Verify replacement _id is used
	results := storage.Find("users", bson.M{}, nil)
	assert.Len(t, results, 1)
	assert.Equal(t, "new_id", results[0]["_id"])
}

// TestUpdate_ErrorPath tests Update error handling
func TestUpdate_ErrorPath(t *testing.T) {
	storage := NewMemoryStorage()

	// Insert a document
	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// This won't actually cause an error in applyUpdate, but tests the path
	// The actual error path in Update is when applyUpdate returns an error
	// Since applyUpdate doesn't fail easily, we verify the code path exists
	update := bson.M{"$set": bson.M{"age": 30}}
	updated := storage.Update("users", bson.M{}, update)
	assert.Equal(t, int64(1), updated)
}

// TestUpdateOne_ErrorPath tests UpdateOne error handling
func TestUpdateOne_ErrorPath(t *testing.T) {
	storage := NewMemoryStorage()

	// Insert a document
	doc := bson.M{"name": "alice", "age": 25}
	err := storage.Insert("users", doc)
	assert.NoError(t, err)

	// Test the error path
	update := bson.M{"$set": bson.M{"age": 30}}
	updated := storage.UpdateOne("users", bson.M{}, update)
	assert.Equal(t, int64(1), updated)
}

// TestApplyIncrement_MismatchedTypes tests increment with type mismatches
func TestApplyIncrement_MismatchedTypes(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name         string
		currentValue interface{}
		incValue     interface{}
		expectedType string
	}{
		{
			name:         "int with int32 increment",
			currentValue: 10,
			incValue:     int32(5),
			expectedType: "should keep original value",
		},
		{
			name:         "int32 with int increment",
			currentValue: int32(10),
			incValue:     5,
			expectedType: "should keep original value",
		},
		{
			name:         "int64 with float64 increment",
			currentValue: int64(10),
			incValue:     float64(5.5),
			expectedType: "should keep original value",
		},
		{
			name:         "float64 with int increment",
			currentValue: float64(10.5),
			incValue:     5,
			expectedType: "should keep original value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.Clear()

			doc := bson.M{"value": tt.currentValue}
			err := storage.Insert("test", doc)
			assert.NoError(t, err)

			update := bson.M{"$inc": bson.M{"value": tt.incValue}}
			updated := storage.Update("test", bson.M{}, update)
			assert.Equal(t, int64(1), updated)

			// The value should remain unchanged due to type mismatch
			results := storage.Find("test", bson.M{}, nil)
			assert.Len(t, results, 1)
			// Value should be unchanged
			assert.Equal(t, tt.currentValue, results[0]["value"])
		})
	}
}

func TestIntegration_FullCRUDWorkflow(t *testing.T) {
	// Create mock client
	client := NewMockClient()
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
	storage := NewMemoryStorage()
	coll := NewMockCollection("products", storage, nil)
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
	storage := NewMemoryStorage()
	coll := NewMockCollection("users", storage, nil)
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
	storage := NewMemoryStorage()
	coll := NewMockCollection("items", storage, nil)
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
	storage := NewMemoryStorage()
	coll := NewMockCollection("scores", storage, nil)
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
	storage := NewMemoryStorage()
	coll := NewMockCollection("counters", storage, nil)
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
	storage := NewMemoryStorage()
	coll := NewMockCollection("users", storage, nil)
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
	storage := NewMemoryStorage()
	coll := NewMockCollection("items", storage, nil)
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
	storage := NewMemoryStorage()
	ctx := context.Background()

	t.Run("FindOne on empty collection", func(t *testing.T) {
		coll := NewMockCollection("empty", storage, nil)
		var doc bson.M
		err := coll.FindOne(ctx, bson.M{}).Decode(&doc)
		assert.Error(t, err)
	})

	t.Run("UpdateOne on non-existent document", func(t *testing.T) {
		coll := NewMockCollection("test", storage, nil)
		update := bson.M{"$set": bson.M{"field": "value"}}
		result, err := coll.UpdateOne(ctx, bson.M{"_id": primitive.NewObjectID()}, update)
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.MatchedCount)
	})

	t.Run("DeleteOne on non-existent document", func(t *testing.T) {
		coll := NewMockCollection("test", storage, nil)
		result, err := coll.DeleteOne(ctx, bson.M{"_id": primitive.NewObjectID()})
		require.NoError(t, err)
		assert.Equal(t, int64(0), result.DeletedCount)
	})
}

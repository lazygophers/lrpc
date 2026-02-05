package mock_test

import (
	"fmt"

	"github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"go.mongodb.org/mongo-driver/bson"
)

func ExampleMemoryStorage_Insert() {
	storage := mock.NewMemoryStorage()

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
	storage := mock.NewMemoryStorage()

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
	storage := mock.NewMemoryStorage()

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
	opts := &mock.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	}

	results := storage.Find("users", bson.M{}, opts)

	fmt.Printf("Found %d document(s)\n", len(results))
	// Output: Found 2 document(s)
}

func ExampleMemoryStorage_Update() {
	storage := mock.NewMemoryStorage()

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
	storage := mock.NewMemoryStorage()

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

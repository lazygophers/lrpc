package mock_test

import (
	"context"
	"fmt"

	"github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"go.mongodb.org/mongo-driver/bson"
)

// ExampleNewMockClient demonstrates basic usage of MockClient
func ExampleNewMockClient() {
	// Create a new mock client
	client := mock.NewMockClient()

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
	client := mock.NewMockClient()
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
	client := mock.NewMockClient()
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
	mockClient := client.(*mock.MockClient)
	mockClient.Clear()

	// Check count after cleanup
	count, _ = coll.CountDocuments(ctx, bson.M{})
	fmt.Printf("After cleanup: %d items\n", count)

	// Output:
	// Before cleanup: 2 items
	// After cleanup: 0 items
}

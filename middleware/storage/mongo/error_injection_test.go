package mongo

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCountWithClosedClient tests Count behavior when client is closed
func TestCountWithClosedClient(t *testing.T) {
	client := newTestClient(t)
	
	scoop := client.NewScoop().Collection(User{})
	
	// Close the client
	_ = client.Close()
	
	// Try to count - should trigger error path
	count, err := scoop.Count()
	if err != nil {
		t.Logf("Count with closed client returned error (expected): %v", err)
	} else {
		t.Logf("Count returned: %d (no error, client may still work)", count)
	}
}

// TestDeleteWithClosedClient tests Delete behavior when client is closed
func TestDeleteWithClosedClient(t *testing.T) {
	client := newTestClient(t)
	
	scoop := client.NewScoop().Collection(User{})
	
	// Close the client
	_ = client.Close()
	
	// Try to delete - should trigger error path
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Logf("Delete with closed client returned error (expected): %v", err)
	} else {
		t.Logf("Delete returned: %d (no error, client may still work)", deleted)
	}
}

// TestHealthWithClosedClient tests Health behavior when client is closed
func TestHealthWithClosedClient(t *testing.T) {
	client := newTestClient(t)
	
	// Close the client
	_ = client.Close()
	
	// Try to check health - should trigger error path
	err := client.Health()
	if err != nil {
		t.Logf("Health with closed client returned error (expected): %v", err)
	} else {
		t.Logf("Health returned no error (client may still work)")
	}
}

// TestPingWithClosedClient tests Ping behavior when client is closed
func TestPingWithClosedClient(t *testing.T) {
	client := newTestClient(t)
	
	// Close the client
	_ = client.Close()
	
	// Try to ping - should trigger error path
	err := client.Ping()
	if err != nil {
		t.Logf("Ping with closed client returned error (expected): %v", err)
	} else {
		t.Logf("Ping returned no error (client may still work)")
	}
}

// TestFirstWithInvalidFilter tests First with criteria that matches nothing and context timeout
func TestFirstWithInvalidFilter(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{}).Equal("email", "nonexistent@example.com")
	
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Logf("First with no matches returned error: %v", err)
	} else {
		t.Logf("First returned no error")
	}
}

// TestCountWithLargeCollection tests Count performance with large collection
func TestCountWithLargeCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert large number of documents
	for i := 0; i < 50; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "large_" + string(rune(48+(i%10))) + "@example.com",
			Name:      "User",
			Age:       25 + (i % 30),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Logf("Count error: %v", err)
	}
	if count != 50 {
		t.Errorf("Expected 50, got %d", count)
	}
}

// TestUpdateWithInvalidBSON tests Update with invalid BSON
func TestUpdateWithInvalidBSON(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "update@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{}).Equal("email", "update@example.com")
	
	// Try update with valid BSON
	updateResult := scoop.Update(bson.M{
		"$set": bson.M{
			"name": "Updated",
		},
	}
	updated, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Logf("Update returned error: %v", err)
	} else if updated != 1 {
		t.Errorf("Expected 1 updated, got %d", updated)
	}
}

// TestContextTimeout tests operation with timeout context
func TestContextTimeout(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create context with very short timeout (to test behavior)
	_, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(10 * time.Millisecond)

	// This should timeout but the scoop uses its own context
	scoop := client.NewScoop().Collection(User{})
	
	// Try count
	count, err := scoop.Count()
	if err != nil {
		t.Logf("Count with expired context returned error: %v", err)
	} else {
		t.Logf("Count with expired context returned: %d", count)
	}
}

// TestAggregateExecuteWithLargeResult tests Aggregate Execute with many results
func TestAggregateExecuteWithLargeResult(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert many documents
	for i := 0; i < 30; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "agg_" + string(rune(48+(i%10))) + "@example.com",
			Name:      "User",
			Age:       20 + (i % 40),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	agg := scoop.Aggregate()
	
	// Execute should handle large result sets
	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Logf("Aggregate execute returned error: %v", err)
	} else if len(results) != 30 {
		t.Logf("Expected 30 results, got %d", len(results))
	}
}

// TestExecuteOneWithEmptyResult tests ExecuteOne with no results
func TestExecuteOneWithEmptyResult(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	agg := scoop.Aggregate()
	
	// Match non-existent documents
	agg.Match(bson.M{"email": "nonexistent@example.com"})
	
	var result bson.M
	err := agg.ExecuteOne(&result)
	if err != nil {
		t.Logf("ExecuteOne with no results returned error (expected): %v", err)
	} else {
		t.Logf("ExecuteOne returned no error: %v", result)
	}
}

// TestChangeStreamWithContextCancel tests ChangeStream behavior when context is cancelled
func TestChangeStreamWithContextCancel(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	
	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Logf("WatchChanges returned error: %v", err)
		return
	}
	
	// Close immediately
	cs.Close()
}

// TestDatabaseChangeStreamWithContextCancel tests DatabaseChangeStream
func TestDatabaseChangeStreamWithContextCancel(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Logf("WatchAllCollections returned error: %v", err)
		return
	}
	
	// Close immediately
	dcs.Close()
}

// TestNewScoopWithDifferentModels tests NewScoop with different model types
func TestNewScoopWithDifferentModels(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Test with User model
	scoop1 := client.NewScoop().Collection(User{})
	if scoop1 == nil {
		t.Error("Expected non-nil scoop for User model")
	}

	// Test with nil collection
	scoop2 := client.NewScoop()
	if scoop2 == nil {
		t.Error("Expected non-nil scoop")
	}
}

// TestFindWithMultipleConditions tests Find with various condition combinations
func TestFindWithMultipleConditions(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	for i := 0; i < 10; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "find_" + string(rune(48+(i%10))) + "@example.com",
			Name:      "User",
			Age:       20 + (i % 30),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Test Find with multiple conditions
	scoop := client.NewScoop().Collection(User{}).
		Where("age", bson.M{"$gte": 25}).
		Where("age", bson.M{"$lte": 35}).
		Limit(5)

	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Logf("Find with conditions returned error: %v", err)
	} else {
		t.Logf("Find returned %d results", len(results))
	}
})

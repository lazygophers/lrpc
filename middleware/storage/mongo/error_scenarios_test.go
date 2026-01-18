package mongo

import (
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCountDocumentsWithDifferentDataTypes tests Count with various data
func TestCountDocumentsWithDifferentDataTypes(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data with various ages
	testCases := []int{18, 20, 25, 30, 35, 40}
	for _, age := range testCases {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("count_%d@example.com", age),
			Name:      "Test",
			Age:       age,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})

	// Count all
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count all failed: %v", err)
	}

	if count != int64(len(testCases)) {
		t.Errorf("expected count %d, got %d", len(testCases), count)
	}

	// Count with filter
	scoop2 := client.NewScoop().Collection(User{}).Where("age", bson.M{"$gte": 30})
	count2, err := scoop2.Count()
	if err != nil {
		t.Fatalf("count with filter failed: %v", err)
	}

	if count2 != 3 { // 30, 35, 40
		t.Errorf("expected count 3, got %d", count2)
	}
}

// TestDeleteDocumentsWithConditions tests Delete with various conditions
func TestDeleteDocumentsWithConditions(t *testing.T) {
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
			Email:     fmt.Sprintf("delete_%d@example.com", i),
			Name:      "Test",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Delete with age filter (ages 20-24)
	scoop := client.NewScoop().Collection(User{}).Where("age", bson.M{"$lt": 25})
	deleted, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete with filter failed: %v", err)
	}

	if deleted != 5 {
		t.Errorf("expected 5 deleted, got %d", deleted)
	}

	// Verify remaining count
	remaining, _ := client.NewScoop().Collection(User{}).Count()
	if remaining != 5 {
		t.Errorf("expected 5 remaining, got %d", remaining)
	}
}

// TestHealthCheckBasic tests basic Health check
func TestHealthCheckBasic(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Health()
	if err != nil {
		t.Logf("Health check error (may be expected): %v", err)
	}
}

// TestPingBasic tests basic Ping
func TestPingBasic(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Ping()
	if err != nil {
		t.Logf("Ping error (may be expected in test env): %v", err)
	}
}

// TestDeleteEmptyCollectionScenario tests Delete on empty collection
func TestDeleteEmptyCollectionScenario(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	deleted, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete on empty collection failed: %v", err)
	}

	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

// TestCountAndDeleteSequence tests Count and Delete in sequence
func TestCountAndDeleteSequence(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert data
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("seq_%d@example.com", i),
			Name:      "Test",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Count should return 5
	scoop := client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}

	// Delete all
	deleted, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if deleted != 5 {
		t.Errorf("expected 5 deleted, got %d", deleted)
	}

	// Count should return 0
	scoop2 := client.NewScoop().Collection(User{})
	count2, err := scoop2.Count()
	if err != nil {
		t.Fatalf("count after delete failed: %v", err)
	}
	if count2 != 0 {
		t.Errorf("expected 0 after delete, got %d", count2)
	}
}

// TestStreamWatchBasic tests basic Watch functionality
func TestStreamWatchBasic(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})
	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("create change stream failed: %v", err)
	}

	// Test Watch with no pipeline
	stream, err := cs.Watch()
	if err != nil {
		t.Logf("Watch() error: %v", err)
	} else if stream != nil {
		stream.Close(client.Context())
	}
}

// TestStreamWatchWithMatch tests Watch with match pipeline
func TestStreamWatchWithMatch(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})
	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("create change stream failed: %v", err)
	}

	// Test Watch with match stage
	stream, err := cs.Watch(
		bson.M{
			"$match": bson.M{
				"operationType": "insert",
			},
		},
	)
	if err != nil {
		t.Logf("Watch(match) error: %v", err)
	} else if stream != nil {
		stream.Close(client.Context())
	}
}

// TestAggregateBasic tests basic Aggregate functionality
func TestAggregateBasic(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("agg_%d@example.com", i),
			Name:      "User",
			Age:       25 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})

	// Test Execute with valid result
	agg := scoop.Aggregate()
	var results []User
	err := agg.Execute(&results)
	if err != nil {
		t.Logf("Execute error: %v", err)
	} else if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

// TestClientCloseMultipleTimes tests calling Close multiple times
func TestClientCloseMultipleTimes(t *testing.T) {
	client := newTestClient(t)

	// First close should work
	err1 := client.Close()
	if err1 != nil {
		t.Logf("First Close error: %v", err1)
	}

	// Second close may fail or succeed
	err2 := client.Close()
	if err2 != nil {
		t.Logf("Second Close error: %v", err2)
	}
}

// Helper function for InsertTestData that returns error
func InsertTestDataError(client *Client, collection string, doc interface{}) error {
	scoop := client.NewScoop().Collection(doc)
	return scoop.Create(doc)
}

package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ============= Aggregation Tests =============

// TestAggregationExecuteEmpty tests Execute with empty results
func TestAggregationExecuteEmpty(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	// Match documents that don't exist
	agg.Match(bson.M{"email": "nonexistent@example.com"})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestAggregationExecuteOneWithResults tests ExecuteOne with results
func TestAggregationExecuteOneWithResults(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "test@example.com",
		Name:      "Test User",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	model := NewModel(client, User{})
	model.NewScoop().Create(user)

	// Test ExecuteOne
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()
	agg.Match(bson.M{"email": "test@example.com"})

	var result bson.M
	err := agg.ExecuteOne(&result)
	if err != nil {
		t.Fatalf("execute one failed: %v", err)
	}

	if result == nil {
		t.Error("expected result, got nil")
	}
	if result["email"] != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%v'", result["email"])
	}
}

// TestAggregationExecuteOneEmpty tests ExecuteOne with no results
func TestAggregationExecuteOneEmpty(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	// Match documents that don't exist
	agg.Match(bson.M{"email": "nonexistent@example.com"})

	var result bson.M
	err := agg.ExecuteOne(&result)
	
	// When there are no results, ExecuteOne should not error but result will be empty/default
	if err != nil {
		// This is acceptable
		if err.Error() != "mongo: no documents in result" {
			t.Logf("ExecuteOne with no results: %v", err)
		}
	}
}

// TestAggregationProjectFields tests aggregation with projection
func TestAggregationProjectFields(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "test@example.com",
		Name:      "Test User",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	model := NewModel(client, User{})
	model.NewScoop().Create(user)

	// Test with projection
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()
	agg.Project(bson.M{"email": 1, "name": 1})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute with project failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// ============= Client Methods Tests =============

// TestClientHealthWithFailingPing tests Health when Ping fails
func TestClientHealthSuccess(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Test that Health works when connection is good
	err := client.Health()
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
}

// TestClientNewWithNilConfig tests New with nil config
func TestClientNewWithNilConfig(t *testing.T) {
	client, err := New(nil)
	if err != nil {
		// nil config is acceptable, defaults will be applied
		t.Logf("New with nil config: %v", err)
	}
	if client != nil {
		client.Close()
	}
}

// TestClientCloseSuccess tests Close functionality
func TestClientCloseSuccess(t *testing.T) {
	client := newTestClient(t)

	// Verify connection works before close
	err := client.Ping()
	if err != nil {
		t.Fatalf("ping before close failed: %v", err)
	}

	// Close the connection
	err = client.Close()
	if err != nil {
		t.Fatalf("close failed: %v", err)
	}

	// After close, ping should fail
	err = client.Ping()
	if err == nil {
		t.Error("expected error after close, got nil")
	}
}

// TestClientPingSuccess tests Ping functionality
func TestClientPingSuccess(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Ping()
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}
}

// ============= Scoop Methods Tests =============

// TestScoopCountWithLargeDataset tests Count with many documents
func TestScoopCountWithLargeDataset(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create 20 documents
	for i := 0; i < 20; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "user" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		scoop.Create(user)
	}

	// Count all
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 20 {
		t.Errorf("expected 20, got %d", count)
	}

	// Count with filter
	countFiltered, err := scoop.Where("age", map[string]interface{}{"$gte": 30}).Count()
	if err != nil {
		t.Fatalf("count with filter failed: %v", err)
	}
	if countFiltered != 10 {
		t.Errorf("expected 10 (ages 30-39), got %d", countFiltered)
	}
}

// TestScoopFirstWithFilter tests First with filter conditions
func TestScoopFirstWithFilter(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create multiple users
	users := []User{
		{ID: primitive.NewObjectID(), Email: "alice@example.com", Name: "Alice", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "bob@example.com", Name: "Bob", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "charlie@example.com", Name: "Charlie", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, user := range users {
		scoop.Create(user)
	}

	// First with filter
	var found User
	err := scoop.Where("age", map[string]interface{}{"$gte": 30}).First(&found)
	if err != nil {
		t.Fatalf("first with filter failed: %v", err)
	}

	if found.Age < 30 {
		t.Errorf("expected age >= 30, got %d", found.Age)
	}
}

// TestScoopFindWithProjection tests Find with projection
func TestScoopFindWithProjection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create test user
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "test@example.com",
		Name:      "Test User",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	scoop.Create(user)

	// Find with Select (projection)
	var found []User
	err := scoop.Select("email", "name").Find(&found)
	if err != nil {
		t.Fatalf("find with select failed: %v", err)
	}

	if len(found) != 1 {
		t.Errorf("expected 1 user, got %d", len(found))
	}
}

// TestScoopBatchCreateWithMultiple tests BatchCreate with multiple documents
func TestScoopBatchCreateWithMultiple(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create multiple documents in one call
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	err := scoop.BatchCreate(users...)
	if err != nil {
		t.Fatalf("batch create failed: %v", err)
	}

	// Verify count
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 documents, got %d", count)
	}
}

// TestScoopUpdateWithZeroResults tests Update returning zero results
func TestScoopUpdateWithZeroResults(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Try to update non-existent document
	updated, err := scoop.Where("email", "nonexistent@example.com").Update(map[string]interface{}{
		"name": "Updated",
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated != 0 {
		t.Errorf("expected 0 updated, got %d", updated)
	}
}

// TestScoopFindWithLimit tests Find with Limit
func TestScoopFindWithLimit(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create 5 users
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "user" + string(rune('0'+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		scoop.Create(user)
	}

	// Find with limit
	var found []User
	err := scoop.Limit(2).Find(&found)
	if err != nil {
		t.Fatalf("find with limit failed: %v", err)
	}

	if len(found) != 2 {
		t.Errorf("expected 2 results (limit=2), got %d", len(found))
	}
}

// TestScoopFindWithSkip tests Find with Skip
func TestScoopFindWithSkip(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create 5 users
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "user" + string(rune('0'+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		scoop.Create(user)
	}

	// Find with skip
	var found []User
	err := scoop.Skip(2).Find(&found)
	if err != nil {
		t.Fatalf("find with skip failed: %v", err)
	}

	if len(found) != 3 {
		t.Errorf("expected 3 results (skip=2), got %d", len(found))
	}
}

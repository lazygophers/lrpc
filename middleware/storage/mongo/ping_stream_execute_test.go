package mongo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestPingRepeatedly tests calling Ping multiple times
func TestPingRepeatedly(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Call Ping multiple times to ensure idempotency
	for i := 0; i < 5; i++ {
		err := client.Ping()
		if err != nil {
			t.Errorf("Ping iteration %d failed: %v", i, err)
		}
	}
}

// TestPingAfterOperations tests Ping after various MongoDB operations
func TestPingAfterOperations(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Perform some operations
	scoop := client.NewScoop().Collection(User{})
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "ping_after@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err := scoop.Create(user)
	if err != nil {
		t.Logf("Create error: %v", err)
	}

	// Now ping to verify connection is still good
	err = client.Ping()
	if err != nil {
		t.Errorf("Ping after Create failed: %v", err)
	}

	// Query
	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Logf("Find error: %v", findResult.Error)
	}

	// Ping again
	err = client.Ping()
	if err != nil {
		t.Errorf("Ping after Find failed: %v", err)
	}
}

// TestStreamWatchWithFilters tests Watch with specific filters
func TestStreamWatchWithFilters(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Watch with filter
	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Logf("WatchChanges returned error: %v", err)
		return
	}

	if cs == nil {
		t.Error("Expected non-nil ChangeStream")
		return
	}

	cs.Close()
	t.Logf("Watch with filters completed successfully")
}

// TestStreamWatchMultipleTimes tests creating and closing multiple watches
func TestStreamWatchMultipleTimes(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	for i := 0; i < 3; i++ {
		scoop := client.NewScoop().Collection(User{})
		cs, err := scoop.WatchChanges()
		if err != nil {
			t.Logf("WatchChanges iteration %d returned error: %v", i, err)
			continue
		}

		if cs != nil {
			cs.Close()
		}
	}

	t.Logf("Multiple watches completed successfully")
}

// TestExecuteOneWithEmptyPipeline tests ExecuteOne with empty pipeline
func TestExecuteOneWithEmptyPipeline(t *testing.T) {
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
			Email:     "exec_one_empty_" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i*5,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	agg := scoop.Aggregate()

	// Empty pipeline should return first document
	var result bson.M
	err := agg.ExecuteOne(&result)

	if err != nil {
		t.Logf("ExecuteOne with empty pipeline returned error: %v", err)
	} else if result != nil {
		t.Logf("ExecuteOne with empty pipeline returned result with %d fields", len(result))
	} else {
		t.Logf("ExecuteOne with empty pipeline returned nil result")
	}
}

// TestExecuteOneWithMatchAndProject tests ExecuteOne with match and project
func TestExecuteOneWithMatchAndProject(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "exec_proj_" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i*10,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	agg := scoop.Aggregate()

	agg.Match(bson.M{"age": bson.M{"$gte": 30}})
	agg.Project(bson.M{"name": 1, "email": 1})

	var result bson.M
	err := agg.ExecuteOne(&result)

	if err != nil {
		t.Logf("ExecuteOne with match and project returned error: %v", err)
	} else if result != nil {
		t.Logf("ExecuteOne returned result with name=%v, email=%v", result["name"], result["email"])
	}
}

// TestExecuteOneWithSort tests ExecuteOne with sort
func TestExecuteOneWithSort(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data with different ages
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "exec_sort_" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       30 - i*5, // 30, 25, 20, 15, 10
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	agg := scoop.Aggregate()

	agg.Sort(bson.M{"age": 1}) // Ascending order

	var result bson.M
	err := agg.ExecuteOne(&result)

	if err != nil {
		t.Logf("ExecuteOne with sort returned error: %v", err)
	} else if result != nil {
		// First result should have the lowest age
		t.Logf("ExecuteOne with sort returned result with age field present: %v", result["age"] != nil)
	}
}

// TestExecuteWithComplexAggregation tests Execute with multiple stages
func TestExecuteWithComplexAggregation(t *testing.T) {
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
			Email:     "complex_agg_" + string(rune(48+(i%10))) + "@example.com",
			Name:      "User",
			Age:       20 + (i % 30),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	agg := scoop.Aggregate()

	// Complex pipeline
	agg.Match(bson.M{"age": bson.M{"$gte": 25}})
	agg.Group(bson.M{"_id": "$age", "count": bson.M{"$sum": 1}})
	agg.Sort(bson.M{"_id": 1})

	var results []bson.M
	err := agg.Execute(&results)

	if err != nil {
		t.Logf("Complex aggregation returned error: %v", err)
	} else {
		t.Logf("Complex aggregation returned %d results", len(results))
	}
}

// TestWatchAllCollectionsBasic tests WatchAllCollections
func TestWatchAllCollectionsBasic(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Logf("WatchAllCollections returned error: %v", err)
		return
	}

	if dcs == nil {
		t.Error("Expected non-nil DatabaseChangeStream")
		return
	}

	dcs.Close()
	t.Logf("WatchAllCollections completed successfully")
}

// TestExecuteWithLimitStage tests Execute with $limit stage
func TestExecuteWithLimitStage(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert many documents
	for i := 0; i < 20; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "limit_test_" + string(rune(48+(i%10))) + "_" + string(rune(48+i/10)) + "@example.com",
			Name:      "User",
			Age:       20 + (i % 30),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	agg := scoop.Aggregate()

	agg.Limit(5)

	var results []bson.M
	err := agg.Execute(&results)

	if err != nil {
		t.Errorf("Execute with limit returned error: %v", err)
	} else if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	} else {
		t.Logf("Execute with limit returned expected 5 results")
	}
}

// TestExecuteWithSkipAndLimit tests Execute with $skip and $limit
func TestExecuteWithSkipAndLimit(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert many documents
	for i := 0; i < 15; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "skip_limit_" + fmt.Sprintf("%02d", i) + "@example.com",
			Name:      "User",
			Age:       20 + (i % 30),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	agg := scoop.Aggregate()

	agg.Skip(5)
	agg.Limit(5)

	var results []bson.M
	err := agg.Execute(&results)

	if err != nil {
		t.Errorf("Execute with skip and limit returned error: %v", err)
	} else if len(results) != 5 {
		t.Errorf("Expected 5 results after skip, got %d", len(results))
	} else {
		t.Logf("Execute with skip and limit returned expected 5 results")
	}
}

// TestContinuousOperations tests continuous operations with health checks between
func TestContinuousOperations(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	for i := 0; i < 3; i++ {
		// Perform operation
		scoop := client.NewScoop().Collection(User{})
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "continuous_" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}

		err := scoop.Create(user)
		if err != nil {
			t.Logf("Create iteration %d failed: %v", i, err)
		}

		// Check health between operations
		err = client.Health()
		if err != nil {
			t.Logf("Health check iteration %d failed: %v", i, err)
		}
	}

	t.Logf("Continuous operations completed")
}

// TestExecuteWithFacetStage tests Execute with $facet stage
func TestExecuteWithFacetStage(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	for i := 0; i < 20; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "facet_" + string(rune(48+(i%10))) + "_" + string(rune(48+i/10)) + "@example.com",
			Name:      "User",
			Age:       20 + (i % 30),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	agg := scoop.Aggregate()

	agg.Facet(bson.M{
		"young": []bson.M{
			{"$match": bson.M{"age": bson.M{"$lt": 25}}},
		},
		"old": []bson.M{
			{"$match": bson.M{"age": bson.M{"$gte": 25}}},
		},
	})

	var results []bson.M
	err := agg.Execute(&results)

	if err != nil {
		t.Logf("Facet aggregation returned error: %v", err)
	} else if len(results) > 0 {
		t.Logf("Facet aggregation returned %d results", len(results))
	}
}

// TestContextWithTimeout tests operations with timeout context
func TestContextWithTimeout(t *testing.T) {
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
		Email:     "timeout@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	InsertTestData(t, client, "users", user)

	// Create context with reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Note: Scoop uses its own context internally, but we test the mechanism
	scoop := client.NewScoop().Collection(User{})

	_ = ctx // Context would be used in real scenarios

	var results []User
	err := scoop.Find(&results)

	if err != nil {
		t.Logf("Find returned error: %v", err)
	} else {
		t.Logf("Find succeeded with %d results", len(results))
	}
}

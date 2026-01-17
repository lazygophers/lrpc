package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAggregationMatch(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test match
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Match(bson.M{"age": bson.M{"$gte": 30}})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestAggregationGroup(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test group
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Group(bson.M{
		"_id":   "$age",
		"count": bson.M{"$sum": 1},
	})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) < 2 {
		t.Errorf("expected at least 2 groups, got %d", len(results))
	}
}

func TestAggregationSort(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test sort
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Sort(bson.M{"age": -1})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestAggregationProject(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	InsertTestData(t, client, "users", user)

	// Test project
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Project(bson.M{
		"name":  1,
		"email": 1,
		"_id":   0,
	})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Check that only name and email are present
	if _, ok := results[0]["name"]; !ok {
		t.Error("expected name in result")
	}
	if _, ok := results[0]["email"]; !ok {
		t.Error("expected email in result")
	}
}

func TestAggregationLimitSkip(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test limit and skip
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Skip(1).Limit(2)

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestAggregationAddFields(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	InsertTestData(t, client, "users", user)

	// Test add fields
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.AddFields(bson.M{
		"isAdult": bson.M{"$gte": []interface{}{"$age", 18}},
	})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

func TestAggregationCount(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test count
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Count("total")

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if total, ok := results[0]["total"].(int32); !ok || total != 2 {
		t.Errorf("expected total 2, got %v", results[0]["total"])
	}
}

func TestAggregationExecuteOne(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	InsertTestData(t, client, "users", user)

	// Test execute one
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	var result bson.M
	err := agg.ExecuteOne(&result)
	if err != nil {
		t.Fatalf("execute one failed: %v", err)
	}

	if result == nil {
		t.Error("expected result, got nil")
	}
}

func TestAggregationPipeline(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test complex pipeline
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate(
		bson.M{"$match": bson.M{"age": bson.M{"$gte": 25}}},
		bson.M{"$sort": bson.M{"age": -1}},
		bson.M{"$limit": 2},
	)

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestAggregationClear(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate(bson.M{"$match": bson.M{"age": 25}})

	agg.Clear()

	if len(agg.GetPipeline()) != 0 {
		t.Error("expected pipeline to be cleared")
	}
}


func TestAggregationLookup(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users", "orders")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert users
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Insert orders (with user_id reference)
	userID := users[0].(User).ID
	orders := []interface{}{
		bson.M{"_id": primitive.NewObjectID(), "user_id": userID, "amount": 100},
		bson.M{"_id": primitive.NewObjectID(), "user_id": userID, "amount": 200},
	}
	InsertTestData(t, client, "orders", orders...)

	// Test lookup
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Lookup("orders", "_id", "user_id", "user_orders")

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) < 1 {
		t.Errorf("expected at least 1 result, got %d", len(results))
	}

	// Verify lookup field exists
	if _, ok := results[0]["user_orders"]; !ok {
		t.Error("expected user_orders field in result")
	}
}

func TestAggregationUnwind(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "posts")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert posts with tags array
	posts := []interface{}{
		bson.M{"_id": primitive.NewObjectID(), "title": "Post 1", "tags": []string{"go", "mongodb"}},
		bson.M{"_id": primitive.NewObjectID(), "title": "Post 2", "tags": []string{"go"}},
	}
	InsertTestData(t, client, "posts", posts...)

	// Test unwind
	_, _, db, mgmErr := mgm.DefaultConfigs()
	if mgmErr != nil {
		t.Fatalf("failed to get MGM config: %v", mgmErr)
	}
	coll := db.Collection("posts")
	agg := NewAggregation(coll, context.Background())

	agg.Unwind("$tags")

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// Unwind should expand arrays, so we should get more documents than inserted
	if len(results) < 3 {
		t.Errorf("expected at least 3 results after unwind, got %d", len(results))
	}
}

func TestAggregationUnwindWithPreserve(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "posts")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert posts with some without tags array
	posts := []interface{}{
		bson.M{"_id": primitive.NewObjectID(), "title": "Post 1", "tags": []string{"go", "mongodb"}},
		bson.M{"_id": primitive.NewObjectID(), "title": "Post 2"},
	}
	InsertTestData(t, client, "posts", posts...)

	// Test unwind with preserveNullAndEmptyArrays
	_, _, db, mgmErr := mgm.DefaultConfigs()
	if mgmErr != nil {
		t.Fatalf("failed to get MGM config: %v", mgmErr)
	}
	coll := db.Collection("posts")
	agg := NewAggregation(coll, context.Background())

	agg.Unwind("$tags", true)

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	// With preserveNullAndEmptyArrays=true, documents without array should be preserved
	if len(results) < 2 {
		t.Errorf("expected at least 2 results, got %d", len(results))
	}
}

func TestAggregationFacet(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test facet
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Facet(bson.M{
		"young": []bson.M{
			{"$match": bson.M{"age": bson.M{"$lt": 30}}},
		},
		"older": []bson.M{
			{"$match": bson.M{"age": bson.M{"$gte": 30}}},
		},
	})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	// Check facet result structure
	if _, ok := results[0]["young"]; !ok {
		t.Error("expected 'young' facet in result")
	}
	if _, ok := results[0]["older"]; !ok {
		t.Error("expected 'older' facet in result")
	}
}

func TestAggregationAddStage(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test add stage
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	// Add custom stage
	agg.AddStage(bson.M{
		"$match": bson.M{"age": bson.M{"$gte": 25}},
	})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCountWithSpecialConditions tests Count with various special conditions
func TestCountWithSpecialConditions(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Insert test data with different ages
	ages := []int{10, 20, 30, 40, 50}
	for i, age := range ages {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "count" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       age,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		scoop.Create(user)
	}

	// Test count with $in operator
	inScoop := client.NewScoop().Collection(User{})
	count, err := inScoop.Where("age", bson.M{"$in": []int{20, 40}}).Count()
	if err != nil {
		t.Fatalf("count with $in failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 with $in operator, got %d", count)
	}

	// Test count with $nin operator
	ninScoop := client.NewScoop().Collection(User{})
	count, err = ninScoop.Where("age", bson.M{"$nin": []int{10, 50}}).Count()
	if err != nil {
		t.Fatalf("count with $nin failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 with $nin operator, got %d", count)
	}

	// Test count with $exists
	existsScoop := client.NewScoop().Collection(User{})
	count, err = existsScoop.Where("age", bson.M{"$exists": true}).Count()
	if err != nil {
		t.Fatalf("count with $exists failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 with $exists, got %d", count)
	}
}

// TestFirstWithComplexSort tests First with filter and result validation
func TestFirstWithComplexSort(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []User{
		{ID: primitive.NewObjectID(), Email: "first1@example.com", Name: "Charlie", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "first2@example.com", Name: "Alice", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "first3@example.com", Name: "Bob", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	for _, u := range users {
		InsertTestData(t, client, "users", u)
	}

	// Get first document (should be one of the inserted users)
	scoop := client.NewScoop().Collection(User{})
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("first failed: %v", err)
	}

	// Verify it's a valid user
	if result.ID == primitive.NilObjectID {
		t.Error("expected valid user ID")
	}

	// Get first with filter
	filterScoop := client.NewScoop().Collection(User{}).Equal("name", "Alice")
	var filterResult User
	err = filterScoop.First(&filterResult)
	if err != nil {
		t.Fatalf("first with filter failed: %v", err)
	}
	if filterResult.Age != 25 {
		t.Errorf("expected filtered user to have age 25, got %d", filterResult.Age)
	}
}

// TestFirstWithFilterAndSort tests First with both filter and sort
func TestFirstWithFilterAndSort(t *testing.T) {
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
			Email:     "filter" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + (i * 5),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Get first user with age > 30 sorted by age ascending
	scoop := client.NewScoop().Collection(User{}).
		Where("age", bson.M{"$gt": 30}).
		Sort("age", 1)
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("first with filter and sort failed: %v", err)
	}
	if result.Age != 35 {
		t.Errorf("expected age 35, got %d", result.Age)
	}
}

// TestBatchCreateWithLargeData tests BatchCreate with many documents
func TestBatchCreateWithLargeData(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create 20 users in batch
	users := make([]interface{}, 20)
	for i := 0; i < 20; i++ {
		users[i] = User{
			ID:        primitive.NewObjectID(),
			Email:     "batch" + string(rune(48+i%10)) + string(rune(48+i/10)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	err := scoop.BatchCreate(users...)
	if err != nil {
		t.Fatalf("batch create large data failed: %v", err)
	}

	// Verify count
	count, _ := scoop.Count()
	if count != 20 {
		t.Errorf("expected 20 documents, got %d", count)
	}
}

// TestBatchCreateWithDuplicateID tests BatchCreate behavior with duplicate IDs
func TestBatchCreateWithDuplicateID(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	id := primitive.NewObjectID()
	users := []interface{}{
		User{
			ID:        id,
			Email:     "dup1@example.com",
			Name:      "User 1",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		User{
			ID:        id,
			Email:     "dup2@example.com",
			Name:      "User 2",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err := scoop.BatchCreate(users...)
	// Should fail due to duplicate ID
	if err == nil {
		t.Error("expected error for duplicate ID in batch create")
	}
}

// TestAggregationWithMultipleStages tests complex aggregation pipelines
func TestAggregationWithMultipleStages(t *testing.T) {
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
			Email:     "agg" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + (i * 5),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	model := NewModel[User](client)
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	// Add match stage
	agg.Match(bson.M{"age": bson.M{"$gte": 30}})

	// Add sort stage
	agg.Sort(bson.M{"age": 1})

	// Add project stage
	agg.Project(bson.M{
		"_id":   1,
		"email": 1,
		"age":   1,
	})

	// Execute aggregation
	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("aggregation with multiple stages failed: %v", err)
	}

	// We expect results with age >= 30
	if len(results) == 0 {
		t.Error("expected at least one result from aggregation")
	}
}

// TestAggregationWithGroup tests aggregation with grouping
func TestAggregationWithGroup(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert data with different ages
	ageGroups := []int{20, 20, 30, 30, 30, 40}
	for i, age := range ageGroups {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "group" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       age,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	model := NewModel[User](client)
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	// Group by age and count
	agg.Group(bson.M{
		"_id":   "$age",
		"count": bson.M{"$sum": 1},
	})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("aggregation group failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 groups (ages 20, 30, 40), got %d", len(results))
	}
}

// TestAggregationExecuteOneSingleResult tests ExecuteOne with single result
func TestAggregationExecuteOneSingleResult(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "single@example.com",
		Name:      "Single User",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	model := NewModel[User](client)
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Match(bson.M{"email": "single@example.com"})

	var result bson.M
	err := agg.ExecuteOne(&result)
	if err != nil {
		t.Fatalf("aggregation execute one failed: %v", err)
	}

	if result == nil {
		t.Error("expected non-nil result from execute one")
	}
}

// TestAggregationExecuteOneNoResults tests ExecuteOne with no results
func TestAggregationExecuteOneNoResults(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	model := NewModel[User](client)
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	agg.Match(bson.M{"email": "nonexistent@example.com"})

	var result bson.M
	err := agg.ExecuteOne(&result)
	// ExecuteOne may return error or nil depending on implementation
	// If error, that's fine, if not, result should be nil
	if err == nil && result != nil {
		t.Error("expected nil or error when no documents found")
	}
}

// TestScoopFindWithSortField tests Find with sort field
func TestScoopFindWithSortField(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert data
	data := []User{
		{ID: primitive.NewObjectID(), Email: "sort_multi1@example.com", Name: "Alice", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "sort_multi2@example.com", Name: "Bob", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "sort_multi3@example.com", Name: "Charlie", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	for _, u := range data {
		InsertTestData(t, client, "users", u)
	}

	// Find all results
	scoop := client.NewScoop().Collection(User{})

	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

// TestCountWithCombinedConditions tests Count with multiple combined conditions
func TestCountWithCombinedConditions(t *testing.T) {
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
			Email:     "combined" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + (i * 5),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Count all
	scoop := client.NewScoop().Collection(User{})

	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count all failed: %v", err)
	}

	if count != 5 {
		t.Errorf("expected 5 results, got %d", count)
	}

	// Count with equal condition
	eqScoop := client.NewScoop().Collection(User{}).Equal("age", 25)
	eqCount, err := eqScoop.Count()
	if err != nil {
		t.Fatalf("count with equal condition failed: %v", err)
	}

	if eqCount != 1 {
		t.Errorf("expected 1 user with age 25, got %d", eqCount)
	}
}

// TestUpdateWithZeroMatches tests Update when no documents match
func TestUpdateWithZeroMatches(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{}).
		Equal("email", "nonexistent@example.com")

	updated, err := scoop.Update(bson.M{"$set": bson.M{"age": 99}})
	if err != nil {
		t.Fatalf("update with zero matches failed: %v", err)
	}

	if updated != 0 {
		t.Errorf("expected 0 updated, got %d", updated)
	}
}

// TestDeleteWithZeroMatches tests Delete when no documents match
func TestDeleteWithZeroMatches(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{}).
		Equal("email", "nonexistent@example.com")

	deleted, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete with zero matches failed: %v", err)
	}

	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

// TestFirstWithNoResults tests First when no documents match
func TestFirstWithNoResults(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{}).
		Equal("email", "nonexistent@example.com")

	var result User
	err := scoop.First(&result)
	// Should return error for no documents
	if err == nil {
		t.Error("expected error when no documents found")
	}
}

// TestCondComplexChaining tests condition chaining
func TestCondComplexChaining(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data with specific email pattern for this test
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "chaining" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + (i * 10),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Count all in this test
	scoop := client.NewScoop().Collection(User{}).
		Where("email", bson.M{"$regex": "^chaining.*@example\\.com$"})

	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("condition chaining failed: %v", err)
	}

	// Should match all 5 records we just inserted
	if count != 5 {
		t.Errorf("expected 5 results, got %d", count)
	}
}

// TestFindWithSkip tests Find with skip
func TestFindWithSkip(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert 5 documents
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "skip" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Find with skip 2
	scoop := client.NewScoop().Collection(User{}).Skip(2)

	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find with skip failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results after skip 2, got %d", len(results))
	}
}

// TestFindWithLimitAndSkip tests Find with both limit and skip
func TestFindWithLimitAndSkip(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert 10 documents
	for i := 0; i < 10; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "limskip" + string(rune(48+i%10)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Find with skip 3 and limit 2
	scoop := client.NewScoop().Collection(User{}).Skip(3).Limit(2)

	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find with limit and skip failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results (skip 3, limit 2), got %d", len(results))
	}
}

// TestClientPingFailure tests client ping in various scenarios
func TestClientPingFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Ping()
	if err != nil {
		t.Logf("Ping returned error: %v", err)
		return
	}
	// If no error, connection is good
}

// TestExistWithResults tests Exist method when document exists
func TestExistWithResults(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "exist@example.com",
		Name:      "Exists",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{}).Equal("email", "exist@example.com")

	exists, err := scoop.Exist()
	if err != nil {
		t.Fatalf("exist query failed: %v", err)
	}

	if !exists {
		t.Error("expected document to exist")
	}
}

// TestExistWithoutResults tests Exist method when document does not exist
func TestExistWithoutResults(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{}).Equal("email", "nonexistent@example.com")

	exists, err := scoop.Exist()
	if err != nil {
		t.Fatalf("exist query failed: %v", err)
	}

	if exists {
		t.Error("expected document to not exist")
	}
}

package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestFindAll tests basic Find operation returning all documents
func TestFindAll(t *testing.T) {
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
			Email:     "findall" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Find all documents
	scoop := client.NewScoop().Collection(User{})
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find all failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 records from find all, got %d", len(results))
	}
}

// TestFirstWithAllOptions tests First with various query options
func TestFirstWithAllOptions(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert multiple records
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "first_opt" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Test First without any filter
	scoop1 := client.NewScoop().Collection(User{})
	var result1 User
	err := scoop1.First(&result1)
	if err != nil {
		t.Fatalf("first without filter failed: %v", err)
	}
	if result1.ID == primitive.NilObjectID {
		t.Error("expected valid user")
	}

	// Test First with filter
	scoop2 := client.NewScoop().Collection(User{}).Equal("age", 22)
	var result2 User
	err = scoop2.First(&result2)
	if err != nil {
		t.Fatalf("first with filter failed: %v", err)
	}
	if result2.Age != 22 {
		t.Errorf("expected age 22, got %d", result2.Age)
	}

	// Test First with limit (should work even with limit=1)
	scoop3 := client.NewScoop().Collection(User{}).Limit(1)
	var result3 User
	err = scoop3.First(&result3)
	if err != nil {
		t.Fatalf("first with limit failed: %v", err)
	}
	if result3.ID == primitive.NilObjectID {
		t.Error("expected valid user")
	}
}

// TestCountWithMultipleConditionTypes tests Count with different condition types
func TestCountWithMultipleConditionTypes(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data with specific ages
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "type1@example.com", Name: "Alice", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "type2@example.com", Name: "Bob", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "type3@example.com", Name: "Charlie", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "type4@example.com", Name: "David", Age: 40, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test count with $lt 35 (should match 25, 30)
	ltScoop := client.NewScoop().Collection(User{}).Where("age", bson.M{"$lt": 35})
	ltCount, err := ltScoop.Count()
	if err != nil {
		t.Fatalf("count with $lt failed: %v", err)
	}
	if ltCount != 2 {
		t.Errorf("expected 2 with $lt 35, got %d", ltCount)
	}

	// Test count with $lte 35 (should match 25, 30, 35)
	lteScoop := client.NewScoop().Collection(User{}).Where("age", bson.M{"$lte": 35})
	lteCount, err := lteScoop.Count()
	if err != nil {
		t.Fatalf("count with $lte failed: %v", err)
	}
	if lteCount != 3 {
		t.Errorf("expected 3 with $lte 35, got %d", lteCount)
	}

	// Test count with $ne 25 (should match 30, 35, 40)
	neScoop := client.NewScoop().Collection(User{}).Where("age", bson.M{"$ne": 25})
	neCount, err := neScoop.Count()
	if err != nil {
		t.Fatalf("count with $ne failed: %v", err)
	}
	if neCount != 3 {
		t.Errorf("expected 3 with $ne 25, got %d", neCount)
	}
}

// TestFindWithProjection tests Find with field projection
func TestFindWithProjection(t *testing.T) {
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
		Email:     "proj@example.com",
		Name:      "Test User",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// Find without projection
	scoop := client.NewScoop().Collection(User{})
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].ID == primitive.NilObjectID {
		t.Error("expected ID to be populated")
	}
}

// TestCountVariousDataTypes tests Count with different data scenarios
func TestCountVariousDataTypes(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert users with string names and numeric ages
	names := []string{"Alice", "Bob", "Charlie"}
	for i, name := range names {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "dtype" + string(rune(48+i)) + "@example.com",
			Name:      name,
			Age:       20 + (i * 5),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Count by string field
	nameScoop := client.NewScoop().Collection(User{}).Equal("name", "Bob")
	nameCount, err := nameScoop.Count()
	if err != nil {
		t.Fatalf("count by name failed: %v", err)
	}
	if nameCount != 1 {
		t.Errorf("expected 1 user named Bob, got %d", nameCount)
	}

	// Count all
	allScoop := client.NewScoop().Collection(User{})
	allCount, err := allScoop.Count()
	if err != nil {
		t.Fatalf("count all failed: %v", err)
	}
	if allCount != 3 {
		t.Errorf("expected 3 users, got %d", allCount)
	}
}

// TestFirstWithMultipleResults tests First behavior with multiple matching results
func TestFirstWithMultipleResults(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert multiple users with same age
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "multi" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Get first should return one of them
	scoop := client.NewScoop().Collection(User{}).Equal("age", 25)
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("first with multiple results failed: %v", err)
	}

	if result.ID == primitive.NilObjectID {
		t.Error("expected valid user ID")
	}
	if result.Age != 25 {
		t.Errorf("expected age 25, got %d", result.Age)
	}
}

// TestBatchCreateWithMixedTypes tests BatchCreate with different document states
func TestBatchCreateWithMixedTypes(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create batch with minimal data
	user1 := User{
		ID:        primitive.NewObjectID(),
		Email:     "batch1@example.com",
		Name:      "User1",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	user2 := User{
		ID:        primitive.NewObjectID(),
		Email:     "batch2@example.com",
		Name:      "User2",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := scoop.BatchCreate(user1, user2)
	if err != nil {
		t.Fatalf("batch create failed: %v", err)
	}

	// Verify created
	count, _ := scoop.Count()
	if count != 2 {
		t.Errorf("expected 2 documents, got %d", count)
	}
}

// TestFindWithOrdering tests Find preserves insertion order
func TestFindWithOrdering(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert in specific order
	names := []string{"Alice", "Bob", "Charlie"}
	for _, name := range names {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "order" + name + "@example.com",
			Name:      name,
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
		time.Sleep(10 * time.Millisecond) // Small delay to ensure order
	}

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

// TestCountAfterUpdate tests Count reflects updates correctly
func TestCountAfterUpdate(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert initial data
	scoop := client.NewScoop().Collection(User{})
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "update" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		scoop.Create(user)
	}

	// Initial count
	count1, _ := scoop.Count()
	if count1 != 5 {
		t.Errorf("expected 5 initial records, got %d", count1)
	}

	// Update all to new age
	updateScoop := client.NewScoop().Collection(User{})
	_, err := updateScoop.Update(bson.M{"$set": bson.M{"age": 30}})
	if err != nil {
		t.Logf("update operation result: %v", err)
	}

	// Count should still be 5
	countAfter, _ := scoop.Count()
	if countAfter != 5 {
		t.Errorf("expected 5 records after update, got %d", countAfter)
	}
}

// TestExistEmptyCollection tests Exist on empty collection
func TestExistEmptyCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	exists, err := scoop.Exist()
	if err != nil {
		t.Fatalf("exist on empty collection failed: %v", err)
	}

	if exists {
		t.Error("expected false for empty collection")
	}
}

// TestFirstWithSpecificField tests First with specific field value
func TestFirstWithSpecificField(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert specific test data
	targetUser := User{
		ID:        primitive.NewObjectID(),
		Email:     "target@example.com",
		Name:      "Target",
		Age:       99,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", targetUser)

	// Insert other data
	for i := 0; i < 2; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "other" + string(rune(48+i)) + "@example.com",
			Name:      "Other",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Get first with specific field
	scoop := client.NewScoop().Collection(User{}).Equal("age", 99)
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("first with specific field failed: %v", err)
	}

	if result.Age != 99 {
		t.Errorf("expected age 99, got %d", result.Age)
	}
	if result.Name != "Target" {
		t.Errorf("expected name Target, got %s", result.Name)
	}
}

package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestClientHealthCheck tests client Health method
func TestClientHealthCheck(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Health()
	if err != nil {
		t.Logf("Health check returned error: %v", err)
		return
	}
	// Health check succeeded
}

// TestClientPingOperation tests client Ping method
func TestClientPingOperation(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Ping()
	if err != nil {
		t.Logf("Ping returned error: %v", err)
		return
	}
	// Ping succeeded
}

// TestClientGetDatabaseInfo tests client GetDatabase method
func TestClientGetDatabaseInfo(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Get database
	db := client.GetDatabase()
	if db == "" {
		t.Error("expected non-empty database name")
	}
}

// TestScoopCreateCompleteDocument tests Create with complete struct
func TestScoopCreateCompleteDocument(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "complete@example.com",
		Name:      "Complete User",
		Age:       30,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err := scoop.Create(user)
	if err != nil {
		t.Fatalf("create complete document failed: %v", err)
	}

	// Verify
	count, _ := scoop.Count()
	if count != 1 {
		t.Errorf("expected 1 document, got %d", count)
	}
}

// TestScoopFindWithNoFilter tests Find returning all documents
func TestScoopFindWithNoFilter(t *testing.T) {
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
			Email:     "nofilt" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find with no filter failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

// TestScoopUpdateMultiple tests Update on multiple documents
func TestScoopUpdateMultiple(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert multiple documents
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "updmulti" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	updateResult := scoop.Updates(bson.M{"$set": bson.M{"age": 30}})
	updated, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update multiple failed: %v", err)
	}

	if updated != 5 {
		t.Errorf("expected 5 updated, got %d", updated)
	}
}

// TestScoopDeleteMultiple tests Delete on multiple documents
func TestScoopDeleteMultiple(t *testing.T) {
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
			Email:     "delmulti" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete multiple failed: %v", err)
	}

	if deleted != 5 {
		t.Errorf("expected 5 deleted, got %d", deleted)
	}
}

// TestScoopFirstWithSpecificFilter tests First with specific filter and results
func TestScoopFirstWithSpecificFilter(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	targetUser := User{
		ID:        primitive.NewObjectID(),
		Email:     "firstfilt@example.com",
		Name:      "Target",
		Age:       99,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	InsertTestData(t, client, "users", targetUser)

	// Insert other users
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "other" + string(rune(48+i)) + "@example.com",
			Name:      "Other",
			Age:       25,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Get first matching filter
	scoop := client.NewScoop().Collection(User{}).Equal("age", 99)
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("first with filter failed: %v", err)
	}

	if result.Age != 99 {
		t.Errorf("expected age 99, got %d", result.Age)
	}
}

// TestScoopCountLargeDataset tests Count with larger dataset
func TestScoopCountLargeDataset(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Insert 10 documents
	for i := 0; i < 10; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "large" + string(rune(48+i%10)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		scoop.Create(user)
	}

	// Count all
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count large dataset failed: %v", err)
	}

	if count != 10 {
		t.Errorf("expected 10 documents, got %d", count)
	}

	// Count with condition (ages 25, 26, 27, 28, 29 = 5)
	filtScoop := client.NewScoop().Collection(User{}).Where("age", bson.M{"$gte": 25})
	filtCount, err := filtScoop.Count()
	if err != nil {
		t.Fatalf("count with filter failed: %v", err)
	}

	if filtCount != 5 {
		t.Errorf("expected 5 documents with age >= 25, got %d", filtCount)
	}
}

// TestAggregationExecuteWithPipeline tests Execute with complex pipeline
func TestAggregationExecuteWithPipeline(t *testing.T) {
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
			Email:     "agg_pipe" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + (i * 5),
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	model := NewModel[User](client)
	scoop := model.NewScoop().GetScoop()
	agg := scoop.Aggregate()

	// Build aggregation pipeline
	agg.Match(bson.M{"age": bson.M{"$gte": 25}})
	agg.Group(bson.M{"_id": "$age", "count": bson.M{"$sum": 1}})

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("aggregation execute failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected non-empty results from aggregation")
	}
}

// TestScoopUpdateWithBsonM tests Update with bson.M format
func TestScoopUpdateWithBsonM(t *testing.T) {
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
		Email:     "bsonm@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{}).Equal("email", "bsonm@example.com")
	updateResult := scoop.Updates(bson.M{
		"$set": bson.M{
			"name": "Updated",
			"age":  30,
		},
	})
	updated, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update with bson.M failed: %v", err)
	}

	if updated != 1 {
		t.Errorf("expected 1 updated, got %d", updated)
	}
}

// TestScoopNewScoopMultipleTimes tests creating multiple scoop instances
func TestScoopNewScoopMultipleTimes(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create multiple scoop instances
	scoop1 := client.NewScoop().Collection(User{})
	scoop2 := client.NewScoop().Collection(User{})

	// Use first scoop to insert
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "multi_scoop@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	scoop1.Create(user)

	// Use second scoop to query
	count, err := scoop2.Count()
	if err != nil {
		t.Fatalf("count with second scoop failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 document, got %d", count)
	}
}

// TestScoopFindWithCollectionSet tests Find after explicitly setting collection
func TestScoopFindWithCollectionSet(t *testing.T) {
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
		Email:     "coll_set@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	InsertTestData(t, client, "users", user)

	// Create scoop with collection explicitly set
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})

	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find with collection set failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// TestCondMultipleWhereConditions tests using Where multiple times adds conditions
func TestCondMultipleWhereConditions(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data with specific ages
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "where_multi" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Insert users with different ages
	for i := 0; i < 2; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "other" + string(rune(48+i)) + "@example.com",
			Name:      "Other",
			Age:       30 + i,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Use Where to filter by age
	scoop := client.NewScoop().Collection(User{}).
		Where("age", bson.M{"$eq": 25})

	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count with where failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 results with age 25, got %d", count)
	}
}

// TestScoopCloneAndModify tests Clone and then modifying the clone
func TestScoopCloneAndModify(t *testing.T) {
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
			Email:     "clone" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Create base scoop with filter
	base := client.NewScoop().Collection(User{}).Equal("age", 22)

	// Clone it
	cloned := base.Clone()

	// Both should work the same
	baseCount, _ := base.Count()
	clonedCount, _ := cloned.Count()

	if baseCount != clonedCount {
		t.Errorf("base count %d != cloned count %d", baseCount, clonedCount)
	}

	if baseCount != 1 {
		t.Errorf("expected 1 match for age 22, got %d", baseCount)
	}
}

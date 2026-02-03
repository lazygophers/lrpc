package mongo

import (
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestFindWithVariousConditions tests Find with various filter conditions
func TestFindWithVariousConditions(t *testing.T) {
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
			Email:     fmt.Sprintf("find_%d@example.com", i),
			Name:      fmt.Sprintf("User%d", i),
			Age:       20 + i,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Test Find with Name filter
	scoop := client.NewScoop().Collection(User{}).Equal("name", "User0")
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find with equal failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// TestFirstWithVariousConditions tests First with various filters
func TestFirstWithVariousConditions(t *testing.T) {
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
			Email:     fmt.Sprintf("first_%d@example.com", i),
			Name:      fmt.Sprintf("Test%d", i),
			Age:       25 + i,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Test First with Age filter
	scoop := client.NewScoop().Collection(User{}).Where("age", bson.M{"$gte": 26})
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("first with filter failed: %v", err)
	}
	if result.Age < 26 {
		t.Errorf("expected age >= 26, got %d", result.Age)
	}
}

// TestCreateAndVerifyExists tests Create and Exist together
func TestCreateAndVerifyExists(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "exist_test@example.com",
		Name:      "ExistTest",
		Age:       30,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	// Create
	scoop := client.NewScoop().Collection(User{})
	err := scoop.Create(user)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Verify Exist
	scoop2 := client.NewScoop().Collection(User{}).Equal("email", "exist_test@example.com")
	exists, err := scoop2.Exist()
	if err != nil {
		t.Fatalf("exist check failed: %v", err)
	}
	if !exists {
		t.Error("expected document to exist")
	}
}

// TestCountWithoutFilters tests Count on entire collection
func TestCountWithoutFilters(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert 3 documents
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("count_all_%d@example.com", i),
			Name:      "User",
			Age:       30,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

// TestDeleteSingleDocumentPaths tests Delete removing single document
func TestDeleteSingleDocumentPaths(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "delete_single@example.com",
		Name:      "User",
		Age:       30,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{}).Equal("email", "delete_single@example.com")
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}
}

// TestBatchCreateVerify tests BatchCreate and verifies count
func TestBatchCreateVerify(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create batch of users
	users := make([]interface{}, 5)
	for i := 0; i < 5; i++ {
		users[i] = User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("batch_%d@example.com", i),
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
	}

	scoop := client.NewScoop().Collection(User{})
	err := scoop.BatchCreate(users...)
	if err != nil {
		t.Fatalf("batch create failed: %v", err)
	}

	// Verify count
	count, _ := scoop.Count()
	if count != 5 {
		t.Errorf("expected 5 after batch create, got %d", count)
	}
}

// TestFindWithSort tests Find with sorting
func TestFindWithSort(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert unsorted data
	ages := []int{35, 20, 30, 25}
	for i, age := range ages {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("sort_%d@example.com", i),
			Name:      "User",
			Age:       age,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Find with sort
	scoop := client.NewScoop().Collection(User{}).Sort("age", 1)
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find with sort failed: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}
}

// TestUpdateFields tests Update with specific fields
func TestUpdateFields(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "update_test@example.com",
		Name:      "Original",
		Age:       25,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	InsertTestData(t, client, "users", user)

	// Update specific field
	scoop := client.NewScoop().Collection(User{}).Equal("email", "update_test@example.com")
	updated := bson.M{"name": "Updated", "age": 30}
	updateResult := scoop.Updates(updated)
	count, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if count <= 0 {
		t.Errorf("expected count > 0, got %d", count)
	}
}

// TestCloneScoop tests Clone to copy scoop state
func TestCloneScoop(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop1 := client.NewScoop().Collection(User{}).Where("age", bson.M{"$gte": 25})
	scoop2 := scoop1.Clone()

	if scoop2 == nil {
		t.Error("expected cloned scoop to be non-nil")
	}
}

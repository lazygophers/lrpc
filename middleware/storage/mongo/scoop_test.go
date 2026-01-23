package mongo

import (
	"testing"
	"time"

	"github.com/lazygophers/lrpc/middleware/core"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestScoopFind(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Find with scoop
	scoop := client.NewScoop()
	scoop = scoop.Equal("age", 25)

	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Fatalf("find failed: %v", findResult.Error)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if results[0].Age != 25 {
		t.Errorf("expected age 25, got %d", results[0].Age)
	}
}

func TestScoopFirst(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}
	InsertTestData(t, client, "users", user)

	// FindOne with scoop
	scoop := client.NewScoop()
	scoop = scoop.Equal("email", "test@example.com")

	var result User
	firstResult := scoop.First(&result)
	if firstResult.Error != nil {
		t.Fatalf("find one failed: %v", firstResult.Error)
	}

	if result.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", result.Email)
	}
}

func TestScoopCount(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Count with scoop
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("age", 25)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 documents, got %d", count)
	}
}

func TestScoopExist(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}
	InsertTestData(t, client, "users", user)

	// Test exist
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("email", "test@example.com")

	exists, err := scoop.Exist()
	if err != nil {
		t.Fatalf("exist failed: %v", err)
	}

	if !exists {
		t.Error("expected document to exist")
	}

	// Test not exist
	scoop = client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("email", "notexist@example.com")

	exists, err = scoop.Exist()
	if err != nil {
		t.Fatalf("exist failed: %v", err)
	}

	if exists {
		t.Error("expected document to not exist")
	}
}

func TestScoopCreate(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create with scoop
	scoop := client.NewScoop()

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "newuser@example.com",
		Name:      "New User",
		Age:       28,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err := scoop.Create(user)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Verify
	AssertDocumentExists(t, client, "users", bson.M{"email": "newuser@example.com"})
}

func TestScoopUpdate(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}
	InsertTestData(t, client, "users", user)

	// Update with scoop
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("email", "test@example.com")

	updateResult := scoop.Updates(bson.M{"age": 30, "name": "Updated User"})
	count, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 updated document, got %d", count)
	}

	// Verify
	doc := GetTestDocument(t, client, "users", bson.M{"email": "test@example.com"})
	if age, ok := doc["age"].(int32); !ok || age != 30 {
		t.Errorf("expected age 30, got %v", doc["age"])
	}
}

func TestScoopDelete(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Delete with scoop
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("age", 25)

	deleteResult := scoop.Delete()
	if deleteResult.Error != nil {
		t.Fatalf("delete failed: %v", deleteResult.Error)
	}

	if deleteResult.DocsAffected != 2 {
		t.Errorf("expected 2 deleted documents, got %d", deleteResult.DocsAffected)
	}

	// Verify
	AssertCount(t, 0, client, "users", bson.M{})
}

func TestScoopLimitOffset(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Find with limit and offset
	scoop := client.NewScoop()
	scoop = scoop.Equal("age", 25).Limit(2).Offset(1)

	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Fatalf("find failed: %v", findResult.Error)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestScoopSelect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}
	InsertTestData(t, client, "users", user)

	// Find with select
	scoop := client.NewScoop()
	scoop = scoop.Select("email", "name")

	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Fatalf("find failed: %v", findResult.Error)
	}

	if len(results) == 0 {
		t.Error("expected results, got empty")
	}
}

func TestScoopComparison(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Gt
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Gt("age", 25)
	count, _ := scoop.Count()
	if count != 1 {
		t.Errorf("expected 1 result for Gt, got %d", count)
	}

	// Test Lt
	scoop = client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Lt("age", 25)
	count, _ = scoop.Count()
	if count != 0 {
		t.Errorf("expected 0 results for Lt, got %d", count)
	}

	// Test Between
	scoop = client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Between("age", 20, 30)
	count, _ = scoop.Count()
	if count != 3 {
		t.Errorf("expected 3 results for Between, got %d", count)
	}
}

func TestScoopIn(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test In
	scoop := client.NewScoop()
	scoop = scoop.In("age", 25, 30)
	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Fatalf("find failed: %v", findResult.Error)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for In, got %d", len(results))
	}

	// Test NotIn
	scoop = client.NewScoop()
	scoop = scoop.NotIn("age", 25, 30)
	results = nil
	findResult = scoop.Find(&results)
	if findResult.Error != nil {
		t.Fatalf("find failed: %v", findResult.Error)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for NotIn, got %d", len(results))
	}
}

// TestSortAscending tests sort with ascending order
func TestSortAscending(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data with different ages
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 20, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Sort ascending - default (no direction param)
	scoop := client.NewScoop().Collection(User{}).Sort("age")
	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Fatalf("find failed: %v", findResult.Error)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Verify ascending order: 20, 25, 30
	if results[0].Age != 20 || results[1].Age != 25 || results[2].Age != 30 {
		t.Errorf("expected ascending order [20, 25, 30], got [%d, %d, %d]", results[0].Age, results[1].Age, results[2].Age)
	}
}

// TestSortDescending tests sort with descending order
func TestSortDescending(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data with different ages
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 20, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Sort descending
	scoop := client.NewScoop().Collection(User{}).Sort("age", -1)
	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Fatalf("find failed: %v", findResult.Error)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Verify descending order: 30, 25, 20
	if results[0].Age != 30 || results[1].Age != 25 || results[2].Age != 20 {
		t.Errorf("expected descending order [30, 25, 20], got [%d, %d, %d]", results[0].Age, results[1].Age, results[2].Age)
	}
}

func TestScoopGte(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 20, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Gte
	scoop := client.NewScoop().Collection(User{}).Gte("age", 25)
	count, err := scoop.Count()
	if err != nil {
		t.Errorf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 results for Gte(age >= 25), got %d", count)
	}
}

func TestScoopLte(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 20, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Lte
	scoop := client.NewScoop().Collection(User{}).Lte("age", 25)
	count, err := scoop.Count()
	if err != nil {
		t.Errorf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 results for Lte(age <= 25), got %d", count)
	}
}

func TestScoopLike(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "demo@example.com", Name: "Demo User", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "other@example.com", Name: "Other User", Age: 20, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Like with pattern
	scoop := client.NewScoop().Like("name", "Test")
	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Errorf("find failed: %v", findResult.Error)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for Like pattern, got %d", len(results))
	}
}

func TestScoopSkip(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Skip (should be equivalent to Offset)
	scoop := client.NewScoop().Equal("age", 25).Skip(1).Limit(2)
	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Errorf("find failed: %v", findResult.Error)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results with Skip(1), got %d", len(results))
	}
}

func TestScoopClear(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Create a scoop with filter
	scoop := client.NewScoop().Collection(User{}).Equal("age", 25)
	count, err := scoop.Count()
	if err != nil {
		t.Errorf("count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 result before clear, got %d", count)
	}

	// Clear the scoop and verify no filters
	scoop = scoop.Clear()
	count, err = scoop.Count()
	if err != nil {
		t.Errorf("count after clear failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 results after clear, got %d", count)
	}
}

func TestScoopBatchCreate(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create batch of users as interface{} varargs
	user1 := User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}
	user2 := User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}
	user3 := User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}

	// BatchCreate with varargs
	scoop := client.NewScoop()
	err := scoop.BatchCreate(user1, user2, user3)
	if err != nil {
		t.Errorf("batch create failed: %v", err)
	}

	// Verify all users were created
	scoop = client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Errorf("count failed: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 users after batch create, got %d", count)
	}
}

func TestScoopIsNotFound(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create a scoop
	scoop := client.NewScoop()

	// Insert a user
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "test@example.com",
		Name:      "Test User",
		Age:       25,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}
	err := scoop.Create(user)
	if err != nil {
		t.Errorf("create failed: %v", err)
	}

	// Try to find non-existent user
	scoop = client.NewScoop().Equal("email", "nonexistent@example.com")
	var result User
	firstResult := scoop.First(&result)

	// Verify IsNotFound
	if !scoop.IsNotFound(firstResult.Error) {
		t.Errorf("expected IsNotFound to return true for no documents found")
	}
}

func TestScoopFindByPage(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data - 25 users for pagination testing
	users := make([]interface{}, 25)
	for i := 0; i < 25; i++ {
		users[i] = User{
			ID:        primitive.NewObjectID(),
			Email:     "user" + string(rune(i)) + "@example.com",
			Name:      "User " + string(rune(i)),
			Age:       20 + i,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
	}
	InsertTestData(t, client, "users", users...)

	// Test 1: Basic FindByPage with default options
	scoop := client.NewScoop()
	opt := &core.ListOption{
		Offset:    0,
		Limit:     10,
		ShowTotal: true,
	}

	var results []User
	paginate, err := scoop.FindByPage(opt, &results)
	if err != nil {
		t.Fatalf("FindByPage failed: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("expected 10 results, got %d", len(results))
	}

	if paginate.Offset != 0 {
		t.Errorf("expected offset 0, got %d", paginate.Offset)
	}

	if paginate.Limit != 10 {
		t.Errorf("expected limit 10, got %d", paginate.Limit)
	}

	if paginate.Total != 25 {
		t.Errorf("expected total 25, got %d", paginate.Total)
	}

	// Test 2: FindByPage with offset
	scoop = client.NewScoop()
	opt = &core.ListOption{
		Offset:    10,
		Limit:     10,
		ShowTotal: true,
	}

	results = []User{}
	paginate, err = scoop.FindByPage(opt, &results)
	if err != nil {
		t.Fatalf("FindByPage with offset failed: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("expected 10 results, got %d", len(results))
	}

	if paginate.Total != 25 {
		t.Errorf("expected total 25, got %d", paginate.Total)
	}

	// Test 3: FindByPage without showing total
	scoop = client.NewScoop()
	opt = &core.ListOption{
		Offset:    0,
		Limit:     5,
		ShowTotal: false,
	}

	results = []User{}
	paginate, err = scoop.FindByPage(opt, &results)
	if err != nil {
		t.Fatalf("FindByPage without total failed: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("expected 5 results, got %d", len(results))
	}

	if paginate.Total != 0 {
		t.Errorf("expected total 0 (ShowTotal=false), got %d", paginate.Total)
	}

	// Test 4: FindByPage with where condition
	scoop = client.NewScoop().Gte("age", 30)
	opt = &core.ListOption{
		Offset:    0,
		Limit:     10,
		ShowTotal: true,
	}

	results = []User{}
	paginate, err = scoop.FindByPage(opt, &results)
	if err != nil {
		t.Fatalf("FindByPage with condition failed: %v", err)
	}

	// We should have 20 users with age >= 30 (users with age 20-44, so 24-44 is 20 users)
	expectedCount := uint64(20)
	if paginate.Total != expectedCount {
		t.Errorf("expected total %d, got %d", expectedCount, paginate.Total)
	}

	// Test 5: FindByPage with nil options (should use defaults)
	scoop = client.NewScoop()
	results = []User{}
	paginate, err = scoop.FindByPage(nil, &results)
	if err != nil {
		t.Fatalf("FindByPage with nil options failed: %v", err)
	}

	if len(results) != 20 { // Default limit is 20
		t.Errorf("expected 20 results (default limit), got %d", len(results))
	}

	if paginate.Offset != 0 {
		t.Errorf("expected offset 0 (default), got %d", paginate.Offset)
	}

	if paginate.Limit != 20 {
		t.Errorf("expected limit 20 (default), got %d", paginate.Limit)
	}
}

func TestScoopFindByPageDepth(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test depth tracking
	scoop := client.NewScoop()
	initialDepth := scoop.depth
	if initialDepth != 3 {
		t.Errorf("expected initial depth 3, got %d", initialDepth)
	}

	opt := &core.ListOption{
		Offset:    0,
		Limit:     10,
		ShowTotal: true,
	}

	var results []User
	_, err := scoop.FindByPage(opt, &results)
	if err != nil {
		t.Fatalf("FindByPage failed: %v", err)
	}

	// After FindByPage, depth should be restored to initial value
	if scoop.depth != initialDepth {
		t.Errorf("expected depth %d after FindByPage, got %d", initialDepth, scoop.depth)
	}
}

func TestScoopLogging(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}
	InsertTestData(t, client, "users", users...)

	// Test logging in various operations
	scoop := client.NewScoop()

	// Test Find logging
	var results []User
	findResult := scoop.Find(&results)
	if findResult.Error != nil {
		t.Fatalf("Find failed: %v", findResult.Error)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Test Count logging
	scoop = client.NewScoop()
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}

	// Test Update logging
	scoop = client.NewScoop().Equal("age", 25)
	updateResult := scoop.Updates(bson.M{"name": "Updated User 1"})
	if updateResult.Error != nil {
		t.Fatalf("Update failed: %v", updateResult.Error)
	}

	if updateResult.DocsAffected != 1 {
		t.Errorf("expected 1 modified, got %d", updateResult.DocsAffected)
	}

	// Test Delete logging
	scoop = client.NewScoop().Equal("age", 30)
	deleteResult2 := scoop.Delete()
	if deleteResult2.Error != nil {
		t.Fatalf("Delete failed: %v", deleteResult2.Error)
	}

	if deleteResult2.DocsAffected != 1 {
		t.Errorf("expected 1 deleted, got %d", deleteResult2.DocsAffected)
	}
}

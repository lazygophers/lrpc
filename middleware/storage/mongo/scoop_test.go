package mongo

import (
	"testing"
	"time"

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
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Find with scoop
	scoop := client.NewScoop()
	scoop = scoop.Equal("age", 25)

	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
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
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	InsertTestData(t, client, "users", user)

	// FindOne with scoop
	scoop := client.NewScoop()
	scoop = scoop.Equal("email", "test@example.com")

	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("find one failed: %v", err)
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
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
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
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	InsertTestData(t, client, "users", user)

	// Update with scoop
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("email", "test@example.com")

	count, err := scoop.Update(bson.M{"age": 30, "name": "Updated User"})
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
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Delete with scoop
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("age", 25)

	count, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 deleted documents, got %d", count)
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
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Find with limit and offset
	scoop := client.NewScoop()
	scoop = scoop.Equal("age", 25).Limit(2).Offset(1)

	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
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
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	InsertTestData(t, client, "users", user)

	// Find with select
	scoop := client.NewScoop()
	scoop = scoop.Select("email", "name")

	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
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
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
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
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test In
	scoop := client.NewScoop()
	scoop = scoop.In("age", 25, 30)
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for In, got %d", len(results))
	}

	// Test NotIn
	scoop = client.NewScoop()
	scoop = scoop.NotIn("age", 25, 30)
	results = nil
	err = scoop.Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
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
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 20, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Sort ascending - default (no direction param)
	scoop := client.NewScoop().Collection(User{}).Sort("age")
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
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
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 20, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Sort descending
	scoop := client.NewScoop().Collection(User{}).Sort("age", -1)
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Verify descending order: 30, 25, 20
	if results[0].Age != 30 || results[1].Age != 25 || results[2].Age != 20 {
		t.Errorf("expected descending order [30, 25, 20], got [%d, %d, %d]", results[0].Age, results[1].Age, results[2].Age)
	}
}

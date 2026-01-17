package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCondEqual tests the Equal condition
func TestCondEqual(t *testing.T) {
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

	// Test with Cond
	cond := Where("age", "=", 25)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 result, got %d", count)
	}
}

// TestCondNe tests the Ne condition
func TestCondNe(t *testing.T) {
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

	// Test with Cond
	cond := Where("age", "$ne", 25)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 result, got %d", count)
	}
}

// TestCondGt tests the Gt condition
func TestCondGt(t *testing.T) {
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

	// Test with Cond
	cond := Gt("age", 25)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 results, got %d", count)
	}
}

// TestCondLt tests the Lt condition
func TestCondLt(t *testing.T) {
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

	// Test with Cond
	cond := Lt("age", 30)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 result, got %d", count)
	}
}

// TestCondIn tests the In condition
func TestCondIn(t *testing.T) {
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

	// Test with Cond
	cond := In("age", 25, 35)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 results, got %d", count)
	}
}

// TestCondBetween tests the Between condition
func TestCondBetween(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 20, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user4@example.com", Name: "User 4", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test with Cond
	cond := Between("age", 25, 30)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 results, got %d", count)
	}
}

// TestCondLike tests the Like condition
func TestCondLike(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "Alice", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "Bob", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test with Cond
	cond := Like("name", "ice")
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 result, got %d", count)
	}
}

// TestCondOrConditions tests OR conditions
func TestCondOrConditions(t *testing.T) {
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

	// Test: a = 1 or a = 2
	cond := Or(Equal("age", 25), Equal("age", 30))
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 results for OR condition, got %d", count)
	}
}

// TestCondComplexConditions tests complex nested OR/AND conditions
func TestCondComplexConditions(t *testing.T) {
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

	// Test: (age = 25 or age = 30) and name != ""
	cond := Where(
		Or(Equal("age", 25), Equal("age", 30)),
		Ne("name", ""),
	)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 results for complex condition, got %d", count)
	}
}

// TestCondUpdate tests updating with condition
func TestCondUpdate(t *testing.T) {
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

	// Update with condition
	cond := Gt("age", 25)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Update(bson.M{"name": "Updated"})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 updated document, got %d", count)
	}
}

// TestCondDelete tests deleting with condition
func TestCondDelete(t *testing.T) {
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

	// Delete with condition
	cond := Gt("age", 25)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 deleted document, got %d", count)
	}

	// Verify remaining count
	remainingCount, err := client.NewScoop().Collection(User{}).Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if remainingCount != 1 {
		t.Errorf("expected 1 remaining document, got %d", remainingCount)
	}
}

// TestCondFind tests finding with condition
func TestCondFind(t *testing.T) {
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

	// Find with condition
	cond := Between("age", 25, 30)
	var results []User
	err := client.NewScoop().Collection(User{}).Where(cond).Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

// TestCondMapInput tests condition with map input
func TestCondMapInput(t *testing.T) {
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

	// Test with map input
	cond := Where(map[string]interface{}{
		"age":  25,
		"name": "User 1",
	})
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 result, got %d", count)
	}
}

// TestScoopOrFilter tests Or function combining multiple conditions
func TestScoopOrFilter(t *testing.T) {
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

	// Test Or function: age = 25 OR age = 35
	cond1 := Equal("age", 25)
	cond2 := Equal("age", 35)
	orCond := Or(cond1, cond2)
	scoop := client.NewScoop().
		Collection(User{}).
		Where(orCond)

	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 results for Or function, got %d", count)
	}
}

// TestScoopWhereWithCond tests new Where method with Cond
func TestScoopWhereWithCond(t *testing.T) {
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

	// Test Where with Cond
	cond := Gt("age", 25)
	scoop := client.NewScoop().Collection(User{}).Where(cond)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 result with Where(Cond), got %d", count)
	}
}

// TestScoopWhereWithTraditional tests backward compatibility of Where method
func TestScoopWhereWithTraditional(t *testing.T) {
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

	// Test traditional Where method (backward compatibility)
	scoop := client.NewScoop().Collection(User{}).Where("age", 25)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 result with traditional Where, got %d", count)
	}
}

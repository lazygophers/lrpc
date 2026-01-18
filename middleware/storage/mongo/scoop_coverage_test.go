package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestFirstWithResults tests First when documents exist
func TestFirstWithResults(t *testing.T) {
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

	// Find first with filter
	scoop := client.NewScoop().Equal("age", 30)
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("find first failed: %v", err)
	}

	if result.Age != 30 {
		t.Errorf("expected age 30, got %d", result.Age)
	}
}

// TestCountWithFilters tests Count with various filters
func TestCountWithFilters(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data with different ages
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 20, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user4@example.com", Name: "User 4", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Count with equality filter
	scoop := client.NewScoop().Collection(User{}).Equal("age", 25)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}

	// Count with greater than filter
	scoop = client.NewScoop().Collection(User{}).Gt("age", 25)
	count, err = scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected count 1 for age > 25, got %d", count)
	}
}

// TestDeleteWithFilters tests Delete with various filters
func TestDeleteWithFilters(t *testing.T) {
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
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Delete with filter
	scoop := client.NewScoop().Collection(User{}).Gt("age", 25)
	count, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 deleted document, got %d", count)
	}

	// Verify remaining count
	scoop = client.NewScoop().Collection(User{})
	remaining, err := scoop.Count()
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if remaining != 2 {
		t.Errorf("expected 2 remaining documents, got %d", remaining)
	}
}

// TestFindWithMultipleFilters tests Find with chained filters
func TestFindWithMultipleFilters(t *testing.T) {
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
	}
	InsertTestData(t, client, "users", users...)

	// Find all documents
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

// TestFindWithSelectProjection tests Find with field selection
func TestFindWithSelectProjection(t *testing.T) {
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

	// Find with projection
	scoop := client.NewScoop().Collection(User{}).Select("name", "email")
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find with select failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected results with projection")
	}
}

// TestCloneAndModifyIndependently tests Clone creates independent instances
func TestCloneAndModifyIndependently(t *testing.T) {
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

	// Create original scoop
	scoop := client.NewScoop().Collection(User{}).Equal("age", 25)
	count1, _ := scoop.Count()

	// Clone and modify with different filter - should be independent
	cloned := scoop.Clone().Clear().Gt("age", 30)
	count2, _ := cloned.Count()

	// Verify original scoop is still the same
	count1Again, _ := scoop.Count()

	if count1 != count1Again {
		t.Errorf("expected original scoop unchanged, got %d then %d", count1, count1Again)
	}

	if count1 != 1 || count2 != 1 {
		t.Errorf("expected count1=1, count2=1, got count1=%d, count2=%d", count1, count2)
	}
}

// TestGetCollectionMethod tests GetCollection method
func TestGetCollectionMethod(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})
	col := scoop.GetCollection()

	if col == nil {
		t.Error("expected collection, got nil")
	}

	if col.Name() != "users" {
		t.Errorf("expected collection name 'users', got %s", col.Name())
	}
}

// TestSortDescendingMultiple tests Sort with multiple fields
func TestSortDescendingMultiple(t *testing.T) {
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
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "Bob", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "Charlie", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Sort by age descending
	scoop := client.NewScoop().Collection(User{}).Sort("age", -1)
	var results []User
	err := scoop.Find(&results)
	if err != nil {
		t.Fatalf("find with sort failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Verify descending order
	if results[0].Age < results[1].Age {
		t.Error("expected descending order by age")
	}
}

// TestExistFiltered tests Exist with filters
func TestExistFiltered(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	InsertTestData(t, client, "users", user)

	// Test exist with matching filter
	scoop := client.NewScoop().Collection(User{}).Equal("age", 25)
	exists, err := scoop.Exist()
	if err != nil {
		t.Fatalf("exist failed: %v", err)
	}

	if !exists {
		t.Error("expected document to exist")
	}

	// Test exist with non-matching filter
	scoop = client.NewScoop().Collection(User{}).Equal("age", 99)
	exists, err = scoop.Exist()
	if err != nil {
		t.Fatalf("exist failed: %v", err)
	}

	if exists {
		t.Error("expected document not to exist")
	}
}

// TestCondEqualOperator tests Equal operator on Cond
func TestCondEqualOperator(t *testing.T) {
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

	// Test Equal on Cond
	cond := NewCond().Equal("age", 25)
	scoop := client.NewScoop().Collection(User{})
	var results []User
	err := scoop.Where(cond).Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// TestCondNeOperator tests Ne operator on Cond
func TestCondNeOperator(t *testing.T) {
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

	// Test Ne on Cond
	cond := NewCond().Ne("age", 25)
	scoop := client.NewScoop().Collection(User{})
	var results []User
	err := scoop.Where(cond).Find(&results)
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if results[0].Age == 25 {
		t.Error("expected result with age != 25")
	}
}

// TestClear resets scoop state
func TestClearScoop(t *testing.T) {
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

	// Create scoop with filters
	scoop := client.NewScoop().Collection(User{}).Equal("age", 25)
	count, _ := scoop.Count()
	if count != 1 {
		t.Errorf("expected 1 before clear, got %d", count)
	}

	// Clear and verify all documents are returned
	scoop.Clear()
	count, _ = scoop.Count()
	if count != 2 {
		t.Errorf("expected 2 after clear, got %d", count)
	}
}

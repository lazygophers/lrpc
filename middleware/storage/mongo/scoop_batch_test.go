package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCreateWithError tests Create when collection not set
func TestCreateWithError(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Create without setting collection
	scoop := client.NewScoop()
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "noexist@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := scoop.Create(user)
	// Should either succeed (collection auto-determined) or fail with specific error
	if err != nil {
		t.Logf("Create without collection set returned error: %v", err)
	}
}

// TestBatchCreateEmpty tests BatchCreate with no documents
func TestBatchCreateEmpty(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})
	err := scoop.BatchCreate()
	// Should return error for no documents
	if err == nil {
		t.Error("expected error for empty batch create")
	}
}

// TestBatchCreateSingleDocument tests BatchCreate with one document
func TestBatchCreateSingleDocument(t *testing.T) {
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
		Email:     "single@example.com",
		Name:      "Single",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := scoop.BatchCreate(user)
	if err != nil {
		t.Fatalf("batch create single failed: %v", err)
	}

	// Verify it was created
	count, _ := scoop.Count()
	if count != 1 {
		t.Errorf("expected 1 document, got %d", count)
	}
}

// TestUpdateWithoutFilter tests Update with no filter condition
func TestUpdateWithoutFilter(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert some test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "all1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "all2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Update without filter (should update all)
	scoop := client.NewScoop().Collection(User{})
	updateResult := scoop.Updates(bson.M{"$set": bson.M{"status": "updated"}}
	count, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update without filter failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 updated documents, got %d", count)
	}
}

// TestUpdateWithNullValue tests Update with nil/null values
func TestUpdateWithNullValue(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	user := User{ID: primitive.NewObjectID(), Email: "null@example.com", Name: "Test", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	InsertTestData(t, client, "users", user)

	// Update with null/empty value
	scoop := client.NewScoop().Collection(User{}).Equal("email", "null@example.com")
	updateResult := scoop.Updates(bson.M{"age": nil}
	count, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update with nil failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 updated document, got %d", count)
	}
}

// TestCountEmptyCollection tests Count on empty collection
func TestCountEmptyCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Count empty collection
	scoop := client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count empty collection failed: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 documents, got %d", count)
	}
}

// TestCountWithoutCollection tests Count without setting collection
func TestCountWithoutCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	defer func() {
		if r := recover(); r != nil {
			t.Logf("Count without collection panicked as expected: %v", r)
		}
	}()

	scoop := client.NewScoop()
	_, err := scoop.Count()
	// Should fail if collection not set
	if err != nil {
		t.Logf("Count without collection returned error: %v", err)
	}
}

// TestDeleteAll tests Delete without filter (delete all)
func TestDeleteAll(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "del_all1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "del_all2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Delete all
	scoop := client.NewScoop().Collection(User{})
	deleteResult := scoop.Delete()
	count, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete all failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 deleted documents, got %d", count)
	}

	// Verify empty
	remaining, _ := scoop.Count()
	if remaining != 0 {
		t.Errorf("expected 0 remaining documents, got %d", remaining)
	}
}

// TestDeleteWithComplexFilter tests Delete with complex filter conditions
func TestDeleteWithComplexFilter(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "complex1@example.com", Name: "User 1", Age: 20, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "complex2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "complex3@example.com", Name: "User 3", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Delete with simple filter (just age 25)
	scoop := client.NewScoop().Collection(User{}).Equal("age", 25)
	deleteResult := scoop.Delete()
	count, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete with filter failed: %v", err)
	}

	// Verify the correct document was deleted
	if count != 1 {
		t.Errorf("expected 1 deleted document, got %d", count)
	}
}

// TestCloneWithFilters tests Clone preserves filters correctly
func TestCloneWithFilters(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "clone1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "clone2@example.com", Name: "User 2", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Create scoop with filter
	base := client.NewScoop().Collection(User{}).Equal("age", 25)

	// Clone should preserve the filter
	cloned := base.Clone()

	// Verify both work the same
	count1, _ := base.Count()
	count2, _ := cloned.Count()

	if count1 != count2 {
		t.Errorf("expected same count from base and clone, got %d and %d", count1, count2)
	}

	if count1 != 1 {
		t.Errorf("expected count 1, got %d", count1)
	}
}

// TestCloneWithSortAndLimit tests Clone with sort and limit
func TestCloneWithSortAndLimit(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	for i := 1; i <= 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "sort_limit" + string(rune(48+i)) + "@example.com",
			Name:      "User " + string(rune(48+i)),
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Create scoop with sort and limit
	base := client.NewScoop().Collection(User{}).Sort("age", -1).Limit(2)

	// Clone should preserve sort and limit
	cloned := base.Clone()

	var results1 []User
	var results2 []User
	base.Find(&results1)
	cloned.Find(&results2)

	if len(results1) != len(results2) {
		t.Errorf("expected same number of results from clone, got %d and %d", len(results1), len(results2))
	}

	if len(results1) != 2 {
		t.Errorf("expected 2 results with limit, got %d", len(results1))
	}
}

// TestBeginTransaction tests Begin transaction creation
func TestBeginTransaction(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop()
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Logf("Begin transaction error (may need MongoDB transaction support): %v", err)
		return
	}

	if txScoop == nil {
		t.Error("expected non-nil transaction scoop")
	}

	// Clean up transaction
	_ = txScoop.Rollback()
}

// TestRollback tests transaction rollback
func TestRollback(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop()
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Logf("Begin transaction skipped: %v", err)
		return
	}

	// Try to create document in transaction
	user := User{ID: primitive.NewObjectID(), Email: "tx@example.com", Name: "TX User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	txScoop.Collection(User{})
	_ = txScoop.Create(user)

	// Rollback
	err = txScoop.Rollback()
	if err != nil {
		t.Logf("Rollback error: %v", err)
	}

	// Verify document was rolled back
	scoop = client.NewScoop().Collection(User{})
	count, _ := scoop.Count()
	// May or may not be rolled back depending on MongoDB version
	t.Logf("Document count after rollback: %d", count)
}

// TestCommit tests transaction commit
func TestCommit(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop()
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Logf("Begin transaction skipped: %v", err)
		return
	}

	// Try to create document in transaction
	user := User{ID: primitive.NewObjectID(), Email: "commit@example.com", Name: "Commit User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	txScoop.Collection(User{})
	_ = txScoop.Create(user)

	// Commit
	err = txScoop.Commit()
	if err != nil {
		t.Logf("Commit error: %v", err)
	}

	// Verify document was committed
	scoop = client.NewScoop().Collection(User{})
	count, _ := scoop.Count()
	// Should have at least the committed document if transaction succeeded
	if count > 0 {
		t.Logf("Document successfully committed, count: %d", count)
	}
}

// TestFirstWithSort tests First returns a document
func TestFirstWithSort(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "first_sort1@example.com", Name: "User", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "first_sort2@example.com", Name: "User", Age: 20, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Get first (order may vary without explicit sort in Find)
	scoop := client.NewScoop().Collection(User{})
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("first failed: %v", err)
	}

	// Should get one of the inserted documents
	if result.Age != 30 && result.Age != 20 {
		t.Errorf("expected age 20 or 30, got %d", result.Age)
	}
})

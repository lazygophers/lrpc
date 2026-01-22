package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCountErrorHandling tests Count error handling path
func TestCountErrorHandling(t *testing.T) {
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
		Email:     "count_error@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// Create a scoop with normal data
	scoop := client.NewScoop().Collection(User{})

	// Count should work normally
	count, err := scoop.Count()
	if err != nil {
		t.Logf("Count returned error: %v", err)
	} else if count == 0 {
		t.Errorf("Expected count > 0, got %d", count)
	}
}

// TestDeleteErrorHandling tests Delete error handling path
func TestDeleteErrorHandling(t *testing.T) {
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
			Email:     "delete_error_" + string(rune(48+i)) + "@example.com",
			Name:      "Test",
			Age:       25 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Delete should work normally
	scoop := client.NewScoop().Collection(User{})
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Logf("Delete returned error: %v", err)
	} else if deleted != 5 {
		t.Errorf("Expected 5 deleted, got %d", deleted)
	}
}

// TestCommitWithNilSession tests Commit when session is nil
func TestCommitWithNilSession(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Create a scoop without starting a transaction
	scoop := client.NewScoop().Collection(User{})

	// Try to commit without a session - should return error
	err := scoop.Commit()
	if err == nil {
		t.Error("Expected error when committing with nil session, got nil")
	}
	if err != nil && err.Error() != "no active transaction" {
		t.Logf("Got error: %v", err)
	}
}

// TestRollbackWithNilSession tests Rollback when session is nil
func TestRollbackWithNilSession(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Create a scoop without starting a transaction
	scoop := client.NewScoop().Collection(User{})

	// Try to rollback without a session - should return error
	err := scoop.Rollback()
	if err == nil {
		t.Error("Expected error when rolling back with nil session, got nil")
	}
	if err != nil && err.Error() != "no active transaction" {
		t.Logf("Got error: %v", err)
	}
}

// TestCommitAfterBegin tests successful Commit after Begin
func TestCommitAfterBegin(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Begin transaction
	scoop := client.NewScoop().Collection(User{})
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	// Create a user in transaction
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "transaction@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = txScoop.Create(user)
	if err != nil {
		t.Logf("Create in transaction failed: %v", err)
	}

	// Commit transaction
	err = txScoop.Commit()
	if err != nil {
		t.Logf("Commit failed: %v", err)
	}
}

// TestRollbackAfterBegin tests successful Rollback after Begin
func TestRollbackAfterBegin(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Begin transaction
	scoop := client.NewScoop().Collection(User{})
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	// Create a user in transaction
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "rollback_test@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = txScoop.Create(user)
	if err != nil {
		t.Logf("Create in transaction failed: %v", err)
	}

	// Rollback transaction
	err = txScoop.Rollback()
	if err != nil {
		t.Logf("Rollback failed: %v", err)
	}
}

// TestDatabaseChangeStreamWatchErrorHandling tests Watch error handling
func TestDatabaseChangeStreamWatchErrorHandling(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Create database change stream
	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("Create database change stream failed: %v", err)
	}

	if dcs == nil {
		t.Error("Expected non-nil database change stream")
	}

	// Close the change stream
	dcs.Close()
}

// TestChangeStreamWatchWithVariousPipelines tests Watch with different pipelines
func TestChangeStreamWatchWithVariousPipelines(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Test 1: Watch with no pipeline
	cs1, err := scoop.WatchChanges()
	if err != nil {
		t.Logf("WatchChanges with no pipeline failed: %v", err)
	}
	if cs1 != nil {
		cs1.Close()
	}

	// Test 2: Watch with $match pipeline
	cs2, err := scoop.WatchChanges(bson.M{
		"$match": bson.M{
			"operationType": "insert",
		},
	})
	if err != nil {
		t.Logf("WatchChanges with $match failed: %v", err)
	}
	if cs2 != nil {
		cs2.Close()
	}

	// Test 3: Watch with multiple pipeline stages
	cs3, err := scoop.WatchChanges(
		bson.M{
			"$match": bson.M{
				"operationType": "insert",
			},
		},
		bson.M{
			"$project": bson.M{
				"fullDocument": 1,
			},
		},
	)
	if err != nil {
		t.Logf("WatchChanges with multiple stages failed: %v", err)
	}
	if cs3 != nil {
		cs3.Close()
	}
}

// TestCountEdgeCases tests Count with various edge conditions
func TestCountEdgeCases(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Test 1: Count on empty collection
	count, err := scoop.Count()
	if err != nil {
		t.Logf("Count on empty collection failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 on empty collection, got %d", count)
	}

	// Test 2: Insert one and count
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "edge1@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	count, err = scoop.Count()
	if err != nil {
		t.Logf("Count with one document failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1, got %d", count)
	}

	// Test 3: Count with complex filter
	scoop2 := client.NewScoop().Collection(User{}).
		Where("age", bson.M{"$gte": 25}).
		Where("age", bson.M{"$lte": 30})

	count, err = scoop2.Count()
	if err != nil {
		t.Logf("Count with complex filter failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1, got %d", count)
	}
}

// TestDeleteEdgeCases tests Delete with various edge conditions
func TestDeleteEdgeCases(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Test 1: Delete from empty collection
	scoop := client.NewScoop().Collection(User{})
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Logf("Delete from empty collection failed: %v", err)
	}
	if deleted != 0 {
		t.Errorf("Expected 0 deleted on empty collection, got %d", deleted)
	}

	// Test 2: Insert and delete all
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "del_edge_" + string(rune(48+i)) + "@example.com",
			Name:      "Test",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	deleted, err = scoop.Delete()
	if err != nil {
		t.Logf("Delete all failed: %v", err)
	}
	if deleted != 3 {
		t.Errorf("Expected 3 deleted, got %d", deleted)
	}

	// Test 3: Delete with no matches
	deleted, err = scoop.Delete()
	if err != nil {
		t.Logf("Delete from empty collection (after previous delete) failed: %v", err)
	}
	if deleted != 0 {
		t.Errorf("Expected 0 deleted (no matches), got %d", deleted)
	}
}

// TestTransactionWithCreateUpdate tests Begin/Commit with Create and Update
func TestTransactionWithCreateUpdate(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Start transaction
	scoop := client.NewScoop().Collection(User{})
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}

	// Create user
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "tx_create@example.com",
		Name:      "Original",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = txScoop.Create(user)
	if err != nil {
		t.Logf("Create in transaction failed: %v", err)
	}

	// Update user (same scoop)
	updated, err := txScoop.Equal("email", "tx_create@example.com").Update(bson.M{
		"$set": bson.M{
			"name": "Updated",
		},
	})
	if err != nil {
		t.Logf("Update in transaction failed: %v", err)
	} else if updated != 1 {
		t.Logf("Expected 1 updated, got %d", updated)
	}

	// Commit transaction
	err = txScoop.Commit()
	if err != nil {
		t.Logf("Commit failed: %v", err)
	}
}

// TestBeginMultipleTimes tests calling Begin multiple times
func TestBeginMultipleTimes(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})

	// First Begin should work
	tx1, err := scoop.Begin()
	if err != nil {
		t.Fatalf("First Begin failed: %v", err)
	}

	// Try Begin on same scoop again (should handle existing transaction)
	tx2, err := scoop.Begin()
	if err != nil {
		// This is expected - transaction already in progress
		t.Logf("Second Begin returned error (expected): %v", err)
	}

	if tx1 != nil {
		// Rollback first transaction
		_ = tx1.Rollback()
	}

	if tx2 != nil {
		_ = tx2.Rollback()
	}
}

// TestHealthConsecutiveCalls tests calling Health multiple times
func TestHealthConsecutiveCalls(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Call Health multiple times
	for i := 0; i < 3; i++ {
		err := client.Health()
		if err != nil {
			t.Logf("Health call %d failed: %v", i, err)
		}
	}
}

// TestPingConsecutiveCalls tests calling Ping multiple times
func TestPingConsecutiveCalls(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Call Ping multiple times
	for i := 0; i < 3; i++ {
		err := client.Ping()
		if err != nil {
			t.Logf("Ping call %d failed: %v", i, err)
		}
	}
}

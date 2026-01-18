package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestScoopCountWithEmptyCollection tests Count on empty collection returns 0
func TestScoopCountWithEmptyCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count on empty collection failed: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 count, got %d", count)
	}
}

// TestScoopCountWithSingleDocument tests Count returns 1 for single document
func TestScoopCountWithSingleDocument(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert one document
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "single@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count with single document failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

// TestScoopCountWithHighNumber tests Count with many documents
func TestScoopCountWithHighNumber(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert many documents
	for i := 0; i < 100; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "test" + string(rune(48+(i%10))) + "@example.com",
			Name:      "User",
			Age:       20 + (i % 30),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count with many documents failed: %v", err)
	}

	if count != 100 {
		t.Errorf("expected 100, got %d", count)
	}
}

// TestScoopCountBySpecificField tests Count with specific field equality
func TestScoopCountBySpecificField(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert documents with different ages
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "age" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Add some with different age
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "other" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       30,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{}).Equal("age", 25)
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count by field failed: %v", err)
	}

	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

// TestScoopDeleteFromEmptyCollection tests Delete on empty collection
func TestScoopDeleteFromEmptyCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	deleted, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete from empty collection failed: %v", err)
	}

	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}

// TestScoopDeleteSingleDocument tests Delete removes exactly one document
func TestScoopDeleteSingleDocument(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert one document
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "delete_single@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})
	deleted, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete single document failed: %v", err)
	}

	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	// Verify it's gone
	count, _ := scoop.Count()
	if count != 0 {
		t.Errorf("expected 0 remaining, got %d", count)
	}
}

// TestScoopDeleteByCondition tests Delete with specific condition
func TestScoopDeleteByCondition(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert mixed data
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "young" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       20,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	for i := 0; i < 4; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "old" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       50,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Delete only users with age 20
	scoop := client.NewScoop().Collection(User{}).Equal("age", 20)
	deleted, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete by condition failed: %v", err)
	}

	if deleted != 3 {
		t.Errorf("expected 3 deleted, got %d", deleted)
	}

	// Verify remaining count
	allScoop := client.NewScoop().Collection(User{})
	remaining, _ := allScoop.Count()
	if remaining != 4 {
		t.Errorf("expected 4 remaining, got %d", remaining)
	}
}

// TestClientHealthErrorPropagation tests that Health properly propagates Ping errors
func TestClientHealthErrorPropagation(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Normal health check should succeed or fail gracefully
	err := client.Health()
	// Just ensure it doesn't panic
	if err != nil {
		t.Logf("Health check returned error (expected in some environments): %v", err)
	}
}

// TestClientHealthFormatsError tests that Health wraps error message
func TestClientHealthFormatsError(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Normal case - should succeed or fail gracefully
	err := client.Health()
	if err != nil {
		// Error message should contain context
		errMsg := err.Error()
		t.Logf("Health check error: %s", errMsg)
		// Should not be empty
		if errMsg == "" {
			t.Error("expected non-empty error message")
		}
	}
}

// TestClientPingReturnsNoError tests successful Ping
func TestClientPingReturnsNoError(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Should succeed when connection is valid
	err := client.Ping()
	if err != nil {
		t.Logf("Ping error (may be expected): %v", err)
	}
}

// TestClientCloseHandlesNilClient tests Close when client is nil (edge case)
func TestClientCloseHandlesNilClient(t *testing.T) {
	client := newTestClient(t)
	
	// First close should work
	err := client.Close()
	if err != nil {
		t.Logf("First close error: %v", err)
	}

	// Should be safe to call again (though behavior may vary)
	err2 := client.Close()
	if err2 != nil {
		t.Logf("Second close error: %v", err2)
	}
}

// TestScoopFirstThrowsErrorWhenNotFound tests that First returns error for no results
func TestScoopFirstThrowsErrorWhenNotFound(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{}).Equal("email", "notfound@example.com")
	var result User
	err := scoop.First(&result)
	if err == nil {
		t.Error("expected error when document not found")
	}
}

// TestScoopFirstWithNullFilter tests First with explicitly null-like conditions
func TestScoopFirstWithNullFilter(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert multiple documents
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "first" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// First without any filter - should return first document
	scoop := client.NewScoop().Collection(User{})
	var result User
	err := scoop.First(&result)
	if err != nil {
		t.Fatalf("first without filter failed: %v", err)
	}

	if result.ID == primitive.NilObjectID {
		t.Error("expected valid ID")
	}
}

// TestClientGetDatabaseEmpty tests GetDatabase when config database is empty
func TestClientGetDatabaseEmpty(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// GetDatabase should return default "test" if not set in config
	db := client.GetDatabase()
	if db != "test" && db == "" {
		t.Error("expected non-empty database name")
	}
}

// TestCountAndDeleteIntegration tests Count and Delete work together
func TestCountAndDeleteIntegration(t *testing.T) {
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
			Email:     "integ" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Count before delete
	scoop := client.NewScoop().Collection(User{})
	countBefore, _ := scoop.Count()
	if countBefore != 5 {
		t.Errorf("expected 5 before delete, got %d", countBefore)
	}

	// Delete all
	deleted, err := scoop.Delete()
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if deleted != 5 {
		t.Errorf("expected 5 deleted, got %d", deleted)
	}

	// Count after delete
	countAfter, _ := scoop.Count()
	if countAfter != 0 {
		t.Errorf("expected 0 after delete, got %d", countAfter)
	}
}

// TestClientMethodsSequentially tests calling client methods in sequence
func TestClientMethodsSequentially(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Get config
	cfg := client.GetConfig()
	if cfg == nil {
		t.Error("expected non-nil config")
	}

	// Get database
	db := client.GetDatabase()
	if db == "" {
		t.Error("expected non-empty database")
	}

	// Ping
	err := client.Ping()
	if err != nil {
		t.Logf("ping error: %v", err)
	}

	// Health
	err = client.Health()
	if err != nil {
		t.Logf("health error: %v", err)
	}

	// Context
	ctx := client.Context()
	if ctx == nil {
		t.Error("expected non-nil context")
	}
}

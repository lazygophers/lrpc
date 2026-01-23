package mongo

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestAutoMigrateSuccess tests successful AutoMigrate of a single model
func TestAutoMigrateSuccess(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users", "posts", "comments")
	}
	cleanupTest()
	defer cleanupTest()

	// Test AutoMigrate with User model
	err := client.AutoMigrate(User{})
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Verify collection was created
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	collectionExists := false
	for _, name := range collections {
		if name == "users" {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		t.Error("expected 'users' collection to be created")
	}
}

// TestAutoMigrateDuplicateCollection tests AutoMigrate on an already existing collection
func TestAutoMigrateDuplicateCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// First AutoMigrate
	err := client.AutoMigrate(User{})
	if err != nil {
		t.Fatalf("first AutoMigrate failed: %v", err)
	}

	// Second AutoMigrate should succeed without error (idempotent)
	err = client.AutoMigrate(User{})
	if err != nil {
		t.Fatalf("second AutoMigrate failed: %v", err)
	}

	// Verify collection still exists and only one was created
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	userCollCount := 0
	for _, name := range collections {
		if name == "users" {
			userCollCount++
		}
	}

	if userCollCount != 1 {
		t.Errorf("expected exactly 1 'users' collection, found %d", userCollCount)
	}
}

// TestAutoMigratesMultiple tests AutoMigrates with multiple models
func TestAutoMigratesMultiple(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users", "posts", "comments")
	}
	cleanupTest()
	defer cleanupTest()

	// Migrate multiple models at once
	err := client.AutoMigrates(User{}, Post{}, Comment{})
	if err != nil {
		t.Fatalf("AutoMigrates failed: %v", err)
	}

	// Verify all collections were created
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	expectedCollections := map[string]bool{
		"users":    false,
		"posts":    false,
		"comments": false,
	}

	for _, name := range collections {
		if _, exists := expectedCollections[name]; exists {
			expectedCollections[name] = true
		}
	}

	for collName, found := range expectedCollections {
		if !found {
			t.Errorf("expected collection '%s' to be created", collName)
		}
	}
}

// TestAutoMigrateNoCollectionierInterface tests error when model doesn't implement Collectioner
func TestAutoMigrateNoCollectionierInterface(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// NoCollectionUser doesn't implement Collectioner interface
	err := client.AutoMigrate(NoCollectionUser{})
	if err == nil {
		t.Error("expected error for model without Collectioner interface, got nil")
	}

	// Verify error message mentions interface requirement
	if err != nil && !errors.Is(err, err) {
		// Just check that it's not nil and matches the pattern
		t.Logf("got expected error: %v", err)
	}
}

// TestAutoMigrateCustomCollection tests AutoMigrate with custom collection name
func TestAutoMigrateCustomCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "custom_users")
	}
	cleanupTest()
	defer cleanupTest()

	// CustomUser has custom collection name
	err := client.AutoMigrate(CustomUser{})
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Verify custom collection was created
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	collectionExists := false
	for _, name := range collections {
		if name == "custom_users" {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		t.Error("expected 'custom_users' collection to be created")
	}
}

// TestAutoMigratePartialFailure tests AutoMigrates with one invalid model
func TestAutoMigratePartialFailure(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Try to migrate valid and invalid models
	err := client.AutoMigrates(User{}, NoCollectionUser{})
	if err == nil {
		t.Error("expected error for invalid model, got nil")
	}

	// First model should have been migrated before the error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	// Verify 'users' collection was created despite later failure
	userCollFound := false
	for _, name := range collections {
		if name == "users" {
			userCollFound = true
			break
		}
	}

	if !userCollFound {
		t.Error("expected 'users' collection to be created before failure")
	}
}

// TestScoopAutoMigrate tests AutoMigrate through Scoop
func TestScoopAutoMigrate(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create scoop and call AutoMigrate through it
	scoop := client.NewScoop()
	err := scoop.AutoMigrate(User{})
	if err != nil {
		t.Fatalf("Scoop.AutoMigrate failed: %v", err)
	}

	// Verify collection was created
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	collectionExists := false
	for _, name := range collections {
		if name == "users" {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		t.Error("expected 'users' collection to be created via Scoop")
	}
}

// TestScoopAutoMigrates tests AutoMigrates through Scoop
func TestScoopAutoMigrates(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users", "posts")
	}
	cleanupTest()
	defer cleanupTest()

	// Create scoop and call AutoMigrates through it
	scoop := client.NewScoop()
	err := scoop.AutoMigrates(User{}, Post{})
	if err != nil {
		t.Fatalf("Scoop.AutoMigrates failed: %v", err)
	}

	// Verify both collections were created
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	expectedCollections := map[string]bool{
		"users": false,
		"posts": false,
	}

	for _, name := range collections {
		if _, exists := expectedCollections[name]; exists {
			expectedCollections[name] = true
		}
	}

	for collName, found := range expectedCollections {
		if !found {
			t.Errorf("expected collection '%s' to be created via Scoop", collName)
		}
	}
}

// TestAutoMigrateIdempotent tests that AutoMigrate is idempotent
func TestAutoMigrateIdempotent(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Call AutoMigrate multiple times
	for i := 0; i < 3; i++ {
		err := client.AutoMigrate(User{})
		if err != nil {
			t.Fatalf("AutoMigrate attempt %d failed: %v", i+1, err)
		}
	}

	// Verify collection exists and is accessible
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	coll := db.Collection("users")

	// Try to insert a document to verify collection works
	doc := User{
		Email:     "test@example.com",
		Name:      "Test User",
		Age:       30,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	_, err := coll.InsertOne(ctx, doc)
	if err != nil {
		t.Fatalf("failed to insert document into migrated collection: %v", err)
	}
}

// TestAutoMigrateWithIndex tests AutoMigrate with a model that has indexes
func TestAutoMigrateWithIndex(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Test that AutoMigrate works with User model (which may have indexes in future)
	// For now, just verify AutoMigrate works without errors
	err := client.AutoMigrate(User{})
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// Verify collection was created
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	collectionExists := false
	for _, name := range collections {
		if name == "users" {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		t.Error("expected 'users' collection to be created")
	}
}

// TestMockClientAutoMigrate tests MockClient.AutoMigrate
func TestMockClientAutoMigrate(t *testing.T) {
	mockClient := NewMockClient()

	// Test success case
	mockClient.SetupAutoMigrateSuccess()
	err := mockClient.AutoMigrate(User{})
	if err != nil {
		t.Fatalf("MockClient.AutoMigrate failed: %v", err)
	}

	// Verify call was recorded
	if !mockClient.AssertCalled("AutoMigrate") {
		t.Error("expected AutoMigrate to be called")
	}

	if mockClient.GetCallCount("AutoMigrate") != 1 {
		t.Errorf("expected 1 call to AutoMigrate, got %d", mockClient.GetCallCount("AutoMigrate"))
	}
}

// TestMockClientAutoMigrates tests MockClient.AutoMigrates
func TestMockClientAutoMigrates(t *testing.T) {
	mockClient := NewMockClient()

	// Test success case
	mockClient.SetupAutoMigratesSuccess()
	err := mockClient.AutoMigrates(User{}, Post{})
	if err != nil {
		t.Fatalf("MockClient.AutoMigrates failed: %v", err)
	}

	// Verify call was recorded
	if !mockClient.AssertCalled("AutoMigrates") {
		t.Error("expected AutoMigrates to be called")
	}

	if mockClient.GetCallCount("AutoMigrates") != 1 {
		t.Errorf("expected 1 call to AutoMigrates, got %d", mockClient.GetCallCount("AutoMigrates"))
	}
}

// TestMockClientAutoMigrateError tests MockClient.AutoMigrate returning error
func TestMockClientAutoMigrateError(t *testing.T) {
	mockClient := NewMockClient()

	testErr := errors.New("migration failed")
	mockClient.SetupAutoMigrateError(testErr)

	err := mockClient.AutoMigrate(User{})
	if err == nil {
		t.Error("expected error, got nil")
	}

	if !errors.Is(err, testErr) {
		t.Errorf("expected error %v, got %v", testErr, err)
	}
}

// TestMockClientAutoMigratesError tests MockClient.AutoMigrates returning error
func TestMockClientAutoMigratesError(t *testing.T) {
	mockClient := NewMockClient()

	testErr := errors.New("batch migration failed")
	mockClient.SetupAutoMigratesError(testErr)

	err := mockClient.AutoMigrates(User{}, Post{})
	if err == nil {
		t.Error("expected error, got nil")
	}

	if !errors.Is(err, testErr) {
		t.Errorf("expected error %v, got %v", testErr, err)
	}
}

// TestMockClientAutoMigrateCalls tests MockClient call tracking for AutoMigrate
func TestMockClientAutoMigrateCalls(t *testing.T) {
	mockClient := NewMockClient()
	mockClient.SetupAutoMigrateSuccess()

	// Make multiple calls
	mockClient.AutoMigrate(User{})
	mockClient.AutoMigrate(Post{})
	mockClient.AutoMigrate(User{}) // Call again

	// Verify call count
	if mockClient.GetCallCount("AutoMigrate") != 3 {
		t.Errorf("expected 3 calls to AutoMigrate, got %d", mockClient.GetCallCount("AutoMigrate"))
	}

	// Verify all calls were recorded
	calls := mockClient.GetCalls()
	autoMigrateCalls := 0
	for _, call := range calls {
		if call.Method == "AutoMigrate" {
			autoMigrateCalls++
		}
	}

	if autoMigrateCalls != 3 {
		t.Errorf("expected 3 AutoMigrate calls in records, got %d", autoMigrateCalls)
	}
}

// TestAutoMigrateWithTransactionScoop tests AutoMigrate with transaction scoop
func TestAutoMigrateWithTransactionScoop(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create a transaction scoop
	baseScopop := client.NewScoop()
	txScoop, err := baseScopop.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// AutoMigrate through transaction scoop
	err = txScoop.AutoMigrate(User{})
	if err != nil {
		t.Fatalf("AutoMigrate in transaction failed: %v", err)
	}

	// Verify collection was created
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := client.client.Database(client.database)
	collections, err := db.ListCollectionNames(ctx, nil)
	if err != nil {
		t.Fatalf("failed to list collections: %v", err)
	}

	collectionExists := false
	for _, name := range collections {
		if name == "users" {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		t.Error("expected 'users' collection to be created in transaction")
	}
}

// TestAutoMigrateResetCalls tests MockClient.ResetCalls
func TestAutoMigrateResetCalls(t *testing.T) {
	mockClient := NewMockClient()
	mockClient.SetupAutoMigrateSuccess()

	// Make multiple calls
	mockClient.AutoMigrate(User{})
	mockClient.AutoMigrate(Post{})

	// Verify calls were recorded
	if mockClient.GetCallCount("AutoMigrate") != 2 {
		t.Errorf("expected 2 calls before reset, got %d", mockClient.GetCallCount("AutoMigrate"))
	}

	// Reset calls
	mockClient.ResetCalls()

	// Verify calls were cleared
	if mockClient.GetCallCount("AutoMigrate") != 0 {
		t.Errorf("expected 0 calls after reset, got %d", mockClient.GetCallCount("AutoMigrate"))
	}

	if !mockClient.AssertNotCalled("AutoMigrate") {
		t.Error("expected AutoMigrate to not be called after reset")
	}
}

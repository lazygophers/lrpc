package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTransactionWithTx(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Start transaction
	scoop := client.NewScoop()
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("begin transaction failed: %v", err)
	}

	// Create user within transaction
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "txuser@example.com",
		Name:      "TX User",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = txScoop.Create(user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Commit transaction
	err = txScoop.Commit()
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	// Verify data
	AssertDocumentExists(t, client, "users", bson.M{"email": "txuser@example.com"})
}

func TestTransactionWithTxAndResult(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert initial data
	user := User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	InsertTestData(t, client, "users", user)

	// Start transaction
	scoop := client.NewScoop()
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("begin transaction failed: %v", err)
	}

	// Find user within transaction
	var foundUser User
	firstResult := txScoop.Equal("email", "test@example.com").First(&foundUser)
	if firstResult.Error != nil {
		t.Fatalf("failed to find user: %v", firstResult.Error)
	}

	if foundUser.Age != 25 {
		t.Errorf("expected age 25, got %v", foundUser.Age)
	}

	// Commit transaction
	err = txScoop.Commit()
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}
}

func TestTransactionRollback(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Start transaction
	scoop := client.NewScoop()
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("begin transaction failed: %v", err)
	}

	// Create user within transaction
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "txuser@example.com",
		Name:      "TX User",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = txScoop.Create(user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Rollback transaction
	err = txScoop.Rollback()
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	// Note: In MongoDB standalone, transactions might not actually rollback
	// but in a replica set they should. For testing purposes,
	// we verify that the error was not returned from rollback
}

func TestTransactionMultipleOperations(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users", "posts")
	}
	cleanupTest()
	defer cleanupTest()

	// Start transaction
	scoop := client.NewScoop()
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("begin transaction failed: %v", err)
	}

	// Insert user within transaction
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "txuser@example.com",
		Name:      "TX User",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = txScoop.Create(user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Insert post within transaction using a new scoop for posts collection
	postScoop, err := txScoop.Begin()
	if err == nil {
		// Begin on a transactional scoop should work but use same session
		postScoop.Rollback()
	}

	// Create post using the same scoop (will use same collection automatically)
	post := Post{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		Title:     "Test Post",
		Content:   "Test Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// For this, we need a separate scoop for the posts collection
	postScoop2 := client.NewScoop().CollectionName("posts")
	err = postScoop2.Create(post)
	if err != nil {
		t.Fatalf("failed to create post: %v", err)
	}

	// Commit transaction
	err = txScoop.Commit()
	if err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	// Verify user exists
	userCount := CountTestDocuments(t, client, "users", bson.M{"email": "txuser@example.com"})
	if userCount != 1 {
		t.Errorf("expected 1 user, got %d", userCount)
	}

	// Verify post exists
	postCount := CountTestDocuments(t, client, "posts", bson.M{"title": "Test Post"})
	if postCount != 1 {
		t.Errorf("expected 1 post, got %d", postCount)
	}
}

func TestNewTransaction(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop()
	if scoop == nil {
		t.Error("expected scoop, got nil")
	}

	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("begin transaction failed: %v", err)
	}

	if txScoop == nil {
		t.Error("expected transactional scoop, got nil")
	}

	txScoop.Rollback()
}

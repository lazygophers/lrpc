package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCommitWithoutTransaction tests Commit when no transaction is active
func TestCommitWithoutTransaction(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})

	// Try to commit without starting a transaction
	err := scoop.Commit()
	if err == nil {
		t.Error("expected error when committing without transaction, got nil")
	}
	if err.Error() != "no active transaction" {
		t.Errorf("expected 'no active transaction' error, got '%v'", err)
	}
}

// TestRollbackWithoutTransaction tests Rollback when no transaction is active
func TestRollbackWithoutTransaction(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})

	// Try to rollback without starting a transaction
	err := scoop.Rollback()
	if err == nil {
		t.Error("expected error when rolling back without transaction, got nil")
	}
	if err.Error() != "no active transaction" {
		t.Errorf("expected 'no active transaction' error, got '%v'", err)
	}
}

// TestBeginTransactionSuccess tests successful transaction start
func TestBeginTransactionSuccess(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Start a transaction
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	if txScoop == nil {
		t.Error("expected transaction scoop, got nil")
	}

	// Rollback the transaction
	err = txScoop.Rollback()
	if err != nil {
		t.Fatalf("failed to rollback transaction: %v", err)
	}
}

// TestTransactionWithInsertAndRollback tests inserting within a transaction and rolling back
func TestTransactionWithInsertAndRollback(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Start a transaction
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// Insert a document within the transaction
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "tx@example.com",
		Name:      "Transaction User",
		Age:       25,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err = txScoop.Create(user)
	if err != nil {
		t.Fatalf("failed to create user in transaction: %v", err)
	}

	// Rollback the transaction
	err = txScoop.Rollback()
	if err != nil {
		t.Fatalf("failed to rollback transaction: %v", err)
	}

	// Use a new scoop to verify the document was not persisted (rollback worked)
	countScoop := client.NewScoop().Collection(User{})
	count, err := countScoop.Count()
	if err != nil {
		t.Fatalf("failed to count documents: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 documents after rollback, got %d", count)
	}
}

// TestTransactionWithInsertAndCommit tests inserting within a transaction and committing
func TestTransactionWithInsertAndCommit(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Start a transaction
	txScoop, err := scoop.Begin()
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// Insert a document within the transaction
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "commit@example.com",
		Name:      "Commit User",
		Age:       30,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err = txScoop.Create(user)
	if err != nil {
		t.Fatalf("failed to create user in transaction: %v", err)
	}

	// Commit the transaction
	err = txScoop.Commit()
	if err != nil {
		t.Fatalf("failed to commit transaction: %v", err)
	}

	// Use a new scoop to verify the document was persisted (after commit, session is closed)
	countScoop := client.NewScoop().Collection(User{})
	count, err := countScoop.Count()
	if err != nil {
		t.Fatalf("failed to count documents: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 document after commit, got %d", count)
	}

	// Verify the document has correct data
	var foundUser User
	firstResult := countScoop.Where("email", "commit@example.com").First(&foundUser)
	if firstResult.Error != nil {
		t.Fatalf("failed to find user: %v", firstResult.Error)
	}
	if foundUser.Name != "Commit User" {
		t.Errorf("expected name 'Commit User', got '%s'", foundUser.Name)
	}
}

// TestWhereWithNoFilters tests Where with empty parameters
func TestWhereWithNoFilters(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop()
	result := scoop.Where()

	if result != scoop {
		t.Error("Where with no parameters should return the same Scoop instance")
	}
}

// TestNewScoopFromClient tests creating Scoop from Client.NewScoop
func TestNewScoopFromClient(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop()
	if scoop == nil {
		t.Error("expected scoop, got nil")
	}
}

// TestNewScoopWithCollection tests NewScoop with a specific collection
func TestNewScoopWithCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})
	if scoop == nil {
		t.Error("expected scoop, got nil")
	}
}

// TestFirstWithoutCollection tests First when collection is not set
func TestFirstWithoutCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop()

	// Create a test User struct to pass to First
	var user User

	// Try to call First without setting collection
	err := scoop.First(&user)
	if err == nil {
		t.Error("expected error when calling First without collection, got nil")
	}
}

// TestWhereAndFirstChaining tests chaining Where and First
func TestWhereAndFirstChaining(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create test documents
	users := []User{
		{ID: primitive.NewObjectID(), Email: "alice@example.com", Name: "Alice", Age: 25, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		{ID: primitive.NewObjectID(), Email: "bob@example.com", Name: "Bob", Age: 30, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		{ID: primitive.NewObjectID(), Email: "charlie@example.com", Name: "Charlie", Age: 35, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
	}

	for _, user := range users {
		scoop.Create(user)
	}

	// Test Where and First chaining
	var found User
	err := scoop.Where("age", map[string]interface{}{"$gte": 30}).First(&found)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	// Should get Bob or Charlie
	if found.Age < 30 {
		t.Errorf("expected age >= 30, got %d", found.Age)
	}
}

// TestScoopNewScoopInitialization tests NewScoop initializes properly
func TestScoopNewScoopInitialization(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop()
	if scoop == nil {
		t.Error("expected scoop, got nil")
	}
}

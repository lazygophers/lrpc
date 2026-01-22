package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestUpdateWithBsonM tests Update with bson.M input
func TestUpdateWithBsonM(t *testing.T) {
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

	// Update with bson.M
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("email", "test@example.com")

	updateData := bson.M{"age": 35, "name": "Updated Name"}
	updateResult := scoop.Update(updateData
	count, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 updated document, got %d", count)
	}

	// Verify
	doc := GetTestDocument(t, client, "users", bson.M{"email": "test@example.com"})
	if age, ok := doc["age"].(int32); !ok || age != 35 {
		t.Errorf("expected age 35, got %v", doc["age"])
	}
}

// TestUpdateWithMapOperators tests Update with map containing operators
func TestUpdateWithMapOperators(t *testing.T) {
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

	// Update with map operators ($set)
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("email", "test@example.com")

	updateData := map[string]interface{}{"$set": map[string]interface{}{"age": 40}}
	updateResult := scoop.Update(updateData
	count, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 updated document, got %d", count)
	}

	// Verify
	doc := GetTestDocument(t, client, "users", bson.M{"email": "test@example.com"})
	if age, ok := doc["age"].(int32); !ok || age != 40 {
		t.Errorf("expected age 40, got %v", doc["age"])
	}
}

// TestUpdateWithMapNoOperators tests Update with plain map (no operators)
func TestUpdateWithMapNoOperators(t *testing.T) {
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

	// Update with plain map
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("email", "test@example.com")

	updateData := map[string]interface{}{"age": 45, "name": "Another Update"}
	updateResult := scoop.Update(updateData
	count, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 updated document, got %d", count)
	}

	// Verify
	doc := GetTestDocument(t, client, "users", bson.M{"email": "test@example.com"})
	if age, ok := doc["age"].(int32); !ok || age != 45 {
		t.Errorf("expected age 45, got %v", doc["age"])
	}
}

// TestUpdateWithoutCollection tests Update error when collection not set
func TestUpdateWithoutCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Create scoop without setting collection
	scoop := client.NewScoop()

	updateResult := scoop.Update(bson.M{"age": 30}
	count, err := updateResult.DocsAffected, updateResult.Error
	if err == nil {
		t.Error("expected error when collection not set")
	}

	if count != 0 {
		t.Errorf("expected 0 updated documents on error, got %d", count)
	}
}

// TestUpdateMultipleDocuments tests Update affecting multiple documents
func TestUpdateMultipleDocuments(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert multiple test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Update multiple documents
	scoop := client.NewScoop()
	scoop = scoop.Collection(User{})
	scoop = scoop.Equal("age", 25)

	updateResult := scoop.Update(bson.M{"status": "updated"}
	count, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 updated documents, got %d", count)
	}
}

// TestUpdateZeroResults tests Update when no documents match
func TestUpdateZeroResults(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().
		Collection(User{}).
		Equal("email", "nonexistent@example.com")

	updateResult := scoop.Update(bson.M{"age": 99}
	count, err := updateResult.DocsAffected, updateResult.Error
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 updated documents, got %d", count)
	}
})

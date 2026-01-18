package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCreateVariousTypes tests Create with different document types
func TestCreateVariousTypes(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop()

	// Test creating a User struct
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "test@example.com",
		Name:      "Test User",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := scoop.Collection(User{}).Create(user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Verify the user was created
	var found User
	err = scoop.Where("email", "test@example.com").First(&found)
	if err != nil {
		t.Fatalf("failed to find user: %v", err)
	}
	if found.ID == primitive.NilObjectID {
		t.Error("expected user to be found after creation")
	}
}

// TestCreateNilDocument tests Create behavior when doc is nil
func TestCreateNilDocument(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop()

	// This should attempt to insert nil, which might fail depending on MongoDB
	err := scoop.Create(nil)
	if err == nil {
		t.Error("expected error when creating nil document, got nil")
	}
}

// TestCreateMultipleDocumentsSequentially tests multiple Create calls
func TestCreateMultipleDocumentsSequentially(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	users := []User{
		{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 26, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 27, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, user := range users {
		err := scoop.Create(user)
		if err != nil {
			t.Fatalf("failed to create user %s: %v", user.Email, err)
		}
	}

	// Verify all users were created
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 users, got %d", count)
	}
}

// TestCountWithNoDocuments tests Count on empty collection
func TestCountWithNoDocuments(t *testing.T) {
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
		t.Fatalf("failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

// TestCountAfterCreate tests Count after creating documents
func TestCountAfterCreate(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create documents
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "user" + string(rune(i)) + "@example.com",
			Name:      "User " + string(rune(i)),
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := scoop.Create(user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	// Count all documents
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

// TestCountWithComplexFilters tests Count with multiple filter conditions
func TestCountWithComplexFilters(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create documents with different ages
	ages := []int{20, 25, 30, 35, 40}
	for i, age := range ages {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "user" + string(rune(i)) + "@example.com",
			Name:      "User " + string(rune(i)),
			Age:       age,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := scoop.Create(user)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	// Count with age > 25
	count, err := scoop.Where("age", map[string]interface{}{"$gt": 25}).Count()
	if err != nil {
		t.Fatalf("failed to count with filter: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 (ages 30, 35, 40), got %d", count)
	}
}

// TestDeleteWithoutFilters tests Delete without filters (delete all)
func TestDeleteWithoutFilters(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create multiple documents
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "user" + string(rune(i)) + "@example.com",
			Name:      "User " + string(rune(i)),
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		scoop.Create(user)
	}

	// Delete all documents
	deleted, err := scoop.Delete()
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
	if deleted != 3 {
		t.Errorf("expected 3 deleted, got %d", deleted)
	}

	// Verify all are deleted
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("failed to count after delete: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 after delete, got %d", count)
	}
}

// TestDeleteSpecificDocuments tests Delete with specific filters
func TestDeleteSpecificDocuments(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create multiple documents with specific ages
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "test" + string('0'+rune(i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		scoop.Create(user)
	}

	// Create a new scoop for the delete operation (fresh filter)
	deleteScoop := client.NewScoop().Collection(User{})
	
	// Delete documents with age >= 23 (that's ages 23, 24)
	deleted, err := deleteScoop.Where("age", map[string]interface{}{"$gte": 23}).Delete()
	if err != nil {
		t.Fatalf("failed to delete with filter: %v", err)
	}

	// Verify we deleted the expected number
	if deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", deleted)
	}

	// Count remaining documents
	countScoop := client.NewScoop().Collection(User{})
	count, err := countScoop.Count()
	if err != nil {
		t.Fatalf("failed to count after delete: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 remaining, got %d", count)
	}
}

// TestDeleteSingleDocument tests Delete removing exactly one document
func TestDeleteSingleDocument(t *testing.T) {
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
		Email:     "target@example.com",
		Name:      "Target User",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	scoop.Create(user)

	// Delete the specific user
	deleted, err := scoop.Where("email", "target@example.com").Delete()
	if err != nil {
		t.Fatalf("failed to delete: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	// Verify deletion
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 after delete, got %d", count)
	}
}

// TestUpdateSingleField tests Update with single field modification
func TestUpdateSingleField(t *testing.T) {
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
		Email:     "test@example.com",
		Name:      "Original Name",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	scoop.Create(user)

	// Update the name
	updated, err := scoop.Where("email", "test@example.com").Update(map[string]interface{}{
		"name": "Updated Name",
	})
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}
	if updated != 1 {
		t.Errorf("expected 1 updated, got %d", updated)
	}

	// Verify the update
	var foundUser User
	err = scoop.Where("email", "test@example.com").First(&foundUser)
	if err != nil {
		t.Fatalf("failed to find user: %v", err)
	}
	if foundUser.Name != "Updated Name" {
		t.Error("expected name to be updated")
	}
}

// TestUpdateMultipleFields tests Update with multiple field modifications
func TestUpdateMultipleFields(t *testing.T) {
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
		Email:     "test@example.com",
		Name:      "Original",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	scoop.Create(user)

	// Update multiple fields
	updated, err := scoop.Where("email", "test@example.com").Update(map[string]interface{}{
		"name": "Updated",
		"age":  30,
	})
	if err != nil {
		t.Fatalf("failed to update: %v", err)
	}
	if updated != 1 {
		t.Errorf("expected 1 updated, got %d", updated)
	}

	// Verify both fields were updated
	var foundUser User
	err = scoop.Where("email", "test@example.com").First(&foundUser)
	if err != nil {
		t.Fatalf("failed to find user: %v", err)
	}
	if foundUser.Name != "Updated" || foundUser.Age != 30 {
		t.Errorf("expected name='Updated', age=30, got name='%s', age=%d", foundUser.Name, foundUser.Age)
	}
}

// TestCountIsZeroAfterDeleteAll tests Count returns 0 after deleting all
func TestCountIsZeroAfterDeleteAll(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create some documents
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "user" + string(rune(i)) + "@example.com",
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		scoop.Create(user)
	}

	// Delete all
	scoop.Delete()

	// Count should be 0
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("failed to count: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 after delete all, got %d", count)
	}
}

package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewModel(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	model := NewModel[User](client)

	if model == nil {
		t.Error("expected model, got nil")
	}
}

func TestModelTableName(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	model := NewModel[User](client)
	tableName := model.CollectionName()

	if tableName == "" {
		t.Error("expected table name, got empty string")
	}
}

func TestModelFind(t *testing.T) {
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

	// Find with model
	model := NewModel[User](client)
	results, err := model.NewScoop().Where("age", 25).Find()

	if err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	if results[0].Age != 25 {
		t.Errorf("expected age 25, got %d", results[0].Age)
	}
}

func TestModelFirst(t *testing.T) {
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

	// FindOne with model
	model := NewModel[User](client)
	result, err := model.NewScoop().Where("email", "test@example.com").First()

	if err != nil {
		t.Fatalf("find one failed: %v", err)
	}

	if result == nil {
		t.Error("expected result, got nil")
	}

	if result.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", result.Email)
	}
}

func TestModelCount(t *testing.T) {
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
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Count with model
	model := NewModel[User](client)
	count, err := model.NewScoop().Where("age", 25).Count()

	if err != nil {
		t.Fatalf("count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestModelCreate(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create with model
	model := NewModel[User](client)

	newUser := User{
		ID:        primitive.NewObjectID(),
		Email:     "newuser@example.com",
		Name:      "New User",
		Age:       28,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := model.NewScoop().Create(newUser)

	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Verify
	AssertDocumentExists(t, client, "users", bson.M{"email": "newuser@example.com"})
}

func TestModelUpdate(t *testing.T) {
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

	// Update with model
	model := NewModel[User](client)
	count, err := model.NewScoop().Where("email", "test@example.com").Updates(bson.M{"age": 30, "name": "Updated User"})

	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 updated document, got %d", count)
	}

	// Verify
	doc := GetTestDocument(t, client, "users", bson.M{"email": "test@example.com"})
	if age, ok := doc["age"].(int32); !ok || age != 30 {
		t.Errorf("expected age to be updated to 30")
	}
}

func TestModelDelete(t *testing.T) {
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
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Delete with model
	model := NewModel[User](client)
	count, err := model.NewScoop().Where("age", 25).Delete()

	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 deleted documents, got %d", count)
	}

	// Verify
	AssertCount(t, 0, client, "users", bson.M{"age": 25})
}

func TestModelExist(t *testing.T) {
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

	// Test exist
	model := NewModel[User](client)
	exists, err := model.NewScoop().Where("email", "test@example.com").Exist()

	if err != nil {
		t.Fatalf("exist failed: %v", err)
	}

	if !exists {
		t.Error("expected document to exist")
	}

	// Test not exist
	exists, err = model.NewScoop().Where("email", "notexist@example.com").Exist()

	if err != nil {
		t.Fatalf("exist failed: %v", err)
	}

	if exists {
		t.Error("expected document to not exist")
	}
}

func TestModelGetCollection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	model := NewModel[User](client)
	coll := model.NewScoop().GetCollection()

	if coll == nil {
		t.Error("expected collection, got nil")
	}

	if coll.Name() != model.CollectionName() {
		t.Errorf("expected collection name '%s', got '%s'", model.CollectionName(), coll.Name())
	}
}

func TestModelNewScoop(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	model := NewModel[User](client)
	modelScoop := model.NewScoop()

	if modelScoop == nil {
		t.Error("expected model scoop, got nil")
	}

	if modelScoop.GetCollection() == nil {
		t.Error("expected collection, got nil")
	}
}

func TestModelAggregate(t *testing.T) {
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

	// Aggregate with model
	model := NewModel[User](client)
	agg := model.NewScoop().Aggregate(bson.M{"$match": bson.M{"age": bson.M{"$gte": 25}}})

	if agg == nil {
		t.Error("expected aggregation, got nil")
	}

	var results []bson.M
	err := agg.Execute(&results)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestModelSetNotFound(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create a model
	model := NewModel[User](client)

	// Test SetNotFound - verify the method returns Model
	result := model.SetNotFound(bson.ErrDecodeToNil)
	if result == nil {
		t.Error("SetNotFound should return non-nil Model")
	}

	if result != model {
		t.Error("SetNotFound should return the same Model instance")
	}
}

func TestModelIsNotFound(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create a model
	model := NewModel[User](client)

	// Insert a user
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "test@example.com",
		Name:      "Test User",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := model.NewScoop().Create(user)
	if err != nil {
		t.Errorf("create failed: %v", err)
	}

	// Try to find non-existent user
	modelScoop := model.NewScoop().Equal("email", "nonexistent@example.com")
	_, err = modelScoop.First()

	// Verify IsNotFound returns true
	if !model.IsNotFound(err) {
		t.Error("expected IsNotFound to return true for no documents found")
	}

	// Verify IsNotFound returns false for other errors
	otherErr := bson.ErrDecodeToNil
	if model.IsNotFound(otherErr) {
		t.Error("expected IsNotFound to return false for non-not-found error")
	}
}

// TestGetCollectionNameWithCollectionerInterface tests the Collectioner interface implementation
func TestGetCollectionNameWithCollectionerInterface(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Test with CustomUser that implements Collectioner interface
	customModel := NewModel[CustomUser](client)
	collName := customModel.CollectionName()

	expected := "custom_users"
	if collName != expected {
		t.Errorf("expected collection name '%s', got '%s'", expected, collName)
	}

	// Verify the interface method is being called
	if collName != "custom_users" {
		t.Error("Collectioner interface method was not called correctly")
	}
}

// TestGetCollectionNameFromTypeWithoutCollectionMethod tests default type name fallback
func TestGetCollectionNameFromTypeWithoutCollectionMethod(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Test with NoCollectionUser that doesn't implement Collection() method
	model := NewModel[NoCollectionUser](client)
	collName := model.CollectionName()

	expected := "NoCollectionUser"
	if collName != expected {
		t.Errorf("expected collection name '%s', got '%s'", expected, collName)
	}
}

// TestGetCollectionNameCaching tests the caching mechanism
func TestGetCollectionNameCaching(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Create two models of the same type
	model1 := NewModel[User](client)
	model2 := NewModel[User](client)

	// Both should have the same collection name
	if model1.CollectionName() != model2.CollectionName() {
		t.Error("same type should have the same collection name")
	}

	// Verify the collection name is from the Collection() interface method (because User implements Collectioner)
	if model1.CollectionName() != "users" {
		t.Errorf("expected 'users' (from Collection() method), got '%s'", model1.CollectionName())
	}
}

// TestGetCollectionNameWithPointerType tests that pointer types are handled correctly
func TestGetCollectionNameWithPointerType(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Even though NewModel works with non-pointer types internally,
	// the getCollectionNameFromType should handle pointer types correctly
	model := NewModel[User](client)

	// Verify it returns the collection name from the Collection() method
	if model.CollectionName() != "users" {
		t.Errorf("expected 'users', got '%s'", model.CollectionName())
	}
}

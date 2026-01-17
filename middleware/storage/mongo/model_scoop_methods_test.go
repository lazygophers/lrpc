package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestModelScoopLimit tests the Limit chain method
func TestModelScoopLimit(t *testing.T) {
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

	// Test Limit returns ModelScoop
	model := NewModel(client, User{})
	modelScoop := model.NewScoop().Limit(2)

	// Verify it returns ModelScoop by checking Find
	results, err := modelScoop.Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results with Limit(2), got %d", len(results))
	}
}

// TestModelScoopOffset tests the Offset chain method
func TestModelScoopOffset(t *testing.T) {
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

	// Test Offset returns ModelScoop
	model := NewModel(client, User{})
	modelScoop := model.NewScoop().Offset(1).Limit(2)

	results, err := modelScoop.Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results with Offset(1), got %d", len(results))
	}
}

// TestModelScoopSort tests the Sort chain method
func TestModelScoopSort(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "user1@example.com", Name: "User 1", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user2@example.com", Name: "User 2", Age: 20, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Sort returns ModelScoop and works
	model := NewModel(client, User{})
	results, err := model.NewScoop().Sort("age", 1).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Verify ascending order
	if results[0].Age != 20 || results[1].Age != 25 || results[2].Age != 30 {
		t.Errorf("expected ascending order [20, 25, 30], got [%d, %d, %d]", results[0].Age, results[1].Age, results[2].Age)
	}
}

// TestModelScoopSelect tests the Select chain method
func TestModelScoopSelect(t *testing.T) {
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

	// Test Select returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().Select("email", "name").Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected results, got empty")
	}
}

// TestModelScoopEqual tests the Equal chain method
func TestModelScoopEqual(t *testing.T) {
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

	// Test Equal returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().Equal("age", 25).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for Equal, got %d", len(results))
	}
}

// TestModelScoopNe tests the Ne chain method
func TestModelScoopNe(t *testing.T) {
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

	// Test Ne returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().Ne("age", 25).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for Ne, got %d", len(results))
	}
}

// TestModelScoopIn tests the In chain method
func TestModelScoopIn(t *testing.T) {
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

	// Test In returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().In("age", 25, 30).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for In, got %d", len(results))
	}
}

// TestModelScoopNotIn tests the NotIn chain method
func TestModelScoopNotIn(t *testing.T) {
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

	// Test NotIn returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().NotIn("age", 25, 30).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for NotIn, got %d", len(results))
	}
}

// TestModelScoopLike tests the Like chain method
func TestModelScoopLike(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data
	users := []interface{}{
		User{ID: primitive.NewObjectID(), Email: "test@example.com", Name: "Test User", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		User{ID: primitive.NewObjectID(), Email: "demo@example.com", Name: "Demo User", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Like returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().Like("name", "Test").Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for Like, got %d", len(results))
	}
}

// TestModelScoopGt tests the Gt chain method
func TestModelScoopGt(t *testing.T) {
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

	// Test Gt returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().Gt("age", 25).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for Gt, got %d", len(results))
	}
}

// TestModelScoopLt tests the Lt chain method
func TestModelScoopLt(t *testing.T) {
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

	// Test Lt returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().Lt("age", 30).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for Lt, got %d", len(results))
	}
}

// TestModelScoopGte tests the Gte chain method
func TestModelScoopGte(t *testing.T) {
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

	// Test Gte returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().Gte("age", 25).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for Gte, got %d", len(results))
	}
}

// TestModelScoopLte tests the Lte chain method
func TestModelScoopLte(t *testing.T) {
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

	// Test Lte returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().Lte("age", 25).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result for Lte, got %d", len(results))
	}
}

// TestModelScoopBetween tests the Between chain method
func TestModelScoopBetween(t *testing.T) {
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

	// Test Between returns ModelScoop
	model := NewModel(client, User{})
	results, err := model.NewScoop().Between("age", 25, 30).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results for Between, got %d", len(results))
	}
}

// TestModelScoopSkip tests the Skip chain method
func TestModelScoopSkip(t *testing.T) {
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
		User{ID: primitive.NewObjectID(), Email: "user3@example.com", Name: "User 3", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	InsertTestData(t, client, "users", users...)

	// Test Skip returns ModelScoop and is equivalent to Offset
	model := NewModel(client, User{})
	results, err := model.NewScoop().Skip(1).Limit(2).Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results with Skip(1), got %d", len(results))
	}
}

// TestModelScoopClear tests the Clear chain method
func TestModelScoopClear(t *testing.T) {
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

	// Create a scoop with filter
	model := NewModel(client, User{})
	modelScoop := model.NewScoop().Equal("age", 25)

	// Verify filter works
	results, err := modelScoop.Find()
	if err != nil {
		t.Errorf("find failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result before clear, got %d", len(results))
	}

	// Clear and verify no filters
	modelScoop = modelScoop.Clear()
	results, err = modelScoop.Find()
	if err != nil {
		t.Errorf("find after clear failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results after clear, got %d", len(results))
	}
}

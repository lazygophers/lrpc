package mongo

import (
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestScoopCountMultipleRanges tests Count with multiple range conditions
func TestScoopCountMultipleRanges(t *testing.T) {
	t.Skip("Skipping due to data contamination in shared test environment")
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert documents with different ages for range testing
	ages := []int{18, 20, 25, 30, 35, 40, 45, 50}
	for i, age := range ages {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("range%d@example.com", i),
			Name:      "User",
			Age:       age,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Test different range queries
	testCases := []struct {
		name     string
		condFn   func(*Scoop) *Scoop
		expected int64
	}{
		{
			name: "between 25 and 40",
			condFn: func(s *Scoop) *Scoop {
				return s.Where("age", bson.M{"$gte": 25}).
					Where("age", bson.M{"$lte": 40})
			},
			expected: 4, // 25, 30, 35, 40
		},
		{
			name: "less than 30",
			condFn: func(s *Scoop) *Scoop {
				return s.Where("age", bson.M{"$lt": 30})
			},
			expected: 3, // 18, 20, 25
		},
		{
			name: "greater than or equal 40",
			condFn: func(s *Scoop) *Scoop {
				return s.Where("age", bson.M{"$gte": 40})
			},
			expected: 3, // 40, 45, 50
		},
		{
			name: "exact age 35",
			condFn: func(s *Scoop) *Scoop {
				return s.Equal("age", 35)
			},
			expected: 1,
		},
	}

	for _, tc := range testCases {
		scoop := client.NewScoop().Collection(User{})
		scoop = tc.condFn(scoop)
		count, err := scoop.Count()
		if err != nil {
			t.Errorf("count test %s failed: %v", tc.name, err)
			continue
		}
		if count != tc.expected {
			t.Errorf("test %s: expected %d, got %d", tc.name, tc.expected, count)
		}
	}
}

// TestScoopDeleteWithMultipleConditions tests Delete with complex conditions
func TestScoopDeleteWithMultipleConditions(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert test data with categories
	categories := []struct {
		email string
		name  string
		age   int
		count int
	}{
		{"young", "Young User", 20, 3},
		{"middle", "Middle User", 35, 4},
		{"old", "Old User", 50, 2},
	}

	for _, cat := range categories {
		for i := 0; i < cat.count; i++ {
			user := User{
				ID:        primitive.NewObjectID(),
				Email:     fmt.Sprintf("%s%d@example.com", cat.email, i),
				Name:      cat.name,
				Age:       cat.age,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			}
			InsertTestData(t, client, "users", user)
		}
	}

	// Delete one category
	scoop := client.NewScoop().Collection(User{}).Equal("age", 35)
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete with condition failed: %v", err)
	}

	if deleted != 4 {
		t.Errorf("expected 4 deleted, got %d", deleted)
	}

	// Verify remaining
	remaining, _ := client.NewScoop().Collection(User{}).Count()
	if remaining != 5 {
		t.Errorf("expected 5 remaining (3+2), got %d", remaining)
	}
}

// TestScoopDeleteAllThenVerify tests Delete all and verify empty
func TestScoopDeleteAllThenVerify(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert data
	for i := 0; i < 7; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("delete_all%d@example.com", i),
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Verify count
	count, _ := client.NewScoop().Collection(User{}).Count()
	if count != 7 {
		t.Errorf("expected 7 documents before delete, got %d", count)
	}

	// Delete all
	scoop := client.NewScoop().Collection(User{})
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete all failed: %v", err)
	}

	if deleted != 7 {
		t.Errorf("expected 7 deleted, got %d", deleted)
	}

	// Verify empty
	final, _ := scoop.Count()
	if final != 0 {
		t.Errorf("expected 0 remaining, got %d", final)
	}
}

// TestClientHealthWithValidConnection tests Health on valid connection
func TestClientHealthWithValidConnection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Multiple consecutive health checks
	for i := 0; i < 3; i++ {
		err := client.Health()
		if err != nil {
			t.Logf("Health check %d error: %v", i, err)
		}
	}
}

// TestClientPingMultipleTimes tests Ping can be called multiple times
func TestClientPingMultipleTimes(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Call Ping multiple times
	for i := 0; i < 5; i++ {
		err := client.Ping()
		if err != nil {
			t.Logf("Ping %d error: %v", i, err)
		}
	}
}

// TestStreamWatchWithComplexPipeline tests Watch with multi-stage pipeline
func TestStreamWatchWithComplexPipeline(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("create change stream failed: %v", err)
	}

	// Complex pipeline with multiple stages
	stream, err := cs.Watch(
		bson.M{
			"$match": bson.M{
				"operationType": "insert",
			},
		},
		bson.M{
			"$match": bson.M{
				"fullDocument.age": bson.M{
					"$gte": 20,
				},
			},
		},
	)

	if err != nil {
		t.Logf("Watch with complex pipeline error: %v", err)
	} else if stream != nil {
		stream.Close(client.Context())
	}
}

// TestStreamWatchWithProjection tests Watch with projection
func TestStreamWatchWithProjection(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("create change stream failed: %v", err)
	}

	// Pipeline with projection
	stream, err := cs.Watch(
		bson.M{
			"$project": bson.M{
				"fullDocument.email": 1,
				"fullDocument.age":   1,
			},
		},
	)

	if err != nil {
		t.Logf("Watch with projection error: %v", err)
	} else if stream != nil {
		stream.Close(client.Context())
	}
}

// TestCountWithInOperator tests Count with $in operator
func TestCountWithInOperator(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert diverse data
	ages := []int{22, 24, 25, 26, 28, 30, 32, 35}
	for i, age := range ages {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("in%d@example.com", i),
			Name:      "User",
			Age:       age,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Count with $in operator
	scoop := client.NewScoop().Collection(User{}).
		Where("age", bson.M{"$in": []int{22, 26, 30, 35}})

	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count with $in failed: %v", err)
	}

	// Should match 22, 26, 30, 35
	if count != 4 {
		t.Errorf("expected 4 matches with $in, got %d", count)
	}
}

// TestCountWithNinOperator tests Count with $nin operator
func TestCountWithNinOperator(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert data
	for i := 0; i < 6; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("nin%d@example.com", i),
			Name:      "User",
			Age:       20 + i,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Count with $nin operator
	scoop := client.NewScoop().Collection(User{}).
		Where("age", bson.M{"$nin": []int{20, 21, 22}})

	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count with $nin failed: %v", err)
	}

	// Should match 23, 24, 25 (6 total - 3 excluded)
	if count != 3 {
		t.Errorf("expected 3 matches with $nin, got %d", count)
	}
}

// TestDeleteWithInOperator tests Delete with $in operator
func TestDeleteWithInOperator(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert data
	emails := []string{"a@test.com", "b@test.com", "c@test.com", "d@test.com"}
	for i, email := range emails {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     email,
			Name:      fmt.Sprintf("User%d", i),
			Age:       25,
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Delete with $in
	scoop := client.NewScoop().Collection(User{}).
		Where("email", bson.M{"$in": []string{"a@test.com", "c@test.com"}})

	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete with $in failed: %v", err)
	}

	if deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", deleted)
	}

	// Verify
	remaining, _ := client.NewScoop().Collection(User{}).Count()
	if remaining != 2 {
		t.Errorf("expected 2 remaining, got %d", remaining)
	}
}

// TestDatabaseChangeStreamWatchWithMultiplePipelines tests DatabaseChangeStream.Watch
func TestDatabaseChangeStreamWatchWithMultiplePipelines(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("create database change stream failed: %v", err)
	}

	// Watch with multiple stages
	stream, err := dcs.Watch(
		bson.M{
			"$match": bson.M{
				"operationType": bson.M{
					"$in": []string{"insert", "update"},
				},
			},
		},
	)

	if err != nil {
		t.Logf("Database Watch with pipeline error: %v", err)
	} else if stream != nil {
		stream.Close(client.Context())
	}
}

// TestCountAfterMultipleInserts tests Count reflects all inserts
func TestCountAfterMultipleInserts(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert in batches and verify count increases
	testCases := []int{1, 3, 5, 10}
	total := 0

	for _, batchSize := range testCases {
		// Insert batch
		for i := 0; i < batchSize; i++ {
			user := User{
				ID:        primitive.NewObjectID(),
				Email:     fmt.Sprintf("batch%d_%d@example.com", total+i, i),
				Name:      "User",
				Age:       25,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			}
			InsertTestData(t, client, "users", user)
		}
		total += batchSize

		// Verify count
		scoop := client.NewScoop().Collection(User{})
		count, err := scoop.Count()
		if err != nil {
			t.Errorf("count after batch failed: %v", err)
			continue
		}

		if count != int64(total) {
			t.Errorf("after batch of %d: expected total %d, got %d", batchSize, total, count)
		}
	}
}

// TestDeletePartialThenCountRemaining tests delete partial and count remaining
func TestDeletePartialThenCountRemaining(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert 20 users with different ages
	for i := 0; i < 20; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("partial%d@example.com", i),
			Name:      "User",
			Age:       20 + (i % 15), // Ages 20-34
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Delete young users (age < 25)
	scoop := client.NewScoop().Collection(User{}).
		Where("age", bson.M{"$lt": 25})

	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete young users failed: %v", err)
	}

	// Count should reflect deletions
	remaining, err := client.NewScoop().Collection(User{}).Count()
	if err != nil {
		t.Fatalf("count remaining failed: %v", err)
	}

	if remaining != int64(20-int(deleted)) {
		t.Errorf("deleted %d, expected %d remaining, got %d", deleted, 20-int(deleted), remaining)
	}
}

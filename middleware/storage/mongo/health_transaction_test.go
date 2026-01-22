package mongo

import (
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestClientHealthCheckDirect tests Health method directly
func TestClientHealthCheckDirect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	err := client.Health()
	if err != nil {
		t.Logf("Health check error: %v", err)
	}
}

// TestClientHealthConsecutiveCalls tests multiple Health calls
func TestClientHealthConsecutiveCalls(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	for i := 0; i < 3; i++ {
		err := client.Health()
		if err != nil {
			t.Logf("Health call %d error: %v", i, err)
		}
	}
}

// TestCountEmptyCollectionPathway tests Count branch on empty
func TestCountEmptyCollectionPathway(t *testing.T) {
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
		t.Fatalf("count empty failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

// TestCountSingleItemPathway tests Count with one document
func TestCountSingleItemPathway(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "count_pathway_single@example.com",
		Name:      "User",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count single failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

// TestCountManyItemsPathway tests Count with multiple documents
func TestCountManyItemsPathway(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("count_pathway_many_%d@example.com", i),
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	count, err := scoop.Count()
	if err != nil {
		t.Fatalf("count many failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5, got %d", count)
	}
}

// TestDeleteEmptyCollectionPathway tests Delete branch on empty
func TestDeleteEmptyCollectionPathway(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete empty failed: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0, got %d", deleted)
	}
}

// TestDeleteSingleItemPathway tests Delete with one document
func TestDeleteSingleItemPathway(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "delete_pathway_single@example.com",
		Name:      "User",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	scoop := client.NewScoop().Collection(User{})
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete single failed: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1, got %d", deleted)
	}
}

// TestDeleteManyItemsPathway tests Delete with multiple documents
func TestDeleteManyItemsPathway(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("delete_pathway_many_%d@example.com", i),
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	deleteResult := scoop.Delete()
	deleted, err := deleteResult.DocsAffected, deleteResult.Error
	if err != nil {
		t.Fatalf("delete many failed: %v", err)
	}
	if deleted != 5 {
		t.Errorf("expected 5, got %d", deleted)
	}
}

// TestDatabaseChangeStreamWatchNoArgs tests Watch with no arguments
func TestDatabaseChangeStreamWatchNoArgs(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("create database change stream failed: %v", err)
	}

	stream, err := dcs.Watch()
	if err != nil {
		t.Logf("Watch no args error: %v", err)
	} else if stream != nil {
		stream.Close(client.Context())
	}
}

// TestDatabaseChangeStreamWatchSingleStage tests Watch with one pipeline stage
func TestDatabaseChangeStreamWatchSingleStage(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("create database change stream failed: %v", err)
	}

	stream, err := dcs.Watch(
		bson.M{
			"$match": bson.M{
				"operationType": "insert",
			},
		},
	)
	if err != nil {
		t.Logf("Watch single stage error: %v", err)
	} else if stream != nil {
		stream.Close(client.Context())
	}
}

// TestDatabaseChangeStreamWatchTwoStages tests Watch with two pipeline stages
func TestDatabaseChangeStreamWatchTwoStages(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("create database change stream failed: %v", err)
	}

	stream, err := dcs.Watch(
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
		t.Logf("Watch two stages error: %v", err)
	} else if stream != nil {
		stream.Close(client.Context())
	}
}

// TestTransactionCommitPathway tests Begin and Commit
func TestTransactionCommitPathway(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	_, err := scoop.Begin()
	if err != nil {
		t.Logf("Begin error: %v", err)
		return
	}

	err = scoop.Commit()
	if err != nil {
		t.Logf("Commit error: %v", err)
	}
}

// TestTransactionRollbackPathway tests Begin and Rollback
func TestTransactionRollbackPathway(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	_, err := scoop.Begin()
	if err != nil {
		t.Logf("Begin error: %v", err)
		return
	}

	err = scoop.Rollback()
	if err != nil {
		t.Logf("Rollback error: %v", err)
	}
}

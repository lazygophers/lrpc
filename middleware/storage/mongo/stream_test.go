package mongo

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestChangeStreamWatchChanges(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create change stream
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()

	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("watch changes failed: %v", err)
	}

	if cs == nil {
		t.Error("expected change stream, got nil")
	}
}

func TestChangeStreamWatch(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create change stream
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()

	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("watch changes failed: %v", err)
	}

	// Test Watch method - should return mongo.ChangeStream
	stream, err := cs.Watch()
	if err != nil {
		t.Fatalf("watch failed: %v", err)
	}

	if stream == nil {
		t.Error("expected stream, got nil")
	}

	// Close the stream
	stream.Close(context.Background())
}

func TestChangeStreamClose(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()

	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("watch changes failed: %v", err)
	}

	// Close should not panic or error (it's a no-op)
	cs.Close()
}

func TestModelScoopWatch(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Test ModelScoop Watch method
	model := NewModel(client, User{})
	modelScoop := model.NewScoop()

	cs, err := modelScoop.Watch()
	if err != nil {
		t.Fatalf("watch failed: %v", err)
	}

	if cs == nil {
		t.Error("expected change stream, got nil")
	}
}

func TestChangeStreamWatchWithPipeline(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Create change stream
	model := NewModel(client, User{})
	scoop := model.NewScoop().GetScoop()

	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("watch changes failed: %v", err)
	}

	// Test Watch with pipeline (match operation types)
	stream, err := cs.Watch(bson.M{
		"$match": bson.M{
			"operationType": "insert",
		},
	})
	if err != nil {
		t.Fatalf("watch with pipeline failed: %v", err)
	}

	if stream == nil {
		t.Error("expected stream, got nil")
	}

	stream.Close(context.Background())
}

func TestDatabaseChangeStreamWatchAllCollections(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Watch all collections in database
	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("watch all collections failed: %v", err)
	}

	if dcs == nil {
		t.Error("expected database change stream, got nil")
	}
}

func TestDatabaseChangeStreamWatch(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Watch all collections
	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("watch all collections failed: %v", err)
	}

	// Test Watch method
	stream, err := dcs.Watch()
	if err != nil {
		t.Fatalf("database watch failed: %v", err)
	}

	if stream == nil {
		t.Error("expected stream, got nil")
	}

	stream.Close(context.Background())
}

func TestDatabaseChangeStreamClose(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("watch all collections failed: %v", err)
	}

	// Close should not panic or error (it's a no-op)
	dcs.Close()
}

func TestPrintEvent(t *testing.T) {
	event := &ChangeEvent{
		OperationType: "insert",
		FullDocument: map[string]interface{}{
			"_id":   primitive.NewObjectID(),
			"name":  "test",
		},
	}

	// PrintEvent should not panic
	err := PrintEvent(event)
	if err != nil {
		t.Errorf("print event failed: %v", err)
	}
}

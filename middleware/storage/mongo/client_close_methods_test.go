package mongo

import (
	"testing"
)

// TestChangeStreamCloseMethod tests ChangeStream.Close()
func TestChangeStreamCloseMethod(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})
	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("create change stream failed: %v", err)
	}

	// Call Close - should not panic
	cs.Close()
}

// TestDatabaseChangeStreamCloseMethod tests DatabaseChangeStream.Close()
func TestDatabaseChangeStreamCloseMethod(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("create database change stream failed: %v", err)
	}

	// Call Close - should not panic
	dcs.Close()
}

// TestStreamCloseMultipleCalls tests calling Close multiple times
func TestStreamCloseMultipleCalls(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	scoop := client.NewScoop().Collection(User{})
	cs, err := scoop.WatchChanges()
	if err != nil {
		t.Fatalf("create change stream failed: %v", err)
	}

	// Call Close multiple times - should not panic
	cs.Close()
	cs.Close()
	cs.Close()
}

// TestDatabaseChangeStreamCloseMultipleCalls tests calling DatabaseChangeStream.Close multiple times
func TestDatabaseChangeStreamCloseMultipleCalls(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("create database change stream failed: %v", err)
	}

	// Call Close multiple times - should not panic
	dcs.Close()
	dcs.Close()
	dcs.Close()
}

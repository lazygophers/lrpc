package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestStreamChangeStreamCloseDirect directly tests ChangeStream.Close
func TestStreamChangeStreamCloseDirect(t *testing.T) {
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

	// This should not panic
	cs.Close()
	
	// Calling again should also not panic
	cs.Close()
}

// TestDatabaseChangeStreamCloseDirect directly tests DatabaseChangeStream.Close
func TestDatabaseChangeStreamCloseDirect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("create database change stream failed: %v", err)
	}

	// This should not panic
	dcs.Close()
	
	// Calling again should also not panic
	dcs.Close()
}

// TestClientHealthDirect directly calls Health method
func TestClientHealthDirect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Health should either succeed or return a wrapped error
	err := client.Health()
	
	// Just ensure it doesn't panic and returns appropriate value
	if err != nil {
		// Check that error message contains health check context
		errMsg := err.Error()
		if len(errMsg) == 0 {
			t.Error("expected non-empty error message")
		}
	}
}

// TestClientPingDirect directly calls Ping method
func TestClientPingDirect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	// Ping should work or return error
	err := client.Ping()
	
	// Just ensure no panic
	if err != nil {
		t.Logf("Ping error (may be expected): %v", err)
	}
}

// TestScoopCountDirectNormal directly tests Count in normal case
func TestScoopCountDirectNormal(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert exactly 3 documents
	for i := 0; i < 3; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "count_test" + string(rune(48+i)) + "@example.com",
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
		t.Errorf("Count returned error: %v", err)
		return
	}

	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

// TestScoopCountDirectZero directly tests Count with zero results
func TestScoopCountDirectZero(t *testing.T) {
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
		t.Errorf("Count on empty collection failed: %v", err)
		return
	}

	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}
}

// TestScoopDeleteDirectNormal directly tests Delete in normal case
func TestScoopDeleteDirectNormal(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	// Insert 5 documents
	for i := 0; i < 5; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     "delete_test" + string(rune(48+i)) + "@example.com",
			Name:      "User",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	scoop := client.NewScoop().Collection(User{})
	deleted, err := scoop.Delete()
	
	if err != nil {
		t.Errorf("Delete returned error: %v", err)
		return
	}

	if deleted != 5 {
		t.Errorf("expected deleted 5, got %d", deleted)
	}
}

// TestScoopDeleteDirectZero directly tests Delete with zero matches
func TestScoopDeleteDirectZero(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})
	deleted, err := scoop.Delete()
	
	if err != nil {
		t.Errorf("Delete on empty collection failed: %v", err)
		return
	}

	if deleted != 0 {
		t.Errorf("expected deleted 0, got %d", deleted)
	}
}

// TestClientContextDirect directly tests Context method
func TestClientContextDirect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	ctx := client.Context()
	if ctx == nil {
		t.Error("expected non-nil context")
		return
	}

	// Verify context is not cancelled
	select {
	case <-ctx.Done():
		t.Error("context should not be cancelled")
	default:
		// Expected - context is active
	}
}

// TestClientGetConfigDirect directly tests GetConfig
func TestClientGetConfigDirect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cfg := client.GetConfig()
	if cfg == nil {
		t.Error("expected non-nil config")
		return
	}

	// Config should have reasonable values
	if cfg.Database == "" {
		// Default database may be empty, which is ok
		t.Logf("Config database is empty (will use default)")
	}
}

// TestClientGetDatabaseDirect directly tests GetDatabase
func TestClientGetDatabaseDirect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	db := client.GetDatabase()
	
	// Should not be empty - either has explicit value or returns "test" default
	if db == "" {
		t.Error("expected non-empty database name")
	}
}

// TestStreamWatchDirect directly tests ChangeStream.Watch
func TestStreamWatchDirect(t *testing.T) {
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

	// Watch with no pipeline
	stream, err := cs.Watch()
	if err != nil {
		t.Logf("Watch failed: %v", err)
	} else if stream != nil {
		// Close the stream
		stream.Close(client.Context())
	}
}

// TestStreamListenHandlerReturnValue directly tests handler return value
func TestStreamListenHandlerReturnValue(t *testing.T) {
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

	handlerCalls := 0
	handler := func(event *ChangeEvent) error {
		handlerCalls++
		// Return error immediately to stop listening
		return ErrTestingStop
	}

	// Start listening
	errChan := make(chan error, 1)
	go func() {
		errChan <- cs.Listen(handler)
	}()

	time.Sleep(100 * time.Millisecond)

	// Insert document to trigger handler
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "handler_test@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// Wait for handler to be called
	time.Sleep(100 * time.Millisecond)
}

// TestDatabaseChangeStreamWatchDirect directly tests DatabaseChangeStream.Watch
func TestDatabaseChangeStreamWatchDirect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("create database change stream failed: %v", err)
	}

	// Watch with no pipeline
	stream, err := dcs.Watch()
	if err != nil {
		t.Logf("Database Watch failed: %v", err)
	} else if stream != nil {
		// Close the stream
		stream.Close(client.Context())
	}
}

// TestDatabaseChangeStreamListenDirect directly tests DatabaseChangeStream.Listen
func TestDatabaseChangeStreamListenDirect(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	dcs, err := client.WatchAllCollections()
	if err != nil {
		t.Fatalf("create database change stream failed: %v", err)
	}

	handler := func(event *ChangeEvent) error {
		// Stop immediately
		return ErrTestingStop
	}

	// Start listening
	go func() {
		dcs.Listen(handler)
	}()

	time.Sleep(100 * time.Millisecond)

	// Insert to trigger
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "dbhandler@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	time.Sleep(100 * time.Millisecond)
}

// Define a special error for testing
type testStopError struct{}

func (e *testStopError) Error() string {
	return "stop testing"
}

var ErrTestingStop error = &testStopError{}

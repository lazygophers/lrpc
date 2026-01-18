package mongo

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestChangeStreamListenBasic tests ChangeStream Listen method with actual events
func TestChangeStreamListenBasic(t *testing.T) {
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

	// Handler function to receive events
	handler := func(event *ChangeEvent) error {
		// Receive event but don't do anything with it
		return nil
	}

	// Start listening in a goroutine
	errChan := make(chan error, 1)
	go func() {
		// This will block until events are received or timeout
		errChan <- cs.Listen(handler, bson.M{
			"$match": bson.M{
				"operationType": "insert",
			},
		})
	}()

	// Give listener time to start
	time.Sleep(100 * time.Millisecond)

	// Insert a document to trigger change event
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "listen@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// Wait briefly for event processing
	time.Sleep(100 * time.Millisecond)

	// Note: Listen will block, so this is mainly to ensure no panic
	// In a real scenario, you'd implement stream cancellation
}

// TestChangeStreamListenWithFiltersBasic tests ListenWithFilters
func TestChangeStreamListenWithFiltersBasic(t *testing.T) {
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

	eventCount := 0
	handler := func(event *ChangeEvent) error {
		eventCount++
		return nil
	}

	// Start listening in a goroutine with operation type filter
	errChan := make(chan error, 1)
	go func() {
		errChan <- cs.ListenWithFilters(handler, "insert")
	}()

	// Give listener time to start
	time.Sleep(100 * time.Millisecond)

	// Insert documents to trigger change events
	for i := 0; i < 2; i++ {
		user := User{
			ID:        primitive.NewObjectID(),
			Email:     fmt.Sprintf("filter%d@example.com", i),
			Name:      "Test",
			Age:       25,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		InsertTestData(t, client, "users", user)
	}

	// Wait briefly
	time.Sleep(100 * time.Millisecond)
}

// TestChangeStreamListenWithMultipleOperationTypes tests ListenWithFilters with multiple operation types
func TestChangeStreamListenWithMultipleOperationTypes(t *testing.T) {
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

	eventCount := 0
	handler := func(event *ChangeEvent) error {
		eventCount++
		return nil
	}

	// Start listening with multiple operation type filters
	errChan := make(chan error, 1)
	go func() {
		errChan <- cs.ListenWithFilters(handler, "insert", "update")
	}()

	// Give listener time to start
	time.Sleep(100 * time.Millisecond)

	// Insert a document
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "multi@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// Update the document
	updateScoop := client.NewScoop().Collection(User{}).Equal("email", "multi@example.com")
	updateScoop.Update(bson.M{"$set": bson.M{"age": 30}})

	// Wait briefly
	time.Sleep(200 * time.Millisecond)
}

// TestChangeStreamListenWithErrorHandler tests Listen with handler that returns error
func TestChangeStreamListenWithErrorHandler(t *testing.T) {
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

	callCount := 0
	handler := func(event *ChangeEvent) error {
		callCount++
		if callCount >= 1 {
			// Return error to stop listening
			return fmt.Errorf("stop listening")
		}
		return nil
	}

	// Start listening
	errChan := make(chan error, 1)
	go func() {
		errChan <- cs.Listen(handler, bson.M{
			"$match": bson.M{
				"operationType": "insert",
			},
		})
	}()

	// Give listener time to start
	time.Sleep(100 * time.Millisecond)

	// Insert a document to trigger error handler
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "error@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// Wait for error
	time.Sleep(100 * time.Millisecond)
}

// TestDatabaseChangeStreamListenBasic tests DatabaseChangeStream Listen
func TestDatabaseChangeStreamListenBasic(t *testing.T) {
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
		// Receive event but don't do anything with it
		return nil
	}

	// Start listening in a goroutine
	go func() {
		dcs.Listen(handler, bson.M{
			"$match": bson.M{
				"operationType": "insert",
			},
		})
	}()

	// Give listener time to start
	time.Sleep(100 * time.Millisecond)

	// Insert a document
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "dbwatch@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// Wait briefly
	time.Sleep(100 * time.Millisecond)
}

// TestStreamListenConcurrentCalls tests multiple concurrent listen calls
func TestStreamListenConcurrentCalls(t *testing.T) {
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

	var wg sync.WaitGroup
	eventCounts := make([]int, 3)
	var mu sync.Mutex

	// Start multiple listeners
	for i := 0; i < 3; i++ {
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()

			handler := func(event *ChangeEvent) error {
				mu.Lock()
				eventCounts[idx]++
				mu.Unlock()
				// Stop after first event
				return fmt.Errorf("stop")
			}

			cs.Listen(handler)
		}()
	}

	// Give listeners time to start
	time.Sleep(100 * time.Millisecond)

	// Insert a document
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "concurrent@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// Wait briefly for events
	time.Sleep(200 * time.Millisecond)

	// Wait for all goroutines (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected
	case <-time.After(1 * time.Second):
		// Timeout is ok for this test
	}
}

// TestChangeStreamWatchEmptyPipeline tests Watch with empty pipeline
func TestChangeStreamWatchEmptyPipeline(t *testing.T) {
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

	// Watch with empty pipeline
	stream, err := cs.Watch()
	if err != nil {
		t.Logf("watch with empty pipeline: %v", err)
	}
	if stream != nil {
		stream.Close(client.Context())
	}
}

// TestChangeStreamListenWithEmptyOperationTypes tests ListenWithFilters with no operation types
func TestChangeStreamListenWithEmptyOperationTypes(t *testing.T) {
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

	eventCount := 0
	handler := func(event *ChangeEvent) error {
		eventCount++
		return fmt.Errorf("stop")
	}

	// Start listening without specifying operation types
	errChan := make(chan error, 1)
	go func() {
		errChan <- cs.ListenWithFilters(handler)
	}()

	// Give listener time to start
	time.Sleep(100 * time.Millisecond)

	// Insert a document (should be captured even without filter)
	user := User{
		ID:        primitive.NewObjectID(),
		Email:     "nofilter@example.com",
		Name:      "Test",
		Age:       25,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	InsertTestData(t, client, "users", user)

	// Wait briefly
	time.Sleep(100 * time.Millisecond)
}

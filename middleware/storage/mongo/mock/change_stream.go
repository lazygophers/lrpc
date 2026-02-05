package mock

import (
	"context"
	"sync"
	"time"

	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChangeEventType represents the type of change event
type ChangeEventType string

const (
	// ChangeEventInsert represents an insert operation
	ChangeEventInsert ChangeEventType = "insert"
	// ChangeEventUpdate represents an update operation
	ChangeEventUpdate ChangeEventType = "update"
	// ChangeEventReplace represents a replace operation
	ChangeEventReplace ChangeEventType = "replace"
	// ChangeEventDelete represents a delete operation
	ChangeEventDelete ChangeEventType = "delete"
)

// ChangeEvent represents a change stream event
type ChangeEvent struct {
	OperationType     ChangeEventType
	CollectionName    string
	DocumentKey       bson.M
	FullDocument      bson.M
	UpdateDescription *UpdateDescription
}

// UpdateDescription describes the fields updated in an update operation
type UpdateDescription struct {
	UpdatedFields bson.M
	RemovedFields []string
}

// MockChangeStream implements a mock change stream for testing
// It receives events from MemoryStorage and provides them through Next/Decode
type MockChangeStream struct {
	events     chan ChangeEvent
	current    *ChangeEvent
	closed     bool
	mu         sync.RWMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
	filter     ChangeStreamFilter
}

// ChangeStreamFilter defines filtering criteria for change events
type ChangeStreamFilter struct {
	// CollectionName filters events by collection name
	// Empty string means no filter (all collections)
	CollectionName string
	// OperationTypes filters events by operation type
	// Empty slice means no filter (all operations)
	OperationTypes []ChangeEventType
}

// NewMockChangeStream creates a new MockChangeStream instance
// Parameters:
//   - ctx: context for stream lifecycle
//   - filter: filtering criteria for events
//   - bufferSize: size of the event buffer channel
func NewMockChangeStream(ctx context.Context, filter ChangeStreamFilter, bufferSize int) *MockChangeStream {
	if bufferSize <= 0 {
		bufferSize = 100 // default buffer size
	}

	streamCtx, cancel := context.WithCancel(ctx)

	stream := &MockChangeStream{
		events:     make(chan ChangeEvent, bufferSize),
		closed:     false,
		ctx:        streamCtx,
		cancelFunc: cancel,
		filter:     filter,
	}

	log.Debugf("created MockChangeStream with filter: %+v, buffer size: %d", filter, bufferSize)
	return stream
}

// Next advances the change stream to the next event
// Returns true if an event is available, false if stream is closed or context is cancelled
func (m *MockChangeStream) Next(ctx context.Context) bool {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		log.Debugf("MockChangeStream.Next() - stream is closed")
		return false
	}
	m.mu.RUnlock()

	select {
	case <-ctx.Done():
		log.Debugf("MockChangeStream.Next() - context cancelled")
		return false

	case <-m.ctx.Done():
		log.Debugf("MockChangeStream.Next() - stream context cancelled")
		return false

	case event, ok := <-m.events:
		if !ok {
			log.Debugf("MockChangeStream.Next() - events channel closed")
			return false
		}

		m.mu.Lock()
		m.current = &event
		m.mu.Unlock()

		log.Debugf("MockChangeStream.Next() - received event: %s on %s", event.OperationType, event.CollectionName)
		return true
	}
}

// Decode decodes the current event into val
// val should be a pointer to a struct or bson.M
func (m *MockChangeStream) Decode(val interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.current == nil {
		log.Debugf("MockChangeStream.Decode() - no current event")
		return nil
	}

	// Create a MongoDB change stream document format
	changeDoc := m.createChangeDocument(m.current)

	// Marshal and unmarshal to convert to target type
	data, err := bson.Marshal(changeDoc)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = bson.Unmarshal(data, val)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	log.Debugf("MockChangeStream.Decode() - decoded event successfully")
	return nil
}

// createChangeDocument creates a MongoDB change stream document format
func (m *MockChangeStream) createChangeDocument(event *ChangeEvent) bson.M {
	doc := bson.M{
		"_id": bson.M{
			"_data": primitive.NewObjectID().Hex(),
		},
		"operationType": string(event.OperationType),
		"ns": bson.M{
			"db":   "mock_database",
			"coll": event.CollectionName,
		},
		"clusterTime": primitive.Timestamp{T: uint32(time.Now().Unix()), I: 1},
	}

	// Add documentKey if available
	if event.DocumentKey != nil {
		doc["documentKey"] = event.DocumentKey
	}

	// Add fullDocument for insert and replace operations
	if event.OperationType == ChangeEventInsert || event.OperationType == ChangeEventReplace {
		if event.FullDocument != nil {
			doc["fullDocument"] = event.FullDocument
		}
	}

	// Add updateDescription for update operations
	if event.OperationType == ChangeEventUpdate && event.UpdateDescription != nil {
		doc["updateDescription"] = bson.M{
			"updatedFields": event.UpdateDescription.UpdatedFields,
			"removedFields": event.UpdateDescription.RemovedFields,
		}
	}

	return doc
}

// Close closes the change stream
func (m *MockChangeStream) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		log.Debugf("MockChangeStream.Close() - already closed")
		return nil
	}

	m.closed = true
	m.cancelFunc()
	close(m.events)

	log.Debugf("MockChangeStream.Close() - stream closed successfully")
	return nil
}

// Err returns the last error that occurred
// In mock implementation, always returns nil
func (m *MockChangeStream) Err() error {
	return nil
}

// ID returns the cursor ID for this change stream
// In mock implementation, returns 0
func (m *MockChangeStream) ID() int64 {
	return 0
}

// TryNext attempts to get the next event without blocking
// Returns true if an event is available, false otherwise
func (m *MockChangeStream) TryNext(ctx context.Context) bool {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return false
	}
	m.mu.RUnlock()

	select {
	case event, ok := <-m.events:
		if !ok {
			return false
		}

		m.mu.Lock()
		m.current = &event
		m.mu.Unlock()

		log.Debugf("MockChangeStream.TryNext() - received event: %s on %s", event.OperationType, event.CollectionName)
		return true

	default:
		return false
	}
}

// ResumeToken returns the resume token for the most recently returned document
// In mock implementation, returns an empty document
func (m *MockChangeStream) ResumeToken() bson.Raw {
	return bson.Raw{}
}

// SetBatchSize sets the batch size for the change stream
// In mock implementation, this is a no-op
func (m *MockChangeStream) SetBatchSize(batchSize int32) {
	log.Debugf("MockChangeStream.SetBatchSize() - no-op in mock")
}

// publishEvent publishes an event to the change stream
// This is called by MemoryStorage when data changes occur
// Returns true if event was published, false if stream is closed or buffer is full
func (m *MockChangeStream) publishEvent(event ChangeEvent) bool {
	m.mu.RLock()
	closed := m.closed
	m.mu.RUnlock()

	if closed {
		log.Debugf("cannot publish event - stream is closed")
		return false
	}

	// Apply filter
	if !m.matchesFilter(event) {
		log.Debugf("event filtered out: %s on %s", event.OperationType, event.CollectionName)
		return false
	}

	select {
	case m.events <- event:
		log.Debugf("published event: %s on %s", event.OperationType, event.CollectionName)
		return true

	case <-m.ctx.Done():
		log.Debugf("cannot publish event - context cancelled")
		return false

	default:
		// Channel buffer is full, drop the event
		log.Warnf("event buffer full, dropping event: %s on %s", event.OperationType, event.CollectionName)
		return false
	}
}

// matchesFilter checks if an event matches the stream's filter
func (m *MockChangeStream) matchesFilter(event ChangeEvent) bool {
	// Check collection filter
	if m.filter.CollectionName != "" && m.filter.CollectionName != event.CollectionName {
		return false
	}

	// Check operation type filter
	if len(m.filter.OperationTypes) > 0 {
		matched := false
		for _, opType := range m.filter.OperationTypes {
			if opType == event.OperationType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// extractDocumentKey extracts the document key (_id) from a document
func extractDocumentKey(doc bson.M) bson.M {
	if doc == nil {
		return bson.M{}
	}

	if id, exists := doc["_id"]; exists {
		return bson.M{"_id": id}
	}

	return bson.M{}
}

package mock

import (
	"context"
	"time"

	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
)

// MockCursor implements MongoCursor interface for testing
// It provides an in-memory cursor implementation with configurable behavior
type MockCursor struct {
	documents []bson.M // All documents in the cursor
	position  int      // Current position in documents (-1 means before first)
	closed    bool     // Whether the cursor is closed
	err       error    // Error to return (if any)
}

// NewMockCursor creates a new MockCursor with the given documents
// The cursor position starts at -1 (before the first document)
// Call Next() to move to the first document
func NewMockCursor(documents []bson.M) *MockCursor {
	return &MockCursor{
		documents: documents,
		position:  -1,
		closed:    false,
	}
}

// All decodes all documents in the cursor to the results parameter
// results must be a pointer to a slice
// This method processes all remaining documents and closes the cursor
func (m *MockCursor) All(ctx context.Context, results interface{}) error {
	if m.closed {
		log.Errorf("err:cursor is closed")
		return ErrInvalidArgument
	}

	if results == nil {
		log.Errorf("err:results is nil")
		return ErrInvalidArgument
	}

	// Get all remaining documents
	remaining := m.documents[m.position+1:]

	err := bsonArrayToStruct(remaining, results)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// Move position to end and close cursor
	m.position = len(m.documents) - 1
	m.closed = true

	return nil
}

// Close marks the cursor as closed
// After closing, most cursor operations will fail
func (m *MockCursor) Close(ctx context.Context) error {
	m.closed = true
	return nil
}

// Decode decodes the current document into val
// val must be a pointer to the target type
// Returns error if cursor is not positioned on a valid document
func (m *MockCursor) Decode(val interface{}) error {
	if m.closed {
		log.Errorf("err:cursor is closed")
		return ErrInvalidArgument
	}

	if val == nil {
		log.Errorf("err:val is nil")
		return ErrInvalidArgument
	}

	if m.position < 0 || m.position >= len(m.documents) {
		log.Errorf("err:invalid cursor position")
		return ErrInvalidArgument
	}

	err := bsonToStruct(m.documents[m.position], val)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

// Err returns the error state of the cursor
// Returns the configured error (if any)
func (m *MockCursor) Err() error {
	return m.err
}

// ID returns the cursor ID
// MockCursor always returns 0 as it's not a real server cursor
func (m *MockCursor) ID() int64 {
	return 0
}

// Next advances the cursor to the next document
// Returns true if there is a next document, false otherwise
// Call Decode() to retrieve the current document after Next() returns true
func (m *MockCursor) Next(ctx context.Context) bool {
	if m.closed {
		return false
	}

	if m.err != nil {
		return false
	}

	m.position++
	return m.position < len(m.documents)
}

// RemainingBatchLength returns the number of documents remaining in the cursor
// This counts all documents from the current position to the end
func (m *MockCursor) RemainingBatchLength() int {
	if m.closed {
		return 0
	}

	remaining := len(m.documents) - (m.position + 1)
	if remaining < 0 {
		return 0
	}

	return remaining
}

// SetBatchSize sets the batch size for the cursor
// MockCursor ignores this setting as it loads all documents in memory
func (m *MockCursor) SetBatchSize(batchSize int32) {
	// Mock implementation - no-op
}

// SetComment sets a comment for the cursor operation
// MockCursor ignores this setting as it's for testing only
func (m *MockCursor) SetComment(comment interface{}) {
	// Mock implementation - no-op
}

// SetMaxTime sets the maximum execution time for operations
// MockCursor ignores this setting as it's for testing only
func (m *MockCursor) SetMaxTime(dur time.Duration) {
	// Mock implementation - no-op
}

// TryNext attempts to advance the cursor to the next document
// Similar to Next(), but won't block waiting for documents
// In MockCursor, this behaves identically to Next()
func (m *MockCursor) TryNext(ctx context.Context) bool {
	return m.Next(ctx)
}

// Current returns the current document as BSON raw bytes
// Returns nil if cursor is not positioned on a valid document
func (m *MockCursor) Current() bson.Raw {
	if m.closed {
		return nil
	}

	if m.position < 0 || m.position >= len(m.documents) {
		return nil
	}

	// Marshal current document to BSON bytes
	bytes, err := bson.Marshal(m.documents[m.position])
	if err != nil {
		log.Errorf("err:%v", err)
		return nil
	}

	return bson.Raw(bytes)
}

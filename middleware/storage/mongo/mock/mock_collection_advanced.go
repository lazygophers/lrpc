package mock

import (
	"context"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	"go.mongodb.org/mongo-driver/bson"
	gomongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Aggregate executes an aggregation pipeline on the collection
// Supports common aggregation stages: $match, $project, $sort, $limit, $skip, $group, $unwind
// Returns a cursor containing the aggregated results
func (m *MockCollection) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (mongo.MongoCursor, error) {
	// Convert pipeline to []bson.M
	pipelineSlice, err := convertPipelineToBsonM(pipeline)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Get all documents from collection
	allDocs, err := m.storage.Find(m.name, bson.M{}, nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Process aggregation pipeline
	result, err := processAggregationPipeline(allDocs, pipelineSlice)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Return cursor with results
	return NewMockCursor(result), nil
}

// Distinct returns the distinct values for a specified field across the collection
// Supports simple field names (e.g. "name") and nested field names (e.g. "user.name")
// Applies filter to select documents before extracting distinct values
func (m *MockCollection) Distinct(ctx context.Context, fieldName string, filter interface{}, opts ...*options.DistinctOptions) ([]interface{}, error) {
	filterDoc, err := toBsonM(filter)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Find all matching documents
	documents, err := m.storage.Find(m.name, filterDoc, nil)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	// Use map to track distinct values
	// We need to use a unique key for each value
	distinctMap := make(map[string]interface{})
	var result []interface{}

	for _, doc := range documents {
		// Extract field value using getNestedValue for dot notation support
		value, exists := m.storage.getNestedValue(doc, fieldName)
		if !exists {
			continue
		}

		// Create a unique key for the value
		// We use bson.Marshal to create a stable key representation
		keyBytes, err := bson.Marshal(bson.M{"v": value})
		if err != nil {
			log.Errorf("err:%v", err)
			continue
		}
		key := string(keyBytes)

		// Check if we've seen this value before
		if _, seen := distinctMap[key]; !seen {
			distinctMap[key] = value
			result = append(result, value)
		}
	}

	return result, nil
}

// Watch creates a change stream to watch for changes to the collection
// Returns ErrNotImplemented as mongo.ChangeStream is a concrete type that cannot be mocked
// Use WatchMock() instead for testing with change streams
func (m *MockCollection) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*gomongo.ChangeStream, error) {
	err := ErrNotImplemented
	log.Errorf("err:%v - use WatchMock() instead for testing", err)
	return nil, err
}

// WatchMock creates a mock change stream to watch for changes to the collection
// This is a testing-specific method that returns MockChangeStream directly
// Use this method in tests instead of Watch()
func (m *MockCollection) WatchMock(ctx context.Context) *MockChangeStream {
	log.Debugf("creating mock change stream for collection: %s", m.name)

	// Create filter for this collection only
	filter := ChangeStreamFilter{
		CollectionName: m.name,
	}

	// Create the change stream
	stream := NewMockChangeStream(ctx, filter, 100)

	// Register with storage
	m.storage.registerChangeStream(stream)

	return stream
}

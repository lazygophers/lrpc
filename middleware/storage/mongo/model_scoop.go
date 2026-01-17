package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ModelScoop[M] is a type-safe query builder wrapping Scoop with model-specific methods
type ModelScoop[M any] struct {
	*Scoop
	m M
}

// Find finds documents matching the filter and returns typed results
func (ms *ModelScoop[M]) Find() ([]M, error) {
	var results []M
	err := ms.Scoop.Find(&results)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// First finds a single document matching the filter
func (ms *ModelScoop[M]) First() (*M, error) {
	var result M
	err := ms.Scoop.First(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Count counts documents matching the filter
func (ms *ModelScoop[M]) Count() (int64, error) {
	ms.Scoop.Collection(ms.m)
	return ms.Scoop.Count()
}

// Create creates a new document
func (ms *ModelScoop[M]) Create(doc M) error {
	return ms.Scoop.Create(doc)
}

// Update updates documents matching the filter
func (ms *ModelScoop[M]) Update(update interface{}) (int64, error) {
	ms.Scoop.Collection(ms.m)
	return ms.Scoop.Update(update)
}

// Delete deletes documents matching the filter
func (ms *ModelScoop[M]) Delete() (int64, error) {
	ms.Scoop.Collection(ms.m)
	return ms.Scoop.Delete()
}

// Exist checks if documents matching the filter exist
func (ms *ModelScoop[M]) Exist() (bool, error) {
	ms.Scoop.Collection(ms.m)
	return ms.Scoop.Exist()
}

// Watch watches for changes on this model's collection
func (ms *ModelScoop[M]) Watch(pipeline ...bson.M) (*ChangeStream, error) {
	return ms.Scoop.WatchChanges(pipeline...)
}

// Aggregate creates an aggregation pipeline for this model
func (ms *ModelScoop[M]) Aggregate(pipeline ...bson.M) *Aggregation {
	return ms.Scoop.Aggregate(pipeline...)
}

// GetCollection returns the underlying MongoDB collection
func (ms *ModelScoop[M]) GetCollection() *mongo.Collection {
	return ms.Scoop.GetCollection()
}

// GetScoop returns the underlying Scoop for advanced operations
func (ms *ModelScoop[M]) GetScoop() *Scoop {
	return ms.Scoop
}

// Where adds a filter condition
func (ms *ModelScoop[M]) Where(key string, value interface{}) *ModelScoop[M] {
	ms.Scoop.Where(key, value)
	return ms
}

// Limit sets the result limit
func (ms *ModelScoop[M]) Limit(limit int64) *ModelScoop[M] {
	ms.Scoop.Limit(limit)
	return ms
}

// Offset sets the result offset
func (ms *ModelScoop[M]) Offset(offset int64) *ModelScoop[M] {
	ms.Scoop.Offset(offset)
	return ms
}

// Sort adds sorting
// direction: 1 for ascending (default), -1 for descending
func (ms *ModelScoop[M]) Sort(key string, direction ...int) *ModelScoop[M] {
	ms.Scoop.Sort(key, direction...)
	return ms
}

// Select specifies which fields to return
func (ms *ModelScoop[M]) Select(fields ...string) *ModelScoop[M] {
	ms.Scoop.Select(fields...)
	return ms
}

// Equal adds an equality condition
func (ms *ModelScoop[M]) Equal(key string, value interface{}) *ModelScoop[M] {
	ms.Scoop.Equal(key, value)
	return ms
}

// Ne adds a != condition using $ne operator
func (ms *ModelScoop[M]) Ne(key string, value interface{}) *ModelScoop[M] {
	ms.Scoop.Ne(key, value)
	return ms
}

// In adds an $in condition
func (ms *ModelScoop[M]) In(key string, values ...interface{}) *ModelScoop[M] {
	ms.Scoop.In(key, values...)
	return ms
}

// NotIn adds a $nin condition
func (ms *ModelScoop[M]) NotIn(key string, values ...interface{}) *ModelScoop[M] {
	ms.Scoop.NotIn(key, values...)
	return ms
}

// Like adds a regex pattern match (case-insensitive)
func (ms *ModelScoop[M]) Like(key string, pattern string) *ModelScoop[M] {
	ms.Scoop.Like(key, pattern)
	return ms
}

// Gt adds a > condition using $gt operator
func (ms *ModelScoop[M]) Gt(key string, value interface{}) *ModelScoop[M] {
	ms.Scoop.Gt(key, value)
	return ms
}

// Lt adds a < condition using $lt operator
func (ms *ModelScoop[M]) Lt(key string, value interface{}) *ModelScoop[M] {
	ms.Scoop.Lt(key, value)
	return ms
}

// Gte adds a >= condition using $gte operator
func (ms *ModelScoop[M]) Gte(key string, value interface{}) *ModelScoop[M] {
	ms.Scoop.Gte(key, value)
	return ms
}

// Lte adds a <= condition using $lte operator
func (ms *ModelScoop[M]) Lte(key string, value interface{}) *ModelScoop[M] {
	ms.Scoop.Lte(key, value)
	return ms
}

// Between adds a $gte and $lte condition
func (ms *ModelScoop[M]) Between(key string, min interface{}, max interface{}) *ModelScoop[M] {
	ms.Scoop.Between(key, min, max)
	return ms
}

// Skip is an alias for Offset
func (ms *ModelScoop[M]) Skip(skip int64) *ModelScoop[M] {
	return ms.Offset(skip)
}

// Clear resets the scoop
func (ms *ModelScoop[M]) Clear() *ModelScoop[M] {
	ms.Scoop.Clear()
	return ms
}

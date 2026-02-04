package mongo

import (
	"reflect"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	collectionNameCacheMu sync.RWMutex
	collectionNameCache   = make(map[reflect.Type]string)
)

// Model[M] is a lightweight model wrapper managing model metadata and client reference
type Model[M any] struct {
	client         *Client
	model          M
	collectionName string
	notFoundError  error
}

// NewModel creates a new model wrapper
func NewModel[M any](client *Client) *Model[M] {
	// Get the type of M
	var m M
	rt := reflect.TypeOf(m)
	
	return &Model[M]{
		client:         client,
		model:          *new(M),
		collectionName: getCollectionNameFromType(rt),
		notFoundError:  mongo.ErrNoDocuments,
	}
}

// getCollectionNameFromType retrieves the collection name from a type
// It supports multiple methods:
// 1. Caching - checks cache first for performance
// 2. Collectioner interface - Collection() method
// 3. Reflection - looks for Collection() method on the type
// 4. Default - uses the type name (e.g., "User" -> "User")
func getCollectionNameFromType(t reflect.Type) string {
	if t == nil {
		return ""
	}

	// Unwrap pointer types to get the actual type
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check cache first with read lock
	collectionNameCacheMu.RLock()
	cached, ok := collectionNameCache[t]
	collectionNameCacheMu.RUnlock()
	if ok {
		return cached
	}

	// Determine collection name
	collName := determineCollectionName(t)

	// Cache the result with write lock
	collectionNameCacheMu.Lock()
	collectionNameCache[t] = collName
	collectionNameCacheMu.Unlock()

	return collName
}

// determineCollectionName determines the collection name using multiple strategies
func determineCollectionName(t reflect.Type) string {
	// Strategy 1: Look for Collection() method via reflection and try to call it
	if method, ok := t.MethodByName("Collection"); ok && method.Type.NumOut() == 1 && method.Type.Out(0).Kind() == reflect.String {
		// Create a value to call the method on
		v := reflect.New(t).Elem()
		if results := method.Func.Call([]reflect.Value{v}); len(results) == 1 {
			if collName := results[0].String(); collName != "" {
				return collName
			}
		}
	}

	// Strategy 2: Use the type name as collection name (default)
	return t.Name()
}

// getCollectionName retrieves the collection name from a model instance
// This function is kept for backward compatibility
// Deprecated: Use getCollectionNameFromType instead
func getCollectionName(m any) string {
	if m == nil {
		return ""
	}

	// Use reflection to get the type and retrieve collection name
	t := reflect.TypeOf(m)
	return getCollectionNameFromType(t)
}

// NewScoop creates a type-safe query builder for this model, optionally accepting a transaction scoop
func (m *Model[M]) NewScoop(tx ...*Scoop) *ModelScoop[M] {
	var baseScoop *Scoop
	if len(tx) > 0 && tx[0] != nil {
		baseScoop = m.client.NewScoop(tx[0]).CollectionName(m.collectionName)
	} else {
		baseScoop = m.client.NewScoop().CollectionName(m.collectionName)
	}

	baseScoop.SetNotFound(m.notFoundError)

	return &ModelScoop[M]{
		Scoop: baseScoop,
		m:     m.model,
	}
}

// CollectionName returns the collection name for this model
func (m *Model[M]) CollectionName() string {
	return m.collectionName
}

// SetNotFound sets the not found error for this model
func (m *Model[M]) SetNotFound(err error) *Model[M] {
	m.notFoundError = err
	return m
}

// IsNotFound checks if the error is a not found error
func (m *Model[M]) IsNotFound(err error) bool {
	return err == m.notFoundError || err == mongo.ErrNoDocuments
}

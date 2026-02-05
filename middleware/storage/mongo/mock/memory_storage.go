package mock

import (
	"fmt"
	"sort"
	"sync"

	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
)

// MemoryStorage represents an in-memory storage for MongoDB mock
// It stores collections and documents in memory using thread-safe structures
type MemoryStorage struct {
	mu          sync.RWMutex
	collections map[string]*Collection
	// Change stream support
	changeStreams []*MockChangeStream
	streamsMu     sync.RWMutex
}

// Collection represents a collection in memory storage
// It stores documents as BSON maps for flexibility
type Collection struct {
	name      string
	documents []bson.M
}

// FindOptions represents options for Find operations
type FindOptions struct {
	Limit      *int64
	Skip       *int64
	Sort       bson.M
	Projection bson.M
}

// NewMemoryStorage creates a new in-memory storage instance
// Returns a pointer to MemoryStorage with initialized collections map
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		collections:   make(map[string]*Collection),
		changeStreams: make([]*MockChangeStream, 0),
	}
}

// CreateCollection creates a new collection with the given name
// Returns error if collection already exists
func (m *MemoryStorage) CreateCollection(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.collections[name]; exists {
		err := fmt.Errorf("collection %s already exists", name)
		log.Errorf("err:%v", err)
		return err
	}

	m.collections[name] = &Collection{
		name:      name,
		documents: make([]bson.M, 0),
	}

	log.Infof("created collection: %s", name)
	return nil
}

// getOrCreateCollection retrieves an existing collection or creates a new one
// This is a private helper method for internal use
// Always returns a valid collection pointer, never nil
func (m *MemoryStorage) getOrCreateCollection(name string) *Collection {
	m.mu.Lock()
	defer m.mu.Unlock()

	if coll, exists := m.collections[name]; exists {
		return coll
	}

	coll := &Collection{
		name:      name,
		documents: make([]bson.M, 0),
	}
	m.collections[name] = coll

	log.Infof("auto-created collection: %s", name)
	return coll
}

// getCollection retrieves an existing collection by name
// Returns nil if collection doesn't exist
func (m *MemoryStorage) getCollection(name string) *Collection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.collections[name]
}

// Insert inserts a single document into the specified collection
// Creates the collection if it doesn't exist
// Returns error if document insertion fails
func (m *MemoryStorage) Insert(collName string, doc bson.M) error {
	if doc == nil {
		err := fmt.Errorf("cannot insert nil document")
		log.Errorf("err:%v", err)
		return err
	}

	coll := m.getOrCreateCollection(collName)

	m.mu.Lock()
	coll.documents = append(coll.documents, doc)
	m.mu.Unlock()

	// Publish change event
	m.publishChangeEvent(ChangeEvent{
		OperationType:  ChangeEventInsert,
		CollectionName: collName,
		DocumentKey:    extractDocumentKey(doc),
		FullDocument:   m.copyDocument(doc),
	})

	log.Debugf("inserted document into collection %s: %v", collName, doc)
	return nil
}

// InsertMany inserts multiple documents into the specified collection
// Creates the collection if it doesn't exist
// Returns error if any document insertion fails
func (m *MemoryStorage) InsertMany(collName string, docs []bson.M) error {
	if len(docs) == 0 {
		return nil
	}

	for _, doc := range docs {
		if doc == nil {
			err := fmt.Errorf("cannot insert nil document")
			log.Errorf("err:%v", err)
			return err
		}
	}

	coll := m.getOrCreateCollection(collName)

	m.mu.Lock()
	coll.documents = append(coll.documents, docs...)
	m.mu.Unlock()

	// Publish change events for each inserted document
	for _, doc := range docs {
		m.publishChangeEvent(ChangeEvent{
			OperationType:  ChangeEventInsert,
			CollectionName: collName,
			DocumentKey:    extractDocumentKey(doc),
			FullDocument:   m.copyDocument(doc),
		})
	}

	log.Debugf("inserted %d documents into collection %s", len(docs), collName)
	return nil
}

// Find finds documents in the specified collection that match the filter
// Applies sorting, pagination, and projection options if provided
// Returns matched documents and any error encountered
func (m *MemoryStorage) Find(collName string, filter bson.M, opts *FindOptions) ([]bson.M, error) {
	coll := m.getCollection(collName)
	if coll == nil {
		log.Debugf("collection %s not found, returning empty result", collName)
		return []bson.M{}, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Filter documents
	var matched []bson.M
	for _, doc := range coll.documents {
		if m.matchFilter(doc, filter) {
			// Deep copy document to prevent external modifications
			matched = append(matched, m.copyDocument(doc))
		}
	}

	// Apply sorting if specified
	if opts != nil && len(opts.Sort) > 0 {
		matched = m.applySorting(matched, opts.Sort)
	}

	// Apply pagination
	if opts != nil {
		if opts.Skip != nil && *opts.Skip > 0 {
			skip := int(*opts.Skip)
			if skip >= len(matched) {
				matched = []bson.M{}
			} else {
				matched = matched[skip:]
			}
		}

		if opts.Limit != nil && *opts.Limit > 0 {
			limit := int(*opts.Limit)
			if limit < len(matched) {
				matched = matched[:limit]
			}
		}
	}

	// Apply projection if specified
	if opts != nil && len(opts.Projection) > 0 {
		matched = m.applyProjection(matched, opts.Projection)
	}

	log.Debugf("found %d documents in collection %s", len(matched), collName)
	return matched, nil
}

// Count counts the number of documents in the specified collection that match the filter
// Returns 0 if collection doesn't exist
func (m *MemoryStorage) Count(collName string, filter bson.M) (int64, error) {
	coll := m.getCollection(collName)
	if coll == nil {
		log.Debugf("collection %s not found, returning count 0", collName)
		return 0, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	count := int64(0)
	for _, doc := range coll.documents {
		if m.matchFilter(doc, filter) {
			count++
		}
	}

	log.Debugf("counted %d documents in collection %s", count, collName)
	return count, nil
}

// Delete deletes documents from the specified collection that match the filter
// Returns the number of documents deleted and any error encountered
func (m *MemoryStorage) Delete(collName string, filter bson.M) (int64, error) {
	coll := m.getCollection(collName)
	if coll == nil {
		log.Debugf("collection %s not found, no documents deleted", collName)
		return 0, nil
	}

	m.mu.Lock()
	var remaining []bson.M
	deletedDocs := make([]bson.M, 0)
	deleted := int64(0)

	for _, doc := range coll.documents {
		if m.matchFilter(doc, filter) {
			deletedDocs = append(deletedDocs, extractDocumentKey(doc))
			deleted++
		} else {
			remaining = append(remaining, doc)
		}
	}

	coll.documents = remaining
	m.mu.Unlock()

	// Publish change events for each deleted document
	for _, docKey := range deletedDocs {
		m.publishChangeEvent(ChangeEvent{
			OperationType:  ChangeEventDelete,
			CollectionName: collName,
			DocumentKey:    docKey,
		})
	}

	log.Debugf("deleted %d documents from collection %s", deleted, collName)
	return deleted, nil
}

// DeleteOne deletes a single document from the specified collection that matches the filter
// Returns the number of documents deleted (0 or 1) and any error encountered
func (m *MemoryStorage) DeleteOne(collName string, filter bson.M) (int64, error) {
	coll := m.getCollection(collName)
	if coll == nil {
		log.Debugf("collection %s not found, no documents deleted", collName)
		return 0, nil
	}

	m.mu.Lock()
	var docKey bson.M

	for i, doc := range coll.documents {
		if m.matchFilter(doc, filter) {
			docKey = extractDocumentKey(doc)
			// Remove the document at index i
			coll.documents = append(coll.documents[:i], coll.documents[i+1:]...)
			break
		}
	}
	m.mu.Unlock()

	if docKey != nil {
		// Publish change event
		m.publishChangeEvent(ChangeEvent{
			OperationType:  ChangeEventDelete,
			CollectionName: collName,
			DocumentKey:    docKey,
		})

		log.Debugf("deleted 1 document from collection %s", collName)
		return 1, nil
	}

	log.Debugf("no matching document found in collection %s", collName)
	return 0, nil
}

// DropCollection drops the specified collection and all its documents
// Returns error if collection doesn't exist
func (m *MemoryStorage) DropCollection(collName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.collections[collName]; !exists {
		err := fmt.Errorf("collection %s does not exist", collName)
		log.Errorf("err:%v", err)
		return err
	}

	delete(m.collections, collName)
	log.Infof("dropped collection: %s", collName)
	return nil
}

// ListCollections returns the names of all collections in the storage
func (m *MemoryStorage) ListCollections() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.collections))
	for name := range m.collections {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// Clear removes all collections and documents from storage
// Useful for test cleanup
func (m *MemoryStorage) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.collections = make(map[string]*Collection)
	log.Debugf("cleared all collections from memory storage")
}

package mock

import (
	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
)

// Update updates documents in the specified collection that match the filter
// Applies update operations to matched documents
// Returns the number of documents updated
func (m *MemoryStorage) Update(collName string, filter bson.M, update bson.M) int64 {
	coll := m.getCollection(collName)
	if coll == nil {
		log.Debugf("collection %s not found, no documents updated", collName)
		return 0
	}

	m.mu.Lock()
	updated := int64(0)
	updatedDocs := make([]bson.M, 0)

	for i := range coll.documents {
		if m.matchFilter(coll.documents[i], filter) {
			docKey := extractDocumentKey(coll.documents[i])
			m.applyUpdate(&coll.documents[i], update)
			updatedDocs = append(updatedDocs, docKey)
			updated++
		}
	}
	m.mu.Unlock()

	// Publish change events for each updated document
	for _, docKey := range updatedDocs {
		m.publishChangeEvent(ChangeEvent{
			OperationType:  ChangeEventUpdate,
			CollectionName: collName,
			DocumentKey:    docKey,
			UpdateDescription: &UpdateDescription{
				UpdatedFields: m.extractUpdatedFields(update),
				RemovedFields: m.extractRemovedFields(update),
			},
		})
	}

	log.Debugf("updated %d documents in collection %s", updated, collName)
	return updated
}

// UpdateOne updates a single document in the specified collection that matches the filter
// Applies update operations to the first matched document
// Returns the number of documents updated (0 or 1)
func (m *MemoryStorage) UpdateOne(collName string, filter bson.M, update bson.M) int64 {
	coll := m.getCollection(collName)
	if coll == nil {
		log.Debugf("collection %s not found, no documents updated", collName)
		return 0
	}

	m.mu.Lock()
	var updatedDoc *bson.M
	var docKey bson.M

	for i := range coll.documents {
		if m.matchFilter(coll.documents[i], filter) {
			docKey = extractDocumentKey(coll.documents[i])
			m.applyUpdate(&coll.documents[i], update)
			updatedDoc = &coll.documents[i]
			break
		}
	}
	m.mu.Unlock()

	if updatedDoc != nil {
		// Publish change event
		m.publishChangeEvent(ChangeEvent{
			OperationType:  ChangeEventUpdate,
			CollectionName: collName,
			DocumentKey:    docKey,
			UpdateDescription: &UpdateDescription{
				UpdatedFields: m.extractUpdatedFields(update),
				RemovedFields: m.extractRemovedFields(update),
			},
		})

		log.Debugf("updated 1 document in collection %s", collName)
		return 1
	}

	log.Debugf("no matching document found in collection %s", collName)
	return 0
}

// ReplaceOne replaces a single document in the specified collection that matches the filter
// This method publishes a replace event instead of separate delete+insert events
// Returns the number of documents replaced (0 or 1)
func (m *MemoryStorage) ReplaceOne(collName string, filter bson.M, replacement bson.M) int64 {
	coll := m.getCollection(collName)
	if coll == nil {
		log.Debugf("collection %s not found, no documents replaced", collName)
		return 0
	}

	m.mu.Lock()
	var docKey bson.M
	replaced := false

	for i, doc := range coll.documents {
		if m.matchFilter(doc, filter) {
			docKey = extractDocumentKey(doc)
			// Preserve _id from original document if replacement doesn't have one
			if _, hasID := replacement["_id"]; !hasID {
				if id, hasOriginalID := doc["_id"]; hasOriginalID {
					replacement["_id"] = id
				}
			}
			// Replace the document
			coll.documents[i] = replacement
			replaced = true
			break
		}
	}
	m.mu.Unlock()

	if replaced {
		// Publish replace event
		m.publishChangeEvent(ChangeEvent{
			OperationType:  ChangeEventReplace,
			CollectionName: collName,
			DocumentKey:    docKey,
			FullDocument:   m.copyDocument(replacement),
		})

		log.Debugf("replaced 1 document in collection %s", collName)
		return 1
	}

	log.Debugf("no matching document found in collection %s", collName)
	return 0
}

// applyUpdate applies update operations to a document
// Supports MongoDB update operators like $set, $inc, $unset, etc.
func (m *MemoryStorage) applyUpdate(doc *bson.M, update bson.M) {
	// Handle update operators
	for operator, value := range update {
		switch operator {
		case "$set":
			if setFields, ok := value.(bson.M); ok {
				for field, val := range setFields {
					(*doc)[field] = val
				}
			}

		case "$unset":
			if unsetFields, ok := value.(bson.M); ok {
				for field := range unsetFields {
					delete(*doc, field)
				}
			}

		case "$inc":
			if incFields, ok := value.(bson.M); ok {
				for field, incVal := range incFields {
					m.applyIncrement(doc, field, incVal)
				}
			}

		case "$mul":
			if mulFields, ok := value.(bson.M); ok {
				for field, mulVal := range mulFields {
					m.applyMultiply(doc, field, mulVal)
				}
			}

		default:
			// If no operator is specified, treat as direct replacement
			// This handles the case where update is just { "field": "value" }
			if !isOperator(operator) {
				(*doc)[operator] = value
			}
		}
	}
}

// isOperator checks if a string is a MongoDB update operator
func isOperator(s string) bool {
	return len(s) > 0 && s[0] == '$'
}

// applyIncrement increments a numeric field by the specified value
func (m *MemoryStorage) applyIncrement(doc *bson.M, field string, incVal interface{}) {
	currentVal, exists := (*doc)[field]
	if !exists {
		(*doc)[field] = incVal
		return
	}

	// Handle different numeric types
	switch current := currentVal.(type) {
	case int:
		if inc, ok := incVal.(int); ok {
			(*doc)[field] = current + inc
		}
	case int32:
		if inc, ok := incVal.(int32); ok {
			(*doc)[field] = current + inc
		}
	case int64:
		if inc, ok := incVal.(int64); ok {
			(*doc)[field] = current + inc
		}
	case float64:
		if inc, ok := incVal.(float64); ok {
			(*doc)[field] = current + inc
		}
	default:
		// If type doesn't match, set to incVal
		(*doc)[field] = incVal
	}
}

// applyMultiply multiplies a numeric field by the specified value
func (m *MemoryStorage) applyMultiply(doc *bson.M, field string, mulVal interface{}) {
	currentVal, exists := (*doc)[field]
	if !exists {
		(*doc)[field] = 0
		return
	}

	// Handle different numeric types
	switch current := currentVal.(type) {
	case int:
		if mul, ok := mulVal.(int); ok {
			(*doc)[field] = current * mul
		}
	case int32:
		if mul, ok := mulVal.(int32); ok {
			(*doc)[field] = current * mul
		}
	case int64:
		if mul, ok := mulVal.(int64); ok {
			(*doc)[field] = current * mul
		}
	case float64:
		if mul, ok := mulVal.(float64); ok {
			(*doc)[field] = current * mul
		}
	default:
		// If type doesn't match, set to 0
		(*doc)[field] = 0
	}
}

// extractUpdatedFields extracts updated fields from an update document
func (m *MemoryStorage) extractUpdatedFields(update bson.M) bson.M {
	updatedFields := make(bson.M)

	if setFields, ok := update["$set"].(bson.M); ok {
		for k, v := range setFields {
			updatedFields[k] = v
		}
	}

	if incFields, ok := update["$inc"].(bson.M); ok {
		for k, v := range incFields {
			updatedFields[k] = v
		}
	}

	if mulFields, ok := update["$mul"].(bson.M); ok {
		for k, v := range mulFields {
			updatedFields[k] = v
		}
	}

	return updatedFields
}

// extractRemovedFields extracts removed field names from an update document
func (m *MemoryStorage) extractRemovedFields(update bson.M) []string {
	removedFields := make([]string, 0)

	if unsetFields, ok := update["$unset"].(bson.M); ok {
		for k := range unsetFields {
			removedFields = append(removedFields, k)
		}
	}

	return removedFields
}

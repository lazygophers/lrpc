package mock

import (
	"sort"

	"go.mongodb.org/mongo-driver/bson"
)

// copyDocument creates a deep copy of a document to prevent external modifications
func (m *MemoryStorage) copyDocument(doc bson.M) bson.M {
	copied := make(bson.M, len(doc))
	for key, value := range doc {
		copied[key] = value
	}
	return copied
}

// applySorting sorts documents based on sort specification
// sortSpec is a bson.M where keys are field names and values are 1 (ascending) or -1 (descending)
func (m *MemoryStorage) applySorting(docs []bson.M, sortSpec bson.M) []bson.M {
	if len(docs) == 0 || len(sortSpec) == 0 {
		return docs
	}

	// For simplicity, we only support single field sorting for now
	// Multi-field sorting can be added later
	var sortField string
	var sortOrder int

	for field, order := range sortSpec {
		sortField = field
		if orderInt, ok := order.(int); ok {
			sortOrder = orderInt
		} else {
			sortOrder = 1 // default ascending
		}
		break // Only take the first sort field
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]bson.M, len(docs))
	copy(sorted, docs)

	sort.Slice(sorted, func(i, j int) bool {
		valI := sorted[i][sortField]
		valJ := sorted[j][sortField]

		// Handle nil values
		if valI == nil && valJ == nil {
			return false
		}
		if valI == nil {
			return sortOrder > 0 // nil comes first in ascending
		}
		if valJ == nil {
			return sortOrder < 0 // nil comes last in ascending
		}

		// Compare based on type
		result := m.compareValuesForSort(valI, valJ)
		if sortOrder < 0 {
			result = -result
		}
		return result < 0
	})

	return sorted
}

// compareValuesForSort compares two values for sorting
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func (m *MemoryStorage) compareValuesForSort(a, b interface{}) int {
	// Type-based comparison
	switch aVal := a.(type) {
	case int:
		if bVal, ok := b.(int); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	case int32:
		if bVal, ok := b.(int32); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	case int64:
		if bVal, ok := b.(int64); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	case string:
		if bVal, ok := b.(string); ok {
			if aVal < bVal {
				return -1
			} else if aVal > bVal {
				return 1
			}
			return 0
		}
	}

	// Default: consider equal
	return 0
}

// applyProjection applies field projection to documents
// projectionSpec is a bson.M where keys are field names and values are 1 (include) or 0 (exclude)
func (m *MemoryStorage) applyProjection(docs []bson.M, projectionSpec bson.M) []bson.M {
	if len(docs) == 0 || len(projectionSpec) == 0 {
		return docs
	}

	// Determine if this is inclusion or exclusion projection
	isInclusion := false
	for _, value := range projectionSpec {
		if intVal, ok := value.(int); ok && intVal == 1 {
			isInclusion = true
			break
		}
	}

	projected := make([]bson.M, len(docs))
	for i, doc := range docs {
		projectedDoc := make(bson.M)

		if isInclusion {
			// Include only specified fields
			for field, include := range projectionSpec {
				if intVal, ok := include.(int); ok && intVal == 1 {
					if value, exists := doc[field]; exists {
						projectedDoc[field] = value
					}
				}
			}
			// Always include _id unless explicitly excluded
			if _, hasId := projectionSpec["_id"]; !hasId {
				if idVal, exists := doc["_id"]; exists {
					projectedDoc["_id"] = idVal
				}
			}
		} else {
			// Exclude specified fields
			for key, value := range doc {
				excluded := false
				if excludeVal, exists := projectionSpec[key]; exists {
					if intVal, ok := excludeVal.(int); ok && intVal == 0 {
						excluded = true
					}
				}
				if !excluded {
					projectedDoc[key] = value
				}
			}
		}

		projected[i] = projectedDoc
	}

	return projected
}

// registerChangeStream registers a change stream to receive events
func (m *MemoryStorage) registerChangeStream(stream *MockChangeStream) {
	m.streamsMu.Lock()
	defer m.streamsMu.Unlock()

	m.changeStreams = append(m.changeStreams, stream)
}

// unregisterChangeStream removes a change stream from receiving events
func (m *MemoryStorage) unregisterChangeStream(stream *MockChangeStream) {
	m.streamsMu.Lock()
	defer m.streamsMu.Unlock()

	for i, s := range m.changeStreams {
		if s == stream {
			// Remove from slice
			m.changeStreams = append(m.changeStreams[:i], m.changeStreams[i+1:]...)
			return
		}
	}
}

// publishChangeEvent publishes a change event to all registered change streams
func (m *MemoryStorage) publishChangeEvent(event ChangeEvent) {
	m.streamsMu.RLock()
	defer m.streamsMu.RUnlock()

	if len(m.changeStreams) == 0 {
		return
	}

	for _, stream := range m.changeStreams {
		stream.publishEvent(event)
	}
}

package mock

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/lazygophers/log"
	"go.mongodb.org/mongo-driver/bson"
)

// matchFilter checks if a document matches the filter criteria
// Supports MongoDB query operators and nested field queries
// Returns true if document matches all filter conditions, false otherwise
func (m *MemoryStorage) matchFilter(doc bson.M, filter bson.M) bool {
	// Empty filter matches all documents
	if len(filter) == 0 {
		return true
	}

	// Check each filter condition
	for field, filterValue := range filter {
		// Handle logical operators
		if strings.HasPrefix(field, "$") {
			matched := m.matchLogicalOperator(doc, field, filterValue)
			if !matched {
				return false
			}
			continue
		}

		// Get document value (supports nested fields like "user.name")
		docValue, exists := m.getNestedValue(doc, field)

		// Check if filterValue contains operators
		if filterMap, ok := filterValue.(bson.M); ok {
			matched := m.matchFieldOperators(docValue, exists, filterMap)
			if !matched {
				return false
			}
			continue
		}

		// Simple equality check
		if !exists {
			return false
		}
		if !m.matchValue(docValue, filterValue) {
			return false
		}
	}

	return true
}

// matchLogicalOperator handles logical operators: $and, $or, $not
func (m *MemoryStorage) matchLogicalOperator(doc bson.M, operator string, value interface{}) bool {
	switch operator {
	case "$and":
		conditions, ok := value.([]interface{})
		if !ok {
			log.Errorf("err:%v", fmt.Errorf("$and operator requires array of conditions"))
			return false
		}
		for _, condition := range conditions {
			conditionMap, ok := condition.(bson.M)
			if !ok {
				log.Errorf("err:%v", fmt.Errorf("$and condition must be bson.M"))
				return false
			}
			if !m.matchFilter(doc, conditionMap) {
				return false
			}
		}
		return true

	case "$or":
		conditions, ok := value.([]interface{})
		if !ok {
			log.Errorf("err:%v", fmt.Errorf("$or operator requires array of conditions"))
			return false
		}
		for _, condition := range conditions {
			conditionMap, ok := condition.(bson.M)
			if !ok {
				log.Errorf("err:%v", fmt.Errorf("$or condition must be bson.M"))
				return false
			}
			if m.matchFilter(doc, conditionMap) {
				return true
			}
		}
		return false

	case "$nor":
		conditions, ok := value.([]interface{})
		if !ok {
			log.Errorf("err:%v", fmt.Errorf("$nor operator requires array of conditions"))
			return false
		}
		for _, condition := range conditions {
			conditionMap, ok := condition.(bson.M)
			if !ok {
				log.Errorf("err:%v", fmt.Errorf("$nor condition must be bson.M"))
				return false
			}
			if m.matchFilter(doc, conditionMap) {
				return false
			}
		}
		return true

	case "$not":
		conditionMap, ok := value.(bson.M)
		if !ok {
			log.Errorf("err:%v", fmt.Errorf("$not operator requires bson.M"))
			return false
		}
		return !m.matchFilter(doc, conditionMap)

	default:
		log.Errorf("err:%v", fmt.Errorf("unsupported logical operator: %s", operator))
		return false
	}
}

// matchFieldOperators checks if a field value matches operator conditions
// Supports comparison and array operators
func (m *MemoryStorage) matchFieldOperators(docValue interface{}, exists bool, operators bson.M) bool {
	for operator, operatorValue := range operators {
		switch operator {
		case "$eq":
			if !exists {
				return false
			}
			if !m.matchValue(docValue, operatorValue) {
				return false
			}

		case "$ne":
			if exists && m.matchValue(docValue, operatorValue) {
				return false
			}

		case "$gt":
			if !exists {
				return false
			}
			if !m.compareValues(docValue, operatorValue, "$gt") {
				return false
			}

		case "$gte":
			if !exists {
				return false
			}
			if !m.compareValues(docValue, operatorValue, "$gte") {
				return false
			}

		case "$lt":
			if !exists {
				return false
			}
			if !m.compareValues(docValue, operatorValue, "$lt") {
				return false
			}

		case "$lte":
			if !exists {
				return false
			}
			if !m.compareValues(docValue, operatorValue, "$lte") {
				return false
			}

		case "$in":
			if !exists {
				return false
			}
			if !m.matchInOperator(docValue, operatorValue) {
				return false
			}

		case "$nin":
			if exists && m.matchInOperator(docValue, operatorValue) {
				return false
			}

		case "$exists":
			expectExists, ok := operatorValue.(bool)
			if !ok {
				log.Errorf("err:%v", fmt.Errorf("$exists operator requires boolean value"))
				return false
			}
			if exists != expectExists {
				return false
			}

		default:
			log.Errorf("err:%v", fmt.Errorf("unsupported field operator: %s", operator))
			return false
		}
	}

	return true
}

// matchValue checks if two values are equal
// Handles different types and performs deep equality comparison
func (m *MemoryStorage) matchValue(docValue, filterValue interface{}) bool {
	// Handle nil cases
	if docValue == nil && filterValue == nil {
		return true
	}
	if docValue == nil || filterValue == nil {
		return false
	}

	// Use reflect for deep equality
	return reflect.DeepEqual(docValue, filterValue)
}

// compareValues compares two values using the specified operator
// Supports comparison operators: $gt, $gte, $lt, $lte
// Returns true if comparison succeeds, false otherwise
func (m *MemoryStorage) compareValues(a, b interface{}, op string) bool {
	// Handle nil cases
	if a == nil || b == nil {
		return false
	}

	// Compare based on type
	switch aVal := a.(type) {
	case int:
		bVal, ok := m.toInt(b)
		if !ok {
			return false
		}
		return m.compareInts(int64(aVal), int64(bVal), op)

	case int32:
		bVal, ok := m.toInt(b)
		if !ok {
			return false
		}
		return m.compareInts(int64(aVal), int64(bVal), op)

	case int64:
		bVal, ok := m.toInt(b)
		if !ok {
			return false
		}
		return m.compareInts(aVal, int64(bVal), op)

	case float32:
		bVal, ok := m.toFloat(b)
		if !ok {
			return false
		}
		return m.compareFloats(float64(aVal), bVal, op)

	case float64:
		bVal, ok := m.toFloat(b)
		if !ok {
			return false
		}
		return m.compareFloats(aVal, bVal, op)

	case string:
		bVal, ok := b.(string)
		if !ok {
			return false
		}
		return m.compareStrings(aVal, bVal, op)

	default:
		log.Errorf("err:%v", fmt.Errorf("unsupported type for comparison: %T", a))
		return false
	}
}

// compareInts compares two int64 values based on operator
func (m *MemoryStorage) compareInts(a, b int64, op string) bool {
	switch op {
	case "$gt":
		return a > b
	case "$gte":
		return a >= b
	case "$lt":
		return a < b
	case "$lte":
		return a <= b
	default:
		return false
	}
}

// compareFloats compares two float64 values based on operator
func (m *MemoryStorage) compareFloats(a, b float64, op string) bool {
	switch op {
	case "$gt":
		return a > b
	case "$gte":
		return a >= b
	case "$lt":
		return a < b
	case "$lte":
		return a <= b
	default:
		return false
	}
}

// compareStrings compares two string values based on operator
func (m *MemoryStorage) compareStrings(a, b string, op string) bool {
	switch op {
	case "$gt":
		return a > b
	case "$gte":
		return a >= b
	case "$lt":
		return a < b
	case "$lte":
		return a <= b
	default:
		return false
	}
}

// toInt converts various numeric types to int64
func (m *MemoryStorage) toInt(v interface{}) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case float32:
		return int64(val), true
	case float64:
		return int64(val), true
	default:
		return 0, false
	}
}

// toFloat converts various numeric types to float64
func (m *MemoryStorage) toFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

// matchInOperator checks if value is in the provided array
// Supports $in operator for array membership test
func (m *MemoryStorage) matchInOperator(value interface{}, arrayValue interface{}) bool {
	// arrayValue should be a slice
	arr := reflect.ValueOf(arrayValue)
	if arr.Kind() != reflect.Slice && arr.Kind() != reflect.Array {
		log.Errorf("err:%v", fmt.Errorf("$in operator requires array value"))
		return false
	}

	// Check if value exists in array
	for i := 0; i < arr.Len(); i++ {
		item := arr.Index(i).Interface()
		if m.matchValue(value, item) {
			return true
		}
	}

	return false
}

// getNestedValue retrieves a value from a document using dot notation
// Supports nested field access like "user.name" or "address.city"
// Returns the value and a boolean indicating if the field exists
func (m *MemoryStorage) getNestedValue(doc bson.M, field string) (interface{}, bool) {
	// Handle simple field (no dot notation)
	if !strings.Contains(field, ".") {
		value, exists := doc[field]
		return value, exists
	}

	// Handle nested field with dot notation
	parts := strings.Split(field, ".")
	current := interface{}(doc)

	for i, part := range parts {
		// Check if current is a map
		currentMap, ok := current.(bson.M)
		if !ok {
			return nil, false
		}

		// Get the next level value
		value, exists := currentMap[part]
		if !exists {
			return nil, false
		}

		// If this is the last part, return the value
		if i == len(parts)-1 {
			return value, true
		}

		// Move to next level
		current = value
	}

	return nil, false
}

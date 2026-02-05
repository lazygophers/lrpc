package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

// TestMatchFilter tests the basic matchFilter function
func TestMatchFilter(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "empty filter matches all",
			doc:      bson.M{"name": "Alice", "age": 25},
			filter:   bson.M{},
			expected: true,
		},
		{
			name:     "simple equality match",
			doc:      bson.M{"name": "Alice", "age": 25},
			filter:   bson.M{"name": "Alice"},
			expected: true,
		},
		{
			name:     "simple equality no match",
			doc:      bson.M{"name": "Alice", "age": 25},
			filter:   bson.M{"name": "Bob"},
			expected: false,
		},
		{
			name:     "multiple fields match",
			doc:      bson.M{"name": "Alice", "age": 25},
			filter:   bson.M{"name": "Alice", "age": 25},
			expected: true,
		},
		{
			name:     "field not exists",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": 25},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchFieldOperators_Equality tests $eq and $ne operators
func TestMatchFieldOperators_Equality(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "$eq operator match",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$eq": 25}},
			expected: true,
		},
		{
			name:     "$eq operator no match",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$eq": 30}},
			expected: false,
		},
		{
			name:     "$ne operator match",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$ne": 30}},
			expected: true,
		},
		{
			name:     "$ne operator no match",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$ne": 25}},
			expected: false,
		},
		{
			name:     "$ne on non-existent field",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$ne": 25}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchFieldOperators_Comparison tests $gt, $gte, $lt, $lte operators
func TestMatchFieldOperators_Comparison(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "$gt integer match",
			doc:      bson.M{"age": 30},
			filter:   bson.M{"age": bson.M{"$gt": 25}},
			expected: true,
		},
		{
			name:     "$gt integer no match",
			doc:      bson.M{"age": 20},
			filter:   bson.M{"age": bson.M{"$gt": 25}},
			expected: false,
		},
		{
			name:     "$gte equal match",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$gte": 25}},
			expected: true,
		},
		{
			name:     "$gte greater match",
			doc:      bson.M{"age": 30},
			filter:   bson.M{"age": bson.M{"$gte": 25}},
			expected: true,
		},
		{
			name:     "$lt integer match",
			doc:      bson.M{"age": 20},
			filter:   bson.M{"age": bson.M{"$lt": 25}},
			expected: true,
		},
		{
			name:     "$lt integer no match",
			doc:      bson.M{"age": 30},
			filter:   bson.M{"age": bson.M{"$lt": 25}},
			expected: false,
		},
		{
			name:     "$lte equal match",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$lte": 25}},
			expected: true,
		},
		{
			name:     "$lte less match",
			doc:      bson.M{"age": 20},
			filter:   bson.M{"age": bson.M{"$lte": 25}},
			expected: true,
		},
		{
			name:     "comparison on float",
			doc:      bson.M{"price": 19.99},
			filter:   bson.M{"price": bson.M{"$gt": 10.0}},
			expected: true,
		},
		{
			name:     "comparison on string",
			doc:      bson.M{"name": "Bob"},
			filter:   bson.M{"name": bson.M{"$gt": "Alice"}},
			expected: true,
		},
		{
			name:     "comparison on non-existent field",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$gt": 25}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchFieldOperators_InAndNin tests $in and $nin operators
func TestMatchFieldOperators_InAndNin(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "$in operator match",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$in": []interface{}{20, 25, 30}}},
			expected: true,
		},
		{
			name:     "$in operator no match",
			doc:      bson.M{"age": 35},
			filter:   bson.M{"age": bson.M{"$in": []interface{}{20, 25, 30}}},
			expected: false,
		},
		{
			name:     "$in with string",
			doc:      bson.M{"name": "Bob"},
			filter:   bson.M{"name": bson.M{"$in": []interface{}{"Alice", "Bob", "Charlie"}}},
			expected: true,
		},
		{
			name:     "$nin operator match",
			doc:      bson.M{"age": 35},
			filter:   bson.M{"age": bson.M{"$nin": []interface{}{20, 25, 30}}},
			expected: true,
		},
		{
			name:     "$nin operator no match",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$nin": []interface{}{20, 25, 30}}},
			expected: false,
		},
		{
			name:     "$in on non-existent field",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$in": []interface{}{20, 25, 30}}},
			expected: false,
		},
		{
			name:     "$nin on non-existent field",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$nin": []interface{}{20, 25, 30}}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchFieldOperators_Exists tests $exists operator
func TestMatchFieldOperators_Exists(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "$exists true - field exists",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$exists": true}},
			expected: true,
		},
		{
			name:     "$exists true - field not exists",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$exists": true}},
			expected: false,
		},
		{
			name:     "$exists false - field exists",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$exists": false}},
			expected: false,
		},
		{
			name:     "$exists false - field not exists",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$exists": false}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchLogicalOperator_And tests $and operator
func TestMatchLogicalOperator_And(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name: "$and all conditions match",
			doc:  bson.M{"name": "Alice", "age": 25},
			filter: bson.M{
				"$and": []interface{}{
					bson.M{"name": "Alice"},
					bson.M{"age": 25},
				},
			},
			expected: true,
		},
		{
			name: "$and one condition fails",
			doc:  bson.M{"name": "Alice", "age": 25},
			filter: bson.M{
				"$and": []interface{}{
					bson.M{"name": "Alice"},
					bson.M{"age": 30},
				},
			},
			expected: false,
		},
		{
			name: "$and with comparison operators",
			doc:  bson.M{"age": 25},
			filter: bson.M{
				"$and": []interface{}{
					bson.M{"age": bson.M{"$gte": 20}},
					bson.M{"age": bson.M{"$lte": 30}},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchLogicalOperator_Or tests $or operator
func TestMatchLogicalOperator_Or(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name: "$or one condition matches",
			doc:  bson.M{"name": "Alice", "age": 25},
			filter: bson.M{
				"$or": []interface{}{
					bson.M{"name": "Bob"},
					bson.M{"age": 25},
				},
			},
			expected: true,
		},
		{
			name: "$or all conditions match",
			doc:  bson.M{"name": "Alice", "age": 25},
			filter: bson.M{
				"$or": []interface{}{
					bson.M{"name": "Alice"},
					bson.M{"age": 25},
				},
			},
			expected: true,
		},
		{
			name: "$or no conditions match",
			doc:  bson.M{"name": "Alice", "age": 25},
			filter: bson.M{
				"$or": []interface{}{
					bson.M{"name": "Bob"},
					bson.M{"age": 30},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchLogicalOperator_Nor tests $nor operator
func TestMatchLogicalOperator_Nor(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name: "$nor no conditions match",
			doc:  bson.M{"name": "Alice", "age": 25},
			filter: bson.M{
				"$nor": []interface{}{
					bson.M{"name": "Bob"},
					bson.M{"age": 30},
				},
			},
			expected: true,
		},
		{
			name: "$nor one condition matches",
			doc:  bson.M{"name": "Alice", "age": 25},
			filter: bson.M{
				"$nor": []interface{}{
					bson.M{"name": "Alice"},
					bson.M{"age": 30},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchLogicalOperator_Not tests $not operator
func TestMatchLogicalOperator_Not(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name: "$not condition matches",
			doc:  bson.M{"age": 25},
			filter: bson.M{
				"$not": bson.M{"age": 30},
			},
			expected: true,
		},
		{
			name: "$not condition fails",
			doc:  bson.M{"age": 25},
			filter: bson.M{
				"$not": bson.M{"age": 25},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetNestedValue tests nested field access with dot notation
func TestGetNestedValue(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name        string
		doc         bson.M
		field       string
		expectValue interface{}
		expectFound bool
	}{
		{
			name:        "simple field",
			doc:         bson.M{"name": "Alice"},
			field:       "name",
			expectValue: "Alice",
			expectFound: true,
		},
		{
			name: "nested field",
			doc: bson.M{
				"user": bson.M{
					"name": "Alice",
				},
			},
			field:       "user.name",
			expectValue: "Alice",
			expectFound: true,
		},
		{
			name: "deep nested field",
			doc: bson.M{
				"user": bson.M{
					"profile": bson.M{
						"name": "Alice",
					},
				},
			},
			field:       "user.profile.name",
			expectValue: "Alice",
			expectFound: true,
		},
		{
			name:        "non-existent field",
			doc:         bson.M{"name": "Alice"},
			field:       "age",
			expectValue: nil,
			expectFound: false,
		},
		{
			name: "non-existent nested field",
			doc: bson.M{
				"user": bson.M{
					"name": "Alice",
				},
			},
			field:       "user.age",
			expectValue: nil,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found := storage.getNestedValue(tt.doc, tt.field)
			assert.Equal(t, tt.expectFound, found)
			if tt.expectFound {
				assert.Equal(t, tt.expectValue, value)
			}
		})
	}
}

// TestMatchNestedFilter tests filter with nested fields
func TestMatchNestedFilter(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name: "nested field match",
			doc: bson.M{
				"user": bson.M{
					"name": "Alice",
					"age":  25,
				},
			},
			filter:   bson.M{"user.name": "Alice"},
			expected: true,
		},
		{
			name: "nested field no match",
			doc: bson.M{
				"user": bson.M{
					"name": "Alice",
				},
			},
			filter:   bson.M{"user.name": "Bob"},
			expected: false,
		},
		{
			name: "nested field with operator",
			doc: bson.M{
				"user": bson.M{
					"age": 25,
				},
			},
			filter:   bson.M{"user.age": bson.M{"$gt": 20}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCompareValues tests comparison helper functions
func TestCompareValues(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		op       string
		expected bool
	}{
		{
			name:     "int comparison $gt",
			a:        30,
			b:        20,
			op:       "$gt",
			expected: true,
		},
		{
			name:     "int32 comparison $gte",
			a:        int32(25),
			b:        int32(25),
			op:       "$gte",
			expected: true,
		},
		{
			name:     "int64 comparison $lt",
			a:        int64(20),
			b:        int64(30),
			op:       "$lt",
			expected: true,
		},
		{
			name:     "float32 comparison $lte",
			a:        float32(19.99),
			b:        float32(20.0),
			op:       "$lte",
			expected: true,
		},
		{
			name:     "float64 comparison $gt",
			a:        29.99,
			b:        19.99,
			op:       "$gt",
			expected: true,
		},
		{
			name:     "string comparison $gt",
			a:        "Bob",
			b:        "Alice",
			op:       "$gt",
			expected: true,
		},
		{
			name:     "string comparison $lt",
			a:        "Alice",
			b:        "Bob",
			op:       "$lt",
			expected: true,
		},
		{
			name:     "nil comparison",
			a:        nil,
			b:        20,
			op:       "$gt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.compareValues(tt.a, tt.b, tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchValue tests value equality matching
func TestMatchValue(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name      string
		docVal    interface{}
		filterVal interface{}
		expected  bool
	}{
		{
			name:      "equal strings",
			docVal:    "Alice",
			filterVal: "Alice",
			expected:  true,
		},
		{
			name:      "not equal strings",
			docVal:    "Alice",
			filterVal: "Bob",
			expected:  false,
		},
		{
			name:      "equal integers",
			docVal:    25,
			filterVal: 25,
			expected:  true,
		},
		{
			name:      "both nil",
			docVal:    nil,
			filterVal: nil,
			expected:  true,
		},
		{
			name:      "one nil",
			docVal:    nil,
			filterVal: 25,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchValue(tt.docVal, tt.filterVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchInOperator tests $in operator matching
func TestMatchInOperator(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		value    interface{}
		arrayVal interface{}
		expected bool
	}{
		{
			name:     "value in integer array",
			value:    25,
			arrayVal: []interface{}{20, 25, 30},
			expected: true,
		},
		{
			name:     "value not in integer array",
			value:    35,
			arrayVal: []interface{}{20, 25, 30},
			expected: false,
		},
		{
			name:     "value in string array",
			value:    "Bob",
			arrayVal: []interface{}{"Alice", "Bob", "Charlie"},
			expected: true,
		},
		{
			name:     "empty array",
			value:    25,
			arrayVal: []interface{}{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchInOperator(tt.value, tt.arrayVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestComplexQueryScenarios tests complex real-world query scenarios
func TestComplexQueryScenarios(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name: "complex filter with $and and comparison",
			doc:  bson.M{"age": 25, "score": 85},
			filter: bson.M{
				"$and": []interface{}{
					bson.M{"age": bson.M{"$gte": 18}},
					bson.M{"score": bson.M{"$gt": 80}},
				},
			},
			expected: true,
		},
		{
			name: "complex filter with $or and $in",
			doc:  bson.M{"status": "active", "role": "admin"},
			filter: bson.M{
				"$or": []interface{}{
					bson.M{"status": bson.M{"$in": []interface{}{"active", "pending"}}},
					bson.M{"role": "admin"},
				},
			},
			expected: true,
		},
		{
			name: "nested fields with operators",
			doc: bson.M{
				"user": bson.M{
					"profile": bson.M{
						"age": 25,
					},
				},
			},
			filter: bson.M{
				"user.profile.age": bson.M{"$gte": 18},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestLogicalOperator_ErrorCases tests error handling in logical operators
func TestLogicalOperator_ErrorCases(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "$and with non-array value",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"$and": "invalid"},
			expected: false,
		},
		{
			name:     "$and with non-bson.M condition",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"$and": []interface{}{"invalid"}},
			expected: false,
		},
		{
			name:     "$or with non-array value",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"$or": "invalid"},
			expected: false,
		},
		{
			name:     "$or with non-bson.M condition",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"$or": []interface{}{"invalid"}},
			expected: false,
		},
		{
			name:     "$nor with non-array value",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"$nor": "invalid"},
			expected: false,
		},
		{
			name:     "$nor with non-bson.M condition",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"$nor": []interface{}{"invalid"}},
			expected: false,
		},
		{
			name:     "$not with non-bson.M value",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"$not": "invalid"},
			expected: false,
		},
		{
			name:     "unsupported logical operator",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"$unknown": bson.M{"age": 25}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFieldOperators_ErrorCases tests error handling in field operators
func TestFieldOperators_ErrorCases(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "$eq on non-existent field",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$eq": 25}},
			expected: false,
		},
		{
			name:     "$exists with non-boolean value",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$exists": "invalid"}},
			expected: false,
		},
		{
			name:     "unsupported field operator",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$unknown": 25}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCompareValues_EdgeCases tests edge cases in value comparison
func TestCompareValues_EdgeCases(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		op       string
		expected bool
	}{
		{
			name:     "nil value in a",
			a:        nil,
			b:        20,
			op:       "$gt",
			expected: false,
		},
		{
			name:     "nil value in b",
			a:        20,
			b:        nil,
			op:       "$gt",
			expected: false,
		},
		{
			name:     "unsupported type",
			a:        true,
			b:        false,
			op:       "$gt",
			expected: false,
		},
		{
			name:     "int32 with mismatched type",
			a:        int32(25),
			b:        "invalid",
			op:       "$gt",
			expected: false,
		},
		{
			name:     "int64 with mismatched type",
			a:        int64(25),
			b:        "invalid",
			op:       "$gt",
			expected: false,
		},
		{
			name:     "float32 with mismatched type",
			a:        float32(25.5),
			b:        "invalid",
			op:       "$gt",
			expected: false,
		},
		{
			name:     "float64 with mismatched type",
			a:        float64(25.5),
			b:        "invalid",
			op:       "$gt",
			expected: false,
		},
		{
			name:     "string with mismatched type",
			a:        "Alice",
			b:        123,
			op:       "$gt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.compareValues(tt.a, tt.b, tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCompareInts_AllOperators tests all comparison operators for integers
func TestCompareInts_AllOperators(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		a        int64
		b        int64
		op       string
		expected bool
	}{
		{
			name:     "$gt true",
			a:        30,
			b:        20,
			op:       "$gt",
			expected: true,
		},
		{
			name:     "$gt false",
			a:        20,
			b:        30,
			op:       "$gt",
			expected: false,
		},
		{
			name:     "$gte equal",
			a:        25,
			b:        25,
			op:       "$gte",
			expected: true,
		},
		{
			name:     "$gte greater",
			a:        30,
			b:        25,
			op:       "$gte",
			expected: true,
		},
		{
			name:     "$gte less",
			a:        20,
			b:        25,
			op:       "$gte",
			expected: false,
		},
		{
			name:     "$lt true",
			a:        20,
			b:        30,
			op:       "$lt",
			expected: true,
		},
		{
			name:     "$lt false",
			a:        30,
			b:        20,
			op:       "$lt",
			expected: false,
		},
		{
			name:     "$lte equal",
			a:        25,
			b:        25,
			op:       "$lte",
			expected: true,
		},
		{
			name:     "$lte less",
			a:        20,
			b:        25,
			op:       "$lte",
			expected: true,
		},
		{
			name:     "$lte greater",
			a:        30,
			b:        25,
			op:       "$lte",
			expected: false,
		},
		{
			name:     "unknown operator",
			a:        25,
			b:        25,
			op:       "$unknown",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.compareInts(tt.a, tt.b, tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCompareFloats_AllOperators tests all comparison operators for floats
func TestCompareFloats_AllOperators(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		a        float64
		b        float64
		op       string
		expected bool
	}{
		{
			name:     "$gt true",
			a:        30.5,
			b:        20.5,
			op:       "$gt",
			expected: true,
		},
		{
			name:     "$gt false",
			a:        20.5,
			b:        30.5,
			op:       "$gt",
			expected: false,
		},
		{
			name:     "$gte equal",
			a:        25.5,
			b:        25.5,
			op:       "$gte",
			expected: true,
		},
		{
			name:     "$gte greater",
			a:        30.5,
			b:        25.5,
			op:       "$gte",
			expected: true,
		},
		{
			name:     "$gte less",
			a:        20.5,
			b:        25.5,
			op:       "$gte",
			expected: false,
		},
		{
			name:     "$lt true",
			a:        20.5,
			b:        30.5,
			op:       "$lt",
			expected: true,
		},
		{
			name:     "$lt false",
			a:        30.5,
			b:        20.5,
			op:       "$lt",
			expected: false,
		},
		{
			name:     "$lte equal",
			a:        25.5,
			b:        25.5,
			op:       "$lte",
			expected: true,
		},
		{
			name:     "$lte less",
			a:        20.5,
			b:        25.5,
			op:       "$lte",
			expected: true,
		},
		{
			name:     "$lte greater",
			a:        30.5,
			b:        25.5,
			op:       "$lte",
			expected: false,
		},
		{
			name:     "unknown operator",
			a:        25.5,
			b:        25.5,
			op:       "$unknown",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.compareFloats(tt.a, tt.b, tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCompareStrings_AllOperators tests all comparison operators for strings
func TestCompareStrings_AllOperators(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		a        string
		b        string
		op       string
		expected bool
	}{
		{
			name:     "$gt true",
			a:        "Charlie",
			b:        "Bob",
			op:       "$gt",
			expected: true,
		},
		{
			name:     "$gt false",
			a:        "Alice",
			b:        "Bob",
			op:       "$gt",
			expected: false,
		},
		{
			name:     "$gte equal",
			a:        "Bob",
			b:        "Bob",
			op:       "$gte",
			expected: true,
		},
		{
			name:     "$gte greater",
			a:        "Charlie",
			b:        "Bob",
			op:       "$gte",
			expected: true,
		},
		{
			name:     "$gte less",
			a:        "Alice",
			b:        "Bob",
			op:       "$gte",
			expected: false,
		},
		{
			name:     "$lt true",
			a:        "Alice",
			b:        "Bob",
			op:       "$lt",
			expected: true,
		},
		{
			name:     "$lt false",
			a:        "Charlie",
			b:        "Bob",
			op:       "$lt",
			expected: false,
		},
		{
			name:     "$lte equal",
			a:        "Bob",
			b:        "Bob",
			op:       "$lte",
			expected: true,
		},
		{
			name:     "$lte less",
			a:        "Alice",
			b:        "Bob",
			op:       "$lte",
			expected: true,
		},
		{
			name:     "$lte greater",
			a:        "Charlie",
			b:        "Bob",
			op:       "$lte",
			expected: false,
		},
		{
			name:     "unknown operator",
			a:        "Bob",
			b:        "Bob",
			op:       "$unknown",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.compareStrings(tt.a, tt.b, tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestToInt_AllTypes tests type conversion to int64
func TestToInt_AllTypes(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		value    interface{}
		expected int64
		ok       bool
	}{
		{
			name:     "int to int64",
			value:    int(25),
			expected: 25,
			ok:       true,
		},
		{
			name:     "int32 to int64",
			value:    int32(25),
			expected: 25,
			ok:       true,
		},
		{
			name:     "int64 to int64",
			value:    int64(25),
			expected: 25,
			ok:       true,
		},
		{
			name:     "float32 to int64",
			value:    float32(25.5),
			expected: 25,
			ok:       true,
		},
		{
			name:     "float64 to int64",
			value:    float64(25.5),
			expected: 25,
			ok:       true,
		},
		{
			name:     "unsupported type",
			value:    "invalid",
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := storage.toInt(tt.value)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestToFloat_AllTypes tests type conversion to float64
func TestToFloat_AllTypes(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		value    interface{}
		expected float64
		ok       bool
	}{
		{
			name:     "int to float64",
			value:    int(25),
			expected: 25.0,
			ok:       true,
		},
		{
			name:     "int32 to float64",
			value:    int32(25),
			expected: 25.0,
			ok:       true,
		},
		{
			name:     "int64 to float64",
			value:    int64(25),
			expected: 25.0,
			ok:       true,
		},
		{
			name:     "float32 to float64",
			value:    float32(25.5),
			expected: float64(float32(25.5)),
			ok:       true,
		},
		{
			name:     "float64 to float64",
			value:    float64(25.5),
			expected: 25.5,
			ok:       true,
		},
		{
			name:     "unsupported type",
			value:    "invalid",
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := storage.toFloat(tt.value)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestMatchInOperator_ErrorCases tests error handling in $in operator
func TestMatchInOperator_ErrorCases(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		value    interface{}
		arrayVal interface{}
		expected bool
	}{
		{
			name:     "non-array value",
			value:    25,
			arrayVal: "invalid",
			expected: false,
		},
		{
			name:     "non-slice non-array value",
			value:    25,
			arrayVal: 123,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchInOperator(tt.value, tt.arrayVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetNestedValue_EdgeCases tests edge cases in nested value retrieval
func TestGetNestedValue_EdgeCases(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name        string
		doc         bson.M
		field       string
		expectValue interface{}
		expectFound bool
	}{
		{
			name: "nested field with non-map intermediate value",
			doc: bson.M{
				"user": "not a map",
			},
			field:       "user.name",
			expectValue: nil,
			expectFound: false,
		},
		{
			name: "deeply nested field with missing intermediate",
			doc: bson.M{
				"user": bson.M{
					"profile": bson.M{},
				},
			},
			field:       "user.profile.name.first",
			expectValue: nil,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found := storage.getNestedValue(tt.doc, tt.field)
			assert.Equal(t, tt.expectFound, found)
			if tt.expectFound {
				assert.Equal(t, tt.expectValue, value)
			}
		})
	}
}

// TestMatchValue_EdgeCases tests edge cases in value matching
func TestMatchValue_EdgeCases(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name      string
		docVal    interface{}
		filterVal interface{}
		expected  bool
	}{
		{
			name:      "doc value nil, filter not nil",
			docVal:    nil,
			filterVal: 25,
			expected:  false,
		},
		{
			name:      "doc value not nil, filter nil",
			docVal:    25,
			filterVal: nil,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchValue(tt.docVal, tt.filterVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCompareValues_AllIntTypes tests all integer type comparisons
func TestCompareValues_AllIntTypes(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		op       string
		expected bool
	}{
		{
			name:     "int with int comparison",
			a:        int(30),
			b:        int(20),
			op:       "$gt",
			expected: true,
		},
		{
			name:     "int32 with int32 comparison",
			a:        int32(30),
			b:        int32(20),
			op:       "$gte",
			expected: true,
		},
		{
			name:     "int64 with int64 comparison",
			a:        int64(20),
			b:        int64(30),
			op:       "$lt",
			expected: true,
		},
		{
			name:     "float32 with float32 comparison",
			a:        float32(20.5),
			b:        float32(30.5),
			op:       "$lte",
			expected: true,
		},
		{
			name:     "float64 with float64 comparison",
			a:        float64(30.5),
			b:        float64(20.5),
			op:       "$gt",
			expected: true,
		},
		{
			name:     "string with string comparison",
			a:        "Bob",
			b:        "Alice",
			op:       "$gt",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.compareValues(tt.a, tt.b, tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetNestedValue_FullCoverage tests all edge cases for nested value retrieval
func TestGetNestedValue_FullCoverage(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name        string
		doc         bson.M
		field       string
		expectValue interface{}
		expectFound bool
	}{
		{
			name:        "simple field exists",
			doc:         bson.M{"name": "Alice"},
			field:       "name",
			expectValue: "Alice",
			expectFound: true,
		},
		{
			name:        "simple field not exists",
			doc:         bson.M{"name": "Alice"},
			field:       "age",
			expectValue: nil,
			expectFound: false,
		},
		{
			name: "nested field exists",
			doc: bson.M{
				"user": bson.M{
					"name": "Alice",
				},
			},
			field:       "user.name",
			expectValue: "Alice",
			expectFound: true,
		},
		{
			name: "nested field with intermediate non-map",
			doc: bson.M{
				"user": "not a map",
			},
			field:       "user.name",
			expectValue: nil,
			expectFound: false,
		},
		{
			name: "nested field with missing intermediate field",
			doc: bson.M{
				"user": bson.M{
					"email": "alice@example.com",
				},
			},
			field:       "user.profile.name",
			expectValue: nil,
			expectFound: false,
		},
		{
			name: "deeply nested field exists",
			doc: bson.M{
				"user": bson.M{
					"profile": bson.M{
						"details": bson.M{
							"name": "Alice",
						},
					},
				},
			},
			field:       "user.profile.details.name",
			expectValue: "Alice",
			expectFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found := storage.getNestedValue(tt.doc, tt.field)
			assert.Equal(t, tt.expectFound, found)
			if tt.expectFound {
				assert.Equal(t, tt.expectValue, value)
			}
		})
	}
}

// TestMatchFieldOperators_AllBranches tests all branches in field operators
func TestMatchFieldOperators_AllBranches(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "$eq with nil field",
			doc:      bson.M{"name": nil},
			filter:   bson.M{"name": bson.M{"$eq": nil}},
			expected: true,
		},
		{
			name:     "$gt on string comparison",
			doc:      bson.M{"name": "Charlie"},
			filter:   bson.M{"name": bson.M{"$gt": "Bob"}},
			expected: true,
		},
		{
			name:     "$gte on string comparison",
			doc:      bson.M{"name": "Bob"},
			filter:   bson.M{"name": bson.M{"$gte": "Bob"}},
			expected: true,
		},
		{
			name:     "$lt on string comparison",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"name": bson.M{"$lt": "Bob"}},
			expected: true,
		},
		{
			name:     "$lte on string comparison",
			doc:      bson.M{"name": "Bob"},
			filter:   bson.M{"name": bson.M{"$lte": "Bob"}},
			expected: true,
		},
		{
			name:     "$gt with nil value in a",
			doc:      bson.M{"age": nil},
			filter:   bson.M{"age": bson.M{"$gt": 20}},
			expected: false,
		},
		{
			name:     "$gte with nil value in a",
			doc:      bson.M{"age": nil},
			filter:   bson.M{"age": bson.M{"$gte": 20}},
			expected: false,
		},
		{
			name:     "$lt with nil value in a",
			doc:      bson.M{"age": nil},
			filter:   bson.M{"age": bson.M{"$lt": 20}},
			expected: false,
		},
		{
			name:     "$lte with nil value in a",
			doc:      bson.M{"age": nil},
			filter:   bson.M{"age": bson.M{"$lte": 20}},
			expected: false,
		},
		{
			name:     "$gt with unsupported type bool",
			doc:      bson.M{"flag": true},
			filter:   bson.M{"flag": bson.M{"$gt": false}},
			expected: false,
		},
		{
			name:     "$gte with unsupported type bool",
			doc:      bson.M{"flag": true},
			filter:   bson.M{"flag": bson.M{"$gte": false}},
			expected: false,
		},
		{
			name:     "$lt with unsupported type bool",
			doc:      bson.M{"flag": true},
			filter:   bson.M{"flag": bson.M{"$lt": false}},
			expected: false,
		},
		{
			name:     "$lte with unsupported type bool",
			doc:      bson.M{"flag": true},
			filter:   bson.M{"flag": bson.M{"$lte": false}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCompareValues_CompleteTypes tests compareValues with all type variations
func TestCompareValues_CompleteTypes(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		op       string
		expected bool
	}{
		{
			name:     "int with int $lte",
			a:        int(25),
			b:        int(30),
			op:       "$lte",
			expected: true,
		},
		{
			name:     "int32 with int32 $lt",
			a:        int32(20),
			b:        int32(30),
			op:       "$lt",
			expected: true,
		},
		{
			name:     "int64 with int64 $gte",
			a:        int64(30),
			b:        int64(25),
			op:       "$gte",
			expected: true,
		},
		{
			name:     "float32 with float32 $gt",
			a:        float32(30.5),
			b:        float32(20.5),
			op:       "$gt",
			expected: true,
		},
		{
			name:     "float64 with float64 $lte",
			a:        float64(20.5),
			b:        float64(30.5),
			op:       "$lte",
			expected: true,
		},
		{
			name:     "string with string $lte",
			a:        "Alice",
			b:        "Bob",
			op:       "$lte",
			expected: true,
		},
		{
			name:     "string with string $gte",
			a:        "Charlie",
			b:        "Bob",
			op:       "$gte",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.compareValues(tt.a, tt.b, tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchFieldOperators_ComparisonReturnsTrue tests all comparison operators returning true
func TestMatchFieldOperators_ComparisonReturnsTrue(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "$gt returns true",
			doc:      bson.M{"age": 30},
			filter:   bson.M{"age": bson.M{"$gt": 20}},
			expected: true,
		},
		{
			name:     "$gte returns true",
			doc:      bson.M{"age": 30},
			filter:   bson.M{"age": bson.M{"$gte": 30}},
			expected: true,
		},
		{
			name:     "$lt returns true",
			doc:      bson.M{"age": 20},
			filter:   bson.M{"age": bson.M{"$lt": 30}},
			expected: true,
		},
		{
			name:     "$lte returns true",
			doc:      bson.M{"age": 30},
			filter:   bson.M{"age": bson.M{"$lte": 30}},
			expected: true,
		},
		{
			name:     "$in returns true",
			doc:      bson.M{"status": "active"},
			filter:   bson.M{"status": bson.M{"$in": []interface{}{"active", "pending"}}},
			expected: true,
		},
		{
			name:     "$nin returns true",
			doc:      bson.M{"status": "archived"},
			filter:   bson.M{"status": bson.M{"$nin": []interface{}{"active", "pending"}}},
			expected: true,
		},
		{
			name:     "$exists true returns true",
			doc:      bson.M{"age": 25},
			filter:   bson.M{"age": bson.M{"$exists": true}},
			expected: true,
		},
		{
			name:     "$exists false returns true",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$exists": false}},
			expected: true,
		},
		{
			name:     "$eq returns true",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"name": bson.M{"$eq": "Alice"}},
			expected: true,
		},
		{
			name:     "$ne returns true",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"name": bson.M{"$ne": "Bob"}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchFieldOperators_MultipleOperators tests combining multiple operators
func TestMatchFieldOperators_MultipleOperators(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name: "combined $gte and $lte",
			doc:  bson.M{"age": 25},
			filter: bson.M{"age": bson.M{
				"$gte": 20,
				"$lte": 30,
			}},
			expected: true,
		},
		{
			name: "combined $gt and $lt both true",
			doc:  bson.M{"age": 25},
			filter: bson.M{"age": bson.M{
				"$gt": 20,
				"$lt": 30,
			}},
			expected: true,
		},
		{
			name: "combined $ne and $exists",
			doc:  bson.M{"name": "Alice"},
			filter: bson.M{"name": bson.M{
				"$ne":     "Bob",
				"$exists": true,
			}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchFieldOperators_NonExistentFields tests operators on non-existent fields
func TestMatchFieldOperators_NonExistentFields(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		doc      bson.M
		filter   bson.M
		expected bool
	}{
		{
			name:     "$gte on non-existent field",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$gte": 18}},
			expected: false,
		},
		{
			name:     "$lt on non-existent field",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$lt": 100}},
			expected: false,
		},
		{
			name:     "$lte on non-existent field",
			doc:      bson.M{"name": "Alice"},
			filter:   bson.M{"age": bson.M{"$lte": 100}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.matchFilter(tt.doc, tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCompareValues_TypeMismatch tests comparison with type mismatch for int
func TestCompareValues_TypeMismatch_Int(t *testing.T) {
	storage := NewMemoryStorage()

	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		op       string
		expected bool
	}{
		{
			name:     "int with string",
			a:        int(25),
			b:        "invalid",
			op:       "$gt",
			expected: false,
		},
		{
			name:     "int with bool",
			a:        int(25),
			b:        true,
			op:       "$gte",
			expected: false,
		},
		{
			name:     "int with nil",
			a:        int(25),
			b:        nil,
			op:       "$lt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.compareValues(tt.a, tt.b, tt.op)
			assert.Equal(t, tt.expected, result)
		})
	}
}

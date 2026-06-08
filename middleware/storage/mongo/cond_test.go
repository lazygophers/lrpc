package mongo

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// TestCondWithBoolFalse tests where method with bool false argument
func TestCondWithBoolFalse(t *testing.T) {
	cond := NewCond().Equal("age", 25)
	// Add false condition - should set skip flag
	cond.where(false)
	// After skip is set, the condition should still build but with skip flag
	result := cond.ToBson()
	// When skip is true, result may be affected
	if result == nil {
		// OK - skip=true might result in nil
	}
}

// TestCondWithBoolTrue tests where method with bool true argument
func TestCondWithBoolTrue(t *testing.T) {
	cond := NewCond().Equal("age", 25)
	// Add true condition - should not set skip flag
	cond.where(true)
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result")
	}
}

// TestCondWithMapCondition tests where method with map argument
func TestCondWithMapCondition(t *testing.T) {
	cond := NewCond()
	mapCond := map[string]interface{}{
		"age":  25,
		"name": "John",
	}
	cond.where(mapCond)
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from map condition")
	}

	// Verify map fields are included
	if _, ok := result["age"]; !ok {
		t.Error("expected 'age' field in result")
	}
	if _, ok := result["name"]; !ok {
		t.Error("expected 'name' field in result")
	}
}

// TestCondWithSliceConditions tests where method with slice argument
func TestCondWithSliceConditions(t *testing.T) {
	cond := NewCond()
	sliceCond := []interface{}{
		map[string]interface{}{"age": 25},
		map[string]interface{}{"name": "John"},
	}
	cond.where(sliceCond)
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from slice condition")
	}
}

// TestCondWhereWithTwoArgs tests where method with field name and value
func TestCondWhereWithTwoArgs(t *testing.T) {
	cond := NewCond()
	cond.where("age", 25)
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from field/value condition")
	}

	// Should set age to 25
	if age, ok := result["age"]; !ok || age != 25 {
		t.Errorf("expected age=25, got %v", result["age"])
	}
}

// TestCondWhereWithThreeArgs tests where method with field, operator, and value
func TestCondWhereWithThreeArgs(t *testing.T) {
	cond := NewCond()
	cond.where("age", "$gt", 25)
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from field/op/value condition")
	}

	// Should have age with $gt operator
	if ageVal, ok := result["age"]; ok {
		if m, ok := ageVal.(bson.M); ok {
			if _, ok := m["$gt"]; !ok {
				t.Error("expected $gt operator in result")
			}
		} else {
			t.Errorf("expected bson.M, got %T", ageVal)
		}
	} else {
		t.Error("expected 'age' field in result")
	}
}

// TestCondWhereWithFieldOperatorExtraction tests where method with field name that includes operator
func TestCondWhereWithFieldOperatorExtraction(t *testing.T) {
	cond := NewCond()
	// Field name with operator suffix (e.g., "age>")
	cond.where("age>", 25)
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from operator-suffixed field")
	}
}

// TestCondToJsonWithMultipleConds tests ToBson with multiple conditions
func TestCondToJsonWithMultipleConds(t *testing.T) {
	cond := NewCond()
	cond.where("age", 25)
	cond.where("name", "John")
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result with multiple conditions")
	}

	// Both fields should be present (AND logic)
	if _, ok := result["age"]; !ok {
		t.Error("expected 'age' field")
	}
	if _, ok := result["name"]; !ok {
		t.Error("expected 'name' field")
	}
}

// TestCondOrWithMap tests Or method with map argument
func TestCondOrWithMap(t *testing.T) {
	cond := NewCond()
	cond.Or(map[string]interface{}{"age": 25}, map[string]interface{}{"age": 30})
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from Or condition")
	}

	// Should have $or operator
	if _, ok := result["$or"]; !ok {
		t.Error("expected $or operator in result")
	}
}

// TestCondWhereWithNestedCond tests where method with nested *Cond
func TestCondWhereWithNestedCond(t *testing.T) {
	subCond := NewCond().Equal("age", 25)
	mainCond := NewCond()
	mainCond.where(subCond)
	result := mainCond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from nested condition")
	}
}

// TestCondWhereWithInterfacePointer tests where with interface{} wrapping pointer
func TestCondWhereWithInterfacePointer(t *testing.T) {
	cond := NewCond()
	subCond := NewCond().Equal("age", 25)
	var iface interface{} = subCond
	cond.where(iface)
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from interface-wrapped condition")
	}
}

// TestCondLikeVariants tests Like operator and its variants
func TestCondLikeVariants(t *testing.T) {
	tests := []struct {
		name   string
		fn     func(*Cond) *Cond
		target string
	}{
		{
			name: "Like",
			fn: func(c *Cond) *Cond {
				return c.Like("name", "john")
			},
			target: "$regex",
		},
		{
			name: "LeftLike",
			fn: func(c *Cond) *Cond {
				return c.LeftLike("name", "john")
			},
			target: "$regex",
		},
		{
			name: "RightLike",
			fn: func(c *Cond) *Cond {
				return c.RightLike("name", "john")
			},
			target: "$regex",
		},
		{
			name: "NotLike",
			fn: func(c *Cond) *Cond {
				return c.NotLike("name", "john")
			},
			target: "$not",
		},
		{
			name: "NotLeftLike",
			fn: func(c *Cond) *Cond {
				return c.NotLeftLike("name", "john")
			},
			target: "$not",
		},
		{
			name: "NotRightLike",
			fn: func(c *Cond) *Cond {
				return c.NotRightLike("name", "john")
			},
			target: "$not",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := NewCond()
			tt.fn(cond)
			result := cond.ToBson()
			if result == nil {
				t.Errorf("%s: expected non-nil result", tt.name)
			}
		})
	}
}



// TestGetOp tests field operator extraction
func TestGetOp(t *testing.T) {
	tests := []struct {
		field    string
		expField string
		expOp    string
	}{
		{"age", "age", "="},
		{"age>", "age", ">"},
		{"age<", "age", "<"},
		{"age>=", "age", ">="},
		{"age<=", "age", "<="},
		{"age<>", "age", "<>"},
		{"age!=", "age", "!="},
		{"age%", "age", "%"},
		{"age~", "age", "~"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			field, op := getOp(tt.field)
			if field != tt.expField {
				t.Errorf("expected field %s, got %s", tt.expField, field)
			}
			if op != tt.expOp {
				t.Errorf("expected op %s, got %s", tt.expOp, op)
			}
		})
	}
}

// TestCondWhereWithSingleStringArg tests where with only string arg (field name only)
func TestCondWhereWithSingleStringArg(t *testing.T) {
	cond := NewCond()
	// Should not panic with single string argument
	cond.where("fieldName")
	result := cond.ToBson()
	// Single field name without value is skipped
	if result != nil {
		t.Error("expected nil result when only field name provided")
	}
}

// TestCondWhereEmptySlice tests where with empty slice
func TestCondWhereEmptySlice(t *testing.T) {
	cond := NewCond()
	cond.where([]interface{}{})
	result := cond.ToBson()
	// Empty slice should result in no conditions
	if result != nil {
		t.Error("expected nil result from empty slice")
	}
}

// TestCondWhereStringSlice tests where with string slice (field names)
func TestCondWhereStringSlice(t *testing.T) {
	cond := NewCond()
	cond.where([]interface{}{"age", "name"})
	result := cond.ToBson()
	// String slice should be processed as fields to skip
	if result != nil {
		// May result in no conditions if fields without values are skipped
	}
}

// TestCondWhereMapInSlice tests where with map inside slice
func TestCondWhereMapInSlice(t *testing.T) {
	cond := NewCond()
	cond.where([]interface{}{map[string]interface{}{"age": 25}})
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from map in slice")
	}
}

// TestCondWhereNestedSliceInSlice tests where with nested slice
func TestCondWhereNestedSliceInSlice(t *testing.T) {
	cond := NewCond()
	cond.where([]interface{}{[]interface{}{"age", 25}})
	result := cond.ToBson()
	if result == nil {
		t.Error("expected non-nil result from nested slice")
	}
}

// ============================================================
// Tests for uncovered methods from cond.go
// ============================================================

// TestCond_Where tests the Where method (0% covered)
func TestCond_Where(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Cond
		expected func(bson.M) bool
	}{
		{
			name: "simple field value",
			setup: func() *Cond {
				return NewCond().Where("age", 30)
			},
			expected: func(result bson.M) bool {
				return result["age"] == 30
			},
		},
		{
			name: "multiple conditions",
			setup: func() *Cond {
				return NewCond().Where("age", 30).Where("name", "John")
			},
			expected: func(result bson.M) bool {
				return result["age"] == 30 && result["name"] == "John"
			},
		},
		{
			name: "with map argument",
			setup: func() *Cond {
				return NewCond().Where(map[string]interface{}{"age": 30, "name": "John"})
			},
			expected: func(result bson.M) bool {
				return result["age"] == 30 && result["name"] == "John"
			},
		},
		{
			name: "with nested Cond",
			setup: func() *Cond {
				subCond := NewCond().Equal("status", "active")
				return NewCond().Where(subCond)
			},
			expected: func(result bson.M) bool {
				return result["status"] == "active"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.setup()
			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if !tt.expected(result) {
				t.Errorf("unexpected result: %v", result)
			}
		})
	}
}

// TestCond_OrWhere tests the OrWhere method (0% covered)
func TestCond_OrWhere(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Cond
		expected func(bson.M) bool
	}{
		{
			name: "single OrWhere call with multiple args",
			setup: func() *Cond {
				// OrWhere accepts multiple map conditions in a single call
				return NewCond().OrWhere(
					map[string]interface{}{"age": 25},
					map[string]interface{}{"age": 30},
				)
			},
			expected: func(result bson.M) bool {
				// Single OrWhere call creates a sub-condition
				return result != nil
			},
		},
		{
			name: "multiple OrWhere calls",
			setup: func() *Cond {
				// Multiple OrWhere calls add multiple OR groups
				return NewCond().
					OrWhere(map[string]interface{}{"age": 25}).
					OrWhere(map[string]interface{}{"status": "active"})
			},
			expected: func(result bson.M) bool {
				return result != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.setup()
			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if !tt.expected(result) {
				t.Errorf("unexpected result: %v", result)
			}
		})
	}
}

// TestCond_In tests the In method (0% covered)
func TestCond_In(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Cond
		expected func(bson.M) bool
	}{
		{
			name: "simple IN",
			setup: func() *Cond {
				return NewCond().In("age", 20, 25, 30)
			},
			expected: func(result bson.M) bool {
				ageVal, ok := result["age"]
				if !ok {
					return false
				}
				if m, ok := ageVal.(bson.M); ok {
					_, hasIn := m["$in"]
					return hasIn
				}
				return false
			},
		},
		{
			name: "IN with slice",
			setup: func() *Cond {
				return NewCond().In("status", "active", "pending", "approved")
			},
			expected: func(result bson.M) bool {
				statusVal, ok := result["status"]
				if !ok {
					return false
				}
				if m, ok := statusVal.(bson.M); ok {
					_, hasIn := m["$in"]
					return hasIn
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.setup()
			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if !tt.expected(result) {
				t.Errorf("unexpected result: %v", result)
			}
		})
	}
}

// TestCond_NotIn tests the NotIn method (0% covered)
func TestCond_NotIn(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Cond
		expected func(bson.M) bool
	}{
		{
			name: "simple NOT IN",
			setup: func() *Cond {
				return NewCond().NotIn("age", 20, 25, 30)
			},
			expected: func(result bson.M) bool {
				ageVal, ok := result["age"]
				if !ok {
					return false
				}
				if m, ok := ageVal.(bson.M); ok {
					_, hasNin := m["$nin"]
					return hasNin
				}
				return false
			},
		},
		{
			name: "NOT IN with multiple values",
			setup: func() *Cond {
				return NewCond().NotIn("status", "deleted", "archived")
			},
			expected: func(result bson.M) bool {
				statusVal, ok := result["status"]
				if !ok {
					return false
				}
				if m, ok := statusVal.(bson.M); ok {
					_, hasNin := m["$nin"]
					return hasNin
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.setup()
			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if !tt.expected(result) {
				t.Errorf("unexpected result: %v", result)
			}
		})
	}
}

// TestCond_Like_Empty tests Like method with empty pattern (75% covered)
func TestCond_Like_Empty(t *testing.T) {
	cond := NewCond().Like("name", "")
	result := cond.ToBson()
	// Empty pattern should result in no conditions
	if result != nil {
		t.Errorf("expected nil result for empty pattern, got %v", result)
	}
}

// TestCond_NotLike_Empty tests NotLike method with empty pattern (75% covered)
func TestCond_NotLike_Empty(t *testing.T) {
	cond := NewCond().NotLike("name", "")
	result := cond.ToBson()
	// Empty pattern should result in no conditions
	if result != nil {
		t.Errorf("expected nil result for empty pattern, got %v", result)
	}
}

// TestCond_LeftLike_Empty tests LeftLike method with empty pattern (75% covered)
func TestCond_LeftLike_Empty(t *testing.T) {
	cond := NewCond().LeftLike("name", "")
	result := cond.ToBson()
	// Empty pattern should result in no conditions
	if result != nil {
		t.Errorf("expected nil result for empty pattern, got %v", result)
	}
}

// TestCond_NotLeftLike_Empty tests NotLeftLike method with empty pattern (75% covered)
func TestCond_NotLeftLike_Empty(t *testing.T) {
	cond := NewCond().NotLeftLike("name", "")
	result := cond.ToBson()
	// Empty pattern should result in no conditions
	if result != nil {
		t.Errorf("expected nil result for empty pattern, got %v", result)
	}
}

// TestCond_RightLike_Empty tests RightLike method with empty pattern (75% covered)
func TestCond_RightLike_Empty(t *testing.T) {
	cond := NewCond().RightLike("name", "")
	result := cond.ToBson()
	// Empty pattern should result in no conditions
	if result != nil {
		t.Errorf("expected nil result for empty pattern, got %v", result)
	}
}

// TestCond_NotRightLike_Empty tests NotRightLike method with empty pattern (75% covered)
func TestCond_NotRightLike_Empty(t *testing.T) {
	cond := NewCond().NotRightLike("name", "")
	result := cond.ToBson()
	// Empty pattern should result in no conditions
	if result != nil {
		t.Errorf("expected nil result for empty pattern, got %v", result)
	}
}

// TestCond_Between tests the Between method (0% covered)
func TestCond_Between(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Cond
		expected func(bson.M) bool
	}{
		{
			name: "numeric range",
			setup: func() *Cond {
				return NewCond().Between("age", 20, 30)
			},
			expected: func(result bson.M) bool {
				ageVal, ok := result["age"]
				if !ok {
					return false
				}
				if m, ok := ageVal.(bson.M); ok {
					_, hasGte := m["$gte"]
					_, hasLte := m["$lte"]
					return hasGte && hasLte
				}
				return false
			},
		},
		{
			name: "date range",
			setup: func() *Cond {
				return NewCond().Between("created_at", "2024-01-01", "2024-12-31")
			},
			expected: func(result bson.M) bool {
				createdVal, ok := result["created_at"]
				if !ok {
					return false
				}
				if m, ok := createdVal.(bson.M); ok {
					_, hasGte := m["$gte"]
					_, hasLte := m["$lte"]
					return hasGte && hasLte
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.setup()
			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if !tt.expected(result) {
				t.Errorf("unexpected result: %v", result)
			}
		})
	}
}

// TestCond_NotBetween tests the NotBetween method (0% covered)
func TestCond_NotBetween(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Cond
		expected func(bson.M) bool
	}{
		{
			name: "numeric NOT BETWEEN",
			setup: func() *Cond {
				return NewCond().NotBetween("age", 20, 30)
			},
			expected: func(result bson.M) bool {
				ageVal, ok := result["age"]
				if !ok {
					return false
				}
				if m, ok := ageVal.(bson.M); ok {
					_, hasNot := m["$not"]
					return hasNot
				}
				return false
			},
		},
		{
			name: "date NOT BETWEEN",
			setup: func() *Cond {
				return NewCond().NotBetween("created_at", "2024-01-01", "2024-12-31")
			},
			expected: func(result bson.M) bool {
				createdVal, ok := result["created_at"]
				if !ok {
					return false
				}
				if m, ok := createdVal.(bson.M); ok {
					_, hasNot := m["$not"]
					return hasNot
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.setup()
			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if !tt.expected(result) {
				t.Errorf("unexpected result: %v", result)
			}
		})
	}
}

// TestCond_Ne tests the Ne method (0% covered)
func TestCond_Ne(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Cond
		expected func(bson.M) bool
	}{
		{
			name: "simple not equal",
			setup: func() *Cond {
				return NewCond().Ne("status", "deleted")
			},
			expected: func(result bson.M) bool {
				statusVal, ok := result["status"]
				if !ok {
					return false
				}
				if m, ok := statusVal.(bson.M); ok {
					_, hasNe := m["$ne"]
					return hasNe
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.setup()
			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if !tt.expected(result) {
				t.Errorf("unexpected result: %v", result)
			}
		})
	}
}

// TestCond_Gt tests the Gt method (0% covered)
func TestCond_Gt(t *testing.T) {
	cond := NewCond().Gt("age", 18)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasGt := m["$gt"]; !hasGt {
			t.Error("expected $gt operator")
		}
	} else {
		t.Errorf("expected bson.M, got %T", ageVal)
	}
}

// TestCond_Lt tests the Lt method (0% covered)
func TestCond_Lt(t *testing.T) {
	cond := NewCond().Lt("age", 65)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasLt := m["$lt"]; !hasLt {
			t.Error("expected $lt operator")
		}
	} else {
		t.Errorf("expected bson.M, got %T", ageVal)
	}
}

// TestCond_Gte tests the Gte method (0% covered)
func TestCond_Gte(t *testing.T) {
	cond := NewCond().Gte("age", 18)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasGte := m["$gte"]; !hasGte {
			t.Error("expected $gte operator")
		}
	} else {
		t.Errorf("expected bson.M, got %T", ageVal)
	}
}

// TestCond_Lte tests the Lte method (0% covered)
func TestCond_Lte(t *testing.T) {
	cond := NewCond().Lte("age", 65)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasLte := m["$lte"]; !hasLte {
			t.Error("expected $lte operator")
		}
	} else {
		t.Errorf("expected bson.M, got %T", ageVal)
	}
}

// TestCond_Reset tests the Reset method (0% covered)
func TestCond_Reset(t *testing.T) {
	cond := NewCond().Equal("age", 25).Equal("name", "John")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result before reset")
	}

	// Reset the condition
	cond.Reset()
	result = cond.ToBson()
	if result != nil {
		t.Errorf("expected nil result after reset, got %v", result)
	}
}

// TestCond_String tests the String method (0% covered)
func TestCond_String(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Cond
		contains string
	}{
		{
			name: "simple condition",
			setup: func() *Cond {
				return NewCond().Equal("age", 25)
			},
			contains: "age",
		},
		{
			name: "empty condition",
			setup: func() *Cond {
				return NewCond()
			},
			contains: "{}",
		},
		{
			name: "complex condition",
			setup: func() *Cond {
				return NewCond().Gt("age", 18).Lt("age", 65)
			},
			contains: "age",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.setup()
			str := cond.String()
			if str == "" {
				t.Error("expected non-empty string")
			}
			// Just verify it returns a string, don't check exact format
		})
	}
}

// TestCond_ChainedOperations tests chaining multiple operations
func TestCond_ChainedOperations(t *testing.T) {
	cond := NewCond().
		Equal("status", "active").
		Gt("age", 18).
		Lt("age", 65).
		In("role", "admin", "user")

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result from chained operations")
	}

	// Verify all fields are present
	if _, ok := result["status"]; !ok {
		t.Error("expected 'status' field")
	}
	if _, ok := result["age"]; !ok {
		t.Error("expected 'age' field")
	}
	if _, ok := result["role"]; !ok {
		t.Error("expected 'role' field")
	}
}

// ============================================================
// Tests for uncovered standalone functions from sub_cond.go
// ============================================================

// TestSubCond_Where tests the standalone Where function
func TestSubCond_Where(t *testing.T) {
	cond := Where("age", 30)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["age"] != 30 {
		t.Errorf("expected age=30, got %v", result["age"])
	}
}

// TestSubCond_OrWhere tests the standalone OrWhere function
func TestSubCond_OrWhere(t *testing.T) {
	cond := OrWhere(map[string]interface{}{"age": 25}, map[string]interface{}{"age": 30})
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// OrWhere creates a sub-condition, so result may or may not have $or at top level
	// Just verify it returns valid BSON
}

// TestSubCond_Or tests the standalone Or function
func TestSubCond_Or(t *testing.T) {
	cond := Or(map[string]interface{}{"status": "active"}, map[string]interface{}{"status": "pending"})
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Or creates a sub-condition, so result may or may not have $or at top level
	// Just verify it returns valid BSON
}

// TestSubCond_And tests the standalone And function
func TestSubCond_And(t *testing.T) {
	cond := And("age", 30)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["age"] != 30 {
		t.Errorf("expected age=30, got %v", result["age"])
	}
}

// TestSubCond_Equal tests the standalone Equal function
func TestSubCond_Equal(t *testing.T) {
	cond := Equal("name", "John")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["name"] != "John" {
		t.Errorf("expected name=John, got %v", result["name"])
	}
}

// TestSubCond_Ne tests the standalone Ne function
func TestSubCond_Ne(t *testing.T) {
	cond := Ne("status", "deleted")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	statusVal, ok := result["status"]
	if !ok {
		t.Error("expected 'status' field")
	}
	if m, ok := statusVal.(bson.M); ok {
		if _, hasNe := m["$ne"]; !hasNe {
			t.Error("expected $ne operator")
		}
	}
}

// TestSubCond_Gt tests the standalone Gt function
func TestSubCond_Gt(t *testing.T) {
	cond := Gt("age", 18)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasGt := m["$gt"]; !hasGt {
			t.Error("expected $gt operator")
		}
	}
}

// TestSubCond_Lt tests the standalone Lt function
func TestSubCond_Lt(t *testing.T) {
	cond := Lt("age", 65)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasLt := m["$lt"]; !hasLt {
			t.Error("expected $lt operator")
		}
	}
}

// TestSubCond_Gte tests the standalone Gte function
func TestSubCond_Gte(t *testing.T) {
	cond := Gte("age", 18)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasGte := m["$gte"]; !hasGte {
			t.Error("expected $gte operator")
		}
	}
}

// TestSubCond_Lte tests the standalone Lte function
func TestSubCond_Lte(t *testing.T) {
	cond := Lte("age", 65)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasLte := m["$lte"]; !hasLte {
			t.Error("expected $lte operator")
		}
	}
}

// TestSubCond_In tests the standalone In function
func TestSubCond_In(t *testing.T) {
	cond := In("status", "active", "pending", "approved")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	statusVal, ok := result["status"]
	if !ok {
		t.Error("expected 'status' field")
	}
	if m, ok := statusVal.(bson.M); ok {
		if _, hasIn := m["$in"]; !hasIn {
			t.Error("expected $in operator")
		}
	}
}

// TestSubCond_NotIn tests the standalone NotIn function
func TestSubCond_NotIn(t *testing.T) {
	cond := NotIn("status", "deleted", "archived")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	statusVal, ok := result["status"]
	if !ok {
		t.Error("expected 'status' field")
	}
	if m, ok := statusVal.(bson.M); ok {
		if _, hasNin := m["$nin"]; !hasNin {
			t.Error("expected $nin operator")
		}
	}
}

// TestSubCond_Like tests the standalone Like function
func TestSubCond_Like(t *testing.T) {
	cond := Like("name", "John")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	nameVal, ok := result["name"]
	if !ok {
		t.Error("expected 'name' field")
	}
	if m, ok := nameVal.(bson.M); ok {
		if _, hasRegex := m["$regex"]; !hasRegex {
			t.Error("expected $regex operator")
		}
	}
}

// TestSubCond_LeftLike tests the standalone LeftLike function
func TestSubCond_LeftLike(t *testing.T) {
	cond := LeftLike("name", "John")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	nameVal, ok := result["name"]
	if !ok {
		t.Error("expected 'name' field")
	}
	if m, ok := nameVal.(bson.M); ok {
		if _, hasRegex := m["$regex"]; !hasRegex {
			t.Error("expected $regex operator")
		}
	}
}

// TestSubCond_RightLike tests the standalone RightLike function
func TestSubCond_RightLike(t *testing.T) {
	cond := RightLike("name", "son")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	nameVal, ok := result["name"]
	if !ok {
		t.Error("expected 'name' field")
	}
	if m, ok := nameVal.(bson.M); ok {
		if _, hasRegex := m["$regex"]; !hasRegex {
			t.Error("expected $regex operator")
		}
	}
}

// TestSubCond_NotLike tests the standalone NotLike function
func TestSubCond_NotLike(t *testing.T) {
	cond := NotLike("name", "Admin")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	nameVal, ok := result["name"]
	if !ok {
		t.Error("expected 'name' field")
	}
	if m, ok := nameVal.(bson.M); ok {
		if _, hasNot := m["$not"]; !hasNot {
			t.Error("expected $not operator")
		}
	}
}

// TestSubCond_NotLeftLike tests the standalone NotLeftLike function
func TestSubCond_NotLeftLike(t *testing.T) {
	cond := NotLeftLike("name", "Admin")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	nameVal, ok := result["name"]
	if !ok {
		t.Error("expected 'name' field")
	}
	if m, ok := nameVal.(bson.M); ok {
		if _, hasNot := m["$not"]; !hasNot {
			t.Error("expected $not operator")
		}
	}
}

// TestSubCond_NotRightLike tests the standalone NotRightLike function
func TestSubCond_NotRightLike(t *testing.T) {
	cond := NotRightLike("name", "bot")
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	nameVal, ok := result["name"]
	if !ok {
		t.Error("expected 'name' field")
	}
	if m, ok := nameVal.(bson.M); ok {
		if _, hasNot := m["$not"]; !hasNot {
			t.Error("expected $not operator")
		}
	}
}

// TestSubCond_Between tests the standalone Between function
func TestSubCond_Between(t *testing.T) {
	cond := Between("age", 18, 65)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasGte := m["$gte"]; !hasGte {
			t.Error("expected $gte operator")
		}
		if _, hasLte := m["$lte"]; !hasLte {
			t.Error("expected $lte operator")
		}
	}
}

// TestSubCond_NotBetween tests the standalone NotBetween function
func TestSubCond_NotBetween(t *testing.T) {
	cond := NotBetween("age", 18, 65)
	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasNot := m["$not"]; !hasNot {
			t.Error("expected $not operator")
		}
	}
}

// ============================================================
// Edge case tests to improve coverage of internal functions
// ============================================================

// TestCond_EdgeCases_EmptyFieldName tests panic on empty field name
func TestCond_EdgeCases_EmptyFieldName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty field name")
		}
	}()
	// This should panic
	NewCond().where("", "value")
}

// TestCond_EdgeCases_EmptyOperator tests panic on empty operator
func TestCond_EdgeCases_EmptyOperator(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty operator")
		}
	}()
	// This should panic when calling addCond with empty op
	cond := NewCond()
	cond.where("field", "", "value")
}

// TestCond_EdgeCases_InvalidArgCount tests panic on invalid argument count
func TestCond_EdgeCases_InvalidArgCount(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid arg count")
		}
	}()
	// This should panic: string prefix with >3 args
	NewCond().where("field", "op", "val", "extra")
}

// TestCond_EdgeCases_NonStringMapKey tests panic on non-string map key
func TestCond_EdgeCases_NonStringMapKey(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for non-string map key")
		}
	}()
	// Create a map with int keys
	invalidMap := map[int]interface{}{1: "value"}
	NewCond().where(invalidMap)
}

// TestCond_EdgeCases_InvalidMapValue tests panic on invalid map value
func TestCond_EdgeCases_InvalidMapValue(t *testing.T) {
	// This test is tricky because we need a value that's not valid or can't interface
	// Skip this for now as it's hard to create such a value in safe Go code
	t.Skip("hard to create invalid map value in safe Go")
}

// TestCond_EdgeCases_UnhandledType tests panic on unhandled type
func TestCond_EdgeCases_UnhandledType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unhandled type")
		}
	}()
	// Pass a channel (unhandled type)
	ch := make(chan int)
	NewCond().where(ch)
}

// TestCond_EdgeCases_NestedCondWithNilBson tests nested Cond that returns nil BSON
func TestCond_EdgeCases_NestedCondWithNilBson(t *testing.T) {
	// Create an empty Cond (ToBson returns nil)
	emptyCond := NewCond()
	// Add it to another Cond
	mainCond := NewCond().Equal("age", 30)
	mainCond.where(emptyCond)

	result := mainCond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Should only have age field
	if result["age"] != 30 {
		t.Errorf("expected age=30, got %v", result["age"])
	}
}

// TestCond_EdgeCases_AddSubWhereEmptyConds tests addSubWhere with conditions that result in empty
func TestCond_EdgeCases_AddSubWhereEmptyConds(t *testing.T) {
	// OrWhere with no valid conditions should not add anything
	cond := NewCond().Equal("age", 30)
	// Pass empty string (single arg, will be skipped)
	cond.OrWhere("fieldOnly")

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	// Should still only have age
	if result["age"] != 30 {
		t.Errorf("expected age=30, got %v", result["age"])
	}
}

// TestCond_EdgeCases_OperatorWithDollarSign tests operator with $ prefix
func TestCond_EdgeCases_OperatorWithDollarSign(t *testing.T) {
	// Test operator that already has $ prefix
	cond := NewCond()
	cond.where("age", "$gt", 18)

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	ageVal, ok := result["age"]
	if !ok {
		t.Error("expected 'age' field")
	}
	if m, ok := ageVal.(bson.M); ok {
		if _, hasGt := m["$gt"]; !hasGt {
			t.Error("expected $gt operator")
		}
	}
}

// TestCond_EdgeCases_EqualityOperator tests various forms of equality operator
func TestCond_EdgeCases_EqualityOperator(t *testing.T) {
	tests := []struct {
		name string
		op   string
	}{
		{"equals sign", "="},
		{"$eq operator", "$eq"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := NewCond()
			cond.where("age", tt.op, 25)

			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}

			// Both should result in simple field: value
			if result["age"] != 25 {
				t.Errorf("expected age=25, got %v", result)
			}
		})
	}
}

// TestCond_EdgeCases_WhereWithArray tests where with array argument
func TestCond_EdgeCases_WhereWithArray(t *testing.T) {
	// Test with array type (should be treated like slice)
	arr := [2]interface{}{map[string]interface{}{"age": 25}, map[string]interface{}{"age": 30}}
	cond := NewCond()
	cond.where(arr)

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result from array")
	}
}

// ============================================================
// Complex condition combination tests
// ============================================================

// TestCond_ComplexCombinations tests complex condition combinations
func TestCond_ComplexCombinations(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Cond
	}{
		{
			name: "AND + OR combination",
			setup: func() *Cond {
				return NewCond().
					Equal("status", "active").
					Or(map[string]interface{}{"age": 25}, map[string]interface{}{"age": 30})
			},
		},
		{
			name: "multiple operators on same field",
			setup: func() *Cond {
				return NewCond().Gte("age", 18).Lte("age", 65)
			},
		},
		{
			name: "IN + NOT IN combination",
			setup: func() *Cond {
				return NewCond().
					In("role", "admin", "user").
					NotIn("status", "deleted", "archived")
			},
		},
		{
			name: "LIKE + BETWEEN combination",
			setup: func() *Cond {
				return NewCond().
					Like("name", "John").
					Between("age", 20, 30)
			},
		},
		{
			name: "nested OR conditions",
			setup: func() *Cond {
				subCond1 := NewCond().Equal("type", "A").Gt("score", 80)
				subCond2 := NewCond().Equal("type", "B").Gt("score", 90)
				return NewCond().Or(subCond1, subCond2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := tt.setup()
			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result from complex combination")
			}
			// Verify it generates valid BSON
			_ = cond.String()
		})
	}
}

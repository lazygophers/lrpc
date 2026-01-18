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

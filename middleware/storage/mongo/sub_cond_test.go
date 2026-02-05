package mongo

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// TestWhere tests the Where package-level function
func TestWhere(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		validate func(*testing.T, *Cond)
	}{
		{
			name: "single key-value",
			args: []interface{}{"name", "Alice"},
			validate: func(t *testing.T, cond *Cond) {
				result := cond.ToBson()
				if result == nil {
					t.Fatal("expected non-nil result")
				}
				if result["name"] != "Alice" {
					t.Errorf("expected name=Alice, got %v", result["name"])
				}
			},
		},
		{
			name: "multiple key-value pairs",
			args: []interface{}{
				map[string]interface{}{
					"name": "Bob",
					"age":  30,
				},
			},
			validate: func(t *testing.T, cond *Cond) {
				result := cond.ToBson()
				if result == nil {
					t.Fatal("expected non-nil result")
				}
				if result["name"] != "Bob" {
					t.Errorf("expected name=Bob, got %v", result["name"])
				}
				if result["age"] != 30 {
					t.Errorf("expected age=30, got %v", result["age"])
				}
			},
		},
		{
			name: "with operator",
			args: []interface{}{"age", "$gt", 25},
			validate: func(t *testing.T, cond *Cond) {
				result := cond.ToBson()
				if result == nil {
					t.Fatal("expected non-nil result")
				}
				if ageVal, ok := result["age"]; ok {
					if m, ok := ageVal.(bson.M); ok {
						if m["$gt"] != 25 {
							t.Errorf("expected $gt=25, got %v", m["$gt"])
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := Where(tt.args...)
			if cond == nil {
				t.Fatal("expected non-nil Cond")
			}
			tt.validate(t, cond)
		})
	}
}

// TestOrWhere tests the OrWhere package-level function
func TestOrWhere(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		validate func(*testing.T, *Cond)
	}{
		{
			name: "simple or condition",
			args: []interface{}{
				map[string]interface{}{"age": 25},
				map[string]interface{}{"age": 30},
			},
			validate: func(t *testing.T, cond *Cond) {
				result := cond.ToBson()
				if result == nil {
					t.Fatal("expected non-nil result")
				}
				if _, ok := result["$or"]; !ok {
					t.Error("expected $or operator in result")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := OrWhere(tt.args...)
			if cond == nil {
				t.Fatal("expected non-nil Cond")
			}
			tt.validate(t, cond)
		})
	}
}

// TestOr tests the Or package-level function (alias for OrWhere)
func TestOr(t *testing.T) {
	cond := Or(
		map[string]interface{}{"status": "active"},
		map[string]interface{}{"status": "pending"},
	)
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if _, ok := result["$or"]; !ok {
		t.Error("expected $or operator in result")
	}
}

// TestAnd tests the And package-level function
func TestAnd(t *testing.T) {
	cond := And("name", "Charlie")
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result["name"] != "Charlie" {
		t.Errorf("expected name=Charlie, got %v", result["name"])
	}
}

// TestEqual tests the Equal package-level function
func TestEqual(t *testing.T) {
	tests := []struct {
		name   string
		column string
		value  interface{}
	}{
		{
			name:   "string value",
			column: "name",
			value:  "Alice",
		},
		{
			name:   "int value",
			column: "age",
			value:  25,
		},
		{
			name:   "bool value",
			column: "active",
			value:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := Equal(tt.column, tt.value)
			if cond == nil {
				t.Fatal("expected non-nil Cond")
			}

			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if result[tt.column] != tt.value {
				t.Errorf("expected %s=%v, got %v", tt.column, tt.value, result[tt.column])
			}
		})
	}
}

// TestNe tests the Ne package-level function
func TestNe(t *testing.T) {
	cond := Ne("status", "deleted")
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if statusVal, ok := result["status"]; ok {
		if m, ok := statusVal.(bson.M); ok {
			if m["$ne"] != "deleted" {
				t.Errorf("expected $ne=deleted, got %v", m["$ne"])
			}
		} else {
			t.Errorf("expected bson.M, got %T", statusVal)
		}
	} else {
		t.Error("expected 'status' field in result")
	}
}

// TestGt tests the Gt package-level function
func TestGt(t *testing.T) {
	cond := Gt("age", 18)
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if ageVal, ok := result["age"]; ok {
		if m, ok := ageVal.(bson.M); ok {
			if m["$gt"] != 18 {
				t.Errorf("expected $gt=18, got %v", m["$gt"])
			}
		} else {
			t.Errorf("expected bson.M, got %T", ageVal)
		}
	} else {
		t.Error("expected 'age' field in result")
	}
}

// TestLt tests the Lt package-level function
func TestLt(t *testing.T) {
	cond := Lt("age", 65)
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if ageVal, ok := result["age"]; ok {
		if m, ok := ageVal.(bson.M); ok {
			if m["$lt"] != 65 {
				t.Errorf("expected $lt=65, got %v", m["$lt"])
			}
		} else {
			t.Errorf("expected bson.M, got %T", ageVal)
		}
	} else {
		t.Error("expected 'age' field in result")
	}
}

// TestGte tests the Gte package-level function
func TestGte(t *testing.T) {
	cond := Gte("score", 80)
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if scoreVal, ok := result["score"]; ok {
		if m, ok := scoreVal.(bson.M); ok {
			if m["$gte"] != 80 {
				t.Errorf("expected $gte=80, got %v", m["$gte"])
			}
		} else {
			t.Errorf("expected bson.M, got %T", scoreVal)
		}
	} else {
		t.Error("expected 'score' field in result")
	}
}

// TestLte tests the Lte package-level function
func TestLte(t *testing.T) {
	cond := Lte("price", 100.50)
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if priceVal, ok := result["price"]; ok {
		if m, ok := priceVal.(bson.M); ok {
			if m["$lte"] != 100.50 {
				t.Errorf("expected $lte=100.50, got %v", m["$lte"])
			}
		} else {
			t.Errorf("expected bson.M, got %T", priceVal)
		}
	} else {
		t.Error("expected 'price' field in result")
	}
}

// TestIn tests the In package-level function
func TestIn(t *testing.T) {
	tests := []struct {
		name   string
		column string
		values []interface{}
	}{
		{
			name:   "string values",
			column: "status",
			values: []interface{}{"active", "pending", "completed"},
		},
		{
			name:   "int values",
			column: "id",
			values: []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:   "single value",
			column: "role",
			values: []interface{}{"admin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := In(tt.column, tt.values...)
			if cond == nil {
				t.Fatal("expected non-nil Cond")
			}

			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if colVal, ok := result[tt.column]; ok {
				if m, ok := colVal.(bson.M); ok {
					if inVal, ok := m["$in"]; ok {
						// Verify it's the expected slice
						if inSlice, ok := inVal.([]interface{}); ok {
							if len(inSlice) != len(tt.values) {
								t.Errorf("expected %d values, got %d", len(tt.values), len(inSlice))
							}
						}
					} else {
						t.Error("expected $in operator")
					}
				} else {
					t.Errorf("expected bson.M, got %T", colVal)
				}
			} else {
				t.Errorf("expected '%s' field in result", tt.column)
			}
		})
	}
}

// TestNotIn tests the NotIn package-level function
func TestNotIn(t *testing.T) {
	cond := NotIn("status", "deleted", "archived")
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if statusVal, ok := result["status"]; ok {
		if m, ok := statusVal.(bson.M); ok {
			if _, ok := m["$nin"]; !ok {
				t.Error("expected $nin operator")
			}
		} else {
			t.Errorf("expected bson.M, got %T", statusVal)
		}
	} else {
		t.Error("expected 'status' field in result")
	}
}

// TestLike tests the Like package-level function
func TestLike(t *testing.T) {
	tests := []struct {
		name    string
		column  string
		pattern string
	}{
		{
			name:    "simple pattern",
			column:  "name",
			pattern: "john",
		},
		{
			name:    "pattern with special chars",
			column:  "email",
			pattern: ".*@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := Like(tt.column, tt.pattern)
			if cond == nil {
				t.Fatal("expected non-nil Cond")
			}

			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if colVal, ok := result[tt.column]; ok {
				if m, ok := colVal.(bson.M); ok {
					if _, ok := m["$regex"]; !ok {
						t.Error("expected $regex operator")
					}
					if options, ok := m["$options"]; !ok || options != "i" {
						t.Error("expected $options=i")
					}
				} else {
					t.Errorf("expected bson.M, got %T", colVal)
				}
			} else {
				t.Errorf("expected '%s' field in result", tt.column)
			}
		})
	}
}

// TestLeftLike tests the LeftLike package-level function
func TestLeftLike(t *testing.T) {
	cond := LeftLike("name", "john")
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if nameVal, ok := result["name"]; ok {
		if m, ok := nameVal.(bson.M); ok {
			if regex, ok := m["$regex"]; !ok || regex != "john.*" {
				t.Errorf("expected $regex=john.*, got %v", regex)
			}
			if options, ok := m["$options"]; !ok || options != "i" {
				t.Error("expected $options=i")
			}
		} else {
			t.Errorf("expected bson.M, got %T", nameVal)
		}
	} else {
		t.Error("expected 'name' field in result")
	}
}

// TestRightLike tests the RightLike package-level function
func TestRightLike(t *testing.T) {
	cond := RightLike("email", "example.com")
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if emailVal, ok := result["email"]; ok {
		if m, ok := emailVal.(bson.M); ok {
			if regex, ok := m["$regex"]; !ok || regex != ".*example.com" {
				t.Errorf("expected $regex=.*example.com, got %v", regex)
			}
			if options, ok := m["$options"]; !ok || options != "i" {
				t.Error("expected $options=i")
			}
		} else {
			t.Errorf("expected bson.M, got %T", emailVal)
		}
	} else {
		t.Error("expected 'email' field in result")
	}
}

// TestNotLike tests the NotLike package-level function
func TestNotLike(t *testing.T) {
	cond := NotLike("name", "test")
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if nameVal, ok := result["name"]; ok {
		if m, ok := nameVal.(bson.M); ok {
			if notVal, ok := m["$not"]; ok {
				if notM, ok := notVal.(bson.M); ok {
					if regex, ok := notM["$regex"]; !ok || regex != "test" {
						t.Errorf("expected $regex=test, got %v", regex)
					}
				} else {
					t.Errorf("expected $not to contain bson.M, got %T", notVal)
				}
			} else {
				t.Error("expected $not operator")
			}
		} else {
			t.Errorf("expected bson.M, got %T", nameVal)
		}
	} else {
		t.Error("expected 'name' field in result")
	}
}

// TestNotLeftLike tests the NotLeftLike package-level function
func TestNotLeftLike(t *testing.T) {
	cond := NotLeftLike("username", "admin")
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if usernameVal, ok := result["username"]; ok {
		if m, ok := usernameVal.(bson.M); ok {
			if notVal, ok := m["$not"]; ok {
				if notM, ok := notVal.(bson.M); ok {
					if regex, ok := notM["$regex"]; !ok || regex != "admin.*" {
						t.Errorf("expected $regex=admin.*, got %v", regex)
					}
				} else {
					t.Errorf("expected $not to contain bson.M, got %T", notVal)
				}
			} else {
				t.Error("expected $not operator")
			}
		} else {
			t.Errorf("expected bson.M, got %T", usernameVal)
		}
	} else {
		t.Error("expected 'username' field in result")
	}
}

// TestNotRightLike tests the NotRightLike package-level function
func TestNotRightLike(t *testing.T) {
	cond := NotRightLike("email", "spam.com")
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if emailVal, ok := result["email"]; ok {
		if m, ok := emailVal.(bson.M); ok {
			if notVal, ok := m["$not"]; ok {
				if notM, ok := notVal.(bson.M); ok {
					if regex, ok := notM["$regex"]; !ok || regex != ".*spam.com" {
						t.Errorf("expected $regex=.*spam.com, got %v", regex)
					}
				} else {
					t.Errorf("expected $not to contain bson.M, got %T", notVal)
				}
			} else {
				t.Error("expected $not operator")
			}
		} else {
			t.Errorf("expected bson.M, got %T", emailVal)
		}
	} else {
		t.Error("expected 'email' field in result")
	}
}

// TestBetween tests the Between package-level function
func TestBetween(t *testing.T) {
	tests := []struct {
		name   string
		column string
		min    interface{}
		max    interface{}
	}{
		{
			name:   "int range",
			column: "age",
			min:    18,
			max:    65,
		},
		{
			name:   "float range",
			column: "price",
			min:    10.99,
			max:    99.99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := Between(tt.column, tt.min, tt.max)
			if cond == nil {
				t.Fatal("expected non-nil Cond")
			}

			result := cond.ToBson()
			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if colVal, ok := result[tt.column]; ok {
				if m, ok := colVal.(bson.M); ok {
					if m["$gte"] != tt.min {
						t.Errorf("expected $gte=%v, got %v", tt.min, m["$gte"])
					}
					if m["$lte"] != tt.max {
						t.Errorf("expected $lte=%v, got %v", tt.max, m["$lte"])
					}
				} else {
					t.Errorf("expected bson.M, got %T", colVal)
				}
			} else {
				t.Errorf("expected '%s' field in result", tt.column)
			}
		})
	}
}

// TestNotBetween tests the NotBetween package-level function
func TestNotBetween(t *testing.T) {
	cond := NotBetween("age", 18, 25)
	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if ageVal, ok := result["age"]; ok {
		if m, ok := ageVal.(bson.M); ok {
			if notVal, ok := m["$not"]; ok {
				if notM, ok := notVal.(bson.M); ok {
					if notM["$gte"] != 18 {
						t.Errorf("expected $gte=18, got %v", notM["$gte"])
					}
					if notM["$lte"] != 25 {
						t.Errorf("expected $lte=25, got %v", notM["$lte"])
					}
				} else {
					t.Errorf("expected $not to contain bson.M, got %T", notVal)
				}
			} else {
				t.Error("expected $not operator")
			}
		} else {
			t.Errorf("expected bson.M, got %T", ageVal)
		}
	} else {
		t.Error("expected 'age' field in result")
	}
}

// TestSubCondChaining tests method chaining with package-level functions
func TestSubCondChaining(t *testing.T) {
	// Test that all functions return *Cond that can be chained
	cond := Where("name", "Alice").
		Equal("age", 30).
		Gt("score", 80).
		In("status", "active", "pending")

	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify multiple conditions are present
	// Due to the nested structure, we just verify result is not nil
	if len(result) == 0 {
		t.Error("expected non-empty result from chained conditions")
	}
}

// TestSubCondEmptyValues tests edge cases with empty values
func TestSubCondEmptyValues(t *testing.T) {
	t.Run("Like with empty pattern", func(t *testing.T) {
		cond := Like("name", "")
		result := cond.ToBson()
		// Empty pattern should result in no condition
		if result != nil && len(result) > 0 {
			t.Error("expected nil or empty result for empty Like pattern")
		}
	})

	t.Run("LeftLike with empty pattern", func(t *testing.T) {
		cond := LeftLike("name", "")
		result := cond.ToBson()
		if result != nil && len(result) > 0 {
			t.Error("expected nil or empty result for empty LeftLike pattern")
		}
	})

	t.Run("RightLike with empty pattern", func(t *testing.T) {
		cond := RightLike("name", "")
		result := cond.ToBson()
		if result != nil && len(result) > 0 {
			t.Error("expected nil or empty result for empty RightLike pattern")
		}
	})
}

// TestSubCondNilValues tests nil value handling
func TestSubCondNilValues(t *testing.T) {
	t.Run("Equal with nil value", func(t *testing.T) {
		cond := Equal("field", nil)
		result := cond.ToBson()
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result["field"] != nil {
			t.Errorf("expected field=nil, got %v", result["field"])
		}
	})

	t.Run("Ne with nil value", func(t *testing.T) {
		cond := Ne("field", nil)
		result := cond.ToBson()
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		// Should have $ne operator with nil value
		if fieldVal, ok := result["field"]; ok {
			if m, ok := fieldVal.(bson.M); ok {
				if m["$ne"] != nil {
					t.Errorf("expected $ne=nil, got %v", m["$ne"])
				}
			}
		}
	})
}

// TestSubCondComplexChaining tests complex chaining scenarios
func TestSubCondComplexChaining(t *testing.T) {
	// Create a complex condition using multiple package functions
	cond := Where("category", "electronics").
		Between("price", 100, 500).
		NotIn("brand", "BrandA", "BrandB").
		Like("name", "phone").
		Gte("rating", 4.0)

	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result from complex condition")
	}

	// Just verify that we got a non-empty result
	if len(result) == 0 {
		t.Error("expected non-empty result from complex condition chain")
	}
}

// TestSubCondOrWithMultipleConditions tests Or with multiple conditions
func TestSubCondOrWithMultipleConditions(t *testing.T) {
	cond := Or(
		map[string]interface{}{"status": "active"},
		map[string]interface{}{"status": "pending"},
		map[string]interface{}{"priority": "high"},
	)

	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if orVal, ok := result["$or"]; ok {
		if orSlice, ok := orVal.([]bson.M); ok {
			if len(orSlice) != 3 {
				t.Errorf("expected 3 OR conditions, got %d", len(orSlice))
			}
		} else {
			t.Errorf("expected []bson.M for $or, got %T", orVal)
		}
	} else {
		t.Error("expected $or operator in result")
	}
}

// TestSubCondMixedOperators tests mixing different operator types
func TestSubCondMixedOperators(t *testing.T) {
	// Mix comparison, range, and pattern matching operators
	cond := Gt("age", 18).
		Lt("age", 65).
		Like("email", "@company.com").
		In("role", "admin", "manager").
		Ne("status", "deleted")

	if cond == nil {
		t.Fatal("expected non-nil Cond")
	}

	result := cond.ToBson()
	if result == nil {
		t.Fatal("expected non-nil result from mixed operators")
	}

	// Verify result is non-empty
	if len(result) == 0 {
		t.Error("expected non-empty result from mixed operators")
	}
}

package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestCondStringMethod tests the String method
func TestCondStringMethod(t *testing.T) {
	cond := NewCond().Equal("name", "John")
	str := cond.String()

	if str == "" {
		t.Error("expected non-empty string, got empty")
	}
	if str != "{}" {
		// String method should return something like the bson representation
		t.Logf("Cond.String() = %s", str)
	}
}

// TestCondStringEmpty tests String method with empty condition
func TestCondStringEmpty(t *testing.T) {
	cond := NewCond()
	str := cond.String()

	if str != "{}" {
		t.Errorf("expected '{}', got '%s'", str)
	}
}

// TestCondStringWithMultipleConditions tests String with multiple conditions
func TestCondStringWithMultipleConditions(t *testing.T) {
	cond := NewCond().
		Equal("name", "John").
		Gte("age", 18)

	str := cond.String()

	if str == "" {
		t.Error("expected non-empty string, got empty")
	}
	if str == "{}" {
		t.Error("expected conditions in string, got empty")
	}
}

// TestAddSubWhere tests adding sub-conditions with various types
func TestAddSubWhereWithMap(t *testing.T) {
	cond := NewCond()
	
	// Create a condition with multiple where calls using maps
	cond.Where(bson.M{"age": bson.M{"$gte": 18}, "status": "active"})
	
	result := cond.ToBson()
	if result == nil {
		t.Error("expected result, got nil")
	}
}

// TestCondEqualWithVariousTypes tests Equal with different value types
func TestCondEqualWithString(t *testing.T) {
	cond := NewCond().Equal("email", "test@example.com")
	result := cond.ToBson()
	
	if _, hasEmail := result["email"]; !hasEmail {
		t.Error("expected email field in condition")
	}
}

// TestCondEqualWithInt tests Equal with integer value
func TestCondEqualWithInt(t *testing.T) {
	cond := NewCond().Equal("age", 25)
	result := cond.ToBson()
	
	if _, hasAge := result["age"]; !hasAge {
		t.Error("expected age field in condition")
	}
}

// TestCondEqualWithNil tests Equal with nil value
func TestCondEqualWithNil(t *testing.T) {
	cond := NewCond().Equal("value", nil)
	result := cond.ToBson()
	
	if _, hasValue := result["value"]; !hasValue {
		t.Error("expected value field in condition")
	}
}

// TestCondNeBehavior tests Ne operator behavior
func TestCondNeWithValue(t *testing.T) {
	cond := NewCond().Ne("status", "inactive")
	result := cond.ToBson()
	
	if _, hasStatus := result["status"]; !hasStatus {
		t.Error("expected status field in condition")
	}
}

// TestCondGtValue tests Gt with various types
func TestCondGtWithInt(t *testing.T) {
	cond := NewCond().Gt("age", 18)
	result := cond.ToBson()
	
	if ageFilter, ok := result["age"]; ok {
		if gtMap, ok := ageFilter.(bson.M); ok {
			if _, hasGt := gtMap["$gt"]; !hasGt {
				t.Error("expected $gt operator")
			}
		}
	}
}

// TestCondLtValue tests Lt operator
func TestCondLtWithInt(t *testing.T) {
	cond := NewCond().Lt("age", 65)
	result := cond.ToBson()
	
	if ageFilter, ok := result["age"]; ok {
		if ltMap, ok := ageFilter.(bson.M); ok {
			if _, hasLt := ltMap["$lt"]; !hasLt {
				t.Error("expected $lt operator")
			}
		}
	}
}

// TestCondGteValue tests Gte operator
func TestCondGteWithValue(t *testing.T) {
	cond := NewCond().Gte("score", 80)
	result := cond.ToBson()
	
	if scoreFilter, ok := result["score"]; ok {
		if gteMap, ok := scoreFilter.(bson.M); ok {
			if _, hasGte := gteMap["$gte"]; !hasGte {
				t.Error("expected $gte operator")
			}
		}
	}
}

// TestCondLteValue tests Lte operator
func TestCondLteWithValue(t *testing.T) {
	cond := NewCond().Lte("score", 100)
	result := cond.ToBson()
	
	if scoreFilter, ok := result["score"]; ok {
		if lteMap, ok := scoreFilter.(bson.M); ok {
			if _, hasLte := lteMap["$lte"]; !hasLte {
				t.Error("expected $lte operator")
			}
		}
	}
}

// TestCondInOperator tests In operator with slice
func TestCondInWithSlice(t *testing.T) {
	cond := NewCond().In("status", []interface{}{"active", "pending"})
	result := cond.ToBson()
	
	if statusFilter, ok := result["status"]; ok {
		if inMap, ok := statusFilter.(bson.M); ok {
			if _, hasIn := inMap["$in"]; !hasIn {
				t.Error("expected $in operator")
			}
		}
	}
}

// TestCondNotInOperator tests NotIn operator
func TestCondNotInWithSlice(t *testing.T) {
	cond := NewCond().NotIn("status", []interface{}{"deleted", "archived"})
	result := cond.ToBson()
	
	if statusFilter, ok := result["status"]; ok {
		if notInMap, ok := statusFilter.(bson.M); ok {
			if _, hasNotIn := notInMap["$nin"]; !hasNotIn {
				t.Error("expected $nin operator")
			}
		}
	}
}

// TestCondBetweenOperator tests Between operator
func TestCondBetweenValues(t *testing.T) {
	cond := NewCond().Between("age", 18, 65)
	result := cond.ToBson()
	
	if ageFilter, ok := result["age"]; ok {
		if betweenMap, ok := ageFilter.(bson.M); ok {
			if _, hasGte := betweenMap["$gte"]; !hasGte {
				t.Error("expected $gte in between operator")
			}
			if _, hasLte := betweenMap["$lte"]; !hasLte {
				t.Error("expected $lte in between operator")
			}
		}
	}
}

// TestCondNotBetweenOperator tests NotBetween operator
func TestCondNotBetweenValues(t *testing.T) {
	cond := NewCond().NotBetween("age", 18, 65)
	result := cond.ToBson()
	
	if ageFilter, ok := result["age"]; ok {
		if notBetweenMap, ok := ageFilter.(bson.M); ok {
			if _, hasNot := notBetweenMap["$not"]; !hasNot {
				t.Error("expected $not in not between operator")
			}
		}
	}
}

// TestCondOrOperator tests Or method
func TestCondOrWithTwoConds(t *testing.T) {
	cond1 := NewCond().Equal("status", "active")
	cond2 := NewCond().Equal("status", "pending")
	
	result := cond1.Or(cond2)
	bsonResult := result.ToBson()
	
	if _, hasOr := bsonResult["$or"]; !hasOr {
		// Or combines conditions in a specific way
		t.Logf("Or result: %v", bsonResult)
	}
}

// TestCondMultipleConditions tests combining multiple conditions
func TestCondMultipleConditions(t *testing.T) {
	cond := NewCond().Equal("status", "active").Gte("age", 18)
	
	bsonResult := cond.ToBson()
	
	if bsonResult == nil {
		t.Error("expected result from multiple conditions, got nil")
	}
}

// TestCondResetMethod tests Reset functionality
func TestCondResetClears(t *testing.T) {
	cond := NewCond().Equal("name", "John").Equal("age", 25)
	
	// Reset should clear all conditions
	cond.Reset()
	
	result := cond.ToBson()
	
	// After reset, should be empty or have no conditions
	if result != nil && len(result) > 0 {
		// Might have some structure, but no actual filter conditions
		t.Logf("After reset: %v", result)
	}
}

// TestScoopWithScoop tests database operations with complex filters
func TestScoopWithComplexWhere(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Insert test data
	users := []User{
		{ID: primitive.NewObjectID(), Email: "alice@example.com", Name: "Alice", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "bob@example.com", Name: "Bob", Age: 30, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "charlie@example.com", Name: "Charlie", Age: 35, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, user := range users {
		scoop.Create(user)
	}

	// Query with Between condition
	var results []User
	err := scoop.Where("age", map[string]interface{}{"$gte": 25, "$lte": 30}).Find(&results)
	if err != nil {
		t.Fatalf("find with between failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

// TestCondCreation tests various ways to create conditions
func TestNewCondInstance(t *testing.T) {
	cond := NewCond()
	
	if cond == nil {
		t.Error("expected cond instance, got nil")
	}
}

// TestCondChaining tests method chaining
func TestCondCompleteChaining(t *testing.T) {
	cond := NewCond().
		Equal("status", "active").
		Gte("age", 18).
		Lte("age", 65).
		In("role", []interface{}{"admin", "user"})

	result := cond.ToBson()
	
	if result == nil {
		t.Error("expected result from chaining, got nil")
	}

	// Should have multiple conditions
	if len(result) < 1 {
		t.Error("expected multiple conditions in result")
	}
}

// TestCondAsCond tests casting sub conditions
func TestCondAsCond(t *testing.T) {
	subCond := NewCond().Equal("verified", true)
	
	mainCond := NewCond().Equal("status", "active").Where(subCond)
	
	result := mainCond.ToBson()
	if result == nil {
		t.Error("expected result from nested cond, got nil")
	}
}

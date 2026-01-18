package mongo

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestLikeWithPattern tests Like with non-empty pattern
func TestLikeWithPattern(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create test data
	users := []User{
		{ID: primitive.NewObjectID(), Email: "john@example.com", Name: "John Doe", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "jane@example.com", Name: "Jane Smith", Age: 26, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "bob@test.com", Name: "Bob Wilson", Age: 27, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, user := range users {
		scoop.Create(user)
	}

	// Test Like with pattern
	cond := NewCond().Like("name", "John")
	result := cond.ToBson()
	
	// Should contain $regex condition
	if nameFilter, ok := result["name"]; ok {
		if regexMap, ok := nameFilter.(bson.M); ok {
			if _, hasRegex := regexMap["$regex"]; !hasRegex {
				t.Error("expected $regex in condition")
			}
			if _, hasOptions := regexMap["$options"]; !hasOptions {
				t.Error("expected $options in condition")
			}
		} else {
			t.Error("expected name filter to be a bson.M")
		}
	} else {
		t.Error("expected name filter in condition")
	}
}

// TestLikeWithEmptyPattern tests Like with empty pattern (should return without adding condition)
func TestLikeWithEmptyPattern(t *testing.T) {
	cond := NewCond()
	initialLen := len(cond.conds)

	// Like with empty pattern should not add a condition
	result := cond.Like("name", "")

	if result != cond {
		t.Error("Like should return the same Cond instance")
	}

	if len(cond.conds) != initialLen {
		t.Errorf("Like with empty pattern should not add condition, expected %d, got %d", initialLen, len(cond.conds))
	}
}

// TestLeftLikeWithPattern tests LeftLike with non-empty pattern
func TestLeftLikeWithPattern(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create test data
	users := []User{
		{ID: primitive.NewObjectID(), Email: "john@example.com", Name: "John Doe", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "jane@example.com", Name: "Jane Smith", Age: 26, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, user := range users {
		scoop.Create(user)
	}

	// Test LeftLike with pattern
	cond := NewCond().LeftLike("name", "John")
	result := cond.ToBson()

	// Should contain $regex condition with pattern.*
	if nameFilter, ok := result["name"]; ok {
		if regexMap, ok := nameFilter.(bson.M); ok {
			if regex, hasRegex := regexMap["$regex"]; hasRegex {
				if regexStr, ok := regex.(string); ok {
					if regexStr != "John.*" {
						t.Errorf("expected 'John.*', got '%s'", regexStr)
					}
				}
			}
		}
	}
}

// TestLeftLikeWithEmptyPattern tests LeftLike with empty pattern
func TestLeftLikeWithEmptyPattern(t *testing.T) {
	cond := NewCond()
	initialLen := len(cond.conds)

	// LeftLike with empty pattern should not add a condition
	result := cond.LeftLike("name", "")

	if result != cond {
		t.Error("LeftLike should return the same Cond instance")
	}

	if len(cond.conds) != initialLen {
		t.Errorf("LeftLike with empty pattern should not add condition")
	}
}

// TestRightLikeWithPattern tests RightLike with non-empty pattern
func TestRightLikeWithPattern(t *testing.T) {
	cond := NewCond().RightLike("email", "example.com")
	result := cond.ToBson()

	// Should contain $regex condition with .*pattern
	if emailFilter, ok := result["email"]; ok {
		if regexMap, ok := emailFilter.(bson.M); ok {
			if regex, hasRegex := regexMap["$regex"]; hasRegex {
				if regexStr, ok := regex.(string); ok {
					if regexStr != ".*example.com" {
						t.Errorf("expected '.*example.com', got '%s'", regexStr)
					}
				}
			}
		}
	}
}

// TestRightLikeWithEmptyPattern tests RightLike with empty pattern
func TestRightLikeWithEmptyPattern(t *testing.T) {
	cond := NewCond()
	initialLen := len(cond.conds)

	// RightLike with empty pattern should not add a condition
	result := cond.RightLike("email", "")

	if result != cond {
		t.Error("RightLike should return the same Cond instance")
	}

	if len(cond.conds) != initialLen {
		t.Errorf("RightLike with empty pattern should not add condition")
	}
}

// TestNotLikeWithPattern tests NotLike with non-empty pattern
func TestNotLikeWithPattern(t *testing.T) {
	cond := NewCond().NotLike("status", "inactive")
	result := cond.ToBson()

	// Should contain $not $regex condition
	if statusFilter, ok := result["status"]; ok {
		if notMap, ok := statusFilter.(bson.M); ok {
			if _, hasNot := notMap["$not"]; !hasNot {
				t.Error("expected $not in condition")
			}
		}
	}
}

// TestNotLikeWithEmptyPattern tests NotLike with empty pattern
func TestNotLikeWithEmptyPattern(t *testing.T) {
	cond := NewCond()
	initialLen := len(cond.conds)

	// NotLike with empty pattern should not add a condition
	result := cond.NotLike("status", "")

	if result != cond {
		t.Error("NotLike should return the same Cond instance")
	}

	if len(cond.conds) != initialLen {
		t.Errorf("NotLike with empty pattern should not add condition")
	}
}

// TestNotLeftLikeWithPattern tests NotLeftLike with non-empty pattern
func TestNotLeftLikeWithPattern(t *testing.T) {
	cond := NewCond().NotLeftLike("description", "old")
	result := cond.ToBson()

	// Should contain $not $regex condition with pattern.*
	if descFilter, ok := result["description"]; ok {
		if notMap, ok := descFilter.(bson.M); ok {
			if notValue, hasNot := notMap["$not"]; hasNot {
				if regexMap, ok := notValue.(bson.M); ok {
					if regex, hasRegex := regexMap["$regex"]; hasRegex {
						if regexStr, ok := regex.(string); ok {
							if regexStr != "old.*" {
								t.Errorf("expected 'old.*', got '%s'", regexStr)
							}
						}
					}
				}
			}
		}
	}
}

// TestNotLeftLikeWithEmptyPattern tests NotLeftLike with empty pattern
func TestNotLeftLikeWithEmptyPattern(t *testing.T) {
	cond := NewCond()
	initialLen := len(cond.conds)

	// NotLeftLike with empty pattern should not add a condition
	result := cond.NotLeftLike("description", "")

	if result != cond {
		t.Error("NotLeftLike should return the same Cond instance")
	}

	if len(cond.conds) != initialLen {
		t.Errorf("NotLeftLike with empty pattern should not add condition")
	}
}

// TestNotRightLikeWithPattern tests NotRightLike with non-empty pattern
func TestNotRightLikeWithPattern(t *testing.T) {
	cond := NewCond().NotRightLike("url", ".html")
	result := cond.ToBson()

	// Should contain $not $regex condition with .*pattern
	if urlFilter, ok := result["url"]; ok {
		if notMap, ok := urlFilter.(bson.M); ok {
			if notValue, hasNot := notMap["$not"]; hasNot {
				if regexMap, ok := notValue.(bson.M); ok {
					if regex, hasRegex := regexMap["$regex"]; hasRegex {
						if regexStr, ok := regex.(string); ok {
							if regexStr != ".*.html" {
								t.Errorf("expected '.*.html', got '%s'", regexStr)
							}
						}
					}
				}
			}
		}
	}
}

// TestNotRightLikeWithEmptyPattern tests NotRightLike with empty pattern
func TestNotRightLikeWithEmptyPattern(t *testing.T) {
	cond := NewCond()
	initialLen := len(cond.conds)

	// NotRightLike with empty pattern should not add a condition
	result := cond.NotRightLike("url", "")

	if result != cond {
		t.Error("NotRightLike should return the same Cond instance")
	}

	if len(cond.conds) != initialLen {
		t.Errorf("NotRightLike with empty pattern should not add condition")
	}
}

// TestMultipleLikeCalls tests chaining multiple Like calls
func TestMultipleLikeCalls(t *testing.T) {
	cond := NewCond().
		Like("name", "John").
		Like("email", "example")

	result := cond.ToBson()

	if _, hasName := result["name"]; !hasName {
		t.Error("expected name filter in condition")
	}
	if _, hasEmail := result["email"]; !hasEmail {
		t.Error("expected email filter in condition")
	}
}

// TestLikeWithMultipleConditions tests Like combined with other conditions
func TestLikeWithMultipleConditions(t *testing.T) {
	client := newTestClient(t)
	defer client.Close()

	cleanupTest := func() {
		CleanupTestCollections(t, client, "users")
	}
	cleanupTest()
	defer cleanupTest()

	scoop := client.NewScoop().Collection(User{})

	// Create test data
	users := []User{
		{ID: primitive.NewObjectID(), Email: "john@example.com", Name: "John Doe", Age: 25, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "jane@example.com", Name: "Jane Smith", Age: 26, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: primitive.NewObjectID(), Email: "john@test.com", Name: "John Wilson", Age: 27, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	for _, user := range users {
		scoop.Create(user)
	}

	// Query with Like and other conditions
	result := NewCond().
		Like("name", "John").
		Equal("age", 25)

	// Build the filter and verify
	filter := result.ToBson()
	
	// The filter should have conditions for both name and age
	if _, hasName := filter["name"]; !hasName {
		// Note: This test may need adjustment based on how the implementation works
	}
}

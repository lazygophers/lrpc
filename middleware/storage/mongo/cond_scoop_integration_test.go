package mongo_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock" // Import to register mock factory
)

// ============================================================
// Integration tests with Scoop
// ============================================================

// TestCond_ScoopIntegration tests Cond integration with Scoop
func TestCond_ScoopIntegration(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	scoop := client.NewScoop()

	// Test Where integration
	scoop = scoop.Where("age", 30).Where("status", "active")
	if scoop == nil {
		t.Fatal("expected non-nil scoop")
	}

	// Test In integration
	scoop = client.NewScoop().In("role", "admin", "user", "moderator")
	if scoop == nil {
		t.Fatal("expected non-nil scoop")
	}

	// Test NotIn integration
	scoop = client.NewScoop().NotIn("status", "deleted", "archived")
	if scoop == nil {
		t.Fatal("expected non-nil scoop")
	}

	// Test Like integration
	scoop = client.NewScoop().Like("name", "John")
	if scoop == nil {
		t.Fatal("expected non-nil scoop")
	}

	// Test Between integration
	scoop = client.NewScoop().Between("age", 18, 65)
	if scoop == nil {
		t.Fatal("expected non-nil scoop")
	}

	// Test complex chained conditions
	scoop = client.NewScoop().
		Equal("status", "active").
		Gt("age", 18).
		Lt("age", 65).
		In("role", "admin", "user").
		Like("name", "John")
	if scoop == nil {
		t.Fatal("expected non-nil scoop")
	}
}

// TestCond_ScoopWithNestedCond tests using nested Cond with Scoop
func TestCond_ScoopWithNestedCond(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Create a complex nested condition
	ageCond := mongo.NewCond().Gte("age", 18).Lte("age", 65)
	statusCond := mongo.NewCond().In("status", "active", "pending")

	// Use with Scoop
	scoop := client.NewScoop().
		Where(ageCond).
		Where(statusCond)

	if scoop == nil {
		t.Fatal("expected non-nil scoop")
	}
}

// TestCond_ScoopOrConditions tests OR conditions with Scoop
func TestCond_ScoopOrConditions(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Create OR condition using standalone Or function
	orCond := mongo.Or(
		map[string]interface{}{"age": 25},
		map[string]interface{}{"age": 30},
	)

	scoop := client.NewScoop().Where(orCond)
	if scoop == nil {
		t.Fatal("expected non-nil scoop")
	}
}

// TestCond_ScoopComplexScenarios tests complex real-world scenarios
func TestCond_ScoopComplexScenarios(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	tests := []struct {
		name  string
		setup func() interface{}
	}{
		{
			name: "user search with multiple filters",
			setup: func() interface{} {
				return client.NewScoop().
					Equal("deleted", false).
					In("role", "admin", "user").
					Gte("created_at", "2024-01-01").
					Like("email", "@example.com")
			},
		},
		{
			name: "age range with status filter",
			setup: func() interface{} {
				return client.NewScoop().
					Between("age", 18, 65).
					In("status", "active", "pending")
			},
		},
		{
			name: "complex OR with AND conditions",
			setup: func() interface{} {
				youngUsers := mongo.NewCond().Lt("age", 25).Equal("tier", "free")
				premiumUsers := mongo.NewCond().Equal("tier", "premium")
				return client.NewScoop().
					Equal("status", "active").
					Where(mongo.Or(youngUsers, premiumUsers))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.setup()
			if result == nil {
				t.Fatal("expected non-nil result")
			}
		})
	}
}

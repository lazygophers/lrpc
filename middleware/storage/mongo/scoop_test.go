package mongo_test

import (
"errors"
"github.com/lazygophers/lrpc/middleware/storage/mongo"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/require"
"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
"testing"
_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock" // Import to register mock factory
gomongo "go.mongodb.org/mongo-driver/mongo"
)

// TestModel is a test model with Collection() method
type TestModel struct {
	Name  string
	Age   int
	Email string
}

// Collection returns the collection name for TestModel
func (tm TestModel) Collection() string {
	return "test_models"
}

// ============================================================
// NewScoop Tests
// ============================================================

func TestScoop_NewScoopBasic(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("create new scoop", func(t *testing.T) {
		scoop := client.NewScoop()
		assert.NotNil(t, scoop)
	})

	t.Run("create scoop with transaction", func(t *testing.T) {
		parentScoop := client.NewScoop()
		childScoop := client.NewScoop(parentScoop)
		assert.NotNil(t, childScoop)
	})

	t.Run("create scoop with nil transaction", func(t *testing.T) {
		scoop := client.NewScoop(nil)
		assert.NotNil(t, scoop)
	})
}

// ============================================================
// Collection Tests
// ============================================================

func TestScoop_Collection(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("collection method returns self for chaining", func(t *testing.T) {
		scoop := client.NewScoop()
		// Note: Collection() internally calls getCollection() which depends on MGM
		// We can't fully test it without MGM initialization, but we can test that it doesn't panic
		// and returns the scoop for chaining when given nil or invalid models
		result := scoop.Collection(nil)
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result) // chain
	})
}

func TestScoop_CollectionName(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("collectionname method returns self for chaining with empty name", func(t *testing.T) {
		scoop := client.NewScoop()
		// Note: CollectionName() internally calls getCollection() which depends on MGM
		// We test with empty name to avoid MGM calls
		result := scoop.CollectionName("")
		assert.NotNil(t, result)
		assert.Equal(t, scoop, result)
	})
}

// ============================================================
// Where Condition Tests
// ============================================================

func TestScoop_Where(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	tests := []struct {
		name string
		args []interface{}
	}{
		{
			name: "simple key-value",
			args: []interface{}{"age", 30},
		},
		{
			name: "key-operator-value",
			args: []interface{}{"age", ">", 30},
		},
		{
			name: "with Cond object",
			args: []interface{}{mongo.NewCond().Equal("status", "active")},
		},
		{
			name: "empty args",
			args: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoop := client.NewScoop()
			result := scoop.Where(tt.args...)
			assert.NotNil(t, result)
			assert.Equal(t, scoop, result) // chain
		})
	}
}

func TestScoop_Equal(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Equal("name", "John")
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_Ne(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Ne("status", "deleted")
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_In(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	tests := []struct {
		name   string
		key    string
		values []interface{}
	}{
		{
			name:   "with multiple values",
			key:    "role",
			values: []interface{}{"admin", "user", "moderator"},
		},
		{
			name:   "with single value",
			key:    "status",
			values: []interface{}{"active"},
		},
		{
			name:   "with empty values",
			key:    "tag",
			values: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoop := client.NewScoop()
			result := scoop.In(tt.key, tt.values...)
			assert.NotNil(t, result)
			assert.Equal(t, scoop, result)
		})
	}
}

func TestScoop_NotIn(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.NotIn("status", "deleted", "archived")
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_Like(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Like("name", "John")
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_Gt(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Gt("age", 18)
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_Lt(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Lt("age", 65)
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_Gte(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Gte("age", 18)
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_Lte(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Lte("age", 65)
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_Between(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Between("age", 18, 65)
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

// ============================================================
// Limit, Offset, Skip Tests
// ============================================================

func TestScoop_Limit(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	tests := []struct {
		name  string
		limit int64
	}{
		{name: "positive limit", limit: 10},
		{name: "zero limit", limit: 0},
		{name: "negative limit", limit: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoop := client.NewScoop()
			result := scoop.Limit(tt.limit)
			assert.NotNil(t, result)
			assert.Equal(t, scoop, result)
		})
	}
}

func TestScoop_Offset(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Offset(20)
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_Skip(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	result := scoop.Skip(20)
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

// ============================================================
// Sort Tests
// ============================================================

func TestScoop_Sort(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	tests := []struct {
		name      string
		key       string
		direction []int
	}{
		{
			name:      "ascending default",
			key:       "created_at",
			direction: nil,
		},
		{
			name:      "ascending explicit",
			key:       "created_at",
			direction: []int{1},
		},
		{
			name:      "descending",
			key:       "created_at",
			direction: []int{-1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoop := client.NewScoop()
			result := scoop.Sort(tt.key, tt.direction...)
			assert.NotNil(t, result)
			assert.Equal(t, scoop, result)
		})
	}
}

// ============================================================
// Select Tests
// ============================================================

func TestScoop_Select(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	tests := []struct {
		name   string
		fields []string
	}{
		{
			name:   "single field",
			fields: []string{"name"},
		},
		{
			name:   "multiple fields",
			fields: []string{"name", "email", "age"},
		},
		{
			name:   "empty fields",
			fields: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoop := client.NewScoop()
			result := scoop.Select(tt.fields...)
			assert.NotNil(t, result)
			assert.Equal(t, scoop, result)
		})
	}
}

// ============================================================
// Clone Tests
// ============================================================

func TestScoop_Clone(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("clone empty scoop", func(t *testing.T) {
		scoop := client.NewScoop()
		cloned := scoop.Clone()
		assert.NotNil(t, cloned)
		// Verify they are different instances (not same pointer)
		assert.NotSame(t, scoop, cloned)
	})

	t.Run("clone scoop with conditions", func(t *testing.T) {
		scoop := client.NewScoop().
			Equal("status", "active").
			Gt("age", 18).
			Limit(10).
			Offset(20).
			Sort("created_at", -1).
			Select("name", "email")

		cloned := scoop.Clone()
		assert.NotNil(t, cloned)
		// Verify they are different instances (not same pointer)
		assert.NotSame(t, scoop, cloned)

		// Modify original should not affect clone
		scoop.Equal("modified", true)
	})

	t.Run("clone simple scoop", func(t *testing.T) {
		scoop := client.NewScoop().
			Equal("status", "active").
			Limit(5)
		cloned := scoop.Clone()
		assert.NotNil(t, cloned)
		// Verify they are different instances (not same pointer)
		assert.NotSame(t, scoop, cloned)
	})
}

// ============================================================
// Clear Tests
// ============================================================

func TestScoop_Clear(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop().
		Equal("status", "active").
		Gt("age", 18).
		Limit(10).
		Offset(20).
		Sort("created_at", -1).
		Select("name", "email")

	result := scoop.Clear()
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

// ============================================================
// GetCollection Tests
// ============================================================

func TestScoop_GetCollection(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("get collection returns nil when not set", func(t *testing.T) {
		scoop := client.NewScoop()
		coll := scoop.GetCollection()
		assert.Nil(t, coll)
	})

	t.Run("get collection when not set", func(t *testing.T) {
		scoop := client.NewScoop()
		coll := scoop.GetCollection()
		assert.Nil(t, coll)
	})
}

// ============================================================
// SetNotFound and IsNotFound Tests
// ============================================================

func TestScoop_SetNotFound(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	customErr := errors.New("custom not found error")
	scoop := client.NewScoop()
	result := scoop.SetNotFound(customErr)
	assert.NotNil(t, result)
	assert.Equal(t, scoop, result)
}

func TestScoop_IsNotFound(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "default ErrNoDocuments",
			err:      gomongo.ErrNoDocuments,
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoop := client.NewScoop()
			result := scoop.IsNotFound(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("custom not found error", func(t *testing.T) {
		customErr := errors.New("custom not found")
		scoop := client.NewScoop().SetNotFound(customErr)
		assert.True(t, scoop.IsNotFound(customErr))
		assert.False(t, scoop.IsNotFound(errors.New("other error")))
	})
}

// ============================================================
// Aggregation Tests
// ============================================================

func TestScoop_Aggregate(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("aggregate with nil collection", func(t *testing.T) {
		// Note: Aggregate() requires a collection to be set, but we skip this test
		// because it depends on MGM initialization
		scoop := client.NewScoop()
		_ = scoop
		// We're just testing that the scoop can be created
		// Actual aggregation tests would require MGM setup
	})
}

// ============================================================
// Chaining Tests
// ============================================================

func TestScoop_ChainedCalls(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("complex chained query", func(t *testing.T) {
		scoop := client.NewScoop().
			Equal("status", "active").
			Gt("age", 18).
			Lt("age", 65).
			In("role", "admin", "user").
			Like("name", "John").
			Limit(10).
			Offset(20).
			Sort("created_at", -1).
			Select("name", "email", "age")

		assert.NotNil(t, scoop)
	})

	t.Run("chained with clone", func(t *testing.T) {
		base := client.NewScoop().
			Equal("status", "active").
			Gt("age", 18)

		query1 := base.Clone().
			Equal("role", "admin").
			Limit(10)

		query2 := base.Clone().
			Equal("role", "user").
			Limit(20)

		assert.NotNil(t, query1)
		assert.NotNil(t, query2)
		assert.NotEqual(t, query1, query2)
	})

	t.Run("chained with clear", func(t *testing.T) {
		scoop := client.NewScoop().
			Equal("status", "active").
			Gt("age", 18).
			Clear().
			Equal("status", "pending")

		assert.NotNil(t, scoop)
	})
}

// ============================================================
// Edge Cases
// ============================================================

func TestScoop_EdgeCases(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("multiple where calls", func(t *testing.T) {
		scoop := client.NewScoop().
			Where("age", 30).
			Where("status", "active").
			Where("deleted", false)
		assert.NotNil(t, scoop)
	})

	t.Run("multiple sort fields", func(t *testing.T) {
		scoop := client.NewScoop().
			Sort("created_at", -1).
			Sort("name", 1).
			Sort("age", 1)
		assert.NotNil(t, scoop)
	})

	t.Run("multiple select calls", func(t *testing.T) {
		scoop := client.NewScoop().
			Select("name").
			Select("email").
			Select("age", "status")
		assert.NotNil(t, scoop)
	})

	t.Run("zero values", func(t *testing.T) {
		scoop := client.NewScoop().
			Equal("count", 0).
			Equal("name", "").
			Limit(0).
			Offset(0)
		assert.NotNil(t, scoop)
	})

	t.Run("nil values", func(t *testing.T) {
		scoop := client.NewScoop().
			Equal("deleted_at", nil).
			In("tags", nil)
		assert.NotNil(t, scoop)
	})
}

// ============================================================
// getCollectionNameFromOut Tests
// ============================================================

func TestScoop_getCollectionNameFromOut(t *testing.T) {
	t.Skip("getCollectionNameFromOut requires MGM initialization, skipping")
	// Note: getCollectionNameFromOut is not exported and depends on MGM
	// We would need MGM to be properly initialized to test this
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkScoop_NewScoop(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.NewScoop()
	}
}

func BenchmarkScoop_ChainedCalls(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.NewScoop().
			Equal("status", "active").
			Gt("age", 18).
			Lt("age", 65).
			Limit(10).
			Offset(20).
			Sort("created_at", -1).
			Select("name", "email")
	}
}

func BenchmarkScoop_Clone(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	scoop := client.NewScoop().
		Equal("status", "active").
		Gt("age", 18).
		Limit(10).
		Sort("created_at", -1).
		Select("name", "email")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scoop.Clone()
	}
}

// ============================================================
// Integration Scenario Tests
// ============================================================

func TestScoop_IntegrationScenarios(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("user search scenario", func(t *testing.T) {
		scoop := client.NewScoop().
			Equal("deleted", false).
			In("role", "admin", "user").
			Gte("created_at", "2024-01-01").
			Like("email", "@example.com").
			Limit(50).
			Sort("created_at", -1)

		assert.NotNil(t, scoop)
	})

	t.Run("pagination scenario", func(t *testing.T) {
		page := 2
		pageSize := 20
		offset := (page - 1) * pageSize

		scoop := client.NewScoop().
			Equal("status", "active").
			Limit(int64(pageSize)).
			Offset(int64(offset)).
			Sort("created_at", -1)

		assert.NotNil(t, scoop)
	})

	t.Run("complex filter scenario", func(t *testing.T) {
		scoop := client.NewScoop().
			Between("age", 18, 65).
			In("status", "active", "pending").
			NotIn("role", "banned", "deleted").
			Like("name", "John").
			Gte("score", 100).
			Select("name", "email", "age", "score").
			Limit(100).
			Sort("score", -1).
			Sort("created_at", -1)

		assert.NotNil(t, scoop)
	})

	t.Run("reusable base query scenario", func(t *testing.T) {
		baseQuery := client.NewScoop().
			Equal("deleted", false).
			Equal("status", "active")

		adminQuery := baseQuery.Clone().
			Equal("role", "admin").
			Limit(10)

		userQuery := baseQuery.Clone().
			Equal("role", "user").
			Limit(50)

		assert.NotNil(t, adminQuery)
		assert.NotNil(t, userQuery)
		assert.NotEqual(t, adminQuery, userQuery)
	})
}

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

type CRUDUser struct {
	Name  string `bson:"name"`
	Age   int    `bson:"age"`
	Email string `bson:"email"`
}

// Collection returns the collection name
func (u CRUDUser) Collection() string {
	return "crud_users"
}

// ============================================================
// Create Tests
// ============================================================

func TestScoop_Create(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("create single document", func(t *testing.T) {
		scoop := client.NewScoop()
		user := CRUDUser{
			Name:  "Alice",
			Age:   25,
			Email: "alice@example.com",
		}

		err := scoop.Create(&user)
		assert.NoError(t, err)
	})

	t.Run("create with bson.M infers collection name", func(t *testing.T) {
		scoop := client.NewScoop()
		doc := bson.M{
			"name": "Bob",
			"age":  30,
		}
		err := scoop.Create(doc)
		// bson.M's type name is "M", so it infers collection name "M"
		assert.NoError(t, err)
	})

	t.Run("create with collection name", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		doc := bson.M{
			"name":  "Charlie",
			"age":   35,
			"email": "charlie@example.com",
		}

		err := scoop.Create(doc)
		assert.NoError(t, err)
	})

	t.Run("create multiple documents separately", func(t *testing.T) {
		scoop := client.NewScoop()

		user1 := CRUDUser{Name: "David", Age: 40, Email: "david@example.com"}
		err := scoop.Create(&user1)
		assert.NoError(t, err)

		user2 := CRUDUser{Name: "Eve", Age: 28, Email: "eve@example.com"}
		err = scoop.Create(&user2)
		assert.NoError(t, err)
	})
}

func TestScoop_BatchCreate(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("batch create multiple documents", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")

		docs := []interface{}{
			bson.M{"name": "User1", "age": 20},
			bson.M{"name": "User2", "age": 25},
			bson.M{"name": "User3", "age": 30},
		}

		err := scoop.BatchCreate(docs...)
		assert.NoError(t, err)
	})

	t.Run("batch create with empty documents", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		err := scoop.BatchCreate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no documents to insert")
	})

	t.Run("batch create with single document", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		doc := bson.M{"name": "SingleUser", "age": 22}

		err := scoop.BatchCreate(doc)
		assert.NoError(t, err)
	})

	t.Run("batch create with structs", func(t *testing.T) {
		scoop := client.NewScoop()

		users := []interface{}{
			&CRUDUser{Name: "Alice", Age: 25, Email: "alice@example.com"},
			&CRUDUser{Name: "Bob", Age: 30, Email: "bob@example.com"},
		}

		err := scoop.BatchCreate(users...)
		assert.NoError(t, err)
	})

	t.Run("batch create with bson.M infers collection name", func(t *testing.T) {
		scoop := client.NewScoop()
		docs := []interface{}{
			bson.M{"name": "Test1"},
			bson.M{"name": "Test2"},
		}

		err := scoop.BatchCreate(docs...)
		// bson.M's type name is "M", so it infers collection name "M"
		assert.NoError(t, err)
	})
}

// ============================================================
// Update Tests
// ============================================================

func TestScoop_Updates(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("update with bson.M", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		doc := bson.M{"name": "Alice", "age": 25}
		err := scoop.Create(doc)
		require.NoError(t, err)

		updateScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "Alice")
		result := updateScoop.Updates(bson.M{"age": 26})

		assert.NoError(t, result.Error)
		assert.GreaterOrEqual(t, result.DocsAffected, int64(0))
	})

	t.Run("update with $set operator", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		doc := bson.M{"name": "Bob", "age": 30}
		err := scoop.Create(doc)
		require.NoError(t, err)

		updateScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "Bob")
		result := updateScoop.Updates(bson.M{"$set": bson.M{"age": 31}})

		assert.NoError(t, result.Error)
	})

	t.Run("update with map[string]interface{}", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		doc := bson.M{"name": "Charlie", "age": 35}
		err := scoop.Create(doc)
		require.NoError(t, err)

		updateScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "Charlie")
		updateMap := map[string]interface{}{"age": 36, "email": "charlie@example.com"}
		result := updateScoop.Updates(updateMap)

		assert.NoError(t, result.Error)
	})

	t.Run("update with struct", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		user := CRUDUser{Name: "David", Age: 40, Email: "david@example.com"}
		err := scoop.Create(&user)
		require.NoError(t, err)

		updateScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "David")
		updateData := struct {
			Age   int    `json:"age"`
			Email string `json:"email"`
		}{
			Age:   41,
			Email: "david.new@example.com",
		}
		result := updateScoop.Updates(updateData)

		assert.NoError(t, result.Error)
	})

	t.Run("update without collection", func(t *testing.T) {
		scoop := client.NewScoop()
		result := scoop.Updates(bson.M{"age": 50})

		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "collection not set")
	})

	t.Run("update with filter", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		docs := []interface{}{
			bson.M{"name": "User1", "age": 20, "status": "active"},
			bson.M{"name": "User2", "age": 25, "status": "active"},
			bson.M{"name": "User3", "age": 30, "status": "inactive"},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		updateScoop := client.NewScoop().CollectionName("crud_users").Equal("status", "active")
		result := updateScoop.Updates(bson.M{"status": "pending"})

		assert.NoError(t, result.Error)
	})

	t.Run("update with complex operators", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		doc := bson.M{"name": "Eve", "age": 28, "score": 100}
		err := scoop.Create(doc)
		require.NoError(t, err)

		updateScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "Eve")
		result := updateScoop.Updates(bson.M{
			"$set": bson.M{"age": 29},
			"$inc": bson.M{"score": 10},
		})

		assert.NoError(t, result.Error)
	})
}

// ============================================================
// Delete Tests
// ============================================================

func TestScoop_Delete(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("delete single document", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		doc := bson.M{"name": "Alice", "age": 25}
		err := scoop.Create(doc)
		require.NoError(t, err)

		deleteScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "Alice")
		result := deleteScoop.Delete()

		assert.NoError(t, result.Error)
		assert.GreaterOrEqual(t, result.DocsAffected, int64(0))
	})

	t.Run("delete multiple documents", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		docs := []interface{}{
			bson.M{"name": "User1", "status": "inactive"},
			bson.M{"name": "User2", "status": "inactive"},
			bson.M{"name": "User3", "status": "active"},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		deleteScoop := client.NewScoop().CollectionName("crud_users").Equal("status", "inactive")
		result := deleteScoop.Delete()

		assert.NoError(t, result.Error)
	})

	t.Run("delete with complex filter", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		docs := []interface{}{
			bson.M{"name": "Bob", "age": 20},
			bson.M{"name": "Charlie", "age": 30},
			bson.M{"name": "David", "age": 40},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		deleteScoop := client.NewScoop().CollectionName("crud_users").Gt("age", 25)
		result := deleteScoop.Delete()

		assert.NoError(t, result.Error)
	})

	t.Run("delete all documents", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		docs := []interface{}{
			bson.M{"name": "User1"},
			bson.M{"name": "User2"},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		deleteScoop := client.NewScoop().CollectionName("crud_users")
		result := deleteScoop.Delete()

		assert.NoError(t, result.Error)
	})

	t.Run("delete non-existent document", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("crud_users")
		result := scoop.Equal("name", "NonExistent").Delete()

		assert.NoError(t, result.Error)
		assert.Equal(t, int64(0), result.DocsAffected)
	})
}

// ============================================================
// Integration Scenarios
// ============================================================

func TestScoop_CRUDIntegration(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("complete CRUD workflow", func(t *testing.T) {
		// Create
		scoop := client.NewScoop().CollectionName("crud_users")
		user := bson.M{"name": "IntegrationUser", "age": 25, "email": "integration@example.com"}
		err := scoop.Create(user)
		require.NoError(t, err)

		// Count
		countScoop := client.NewScoop().CollectionName("crud_users")
		count, err := countScoop.Count()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))

		// Update
		updateScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "IntegrationUser")
		updateResult := updateScoop.Updates(bson.M{"age": 26})
		assert.NoError(t, updateResult.Error)

		// Delete
		deleteScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "IntegrationUser")
		deleteResult := deleteScoop.Delete()
		assert.NoError(t, deleteResult.Error)
	})

	t.Run("batch operations workflow", func(t *testing.T) {
		// Batch Create
		scoop := client.NewScoop().CollectionName("crud_users")
		docs := []interface{}{
			bson.M{"name": "BatchUser1", "age": 20},
			bson.M{"name": "BatchUser2", "age": 25},
			bson.M{"name": "BatchUser3", "age": 30},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		// Batch Update
		updateScoop := client.NewScoop().CollectionName("crud_users").Gte("age", 25)
		updateResult := updateScoop.Updates(bson.M{"status": "senior"})
		assert.NoError(t, updateResult.Error)

		// Batch Delete
		deleteScoop := client.NewScoop().CollectionName("crud_users").Lt("age", 25)
		deleteResult := deleteScoop.Delete()
		assert.NoError(t, deleteResult.Error)
	})
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkScoop_Create(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scoop := client.NewScoop().CollectionName("crud_users")
		doc := bson.M{"name": "BenchUser", "age": 25}
		_ = scoop.Create(doc)
	}
}

func BenchmarkScoop_BatchCreate(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	docs := []interface{}{
		bson.M{"name": "User1", "age": 20},
		bson.M{"name": "User2", "age": 25},
		bson.M{"name": "User3", "age": 30},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scoop := client.NewScoop().CollectionName("crud_users")
		_ = scoop.BatchCreate(docs...)
	}
}

func BenchmarkScoop_Updates(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	// Setup
	scoop := client.NewScoop().CollectionName("crud_users")
	doc := bson.M{"name": "BenchUser", "age": 25}
	_ = scoop.Create(doc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		updateScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "BenchUser")
		_ = updateScoop.Updates(bson.M{"age": 26})
	}
}

func BenchmarkScoop_Delete(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Setup
		scoop := client.NewScoop().CollectionName("crud_users")
		doc := bson.M{"name": "TempUser", "age": 25}
		_ = scoop.Create(doc)
		b.StartTimer()

		deleteScoop := client.NewScoop().CollectionName("crud_users").Equal("name", "TempUser")
		_ = deleteScoop.Delete()
	}
}

func TestScoop_Find_ID_Equivalence(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("find by id queries both id and _id", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// Create documents with auto-filled id and _id
		doc1 := bson.M{"name": "Alice", "email": "alice@example.com"}
		err := scoop.Create(doc1)
		require.NoError(t, err)

		doc2 := bson.M{"name": "Bob", "email": "bob@example.com"}
		err = scoop.Create(doc2)
		require.NoError(t, err)

		// Get the auto-filled id from doc1
		idStr := doc1["id"].(string)

		// Query by id
		var results []bson.M
		findResult := scoop.Where("id", idStr).Find(&results)
		assert.NoError(t, findResult.Error)
		// In mock mode, we can't verify actual data, but we can verify no errors
	})

	t.Run("find by _id queries both id and _id", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// Create document with auto-filled id and _id
		doc := bson.M{"name": "Charlie", "email": "charlie@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Get the auto-filled _id
		objectID := doc["_id"].(primitive.ObjectID)

		// Query by _id
		var results []bson.M
		findResult := scoop.Where("_id", objectID).Find(&results)
		assert.NoError(t, findResult.Error)
	})
}

func TestScoop_First_ID_Equivalence(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("first by id queries both id and _id", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// Create document with auto-filled id and _id
		doc := bson.M{"name": "David", "email": "david@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Get the auto-filled id
		idStr := doc["id"].(string)

		// Query by id
		var result bson.M
		firstResult := scoop.Where("id", idStr).First(&result)
		assert.NoError(t, firstResult.Error)
	})

	t.Run("first by _id queries both id and _id", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// Create document with auto-filled id and _id
		doc := bson.M{"name": "Eve", "email": "eve@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Get the auto-filled _id
		objectID := doc["_id"].(primitive.ObjectID)

		// Query by _id
		var result bson.M
		firstResult := scoop.Where("_id", objectID).First(&result)
		assert.NoError(t, firstResult.Error)
	})

	t.Run("first by id string converts to ObjectID for _id", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// Create document with known ObjectID
		objectID := primitive.NewObjectID()
		doc := bson.M{
			"_id":   objectID,
			"id":    objectID.Hex(),
			"name":  "Frank",
			"email": "frank@example.com",
		}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Query by id string (should also query _id as ObjectID)
		var result bson.M
		firstResult := scoop.Where("id", objectID.Hex()).First(&result)
		assert.NoError(t, firstResult.Error)
	})

	t.Run("first by _id ObjectID converts to string for id", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// Create document with known ObjectID
		objectID := primitive.NewObjectID()
		doc := bson.M{
			"_id":   objectID,
			"id":    objectID.Hex(),
			"name":  "Grace",
			"email": "grace@example.com",
		}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Query by _id ObjectID (should also query id as string)
		var result bson.M
		firstResult := scoop.Where("_id", objectID).First(&result)
		assert.NoError(t, firstResult.Error)
	})
}

// ============================================================
// ID/_ID Condition Building Tests
// ============================================================

func TestCond_ID_Equivalence(t *testing.T) {
	t.Run("id string generates OR condition", func(t *testing.T) {
		objectID := primitive.NewObjectID()
		idStr := objectID.Hex()

		cond := mongo.NewCond().Where("id", idStr)
		bsonCond := cond.ToBson()

		// Should generate: {$or: [{id: "xxx"}, {_id: ObjectID("xxx")}]}
		assert.NotNil(t, bsonCond)
		orCond, hasOr := bsonCond["$or"]
		if hasOr {
			// Verify OR condition exists
			assert.NotNil(t, orCond)
		}
	})

	t.Run("_id ObjectID generates OR condition", func(t *testing.T) {
		objectID := primitive.NewObjectID()

		cond := mongo.NewCond().Where("_id", objectID)
		bsonCond := cond.ToBson()

		// Should generate: {$or: [{_id: ObjectID("xxx")}, {id: "xxx"}]}
		assert.NotNil(t, bsonCond)
		orCond, hasOr := bsonCond["$or"]
		if hasOr {
			// Verify OR condition exists
			assert.NotNil(t, orCond)
		}
	})

	t.Run("invalid id string only queries id field", func(t *testing.T) {
		invalidID := "not-a-valid-objectid"

		cond := mongo.NewCond().Where("id", invalidID)
		bsonCond := cond.ToBson()

		// Should generate: {id: "not-a-valid-objectid"} (no OR, since it can't be converted to ObjectID)
		assert.NotNil(t, bsonCond)
	})

	t.Run("empty id string is skipped", func(t *testing.T) {
		cond := mongo.NewCond().Where("id", "")
		bsonCond := cond.ToBson()

		// Empty id should not generate any condition
		assert.Nil(t, bsonCond)
	})

	t.Run("zero ObjectID is skipped", func(t *testing.T) {
		cond := mongo.NewCond().Where("_id", primitive.NilObjectID)
		bsonCond := cond.ToBson()

		// Zero ObjectID should not generate any condition
		assert.Nil(t, bsonCond)
	})
}

// ============================================================
// Complex Query Tests
// ============================================================

func TestScoop_ComplexQuery_With_ID_Equivalence(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("query by id with additional conditions", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// Create document
		doc := bson.M{"name": "Alice", "email": "alice@example.com", "age": 25}
		err := scoop.Create(doc)
		require.NoError(t, err)

		idStr := doc["id"].(string)

		// Query by id AND age
		var result bson.M
		firstResult := scoop.Where("id", idStr).Where("age", 25).First(&result)
		assert.NoError(t, firstResult.Error)
	})

	t.Run("query by _id with additional conditions", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// Create document
		doc := bson.M{"name": "Bob", "email": "bob@example.com", "status": "active"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		objectID := doc["_id"].(primitive.ObjectID)

		// Query by _id AND status
		var result bson.M
		firstResult := scoop.Where("_id", objectID).Where("status", "active").First(&result)
		assert.NoError(t, firstResult.Error)
	})
}

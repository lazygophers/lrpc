package mongo_test

import (
	"errors"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock" // Import to register mock factory
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

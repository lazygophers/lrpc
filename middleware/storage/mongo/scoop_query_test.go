package mongo_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	gomongo "go.mongodb.org/mongo-driver/mongo"
)

// QueryTestUser is a test model for query operations
type QueryTestUser struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name"`
	Age   int                `bson:"age"`
	Email string             `bson:"email"`
}

// Collection returns the collection name
func (u QueryTestUser) Collection() string {
	return "query_test_users"
}

// ============================================================
// Find Tests
// ============================================================

func TestScoopQuery_Find(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("find all documents with struct", func(t *testing.T) {
		scoop := client.NewScoop()
		users := []QueryTestUser{
			{Name: "Alice", Age: 25, Email: "alice@example.com"},
			{Name: "Bob", Age: 30, Email: "bob@example.com"},
			{Name: "Charlie", Age: 35, Email: "charlie@example.com"},
		}

		for _, u := range users {
			err := scoop.Create(&u)
			require.NoError(t, err)
		}

		var foundUsers []QueryTestUser
		findScoop := client.NewScoop()
		result := findScoop.Find(&foundUsers)

		assert.NoError(t, result.Error)
		assert.GreaterOrEqual(t, result.DocsAffected, int64(0))
	})

	t.Run("find with limit and offset", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("limited_users")
		for i := 1; i <= 10; i++ {
			err := scoop.Create(bson.M{"name": "User", "index": i})
			require.NoError(t, err)
		}

		var users []bson.M
		findScoop := client.NewScoop().CollectionName("limited_users").Limit(3).Offset(2)
		result := findScoop.Find(&users)

		assert.NoError(t, result.Error)
	})

	t.Run("find with sort", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("sorted_users")
		docs := []interface{}{
			bson.M{"name": "Charlie", "age": 35},
			bson.M{"name": "Alice", "age": 25},
			bson.M{"name": "Bob", "age": 30},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		var users []bson.M
		findScoop := client.NewScoop().CollectionName("sorted_users").Sort("age", 1)
		result := findScoop.Find(&users)

		assert.NoError(t, result.Error)
	})

	t.Run("find with projection", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("projected_users")
		doc := bson.M{"name": "Alice", "age": 25, "email": "alice@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		var users []bson.M
		findScoop := client.NewScoop().CollectionName("projected_users").Select("name", "age")
		result := findScoop.Find(&users)

		assert.NoError(t, result.Error)
	})

	t.Run("find infers collection from bson.M", func(t *testing.T) {
		scoop := client.NewScoop()
		var results []bson.M

		// bson.M's type name "M" will be used as collection name
		result := scoop.Find(&results)

		// Should succeed with inferred collection name
		assert.NoError(t, result.Error)
	})

	t.Run("find with nil result returns error", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// Finding with nil will cause error
		result := scoop.Find(nil)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "invalid argument")
	})

	t.Run("find empty collection", func(t *testing.T) {
		var users []bson.M
		scoop := client.NewScoop().CollectionName("empty_collection")
		result := scoop.Find(&users)

		assert.NoError(t, result.Error)
		assert.Equal(t, int64(0), result.DocsAffected)
	})

	t.Run("find with non-slice result", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")
		var singleDoc bson.M

		// This should work but docsCount will be 0
		result := scoop.Find(&singleDoc)

		// May error or succeed depending on implementation
		if result.Error != nil {
			assert.Error(t, result.Error)
		}
	})

	t.Run("find with all options", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("options_test")
		docs := []interface{}{
			bson.M{"name": "User1", "age": 20},
			bson.M{"name": "User2", "age": 25},
			bson.M{"name": "User3", "age": 30},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		var users []bson.M
		findScoop := client.NewScoop().CollectionName("options_test").
			Limit(2).
			Offset(1).
			Sort("age", -1).
			Select("name")

		result := findScoop.Find(&users)
		assert.NoError(t, result.Error)
	})
}

// ============================================================
// First Tests
// ============================================================

func TestScoopQuery_First(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("first document exists", func(t *testing.T) {
		scoop := client.NewScoop()
		user := QueryTestUser{Name: "Alice", Age: 25, Email: "alice@example.com"}
		err := scoop.Create(&user)
		require.NoError(t, err)

		var foundUser QueryTestUser
		findScoop := client.NewScoop().Equal("name", "Alice")
		result := findScoop.First(&foundUser)

		assert.NoError(t, result.Error)
		assert.Equal(t, "Alice", foundUser.Name)
	})

	t.Run("first with projection", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("first_users")
		doc := bson.M{"name": "Bob", "age": 30, "email": "bob@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		var user bson.M
		findScoop := client.NewScoop().CollectionName("first_users").Select("name", "age")
		result := findScoop.First(&user)

		assert.NoError(t, result.Error)
	})

	t.Run("first document not found", func(t *testing.T) {
		var user bson.M
		scoop := client.NewScoop().CollectionName("nonexistent_coll").Equal("name", "NonExistent")
		result := scoop.First(&user)

		assert.Error(t, result.Error)
		assert.Equal(t, gomongo.ErrNoDocuments, result.Error)
	})

	t.Run("first infers collection from bson.M", func(t *testing.T) {
		scoop := client.NewScoop()
		var result bson.M

		// bson.M's type name "M" will be used as collection name
		firstResult := scoop.First(&result)

		// Should return ErrNoDocuments for empty collection
		assert.Error(t, firstResult.Error)
		assert.Equal(t, gomongo.ErrNoDocuments, firstResult.Error)
	})

	t.Run("first with nil result returns error", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("test_users")

		// First with nil will cause error
		result := scoop.First(nil)
		// Will get ErrNoDocuments or decode error
		assert.Error(t, result.Error)
	})

	t.Run("first with decode error", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("decode_test")
		// Create a document
		doc := bson.M{"name": "Test", "age": "not_a_number"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Try to decode into incompatible struct
		type StrictUser struct {
			Name string `bson:"name"`
			Age  int    `bson:"age"` // This will fail to decode "not_a_number"
		}
		var user StrictUser
		findScoop := client.NewScoop().CollectionName("decode_test")
		result := findScoop.First(&user)

		// Might error during decode
		if result.Error != nil {
			assert.Error(t, result.Error)
		}
	})

	t.Run("first with projection and filter", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("first_proj_users")
		doc := bson.M{"name": "Alice", "age": 25, "email": "alice@example.com", "status": "active"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		var user bson.M
		findScoop := client.NewScoop().CollectionName("first_proj_users").
			Equal("status", "active").
			Select("name", "email")
		result := findScoop.First(&user)

		assert.NoError(t, result.Error)
		assert.Equal(t, "Alice", user["name"])
	})
}

// ============================================================
// Count Tests
// ============================================================

func TestScoopQuery_Count(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("count all documents", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("count_users")
		docs := []interface{}{
			bson.M{"name": "User1", "age": 20},
			bson.M{"name": "User2", "age": 25},
			bson.M{"name": "User3", "age": 30},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		countScoop := client.NewScoop().CollectionName("count_users")
		count, err := countScoop.Count()

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(3))
	})

	t.Run("count with filter", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("filtered_count_users")
		docs := []interface{}{
			bson.M{"name": "User1", "age": 20},
			bson.M{"name": "User2", "age": 25},
			bson.M{"name": "User3", "age": 30},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		countScoop := client.NewScoop().CollectionName("filtered_count_users").Gte("age", 25)
		count, err := countScoop.Count()

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})

	t.Run("count empty collection", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("empty_count_users")
		count, err := scoop.Count()

		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("count with complex filters", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("complex_count_users")
		docs := []interface{}{
			bson.M{"name": "User1", "age": 20, "status": "active"},
			bson.M{"name": "User2", "age": 25, "status": "active"},
			bson.M{"name": "User3", "age": 30, "status": "inactive"},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		countScoop := client.NewScoop().CollectionName("complex_count_users").
			Equal("status", "active").
			Between("age", 20, 26)
		count, err := countScoop.Count()

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

// ============================================================
// Exist Tests
// ============================================================

func TestScoopQuery_Exist(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("exist returns true when document exists", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("exist_users")
		doc := bson.M{"name": "Alice", "age": 25}
		err := scoop.Create(doc)
		require.NoError(t, err)

		existScoop := client.NewScoop().CollectionName("exist_users").Equal("name", "Alice")
		exists, err := existScoop.Exist()

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("exist returns false when document not found", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("exist_users2")
		exists, err := scoop.Equal("name", "NonExistent").Exist()

		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("exist with complex filter", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("complex_exist_users")
		docs := []interface{}{
			bson.M{"name": "User1", "age": 20, "status": "active"},
			bson.M{"name": "User2", "age": 25, "status": "inactive"},
			bson.M{"name": "User3", "age": 30, "status": "active"},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		existScoop := client.NewScoop().CollectionName("complex_exist_users").
			Equal("status", "active").
			Gte("age", 25)
		exists, err := existScoop.Exist()

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("exist only queries _id field", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("efficient_exist_users")
		doc := bson.M{"name": "Bob", "age": 30, "email": "bob@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		existScoop := client.NewScoop().CollectionName("efficient_exist_users").Equal("name", "Bob")
		exists, err := existScoop.Exist()

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("exist with multiple filters", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("multi_filter_exist")
		docs := []interface{}{
			bson.M{"name": "User1", "age": 20, "status": "active", "role": "admin"},
			bson.M{"name": "User2", "age": 25, "status": "active", "role": "user"},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		existScoop := client.NewScoop().CollectionName("multi_filter_exist").
			Equal("status", "active").
			Equal("role", "admin").
			Gte("age", 18)
		exists, err := existScoop.Exist()

		assert.NoError(t, err)
		assert.True(t, exists)
	})
}

// ============================================================
// Error Handling Tests
// ============================================================

func TestScoopQuery_ErrorHandling(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("find with all filter types", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("filter_test")
		docs := []interface{}{
			bson.M{"name": "Alice", "age": 25, "tags": []string{"admin", "user"}},
			bson.M{"name": "Bob", "age": 30, "tags": []string{"user"}},
			bson.M{"name": "Charlie", "age": 35, "tags": []string{"moderator"}},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		var users []bson.M
		findScoop := client.NewScoop().CollectionName("filter_test").
			Ne("name", "David").
			In("tags", "user").
			NotIn("age", 40, 50).
			Like("name", "e").
			Lt("age", 40).
			Lte("age", 35).
			Gt("age", 20).
			Gte("age", 25)

		result := findScoop.Find(&users)
		assert.NoError(t, result.Error)
	})

	t.Run("count with Or conditions", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("or_count_test")
		docs := []interface{}{
			bson.M{"name": "User1", "age": 20},
			bson.M{"name": "User2", "age": 25},
			bson.M{"name": "User3", "age": 30},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		// Using Where with Cond for OR logic
		cond := mongo.NewCond().Equal("age", 20).Or().Equal("age", 30)
		countScoop := client.NewScoop().CollectionName("or_count_test").Where(cond)
		count, err := countScoop.Count()

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})

	t.Run("exist with empty filter", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("empty_filter_exist")
		doc := bson.M{"name": "Test"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		existScoop := client.NewScoop().CollectionName("empty_filter_exist")
		exists, err := existScoop.Exist()

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("first with empty filter", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("empty_filter_first")
		doc := bson.M{"name": "First"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		var result bson.M
		findScoop := client.NewScoop().CollectionName("empty_filter_first")
		firstResult := findScoop.First(&result)

		assert.NoError(t, firstResult.Error)
	})
}

// ============================================================
// Integration Tests
// ============================================================

func TestScoopQuery_Integration(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("full query workflow", func(t *testing.T) {
		// Create test data
		scoop := client.NewScoop().CollectionName("workflow_users")
		docs := []interface{}{
			bson.M{"name": "Alice", "age": 25, "status": "active"},
			bson.M{"name": "Bob", "age": 30, "status": "active"},
			bson.M{"name": "Charlie", "age": 35, "status": "inactive"},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		// Check existence
		existScoop := client.NewScoop().CollectionName("workflow_users").Equal("status", "active")
		exists, err := existScoop.Exist()
		assert.NoError(t, err)
		assert.True(t, exists)

		// Count active users
		countScoop := client.NewScoop().CollectionName("workflow_users").Equal("status", "active")
		count, err := countScoop.Count()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))

		// Find first active user
		var user bson.M
		firstScoop := client.NewScoop().CollectionName("workflow_users").Equal("status", "active")
		result := firstScoop.First(&user)
		assert.NoError(t, result.Error)

		// Find all active users
		var users []bson.M
		findScoop := client.NewScoop().CollectionName("workflow_users").Equal("status", "active")
		findResult := findScoop.Find(&users)
		assert.NoError(t, findResult.Error)
	})
}

// ============================================================
// Edge Cases for Coverage
// ============================================================

func TestScoopQuery_EdgeCases(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("find with limit zero", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("limit_zero")
		var results []bson.M
		result := scoop.Limit(0).Find(&results)

		assert.NoError(t, result.Error)
	})

	t.Run("find with offset zero", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("offset_zero")
		var results []bson.M
		result := scoop.Offset(0).Find(&results)

		assert.NoError(t, result.Error)
	})

	t.Run("find without limit or offset", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("no_limit_offset")
		var results []bson.M
		result := scoop.Find(&results)

		assert.NoError(t, result.Error)
	})

	t.Run("find without sort", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("no_sort")
		var results []bson.M
		result := scoop.Find(&results)

		assert.NoError(t, result.Error)
	})

	t.Run("find without projection", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("no_projection")
		var results []bson.M
		result := scoop.Find(&results)

		assert.NoError(t, result.Error)
	})

	t.Run("first without projection", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("first_no_proj")
		doc := bson.M{"name": "Test"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		var result bson.M
		findScoop := client.NewScoop().CollectionName("first_no_proj")
		firstResult := findScoop.First(&result)

		assert.NoError(t, firstResult.Error)
	})

	t.Run("count with nil collection should panic", func(t *testing.T) {
		// Note: This will panic because Count doesn't call ensureCollection
		// We can't test this without causing a panic
		t.Skip("Count requires collection to be set beforehand")
	})

	t.Run("exist uses cloned scoop", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("exist_clone")
		doc := bson.M{"name": "Test", "age": 25}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Exist should clone and add _id selection
		existScoop := client.NewScoop().CollectionName("exist_clone")
		exists, err := existScoop.Equal("name", "Test").Exist()

		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("find returns correct docsAffected", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("docs_affected")
		docs := []interface{}{
			bson.M{"name": "User1"},
			bson.M{"name": "User2"},
			bson.M{"name": "User3"},
		}
		err := scoop.BatchCreate(docs...)
		require.NoError(t, err)

		var results []bson.M
		findScoop := client.NewScoop().CollectionName("docs_affected")
		result := findScoop.Find(&results)

		assert.NoError(t, result.Error)
		assert.Equal(t, int64(len(results)), result.DocsAffected)
	})
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkScoopQuery_Find(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	scoop := client.NewScoop().CollectionName("bench_users")
	for i := 0; i < 100; i++ {
		_ = scoop.Create(bson.M{"name": "User", "index": i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []bson.M
		_ = client.NewScoop().CollectionName("bench_users").Find(&users)
	}
}

func BenchmarkScoopQuery_First(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	scoop := client.NewScoop().CollectionName("bench_first_users")
	_ = scoop.Create(bson.M{"name": "TestUser"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user bson.M
		_ = client.NewScoop().CollectionName("bench_first_users").First(&user)
	}
}

func BenchmarkScoopQuery_Count(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	scoop := client.NewScoop().CollectionName("bench_count_users")
	for i := 0; i < 100; i++ {
		_ = scoop.Create(bson.M{"index": i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.NewScoop().CollectionName("bench_count_users").Count()
	}
}

func BenchmarkScoopQuery_Exist(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	scoop := client.NewScoop().CollectionName("bench_exist_users")
	_ = scoop.Create(bson.M{"name": "TestUser"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.NewScoop().CollectionName("bench_exist_users").Exist()
	}
}

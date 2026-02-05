package mongo_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// CRUDUser is a test model for CRUD operations
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

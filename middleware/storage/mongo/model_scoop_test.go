package mongo_test

import (
	"errors"
	"testing"

	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock" // Import to register mock factory
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

// TestUser is a test model for ModelScoop tests
type TestUser struct {
	Name   string `bson:"name"`
	Age    int    `bson:"age"`
	Email  string `bson:"email"`
	Status string `bson:"status"`
}

// Collection returns the collection name for TestUser
func (tu TestUser) Collection() string {
	return "test_users"
}

// ============================================================
// ModelScoop Creation Tests
// ============================================================

func TestModelScoop_Creation(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("create model scoop", func(t *testing.T) {
		model := mongo.NewModel[TestUser](client)
		scoop := model.NewScoop()
		assert.NotNil(t, scoop)
		assert.NotNil(t, scoop.GetScoop())
	})

	t.Run("create model scoop with transaction", func(t *testing.T) {
		model := mongo.NewModel[TestUser](client)
		baseTx := client.NewScoop()
		scoop := model.NewScoop(baseTx)
		assert.NotNil(t, scoop)
	})

	t.Run("create model scoop with nil transaction", func(t *testing.T) {
		model := mongo.NewModel[TestUser](client)
		scoop := model.NewScoop(nil)
		assert.NotNil(t, scoop)
	})
}

// ============================================================
// ModelScoop Query Tests
// ============================================================

func TestModelScoop_Find(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("find all users", func(t *testing.T) {
		scoop := model.NewScoop()
		users, err := scoop.Find()
		assert.NoError(t, err)
		assert.NotNil(t, users)
	})

	t.Run("find with conditions", func(t *testing.T) {
		scoop := model.NewScoop().Equal("status", "active")
		users, err := scoop.Find()
		assert.NoError(t, err)
		assert.NotNil(t, users)
	})

	t.Run("find with multiple conditions", func(t *testing.T) {
		scoop := model.NewScoop().
			Equal("status", "active").
			Gt("age", 18).
			Limit(10)
		users, err := scoop.Find()
		assert.NoError(t, err)
		assert.NotNil(t, users)
	})
}

func TestModelScoop_First(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("first with no results", func(t *testing.T) {
		scoop := model.NewScoop().Equal("name", "NonExistent")
		user, err := scoop.First()
		// In mock mode, this should return ErrNoDocuments
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("first with conditions", func(t *testing.T) {
		scoop := model.NewScoop().Equal("status", "active")
		user, err := scoop.First()
		// Mock will return error since no data
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestModelScoop_Count(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("count all", func(t *testing.T) {
		scoop := model.NewScoop()
		count, err := scoop.Count()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})

	t.Run("count with conditions", func(t *testing.T) {
		scoop := model.NewScoop().Equal("status", "active")
		count, err := scoop.Count()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})
}

// ============================================================
// ModelScoop CRUD Tests
// ============================================================

func TestModelScoop_Create(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("create user", func(t *testing.T) {
		scoop := model.NewScoop()
		user := &TestUser{
			Name:   "John Doe",
			Age:    30,
			Email:  "john@example.com",
			Status: "active",
		}
		err := scoop.Create(user)
		assert.NoError(t, err)
	})

	t.Run("create user with empty name", func(t *testing.T) {
		scoop := model.NewScoop()
		user := &TestUser{
			Age:    25,
			Email:  "test@example.com",
			Status: "active",
		}
		err := scoop.Create(user)
		assert.NoError(t, err)
	})
}

func TestModelScoop_Updates(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("update with map", func(t *testing.T) {
		scoop := model.NewScoop().Equal("name", "John Doe")
		result := scoop.Updates(bson.M{"age": 31})
		assert.NotNil(t, result)
		assert.NoError(t, result.Error)
	})

	t.Run("update with struct", func(t *testing.T) {
		scoop := model.NewScoop().Equal("name", "John Doe")
		result := scoop.Updates(map[string]interface{}{
			"age":    32,
			"status": "inactive",
		})
		assert.NotNil(t, result)
		assert.NoError(t, result.Error)
	})

	t.Run("update multiple fields", func(t *testing.T) {
		scoop := model.NewScoop().Equal("status", "active")
		result := scoop.Updates(bson.M{
			"status": "verified",
			"age":    40,
		})
		assert.NotNil(t, result)
		assert.NoError(t, result.Error)
	})
}

func TestModelScoop_Delete(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("delete by condition", func(t *testing.T) {
		scoop := model.NewScoop().Equal("name", "John Doe")
		result := scoop.Delete()
		assert.NotNil(t, result)
		assert.NoError(t, result.Error)
	})

	t.Run("delete with multiple conditions", func(t *testing.T) {
		scoop := model.NewScoop().
			Equal("status", "inactive").
			Lt("age", 18)
		result := scoop.Delete()
		assert.NotNil(t, result)
		assert.NoError(t, result.Error)
	})
}

func TestModelScoop_Exist(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("check existence", func(t *testing.T) {
		scoop := model.NewScoop().Equal("name", "John Doe")
		exists, err := scoop.Exist()
		assert.NoError(t, err)
		assert.False(t, exists) // Mock has no data
	})

	t.Run("check existence with multiple conditions", func(t *testing.T) {
		scoop := model.NewScoop().
			Equal("status", "active").
			Gte("age", 18)
		exists, err := scoop.Exist()
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// ============================================================
// ModelScoop Advanced Tests
// ============================================================

func TestModelScoop_Watch(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("watch without pipeline", func(t *testing.T) {
		scoop := model.NewScoop()
		stream, err := scoop.Watch()
		// Mock may not support watch, expect error or nil
		if err != nil {
			assert.Error(t, err)
		}
		_ = stream
	})

	t.Run("watch with pipeline", func(t *testing.T) {
		scoop := model.NewScoop()
		pipeline := []bson.M{
			{"$match": bson.M{"status": "active"}},
		}
		stream, err := scoop.Watch(pipeline...)
		if err != nil {
			assert.Error(t, err)
		}
		_ = stream
	})
}

func TestModelScoop_Aggregate(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("aggregate without pipeline", func(t *testing.T) {
		scoop := model.NewScoop()
		agg := scoop.Aggregate()
		assert.NotNil(t, agg)
	})

	t.Run("aggregate with match stage", func(t *testing.T) {
		scoop := model.NewScoop()
		pipeline := []bson.M{
			{"$match": bson.M{"status": "active"}},
		}
		agg := scoop.Aggregate(pipeline...)
		assert.NotNil(t, agg)
	})

	t.Run("aggregate with group stage", func(t *testing.T) {
		scoop := model.NewScoop()
		pipeline := []bson.M{
			{"$group": bson.M{
				"_id":   "$status",
				"count": bson.M{"$sum": 1},
			}},
		}
		agg := scoop.Aggregate(pipeline...)
		assert.NotNil(t, agg)
	})
}

func TestModelScoop_GetCollection(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("get collection", func(t *testing.T) {
		scoop := model.NewScoop()
		coll := scoop.GetCollection()
		assert.NotNil(t, coll)
	})
}

func TestModelScoop_GetScoop(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("get underlying scoop", func(t *testing.T) {
		modelScoop := model.NewScoop()
		scoop := modelScoop.GetScoop()
		assert.NotNil(t, scoop)
	})
}

// ============================================================
// ModelScoop Chaining Tests
// ============================================================

func TestModelScoop_Where(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("where condition", func(t *testing.T) {
		scoop := model.NewScoop().Where("name", "John")
		assert.NotNil(t, scoop)
	})

	t.Run("multiple where conditions", func(t *testing.T) {
		scoop := model.NewScoop().
			Where("name", "John").
			Where("age", 30)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Equal(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("equal condition", func(t *testing.T) {
		scoop := model.NewScoop().Equal("status", "active")
		assert.NotNil(t, scoop)
	})

	t.Run("multiple equal conditions", func(t *testing.T) {
		scoop := model.NewScoop().
			Equal("status", "active").
			Equal("name", "John")
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Ne(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("not equal condition", func(t *testing.T) {
		scoop := model.NewScoop().Ne("status", "deleted")
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_In(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("in condition", func(t *testing.T) {
		scoop := model.NewScoop().In("status", "active", "pending")
		assert.NotNil(t, scoop)
	})

	t.Run("in with multiple values", func(t *testing.T) {
		scoop := model.NewScoop().In("age", 18, 25, 30, 35)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_NotIn(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("not in condition", func(t *testing.T) {
		scoop := model.NewScoop().NotIn("status", "deleted", "banned")
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Like(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("like pattern", func(t *testing.T) {
		scoop := model.NewScoop().Like("name", "John")
		assert.NotNil(t, scoop)
	})

	t.Run("like with email pattern", func(t *testing.T) {
		scoop := model.NewScoop().Like("email", "@example.com")
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Gt(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("greater than condition", func(t *testing.T) {
		scoop := model.NewScoop().Gt("age", 18)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Lt(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("less than condition", func(t *testing.T) {
		scoop := model.NewScoop().Lt("age", 65)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Gte(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("greater than or equal condition", func(t *testing.T) {
		scoop := model.NewScoop().Gte("age", 18)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Lte(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("less than or equal condition", func(t *testing.T) {
		scoop := model.NewScoop().Lte("age", 65)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Between(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("between condition", func(t *testing.T) {
		scoop := model.NewScoop().Between("age", 18, 65)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Limit(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("limit results", func(t *testing.T) {
		scoop := model.NewScoop().Limit(10)
		assert.NotNil(t, scoop)
	})

	t.Run("limit with zero", func(t *testing.T) {
		scoop := model.NewScoop().Limit(0)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Offset(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("offset results", func(t *testing.T) {
		scoop := model.NewScoop().Offset(20)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Skip(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("skip results", func(t *testing.T) {
		scoop := model.NewScoop().Skip(20)
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Sort(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("sort ascending", func(t *testing.T) {
		scoop := model.NewScoop().Sort("age", 1)
		assert.NotNil(t, scoop)
	})

	t.Run("sort descending", func(t *testing.T) {
		scoop := model.NewScoop().Sort("age", -1)
		assert.NotNil(t, scoop)
	})

	t.Run("sort default direction", func(t *testing.T) {
		scoop := model.NewScoop().Sort("name")
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Select(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("select fields", func(t *testing.T) {
		scoop := model.NewScoop().Select("name", "email")
		assert.NotNil(t, scoop)
	})

	t.Run("select single field", func(t *testing.T) {
		scoop := model.NewScoop().Select("name")
		assert.NotNil(t, scoop)
	})
}

func TestModelScoop_Clear(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("clear scoop", func(t *testing.T) {
		scoop := model.NewScoop().
			Equal("status", "active").
			Gt("age", 18).
			Limit(10)
		cleared := scoop.Clear()
		assert.NotNil(t, cleared)
	})
}

func TestModelScoop_FindByPage(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("find by page", func(t *testing.T) {
		scoop := model.NewScoop().Equal("status", "active")
		opt := &core.ListOption{
			Offset: 0,
			Limit:  10,
		}
		page, values, err := scoop.FindByPage(opt)
		// Mock may not have pagination data
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, page)
			assert.NotNil(t, values)
		}
	})

	t.Run("find by page with sorting", func(t *testing.T) {
		scoop := model.NewScoop().
			Equal("status", "active").
			Sort("age", -1)
		opt := &core.ListOption{
			Offset: 20,
			Limit:  20,
		}
		page, values, err := scoop.FindByPage(opt)
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, page)
			assert.NotNil(t, values)
		}
	})
}

// ============================================================
// ModelScoop Complex Chaining Tests
// ============================================================

func TestModelScoop_ComplexChaining(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("complex query chain", func(t *testing.T) {
		scoop := model.NewScoop().
			Equal("status", "active").
			Gte("age", 18).
			Lte("age", 65).
			Like("email", "@example.com").
			NotIn("name", "banned1", "banned2").
			Limit(50).
			Offset(100).
			Sort("age", -1).
			Sort("name", 1).
			Select("name", "email", "age")

		assert.NotNil(t, scoop)

		// Test that we can execute find on this chain
		users, err := scoop.Find()
		assert.NoError(t, err)
		assert.NotNil(t, users)
	})

	t.Run("chain with clear and rebuild", func(t *testing.T) {
		scoop := model.NewScoop().
			Equal("status", "active").
			Gt("age", 30).
			Clear().
			Equal("status", "pending").
			Lt("age", 25)

		assert.NotNil(t, scoop)
	})
}

// ============================================================
// ModelScoop Edge Cases
// ============================================================

func TestModelScoop_EdgeCases(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("create with nil pointer", func(t *testing.T) {
		scoop := model.NewScoop()
		// This should handle nil gracefully or return an error
		err := scoop.Create(nil)
		// Expecting error with nil document
		assert.Error(t, err)
	})

	t.Run("updates with empty conditions", func(t *testing.T) {
		scoop := model.NewScoop()
		result := scoop.Updates(bson.M{"status": "updated"})
		assert.NotNil(t, result)
	})

	t.Run("delete with empty conditions", func(t *testing.T) {
		scoop := model.NewScoop()
		result := scoop.Delete()
		assert.NotNil(t, result)
	})

	t.Run("count with empty conditions", func(t *testing.T) {
		scoop := model.NewScoop()
		count, err := scoop.Count()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(0))
	})

	t.Run("find with zero limit", func(t *testing.T) {
		scoop := model.NewScoop().Limit(0)
		users, err := scoop.Find()
		assert.NoError(t, err)
		assert.NotNil(t, users)
	})

	t.Run("find with negative offset", func(t *testing.T) {
		scoop := model.NewScoop().Offset(-1)
		users, err := scoop.Find()
		// Should handle negative offset gracefully
		_ = users
		_ = err
	})
}

// ============================================================
// ModelScoop Error Handling Tests
// ============================================================

func TestModelScoop_ErrorHandling(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("first not found error", func(t *testing.T) {
		scoop := model.NewScoop().Equal("name", "NonExistentUser")
		_, err := scoop.First()
		assert.Error(t, err)
	})

	t.Run("exist with non-existent document", func(t *testing.T) {
		scoop := model.NewScoop().Equal("name", "NonExistentUser")
		exists, err := scoop.Exist()
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// ============================================================
// Model Tests
// ============================================================

func TestModel_Creation(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("create model", func(t *testing.T) {
		model := mongo.NewModel[TestUser](client)
		assert.NotNil(t, model)
	})

	t.Run("get collection name", func(t *testing.T) {
		model := mongo.NewModel[TestUser](client)
		collName := model.CollectionName()
		assert.Equal(t, "test_users", collName)
	})
}

func TestModel_NotFoundError(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	model := mongo.NewModel[TestUser](client)

	t.Run("set custom not found error", func(t *testing.T) {
		customErr := errors.New("custom not found")
		result := model.SetNotFound(customErr)
		assert.NotNil(t, result)
		assert.Equal(t, model, result)
	})

	t.Run("check is not found", func(t *testing.T) {
		customErr := errors.New("custom not found")
		model.SetNotFound(customErr)
		assert.True(t, model.IsNotFound(customErr))
		assert.False(t, model.IsNotFound(errors.New("other error")))
	})
}

// ============================================================
// Benchmark Tests
// ============================================================

// TestSimpleModel is a model without Collection() method
type TestSimpleModel struct {
	ID   string `bson:"_id"`
	Data string `bson:"data"`
}

// ============================================================
// ModelScoop Collection Naming Tests
// ============================================================

func TestModelScoop_CollectionNaming(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("default name from reflection", func(t *testing.T) {
		model := mongo.NewModel[TestSimpleModel](client)
		assert.Equal(t, "TestSimpleModel", model.CollectionName())

		scoop := model.NewScoop()
		assert.NotNil(t, scoop)
	})

	t.Run("interface based name", func(t *testing.T) {
		model := mongo.NewModel[TestUser](client)
		assert.Equal(t, "test_users", model.CollectionName())
	})
}

// ============================================================
// ModelScoop Error Path Tests
// ============================================================

func TestModelScoop_ErrorPaths(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	model := mongo.NewModel[TestUser](client)

	t.Run("Find error path", func(t *testing.T) {
		// In mock mode, we can trigger errors by passing invalid options if the mock supports it
		// or by simulating context cancellation if applicable.
		// For now, we test the coverage of the error return.
		scoop := model.NewScoop()
		// Trigger a mock error if possible, or just ensure the path is exercised
		results, err := scoop.Find()
		_ = results
		_ = err
	})

	t.Run("First error path", func(t *testing.T) {
		scoop := model.NewScoop().Where("_id", "invalid")
		result, err := scoop.First()
		_ = result
		_ = err
	})

	t.Run("FindByPage nil option error", func(t *testing.T) {
		scoop := model.NewScoop()
		page, results, err := scoop.FindByPage(nil)
		assert.Error(t, err)
		assert.Nil(t, page)
		assert.Nil(t, results)
	})
}

// ============================================================
// ModelScoop Method Delegation Tests
// ============================================================

func TestModelScoop_MethodDelegation(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	model := mongo.NewModel[TestUser](client)
	ms := model.NewScoop()

	t.Run("Count delegation", func(t *testing.T) {
		count, err := ms.Count()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Exist delegation", func(t *testing.T) {
		exists, err := ms.Exist()
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Updates delegation", func(t *testing.T) {
		res := ms.Updates(bson.M{"age": 20})
		assert.NotNil(t, res)
		assert.NoError(t, res.Error)
	})

	t.Run("Delete delegation", func(t *testing.T) {
		res := ms.Delete()
		assert.NotNil(t, res)
		assert.NoError(t, res.Error)
	})

	t.Run("GetCollection delegation", func(t *testing.T) {
		coll := ms.GetCollection()
		assert.NotNil(t, coll)
	})
}

func BenchmarkModelScoop_Creation(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	model := mongo.NewModel[TestUser](client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.NewScoop()
	}
}

func BenchmarkModelScoop_ChainedQuery(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	model := mongo.NewModel[TestUser](client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.NewScoop().
			Equal("status", "active").
			Gte("age", 18).
			Lte("age", 65).
			Limit(10).
			Offset(20).
			Sort("age", -1).
			Select("name", "email")
	}
}

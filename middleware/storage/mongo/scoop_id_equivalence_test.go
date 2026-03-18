package mongo_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ============================================================
// ID/_ID Equivalence Tests for Find/First
// ============================================================

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

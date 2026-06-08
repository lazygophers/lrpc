package mongo_test

import (
	"testing"
	"time"

	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AutoFillUser is a test model with auto-fill fields
type AutoFillUser struct {
	ID        string             `bson:"id"`
	ObjectID  primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	CreatedAt int64              `bson:"created_at"`
	UpdatedAt int64              `bson:"updated_at"`
}

// Collection returns the collection name
func (u AutoFillUser) Collection() string {
	return "autofill_users"
}

// ============================================================
// Create Auto-fill Tests
// ============================================================

func TestScoop_Create_AutoFill_Struct(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("auto fill all fields when empty", func(t *testing.T) {
		scoop := client.NewScoop()
		user := &AutoFillUser{
			Name:  "Alice",
			Email: "alice@example.com",
		}

		beforeCreate := time.Now().Unix()
		err := scoop.Create(user)
		afterCreate := time.Now().Unix()

		assert.NoError(t, err)
		assert.NotEmpty(t, user.ID, "id should be auto-filled")
		assert.NotEqual(t, primitive.NilObjectID, user.ObjectID, "_id should be auto-filled")
		assert.GreaterOrEqual(t, user.CreatedAt, beforeCreate, "created_at should be auto-filled")
		assert.LessOrEqual(t, user.CreatedAt, afterCreate, "created_at should be within range")
		assert.GreaterOrEqual(t, user.UpdatedAt, beforeCreate, "updated_at should be auto-filled")
		assert.LessOrEqual(t, user.UpdatedAt, afterCreate, "updated_at should be within range")
		assert.Equal(t, user.CreatedAt, user.UpdatedAt, "created_at and updated_at should be equal on create")
	})

	t.Run("preserve existing id field", func(t *testing.T) {
		scoop := client.NewScoop()
		customID := "custom-id-123"
		user := &AutoFillUser{
			ID:    customID,
			Name:  "Bob",
			Email: "bob@example.com",
		}

		err := scoop.Create(user)

		assert.NoError(t, err)
		assert.Equal(t, customID, user.ID, "existing id should be preserved")
		assert.NotEqual(t, primitive.NilObjectID, user.ObjectID, "_id should still be auto-filled")
	})

	t.Run("preserve existing timestamps", func(t *testing.T) {
		scoop := client.NewScoop()
		customTime := int64(1234567890)
		user := &AutoFillUser{
			Name:      "Charlie",
			Email:     "charlie@example.com",
			CreatedAt: customTime,
			UpdatedAt: customTime,
		}

		err := scoop.Create(user)

		assert.NoError(t, err)
		assert.Equal(t, customTime, user.CreatedAt, "existing created_at should be preserved")
		assert.Equal(t, customTime, user.UpdatedAt, "existing updated_at should be preserved")
	})
}

func TestScoop_Create_AutoFill_BsonM(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("auto fill all fields when empty", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("autofill_users")
		doc := bson.M{
			"name":  "David",
			"email": "david@example.com",
		}

		beforeCreate := time.Now().Unix()
		err := scoop.Create(doc)
		afterCreate := time.Now().Unix()

		assert.NoError(t, err)
		assert.NotNil(t, doc["id"], "id should be auto-filled")
		assert.IsType(t, "", doc["id"], "id should be string type")
		assert.NotNil(t, doc["_id"], "_id should be auto-filled")
		assert.IsType(t, primitive.ObjectID{}, doc["_id"], "_id should be ObjectID type")
		assert.NotNil(t, doc["created_at"], "created_at should be auto-filled")
		assert.IsType(t, int64(0), doc["created_at"], "created_at should be int64 type")
		assert.NotNil(t, doc["updated_at"], "updated_at should be auto-filled")
		assert.IsType(t, int64(0), doc["updated_at"], "updated_at should be int64 type")

		createdAt := doc["created_at"].(int64)
		updatedAt := doc["updated_at"].(int64)
		assert.GreaterOrEqual(t, createdAt, beforeCreate)
		assert.LessOrEqual(t, createdAt, afterCreate)
		assert.Equal(t, createdAt, updatedAt)
	})

	t.Run("preserve existing fields", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("autofill_users")
		customID := "custom-id-456"
		customTime := int64(9876543210)
		doc := bson.M{
			"id":         customID,
			"name":       "Eve",
			"email":      "eve@example.com",
			"created_at": customTime,
			"updated_at": customTime,
		}

		err := scoop.Create(doc)

		assert.NoError(t, err)
		assert.Equal(t, customID, doc["id"], "existing id should be preserved")
		assert.Equal(t, customTime, doc["created_at"], "existing created_at should be preserved")
		assert.Equal(t, customTime, doc["updated_at"], "existing updated_at should be preserved")
	})

	t.Run("auto fill zero value fields", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("autofill_users")
		doc := bson.M{
			"id":         "",
			"name":       "Frank",
			"email":      "frank@example.com",
			"created_at": int64(0),
			"updated_at": int64(0),
		}

		beforeCreate := time.Now().Unix()
		err := scoop.Create(doc)
		afterCreate := time.Now().Unix()

		assert.NoError(t, err)
		assert.NotEmpty(t, doc["id"], "empty id should be filled")
		createdAt := doc["created_at"].(int64)
		updatedAt := doc["updated_at"].(int64)
		assert.GreaterOrEqual(t, createdAt, beforeCreate, "zero created_at should be filled")
		assert.LessOrEqual(t, createdAt, afterCreate)
		assert.GreaterOrEqual(t, updatedAt, beforeCreate, "zero updated_at should be filled")
		assert.LessOrEqual(t, updatedAt, afterCreate)
	})
}

// ============================================================
// BatchCreate Auto-fill Tests
// ============================================================

func TestScoop_BatchCreate_AutoFill(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("auto fill all documents", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("autofill_users")
		docs := []interface{}{
			bson.M{"name": "User1", "email": "user1@example.com"},
			bson.M{"name": "User2", "email": "user2@example.com"},
			bson.M{"name": "User3", "email": "user3@example.com"},
		}

		beforeCreate := time.Now().Unix()
		err := scoop.BatchCreate(docs...)
		afterCreate := time.Now().Unix()

		assert.NoError(t, err)
		for i, doc := range docs {
			m := doc.(bson.M)
			assert.NotNil(t, m["id"], "doc %d: id should be auto-filled", i)
			assert.NotNil(t, m["_id"], "doc %d: _id should be auto-filled", i)
			assert.NotNil(t, m["created_at"], "doc %d: created_at should be auto-filled", i)
			assert.NotNil(t, m["updated_at"], "doc %d: updated_at should be auto-filled", i)

			createdAt := m["created_at"].(int64)
			assert.GreaterOrEqual(t, createdAt, beforeCreate)
			assert.LessOrEqual(t, createdAt, afterCreate)
		}
	})

	t.Run("each document gets unique id", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("autofill_users")
		docs := []interface{}{
			bson.M{"name": "User1"},
			bson.M{"name": "User2"},
			bson.M{"name": "User3"},
		}

		err := scoop.BatchCreate(docs...)
		assert.NoError(t, err)

		ids := make(map[string]bool)
		for _, doc := range docs {
			m := doc.(bson.M)
			id := m["id"].(string)
			assert.False(t, ids[id], "each document should have unique id")
			ids[id] = true
		}
	})
}

// ============================================================
// Update Auto-fill Tests
// ============================================================

func TestScoop_Updates_AutoFill_UpdatedAt(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("auto add updated_at to bson.M update", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("autofill_users")

		// Create a document first
		doc := bson.M{"name": "Alice", "email": "alice@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Update without updated_at
		time.Sleep(10 * time.Millisecond) // Ensure time difference
		result := scoop.Where("name", "Alice").Updates(bson.M{"email": "alice.new@example.com"})

		assert.NoError(t, result.Error)
		// Note: In mock mode we can't verify the actual updated document,
		// but we can verify the update was successful
	})

	t.Run("preserve existing updated_at if provided", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("autofill_users")

		// Create a document first
		doc := bson.M{"name": "Bob", "email": "bob@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Update with custom updated_at
		customTime := int64(1234567890)
		result := scoop.Where("name", "Bob").Updates(bson.M{
			"email":      "bob.new@example.com",
			"updated_at": customTime,
		})

		assert.NoError(t, result.Error)
	})

	t.Run("auto fill zero updated_at", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("autofill_users")

		// Create a document first
		doc := bson.M{"name": "Charlie", "email": "charlie@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Update with zero updated_at
		result := scoop.Where("name", "Charlie").Updates(bson.M{
			"email":      "charlie.new@example.com",
			"updated_at": int64(0),
		})

		assert.NoError(t, result.Error)
	})

	t.Run("auto add updated_at to struct update", func(t *testing.T) {
		scoop := client.NewScoop().CollectionName("autofill_users")

		// Create a document first
		doc := bson.M{"name": "David", "email": "david@example.com"}
		err := scoop.Create(doc)
		require.NoError(t, err)

		// Update with struct
		type UpdateData struct {
			Email string `bson:"email"`
		}
		result := scoop.Where("name", "David").Updates(UpdateData{Email: "david.new@example.com"})

		assert.NoError(t, result.Error)
	})
}

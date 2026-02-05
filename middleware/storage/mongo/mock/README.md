# MongoDB Mock Implementation

This package provides an in-memory implementation of MongoDB interfaces for testing purposes.

## Features

- ✅ Full CRUD operations (Insert, Find, Update, Delete)
- ✅ Aggregation pipeline support
- ✅ Bulk write operations
- ✅ Cursor iteration
- ✅ Collection management
- ⚠️ **Limited Session support** (see below)

## Session and Transaction Support

### Important Limitation

Due to `mongo.Session` interface having unexported methods, we **cannot fully implement** the Session interface outside the official mongo-driver package.

### Current Implementation

The mock provides basic Session method calls that return `ErrNotImplemented`:

```go
client := mock.NewMockClient()
_, err := client.StartSession()
// err == mock.ErrNotImplemented
```

### Alternatives for Testing Sessions

If you need to test code that uses Sessions and Transactions, you have these options:

1. **Use Real MongoDB with Docker** (Recommended)
   ```bash
   docker run -d -p 27017:27017 mongo:latest
   ```

2. **Use mockgen to generate mocks**
   ```bash
   mockgen -destination=mocks/mongo_session.go go.mongodb.org/mongo-driver/mongo Session
   ```

3. **Refactor code to use dependency injection** and mock at a higher level

### Why This Limitation Exists

The `mongo.Session` interface contains unexported methods (marked as "Has unexported methods" in godoc), which prevents external packages from implementing it. This is an intentional design decision by the MongoDB team to ensure session implementations conform to their internal requirements.

## Usage

### Basic Operations

```go
client := mock.NewMockClient()
db := client.Database("testdb")
coll := db.Collection("users")

// Insert
result, _ := coll.InsertOne(ctx, bson.M{"name": "Alice", "age": 30})

// Find
cursor, _ := coll.Find(ctx, bson.M{"age": bson.M{"$gte": 18}})

// Update
_, _ = coll.UpdateOne(ctx, bson.M{"name": "Alice"}, bson.M{"$set": bson.M{"age": 31}})

// Delete
_, _ = coll.DeleteOne(ctx, bson.M{"name": "Alice"})
```

### Testing Example

```go
func TestUserRepository(t *testing.T) {
    // Create mock client
    client := mock.NewMockClient()

    // Use in your code
    repo := NewUserRepository(client)

    // Test operations
    err := repo.CreateUser(context.Background(), "Alice", 30)
    assert.NoError(t, err)

    user, err := repo.GetUser(context.Background(), "Alice")
    assert.NoError(t, err)
    assert.Equal(t, 30, user.Age)
}
```

## Supported Operations

See individual files for detailed documentation:
- `mock_client.go` - Client operations
- `mock_database.go` - Database operations
- `mock_collection.go` - Collection operations
- `mock_cursor.go` - Cursor operations
- `aggregation.go` - Aggregation pipeline
- `memory_storage.go` - In-memory storage engine

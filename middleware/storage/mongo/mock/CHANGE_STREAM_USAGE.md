# Change Stream Usage Guide

This document explains how to use change streams with the MongoDB mock implementation.

## Overview

Change streams allow you to monitor real-time data changes in MongoDB collections. The mock implementation provides a simplified version that supports the four main operation types:

- **insert**: Document insertion
- **update**: Document updates
- **replace**: Document replacement
- **delete**: Document deletion

## Important Note

Due to limitations in the MongoDB driver (mongo.ChangeStream is a concrete type, not an interface), the standard `Watch()` method returns `ErrNotImplemented`. Instead, use the mock-specific `WatchMock()` methods for testing.

## API Methods

### Collection-Level Watch

Monitors changes to a specific collection:

```go
stream := mockCollection.WatchMock(ctx)
defer stream.Close(ctx)
```

### Database-Level Watch

Monitors changes to all collections (note: database isolation is not implemented in the current mock):

```go
stream := mockDatabase.WatchMock(ctx)
defer stream.Close(ctx)
```

### Client-Level Watch

Monitors changes across all collections:

```go
stream := mockClient.WatchMock(ctx)
defer stream.Close(ctx)
```

## Usage Examples

### Example 1: Watch Insert Operations

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
    client := mock.NewMockClient()
    mockClient := client.(*mock.MockClient)

    ctx := context.Background()
    db := mockClient.Database("testdb")
    coll := db.Collection("users")
    mockColl := coll.(*mock.MockCollection)

    // Create change stream
    stream := mockColl.WatchMock(ctx)
    defer stream.Close(ctx)

    // Start goroutine to insert documents
    go func() {
        time.Sleep(100 * time.Millisecond)
        doc := bson.M{
            "_id":  primitive.NewObjectID(),
            "name": "Alice",
            "age":  30,
        }
        mockColl.InsertOne(ctx, doc)
    }()

    // Wait for change event
    if stream.Next(ctx) {
        var event bson.M
        if err := stream.Decode(&event); err != nil {
            panic(err)
        }

        fmt.Printf("Operation: %s\n", event["operationType"])

        if fullDoc, ok := event["fullDocument"].(bson.M); ok {
            fmt.Printf("Document: %v\n", fullDoc)
        }
    }
}
```

### Example 2: Watch Update Operations

```go
// Insert initial document
docID := primitive.NewObjectID()
doc := bson.M{"_id": docID, "name": "Bob", "age": 25}
mockColl.InsertOne(ctx, doc)

// Create change stream
stream := mockColl.WatchMock(ctx)
defer stream.Close(ctx)

// Update document
go func() {
    time.Sleep(100 * time.Millisecond)
    filter := bson.M{"_id": docID}
    update := bson.M{"$set": bson.M{"age": 26}}
    mockColl.UpdateOne(ctx, filter, update)
}()

// Wait for change event
if stream.Next(ctx) {
    var event bson.M
    stream.Decode(&event)

    fmt.Printf("Operation: %s\n", event["operationType"])

    if updateDesc, ok := event["updateDescription"].(bson.M); ok {
        if updatedFields, ok := updateDesc["updatedFields"].(bson.M); ok {
            fmt.Printf("Updated fields: %v\n", updatedFields)
        }
    }
}
```

### Example 3: Watch Multiple Collections

```go
// Create client-level stream to watch all collections
stream := mockClient.WatchMock(ctx)
defer stream.Close(ctx)

db := mockClient.Database("testdb")

// Insert into different collections
go func() {
    usersCol := db.Collection("users").(*mock.MockCollection)
    usersCol.InsertOne(ctx, bson.M{"name": "User1"})

    time.Sleep(100 * time.Millisecond)

    productsCol := db.Collection("products").(*mock.MockCollection)
    productsCol.InsertOne(ctx, bson.M{"name": "Product1"})
}()

// Receive events from all collections
for i := 0; i < 2; i++ {
    if stream.Next(ctx) {
        var event bson.M
        stream.Decode(&event)

        ns := event["ns"].(bson.M)
        fmt.Printf("Collection: %s, Operation: %s\n",
            ns["coll"], event["operationType"])
    }
}
```

### Example 4: Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

stream := mockColl.WatchMock(ctx)
defer stream.Close(context.Background())

// This will timeout after 2 seconds if no events arrive
if stream.Next(ctx) {
    var event bson.M
    stream.Decode(&event)
    // Process event
} else {
    fmt.Println("No events received within timeout")
}
```

## Change Event Format

Change events follow the MongoDB change stream document format:

```go
{
    "_id": {
        "_data": "ObjectID hex string"
    },
    "operationType": "insert|update|replace|delete",
    "ns": {
        "db": "mock_database",
        "coll": "collection_name"
    },
    "clusterTime": Timestamp,
    "documentKey": {
        "_id": DocumentID
    },

    // For insert and replace operations:
    "fullDocument": {
        // Complete document
    },

    // For update operations:
    "updateDescription": {
        "updatedFields": {
            // Changed fields
        },
        "removedFields": [
            // Removed field names
        ]
    }
}
```

## Filtering

### Collection-Specific Filtering

By default, collection-level streams automatically filter events:

```go
// Only receives events for "users" collection
stream := mockColl.WatchMock(ctx)
```

### Operation Type Filtering

Currently, operation type filtering is not exposed in the public API, but you can filter manually:

```go
for stream.Next(ctx) {
    var event bson.M
    stream.Decode(&event)

    // Filter for inserts only
    if event["operationType"] == "insert" {
        // Process insert event
    }
}
```

## Best Practices

### 1. Always Close Streams

```go
stream := mockColl.WatchMock(ctx)
defer stream.Close(ctx)
```

### 2. Use Context with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

stream := mockColl.WatchMock(ctx)
```

### 3. Handle Closed Streams

```go
if !stream.Next(ctx) {
    // Stream closed or context cancelled
    if stream.Err() != nil {
        // Handle error
    }
    return
}
```

### 4. Process Events in Goroutines

```go
go func() {
    for stream.Next(ctx) {
        var event bson.M
        if err := stream.Decode(&event); err != nil {
            log.Printf("decode error: %v", err)
            continue
        }

        // Process event
        processEvent(event)
    }
}()
```

## Limitations

1. **No pipeline support**: The `pipeline` parameter in `Watch()` is currently ignored
2. **No resume tokens**: Resume token functionality is not implemented
3. **Buffer size**: Events are buffered (default 100 events). If the buffer is full, events may be dropped
4. **No database isolation**: Database-level watch monitors all collections globally
5. **No transaction support**: Change events are not transactional
6. **Concrete type limitation**: Cannot return `*mongo.ChangeStream` directly due to driver design

## Testing Patterns

### Pattern 1: Sync with Goroutines

```go
ready := make(chan struct{})

go func() {
    close(ready)
    // Perform operations
    mockColl.InsertOne(ctx, doc)
}()

<-ready
if stream.Next(ctx) {
    // Process event
}
```

### Pattern 2: Multiple Event Processing

```go
expectedEvents := 3
received := 0

for received < expectedEvents {
    if !stream.Next(ctx) {
        break
    }

    var event bson.M
    stream.Decode(&event)

    // Validate event
    received++
}

if received != expectedEvents {
    t.Errorf("expected %d events, got %d", expectedEvents, received)
}
```

### Pattern 3: Event Ordering

```go
events := make([]string, 0)

for i := 0; i < 3; i++ {
    if stream.Next(ctx) {
        var event bson.M
        stream.Decode(&event)
        events = append(events, event["operationType"].(string))
    }
}

expectedOrder := []string{"insert", "update", "delete"}
// Verify order
```

## Troubleshooting

### Problem: No Events Received

**Possible Causes:**
1. Stream was created after the operation
2. Context timeout too short
3. Event filtered out by collection filter
4. Buffer overflow (events dropped)

**Solution:**
```go
// Create stream BEFORE performing operations
stream := mockColl.WatchMock(ctx)

// Use longer timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

// Check if stream is closed
if stream.Next(ctx) {
    // Process event
} else {
    fmt.Println("Stream closed or timed out")
}
```

### Problem: Missing Update Details

**Cause:** Only `$set`, `$inc`, `$mul`, and `$unset` operators are tracked in update descriptions.

**Solution:**
```go
// Use supported update operators
update := bson.M{
    "$set": bson.M{"age": 26},
    "$unset": bson.M{"oldField": ""},
}
```

## Performance Considerations

1. **Buffer Size**: Default is 100 events. If you expect high throughput, events may be dropped
2. **Goroutine Management**: Each stream creates goroutines. Close streams when done
3. **Memory Usage**: Buffered events consume memory. Process events promptly
4. **Event Publishing**: All registered streams receive every event. Use collection-specific streams when possible

## Future Enhancements

Planned improvements for the change stream implementation:

- [ ] Pipeline filtering support
- [ ] Resume token implementation
- [ ] Configurable buffer size
- [ ] Database namespace isolation
- [ ] Operation type filtering in API
- [ ] Full document lookup option
- [ ] Pre/post-image support

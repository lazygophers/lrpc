package mock

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMockCollection_Aggregate_Match(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("users")

	// Insert test documents
	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice", "age": 30, "city": "NYC"},
		bson.M{"name": "Bob", "age": 25, "city": "LA"},
		bson.M{"name": "Charlie", "age": 35, "city": "NYC"},
		bson.M{"name": "David", "age": 28, "city": "SF"},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $match stage
	pipeline := []bson.M{
		{"$match": bson.M{"city": "NYC"}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestMockCollection_Aggregate_Project(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("users")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice", "age": 30, "email": "alice@example.com"},
		bson.M{"name": "Bob", "age": 25, "email": "bob@example.com"},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $project stage
	pipeline := []bson.M{
		{"$project": bson.M{"name": 1, "age": 1}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	// Check that email is not in results
	for _, result := range results {
		if _, hasEmail := result["email"]; hasEmail {
			t.Errorf("email should not be in projected results")
		}
		if _, hasName := result["name"]; !hasName {
			t.Errorf("name should be in projected results")
		}
		if _, hasAge := result["age"]; !hasAge {
			t.Errorf("age should be in projected results")
		}
	}
}

func TestMockCollection_Aggregate_Sort(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("users")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice", "age": 30},
		bson.M{"name": "Bob", "age": 25},
		bson.M{"name": "Charlie", "age": 35},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $sort stage - ascending
	pipeline := []bson.M{
		{"$sort": bson.M{"age": 1}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	// Check order
	if results[0]["name"] != "Bob" {
		t.Errorf("expected Bob first, got %v", results[0]["name"])
	}
	if results[2]["name"] != "Charlie" {
		t.Errorf("expected Charlie last, got %v", results[2]["name"])
	}
}

func TestMockCollection_Aggregate_Limit(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("users")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice"},
		bson.M{"name": "Bob"},
		bson.M{"name": "Charlie"},
		bson.M{"name": "David"},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $limit stage
	pipeline := []bson.M{
		{"$limit": 2},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestMockCollection_Aggregate_Skip(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("users")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice"},
		bson.M{"name": "Bob"},
		bson.M{"name": "Charlie"},
		bson.M{"name": "David"},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $skip stage
	pipeline := []bson.M{
		{"$skip": 2},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestMockCollection_Aggregate_Group_Sum(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("orders")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"customer": "Alice", "amount": 100},
		bson.M{"customer": "Bob", "amount": 200},
		bson.M{"customer": "Alice", "amount": 150},
		bson.M{"customer": "Bob", "amount": 50},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $group with $sum
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$customer",
				"total": bson.M{"$sum": "$amount"},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 groups, got %d", len(results))
	}

	// Check totals
	for _, result := range results {
		customer := result["_id"].(string)
		total := result["total"].(float64)

		if customer == "Alice" && total != 250 {
			t.Errorf("expected Alice total 250, got %v", total)
		}
		if customer == "Bob" && total != 250 {
			t.Errorf("expected Bob total 250, got %v", total)
		}
	}
}

func TestMockCollection_Aggregate_Group_Count(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("orders")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"customer": "Alice", "amount": 100},
		bson.M{"customer": "Bob", "amount": 200},
		bson.M{"customer": "Alice", "amount": 150},
		bson.M{"customer": "Bob", "amount": 50},
		bson.M{"customer": "Alice", "amount": 75},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $group with $sum: 1 for counting
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$customer",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 groups, got %d", len(results))
	}

	// Check counts
	for _, result := range results {
		customer := result["_id"].(string)
		count := result["count"].(float64)

		if customer == "Alice" && count != 3 {
			t.Errorf("expected Alice count 3, got %v", count)
		}
		if customer == "Bob" && count != 2 {
			t.Errorf("expected Bob count 2, got %v", count)
		}
	}
}

func TestMockCollection_Aggregate_Group_Avg(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("students")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"class": "A", "score": 90},
		bson.M{"class": "A", "score": 80},
		bson.M{"class": "B", "score": 70},
		bson.M{"class": "B", "score": 60},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $group with $avg
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":      "$class",
				"avgScore": bson.M{"$avg": "$score"},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 groups, got %d", len(results))
	}

	// Check averages
	for _, result := range results {
		class := result["_id"].(string)
		avgScore := result["avgScore"].(float64)

		if class == "A" && avgScore != 85 {
			t.Errorf("expected class A average 85, got %v", avgScore)
		}
		if class == "B" && avgScore != 65 {
			t.Errorf("expected class B average 65, got %v", avgScore)
		}
	}
}

func TestMockCollection_Aggregate_Group_MinMax(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("products")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"category": "A", "price": 100},
		bson.M{"category": "A", "price": 200},
		bson.M{"category": "B", "price": 50},
		bson.M{"category": "B", "price": 150},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $group with $min and $max
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":      "$category",
				"minPrice": bson.M{"$min": "$price"},
				"maxPrice": bson.M{"$max": "$price"},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 groups, got %d", len(results))
	}

	// Check min/max
	for _, result := range results {
		category := result["_id"].(string)
		minPrice := toFloat64(result["minPrice"])
		maxPrice := toFloat64(result["maxPrice"])

		if category == "A" {
			if minPrice != 100 {
				t.Errorf("expected category A min price 100, got %v", minPrice)
			}
			if maxPrice != 200 {
				t.Errorf("expected category A max price 200, got %v", maxPrice)
			}
		}
		if category == "B" {
			if minPrice != 50 {
				t.Errorf("expected category B min price 50, got %v", minPrice)
			}
			if maxPrice != 150 {
				t.Errorf("expected category B max price 150, got %v", maxPrice)
			}
		}
	}
}

func TestMockCollection_Aggregate_Group_FirstLast(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("events")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"user": "Alice", "event": "login"},
		bson.M{"user": "Bob", "event": "login"},
		bson.M{"user": "Alice", "event": "click"},
		bson.M{"user": "Bob", "event": "purchase"},
		bson.M{"user": "Alice", "event": "logout"},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $group with $first and $last
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":        "$user",
				"firstEvent": bson.M{"$first": "$event"},
				"lastEvent":  bson.M{"$last": "$event"},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 groups, got %d", len(results))
	}

	// Check first/last events
	for _, result := range results {
		user := result["_id"].(string)
		firstEvent := result["firstEvent"].(string)
		lastEvent := result["lastEvent"].(string)

		if user == "Alice" {
			if firstEvent != "login" {
				t.Errorf("expected Alice first event 'login', got %v", firstEvent)
			}
			if lastEvent != "logout" {
				t.Errorf("expected Alice last event 'logout', got %v", lastEvent)
			}
		}
		if user == "Bob" {
			if firstEvent != "login" {
				t.Errorf("expected Bob first event 'login', got %v", firstEvent)
			}
			if lastEvent != "purchase" {
				t.Errorf("expected Bob last event 'purchase', got %v", lastEvent)
			}
		}
	}
}

func TestMockCollection_Aggregate_Unwind(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("products")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{
			"name": "Product A",
			"tags": []string{"electronics", "sale", "new"},
		},
		bson.M{
			"name": "Product B",
			"tags": []string{"books", "sale"},
		},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $unwind stage
	pipeline := []bson.M{
		{"$unwind": "$tags"},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	// Should have 5 documents (3 from Product A, 2 from Product B)
	if len(results) != 5 {
		t.Errorf("expected 5 results after unwind, got %d", len(results))
	}

	// Check that tags are unwound
	for _, result := range results {
		tag := result["tags"]
		if tagStr, ok := tag.(string); !ok {
			t.Errorf("expected tag to be string, got %T: %v", tag, tag)
		} else {
			if tagStr != "electronics" && tagStr != "sale" && tagStr != "new" && tagStr != "books" {
				t.Errorf("unexpected tag value: %v", tagStr)
			}
		}
	}
}

func TestMockCollection_Aggregate_ComplexPipeline(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("sales")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"product": "A", "region": "East", "amount": 100},
		bson.M{"product": "B", "region": "East", "amount": 200},
		bson.M{"product": "A", "region": "West", "amount": 150},
		bson.M{"product": "B", "region": "West", "amount": 250},
		bson.M{"product": "A", "region": "East", "amount": 120},
		bson.M{"product": "C", "region": "East", "amount": 80},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Complex pipeline: filter, group, sort, limit
	pipeline := []bson.M{
		{"$match": bson.M{"region": "East"}},
		{
			"$group": bson.M{
				"_id":   "$product",
				"total": bson.M{"$sum": "$amount"},
			},
		},
		{"$sort": bson.M{"total": -1}},
		{"$limit": 2},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Check order (highest total first)
	if results[0]["_id"] != "A" {
		t.Errorf("expected first result to be product A, got %v", results[0]["_id"])
	}

	totalA := results[0]["total"].(float64)
	if totalA != 220 {
		t.Errorf("expected product A total 220, got %v", totalA)
	}
}

func TestMockCollection_Aggregate_EmptyPipeline(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("users")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"name": "Alice"},
		bson.M{"name": "Bob"},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Empty pipeline should return all documents
	pipeline := []bson.M{}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestMockCollection_Aggregate_GroupWithNullID(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("orders")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"amount": 100},
		bson.M{"amount": 200},
		bson.M{"amount": 150},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Group all documents into one group with null _id
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   nil,
				"total": bson.M{"$sum": "$amount"},
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 group, got %d", len(results))
	}

	total := results[0]["total"].(float64)
	count := results[0]["count"].(float64)

	if total != 450 {
		t.Errorf("expected total 450, got %v", total)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %v", count)
	}
}

func TestMockCollection_Aggregate_UnwindWithDocument(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("products")

	ctx := context.Background()
	_, err := coll.InsertOne(ctx, bson.M{
		"name":  "Product A",
		"items": []int{1, 2, 3},
	})
	if err != nil {
		t.Fatalf("failed to insert document: %v", err)
	}

	// Test $unwind with document format
	pipeline := []bson.M{
		{"$unwind": bson.M{"path": "$items"}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 results after unwind, got %d", len(results))
	}
}

func TestMockCollection_Aggregate_InvalidPipeline(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("users")

	ctx := context.Background()

	// Test invalid operator
	pipeline := []bson.M{
		{"$invalidOperator": bson.M{}},
	}

	_, err := coll.Aggregate(ctx, pipeline)
	if err == nil {
		t.Errorf("expected error for invalid operator")
	}
}

func TestMockCollection_Aggregate_GroupMultipleFields(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("sales")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"product": "A", "region": "East", "amount": 100},
		bson.M{"product": "A", "region": "East", "amount": 150},
		bson.M{"product": "A", "region": "West", "amount": 200},
		bson.M{"product": "B", "region": "East", "amount": 120},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Group by multiple fields
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"product": "$product",
					"region":  "$region",
				},
				"total": bson.M{"$sum": "$amount"},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 groups, got %d", len(results))
	}
}

func TestProcessAggregationPipeline_ObjectIDComparison(t *testing.T) {
	id1 := primitive.NewObjectID()
	id2 := primitive.NewObjectID()
	id3 := id1 // Same as id1

	docs := []bson.M{
		{"_id": id1, "value": 10},
		{"_id": id2, "value": 20},
		{"_id": id3, "value": 30},
	}

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": "$_id",
				"max": bson.M{"$max": "$value"},
			},
		},
	}

	results, err := processAggregationPipeline(docs, pipeline)
	if err != nil {
		t.Fatalf("processAggregationPipeline failed: %v", err)
	}

	// Should have 2 groups (id1/id3 are the same)
	if len(results) != 2 {
		t.Errorf("expected 2 groups, got %d", len(results))
	}
}

// ============================================================================
// Tests for processCount (0% coverage)
// ============================================================================

func TestProcessCount(t *testing.T) {
	client := NewMockClient()
	db := client.Database("testdb")
	coll := db.Collection("orders")

	ctx := context.Background()
	_, err := coll.InsertMany(ctx, []interface{}{
		bson.M{"customer": "Alice", "amount": 100},
		bson.M{"customer": "Bob", "amount": 200},
		bson.M{"customer": "Alice", "amount": 150},
		bson.M{"customer": "Bob", "amount": 50},
		bson.M{"customer": "Alice", "amount": 75},
	})
	if err != nil {
		t.Fatalf("failed to insert documents: %v", err)
	}

	// Test $group with $count accumulator
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$customer",
				"count": bson.M{"$count": bson.M{}},
			},
		},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("Aggregate failed: %v", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	err = cursor.All(ctx, &results)
	if err != nil {
		t.Fatalf("failed to decode results: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 groups, got %d", len(results))
	}

	// Check counts
	for _, result := range results {
		customer := result["_id"].(string)
		count := result["count"].(int64)

		if customer == "Alice" && count != 3 {
			t.Errorf("expected Alice count 3, got %v", count)
		}
		if customer == "Bob" && count != 2 {
			t.Errorf("expected Bob count 2, got %v", count)
		}
	}
}

// ============================================================================
// Tests for toNumericValue (25% coverage)
// ============================================================================

func TestToNumericValue_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{"nil", nil, 0, false},
		{"int", int(42), 42.0, true},
		{"int8", int8(42), 42.0, true},
		{"int16", int16(42), 42.0, true},
		{"int32", int32(42), 42.0, true},
		{"int64", int64(42), 42.0, true},
		{"uint", uint(42), 42.0, true},
		{"uint8", uint8(42), 42.0, true},
		{"uint16", uint16(42), 42.0, true},
		{"uint32", uint32(42), 42.0, true},
		{"uint64", uint64(42), 42.0, true},
		{"float32", float32(42.5), 42.5, true},
		{"float64", float64(42.5), 42.5, true},
		{"string", "not a number", 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toNumericValue(tt.input)
			if ok != tt.ok {
				t.Errorf("toNumericValue(%v) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("toNumericValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Tests for compareValues (27.3% coverage)
// ============================================================================

func TestCompareValues_AllBranches(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected int
	}{
		// Nil comparisons
		{"both nil", nil, nil, 0},
		{"a nil", nil, 100, -1},
		{"b nil", 100, nil, 1},

		// Numeric comparisons
		{"num a < b", 10, 20, -1},
		{"num a > b", 20, 10, 1},
		{"num a == b", 15, 15, 0},
		{"int vs float", int(10), float64(20.5), -1},
		{"uint vs int", uint(100), int(50), 1},

		// String comparisons
		{"str a < b", "apple", "banana", -1},
		{"str a > b", "zebra", "apple", 1},
		{"str a == b", "test", "test", 0},

		// Mixed types (should return 0)
		{"string vs number", "100", 100, 0},
		{"bool vs number", true, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareValues(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareValues(%v, %v) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCompareValues_ObjectID_Different(t *testing.T) {
	// Create two different ObjectIDs
	id1 := primitive.NewObjectIDFromTimestamp(time.Unix(1000, 0))
	id2 := primitive.NewObjectIDFromTimestamp(time.Unix(2000, 0))

	result := compareValues(id1, id2)
	if result >= 0 {
		t.Errorf("compareValues(id1, id2) should be < 0 for id1 < id2, got %d", result)
	}

	result = compareValues(id2, id1)
	if result <= 0 {
		t.Errorf("compareValues(id2, id1) should be > 0 for id2 > id1, got %d", result)
	}
}

func TestCompareValues_ObjectID_Equal(t *testing.T) {
	// Test with the same ObjectID instance
	id1 := primitive.NewObjectID()

	result := compareValues(id1, id1)
	if result != 0 {
		t.Errorf("compareValues(id1, id1) should be 0 for same instance, got %d", result)
	}
}

// ============================================================================
// Tests for processSkip (33.3% coverage)
// ============================================================================

func TestProcessSkip_AllBranches(t *testing.T) {
	tests := []struct {
		name        string
		docs        []bson.M
		skipValue   interface{}
		expected    int
		expectError bool
	}{
		{
			name:        "skip 0",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			skipValue:   0,
			expected:    3,
			expectError: false,
		},
		{
			name:        "skip as int",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			skipValue:   int(1),
			expected:    2,
			expectError: false,
		},
		{
			name:        "skip as int32",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			skipValue:   int32(1),
			expected:    2,
			expectError: false,
		},
		{
			name:        "skip as int64",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			skipValue:   int64(1),
			expected:    2,
			expectError: false,
		},
		{
			name:        "skip as float64",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			skipValue:   float64(1),
			expected:    2,
			expectError: false,
		},
		{
			name:        "skip >= len(docs)",
			docs:        []bson.M{{"a": 1}, {"a": 2}},
			skipValue:   5,
			expected:    0,
			expectError: false,
		},
		{
			name:        "skip negative",
			docs:        []bson.M{{"a": 1}},
			skipValue:   -1,
			expected:    0,
			expectError: true,
		},
		{
			name:        "skip invalid type",
			docs:        []bson.M{{"a": 1}},
			skipValue:   "invalid",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processSkip(tt.docs, tt.skipValue)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != tt.expected {
					t.Errorf("expected %d documents, got %d", tt.expected, len(result))
				}
			}
		})
	}
}

// ============================================================================
// Tests for processLimit (37.5% coverage)
// ============================================================================

func TestProcessLimit_AllBranches(t *testing.T) {
	tests := []struct {
		name        string
		docs        []bson.M
		limitValue  interface{}
		expected    int
		expectError bool
	}{
		{
			name:        "limit 0",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			limitValue:  0,
			expected:    3,
			expectError: false,
		},
		{
			name:        "limit as int",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			limitValue:  int(2),
			expected:    2,
			expectError: false,
		},
		{
			name:        "limit as int32",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			limitValue:  int32(2),
			expected:    2,
			expectError: false,
		},
		{
			name:        "limit as int64",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			limitValue:  int64(2),
			expected:    2,
			expectError: false,
		},
		{
			name:        "limit as float64",
			docs:        []bson.M{{"a": 1}, {"a": 2}, {"a": 3}},
			limitValue:  float64(2),
			expected:    2,
			expectError: false,
		},
		{
			name:        "limit >= len(docs)",
			docs:        []bson.M{{"a": 1}, {"a": 2}},
			limitValue:  10,
			expected:    2,
			expectError: false,
		},
		{
			name:        "limit negative",
			docs:        []bson.M{{"a": 1}},
			limitValue:  -1,
			expected:    0,
			expectError: true,
		},
		{
			name:        "limit invalid type",
			docs:        []bson.M{{"a": 1}},
			limitValue:  "invalid",
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processLimit(tt.docs, tt.limitValue)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != tt.expected {
					t.Errorf("expected %d documents, got %d", tt.expected, len(result))
				}
			}
		})
	}
}

// ============================================================================
// Tests for convertPipelineToBsonM (38.1% coverage)
// ============================================================================

func TestConvertPipelineToBsonM_AllBranches(t *testing.T) {
	tests := []struct {
		name        string
		pipeline    interface{}
		expected    int
		expectError bool
	}{
		{
			name:        "nil pipeline",
			pipeline:    nil,
			expected:    0,
			expectError: false,
		},
		{
			name: "already bson.M slice",
			pipeline: []bson.M{
				{"$match": bson.M{"status": "active"}},
				{"$limit": 10},
			},
			expected:    2,
			expectError: false,
		},
		{
			name: "interface slice with bson.M",
			pipeline: []interface{}{
				bson.M{"$match": bson.M{"status": "active"}},
				bson.M{"$limit": 10},
			},
			expected:    2,
			expectError: false,
		},
		{
			name:        "not a slice",
			pipeline:    bson.M{"$match": bson.M{"status": "active"}},
			expected:    0,
			expectError: true,
		},
		{
			name: "slice with non-bson.M convertible element",
			pipeline: []interface{}{
				bson.M{"$match": bson.M{"status": "active"}},
				"invalid_element",
			},
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := convertPipelineToBsonM(tt.pipeline)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != tt.expected {
					t.Errorf("expected %d stages, got %d", tt.expected, len(result))
				}
			}
		})
	}
}

// ============================================================================
// Tests for processAggregationPipeline edge cases
// ============================================================================

func TestProcessAggregationPipeline_EmptyStage(t *testing.T) {
	docs := []bson.M{{"a": 1}}
	pipeline := []bson.M{
		{}, // Empty stage
	}

	_, err := processAggregationPipeline(docs, pipeline)
	if err == nil {
		t.Errorf("expected error for empty stage")
	}
}

func TestProcessAggregationPipeline_MultipleOperators(t *testing.T) {
	docs := []bson.M{{"a": 1}}
	pipeline := []bson.M{
		{
			"$match": bson.M{"a": 1},
			"$limit": 10,
		},
	}

	_, err := processAggregationPipeline(docs, pipeline)
	if err == nil {
		t.Errorf("expected error for stage with multiple operators")
	}
}

// ============================================================================
// Tests for processMatch error case
// ============================================================================

func TestProcessMatch_InvalidValue(t *testing.T) {
	docs := []bson.M{{"a": 1}}
	_, err := processMatch(docs, "invalid")
	if err == nil {
		t.Errorf("expected error for non-bson.M match value")
	}
}

// ============================================================================
// Tests for processProject error case
// ============================================================================

func TestProcessProject_InvalidValue(t *testing.T) {
	docs := []bson.M{{"a": 1}}
	_, err := processProject(docs, "invalid")
	if err == nil {
		t.Errorf("expected error for non-bson.M project value")
	}
}

// ============================================================================
// Tests for processSort error case
// ============================================================================

func TestProcessSort_InvalidValue(t *testing.T) {
	docs := []bson.M{{"a": 1}}
	_, err := processSort(docs, "invalid")
	if err == nil {
		t.Errorf("expected error for non-bson.M sort value")
	}
}

// ============================================================================
// Tests for processGroup error cases
// ============================================================================

func TestProcessGroup_InvalidValue(t *testing.T) {
	docs := []bson.M{{"a": 1}}
	_, err := processGroup(docs, "invalid")
	if err == nil {
		t.Errorf("expected error for non-bson.M group value")
	}
}

func TestProcessGroup_MissingID(t *testing.T) {
	docs := []bson.M{{"a": 1}}
	_, err := processGroup(docs, bson.M{"total": bson.M{"$sum": "$a"}})
	if err == nil {
		t.Errorf("expected error for missing _id field")
	}
}

func TestProcessGroup_InvalidAccumulator(t *testing.T) {
	docs := []bson.M{{"a": 1}}
	_, err := processGroup(docs, bson.M{
		"_id":   "$a",
		"total": "invalid", // Not a bson.M
	})
	if err == nil {
		t.Errorf("expected error for invalid accumulator")
	}
}

func TestProcessGroup_UnsupportedAccumulator(t *testing.T) {
	docs := []bson.M{{"a": 1}}
	_, err := processGroup(docs, bson.M{
		"_id":   "$a",
		"total": bson.M{"$unsupported": "$a"},
	})
	if err == nil {
		t.Errorf("expected error for unsupported accumulator")
	}
}

// ============================================================================
// Tests for accumulator error cases
// ============================================================================

func TestProcessSum_InvalidValue(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"value": 100}

	err := processSum(state, doc, stat, []int{1, 2, 3})
	if err == nil {
		t.Errorf("expected error for invalid sum value")
	}
}

func TestProcessAvg_InvalidValue(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"value": 100}

	err := processAvg(state, doc, stat, 123)
	if err == nil {
		t.Errorf("expected error for non-field-reference avg value")
	}
}

func TestProcessMin_InvalidValue(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"value": 100}

	err := processMin(state, doc, stat, 123)
	if err == nil {
		t.Errorf("expected error for non-field-reference min value")
	}
}

func TestProcessMax_InvalidValue(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"value": 100}

	err := processMax(state, doc, stat, 123)
	if err == nil {
		t.Errorf("expected error for non-field-reference max value")
	}
}

func TestProcessFirst_InvalidValue(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"value": 100}

	err := processFirst(state, doc, stat, 123)
	if err == nil {
		t.Errorf("expected error for non-field-reference first value")
	}
}

func TestProcessLast_InvalidValue(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"value": 100}

	err := processLast(state, doc, stat, 123)
	if err == nil {
		t.Errorf("expected error for non-field-reference last value")
	}
}

// ============================================================================
// Tests for processUnwind edge cases
// ============================================================================

func TestProcessUnwind_InvalidValue(t *testing.T) {
	docs := []bson.M{{"items": []int{1, 2, 3}}}
	_, err := processUnwind(docs, 123)
	if err == nil {
		t.Errorf("expected error for invalid unwind value")
	}
}

func TestProcessUnwind_InvalidFieldPath(t *testing.T) {
	docs := []bson.M{{"items": []int{1, 2, 3}}}
	_, err := processUnwind(docs, "items") // Missing $
	if err == nil {
		t.Errorf("expected error for field path without $")
	}
}

func TestProcessUnwind_DocumentWithoutPath(t *testing.T) {
	docs := []bson.M{{"items": []int{1, 2, 3}}}
	_, err := processUnwind(docs, bson.M{"preserveNullAndEmptyArrays": true})
	if err == nil {
		t.Errorf("expected error for unwind document without path")
	}
}

func TestProcessUnwind_DocumentWithInvalidPath(t *testing.T) {
	docs := []bson.M{{"items": []int{1, 2, 3}}}
	_, err := processUnwind(docs, bson.M{"path": 123})
	if err == nil {
		t.Errorf("expected error for non-string path")
	}
}

func TestProcessUnwind_EmptyPath(t *testing.T) {
	docs := []bson.M{{"items": []int{1, 2, 3}}}
	_, err := processUnwind(docs, "")
	if err == nil {
		t.Errorf("expected error for empty path")
	}
}

func TestProcessUnwind_NonArrayValue(t *testing.T) {
	docs := []bson.M{{"value": "not an array"}}
	result, err := processUnwind(docs, "$value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 document for non-array value, got %d", len(result))
	}
}

func TestProcessUnwind_MissingField(t *testing.T) {
	docs := []bson.M{{"other": "value"}}
	result, err := processUnwind(docs, "$missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 documents for missing field, got %d", len(result))
	}
}

func TestProcessUnwind_EmptyArray(t *testing.T) {
	docs := []bson.M{{"items": []int{}}}
	result, err := processUnwind(docs, "$items")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 documents for empty array, got %d", len(result))
	}
}

// ============================================================================
// Tests for evaluateGroupKey with non-existent field
// ============================================================================

func TestEvaluateGroupKey_NonExistentField(t *testing.T) {
	doc := bson.M{"existing": "value"}
	key, err := evaluateGroupKey(doc, "$nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != nil {
		t.Errorf("expected nil for non-existent field, got %v", key)
	}
}

func TestEvaluateGroupKey_LiteralValue(t *testing.T) {
	doc := bson.M{"field": "value"}
	key, err := evaluateGroupKey(doc, "literal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "literal" {
		t.Errorf("expected 'literal', got %v", key)
	}
}

func TestEvaluateGroupKey_ComplexWithNonExistentField(t *testing.T) {
	doc := bson.M{"existing": "value"}
	key, err := evaluateGroupKey(doc, bson.M{
		"field1": "$existing",
		"field2": "$nonexistent",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resultMap := key.(bson.M)
	if resultMap["field1"] != "value" {
		t.Errorf("expected 'value' for field1, got %v", resultMap["field1"])
	}
	if resultMap["field2"] != nil {
		t.Errorf("expected nil for field2, got %v", resultMap["field2"])
	}
}

// ============================================================================
// Tests for formatGroupKey error case
// ============================================================================

func TestFormatGroupKey_Nil(t *testing.T) {
	key := formatGroupKey(nil)
	if key != "<null>" {
		t.Errorf("expected '<null>' for nil key, got %s", key)
	}
}

// ============================================================================
// Tests for processMin/Max with non-existent fields
// ============================================================================

func TestProcessMin_NonExistentField(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"other": 100}

	err := processMin(state, doc, stat, "$nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if stat.result != nil {
		t.Errorf("expected nil result for non-existent field, got %v", stat.result)
	}
}

func TestProcessMax_NonExistentField(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"other": 100}

	err := processMax(state, doc, stat, "$nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if stat.result != nil {
		t.Errorf("expected nil result for non-existent field, got %v", stat.result)
	}
}

// ============================================================================
// Tests for processFirst/Last with non-existent fields
// ============================================================================

func TestProcessFirst_NonExistentField(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"other": 100}

	err := processFirst(state, doc, stat, "$nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if stat.result != nil {
		t.Errorf("expected nil result for non-existent field, got %v", stat.result)
	}
}

func TestProcessLast_NonExistentField(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"other": 100}

	err := processLast(state, doc, stat, "$nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if stat.result != nil {
		t.Errorf("expected nil result for non-existent field, got %v", stat.result)
	}
}

// ============================================================================
// Tests for processAvg with non-existent field
// ============================================================================

func TestProcessAvg_NonExistentField(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"other": 100}

	err := processAvg(state, doc, stat, "$nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if stat.count != 0 {
		t.Errorf("expected count 0 for non-existent field, got %d", stat.count)
	}
}

// ============================================================================
// Tests for processSum with non-existent field
// ============================================================================

func TestProcessSum_NonExistentField(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{}
	doc := bson.M{"other": 100}

	err := processSum(state, doc, stat, "$nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if stat.sum != 0 {
		t.Errorf("expected sum 0 for non-existent field, got %f", stat.sum)
	}
}

// ============================================================================
// Additional tests to reach 100% coverage
// ============================================================================

func TestProcessMin_CompareUpdate(t *testing.T) {
	state := &groupState{fieldStats: make(map[string]*fieldStat)}
	stat := &fieldStat{result: 100}
	doc := bson.M{"value": 50}

	err := processMin(state, doc, stat, "$value")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if toFloat64(stat.result) != 50 {
		t.Errorf("expected min 50, got %v", stat.result)
	}
}

func TestProcessGroup_EvaluateGroupKeyError(t *testing.T) {
	// This test is to ensure evaluateGroupKey is called correctly
	docs := []bson.M{
		{"product": "A", "value": 100},
		{"product": "B", "value": 200},
	}

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": bson.M{
					"product": "$product",
					"status":  "active",
				},
				"total": bson.M{"$sum": "$value"},
			},
		},
	}

	result, err := processAggregationPipeline(docs, pipeline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 groups, got %d", len(result))
	}
}

func TestConvertPipelineToBsonM_WithBsonD(t *testing.T) {
	// Test with bson.D elements
	pipeline := []interface{}{
		bson.D{{"$match", bson.M{"status": "active"}}},
	}

	result, err := convertPipelineToBsonM(pipeline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 stage, got %d", len(result))
	}
}

func TestProcessAccumulator_EmptyDoc(t *testing.T) {
	// Test with complex accumulator scenario
	state := &groupState{
		fieldStats: make(map[string]*fieldStat),
		docs:       []bson.M{},
	}
	doc := bson.M{"value": 100}

	err := processAccumulator(state, doc, "total", bson.M{"$sum": "$value"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEvaluateGroupKey_ComplexNested(t *testing.T) {
	doc := bson.M{
		"user": bson.M{
			"name": "Alice",
			"age":  30,
		},
		"status": "active",
	}

	// Test with literal value in complex grouping
	key, err := evaluateGroupKey(doc, bson.M{
		"userStatus": "active",
		"userName":   "$user.name",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resultMap := key.(bson.M)
	if resultMap["userStatus"] != "active" {
		t.Errorf("expected 'active', got %v", resultMap["userStatus"])
	}
}

func TestProcessGroup_FormatGroupKeyPath(t *testing.T) {
	// Test to ensure formatGroupKey is called with various key types
	docs := []bson.M{
		{"category": bson.M{"type": "A", "sub": "X"}, "value": 100},
		{"category": bson.M{"type": "A", "sub": "Y"}, "value": 200},
		{"category": bson.M{"type": "B", "sub": "X"}, "value": 150},
	}

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$category",
				"total": bson.M{"$sum": "$value"},
			},
		},
	}

	result, err := processAggregationPipeline(docs, pipeline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 groups, got %d", len(result))
	}
}

func TestProcessAccumulator_EmptyAccDoc(t *testing.T) {
	// Test with empty accumulator document to cover the return nil branch
	state := &groupState{
		fieldStats: make(map[string]*fieldStat),
		docs:       []bson.M{},
	}
	doc := bson.M{"value": 100}

	// Empty bson.M should hit the return nil at the end
	err := processAccumulator(state, doc, "empty", bson.M{})
	if err != nil {
		t.Errorf("unexpected error for empty accumulator: %v", err)
	}
}

func TestFormatGroupKey_WithComplexTypes(t *testing.T) {
	// Test formatGroupKey with various types
	tests := []struct {
		name string
		key  interface{}
	}{
		{"nil key", nil},
		{"string key", "test"},
		{"int key", 123},
		{"map key", bson.M{"a": 1, "b": 2}},
		{"array key", []int{1, 2, 3}},
		{"objectid key", primitive.NewObjectID()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatGroupKey(tt.key)
			if tt.key == nil {
				if result != "<null>" {
					t.Errorf("expected '<null>' for nil, got %s", result)
				}
			} else {
				if result == "" {
					t.Errorf("expected non-empty result for %v", tt.key)
				}
			}
		})
	}
}

func TestProcessGroup_StatAlreadyExists(t *testing.T) {
	// Test that processAccumulator correctly handles existing field stats
	docs := []bson.M{
		{"category": "A", "value": 100},
		{"category": "A", "value": 200},
	}

	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$category",
				"total": bson.M{"$sum": "$value"},
				"count": bson.M{"$sum": 1},
			},
		},
	}

	result, err := processAggregationPipeline(docs, pipeline)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 group, got %d", len(result))
	}
	if toFloat64(result[0]["total"]) != 300 {
		t.Errorf("expected total 300, got %v", result[0]["total"])
	}
}

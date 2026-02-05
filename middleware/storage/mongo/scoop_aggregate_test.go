package mongo_test

import (
	"context"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AggProduct is a test model for aggregation tests
type AggProduct struct {
	Name     string  `bson:"name"`
	Price    float64 `bson:"price"`
	Quantity int     `bson:"quantity"`
	Category string  `bson:"category"`
}

func (ap AggProduct) Collection() string {
	return "agg_products"
}

// setupAggTestData is a helper function to setup test data
func setupAggTestData(t *testing.T, client *mongo.Client) {
	ctx := context.Background()
	scoop := client.NewScoop().Collection(AggProduct{})

	// Clear collection
	_, err := scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})
	require.NoError(t, err)

	// Insert test data
	products := []interface{}{
		map[string]interface{}{"name": "Product1", "price": 10.0, "quantity": 2, "category": "A"},
		map[string]interface{}{"name": "Product2", "price": 20.0, "quantity": 3, "category": "A"},
		map[string]interface{}{"name": "Product3", "price": 30.0, "quantity": 5, "category": "B"},
	}
	_, err = scoop.GetCollection().InsertMany(ctx, products)
	require.NoError(t, err)
}

// ============================================================
// Sum Tests
// ============================================================

func TestScoop_SumBasic(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Test sum of price
	scoop := client.NewScoop().Collection(AggProduct{})
	sum, err := scoop.Sum("price")
	assert.NoError(t, err)
	assert.Equal(t, 60.0, sum)
}

func TestScoop_SumWithFilter(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Sum with category filter
	scoop := client.NewScoop().Collection(AggProduct{}).Equal("category", "A")
	sum, err := scoop.Sum("price")
	assert.NoError(t, err)
	assert.Equal(t, 30.0, sum)
}

func TestScoop_SumInteger(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Sum integer field
	scoop := client.NewScoop().Collection(AggProduct{})
	sum, err := scoop.Sum("quantity")
	assert.NoError(t, err)
	assert.Equal(t, 10.0, sum)
}

func TestScoop_SumEmpty(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	// Clear collection
	scoop := client.NewScoop().Collection(AggProduct{})
	_, err = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})
	require.NoError(t, err)

	sum, err := scoop.Clone().Sum("price")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, sum)
}

func TestScoop_SumNoCollection(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	_, err = scoop.Sum("price")
	assert.Error(t, err)
}

// ============================================================
// Avg Tests
// ============================================================

func TestScoop_AvgBasic(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Test average of price
	scoop := client.NewScoop().Collection(AggProduct{})
	avg, err := scoop.Avg("price")
	assert.NoError(t, err)
	assert.Equal(t, 20.0, avg)
}

func TestScoop_AvgWithFilter(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Average with category filter
	scoop := client.NewScoop().Collection(AggProduct{}).Equal("category", "A")
	avg, err := scoop.Avg("price")
	assert.NoError(t, err)
	assert.Equal(t, 15.0, avg)
}

func TestScoop_AvgInteger(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Average integer field
	scoop := client.NewScoop().Collection(AggProduct{})
	avg, err := scoop.Avg("quantity")
	assert.NoError(t, err)
	// (2+3+5)/3 = 3.333...
	assert.InDelta(t, 3.333, avg, 0.01)
}

func TestScoop_AvgEmpty(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	// Clear collection
	scoop := client.NewScoop().Collection(AggProduct{})
	_, err = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})
	require.NoError(t, err)

	avg, err := scoop.Clone().Avg("price")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, avg)
}

func TestScoop_AvgNoCollection(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	_, err = scoop.Avg("price")
	assert.Error(t, err)
}

// ============================================================
// Max Tests
// ============================================================

func TestScoop_MaxBasic(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Test max of price
	scoop := client.NewScoop().Collection(AggProduct{})
	max, err := scoop.Max("price")
	assert.NoError(t, err)
	assert.Equal(t, 30.0, max)
}

func TestScoop_MaxWithFilter(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Max with category filter
	scoop := client.NewScoop().Collection(AggProduct{}).Equal("category", "A")
	max, err := scoop.Max("price")
	assert.NoError(t, err)
	assert.Equal(t, 20.0, max)
}

func TestScoop_MaxInteger(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Max integer field
	scoop := client.NewScoop().Collection(AggProduct{})
	max, err := scoop.Max("quantity")
	assert.NoError(t, err)
	assert.Equal(t, 5.0, max)
}

func TestScoop_MaxEmpty(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	// Clear collection
	scoop := client.NewScoop().Collection(AggProduct{})
	_, err = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})
	require.NoError(t, err)

	max, err := scoop.Clone().Max("price")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, max)
}

func TestScoop_MaxNegative(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	// Setup negative data
	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": -10.0, "quantity": 1},
		map[string]interface{}{"name": "P2", "price": -5.0, "quantity": 2},
		map[string]interface{}{"name": "P3", "price": -20.0, "quantity": 3},
	}
	_, err = scoop.GetCollection().InsertMany(ctx, products)
	require.NoError(t, err)

	// Max should be -5.0
	max, err := scoop.Clone().Max("price")
	assert.NoError(t, err)
	assert.Equal(t, -5.0, max)
}

func TestScoop_MaxNoCollection(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	_, err = scoop.Max("price")
	assert.Error(t, err)
}

// ============================================================
// Min Tests
// ============================================================

func TestScoop_MinBasic(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Test min of price
	scoop := client.NewScoop().Collection(AggProduct{})
	min, err := scoop.Min("price")
	assert.NoError(t, err)
	assert.Equal(t, 10.0, min)
}

func TestScoop_MinWithFilter(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Min with category filter
	scoop := client.NewScoop().Collection(AggProduct{}).Equal("category", "A")
	min, err := scoop.Min("price")
	assert.NoError(t, err)
	assert.Equal(t, 10.0, min)
}

func TestScoop_MinInteger(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	// Min integer field
	scoop := client.NewScoop().Collection(AggProduct{})
	min, err := scoop.Min("quantity")
	assert.NoError(t, err)
	assert.Equal(t, 2.0, min)
}

func TestScoop_MinEmpty(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	// Clear collection
	scoop := client.NewScoop().Collection(AggProduct{})
	_, err = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})
	require.NoError(t, err)

	min, err := scoop.Clone().Min("price")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, min)
}

func TestScoop_MinNegative(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	// Setup negative data
	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": -10.0, "quantity": 1},
		map[string]interface{}{"name": "P2", "price": -5.0, "quantity": 2},
		map[string]interface{}{"name": "P3", "price": -20.0, "quantity": 3},
	}
	_, err = scoop.GetCollection().InsertMany(ctx, products)
	require.NoError(t, err)

	// Min should be -20.0
	min, err := scoop.Clone().Min("price")
	assert.NoError(t, err)
	assert.Equal(t, -20.0, min)
}

func TestScoop_MinNoCollection(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	scoop := client.NewScoop()
	_, err = scoop.Min("price")
	assert.Error(t, err)
}

// ============================================================
// Edge Cases
// ============================================================

func TestScoop_AggregateZeroValues(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	// Insert zero values
	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": 0.0, "quantity": 0},
		map[string]interface{}{"name": "P2", "price": 0.0, "quantity": 0},
	}
	_, err = scoop.GetCollection().InsertMany(ctx, products)
	require.NoError(t, err)

	// All aggregations should return 0
	sum, _ := scoop.Clone().Sum("price")
	assert.Equal(t, 0.0, sum)

	avg, _ := scoop.Clone().Avg("price")
	assert.Equal(t, 0.0, avg)

	max, _ := scoop.Clone().Max("price")
	assert.Equal(t, 0.0, max)

	min, _ := scoop.Clone().Min("price")
	assert.Equal(t, 0.0, min)
}

func TestScoop_AggregateSingleDoc(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	// Insert single document
	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": 42.5, "quantity": 7},
	}
	_, err = scoop.GetCollection().InsertMany(ctx, products)
	require.NoError(t, err)

	// All aggregations should return the same value
	sum, _ := scoop.Clone().Sum("price")
	assert.Equal(t, 42.5, sum)

	avg, _ := scoop.Clone().Avg("price")
	assert.Equal(t, 42.5, avg)

	max, _ := scoop.Clone().Max("price")
	assert.Equal(t, 42.5, max)

	min, _ := scoop.Clone().Min("price")
	assert.Equal(t, 42.5, min)
}

func TestScoop_AggregateMultipleFilters(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	// Insert test data
	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": 10.0, "quantity": 5, "category": "A"},
		map[string]interface{}{"name": "P2", "price": 20.0, "quantity": 10, "category": "A"},
		map[string]interface{}{"name": "P3", "price": 30.0, "quantity": 15, "category": "B"},
		map[string]interface{}{"name": "P4", "price": 40.0, "quantity": 20, "category": "B"},
	}
	_, err = scoop.GetCollection().InsertMany(ctx, products)
	require.NoError(t, err)

	// Aggregate with multiple filters: category="B" AND quantity>=15
	sum, err := scoop.Clone().
		Equal("category", "B").
		Gte("quantity", 15).
		Sum("price")
	assert.NoError(t, err)
	assert.Equal(t, 70.0, sum) // 30.0 + 40.0
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkScoop_Sum(b *testing.B) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, _ := mongo.New(cfg)
	ctx := context.Background()

	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": 10.0, "quantity": 2},
		map[string]interface{}{"name": "P2", "price": 20.0, "quantity": 3},
		map[string]interface{}{"name": "P3", "price": 30.0, "quantity": 4},
	}
	_, _ = scoop.GetCollection().InsertMany(ctx, products)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scoop.Clone().Sum("price")
	}
}

func BenchmarkScoop_Avg(b *testing.B) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, _ := mongo.New(cfg)
	ctx := context.Background()

	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": 10.0, "quantity": 2},
		map[string]interface{}{"name": "P2", "price": 20.0, "quantity": 3},
		map[string]interface{}{"name": "P3", "price": 30.0, "quantity": 4},
	}
	_, _ = scoop.GetCollection().InsertMany(ctx, products)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scoop.Clone().Avg("price")
	}
}

func BenchmarkScoop_Max(b *testing.B) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, _ := mongo.New(cfg)
	ctx := context.Background()

	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": 10.0, "quantity": 2},
		map[string]interface{}{"name": "P2", "price": 20.0, "quantity": 3},
		map[string]interface{}{"name": "P3", "price": 30.0, "quantity": 4},
	}
	_, _ = scoop.GetCollection().InsertMany(ctx, products)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scoop.Clone().Max("price")
	}
}

func BenchmarkScoop_Min(b *testing.B) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, _ := mongo.New(cfg)
	ctx := context.Background()

	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": 10.0, "quantity": 2},
		map[string]interface{}{"name": "P2", "price": 20.0, "quantity": 3},
		map[string]interface{}{"name": "P3", "price": 30.0, "quantity": 4},
	}
	_, _ = scoop.GetCollection().InsertMany(ctx, products)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scoop.Clone().Min("price")
	}
}

func BenchmarkScoop_AggregateWithFilter(b *testing.B) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, _ := mongo.New(cfg)
	ctx := context.Background()

	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	products := []interface{}{
		map[string]interface{}{"name": "P1", "price": 10.0, "quantity": 2, "category": "A"},
		map[string]interface{}{"name": "P2", "price": 20.0, "quantity": 3, "category": "A"},
		map[string]interface{}{"name": "P3", "price": 30.0, "quantity": 4, "category": "B"},
	}
	_, _ = scoop.GetCollection().InsertMany(ctx, products)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scoop.Clone().Equal("category", "A").Sum("price")
	}
}

// ============================================================
// Type Conversion Tests
// ============================================================

func TestScoop_TypeConversions(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("sum int values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": 10, "quantity": 2},
			map[string]interface{}{"name": "P2", "price": 20, "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		sum, err := scoop.Clone().Sum("price")
		assert.NoError(t, err)
		assert.Equal(t, 30.0, sum)
	})

	t.Run("sum int32 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": int32(10), "quantity": int32(2)},
			map[string]interface{}{"name": "P2", "price": int32(20), "quantity": int32(3)},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		sum, err := scoop.Clone().Sum("price")
		assert.NoError(t, err)
		assert.Equal(t, 30.0, sum)
	})

	t.Run("sum int64 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": int64(10), "quantity": int64(2)},
			map[string]interface{}{"name": "P2", "price": int64(20), "quantity": int64(3)},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		sum, err := scoop.Clone().Sum("price")
		assert.NoError(t, err)
		assert.Equal(t, 30.0, sum)
	})

	t.Run("sum float32 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": float32(10.5), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": float32(20.5), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		sum, err := scoop.Clone().Sum("price")
		assert.NoError(t, err)
		assert.InDelta(t, 31.0, sum, 0.1)
	})

	t.Run("avg int values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": 10, "quantity": 2},
			map[string]interface{}{"name": "P2", "price": 20, "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		avg, err := scoop.Clone().Avg("price")
		assert.NoError(t, err)
		assert.Equal(t, 15.0, avg)
	})

	t.Run("avg int32 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": int32(10), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": int32(20), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		avg, err := scoop.Clone().Avg("price")
		assert.NoError(t, err)
		assert.Equal(t, 15.0, avg)
	})

	t.Run("avg int64 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": int64(10), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": int64(20), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		avg, err := scoop.Clone().Avg("price")
		assert.NoError(t, err)
		assert.Equal(t, 15.0, avg)
	})

	t.Run("avg float32 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": float32(10.5), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": float32(20.5), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		avg, err := scoop.Clone().Avg("price")
		assert.NoError(t, err)
		assert.InDelta(t, 15.5, avg, 0.1)
	})

	t.Run("max int values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": 10, "quantity": 2},
			map[string]interface{}{"name": "P2", "price": 20, "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		max, err := scoop.Clone().Max("price")
		assert.NoError(t, err)
		assert.Equal(t, 20.0, max)
	})

	t.Run("max int32 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": int32(10), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": int32(20), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		max, err := scoop.Clone().Max("price")
		assert.NoError(t, err)
		assert.Equal(t, 20.0, max)
	})

	t.Run("max int64 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": int64(10), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": int64(20), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		max, err := scoop.Clone().Max("price")
		assert.NoError(t, err)
		assert.Equal(t, 20.0, max)
	})

	t.Run("max float32 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": float32(10.5), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": float32(20.5), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		max, err := scoop.Clone().Max("price")
		assert.NoError(t, err)
		assert.InDelta(t, 20.5, max, 0.1)
	})

	t.Run("min int values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": 10, "quantity": 2},
			map[string]interface{}{"name": "P2", "price": 20, "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		min, err := scoop.Clone().Min("price")
		assert.NoError(t, err)
		assert.Equal(t, 10.0, min)
	})

	t.Run("min int32 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": int32(10), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": int32(20), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		min, err := scoop.Clone().Min("price")
		assert.NoError(t, err)
		assert.Equal(t, 10.0, min)
	})

	t.Run("min int64 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": int64(10), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": int64(20), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		min, err := scoop.Clone().Min("price")
		assert.NoError(t, err)
		assert.Equal(t, 10.0, min)
	})

	t.Run("min float32 values", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": float32(10.5), "quantity": 2},
			map[string]interface{}{"name": "P2", "price": float32(20.5), "quantity": 3},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		min, err := scoop.Clone().Min("price")
		assert.NoError(t, err)
		assert.InDelta(t, 10.5, min, 0.1)
	})
}

func TestScoop_InvalidFieldNames(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)

	setupAggTestData(t, client)

	t.Run("sum nonexistent field", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		sum, err := scoop.Sum("nonexistent_field")
		// Should not error but return 0
		assert.NoError(t, err)
		assert.Equal(t, 0.0, sum)
	})

	t.Run("avg nonexistent field", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		avg, err := scoop.Avg("nonexistent_field")
		assert.NoError(t, err)
		assert.Equal(t, 0.0, avg)
	})

	t.Run("max nonexistent field", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		max, err := scoop.Max("nonexistent_field")
		assert.NoError(t, err)
		assert.Equal(t, 0.0, max)
	})

	t.Run("min nonexistent field", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		min, err := scoop.Min("nonexistent_field")
		assert.NoError(t, err)
		assert.Equal(t, 0.0, min)
	})
}

func TestScoop_ComplexFilters(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("range filter", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": 5.0, "quantity": 1},
			map[string]interface{}{"name": "P2", "price": 15.0, "quantity": 2},
			map[string]interface{}{"name": "P3", "price": 25.0, "quantity": 3},
			map[string]interface{}{"name": "P4", "price": 35.0, "quantity": 4},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		// Sum with range: price >= 15
		sum, err := scoop.Clone().
			Gte("price", 15.0).
			Sum("price")
		assert.NoError(t, err)
		// 15 + 25 + 35 = 75
		assert.True(t, sum > 0)
	})

	t.Run("multiple equals", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": 10.0, "quantity": 5, "category": "A"},
			map[string]interface{}{"name": "P2", "price": 20.0, "quantity": 5, "category": "B"},
			map[string]interface{}{"name": "P3", "price": 30.0, "quantity": 10, "category": "A"},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		// Category A
		avg, err := scoop.Clone().
			Equal("category", "A").
			Avg("price")
		assert.NoError(t, err)
		assert.Equal(t, 20.0, avg) // (10+30)/2 = 20
	})

	t.Run("in filter", func(t *testing.T) {
		scoop := client.NewScoop().Collection(AggProduct{})
		_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

		products := []interface{}{
			map[string]interface{}{"name": "P1", "price": 10.0, "category": "A"},
			map[string]interface{}{"name": "P2", "price": 20.0, "category": "B"},
			map[string]interface{}{"name": "P3", "price": 30.0, "category": "C"},
		}
		_, err = scoop.GetCollection().InsertMany(ctx, products)
		require.NoError(t, err)

		// Category in [A, C]
		max, err := scoop.Clone().
			In("category", "A", "C").
			Max("price")
		assert.NoError(t, err)
		assert.Equal(t, 30.0, max)
	})
}

func TestScoop_LargeDataset(t *testing.T) {
	cfg := &mongo.Config{Mock: true, Database: "test_db"}
	client, err := mongo.New(cfg)
	require.NoError(t, err)
	ctx := context.Background()

	scoop := client.NewScoop().Collection(AggProduct{})
	_, _ = scoop.GetCollection().DeleteMany(ctx, map[string]interface{}{})

	// Insert 100 documents
	products := make([]interface{}, 100)
	for i := 0; i < 100; i++ {
		products[i] = map[string]interface{}{
			"name":     "Product" + string(rune(i)),
			"price":    float64(i + 1),
			"quantity": i + 1,
		}
	}
	_, err = scoop.GetCollection().InsertMany(ctx, products)
	require.NoError(t, err)

	// Test aggregations
	sum, err := scoop.Clone().Sum("price")
	assert.NoError(t, err)
	assert.Equal(t, 5050.0, sum) // Sum of 1 to 100

	avg, err := scoop.Clone().Avg("price")
	assert.NoError(t, err)
	assert.Equal(t, 50.5, avg)

	max, err := scoop.Clone().Max("price")
	assert.NoError(t, err)
	assert.Equal(t, 100.0, max)

	min, err := scoop.Clone().Min("price")
	assert.NoError(t, err)
	assert.Equal(t, 1.0, min)
}

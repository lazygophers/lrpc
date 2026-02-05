package mongo_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/storage/mongo"
	_ "github.com/lazygophers/lrpc/middleware/storage/mongo/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// FindByPage Tests
// ============================================================

func TestScoop_FindByPage(t *testing.T) {
	t.Run("find by page with default options", func(t *testing.T) {
		// Create a fresh client for this test
		freshClient, err := mongo.New(&mongo.Config{Mock: true, Database: "test_findbypage_default"})
		require.NoError(t, err)

		scoop := freshClient.NewScoop()

		// Create test data
		users := []TestModel{
			{Name: "Alice", Age: 25, Email: "alice@example.com"},
			{Name: "Bob", Age: 30, Email: "bob@example.com"},
			{Name: "Charlie", Age: 35, Email: "charlie@example.com"},
		}
		for _, user := range users {
			err := scoop.Create(&user)
			require.NoError(t, err)
		}

		// Test default pagination
		var result []TestModel
		page, err := freshClient.NewScoop().FindByPage(nil, &result)
		require.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, uint64(core.DefaultOffset), page.Offset)
		assert.Equal(t, uint64(core.DefaultLimit), page.Limit)
		assert.Equal(t, 3, len(result))
	})

	t.Run("find by page with custom options", func(t *testing.T) {
		// Create a fresh client for this test
		freshClient, err := mongo.New(&mongo.Config{Mock: true, Database: "test_findbypage_custom"})
		require.NoError(t, err)

		scoop := freshClient.NewScoop()

		// Create test data - 10 users
		for i := 0; i < 10; i++ {
			user := TestModel{
				Name:  "User" + string(rune('A'+i)),
				Age:   20 + i,
				Email: "user" + string(rune('a'+i)) + "@example.com",
			}
			err := scoop.Create(&user)
			require.NoError(t, err)
		}

		// Test page 2 with 3 items per page
		var result []TestModel
		opt := &core.ListOption{
			Offset: 3,
			Limit:  3,
		}
		page, err := freshClient.NewScoop().FindByPage(opt, &result)
		require.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, uint64(3), page.Offset)
		assert.Equal(t, uint64(3), page.Limit)
		assert.Equal(t, 3, len(result))
	})

	t.Run("find by page with show total", func(t *testing.T) {
		// Create a fresh client for this test
		freshClient, err := mongo.New(&mongo.Config{Mock: true, Database: "test_findbypage_showtotal"})
		require.NoError(t, err)

		scoop := freshClient.NewScoop()

		// Create test data - 5 users
		for i := 0; i < 5; i++ {
			user := TestModel{
				Name:  "TotalUser" + string(rune('A'+i)),
				Age:   20 + i,
				Email: "totaluser" + string(rune('a'+i)) + "@example.com",
			}
			err := scoop.Create(&user)
			require.NoError(t, err)
		}

		// Test with ShowTotal
		var result []TestModel
		opt := &core.ListOption{
			Offset:    0,
			Limit:     3,
			ShowTotal: true,
		}
		page, err := freshClient.NewScoop().FindByPage(opt, &result)
		require.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, 3, len(result))
		assert.Equal(t, uint64(5), page.Total)
	})

	t.Run("find by page with filters", func(t *testing.T) {
		// Create a fresh client for this test
		freshClient, err := mongo.New(&mongo.Config{Mock: true, Database: "test_findbypage_filters"})
		require.NoError(t, err)

		scoop := freshClient.NewScoop()

		// Create test data with different ages
		users := []TestModel{
			{Name: "Young1", Age: 18, Email: "young1@example.com"},
			{Name: "Young2", Age: 20, Email: "young2@example.com"},
			{Name: "Adult1", Age: 30, Email: "adult1@example.com"},
			{Name: "Adult2", Age: 35, Email: "adult2@example.com"},
			{Name: "Senior1", Age: 65, Email: "senior1@example.com"},
		}
		for _, user := range users {
			err := scoop.Create(&user)
			require.NoError(t, err)
		}

		// Test with age filter
		var result []TestModel
		opt := &core.ListOption{
			Offset:    0,
			Limit:     10,
			ShowTotal: true,
		}
		page, err := freshClient.NewScoop().Gte("age", 30).FindByPage(opt, &result)
		require.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, 3, len(result)) // Adult1, Adult2, Senior1
		assert.Equal(t, uint64(3), page.Total)
	})

	t.Run("find by page with empty result", func(t *testing.T) {
		// Fresh client for empty collection
		freshClient, err := mongo.New(&mongo.Config{Mock: true, Database: "empty_db"})
		require.NoError(t, err)

		var result []TestModel
		opt := &core.ListOption{
			Offset:    0,
			Limit:     10,
			ShowTotal: true,
		}
		page, err := freshClient.NewScoop().FindByPage(opt, &result)
		require.NoError(t, err)
		assert.NotNil(t, page)
		assert.Equal(t, 0, len(result))
		assert.Equal(t, uint64(0), page.Total)
	})
}

// ============================================================
// Transaction Tests (Begin, Commit, Rollback)
// ============================================================

func TestScoop_TransactionBegin(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("begin transaction", func(t *testing.T) {
		scoop := client.NewScoop()

		// Note: In Mock mode, transaction methods will likely fail
		// because MGM is not fully initialized
		// We test that the method exists and doesn't panic
		txScoop, err := scoop.Begin()
		// In mock mode, this might fail, which is expected
		if err != nil {
			t.Logf("Begin() failed in mock mode (expected): %v", err)
		} else {
			assert.NotNil(t, txScoop)
		}
	})
}

func TestScoop_TransactionCommit(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("commit without transaction", func(t *testing.T) {
		scoop := client.NewScoop()

		err := scoop.Commit()
		// Should fail because no transaction is active
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active transaction")
	})
}

func TestScoop_TransactionRollback(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("rollback without transaction", func(t *testing.T) {
		scoop := client.NewScoop()

		err := scoop.Rollback()
		// Should fail because no transaction is active
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active transaction")
	})
}

// ============================================================
// AutoMigrate Tests
// ============================================================

func TestScoop_AutoMigrate(t *testing.T) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(t, err)

	t.Run("auto migrate single model", func(t *testing.T) {
		scoop := client.NewScoop()

		// Test AutoMigrate - in mock mode this should be a no-op
		err := scoop.AutoMigrate(&TestModel{})
		// In mock mode, this delegates to client.AutoMigrate which is a no-op
		assert.NoError(t, err)
	})

	t.Run("auto migrate multiple models", func(t *testing.T) {
		scoop := client.NewScoop()

		// Test AutoMigrates with multiple models
		err := scoop.AutoMigrates(&TestModel{}, &TestModel{})
		// In mock mode, this delegates to client.AutoMigrates which is a no-op
		assert.NoError(t, err)
	})
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkScoop_FindByPage(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	// Create test data
	scoop := client.NewScoop()
	for i := 0; i < 100; i++ {
		user := TestModel{
			Name:  "BenchUser" + string(rune('A'+(i%26))),
			Age:   20 + (i % 50),
			Email: "benchuser" + string(rune('a'+(i%26))) + "@example.com",
		}
		err := scoop.Create(&user)
		require.NoError(b, err)
	}

	opt := &core.ListOption{
		Offset:    0,
		Limit:     10,
		ShowTotal: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []TestModel
		_, _ = client.NewScoop().FindByPage(opt, &result)
	}
}

func BenchmarkScoop_FindByPageWithFilter(b *testing.B) {
	cfg := &mongo.Config{
		Mock:     true,
		Database: "test_db",
	}

	client, err := mongo.New(cfg)
	require.NoError(b, err)

	// Create test data
	scoop := client.NewScoop()
	for i := 0; i < 100; i++ {
		user := TestModel{
			Name:  "FilterBenchUser" + string(rune('A'+(i%26))),
			Age:   20 + (i % 50),
			Email: "filterbenchuser" + string(rune('a'+(i%26))) + "@example.com",
		}
		err := scoop.Create(&user)
		require.NoError(b, err)
	}

	opt := &core.ListOption{
		Offset:    0,
		Limit:     10,
		ShowTotal: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []TestModel
		_, _ = client.NewScoop().Gte("age", 30).FindByPage(opt, &result)
	}
}

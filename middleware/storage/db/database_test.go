package db_test

import (
	"os"
	"testing"
	"time"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// TestUser model for testing database operations
type TestUser struct {
	Id        int        `gorm:"primaryKey;autoIncrement"`
	Name      string     `gorm:"size:100;not null"`
	Email     string     `gorm:"size:100;unique"`
	Age       int        `gorm:"default:0"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

func (TestUser) TableName() string {
	return "test_users"
}

// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) *db.Client {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "db_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	config := &db.Config{
		Type:    db.Sqlite,
		Address: tempDir,
		Name:    "test",
	}

	client, err := db.New(config, TestUser{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return client
}

// TestClientFunctions tests the client functions with real database
func TestClientFunctions(t *testing.T) {
	t.Run("test New function", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_new_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_new",
		}

		client, err := db.New(config, TestUser{})
		assert.NilError(t, err)
		assert.Assert(t, client != nil)
	})

	t.Run("test Database method", func(t *testing.T) {
		client := setupTestDB(t)

		gormDB := client.Database()
		assert.Assert(t, gormDB != nil)
	})

	t.Run("test SqlDB method", func(t *testing.T) {
		client := setupTestDB(t)

		sqlDB, err := client.SqlDB()
		assert.NilError(t, err)
		assert.Assert(t, sqlDB != nil)
	})

	t.Run("test DriverType method", func(t *testing.T) {
		client := setupTestDB(t)

		driverType := client.DriverType()
		assert.Equal(t, db.Sqlite, driverType)
	})

	t.Run("test NewScoop method", func(t *testing.T) {
		client := setupTestDB(t)

		scoop := client.NewScoop()
		assert.Assert(t, scoop != nil)
	})

	t.Run("test AutoMigrates method", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_migrate_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_migrate",
		}

		// Test without initial models
		client, err := db.New(config)
		assert.NilError(t, err)

		// Test AutoMigrates with models
		err = client.AutoMigrates(TestUser{})
		assert.NilError(t, err)
	})

	t.Run("test AutoMigrate method", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_single_migrate_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_single_migrate",
		}

		client, err := db.New(config)
		assert.NilError(t, err)

		// Test single AutoMigrate
		err = client.AutoMigrate(TestUser{})
		assert.NilError(t, err)
	})
}

// TestConfigApplyWithDatabase tests the private apply function through client creation
func TestConfigApplyWithDatabase(t *testing.T) {
	t.Run("apply function with sqlite", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_apply_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    "", // Should default to sqlite
			Address: tempDir,
			Name:    "test_apply",
		}

		client, err := db.New(config)
		assert.NilError(t, err)
		assert.Assert(t, client != nil)
		assert.Equal(t, db.Sqlite, config.Type) // apply() should have set this
	})

	t.Run("apply function with sqlite3", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_apply_sqlite3_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    "sqlite3", // Should convert to sqlite
			Address: tempDir,
			Name:    "test_apply_sqlite3",
		}

		client, err := db.New(config)
		assert.NilError(t, err)
		assert.Assert(t, client != nil)
		assert.Equal(t, db.Sqlite, config.Type) // apply() should have converted this
	})

	t.Run("apply function sets defaults", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_apply_defaults_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_apply_defaults",
		}

		client, err := db.New(config)
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		// apply() should have set default name
		assert.Assert(t, config.Name != "")
	})
}

// TestModelScoopFunctions tests model scoop functions with real database
func TestModelScoopFunctions(t *testing.T) {
	t.Run("test model scoop operations", func(t *testing.T) {
		client := setupTestDB(t)
		model := db.NewModel[TestUser](client)

		// Test NewScoop (this should work now with real DB)
		scoop := model.NewScoop()
		assert.Assert(t, scoop != nil)

		// Test basic scoop operations
		scoop2 := scoop.Select("id", "name")
		assert.Assert(t, scoop2 != nil)

		scoop3 := scoop.Where("name", "test")
		assert.Assert(t, scoop3 != nil)

		scoop4 := scoop.Equal("age", 25)
		assert.Assert(t, scoop4 != nil)

		scoop5 := scoop.Limit(10)
		assert.Assert(t, scoop5 != nil)

		scoop6 := scoop.Offset(5)
		assert.Assert(t, scoop6 != nil)
	})
}

// TestScoopFunctions tests scoop functions with real database
func TestScoopFunctions(t *testing.T) {
	t.Run("test scoop creation and basic operations", func(t *testing.T) {
		client := setupTestDB(t)

		scoop := client.NewScoop()
		assert.Assert(t, scoop != nil)

		// Test basic scoop methods
		scoop2 := scoop.Table("test_users")
		assert.Assert(t, scoop2 != nil)

		scoop3 := scoop.Select("id", "name")
		assert.Assert(t, scoop3 != nil)

		scoop4 := scoop.Where("name = ?", "test")
		assert.Assert(t, scoop4 != nil)

		scoop5 := scoop.Equal("id", 1)
		assert.Assert(t, scoop5 != nil)

		scoop6 := scoop.NotEqual("id", 0)
		assert.Assert(t, scoop6 != nil)

		scoop7 := scoop.In("id", []int{1, 2, 3})
		assert.Assert(t, scoop7 != nil)

		scoop8 := scoop.NotIn("id", []int{0})
		assert.Assert(t, scoop8 != nil)

		scoop9 := scoop.Like("name", "test")
		assert.Assert(t, scoop9 != nil)

		scoop10 := scoop.LeftLike("name", "te")
		assert.Assert(t, scoop10 != nil)

		scoop11 := scoop.RightLike("name", "st")
		assert.Assert(t, scoop11 != nil)

		scoop12 := scoop.NotLike("name", "bad")
		assert.Assert(t, scoop12 != nil)

		scoop13 := scoop.Between("age", 18, 65)
		assert.Assert(t, scoop13 != nil)

		scoop14 := scoop.NotBetween("age", 0, 10)
		assert.Assert(t, scoop14 != nil)

		scoop15 := scoop.Unscoped()
		assert.Assert(t, scoop15 != nil)

		scoop16 := scoop.Limit(100)
		assert.Assert(t, scoop16 != nil)

		scoop17 := scoop.Offset(10)
		assert.Assert(t, scoop17 != nil)

		scoop18 := scoop.Group("name")
		assert.Assert(t, scoop18 != nil)

		scoop19 := scoop.Order("id")
		assert.Assert(t, scoop19 != nil)

		scoop20 := scoop.Ignore()
		assert.Assert(t, scoop20 != nil)
	})
}

// TestNewSqlite tests the newSqlite function
func TestNewSqlite(t *testing.T) {
	t.Run("test sqlite driver creation", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_sqlite_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_sqlite",
		}

		// This will indirectly test newSqlite function
		client, err := db.New(config)
		assert.NilError(t, err)
		assert.Assert(t, client != nil)
		assert.Equal(t, db.Sqlite, client.DriverType())
	})
}

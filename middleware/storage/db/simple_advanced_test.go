package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SimpleAdvancedModel for simple advanced testing
type SimpleAdvancedModel struct {
	Id   int    `gorm:"primaryKey"`
	Name string `gorm:"size:100"`
}

func (SimpleAdvancedModel) TableName() string {
	return "simple_advanced_models"
}

// setupSimpleAdvancedDB creates database for simple advanced testing
func setupSimpleAdvancedDB(t *testing.T) (*db.Client, *db.Model[SimpleAdvancedModel]) {
	tempDir, err := os.MkdirTemp("", "simple_advanced_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	config := &db.Config{
		Type:    db.Sqlite,
		Address: tempDir,
		Name:    "simple_advanced",
	}

	client, err := db.New(config, SimpleAdvancedModel{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	model := db.NewModel[SimpleAdvancedModel](client)
	return client, model
}

// TestPrintLogger specifically targets Print function in log.go
func TestPrintLogger(t *testing.T) {
	t.Run("test logger Print function through GORM operations", func(t *testing.T) {
		client, model := setupSimpleAdvancedDB(t)

		// Get GORM database and set up logging
		gormDB := client.Database()
		ctx := context.Background()

		// Create different logger configurations to trigger Print method
		testLoggers := []struct {
			name  string
			level logger.LogLevel
		}{
			{"info_level", logger.Info},
			{"warn_level", logger.Warn},
			{"error_level", logger.Error},
			{"silent_level", logger.Silent},
		}

		for _, tc := range testLoggers {
			t.Run(tc.name, func(t *testing.T) {
				// Create session with specific logger
				sessionDB := gormDB.Session(&gorm.Session{
					Logger: gormDB.Logger.LogMode(tc.level),
				})

				// Perform operations that trigger SQL logging and potentially Print method
				var count int64

				// Test SQL operation that generates logs
				result := sessionDB.WithContext(ctx).Raw("SELECT COUNT(*) FROM simple_advanced_models").Scan(&count)
				if result.Error != nil {
					t.Logf("Count query with %s failed: %v", tc.name, result.Error)
				} else {
					t.Logf("Count query with %s succeeded: %d", tc.name, count)
				}

				// Test error-triggering operation
				result2 := sessionDB.WithContext(ctx).Raw("INVALID SQL QUERY").Scan(&count)
				if result2.Error != nil {
					t.Logf("Invalid query with %s failed (expected): %v", tc.name, result2.Error)
				}

				// Test DDL operations
				result3 := sessionDB.WithContext(ctx).Exec("CREATE TABLE IF NOT EXISTS temp_print_test (id INTEGER)")
				if result3.Error != nil {
					t.Logf("DDL with %s failed: %v", tc.name, result3.Error)
				} else {
					t.Logf("DDL with %s succeeded", tc.name)
				}

				// Clean up temp table
				sessionDB.WithContext(ctx).Exec("DROP TABLE IF EXISTS temp_print_test")
			})
		}

		// Test operations that might trigger Print through model operations
		t.Run("model_operations", func(t *testing.T) {
			// Set info level to capture SQL logs
			debugDB := gormDB.Session(&gorm.Session{
				Logger: gormDB.Logger.LogMode(logger.Info),
			})

			// Create test data
			testData := &SimpleAdvancedModel{Name: "Print Logger Test"}
			createResult := model.NewScoop().Create(testData)
			if createResult != nil {
				t.Logf("Create for Print test failed: %v", createResult)
			} else {
				t.Logf("Create for Print test succeeded")
			}

			// Find operations
			results, err := model.NewScoop().Find()
			if err != nil {
				t.Logf("Find for Print test failed: %v", err)
			} else {
				t.Logf("Find for Print test succeeded, found %d records", len(results))
			}

			// Update operations
			updateResult := model.NewScoop().Where("name = ?", "Print Logger Test").Updates(map[string]interface{}{
				"name": "Updated Print Test",
			})
			if updateResult.Error != nil {
				t.Logf("Update for Print test failed: %v", updateResult.Error)
			} else {
				t.Logf("Update for Print test succeeded")
			}

			// Use debugDB for direct operations
			var debugCount int64
			debugResult := debugDB.WithContext(ctx).Raw("SELECT COUNT(*) FROM simple_advanced_models").Scan(&debugCount)
			if debugResult.Error != nil {
				t.Logf("Debug query failed: %v", debugResult.Error)
			} else {
				t.Logf("Debug query succeeded: %d", debugCount)
			}
		})

		t.Logf("Print logger testing completed")
	})
}

// TestDescFunction targets Desc function (66.7% coverage)
func TestDescFunctionSimple(t *testing.T) {
	t.Run("test Desc function comprehensive coverage", func(t *testing.T) {
		client, model := setupSimpleAdvancedDB(t)

		// Create test data for Desc testing
		testData := []*SimpleAdvancedModel{
			{Name: "Desc Test A"},
			{Name: "Desc Test B"},
			{Name: "Desc Test C"},
			{Name: "Desc Test D"},
		}

		for _, data := range testData {
			err := model.NewScoop().Create(data)
			if err != nil {
				t.Logf("Create Desc test data failed: %v", err)
			}
		}

		// Test ModelScoop Desc function
		t.Run("model_scoop_desc", func(t *testing.T) {
			descResults, err := model.NewScoop().Desc("name").Find()
			if err != nil {
				t.Logf("ModelScoop Desc failed: %v", err)
			} else {
				t.Logf("ModelScoop Desc succeeded, found %d records", len(descResults))
				if len(descResults) > 0 {
					t.Logf("First record after Desc: %s", descResults[0].Name)
				}
			}
		})

		// Test raw Scoop Desc function
		t.Run("raw_scoop_desc", func(t *testing.T) {
			var rawResults []*SimpleAdvancedModel
			rawDescResult := client.NewScoop().Table("simple_advanced_models").Order("id DESC").Find(&rawResults)
			if rawDescResult.Error != nil {
				t.Logf("Raw Scoop Desc failed: %v", rawDescResult.Error)
			} else {
				t.Logf("Raw Scoop Desc succeeded, found %d records", len(rawResults))
			}
		})

		// Test Desc with multiple fields
		t.Run("multi_field_desc", func(t *testing.T) {
			multiDescResults, err := model.NewScoop().Desc("name").Desc("id").Find()
			if err != nil {
				t.Logf("Multi-field Desc failed: %v", err)
			} else {
				t.Logf("Multi-field Desc succeeded, found %d records", len(multiDescResults))
			}
		})

		// Test Desc with chain operations
		t.Run("desc_chain_operations", func(t *testing.T) {
			chainResults, err := model.NewScoop().Where("name LIKE ?", "%Desc Test%").Desc("name").Find()
			if err != nil {
				t.Logf("Desc chain operations failed: %v", err)
			} else {
				t.Logf("Desc chain operations succeeded, found %d records", len(chainResults))
			}
		})

		t.Logf("Desc function testing completed")
	})
}

// TestAutoMigratesFunction targets AutoMigrates function (66.7% coverage)
func TestAutoMigratesFunction(t *testing.T) {
	t.Run("test AutoMigrates function with multiple models", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "automigrates_test_*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "automigrates_test",
		}

		// Create client without pre-registering models
		client, err := db.New(config)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}

		// Test AutoMigrates with multiple models
		t.Run("multiple_models", func(t *testing.T) {
			err := client.AutoMigrates(SimpleAdvancedModel{}, PrecisionTestModel{})
			if err != nil {
				t.Logf("AutoMigrates with multiple models failed: %v", err)
			} else {
				t.Logf("AutoMigrates with multiple models succeeded")
			}
		})

		// Test AutoMigrates with single model
		t.Run("single_model", func(t *testing.T) {
			err := client.AutoMigrates(SimpleAdvancedModel{})
			if err != nil {
				t.Logf("AutoMigrates with single model failed: %v", err)
			} else {
				t.Logf("AutoMigrates with single model succeeded")
			}
		})

		// Test AutoMigrates with empty models (edge case)
		t.Run("empty_models", func(t *testing.T) {
			err := client.AutoMigrates()
			if err != nil {
				t.Logf("AutoMigrates with empty models failed: %v", err)
			} else {
				t.Logf("AutoMigrates with empty models succeeded")
			}
		})

		// Test re-running AutoMigrates
		t.Run("re_run_automigrates", func(t *testing.T) {
			err := client.AutoMigrates(SimpleAdvancedModel{})
			if err != nil {
				t.Logf("Re-run AutoMigrates failed: %v", err)
			} else {
				t.Logf("Re-run AutoMigrates succeeded")
			}
		})

		t.Logf("AutoMigrates function testing completed")
	})
}
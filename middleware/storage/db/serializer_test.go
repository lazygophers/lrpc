package db_test

import (
	"os"
	"testing"
	"time"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// TestModelWithJSON tests a model with map[string]any JSON field
type TestModelWithJSON struct {
	Id        int            `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:100;not null"`
	Raw       map[string]any `gorm:"column:raw;type:json;serializer:json;not null" json:"raw,omitempty"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}

func (TestModelWithJSON) TableName() string {
	return "test_json_models"
}

// TestJSONSerializer tests the JSON serializer with map[string]any fields
func TestJSONSerializer(t *testing.T) {
	t.Run("test map[string]any JSON serialization with Scoop", func(t *testing.T) {
		// Create temporary directory for test database
		tempDir, err := os.MkdirTemp("", "db_test_json_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		// Create database client
		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithJSON{})
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		// Create model scoop
		model := db.NewModel[TestModelWithJSON](client)

		// Test data with map[string]any
		testData := TestModelWithJSON{
			Name: "test_user",
			Raw: map[string]any{
				"key1": "value1",
				"key2": 123,
				"key3": true,
				"key4": map[string]any{
					"nested": "value",
				},
			},
		}

		// Insert data using Scoop
		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)
		assert.Assert(t, testData.Id > 0)

		// Query data back using Scoop
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)

		// Verify the data
		assert.Equal(t, testData.Name, result.Name)
		assert.Assert(t, result.Raw != nil)
		assert.Equal(t, "value1", result.Raw["key1"])
		assert.Equal(t, float64(123), result.Raw["key2"]) // JSON unmarshal converts numbers to float64
		assert.Equal(t, true, result.Raw["key3"])

		// Verify nested map
		nested, ok := result.Raw["key4"].(map[string]any)
		assert.Assert(t, ok)
		assert.Equal(t, "value", nested["nested"])
	})

	t.Run("test empty map[string]any with Scoop", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_empty_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_empty",
		}

		client, err := db.New(config, TestModelWithJSON{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithJSON](client)

		// Test with empty map
		testData := TestModelWithJSON{
			Name: "test_empty",
			Raw:  map[string]any{},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)

		// Query back using Scoop
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result.Raw != nil)
		assert.Equal(t, 0, len(result.Raw))
	})

	t.Run("test update map[string]any with Scoop", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_update_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_update",
		}

		client, err := db.New(config, TestModelWithJSON{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithJSON](client)

		// Insert initial data
		testData := TestModelWithJSON{
			Name: "test_update",
			Raw: map[string]any{
				"initial": "data",
			},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)

		// Update the map
		testData.Raw = map[string]any{
			"updated": "value",
			"count":   42,
		}

		updateResult := model.NewScoop().Updates(&testData)
		assert.NilError(t, updateResult.Error)

		// Query back and verify using Scoop
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)

		assert.Equal(t, "value", result.Raw["updated"])
		assert.Equal(t, float64(42), result.Raw["count"])
		_, exists := result.Raw["initial"]
		assert.Assert(t, !exists) // Old data should be replaced
	})

	t.Run("test Find with JSON serialization", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_find_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_find",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithJSON{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithJSON](client)

		// Insert multiple records
		testData := []TestModelWithJSON{
			{
				Name: "user1",
				Raw: map[string]any{
					"role": "admin",
					"age":  30,
				},
			},
			{
				Name: "user2",
				Raw: map[string]any{
					"role": "user",
					"age":  25,
				},
			},
		}

		for i := range testData {
			err = model.NewScoop().Create(&testData[i])
			assert.NilError(t, err)
		}

		// Find all records
		results, err := model.NewScoop().Find()
		assert.NilError(t, err)
		assert.Equal(t, 2, len(results))

		// Verify JSON deserialization
		assert.Equal(t, "admin", results[0].Raw["role"])
		assert.Equal(t, float64(30), results[0].Raw["age"])
		assert.Equal(t, "user", results[1].Raw["role"])
		assert.Equal(t, float64(25), results[1].Raw["age"])
	})

	t.Run("test CreateInBatches with JSON serialization", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_batch_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_batch",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithJSON{})
		assert.NilError(t, err)

		// Create batch data
		testData := []TestModelWithJSON{
			{
				Name: "batch1",
				Raw: map[string]any{
					"batch": 1,
					"type":  "test",
				},
			},
			{
				Name: "batch2",
				Raw: map[string]any{
					"batch": 2,
					"type":  "test",
				},
			},
			{
				Name: "batch3",
				Raw: map[string]any{
					"batch": 3,
					"type":  "test",
				},
			},
		}

		// Insert in batches
		result := client.NewScoop().CreateInBatches(&testData, 2)
		assert.NilError(t, result.Error)
		assert.Equal(t, int64(3), result.RowsAffected)

		// Verify data
		model := db.NewModel[TestModelWithJSON](client)
		results, err := model.NewScoop().Find()
		assert.NilError(t, err)
		assert.Equal(t, 3, len(results))

		// Check JSON deserialization
		for i, r := range results {
			assert.Equal(t, float64(i+1), r.Raw["batch"])
			assert.Equal(t, "test", r.Raw["type"])
		}
	})

	t.Run("test Where with JSON field", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "db_test_json_where_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_json_where",
			Debug:   true,
		}

		client, err := db.New(config, TestModelWithJSON{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithJSON](client)

		// Insert test data
		testData := TestModelWithJSON{
			Name: "where_test",
			Raw: map[string]any{
				"status": "active",
				"count":  100,
			},
		}

		err = model.NewScoop().Create(&testData)
		assert.NilError(t, err)

		// Query with where clause
		result, err := model.NewScoop().Equal("name", "where_test").First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		assert.Equal(t, "active", result.Raw["status"])
		assert.Equal(t, float64(100), result.Raw["count"])
	})
}

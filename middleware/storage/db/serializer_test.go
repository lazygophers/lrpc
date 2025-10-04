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
	t.Run("test map[string]any JSON serialization", func(t *testing.T) {
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

		// Insert data
		err = client.Database().Create(&testData).Error
		assert.NilError(t, err)
		assert.Assert(t, testData.Id > 0)

		// Query data back
		var result TestModelWithJSON
		err = client.Database().Where("id = ?", testData.Id).First(&result).Error
		assert.NilError(t, err)

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

	t.Run("test empty map[string]any", func(t *testing.T) {
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

		// Test with empty map
		testData := TestModelWithJSON{
			Name: "test_empty",
			Raw:  map[string]any{},
		}

		err = client.Database().Create(&testData).Error
		assert.NilError(t, err)

		// Query back
		var result TestModelWithJSON
		err = client.Database().Where("id = ?", testData.Id).First(&result).Error
		assert.NilError(t, err)
		assert.Assert(t, result.Raw != nil)
		assert.Equal(t, 0, len(result.Raw))
	})

	t.Run("test update map[string]any", func(t *testing.T) {
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

		// Insert initial data
		testData := TestModelWithJSON{
			Name: "test_update",
			Raw: map[string]any{
				"initial": "data",
			},
		}

		err = client.Database().Create(&testData).Error
		assert.NilError(t, err)

		// Update the map
		testData.Raw = map[string]any{
			"updated": "value",
			"count":   42,
		}

		err = client.Database().Save(&testData).Error
		assert.NilError(t, err)

		// Query back and verify
		var result TestModelWithJSON
		err = client.Database().Where("id = ?", testData.Id).First(&result).Error
		assert.NilError(t, err)

		assert.Equal(t, "value", result.Raw["updated"])
		assert.Equal(t, float64(42), result.Raw["count"])
		_, exists := result.Raw["initial"]
		assert.Assert(t, !exists) // Old data should be replaced
	})
}

package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// TestModelWithJSON tests a model with map[string]any JSON field
type TestModelWithJSON struct {
	Id        int            `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"size:100;not null"`
	Raw       map[string]any `gorm:"column:raw;type:json;serializer:json" json:"raw,omitempty"`
	CreatedAt int64          `gorm:"autoCreateTime"`
	UpdatedAt int64          `gorm:"autoUpdateTime"`
}

func (TestModelWithJSON) TableName() string {
	return "test_json_models"
}

// TestModelWithJSONPointer 测试指针类型的struct字段
type UserDetail struct {
	Role     string                 `json:"role"`
	Age      int                    `json:"age"`
	Settings map[string]interface{} `json:"settings"`
}

type TestModelWithJSONPointer struct {
	Id        int         `gorm:"primaryKey;autoIncrement"`
	Name      string      `gorm:"size:100;not null"`
	Raw       *UserDetail `gorm:"column:raw;type:json;serializer:json" json:"raw,omitempty"`
	CreatedAt int64       `gorm:"autoCreateTime"`
	UpdatedAt int64       `gorm:"autoUpdateTime"`
}

func (TestModelWithJSONPointer) TableName() string {
	return "test_json_models_pointer"
}

// getTestConfig returns a MySQL database configuration for testing
func getTestConfig() *db.Config {
	return &db.Config{
		Type:     db.MySQL,
		Address:  "127.0.0.1",
		Port:     3306,
		Name:     "test",
		Username: "root",
		Password: "HNEzz4fang.",
		Debug:    true,
	}
}

// TestJSONSerializer tests the JSON serializer with map[string]any fields
func TestJSONSerializer(t *testing.T) {
	t.Run("test map[string]any JSON serialization with Scoop", func(t *testing.T) {
		// Create temporary directory for test database

		// Use MySQL database for testing
		config := getTestConfig()

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

		config := getTestConfig()

		client, err := db.New(config, TestModelWithJSON{})
		assert.NilError(t, err)

		model := db.NewModel[TestModelWithJSON](client)

		// Test with empty map
		testData := TestModelWithJSON{
			Name: "test_empty",
			Raw:  map[string]any{},
		}

		err = model.NewScoop().Create(&testData)

		// Query back using Scoop
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result.Raw != nil)
		assert.Equal(t, 0, len(result.Raw))
	})

	t.Run("test update map[string]any with Scoop", func(t *testing.T) {

		config := getTestConfig()

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

		config := getTestConfig()

		client, err := db.New(config, TestModelWithJSON{})
		assert.NilError(t, err)

		// Clean up the table before test
		deleteResult := client.NewScoop().Table("test_json_models").Delete()
		assert.NilError(t, deleteResult.Error)

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

		config := getTestConfig()

		client, err := db.New(config, TestModelWithJSON{})
		assert.NilError(t, err)

		// Clean up the table before test
		deleteResult := client.NewScoop().Table("test_json_models").Delete()
		assert.NilError(t, deleteResult.Error)

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

		config := getTestConfig()

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

		// Query with where clause
		result, err := model.NewScoop().Equal("name", "where_test").First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		assert.Equal(t, "active", result.Raw["status"])
		assert.Equal(t, float64(100), result.Raw["count"])
	})

	t.Run("test JSON serializer with pointer struct field", func(t *testing.T) {

		config := getTestConfig()

		client, err := db.New(config, TestModelWithJSONPointer{})
		assert.NilError(t, err)
		model := db.NewModel[TestModelWithJSONPointer](client)

		// Test data with pointer struct
		testData := TestModelWithJSONPointer{
			Name: "pointer_test",
			Raw: &UserDetail{
				Role: "admin",
				Age:  30,
				Settings: map[string]interface{}{
					"theme":         "dark",
					"notifications": true,
				},
			},
		}

		// Test Create
		err = model.NewScoop().Create(&testData)
		assert.Assert(t, testData.Id > 0)

		// Test First - should retrieve non-nil pointer data
		result, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, result != nil)
		assert.Assert(t, result.Raw != nil, "Raw should not be nil after First")
		assert.Equal(t, "admin", result.Raw.Role)
		assert.Equal(t, 30, result.Raw.Age)
		assert.Equal(t, "dark", result.Raw.Settings["theme"])
		assert.Equal(t, true, result.Raw.Settings["notifications"])

		// Test Updates with pointer struct
		result.Raw.Role = "super_admin"
		result.Raw.Age = 35
		updateResult := model.NewScoop().Equal("id", testData.Id).Updates(result)
		assert.NilError(t, updateResult.Error)

		// Verify update by querying again
		updatedResult, err := model.NewScoop().Equal("id", testData.Id).First()
		assert.NilError(t, err)
		assert.Assert(t, updatedResult.Raw != nil, "Raw should not be nil after Updates")
		assert.Equal(t, "super_admin", updatedResult.Raw.Role)
		assert.Equal(t, 35, updatedResult.Raw.Age)
	})
}

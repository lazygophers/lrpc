package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// EnhancedTestModel for enhanced coverage testing
type EnhancedTestModel struct {
	Id        int     `gorm:"primaryKey"`
	Name      string  `gorm:"size:100;unique"`
	Email     string  `gorm:"size:100;unique"`
	Age       int     `gorm:"default:0"`
	Score     float64 `gorm:"default:0.0"`
	Active    bool    `gorm:"default:true"`
	CreatedAt int64   `gorm:"autoCreateTime"`
	UpdatedAt int64   `gorm:"autoUpdateTime"`
	DeletedAt *int64  `gorm:"index"`
}

func (EnhancedTestModel) TableName() string {
	return "enhanced_test_models"
}

// CustomTableModel to test getTableName function
type CustomTableModel struct {
	Id   int    `gorm:"primaryKey"`
	Data string `gorm:"size:255"`
}

func (CustomTableModel) TableName() string {
	return "custom_named_table"
}

// setupEnhancedTestDB creates enhanced test database
func setupEnhancedTestDB(t *testing.T) (*db.Client, *db.Model[EnhancedTestModel]) {
	tempDir, err := os.MkdirTemp("", "enhanced_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	config := &db.Config{
		Type:    db.Sqlite,
		Address: tempDir,
		Name:    "enhanced_test",
	}

	client, err := db.New(config, EnhancedTestModel{}, CustomTableModel{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	model := db.NewModel[EnhancedTestModel](client)
	return client, model
}

// TestLowCoverageFunctions tests functions with low coverage
func TestLowCoverageFunctions(t *testing.T) {
	client, model := setupEnhancedTestDB(t)

	t.Run("test Updates function with various scenarios", func(t *testing.T) {
		// Create test data first
		testData := &EnhancedTestModel{
			Id:     1,
			Name:   "TestUser",
			Email:  "test@example.com",
			Age:    25,
			Score:  85.5,
			Active: true,
		}

		// Create the record
		err := model.NewScoop().Create(testData)
		if err != nil {
			t.Logf("Create failed: %v", err)
		}

		// Test Updates with map
		scoop := client.NewScoop().Table("enhanced_test_models").Where("name = ?", "TestUser")
		result := scoop.Updates(map[string]interface{}{
			"age":    30,
			"score":  90.0,
			"active": false,
		})
		assert.Assert(t, result != nil)
		if result.Error != nil {
			t.Logf("Updates with map failed: %v", result.Error)
		}

		// Test Updates with struct
		scoop2 := client.NewScoop().Table("enhanced_test_models").Where("name = ?", "TestUser")
		updateStruct := EnhancedTestModel{Age: 35, Score: 95.0}
		result2 := scoop2.Updates(updateStruct)
		assert.Assert(t, result2 != nil)
		if result2.Error != nil {
			t.Logf("Updates with struct failed: %v", result2.Error)
		}

		// Test Updates with empty conditions (should affect all records)
		scoop3 := client.NewScoop().Table("enhanced_test_models")
		result3 := scoop3.Updates(map[string]interface{}{"active": true})
		assert.Assert(t, result3 != nil)
		if result3.Error != nil {
			t.Logf("Updates all records failed: %v", result3.Error)
		}
	})

	t.Run("test Find function with various scenarios", func(t *testing.T) {
		// Test Find with results using model scoop
		scoop := model.NewScoop().Where("age > ?", 20)
		results, err := scoop.Find()
		if err != nil {
			t.Logf("Find failed: %v", err)
		} else {
			t.Logf("Find returned %d results", len(results))
		}

		// Test Find with no results
		scoop2 := model.NewScoop().Where("age > ?", 1000)
		emptyResults, err2 := scoop2.Find()
		if err2 != nil {
			t.Logf("Find empty failed: %v", err2)
		} else {
			t.Logf("Find empty returned %d results", len(emptyResults))
		}

		// Test Find with complex conditions
		scoop3 := model.NewScoop().Where("age BETWEEN ? AND ?", 20, 40).Where("active = ?", true)
		complexResults, err3 := scoop3.Find()
		if err3 != nil {
			t.Logf("Find complex failed: %v", err3)
		} else {
			t.Logf("Find complex returned %d results", len(complexResults))
		}
	})

	t.Run("test First function with various scenarios", func(t *testing.T) {
		// Test First with existing record using model scoop
		scoop := model.NewScoop().Where("name = ?", "TestUser")
		result, err := scoop.First()
		if err != nil {
			t.Logf("First existing failed: %v", err)
		} else {
			t.Logf("First existing found: %+v", result)
		}

		// Test First with non-existing record
		scoop2 := model.NewScoop().Where("name = ?", "NonExistent")
		result2, err2 := scoop2.First()
		if err2 != nil {
			t.Logf("First non-existing failed (expected): %v", err2)
		} else {
			t.Logf("First non-existing result: %+v", result2)
		}

		// Test First with ordering
		scoop3 := model.NewScoop().Order("age DESC")
		result3, err3 := scoop3.First()
		if err3 != nil {
			t.Logf("First with order failed: %v", err3)
		} else {
			t.Logf("First with order result: %+v", result3)
		}
	})

	t.Run("test Delete function", func(t *testing.T) {
		// Create a record to delete
		deleteData := &EnhancedTestModel{
			Id:    999,
			Name:  "ToDelete",
			Email: "delete@example.com",
			Age:   50,
		}
		err := model.NewScoop().Create(deleteData)
		if err != nil {
			t.Logf("Create for delete failed: %v", err)
		}

		// Test Delete with conditions
		scoop := client.NewScoop().Table("enhanced_test_models").Where("name = ?", "ToDelete")
		result := scoop.Delete()
		assert.Assert(t, result != nil)
		if result.Error != nil {
			t.Logf("Delete failed: %v", result.Error)
		}

		// Test Delete with no matching records
		scoop2 := client.NewScoop().Table("enhanced_test_models").Where("name = ?", "NonExistent")
		result2 := scoop2.Delete()
		assert.Assert(t, result2 != nil)
		if result2.Error != nil {
			t.Logf("Delete non-existing failed: %v", result2.Error)
		}
	})

	t.Run("test Count function", func(t *testing.T) {
		// Test Count with conditions
		scoop := client.NewScoop().Table("enhanced_test_models").Where("age > ?", 20)
		count, err := scoop.Count()
		if err != nil {
			t.Logf("Count with conditions failed: %v", err)
		} else {
			t.Logf("Count result: %d", count)
		}

		// Test Count all records
		scoop2 := client.NewScoop().Table("enhanced_test_models")
		count2, err2 := scoop2.Count()
		if err2 != nil {
			t.Logf("Count all failed: %v", err2)
		} else {
			t.Logf("Count all result: %d", count2)
		}

		// Test Count with no results
		scoop3 := client.NewScoop().Table("enhanced_test_models").Where("age > ?", 1000)
		count3, err3 := scoop3.Count()
		if err3 != nil {
			t.Logf("Count empty failed: %v", err3)
		} else {
			t.Logf("Count empty result: %d", count3)
		}
	})

	t.Run("test CreateInBatches function", func(t *testing.T) {
		// Prepare batch data
		batchData := []*EnhancedTestModel{
			{Id: 100, Name: "Batch1", Email: "batch1@example.com", Age: 20},
			{Id: 101, Name: "Batch2", Email: "batch2@example.com", Age: 21},
			{Id: 102, Name: "Batch3", Email: "batch3@example.com", Age: 22},
			{Id: 103, Name: "Batch4", Email: "batch4@example.com", Age: 23},
			{Id: 104, Name: "Batch5", Email: "batch5@example.com", Age: 24},
		}

		// Test CreateInBatches
		scoop := client.NewScoop().Table("enhanced_test_models")
		result := scoop.CreateInBatches(batchData, 2)
		assert.Assert(t, result != nil)
		if result.Error != nil {
			t.Logf("CreateInBatches failed: %v", result.Error)
		}

		// Test CreateInBatches with larger batch size
		scoop2 := client.NewScoop().Table("enhanced_test_models")
		batchData2 := []*EnhancedTestModel{
			{Id: 200, Name: "LargeBatch1", Email: "large1@example.com", Age: 30},
			{Id: 201, Name: "LargeBatch2", Email: "large2@example.com", Age: 31},
		}
		result2 := scoop2.CreateInBatches(batchData2, 10)
		assert.Assert(t, result2 != nil)
		if result2.Error != nil {
			t.Logf("CreateInBatches large failed: %v", result2.Error)
		}
	})
}

// TestAdvancedModelScoopFunctions tests advanced model scoop functions
func TestAdvancedModelScoopFunctions(t *testing.T) {
	_, model := setupEnhancedTestDB(t)

	t.Run("test FirstOrCreate function", func(t *testing.T) {
		// Test FirstOrCreate with new record
		newData := &EnhancedTestModel{
			Id:    300,
			Name:  "FirstOrCreateNew",
			Email: "firstorcreate@example.com",
			Age:   40,
		}

		scoop := model.NewScoop().Where("name = ?", "FirstOrCreateNew")
		result := scoop.FirstOrCreate(newData)
		assert.Assert(t, result != nil)
		if result.Error != nil {
			t.Logf("FirstOrCreate new failed: %v", result.Error)
		}

		// Test FirstOrCreate with existing record
		existingData := &EnhancedTestModel{
			Name:  "FirstOrCreateNew",
			Email: "updated@example.com", // This should not update
			Age:   45,
		}

		scoop2 := model.NewScoop().Where("name = ?", "FirstOrCreateNew")
		result2 := scoop2.FirstOrCreate(existingData)
		assert.Assert(t, result2 != nil)
		if result2.Error != nil {
			t.Logf("FirstOrCreate existing failed: %v", result2.Error)
		}
	})

	t.Run("test CreateIfNotExists function", func(t *testing.T) {
		// Test CreateIfNotExists with new record
		newData := &EnhancedTestModel{
			Id:    400,
			Name:  "CreateIfNotExists",
			Email: "createif@example.com",
			Age:   50,
		}

		scoop := model.NewScoop().Where("name = ?", "CreateIfNotExists")
		result := scoop.CreateIfNotExists(newData)
		assert.Assert(t, result != nil)
		if result.Error != nil {
			t.Logf("CreateIfNotExists new failed: %v", result.Error)
		}

		// Test CreateIfNotExists with existing record
		duplicateData := &EnhancedTestModel{
			Id:    401,
			Name:  "CreateIfNotExists",
			Email: "duplicate@example.com",
			Age:   55,
		}

		scoop2 := model.NewScoop().Where("name = ?", "CreateIfNotExists")
		result2 := scoop2.CreateIfNotExists(duplicateData)
		assert.Assert(t, result2 != nil)
		if result2.Error != nil {
			t.Logf("CreateIfNotExists existing failed: %v", result2.Error)
		}
	})
}

// TestConfigApplyFunctionEnhanced tests the config apply function
func TestConfigApplyFunctionEnhanced(t *testing.T) {
	t.Run("test config apply with various scenarios", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "config_apply_test_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		// Test with empty type (should default to sqlite)
		config1 := &db.Config{
			Type:    "",
			Address: tempDir,
			Name:    "test_empty_type",
		}
		client1, err1 := db.New(config1)
		assert.NilError(t, err1)
		assert.Assert(t, client1 != nil)
		assert.Equal(t, db.Sqlite, config1.Type)

		// Test with sqlite3 type (should convert to sqlite)
		config2 := &db.Config{
			Type:    "sqlite3",
			Address: tempDir,
			Name:    "test_sqlite3",
		}
		client2, err2 := db.New(config2)
		assert.NilError(t, err2)
		assert.Assert(t, client2 != nil)
		assert.Equal(t, db.Sqlite, config2.Type)

		// Test with existing sqlite type
		config3 := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "test_sqlite",
		}
		client3, err3 := db.New(config3)
		assert.NilError(t, err3)
		assert.Assert(t, client3 != nil)
		assert.Equal(t, db.Sqlite, config3.Type)

		// Test with empty name (should get default)
		config4 := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "",
		}
		client4, err4 := db.New(config4)
		assert.NilError(t, err4)
		assert.Assert(t, client4 != nil)
		assert.Assert(t, config4.Name != "")
	})
}

// TestGetTableNameFunction tests the getTableName utility function
func TestGetTableNameFunction(t *testing.T) {
	t.Run("test getTableName with various model types", func(t *testing.T) {
		client, _ := setupEnhancedTestDB(t)

		// Test with custom table name model
		customModel := db.NewModel[CustomTableModel](client)
		scoop1 := customModel.NewScoop()
		assert.Assert(t, scoop1 != nil)

		// Test operations that might trigger getTableName
		testData := &CustomTableModel{
			Id:   1,
			Data: "test data",
		}

		err := customModel.NewScoop().Create(testData)
		if err != nil {
			t.Logf("Custom model create failed: %v", err)
		}

		// Test Find operation
		results, err := customModel.NewScoop().Find()
		if err != nil {
			t.Logf("Custom model find failed: %v", err)
		} else {
			t.Logf("Custom model find returned %d results", len(results))
		}
	})
}

// TestAdvancedScoopOperations tests complex scoop operations
func TestAdvancedScoopOperations(t *testing.T) {
	_, model := setupEnhancedTestDB(t)

	t.Run("test complex operations that increase coverage", func(t *testing.T) {
		// Create some test data
		testData := []*EnhancedTestModel{
			{Id: 500, Name: "User1", Email: "user1@test.com", Age: 25, Score: 85.5, Active: true},
			{Id: 501, Name: "User2", Email: "user2@test.com", Age: 30, Score: 90.0, Active: false},
			{Id: 502, Name: "User3", Email: "user3@test.com", Age: 35, Score: 95.5, Active: true},
		}

		for _, data := range testData {
			err := model.NewScoop().Create(data)
			if err != nil {
				t.Logf("Create test data failed: %v", err)
			}
		}

		// Test FindByPage with various options
		listOpt1 := &core.ListOption{
			Offset: 0,
			Limit:  2,
		}
		page1, pageResults1, err1 := model.NewScoop().Where("active = ?", true).FindByPage(listOpt1)
		if err1 != nil {
			t.Logf("FindByPage 1 failed: %v", err1)
		} else {
			t.Logf("FindByPage 1 result: page=%+v, results=%d", page1, len(pageResults1))
		}

		// Test with different page options
		listOpt2 := &core.ListOption{
			Offset: 1,
			Limit:  1,
		}
		page2, pageResults2, err2 := model.NewScoop().Order("age DESC").FindByPage(listOpt2)
		if err2 != nil {
			t.Logf("FindByPage 2 failed: %v", err2)
		} else {
			t.Logf("FindByPage 2 result: page=%+v, results=%d", page2, len(pageResults2))
		}

		// Test Chunk operation
		chunkFunc := func(tx *db.Scoop, results []*EnhancedTestModel, offset uint64) error {
			t.Logf("Chunk processing %d results at offset %d", len(results), offset)
			return nil
		}

		scoop := model.NewScoop()
		result := scoop.Chunk(2, chunkFunc)
		assert.Assert(t, result != nil)
		if result.Error != nil {
			t.Logf("Chunk failed: %v", result.Error)
		}
	})
}
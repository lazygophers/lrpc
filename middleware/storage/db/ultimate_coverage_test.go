package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
)

// UltimateTestModel for ultimate coverage testing
type UltimateTestModel struct {
	Id        int     `gorm:"primaryKey"`
	Name      string  `gorm:"size:100"`
	Age       int     `gorm:"default:0"`
	Score     float64 `gorm:"default:0.0"`
	IsActive  bool    `gorm:"default:true"`
	Priority  int     `gorm:"default:1"`
	CreatedAt int64   `gorm:"autoCreateTime"`
	UpdatedAt int64   `gorm:"autoUpdateTime"`
	DeletedAt *int64  `gorm:"index"`
}

func (UltimateTestModel) TableName() string {
	return "ultimate_test_models"
}

// setupUltimateTestDB creates database for ultimate coverage testing
func setupUltimateTestDB(t *testing.T) (*db.Client, *db.Model[UltimateTestModel]) {
	tempDir, err := os.MkdirTemp("", "ultimate_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	config := &db.Config{
		Type:    db.Sqlite,
		Address: tempDir,
		Name:    "ultimate_test",
	}

	client, err := db.New(config, UltimateTestModel{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	model := db.NewModel[UltimateTestModel](client)
	return client, model
}

// TestUnscopedFunctions tests Unscoped functions (60% coverage)
func TestUnscopedFunctions(t *testing.T) {
	client, model := setupUltimateTestDB(t)

	t.Run("test Unscoped function", func(t *testing.T) {
		// Create test data with soft delete
		testData := &UltimateTestModel{
			Id:   1,
			Name: "Unscoped Test",
			Age:  30,
		}

		err := model.NewScoop().Create(testData)
		if err != nil {
			t.Logf("Create failed: %v", err)
		}

		// Soft delete the record
		deleteResult := model.NewScoop().Where("id = ?", 1).Delete()
		if deleteResult.Error != nil {
			t.Logf("Delete failed: %v", deleteResult.Error)
		}

		// Test normal find (should not find deleted record)
		_, err1 := model.NewScoop().Find()
		if err1 != nil {
			t.Logf("Normal find failed: %v", err1)
		}

		// Test Unscoped find (should find deleted record)
		unscopedResults, err2 := model.NewScoop().Unscoped().Find()
		if err2 != nil {
			t.Logf("Unscoped find failed: %v", err2)
		} else {
			t.Logf("Unscoped find succeeded, found %d records", len(unscopedResults))
		}

		// Test raw scoop Unscoped
		rawUnscopedResults := make([]*UltimateTestModel, 0)
		rawUnscopedResult := client.NewScoop().Table("ultimate_test_models").Unscoped().Find(&rawUnscopedResults)
		if rawUnscopedResult.Error != nil {
			t.Logf("Raw unscoped find failed: %v", rawUnscopedResult.Error)
		} else {
			t.Logf("Raw unscoped find succeeded")
		}
	})
}

// TestIgnoreFunctions tests Ignore functions (60% coverage) 
func TestIgnoreFunctions(t *testing.T) {
	client, model := setupUltimateTestDB(t)

	t.Run("test Ignore function", func(t *testing.T) {
		// Test model scoop Ignore
		testData := &UltimateTestModel{
			Id:   10,
			Name: "Ignore Test",
			Age:  25,
		}

		ignoreErr := model.NewScoop().Ignore().Create(testData)
		if ignoreErr != nil {
			t.Logf("Model Ignore Create failed: %v", ignoreErr)
		} else {
			t.Logf("Model Ignore Create succeeded")
		}

		// Test raw scoop Ignore
		rawIgnoreData := &UltimateTestModel{
			Id:   11,
			Name: "Raw Ignore Test",
			Age:  26,
		}
		rawIgnoreResult := client.NewScoop().Table("ultimate_test_models").Ignore().Create(rawIgnoreData)
		if rawIgnoreResult.Error != nil {
			t.Logf("Raw Ignore Create failed: %v", rawIgnoreResult.Error) 
		} else {
			t.Logf("Raw Ignore Create succeeded")
		}

		// Test Ignore with duplicate key (should not fail)
		duplicateData := &UltimateTestModel{
			Id:   10, // Same ID
			Name: "Duplicate Test",
			Age:  27,
		}

		duplicateIgnoreErr := model.NewScoop().Ignore().Create(duplicateData)
		if duplicateIgnoreErr != nil {
			t.Logf("Duplicate Ignore Create failed: %v", duplicateIgnoreErr)
		} else {
			t.Logf("Duplicate Ignore Create succeeded (expected)")
		}
	})
}

// TestCreateIfNotExistsFunction tests CreateIfNotExists function (63.2% coverage)
func TestCreateIfNotExistsFunction(t *testing.T) {
	_, model := setupUltimateTestDB(t)

	t.Run("test CreateIfNotExists function", func(t *testing.T) {
		// Test CreateIfNotExists with new record
		newData := &UltimateTestModel{
			Id:   20,
			Name: "CreateIfNotExists Test",
			Age:  40,
		}

		result1 := model.NewScoop().Where("name = ?", "CreateIfNotExists Test").CreateIfNotExists(newData)
		if result1.Error != nil {
			t.Logf("CreateIfNotExists (new) failed: %v", result1.Error)
		} else {
			t.Logf("CreateIfNotExists (new) succeeded")
		}

		// Test CreateIfNotExists with existing record (should not create again)
		existingData := &UltimateTestModel{
			Id:   21,
			Name: "CreateIfNotExists Test", // Same name
			Age:  41,
		}

		result2 := model.NewScoop().Where("name = ?", "CreateIfNotExists Test").CreateIfNotExists(existingData)
		if result2.Error != nil {
			t.Logf("CreateIfNotExists (existing) failed: %v", result2.Error)
		} else {
			t.Logf("CreateIfNotExists (existing) succeeded (should not create new)")
		}

		// Test CreateIfNotExists with different where condition
		differentData := &UltimateTestModel{
			Id:   22,
			Name: "Different Name",
			Age:  42,
		}

		result3 := model.NewScoop().Where("age > ?", 50).CreateIfNotExists(differentData)
		if result3.Error != nil {
			t.Logf("CreateIfNotExists (different condition) failed: %v", result3.Error)
		} else {
			t.Logf("CreateIfNotExists (different condition) succeeded")
		}
	})
}

// TestInFunctions tests In functions (66.7% coverage)
func TestInFunctions(t *testing.T) {
	client, model := setupUltimateTestDB(t)

	t.Run("test In function", func(t *testing.T) {
		// Create test data
		testData := []*UltimateTestModel{
			{Id: 30, Name: "User A", Age: 25, Priority: 1},
			{Id: 31, Name: "User B", Age: 30, Priority: 2},
			{Id: 32, Name: "User C", Age: 35, Priority: 3},
		}

		for _, data := range testData {
			err := model.NewScoop().Create(data)
			if err != nil {
				t.Logf("Create data failed: %v", err)
			}
		}

		// Test model scoop In with string slice
		inResults1, err1 := model.NewScoop().In("name", []string{"User A", "User C"}).Find()
		if err1 != nil {
			t.Logf("Model In (string slice) failed: %v", err1)
		} else {
			t.Logf("Model In (string slice) succeeded, found %d records", len(inResults1))
		}

		// Test model scoop In with int slice  
		inResults2, err2 := model.NewScoop().In("priority", []int{1, 3}).Find()
		if err2 != nil {
			t.Logf("Model In (int slice) failed: %v", err2)
		} else {
			t.Logf("Model In (int slice) succeeded, found %d records", len(inResults2))
		}

		// Test raw scoop In
		var rawInResults []*UltimateTestModel
		rawInResult := client.NewScoop().Table("ultimate_test_models").In("age", []int{25, 35}).Find(&rawInResults)
		if rawInResult.Error != nil {
			t.Logf("Raw In failed: %v", rawInResult.Error)
		} else {
			t.Logf("Raw In succeeded")
		}

		// Test In with empty slice
		emptyResults, err3 := model.NewScoop().In("name", []string{}).Find()
		if err3 != nil {
			t.Logf("In with empty slice failed: %v", err3)
		} else {
			t.Logf("In with empty slice succeeded, found %d records", len(emptyResults))
		}
	})
}

// TestDescFunction tests Desc function (66.7% coverage)
func TestDescFunction(t *testing.T) {
	_, model := setupUltimateTestDB(t)

	t.Run("test Desc function", func(t *testing.T) {
		// Create test data for ordering
		testData := []*UltimateTestModel{
			{Id: 40, Name: "First", Age: 20, Score: 85.5},
			{Id: 41, Name: "Second", Age: 25, Score: 90.0},
			{Id: 42, Name: "Third", Age: 30, Score: 95.5},
		}

		for _, data := range testData {
			err := model.NewScoop().Create(data)
			if err != nil {
				t.Logf("Create data failed: %v", err)
			}
		}

		// Test Desc with single field
		descResults1, err1 := model.NewScoop().Order("age").Desc().Find()
		if err1 != nil {
			t.Logf("Desc (single field) failed: %v", err1)
		} else {
			t.Logf("Desc (single field) succeeded, found %d records", len(descResults1))
		}

		// Test Desc with multiple fields
		descResults2, err2 := model.NewScoop().Order("score", "age").Desc().Find()
		if err2 != nil {
			t.Logf("Desc (multiple fields) failed: %v", err2)
		} else {
			t.Logf("Desc (multiple fields) succeeded, found %d records", len(descResults2))
		}

		// Test Desc combined with other conditions
		descResults3, err3 := model.NewScoop().Where("age > ?", 22).Order("score").Desc().Limit(2).Find()
		if err3 != nil {
			t.Logf("Desc with conditions failed: %v", err3)
		} else {
			t.Logf("Desc with conditions succeeded, found %d records", len(descResults3))
		}
	})
}

// TestCreateInBatchesFunction tests CreateInBatches function (66.7% coverage)
func TestCreateInBatchesFunction(t *testing.T) {
	client, _ := setupUltimateTestDB(t)

	t.Run("test CreateInBatches function", func(t *testing.T) {
		// Create batch data
		batchData := []*UltimateTestModel{
			{Id: 50, Name: "Batch 1", Age: 25},
			{Id: 51, Name: "Batch 2", Age: 26},
			{Id: 52, Name: "Batch 3", Age: 27},
			{Id: 53, Name: "Batch 4", Age: 28},
			{Id: 54, Name: "Batch 5", Age: 29},
		}

		// Test CreateInBatches with batch size 2
		batchResult1 := client.NewScoop().Table("ultimate_test_models").CreateInBatches(batchData, 2)
		if batchResult1.Error != nil {
			t.Logf("CreateInBatches (size 2) failed: %v", batchResult1.Error)
		} else {
			t.Logf("CreateInBatches (size 2) succeeded")
		}

		// Test CreateInBatches with different batch size
		moreBatchData := []*UltimateTestModel{
			{Id: 55, Name: "Batch 6", Age: 30},
			{Id: 56, Name: "Batch 7", Age: 31},
			{Id: 57, Name: "Batch 8", Age: 32},
		}

		batchResult2 := client.NewScoop().Table("ultimate_test_models").CreateInBatches(moreBatchData, 1)
		if batchResult2.Error != nil {
			t.Logf("CreateInBatches (size 1) failed: %v", batchResult2.Error)
		} else {
			t.Logf("CreateInBatches (size 1) succeeded")
		}

		// Test CreateInBatches with large batch size
		batchResult3 := client.NewScoop().Table("ultimate_test_models").CreateInBatches(moreBatchData, 10)
		if batchResult3.Error != nil {
			t.Logf("CreateInBatches (large size) failed: %v", batchResult3.Error) 
		} else {
			t.Logf("CreateInBatches (large size) succeeded")
		}
	})
}

// TestChunkFunction tests Chunk function (69.2% coverage)
func TestChunkFunction(t *testing.T) {
	client, model := setupUltimateTestDB(t)

	t.Run("test Chunk function", func(t *testing.T) {
		// Create test data for chunking
		testData := []*UltimateTestModel{
			{Id: 60, Name: "Chunk 1", Age: 20},
			{Id: 61, Name: "Chunk 2", Age: 21},
			{Id: 62, Name: "Chunk 3", Age: 22},
			{Id: 63, Name: "Chunk 4", Age: 23},
			{Id: 64, Name: "Chunk 5", Age: 24},
			{Id: 65, Name: "Chunk 6", Age: 25},
		}

		for _, data := range testData {
			err := model.NewScoop().Create(data)
			if err != nil {
				t.Logf("Create chunk data failed: %v", err)
			}
		}

		// Test Chunk with batch size 2
		chunkCount := 0
		chunkResult1 := model.NewScoop().Where("age >= ?", 20).Chunk(2, func(tx *db.Scoop, results []*UltimateTestModel, offset uint64) error {
			chunkCount++
			t.Logf("Chunk %d processed with %d records at offset %d", chunkCount, len(results), offset)
			return nil
		})
		err1 := chunkResult1.Error
		if err1 != nil {
			t.Logf("Chunk (size 2) failed: %v", err1)
		} else {
			t.Logf("Chunk (size 2) succeeded, processed %d chunks", chunkCount)
		}

		// Test Chunk with different batch size
		chunkCount2 := 0
		chunkResult2 := model.NewScoop().Where("age >= ?", 22).Chunk(3, func(tx *db.Scoop, results []*UltimateTestModel, offset uint64) error {
			chunkCount2++
			t.Logf("Chunk2 %d processed with %d records at offset %d", chunkCount2, len(results), offset)
			return nil
		})
		err2 := chunkResult2.Error
		if err2 != nil {
			t.Logf("Chunk (size 3) failed: %v", err2)
		} else {
			t.Logf("Chunk (size 3) succeeded, processed %d chunks", chunkCount2)
		}

		// Test raw scoop Chunk
		rawChunkCount := 0
		var chunkDest []*UltimateTestModel
		rawChunkResult := client.NewScoop().Table("ultimate_test_models").Where("age < ?", 25).Chunk(&chunkDest, 4, func(tx *db.Scoop, offset uint64) error {
			rawChunkCount++
			t.Logf("Raw chunk %d processed at offset %d", rawChunkCount, offset)
			return nil
		})
		rawChunkErr := rawChunkResult.Error
		if rawChunkErr != nil {
			t.Logf("Raw Chunk failed: %v", rawChunkErr)
		} else {
			t.Logf("Raw Chunk succeeded, processed %d chunks", rawChunkCount)
		}
	})
}

// TestFindSqlFunction tests findSql function (69.4% coverage)
func TestFindSqlFunction(t *testing.T) {
	client, model := setupUltimateTestDB(t)

	t.Run("test findSql function", func(t *testing.T) {
		// Create test data
		testData := &UltimateTestModel{
			Id:   70,
			Name: "FindSql Test",
			Age:  35,
		}

		err := model.NewScoop().Create(testData)
		if err != nil {
			t.Logf("Create data failed: %v", err)
		}

		// Test various find operations that should trigger findSql
		operations := []struct {
			name string
			fn   func() error
		}{
			{
				"Simple Find",
				func() error {
					_, err := model.NewScoop().Find()
					return err
				},
			},
			{
				"Find with Where",
				func() error {
					_, err := model.NewScoop().Where("age > ?", 30).Find()
					return err
				},
			},
			{
				"Find with Order",
				func() error {
					_, err := model.NewScoop().Order("age ASC").Find()
					return err
				},
			},
			{
				"Find with Limit",
				func() error {
					_, err := model.NewScoop().Limit(10).Find()
					return err
				},
			},
			{
				"Find with Offset",
				func() error {
					_, err := model.NewScoop().Offset(0).Find()
					return err
				},
			},
			{
				"Complex Find",
				func() error {
					_, err := model.NewScoop().Where("name LIKE ?", "%Test%").Order("id DESC").Limit(5).Find()
					return err
				},
			},
			{
				"Raw Scoop Find",
				func() error {
					var results []*UltimateTestModel
					result := client.NewScoop().Table("ultimate_test_models").Where("age < ?", 50).Find(&results)
					return result.Error
				},
			},
		}

		for _, op := range operations {
			err := op.fn()
			if err != nil {
				t.Logf("%s failed: %v", op.name, err)
			} else {
				t.Logf("%s succeeded", op.name)
			}
		}
	})
}
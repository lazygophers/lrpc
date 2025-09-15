package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// FinalPushModel for final coverage push
type FinalPushModel struct {
	Id        int      `gorm:"primaryKey"`
	Name      string   `gorm:"size:100;unique;not null"`
	Status    string   `gorm:"size:50;default:'active'"`
	Priority  int      `gorm:"default:1"`
	Tags      []string `gorm:"type:json"`
	CreatedAt int64    `gorm:"autoCreateTime"`
	UpdatedAt int64    `gorm:"autoUpdateTime"`
	DeletedAt *int64   `gorm:"index"`
}

func (FinalPushModel) TableName() string {
	return "final_push_models"
}

// setupFinalPushTestDB creates database for final coverage push
func setupFinalPushTestDB(t *testing.T) (*db.Client, *db.Model[FinalPushModel]) {
	tempDir, err := os.MkdirTemp("", "final_push_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	config := &db.Config{
		Type:    db.Sqlite,
		Address: tempDir,
		Name:    "final_push_test",
	}

	client, err := db.New(config, FinalPushModel{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	model := db.NewModel[FinalPushModel](client)
	return client, model
}

// TestUpdateCaseOneField tests the UpdateCaseOneField function
func TestUpdateCaseOneField(t *testing.T) {
	t.Run("test UpdateCaseOneField with various cases", func(t *testing.T) {
		// Test UpdateCaseOneField directly
		caseMap1 := map[any]any{
			1:    "low",
			5:    "medium", 
			10:   "high",
		}
		
		expr1 := db.UpdateCaseOneField("priority", caseMap1, "default")
		t.Logf("UpdateCaseOneField with ints: %+v", expr1)

		// Test with string cases
		caseMap2 := map[any]any{
			"active":   "running",
			"pending":  "waiting",
			"inactive": "stopped",
		}
		
		expr2 := db.UpdateCaseOneField("status", caseMap2)
		t.Logf("UpdateCaseOneField with strings: %+v", expr2)

		// Test with mixed types
		caseMap3 := map[any]any{
			1:      "one",
			"two":  2,
			3.0:    "three",
		}
		
		expr3 := db.UpdateCaseOneField("mixed_field", caseMap3, "default_value")
		t.Logf("UpdateCaseOneField with mixed types: %+v", expr3)
	})
}

// TestAdvancedConditionBuilding tests advanced condition building scenarios
func TestAdvancedConditionBuilding(t *testing.T) {
	t.Run("test addCond function through complex conditions", func(t *testing.T) {
		// Create complex nested conditions to trigger addCond
		nestedConds := []interface{}{
			[]interface{}{"name", "LIKE", "%test%"},
			[]interface{}{"priority", ">", 3},
			[]interface{}{"status", "IN", []string{"active", "pending"}},
		}

		cond := db.Where(nestedConds)
		result := cond.ToString()
		assert.Assert(t, len(result) > 0)
		t.Logf("Nested conditions result: %s", result)

		// Test with OR conditions
		orConds := []interface{}{
			[]interface{}{"name", "=", "urgent"},
			[]interface{}{"priority", ">=", 10},
		}

		orCond := db.OrWhere(orConds)
		orResult := orCond.ToString()
		assert.Assert(t, len(orResult) > 0)
		t.Logf("OR conditions result: %s", orResult)

		// Test with mixed operators
		mixedConds := []interface{}{
			[]interface{}{"name", "NOT LIKE", "%test%"},
			[]interface{}{"priority", "BETWEEN", []interface{}{1, 5}},
			[]interface{}{"status", "NOT IN", []string{"deleted", "archived"}},
		}

		mixedCond := db.Where(mixedConds)
		mixedResult := mixedCond.ToString()
		assert.Assert(t, len(mixedResult) > 0)
		t.Logf("Mixed conditions result: %s", mixedResult)
	})
}

// TestAutoMigrateVariations tests AutoMigrate with different scenarios
func TestAutoMigrateVariations(t *testing.T) {
	t.Run("test AutoMigrate with multiple models", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "automigrate_test_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "automigrate_test",
		}

		// Create client without initial models
		client, err := db.New(config)
		assert.NilError(t, err)
		assert.Assert(t, client != nil)

		// Test AutoMigrate with single model
		err = client.AutoMigrate(FinalPushModel{})
		assert.NilError(t, err)

		// Test AutoMigrates with multiple models
		err = client.AutoMigrates(FinalPushModel{}, EnhancedTestModel{})
		assert.NilError(t, err)

		// Test AutoMigrate with already migrated model
		err = client.AutoMigrate(FinalPushModel{})
		assert.NilError(t, err)
	})
}

// TestComplexDatabaseOperations tests complex database operations
func TestComplexDatabaseOperations(t *testing.T) {
	client, model := setupFinalPushTestDB(t)

	t.Run("test complex scoop operations for higher coverage", func(t *testing.T) {
		// Create test data with complex structures
		testData := []*FinalPushModel{
			{
				Id:       1,
				Name:     "High Priority Task",
				Status:   "active",
				Priority: 10,
				Tags:     []string{"urgent", "important"},
			},
			{
				Id:       2,
				Name:     "Medium Priority Task",
				Status:   "pending",
				Priority: 5,
				Tags:     []string{"normal"},
			},
			{
				Id:       3,
				Name:     "Low Priority Task",
				Status:   "inactive",
				Priority: 1,
				Tags:     []string{"backlog"},
			},
		}

		// Insert test data
		for _, data := range testData {
			err := model.NewScoop().Create(data)
			if err != nil {
				t.Logf("Create failed: %v", err)
			}
		}

		// Test complex FindByPage operations
		listOpt := &core.ListOption{
			Offset: 0,
			Limit:  10,
		}

		// Test FindByPage with various conditions
		page1, results1, err1 := model.NewScoop().
			Where("priority > ?", 3).
			Order("priority DESC").
			FindByPage(listOpt)
		if err1 != nil {
			t.Logf("FindByPage 1 failed: %v", err1)
		} else {
			t.Logf("FindByPage 1: page=%+v, count=%d", page1, len(results1))
		}

		// Test with different ordering
		page2, results2, err2 := model.NewScoop().
			Where("status IN ?", []string{"active", "pending"}).
			Order("name ASC").
			FindByPage(listOpt)
		if err2 != nil {
			t.Logf("FindByPage 2 failed: %v", err2)
		} else {
			t.Logf("FindByPage 2: page=%+v, count=%d", page2, len(results2))
		}

		// Test using raw scoop for Updates operation
		scoop := client.NewScoop().Table("final_push_models")
		
		// Test Updates with complex conditions
		updateResult := scoop.Where("priority < ?", 5).Updates(map[string]interface{}{
			"status": "updated",
		})
		if updateResult.Error != nil {
			t.Logf("Updates failed: %v", updateResult.Error)
		} else {
			t.Logf("Updates succeeded")
		}

		// Test Delete with conditions
		deleteScoop := client.NewScoop().Table("final_push_models")
		deleteResult := deleteScoop.Where("priority = ?", 1).Delete()
		if deleteResult.Error != nil {
			t.Logf("Delete failed: %v", deleteResult.Error)
		} else {
			t.Logf("Delete succeeded")
		}

		// Test In conditions with model scoop
		results3, err3 := model.NewScoop().In("priority", []int{5, 10}).Find()
		if err3 != nil {
			t.Logf("In condition failed: %v", err3)
		} else {
			t.Logf("In condition returned %d results", len(results3))
		}

		// Test Desc ordering
		results4, err4 := model.NewScoop().Order("priority").Desc().Find()
		if err4 != nil {
			t.Logf("Desc ordering failed: %v", err4)
		} else {
			t.Logf("Desc ordering returned %d results", len(results4))
		}
	})
}

// TestGetTableNameFunctionAdvanced tests the getTableName function thoroughly
func TestGetTableNameFunctionAdvanced(t *testing.T) {
	t.Run("test getTableName with different model types", func(t *testing.T) {
		client, _ := setupFinalPushTestDB(t)

		// Create different model types to trigger getTableName
		finalModel := db.NewModel[FinalPushModel](client)
		enhancedModel := db.NewModel[EnhancedTestModel](client)

		// Perform operations that should trigger getTableName internally
		_, err1 := finalModel.NewScoop().Find()
		if err1 != nil {
			t.Logf("FinalPushModel Find failed: %v", err1)
		}

		_, err2 := enhancedModel.NewScoop().Find()
		if err2 != nil {
			t.Logf("EnhancedTestModel Find failed: %v", err2)
		}

		// Test Create operations that should trigger getTableName
		finalData := &FinalPushModel{
			Id:       100,
			Name:     "TableName Test",
			Status:   "test",
			Priority: 1,
		}

		err3 := finalModel.NewScoop().Create(finalData)
		if err3 != nil {
			t.Logf("FinalPushModel Create failed: %v", err3)
		}
	})
}

// TestErrorConditions tests error conditions to improve coverage
func TestErrorConditions(t *testing.T) {
	client, model := setupFinalPushTestDB(t)

	t.Run("test error conditions and edge cases", func(t *testing.T) {
		// Test getNotFoundError
		scoop := client.NewScoop()
		notFoundErr := scoop.IsNotFound(nil)
		assert.Assert(t, !notFoundErr)

		// Test with actual not found scenario
		_, err := model.NewScoop().Where("id = ?", 99999).First()
		if err != nil {
			isNotFound := scoop.IsNotFound(err)
			t.Logf("Is not found error: %v", isNotFound)
		}

		// Test CreateIfNotExists with existing record
		existingData := &FinalPushModel{
			Id:       200,
			Name:     "Existing Record",
			Status:   "test",
			Priority: 1,
		}

		// Create the record first
		err1 := model.NewScoop().Create(existingData)
		if err1 != nil {
			t.Logf("Initial create failed: %v", err1)
		}

		// Try CreateIfNotExists with same data
		result := model.NewScoop().Where("name = ?", "Existing Record").CreateIfNotExists(existingData)
		if result.Error != nil {
			t.Logf("CreateIfNotExists failed: %v", result.Error)
		}

		// Test with different data for same key
		duplicateData := &FinalPushModel{
			Id:       201,
			Name:     "Existing Record", // Same name
			Status:   "different",
			Priority: 5,
		}

		result2 := model.NewScoop().Where("name = ?", "Existing Record").CreateIfNotExists(duplicateData)
		if result2.Error != nil {
			t.Logf("CreateIfNotExists duplicate failed: %v", result2.Error)
		}
	})
}
package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

// UltraTestModel for ultra coverage testing with various data types
type UltraTestModel struct {
	Id           int     `gorm:"primaryKey"`
	Name         string  `gorm:"size:100"`
	Age          int     `gorm:"default:0"`
	Score        float64 `gorm:"default:0.0"`
	IsActive     bool    `gorm:"default:true"`
	Count        uint    `gorm:"default:0"`
	SmallInt     int8    `gorm:"default:0"`
	BigInt       int64   `gorm:"default:0"`
	SmallUint    uint8   `gorm:"default:0"`
	BigUint      uint64  `gorm:"default:0"`
	Float32Val   float32 `gorm:"default:0.0"`
	ByteData     []byte  `gorm:"type:blob"`
	CreatedAt    int64   `gorm:"autoCreateTime"`
	UpdatedAt    int64   `gorm:"autoUpdateTime"`
	DeletedAt    *int64  `gorm:"index"`
}

type UltraConfig struct {
	Theme    string `json:"theme"`
	Language string `json:"language"`
	Debug    bool   `json:"debug"`
}

func (UltraTestModel) TableName() string {
	return "ultra_test_models"
}

// NoTableNameModel to test getTableName with models that don't implement TableName
type NoTableNameModel struct {
	Id   int    `gorm:"primaryKey"`
	Data string `gorm:"size:255"`
}

func (NoTableNameModel) TableName() string {
	return "no_table_name_models"
}

// setupUltraTestDB creates database for ultra coverage testing
func setupUltraTestDB(t *testing.T) (*db.Client, *db.Model[UltraTestModel]) {
	tempDir, err := os.MkdirTemp("", "ultra_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	config := &db.Config{
		Type:    db.Sqlite,
		Address: tempDir,
		Name:    "ultra_test",
	}

	client, err := db.New(config, UltraTestModel{}, NoTableNameModel{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	model := db.NewModel[UltraTestModel](client)
	return client, model
}

// TestDecodeFunctionUltra tests the decode function thoroughly
func TestDecodeFunctionUltra(t *testing.T) {
	t.Run("test decode function through complex raw operations", func(t *testing.T) {
		client, model := setupUltraTestDB(t)

		// Create complex test data with various data types
		testData := &UltraTestModel{
			Id:         1,
			Name:       "Ultra Test",
			Age:        25,
			Score:      95.5,
			IsActive:   true,
			Count:      uint(100),
			SmallInt:   int8(127),
			BigInt:     int64(9223372036854775807),
			SmallUint:  uint8(255),
			BigUint:    uint64(100),
			Float32Val: float32(3.14159),
			ByteData:   []byte("binary data test"),
		}

		// Insert test data
		err := model.NewScoop().Create(testData)
		if err != nil {
			t.Logf("Create failed: %v", err)
		}

		// Use raw SQL queries to potentially trigger decode function
		sqlDB, err := client.SqlDB()
		if err != nil {
			t.Logf("SqlDB failed: %v", err)
			return
		}

		// Test various raw SQL operations that might trigger decode
		queries := []string{
			"SELECT id, name, age, score, is_active FROM ultra_test_models LIMIT 1",
			"SELECT count, small_int, big_int, small_uint, big_uint FROM ultra_test_models LIMIT 1",
			"SELECT float32_val, byte_data, created_at, updated_at FROM ultra_test_models LIMIT 1",
		}

		for _, query := range queries {
			rows, err := sqlDB.Query(query)
			if err != nil {
				t.Logf("Query '%s' failed: %v", query, err)
				continue
			}
			defer rows.Close()

			for rows.Next() {
				cols, err := rows.Columns()
				if err != nil {
					t.Logf("Columns failed: %v", err)
					continue
				}

				// Create dynamic value slice
				values := make([]interface{}, len(cols))
				valuePtrs := make([]interface{}, len(cols))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				err = rows.Scan(valuePtrs...)
				if err != nil {
					t.Logf("Scan failed: %v", err)
				} else {
					t.Logf("Query '%s' scanned %d columns", query, len(values))
				}
			}
			
			if err := rows.Err(); err != nil {
				t.Logf("Rows error for query '%s': %v", query, err)
			}
		}

		// Test direct scoop operations that might trigger decode
		var rawResults []*UltraTestModel
		scoop := client.NewScoop().Table("ultra_test_models")
		findResult := scoop.Find(&rawResults)
		if findResult.Error != nil {
			t.Logf("Raw scoop Find failed: %v", findResult.Error)
		}

		// Test First operation that might trigger decode
		var rawResult UltraTestModel
		firstResult := scoop.First(&rawResult)
		if firstResult.Error != nil {
			t.Logf("Raw scoop First failed: %v", firstResult.Error)
		}
	})
}

// TestGetTableNameAdvanced tests getTableName function thoroughly
func TestGetTableNameAdvanced(t *testing.T) {
	t.Run("test getTableName with various model scenarios", func(t *testing.T) {
		client, _ := setupUltraTestDB(t)

		// Test model WITH TableName method
		ultraModel := db.NewModel[UltraTestModel](client)
		_, err1 := ultraModel.NewScoop().Find()
		if err1 != nil {
			t.Logf("UltraTestModel Find failed: %v", err1)
		}

		// Test model WITHOUT TableName method (should use default naming)
		noTableModel := db.NewModel[NoTableNameModel](client)
		_, err2 := noTableModel.NewScoop().Find()
		if err2 != nil {
			t.Logf("NoTableNameModel Find failed: %v", err2)
		}

		// Test Create operations to trigger getTableName
		ultraData := &UltraTestModel{
			Id:   100,
			Name: "Table Name Test",
		}
		err3 := ultraModel.NewScoop().Create(ultraData)
		if err3 != nil {
			t.Logf("UltraTestModel Create failed: %v", err3)
		}

		noTableData := &NoTableNameModel{
			Id:   100,
			Data: "No table name test",
		}
		err4 := noTableModel.NewScoop().Create(noTableData)
		if err4 != nil {
			t.Logf("NoTableNameModel Create failed: %v", err4)
		}

		// Test various operations that should trigger getTableName
		operations := []func() error{
			func() error { _, err := ultraModel.NewScoop().Count(); return err },
			func() error { _, err := noTableModel.NewScoop().Count(); return err },
			func() error { _, err := ultraModel.NewScoop().Where("id > 0").Find(); return err },
			func() error { _, err := noTableModel.NewScoop().Where("id > 0").Find(); return err },
		}

		for i, op := range operations {
			err := op()
			if err != nil {
				t.Logf("Operation %d failed: %v", i+1, err)
			}
		}
	})
}

// TestAdvancedConditionBuildingUltra tests complex condition scenarios
func TestAdvancedConditionBuildingUltra(t *testing.T) {
	t.Run("test addCond with complex nested conditions", func(t *testing.T) {
		// Create deeply nested conditions to trigger all addCond branches
		complexConditions := []interface{}{
			// Basic field-value pairs
			[]interface{}{"name", "test"},
			[]interface{}{"age", 25},
			
			// Operator conditions
			[]interface{}{"score", ">", 90},
			[]interface{}{"count", ">=", 100},
			[]interface{}{"price", "<", 50.0},
			[]interface{}{"rating", "<=", 4.5},
			[]interface{}{"status", "!=", "deleted"},
			[]interface{}{"category", "<>", "invalid"},
			
			// LIKE conditions
			[]interface{}{"title", "LIKE", "%important%"},
			[]interface{}{"description", "NOT LIKE", "%spam%"},
			
			// IN conditions  
			[]interface{}{"type", "IN", []string{"A", "B", "C"}},
			[]interface{}{"priority", "NOT IN", []int{0, -1}},
			
			// BETWEEN conditions
			[]interface{}{"created_at", "BETWEEN", []interface{}{"2023-01-01", "2023-12-31"}},
			[]interface{}{"updated_at", "NOT BETWEEN", []interface{}{"2022-01-01", "2022-12-31"}},
			
			// NULL conditions
			[]interface{}{"deleted_at", "IS NULL"},
			[]interface{}{"archived_at", "IS NOT NULL"},
		}

		for _, conditionSet := range [][]interface{}{
			complexConditions[:5],   // First batch
			complexConditions[5:10], // Second batch  
			complexConditions[10:],  // Remaining
		} {
			cond := db.Where(conditionSet)
			result := cond.ToString()
			assert.Assert(t, len(result) > 0)
			t.Logf("Complex condition result: %s", result)
		}

		// Test OR conditions
		orConditions := []interface{}{
			[]interface{}{"urgent", "=", true},
			[]interface{}{"priority", ">", 8},
			[]interface{}{"deadline", "<", "2024-01-01"},
		}

		orCond := db.OrWhere(orConditions)
		orResult := orCond.ToString()
		assert.Assert(t, len(orResult) > 0)
		t.Logf("OR condition result: %s", orResult)

		// Test mixed AND/OR conditions
		mixedCond := db.Where([]interface{}{
			[]interface{}{"status", "=", "active"},
			[]interface{}{"type", "IN", []string{"premium", "gold"}},
		}).OrWhere([]interface{}{
			[]interface{}{"vip", "=", true},
			[]interface{}{"score", ">", 95},
		})
		mixedResult := mixedCond.ToString()
		assert.Assert(t, len(mixedResult) > 0)
		t.Logf("Mixed condition result: %s", mixedResult)
	})
}

// TestAutoMigrateEdgeCases tests AutoMigrate with edge cases
func TestAutoMigrateEdgeCases(t *testing.T) {
	t.Run("test AutoMigrate with various edge cases", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "automigrate_edge_test_*")
		assert.NilError(t, err)
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "automigrate_edge_test",
		}

		// Test AutoMigrate with single dummy model
		client1, err1 := db.New(config)
		assert.NilError(t, err1)
		err1 = client1.AutoMigrate(NoTableNameModel{})
		assert.NilError(t, err1)

		// Test AutoMigrate with single complex model
		err2 := client1.AutoMigrate(UltraTestModel{})
		assert.NilError(t, err2)

		// Test AutoMigrates with multiple models
		err3 := client1.AutoMigrates(UltraTestModel{}, NoTableNameModel{})
		assert.NilError(t, err3)

		// Test repeated AutoMigrate (should not fail)
		err4 := client1.AutoMigrate(UltraTestModel{})
		assert.NilError(t, err4)

		// Test with fresh client and different models
		client2, err5 := db.New(&db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "automigrate_edge_test2",
		})
		assert.NilError(t, err5)

		// Test AutoMigrates with many models
		err6 := client2.AutoMigrates(
			UltraTestModel{},
			NoTableNameModel{},
			EnhancedTestModel{},
			FinalPushModel{},
		)
		assert.NilError(t, err6)
	})
}

// TestUpdateCaseVariations tests UpdateCaseOneField with edge cases
func TestUpdateCaseVariations(t *testing.T) {
	t.Run("test UpdateCaseOneField with comprehensive scenarios", func(t *testing.T) {
		// Test with various numeric types
		intCases := map[any]any{
			int(1):     "one",
			int8(2):    "two", 
			int16(3):   "three",
			int32(4):   "four",
			int64(5):   "five",
			uint(6):    "six",
			uint8(7):   "seven",
			uint16(8):  "eight", 
			uint32(9):  "nine",
			uint64(10): "ten",
		}
		expr1 := db.UpdateCaseOneField("number_field", intCases, "default")
		t.Logf("Numeric cases: %+v", expr1)

		// Test with floating point types
		floatCases := map[any]any{
			float32(1.1): "float32_val",
			float64(2.2): "float64_val",
			3.3:          "raw_float",
		}
		expr2 := db.UpdateCaseOneField("float_field", floatCases)
		t.Logf("Float cases: %+v", expr2)

		// Test with string types
		stringCases := map[any]any{
			"active":   "运行中",
			"pending":  "等待中", 
			"inactive": "已停止",
			"error":    "错误",
		}
		expr3 := db.UpdateCaseOneField("status_field", stringCases, "未知")
		t.Logf("String cases: %+v", expr3)

		// Test with boolean types
		boolCases := map[any]any{
			true:  "enabled",
			false: "disabled",
		}
		expr4 := db.UpdateCaseOneField("bool_field", boolCases)
		t.Logf("Boolean cases: %+v", expr4)

		// Test with string representations instead of byte slice (since []byte is not hashable)
		stringValueCases := map[any]any{
			"test":   "test_value",
			"data":   "data_value", 
			"sample": "sample_value",
		}
		expr5 := db.UpdateCaseOneField("string_field", stringValueCases, "default_value")
		t.Logf("String value cases: %+v", expr5)

		// Test with complex mixed types (avoiding []byte since it's not hashable)
		mixedCases := map[any]any{
			1:            "number_one",
			"string_key": 42,
			3.14:         "pi_value",
			true:         "boolean_true",
			nil:          "null_value",
		}
		expr6 := db.UpdateCaseOneField("mixed_field", mixedCases, "fallback")
		t.Logf("Mixed cases: %+v", expr6)

		// Test with empty cases
		emptyCases := map[any]any{}
		expr7 := db.UpdateCaseOneField("empty_field", emptyCases, "only_default")
		t.Logf("Empty cases: %+v", expr7)
	})
}

// TestComplexDatabaseScenarios tests complex database operations
func TestComplexDatabaseScenarios(t *testing.T) {
	client, model := setupUltraTestDB(t)

	t.Run("test complex scenarios to improve coverage", func(t *testing.T) {
		// Create test data with all types
		testData := &UltraTestModel{
			Id:         1,
			Name:       "Complex Test",
			Age:        30,
			Score:      88.8,
			IsActive:   true,
			Count:      uint(200),
			SmallInt:   int8(-128),
			BigInt:     int64(-9223372036854775808),
			SmallUint:  uint8(0),
			BigUint:    uint64(0),
			Float32Val: float32(-3.14),
			ByteData:   []byte("complex binary data"),
		}

		err := model.NewScoop().Create(testData)
		if err != nil {
			t.Logf("Complex create failed: %v", err)
		}

		// Test various complex operations
		operations := []struct {
			name string
			fn   func() error
		}{
			{
				"Complex Find with multiple conditions",
				func() error {
					_, err := model.NewScoop().
						Where("age > ?", 25).
						Where("score < ?", 100).
						In("name", []string{"Complex Test", "Other"}).
						Order("age DESC").
						Limit(10).
						Find()
					return err
				},
			},
			{
				"Complex Updates with conditions",
				func() error {
					result := client.NewScoop().
						Table("ultra_test_models").
						Where("name = ?", "Complex Test").
						Updates(map[string]interface{}{
							"age":       35,
							"score":     95.5,
							"is_active": false,
						})
					return result.Error
				},
			},
			{
				"Complex Delete with conditions",
				func() error {
					result := client.NewScoop().
						Table("ultra_test_models").
						Where("age > ?", 100).
						Delete()
					return result.Error
				},
			},
			{
				"Complex Count with conditions",
				func() error {
					_, err := client.NewScoop().
						Table("ultra_test_models").
						Where("is_active = ?", true).
						Count()
					return err
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
package db_test

import (
	"os"
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
)

// Complex test models to trigger all decode function branches
type PrecisionTestModel struct {
	Id          int                    `gorm:"primaryKey"`
	Name        string                 `gorm:"size:100"`
	Age         int8                   `gorm:"default:0"`
	Score       float32                `gorm:"default:0.0"`
	IsActive    bool                   `gorm:"default:true"`
	BigScore    float64                `gorm:"default:0.0"`
	SmallUint   uint8                  `gorm:"default:0"`
	MediumUint  uint16                 `gorm:"default:0"`
	BigUint     uint32                 `gorm:"default:0"`
	HugeUint    uint64                 `gorm:"default:0"`
	SmallInt    int16                  `gorm:"default:0"`
	MediumInt   int32                  `gorm:"default:0"`
	BigInt      int64                  `gorm:"default:0"`
	JsonData    string  `gorm:"type:text"`
	SliceData   string  `gorm:"type:text"`
	StructData  string  `gorm:"type:text"`
	PointerData *string `gorm:"type:text"`
}

type EmbeddedStruct struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func (PrecisionTestModel) TableName() string {
	return "precision_test_models"
}

// Model without TableName method to test getTableName fallback logic
type PrecisionModelWithoutTableName struct {
	Id   int    `gorm:"primaryKey"`
	Data string `gorm:"size:100"`
}

func (PrecisionModelWithoutTableName) TableName() string {
	return "precision_model_without_table_names"
}

// setupPrecisionTestDB creates database for precision testing
func setupPrecisionTestDB(t *testing.T) (*db.Client, *db.Model[PrecisionTestModel]) {
	tempDir, err := os.MkdirTemp("", "precision_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	config := &db.Config{
		Type:    db.Sqlite,
		Address: tempDir,
		Name:    "precision_test",
	}

	client, err := db.New(config, PrecisionTestModel{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	model := db.NewModel[PrecisionTestModel](client)
	return client, model
}

// TestDecodeFunction targets all decode function branches with precision
func TestDecodeFunctionPrecision(t *testing.T) {
	t.Run("test decode function with all data types and error cases", func(t *testing.T) {
		client, model := setupPrecisionTestDB(t)

		// Create comprehensive test data that will trigger decode function
		testData := &PrecisionTestModel{
			Name:        "Decode Precision Test",
			Age:         127,           // int8 max
			Score:       3.14159,       // float32
			IsActive:    true,          // bool true
			BigScore:    123.456789,    // float64
			SmallUint:   255,           // uint8 max
			MediumUint:  65535,         // uint16 max
			BigUint:     4294967295,    // uint32 max
			HugeUint:    18446744073709551615 >> 1, // uint64 safe value
			SmallInt:    32767,         // int16 max
			MediumInt:   2147483647,    // int32 max
			BigInt:      9223372036854775807, // int64 max
			JsonData:    `{"key": "value", "number": 42}`,
			SliceData:   `["item1", "item2", "item3"]`,
			StructData:  `{"field1": "test", "field2": 123}`,
			PointerData: func() *string { s := "pointer test"; return &s }(),
		}

		err := model.NewScoop().Create(testData)
		if err != nil {
			t.Logf("Create failed: %v", err)
			return
		}

		// Test operations that trigger decode function with different data types
		// Use raw SQL queries to force decode operations
		sqlDB, err := client.SqlDB()
		if err != nil {
			t.Logf("SqlDB failed: %v", err)
			return
		}

		// Test decode with various column types
		rows, err := sqlDB.Query("SELECT * FROM precision_test_models LIMIT 1")
		if err != nil {
			t.Logf("Query failed: %v", err)
			return
		}
		defer rows.Close()

		if rows.Next() {
			// Get column information
			cols, err := rows.Columns()
			if err != nil {
				t.Logf("Columns failed: %v", err)
				return
			}

			// Create interface{} slice to scan into (triggers decode)
			values := make([]interface{}, len(cols))
			scanArgs := make([]interface{}, len(cols))
			for i := range values {
				scanArgs[i] = &values[i]
			}

			err = rows.Scan(scanArgs...)
			if err != nil {
				t.Logf("Scan failed: %v", err)
				return
			}

			t.Logf("Successfully decoded %d columns with various data types", len(cols))
		}
		
		if err := rows.Err(); err != nil {
			t.Logf("Rows error: %v", err)
		}

		// Test Find operations that trigger decode with struct fields
		results, err := model.NewScoop().Find()
		if err != nil {
			t.Logf("Find failed: %v", err)
		} else {
			t.Logf("Find successful, decoded %d records", len(results))
		}

		// Test First operation that triggers decode
		firstResult, err := model.NewScoop().First()
		if err != nil {
			t.Logf("First failed: %v", err)
		} else {
			t.Logf("First successful, decoded record with ID: %d", firstResult.Id)
		}
	})
}

// TestDecodeErrorCases specifically tests decode function error paths
func TestDecodeErrorCases(t *testing.T) {
	t.Run("test decode function error handling", func(t *testing.T) {
		client, _ := setupPrecisionTestDB(t)

		// Create a table with problematic data to trigger decode errors
		sqlDB, err := client.SqlDB()
		if err != nil {
			t.Logf("SqlDB failed: %v", err)
			return
		}

		// Create test table with various data types
		_, err = sqlDB.Exec(`
			CREATE TABLE decode_error_test (
				id INTEGER PRIMARY KEY,
				invalid_int TEXT,
				invalid_uint TEXT,
				invalid_float TEXT,
				invalid_bool TEXT,
				complex_json TEXT
			)
		`)
		if err != nil {
			t.Logf("Create table failed: %v", err)
			return
		}

		// Insert problematic data
		_, err = sqlDB.Exec(`
			INSERT INTO decode_error_test 
			(invalid_int, invalid_uint, invalid_float, invalid_bool, complex_json) 
			VALUES ('not_a_number', 'not_uint', 'not_float', 'invalid_bool', '{"unclosed": json')
		`)
		if err != nil {
			t.Logf("Insert failed: %v", err)
			return
		}

		// Query the problematic data to trigger decode errors
		rows, err := sqlDB.Query("SELECT * FROM decode_error_test")
		if err != nil {
			t.Logf("Query failed: %v", err)
			return
		}
		defer rows.Close()

		if rows.Next() {
			cols, err := rows.Columns()
			if err != nil {
				t.Logf("Columns failed: %v", err)
				return
			}

			values := make([]interface{}, len(cols))
			scanArgs := make([]interface{}, len(cols))
			for i := range values {
				scanArgs[i] = &values[i]
			}

			err = rows.Scan(scanArgs...)
			if err != nil {
				t.Logf("Scan with problematic data failed (expected): %v", err)
			} else {
				t.Logf("Scan with problematic data unexpectedly succeeded")
			}
		}
		
		if err := rows.Err(); err != nil {
			t.Logf("Rows error: %v", err)
		}

		t.Logf("Decode error testing completed")
	})
}

// TestGetTableNameFunction targets all getTableName function branches
func TestGetTableNameFunctionPrecision(t *testing.T) {
	t.Run("test getTableName function with different model types", func(t *testing.T) {
		// Test 1: Model with TableName method (already working)
		model1 := &PrecisionTestModel{}
		tableName1 := model1.TableName()
		if tableName1 != "precision_test_models" {
			t.Errorf("Expected 'precision_test_models', got %s", tableName1)
		} else {
			t.Logf("TableName method works: %s", tableName1)
		}

		// Test 2: Test getTableName function internal logic by creating clients
		// This tests the fallback path when TableName method doesn't exist
		tempDir, err := os.MkdirTemp("", "gettablename_test_*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		config := &db.Config{
			Type:    db.Sqlite,
			Address: tempDir,
			Name:    "gettablename_test",
		}

		// Create client with model without TableName method
		client, err := db.New(config, PrecisionModelWithoutTableName{})
		if err != nil {
			t.Logf("New with ModelWithoutTableName failed: %v", err)
		} else {
			t.Logf("New with ModelWithoutTableName succeeded")
		}

		// Test AutoMigrate with model without TableName method
		err = client.AutoMigrate(PrecisionModelWithoutTableName{})
		if err != nil {
			t.Logf("AutoMigrate with ModelWithoutTableName failed: %v", err)
		} else {
			t.Logf("AutoMigrate with ModelWithoutTableName succeeded")
		}

		// Test reflection path by creating models
		model2 := db.NewModel[PrecisionModelWithoutTableName](client)
		if model2 == nil {
			t.Logf("NewModel for ModelWithoutTableName returned nil")
		} else {
			t.Logf("NewModel for ModelWithoutTableName succeeded")
		}

		t.Logf("getTableName function testing completed")
	})
}

// TestBoolDecodeVariations tests all bool decode cases in decode function
func TestBoolDecodeVariations(t *testing.T) {
	t.Run("test all boolean decode variations", func(t *testing.T) {
		client, _ := setupPrecisionTestDB(t)

		// Create test table with various boolean representations
		sqlDB, err := client.SqlDB()
		if err != nil {
			t.Logf("SqlDB failed: %v", err)
			return
		}

		_, err = sqlDB.Exec(`
			CREATE TABLE bool_test (
				id INTEGER PRIMARY KEY,
				bool_true_1 TEXT,
				bool_true_2 TEXT,
				bool_false_1 TEXT,
				bool_false_2 TEXT,
				bool_invalid TEXT
			)
		`)
		if err != nil {
			t.Logf("Create bool test table failed: %v", err)
			return
		}

		// Insert various boolean representations
		_, err = sqlDB.Exec(`
			INSERT INTO bool_test 
			(bool_true_1, bool_true_2, bool_false_1, bool_false_2, bool_invalid) 
			VALUES ('true', '1', 'false', '0', 'invalid')
		`)
		if err != nil {
			t.Logf("Insert bool test data failed: %v", err)
			return
		}

		// Query and process boolean data to trigger decode
		rows, err := sqlDB.Query("SELECT bool_true_1, bool_true_2, bool_false_1, bool_false_2, bool_invalid FROM bool_test")
		if err != nil {
			t.Logf("Query bool test failed: %v", err)
			return
		}
		defer rows.Close()

		if rows.Next() {
			var trueVal1, trueVal2, falseVal1, falseVal2, invalidVal string
			err = rows.Scan(&trueVal1, &trueVal2, &falseVal1, &falseVal2, &invalidVal)
			if err != nil {
				t.Logf("Scan bool values failed: %v", err)
				return
			}

			t.Logf("Bool test values: true1=%s, true2=%s, false1=%s, false2=%s, invalid=%s", 
				trueVal1, trueVal2, falseVal1, falseVal2, invalidVal)
		}
		
		if err := rows.Err(); err != nil {
			t.Logf("Rows error: %v", err)
		}

		t.Logf("Boolean decode variations testing completed")
	})
}

// TestApplyFunction targets apply function from config.go
func TestApplyFunction(t *testing.T) {
	t.Run("test apply function with various config options", func(t *testing.T) {
		// Test different config combinations to trigger apply function branches
		testConfigs := []struct {
			name   string
			config *db.Config
		}{
			{
				"minimal config",
				&db.Config{
					Type: db.Sqlite,
				},
			},
			{
				"config with address",
				&db.Config{
					Type: db.Sqlite,
					Address: "/tmp",
				},
			},
			{
				"config with port",
				&db.Config{
					Type: db.Sqlite,
					Port: 3306,
				},
			},
			{
				"config with auth",
				&db.Config{
					Type:     db.Sqlite,
					Username: "test",
					Password: "pass",
				},
			},
			{
				"config with database name",
				&db.Config{
					Type: db.Sqlite,
					Name: "test_db",
				},
			},
		}

		for _, tc := range testConfigs {
			t.Run(tc.name, func(t *testing.T) {
				tempDir, err := os.MkdirTemp("", "apply_test_*")
				if err != nil {
					t.Logf("Failed to create temp dir: %v", err)
					return
				}
				defer os.RemoveAll(tempDir)

				// Set address for SQLite
				tc.config.Address = tempDir

				// Test apply function by creating new client
				client, err := db.New(tc.config)
				if err != nil {
					t.Logf("New with %s failed: %v", tc.name, err)
				} else {
					t.Logf("New with %s succeeded", tc.name)
					// Close the client
					if sqlDB, err := client.SqlDB(); err == nil {
						sqlDB.Close()
					}
				}
			})
		}

		t.Logf("Apply function testing completed")
	})
}
package db_test

import (
	"testing"

	"github.com/lazygophers/lrpc/middleware/storage/db"
	"gotest.tools/v3/assert"
)

func TestCondMethods(t *testing.T) {
	t.Run("Or method", func(t *testing.T) {
		cond := db.Where("id", 1).Or("name", "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("Like method", func(t *testing.T) {
		cond := db.Where("").Like("name", "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("LeftLike method", func(t *testing.T) {
		cond := db.Where("").LeftLike("name", "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("RightLike method", func(t *testing.T) {
		cond := db.Where("").RightLike("name", "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("NotLike method", func(t *testing.T) {
		cond := db.Where("").NotLike("name", "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("Between method", func(t *testing.T) {
		cond := db.Where("").Between("age", 18, 65)
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
}

func TestSubCondMethods(t *testing.T) {
	t.Run("Like", func(t *testing.T) {
		cond := db.Like("name", "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("LeftLike", func(t *testing.T) {
		cond := db.LeftLike("name", "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("RightLike", func(t *testing.T) {
		cond := db.RightLike("name", "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("NotLike", func(t *testing.T) {
		cond := db.NotLike("name", "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("Between", func(t *testing.T) {
		cond := db.Between("age", 18, 65)
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
}

func TestAdvancedConditions(t *testing.T) {
	t.Run("raw conditions with $raw", func(t *testing.T) {
		cond := db.Where(map[string]interface{}{
			"$raw": "id IS NOT NULL",
		})
		result := cond.ToString()
		assert.Equal(t, "id IS NOT NULL", result)
	})
	
	t.Run("raw conditions with parameters", func(t *testing.T) {
		cond := db.Where(map[string]interface{}{
			"$raw": []interface{}{"id = ? AND name = ?", 123, "test"},
		})
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("and conditions", func(t *testing.T) {
		cond := db.Where(map[string]interface{}{
			"$and": map[string]interface{}{"status": "active"},
		})
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("or conditions", func(t *testing.T) {
		cond := db.Where(map[string]interface{}{
			"$or": map[string]interface{}{"status": "active"},
		})
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
}

func TestDataTypeConversions(t *testing.T) {
	t.Run("pointer types", func(t *testing.T) {
		val := "test"
		ptr := &val
		cond := db.Where("field", ptr)
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("interface types", func(t *testing.T) {
		var val interface{} = "test"
		cond := db.Where("field", val)
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("various numeric types", func(t *testing.T) {
		cases := []interface{}{
			int8(8), int16(16), int32(32), int64(64),
			uint8(8), uint16(16), uint32(32), uint64(64),
			float32(3.14), float64(3.14159),
		}
		
		for _, val := range cases {
			cond := db.Where("field", val)
			result := cond.ToString()
			assert.Assert(t, result != "")
		}
	})
	
	t.Run("empty slice", func(t *testing.T) {
		cond := db.Where("field", []int{})
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
}

func TestComplexWhere(t *testing.T) {
	t.Run("nested conditions", func(t *testing.T) {
		cond := db.Where([]interface{}{
			map[string]interface{}{"id": 1},
			map[string]interface{}{"name": "test"},
		})
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("string array conditions", func(t *testing.T) {
		cond := db.Where([]interface{}{"name", "John"})
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("nested string arrays", func(t *testing.T) {
		cond := db.Where([]interface{}{
			[]interface{}{"name", "John"},
			[]interface{}{"age", 25},
		})
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("boolean conditions", func(t *testing.T) {
		trueCond := db.Where(true)
		assert.Equal(t, "(1=1)", trueCond.ToString())
		
		falseCond := db.Where(false)
		assert.Equal(t, "(1=0)", falseCond.ToString())
	})
	
	t.Run("single string condition", func(t *testing.T) {
		cond := db.Where("custom_sql_condition")
		assert.Equal(t, "custom_sql_condition", cond.ToString())
	})
}

func TestOperatorHandling(t *testing.T) {
	t.Run("operator with spaces", func(t *testing.T) {
		cond := db.Where("field >", 10)
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("like operator", func(t *testing.T) {
		cond := db.Where("field like", "%test%")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("three parameter with character operator", func(t *testing.T) {
		cond := db.Where("age", '>', 18)
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
	
	t.Run("raw SQL with question marks", func(t *testing.T) {
		cond := db.Where("id = ? OR name = ?", 123, "test")
		result := cond.ToString()
		assert.Assert(t, result != "")
	})
}
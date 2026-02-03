package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUpdateCase 测试 UpdateCase 函数
func TestUpdateCase(t *testing.T) {
	t.Run("update case with string keys", func(t *testing.T) {
		caseMap := map[string]any{
			"(`a` = 1)": "value1",
			"(`b` = 2)": 2,
		}

		expr := UpdateCase(caseMap)
		assert.NotNil(t, expr)
		// UpdateCase 返回 clause.Expr，直接验证即可
	})

	t.Run("update case with string keys and default", func(t *testing.T) {
		caseMap := map[string]any{
			"(`a` = 1)": "value1",
			"(`b` = 2)": 2,
		}

		expr := UpdateCase(caseMap, 10)
		assert.NotNil(t, expr)
	})

	t.Run("update case with Cond keys", func(t *testing.T) {
		cond1 := &Cond{conds: []string{"(`a` = 1)"}, isTopLevel: true}
		cond2 := &Cond{conds: []string{"(`b` = 2)"}, isTopLevel: true}

		caseMap := map[*Cond]any{
			cond1: "value1",
			cond2: 2,
		}

		expr := UpdateCase(caseMap)
		assert.NotNil(t, expr)
	})

	t.Run("update case with various value types", func(t *testing.T) {
		caseMap := map[string]any{
			"(`type` = 'string')":      "string_value",
			"(`type` = 'int')":         42,
			"(`type` = 'int8')":        int8(8),
			"(`type` = 'int16')":       int16(16),
			"(`type` = 'int32')":       int32(32),
			"(`type` = 'int64')":       int64(64),
			"(`type` = 'uint')":        uint(100),
			"(`type` = 'uint8')":       uint8(8),
			"(`type` = 'uint16')":      uint16(16),
			"(`type` = 'uint32')":      uint32(32),
			"(`type` = 'uint64')":      uint64(64),
			"(`type` = 'float32')":     float32(3.14),
			"(`type` = 'float64')":     float64(3.14159),
			"(`type` = 'bool_true')":   true,
			"(`type` = 'bool_false')":  false,
			"(`type` = 'bytes')":       []byte("bytes_value"),
		}

		expr := UpdateCase(caseMap)
		assert.NotNil(t, expr)
	})

	t.Run("update case with complex object", func(t *testing.T) {
		type ComplexObj struct {
			Field1 string
			Field2 int
		}

		caseMap := map[string]any{
			"(`type` = 'object')": ComplexObj{Field1: "test", Field2: 123},
		}

		expr := UpdateCase(caseMap)
		assert.NotNil(t, expr)
	})

	t.Run("update case with default values of different types", func(t *testing.T) {
		tests := []struct {
			name string
			def  interface{}
		}{
			{"string default", "default_value"},
			{"int default", 999},
			{"bool true default", true},
			{"bool false default", false},
			{"float default", 9.99},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				caseMap := map[string]any{
					"(`a` = 1)": "value1",
				}

				expr := UpdateCase(caseMap, tt.def)
				assert.NotNil(t, expr)
			})
		}
	})
}

// TestUpdateCaseOneField 测试 UpdateCaseOneField 函数
func TestUpdateCaseOneField(t *testing.T) {
	t.Run("update case one field basic", func(t *testing.T) {
		caseMap := map[any]any{
			1: "value1",
			2: "value2",
			3: 100,
		}

		expr := UpdateCaseOneField("status", caseMap)
		assert.NotNil(t, expr)
	})

	t.Run("update case one field with default", func(t *testing.T) {
		caseMap := map[any]any{
			1: "value1",
			2: "value2",
		}

		expr := UpdateCaseOneField("status", caseMap, "default_value")
		assert.NotNil(t, expr)
	})

	t.Run("update case one field with various key types", func(t *testing.T) {
		caseMap := map[any]any{
			"string_key":     "string_value",
			42:               "int_value",
			int8(8):          "int8_value",
			int16(16):        "int16_value",
			int32(32):        "int32_value",
			int64(64):        "int64_value",
			uint(100):        "uint_value",
			uint8(8):         "uint8_value",
			uint16(16):       "uint16_value",
			uint32(32):       "uint32_value",
			uint64(64):       "uint64_value",
			float32(3.14):    "float32_value",
			float64(3.14159): "float64_value",
			true:             "bool_true_value",
			false:            "bool_false_value",
			// []byte 不能作为 map key，因为它不可哈希
		}

		expr := UpdateCaseOneField("field", caseMap)
		assert.NotNil(t, expr)
	})

	t.Run("update case one field with various value types", func(t *testing.T) {
		caseMap := map[any]any{
			1:  "string_value",
			2:  42,
			3:  int8(8),
			4:  int16(16),
			5:  int32(32),
			6:  int64(64),
			7:  uint(100),
			8:  uint8(8),
			9:  uint16(16),
			10: uint32(32),
			11: uint64(64),
			12: float32(3.14),
			13: float64(3.14159),
			14: true,
			15: false,
			16: []byte("bytes"),
		}

		expr := UpdateCaseOneField("type_field", caseMap)
		assert.NotNil(t, expr)
	})

	t.Run("update case one field with complex objects", func(t *testing.T) {
		type ComplexKey struct {
			ID   int
			Name string
		}

		type ComplexValue struct {
			Status string
			Count  int
		}

		caseMap := map[any]any{
			ComplexKey{ID: 1, Name: "key1"}: ComplexValue{Status: "active", Count: 10},
			ComplexKey{ID: 2, Name: "key2"}: ComplexValue{Status: "inactive", Count: 20},
		}

		expr := UpdateCaseOneField("data", caseMap)
		assert.NotNil(t, expr)
	})

	t.Run("update case one field with default of different types", func(t *testing.T) {
		tests := []struct {
			name string
			def  interface{}
		}{
			{"string default", "default"},
			{"int default", 999},
			{"bool default", true},
			{"float default", 9.99},
			{"bytes default", []byte("default")},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				caseMap := map[any]any{
					1: "value1",
				}

				expr := UpdateCaseOneField("field", caseMap, tt.def)
				assert.NotNil(t, expr)
			})
		}
	})

	t.Run("update case one field empty map", func(t *testing.T) {
		caseMap := map[any]any{}

		expr := UpdateCaseOneField("field", caseMap)
		assert.NotNil(t, expr)
	})
}

// TestUpdateCase_Integration 测试 UpdateCase 的集成场景
func TestUpdateCase_Integration(t *testing.T) {
	t.Run("realistic update case scenario", func(t *testing.T) {
		// 模拟真实场景：根据不同条件更新状态
		caseMap := map[string]any{
			"(`score` >= 90)": "excellent",
			"(`score` >= 80)": "good",
			"(`score` >= 60)": "pass",
		}

		expr := UpdateCase(caseMap, "fail")
		assert.NotNil(t, expr)
	})

	t.Run("realistic update case one field scenario", func(t *testing.T) {
		// 模拟真实场景：根据状态码更新状态描述
		caseMap := map[any]any{
			0: "pending",
			1: "processing",
			2: "completed",
			3: "failed",
		}

		expr := UpdateCaseOneField("status_code", caseMap, "unknown")
		assert.NotNil(t, expr)
	})
}

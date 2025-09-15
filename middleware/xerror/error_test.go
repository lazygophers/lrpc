package xerror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock I18n implementation for testing
type mockI18n struct {
	messages map[int32]map[string]string
}

func (m *mockI18n) Localize(key int32, langs ...string) (string, bool) {
	if m.messages == nil {
		return "", false
	}
	
	for _, lang := range langs {
		if langMap, exists := m.messages[key]; exists {
			if msg, found := langMap[lang]; found {
				return msg, true
			}
		}
	}
	
	// Check for default language (empty string key)
	if langMap, exists := m.messages[key]; exists {
		if msg, found := langMap[""]; found {
			return msg, true
		}
	}
	
	return "", false
}

func TestError_Error(t *testing.T) {
	err := &Error{
		Code: 1001,
		Msg:  "Test error message",
	}
	
	assert.Equal(t, "Test error message", err.Error())
}

func TestError_Clone(t *testing.T) {
	original := &Error{
		Code: 1001,
		Msg:  "Original message",
	}
	
	cloned := original.Clone()
	
	assert.Equal(t, original.Code, cloned.Code)
	assert.Equal(t, original.Msg, cloned.Msg)
	
	// Verify it's a different instance
	assert.NotSame(t, original, cloned)
	
	// Verify modifications don't affect original
	cloned.Msg = "Modified message"
	assert.Equal(t, "Original message", original.Msg)
	assert.Equal(t, "Modified message", cloned.Msg)
}

func TestError_Is(t *testing.T) {
	err1 := &Error{Code: 1001, Msg: "Error 1"}
	err2 := &Error{Code: 1001, Msg: "Error 2"}
	err3 := &Error{Code: 1002, Msg: "Error 3"}
	
	t.Run("same code - should be equal", func(t *testing.T) {
		assert.True(t, err1.Is(err2))
		assert.True(t, err2.Is(err1))
	})
	
	t.Run("different code - should not be equal", func(t *testing.T) {
		assert.False(t, err1.Is(err3))
		assert.False(t, err3.Is(err1))
	})
	
	t.Run("non-xerror type - fallback to errors.Is", func(t *testing.T) {
		stdErr := errors.New("standard error")
		assert.False(t, err1.Is(stdErr))
	})
}

func TestError_CheckCode(t *testing.T) {
	err := &Error{Code: 1001, Msg: "Test"}
	
	assert.True(t, err.CheckCode(1001))
	assert.False(t, err.CheckCode(1002))
}

func TestIs(t *testing.T) {
	xerr1 := &Error{Code: 1001, Msg: "XError 1"}
	xerr2 := &Error{Code: 1001, Msg: "XError 2"}
	xerr3 := &Error{Code: 1002, Msg: "XError 3"}
	stdErr := errors.New("standard error")
	
	t.Run("xerror comparison", func(t *testing.T) {
		assert.True(t, Is(xerr1, xerr2))
		assert.False(t, Is(xerr1, xerr3))
	})
	
	t.Run("standard error fallback", func(t *testing.T) {
		assert.False(t, Is(stdErr, xerr1))
	})
}

func TestCheckCode(t *testing.T) {
	xerr := &Error{Code: 1001, Msg: "Test"}
	stdErr := errors.New("standard error")
	
	assert.True(t, CheckCode(xerr, 1001))
	assert.False(t, CheckCode(xerr, 1002))
	assert.False(t, CheckCode(stdErr, 1001))
}

func TestGetCode(t *testing.T) {
	xerr := &Error{Code: 1001, Msg: "Test"}
	stdErr := errors.New("standard error")
	
	assert.Equal(t, int32(1001), GetCode(xerr))
	assert.Equal(t, int32(-1), GetCode(stdErr))
}

func TestGetMsg(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.Equal(t, "", GetMsg(nil))
	})
	
	t.Run("xerror", func(t *testing.T) {
		xerr := &Error{Code: 1001, Msg: "XError message"}
		assert.Equal(t, "XError message", GetMsg(xerr))
	})
	
	t.Run("standard error", func(t *testing.T) {
		stdErr := errors.New("standard error message")
		assert.Equal(t, "standard error message", GetMsg(stdErr))
	})
}

func TestNew(t *testing.T) {
	// Clear errMap for clean test
	originalErrMap := make(map[int32]*Error)
	for k, v := range errMap {
		originalErrMap[k] = v
	}
	defer func() {
		// Restore original errMap
		for k := range errMap {
			delete(errMap, k)
		}
		for k, v := range originalErrMap {
			errMap[k] = v
		}
	}()
	
	t.Run("unregistered error code", func(t *testing.T) {
		err := New(9999)
		assert.Equal(t, int32(9999), err.Code)
		assert.Equal(t, "", err.Msg)
	})
	
	t.Run("registered error code", func(t *testing.T) {
		// Register a test error
		testErr := &Error{Code: 5000, Msg: "Registered error"}
		Register(testErr)
		
		err := New(5000)
		assert.Equal(t, int32(5000), err.Code)
		assert.Equal(t, "Registered error", err.Msg)
		
		// Verify it's a clone, not the same instance
		assert.NotSame(t, testErr, err)
		
		// Modify the returned error shouldn't affect the registered one
		err.Msg = "Modified message"
		assert.Equal(t, "Registered error", testErr.Msg)
	})
}

func TestNewError(t *testing.T) {
	// Clear errMap and i18n for clean test
	originalErrMap := make(map[int32]*Error)
	for k, v := range errMap {
		originalErrMap[k] = v
	}
	originalI18n := i18n
	defer func() {
		// Restore original state
		for k := range errMap {
			delete(errMap, k)
		}
		for k, v := range originalErrMap {
			errMap[k] = v
		}
		i18n = originalI18n
	}()
	
	t.Run("registered error - no i18n", func(t *testing.T) {
		i18n = nil
		testErr := &Error{Code: 5001, Msg: "Registered error"}
		Register(testErr)
		
		err := NewError(5001, "en")
		assert.Equal(t, int32(5001), err.Code)
		assert.Equal(t, "Registered error", err.Msg)
	})
	
	t.Run("unregistered error - no i18n", func(t *testing.T) {
		i18n = nil
		
		err := NewError(9999, "en")
		assert.Equal(t, int32(9999), err.Code)
		assert.Equal(t, "", err.Msg)
	})
	
	t.Run("unregistered error - with i18n - found", func(t *testing.T) {
		mockI18nImpl := &mockI18n{
			messages: map[int32]map[string]string{
				9999: {
					"en": "English error message",
					"zh": "中文错误信息",
				},
			},
		}
		i18n = mockI18nImpl
		
		err := NewError(9999, "en")
		assert.Equal(t, int32(9999), err.Code)
		assert.Equal(t, "English error message", err.Msg)
		
		err = NewError(9999, "zh")
		assert.Equal(t, int32(9999), err.Code)
		assert.Equal(t, "中文错误信息", err.Msg)
	})
	
	t.Run("unregistered error - with i18n - not found", func(t *testing.T) {
		mockI18nImpl := &mockI18n{
			messages: map[int32]map[string]string{},
		}
		i18n = mockI18nImpl
		
		err := NewError(9999, "en")
		assert.Equal(t, int32(9999), err.Code)
		assert.Equal(t, "", err.Msg)
	})
	
	t.Run("multiple languages fallback", func(t *testing.T) {
		mockI18nImpl := &mockI18n{
			messages: map[int32]map[string]string{
				9999: {
					"zh": "中文错误信息",
				},
			},
		}
		i18n = mockI18nImpl
		
		// First language not found, should try fallback
		err := NewError(9999, "en", "zh")
		assert.Equal(t, int32(9999), err.Code)
		assert.Equal(t, "中文错误信息", err.Msg)
	})
}

func TestNewErrorWithMsg(t *testing.T) {
	err := NewErrorWithMsg(1001, "Custom message")
	assert.Equal(t, int32(1001), err.Code)
	assert.Equal(t, "Custom message", err.Msg)
}

func TestNewSystemError(t *testing.T) {
	err := NewSystemError("System failure")
	assert.Equal(t, int32(ErrSystemError), err.Code)
	assert.Equal(t, "System failure", err.Msg)
}

func TestNewInvalidParam(t *testing.T) {
	t.Run("single argument", func(t *testing.T) {
		err := NewInvalidParam("invalid field")
		assert.Equal(t, int32(ErrInvalidParam), err.Code)
		assert.Equal(t, "invalid field", err.Msg)
	})
	
	t.Run("multiple arguments", func(t *testing.T) {
		err := NewInvalidParam("field", " ", "value", " ", 123)
		assert.Equal(t, int32(ErrInvalidParam), err.Code)
		assert.Equal(t, "field value 123", err.Msg)
	})
	
	t.Run("no arguments", func(t *testing.T) {
		err := NewInvalidParam()
		assert.Equal(t, int32(ErrInvalidParam), err.Code)
		assert.Equal(t, "", err.Msg)
	})
}

func TestNewNoData(t *testing.T) {
	t.Run("single argument", func(t *testing.T) {
		err := NewNoData("no records found")
		assert.Equal(t, int32(ErrNoData), err.Code)
		assert.Equal(t, "no records found", err.Msg)
	})
	
	t.Run("multiple arguments", func(t *testing.T) {
		err := NewNoData("table:", "users", " count:", 0)
		assert.Equal(t, int32(ErrNoData), err.Code)
		assert.Equal(t, "table:users count:0", err.Msg)
	})
	
	t.Run("no arguments", func(t *testing.T) {
		err := NewNoData()
		assert.Equal(t, int32(ErrNoData), err.Code)
		assert.Equal(t, "", err.Msg)
	})
}

func TestSetI18n(t *testing.T) {
	originalI18n := i18n
	defer func() {
		i18n = originalI18n
	}()
	
	mockI18nImpl := &mockI18n{}
	SetI18n(mockI18nImpl)
	
	assert.Same(t, mockI18nImpl, i18n)
}

func TestRegister(t *testing.T) {
	// Clear errMap for clean test
	originalErrMap := make(map[int32]*Error)
	for k, v := range errMap {
		originalErrMap[k] = v
	}
	defer func() {
		// Restore original errMap
		for k := range errMap {
			delete(errMap, k)
		}
		for k, v := range originalErrMap {
			errMap[k] = v
		}
	}()
	
	t.Run("register single error", func(t *testing.T) {
		err := &Error{Code: 5000, Msg: "Single error"}
		Register(err)
		
		assert.Equal(t, err, errMap[5000])
	})
	
	t.Run("register multiple errors", func(t *testing.T) {
		err1 := &Error{Code: 5001, Msg: "Error 1"}
		err2 := &Error{Code: 5002, Msg: "Error 2"}
		err3 := &Error{Code: 5003, Msg: "Error 3"}
		
		Register(err1, err2, err3)
		
		assert.Equal(t, err1, errMap[5001])
		assert.Equal(t, err2, errMap[5002])
		assert.Equal(t, err3, errMap[5003])
	})
	
	t.Run("register no errors", func(t *testing.T) {
		// Should not panic
		Register()
		// No assertion needed, just ensure it doesn't crash
	})
	
	t.Run("overwrite existing error", func(t *testing.T) {
		err1 := &Error{Code: 5004, Msg: "Original error"}
		err2 := &Error{Code: 5004, Msg: "Overwritten error"}
		
		Register(err1)
		assert.Equal(t, err1, errMap[5004])
		
		Register(err2)
		assert.Equal(t, err2, errMap[5004])
		assert.NotEqual(t, err1, errMap[5004])
	})
}

func TestConstants(t *testing.T) {
	// Verify error code constants
	assert.Equal(t, -1, ErrSystemError)
	assert.Equal(t, 1001, ErrInvalidParam)
	assert.Equal(t, 1002, ErrNoAuth)
	assert.Equal(t, 1003, ErrNoData)
	assert.Equal(t, 1004, ErrConflict)
}

func TestErrorInterface(t *testing.T) {
	// Verify Error implements error interface
	var err error = &Error{Code: 1001, Msg: "test"}
	assert.NotNil(t, err)
	assert.Equal(t, "test", err.Error())
}

func TestI18nInterface(t *testing.T) {
	// Verify mockI18n implements I18n interface
	var i18nImpl I18n = &mockI18n{}
	assert.NotNil(t, i18nImpl)
	
	// Test the interface method
	msg, found := i18nImpl.Localize(123, "en")
	assert.IsType(t, "", msg)
	assert.IsType(t, false, found)
}

func TestErrMapInitialization(t *testing.T) {
	// Verify errMap is properly initialized (even if empty)
	assert.NotNil(t, errMap)
	assert.IsType(t, map[int32]*Error{}, errMap)
}

func TestComplexScenarios(t *testing.T) {
	// Save original state
	originalErrMap := make(map[int32]*Error)
	for k, v := range errMap {
		originalErrMap[k] = v
	}
	originalI18n := i18n
	defer func() {
		// Restore original state
		for k := range errMap {
			delete(errMap, k)
		}
		for k, v := range originalErrMap {
			errMap[k] = v
		}
		i18n = originalI18n
	}()
	
	t.Run("complete workflow", func(t *testing.T) {
		// Setup i18n
		mockI18nImpl := &mockI18n{
			messages: map[int32]map[string]string{
				1001: {
					"en": "Invalid parameter",
					"zh": "参数无效",
				},
				1002: {
					"en": "Unauthorized",
				},
			},
		}
		SetI18n(mockI18nImpl)
		
		// Register some errors
		Register(
			&Error{Code: 1000, Msg: "Registered error"},
		)
		
		// Test various scenarios
		
		// 1. Registered error (should use registered message)
		err1 := NewError(1000, "en")
		assert.Equal(t, "Registered error", err1.Msg)
		
		// 2. Unregistered error with i18n (should use i18n)
		err2 := NewError(1001, "en")
		assert.Equal(t, "Invalid parameter", err2.Msg)
		
		// 3. Unregistered error with i18n fallback
		err3 := NewError(1001, "fr", "zh")
		assert.Equal(t, "参数无效", err3.Msg)
		
		// 4. Unregistered error without i18n match
		err4 := NewError(9999, "en")
		assert.Equal(t, "", err4.Msg)
		
		// Test error comparisons
		assert.True(t, err1.CheckCode(1000))
		assert.True(t, CheckCode(err2, 1001))
		assert.True(t, Is(err2, &Error{Code: 1001}))
		
		// Test utility functions
		assert.Equal(t, int32(1001), GetCode(err2))
		assert.Equal(t, "Invalid parameter", GetMsg(err2))
	})
}
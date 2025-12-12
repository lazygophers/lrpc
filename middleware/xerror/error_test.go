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

func (m *mockI18n) Register(lang string, code int32, msg string) {
	if m.messages == nil {
		m.messages = make(map[int32]map[string]string)
	}
	if m.messages[code] == nil {
		m.messages[code] = make(map[string]string)
	}
	m.messages[code][lang] = msg
}

func (m *mockI18n) RegisterBatch(lang string, data map[int32]string) {
	if m.messages == nil {
		m.messages = make(map[int32]map[string]string)
	}
	for code, msg := range data {
		if m.messages[code] == nil {
			m.messages[code] = make(map[string]string)
		}
		m.messages[code][lang] = msg
	}
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

// Tests for new functionality

func TestError_Unwrap(t *testing.T) {
	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("root cause")
		err := &Error{
			Code:  ErrSystemError,
			Msg:   "wrapper error",
			Cause: cause,
		}

		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("error without cause", func(t *testing.T) {
		err := &Error{
			Code: ErrSystemError,
			Msg:  "simple error",
		}

		assert.Nil(t, err.Unwrap())
	})
}

func TestError_ErrorWithCause(t *testing.T) {
	t.Run("error message includes cause", func(t *testing.T) {
		cause := errors.New("database connection failed")
		err := &Error{
			Code:  ErrSystemError,
			Msg:   "query failed",
			Cause: cause,
		}

		expected := "query failed: database connection failed"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("error message without cause", func(t *testing.T) {
		err := &Error{
			Code: ErrSystemError,
			Msg:  "simple error",
		}

		assert.Equal(t, "simple error", err.Error())
	})
}

func TestError_CloneWithCause(t *testing.T) {
	cause := errors.New("original cause")
	original := &Error{
		Code:  1001,
		Msg:   "Original message",
		Cause: cause,
	}

	cloned := original.Clone()

	assert.Equal(t, original.Code, cloned.Code)
	assert.Equal(t, original.Msg, cloned.Msg)
	assert.Equal(t, original.Cause, cloned.Cause)

	// Verify it's a different instance
	assert.NotSame(t, original, cloned)

	// Verify modifications don't affect original
	cloned.Msg = "Modified message"
	assert.Equal(t, "Original message", original.Msg)
	assert.Equal(t, "Modified message", cloned.Msg)
}

func TestNewNoAuth(t *testing.T) {
	t.Run("basic usage", func(t *testing.T) {
		err := NewNoAuth("token expired")
		assert.Equal(t, int32(ErrNoAuth), err.Code)
		assert.Equal(t, "token expired", err.Msg)
		assert.Nil(t, err.Cause)
	})

	t.Run("permission denied", func(t *testing.T) {
		err := NewNoAuth("permission denied")
		assert.Equal(t, int32(ErrNoAuth), err.Code)
		assert.Equal(t, "permission denied", err.Msg)
	})

	t.Run("empty message", func(t *testing.T) {
		err := NewNoAuth("")
		assert.Equal(t, int32(ErrNoAuth), err.Code)
		assert.Equal(t, "", err.Msg)
	})
}

func TestNewConflict(t *testing.T) {
	t.Run("basic usage", func(t *testing.T) {
		err := NewConflict("email already exists")
		assert.Equal(t, int32(ErrConflict), err.Code)
		assert.Equal(t, "email already exists", err.Msg)
		assert.Nil(t, err.Cause)
	})

	t.Run("version mismatch", func(t *testing.T) {
		err := NewConflict("version mismatch")
		assert.Equal(t, int32(ErrConflict), err.Code)
		assert.Equal(t, "version mismatch", err.Msg)
	})

	t.Run("empty message", func(t *testing.T) {
		err := NewConflict("")
		assert.Equal(t, int32(ErrConflict), err.Code)
		assert.Equal(t, "", err.Msg)
	})
}

func TestWrap(t *testing.T) {
	t.Run("wrap nil error", func(t *testing.T) {
		err := Wrap(nil, ErrSystemError)
		assert.Nil(t, err)
	})

	t.Run("wrap standard error", func(t *testing.T) {
		stdErr := errors.New("standard error")
		err := Wrap(stdErr, ErrSystemError)

		assert.Equal(t, int32(ErrSystemError), err.Code)
		assert.Equal(t, "standard error", err.Msg)
		assert.Nil(t, err.Cause)
	})

	t.Run("wrap xerror - returns as is", func(t *testing.T) {
		xerr := &Error{Code: ErrInvalidParam, Msg: "invalid param"}
		err := Wrap(xerr, ErrSystemError)

		// Should return the same xerror, ignoring the code parameter
		assert.Same(t, xerr, err)
		assert.Equal(t, int32(ErrInvalidParam), err.Code)
	})
}

func TestWrapError(t *testing.T) {
	t.Run("wrap nil error", func(t *testing.T) {
		err := WrapError(nil, ErrSystemError, "database error")
		assert.Nil(t, err)
	})

	t.Run("wrap standard error with custom message", func(t *testing.T) {
		stdErr := errors.New("connection refused")
		err := WrapError(stdErr, ErrSystemError, "database query failed")

		assert.Equal(t, int32(ErrSystemError), err.Code)
		assert.Equal(t, "database query failed", err.Msg)
		assert.Equal(t, stdErr, err.Cause)

		// Verify error chain
		assert.True(t, errors.Is(err, stdErr))

		// Verify Error() includes cause
		assert.Equal(t, "database query failed: connection refused", err.Error())
	})

	t.Run("wrap xerror preserves original", func(t *testing.T) {
		xerr := &Error{Code: ErrInvalidParam, Msg: "original error"}
		err := WrapError(xerr, ErrSystemError, "wrapped")

		assert.Equal(t, int32(ErrSystemError), err.Code)
		assert.Equal(t, "wrapped", err.Msg)
		assert.Equal(t, xerr, err.Cause)

		// Verify error chain
		assert.True(t, errors.Is(err, xerr))
	})

	t.Run("multiple levels of wrapping", func(t *testing.T) {
		level1 := errors.New("root cause")
		level2 := WrapError(level1, ErrSystemError, "level 2")
		level3 := WrapError(level2, ErrInvalidParam, "level 3")

		// Should be able to trace back through the chain
		assert.True(t, errors.Is(level3, level2))
		assert.True(t, errors.Is(level3, level1))

		// Verify codes at each level
		assert.Equal(t, int32(ErrInvalidParam), level3.Code)
		assert.Equal(t, int32(ErrSystemError), level2.Code)
	})
}

func TestErrorChainCompatibility(t *testing.T) {
	t.Run("errors.Is with wrapped error", func(t *testing.T) {
		rootErr := errors.New("root cause")
		wrapped := WrapError(rootErr, ErrSystemError, "wrapped")

		assert.True(t, errors.Is(wrapped, rootErr))
	})

	t.Run("errors.As with wrapped error", func(t *testing.T) {
		rootErr := &Error{Code: ErrInvalidParam, Msg: "invalid"}
		wrapped := WrapError(rootErr, ErrSystemError, "wrapped")

		var target *Error
		assert.True(t, errors.As(wrapped, &target))
		// errors.As finds the first matching type, which is the wrapper
		assert.Equal(t, wrapped, target)
		// But we can still access the root error through the chain
		assert.Equal(t, rootErr, target.Cause)
	})

	t.Run("multiple unwrap levels", func(t *testing.T) {
		level1 := errors.New("level 1")
		level2 := WrapError(level1, ErrSystemError, "level 2")
		level3 := WrapError(level2, ErrInvalidParam, "level 3")

		// Should be able to unwrap multiple times
		unwrapped1 := errors.Unwrap(level3)
		assert.Equal(t, level2, unwrapped1)

		unwrapped2 := errors.Unwrap(unwrapped1)
		assert.Equal(t, level1, unwrapped2)

		unwrapped3 := errors.Unwrap(unwrapped2)
		assert.Nil(t, unwrapped3)
	})
}

func TestBackwardCompatibility(t *testing.T) {
	t.Run("NewError is equivalent to New", func(t *testing.T) {
		originalErrMap := make(map[int32]*Error)
		for k, v := range errMap {
			originalErrMap[k] = v
		}
		defer func() {
			for k := range errMap {
				delete(errMap, k)
			}
			for k, v := range originalErrMap {
				errMap[k] = v
			}
		}()

		testErr := &Error{Code: 5000, Msg: "Test error"}
		Register(testErr)

		err1 := New(5000, "en")
		err2 := NewError(5000, "en")

		assert.Equal(t, err1.Code, err2.Code)
		assert.Equal(t, err1.Msg, err2.Msg)
	})
}
func TestI18n_Register(t *testing.T) {
	mock := &mockI18n{}

	// Test registering single translation
	mock.Register("en", 1001, "Invalid parameter")
	msg, found := mock.Localize(1001, "en")
	assert.True(t, found)
	assert.Equal(t, "Invalid parameter", msg)

	// Test overwriting existing translation
	mock.Register("en", 1001, "Invalid param")
	msg, found = mock.Localize(1001, "en")
	assert.True(t, found)
	assert.Equal(t, "Invalid param", msg)

	// Test registering for multiple languages
	mock.Register("zh", 1001, "参数无效")
	mock.Register("fr", 1001, "Paramètre invalide")

	msgEn, foundEn := mock.Localize(1001, "en")
	msgZh, foundZh := mock.Localize(1001, "zh")
	msgFr, foundFr := mock.Localize(1001, "fr")

	assert.True(t, foundEn)
	assert.True(t, foundZh)
	assert.True(t, foundFr)
	assert.Equal(t, "Invalid param", msgEn)
	assert.Equal(t, "参数无效", msgZh)
	assert.Equal(t, "Paramètre invalide", msgFr)
}

func TestI18n_RegisterBatch(t *testing.T) {
	mock := &mockI18n{}

	// Test batch registration
	enErrors := map[int32]string{
		1001: "Invalid parameter",
		1002: "Unauthorized",
		1003: "Not found",
		1004: "Conflict",
	}

	mock.RegisterBatch("en", enErrors)

	// Verify all errors were registered
	for code, expectedMsg := range enErrors {
		msg, found := mock.Localize(code, "en")
		assert.True(t, found, "Error code %d not found", code)
		assert.Equal(t, expectedMsg, msg, "Error code %d has wrong message", code)
	}

	// Test batch registration for another language
	zhErrors := map[int32]string{
		1001: "参数无效",
		1002: "未授权",
		1003: "未找到",
	}

	mock.RegisterBatch("zh", zhErrors)

	// Verify Chinese translations
	for code, expectedMsg := range zhErrors {
		msg, found := mock.Localize(code, "zh")
		assert.True(t, found, "Error code %d not found in Chinese", code)
		assert.Equal(t, expectedMsg, msg, "Error code %d has wrong Chinese message", code)
	}

	// Verify language fallback still works
	msg, found := mock.Localize(1001, "zh", "en")
	assert.True(t, found)
	assert.Equal(t, "参数无效", msg) // Should use Chinese

	msg, found = mock.Localize(1004, "zh", "en")
	assert.True(t, found)
	assert.Equal(t, "Conflict", msg) // Should fallback to English
}

func TestI18n_RegisterWithNew(t *testing.T) {
	// Clear global i18n first
	originalI18n := i18n
	defer func() { i18n = originalI18n }()

	mock := &mockI18n{}
	SetI18n(mock)

	// Register translations
	mock.Register("en", 5001, "Custom Error")
	mock.Register("zh", 5001, "自定义错误")

	// Test using New with i18n
	errEn := New(5001, "en")
	assert.Equal(t, int32(5001), errEn.Code)
	assert.Equal(t, "Custom Error", errEn.Msg)

	errZh := New(5001, "zh")
	assert.Equal(t, int32(5001), errZh.Code)
	assert.Equal(t, "自定义错误", errZh.Msg)

	// Test fallback
	errFr := New(5001, "fr", "en")
	assert.Equal(t, int32(5001), errFr.Code)
	assert.Equal(t, "Custom Error", errFr.Msg)
}

func TestI18n_RegisterEmpty(t *testing.T) {
	mock := &mockI18n{}

	// Register empty batch should not panic
	mock.RegisterBatch("en", map[int32]string{})

	// Verify no errors exist
	_, found := mock.Localize(9999, "en")
	assert.False(t, found)
}

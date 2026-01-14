//go:build !httpcode

package xerror

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPErrorsRegistration(t *testing.T) {
	// Test that HTTP errors are properly registered during init

	t.Run("400 Bad Request", func(t *testing.T) {
		err := New(400)
		assert.Equal(t, int32(400), err.Code)
		assert.Equal(t, "Bad Request", err.Msg)
	})

	t.Run("401 Unauthorized", func(t *testing.T) {
		err := New(401)
		assert.Equal(t, int32(401), err.Code)
		assert.Equal(t, "Unauthorized", err.Msg)
	})

	t.Run("403 Forbidden", func(t *testing.T) {
		err := New(403)
		assert.Equal(t, int32(403), err.Code)
		assert.Equal(t, "Forbidden", err.Msg)
	})

	t.Run("404 Not Found", func(t *testing.T) {
		err := New(404)
		assert.Equal(t, int32(404), err.Code)
		assert.Equal(t, "Not Found", err.Msg)
	})

	t.Run("500 Internal Server Error", func(t *testing.T) {
		err := New(500)
		assert.Equal(t, int32(500), err.Code)
		assert.Equal(t, "Internal Server Error", err.Msg)
	})
}

func TestHTTPErrorsWithNewError(t *testing.T) {
	// Test that HTTP errors work with NewError function

	t.Run("NewError with HTTP codes", func(t *testing.T) {
		err := New(404)
		assert.Equal(t, int32(404), err.Code)
		assert.Equal(t, "Not Found", err.Msg)
	})

	t.Run("NewError returns cloned instances", func(t *testing.T) {
		err1 := New(500)
		err2 := New(500)

		assert.Equal(t, err1.Code, err2.Code)
		assert.Equal(t, err1.Msg, err2.Msg)
		assert.NotSame(t, err1, err2) // Different instances

		// Modifying one shouldn't affect the other
		err1.Msg = "Modified message"
		assert.Equal(t, "Internal Server Error", err2.Msg)
	})
}

func TestHTTPErrorsIntegration(t *testing.T) {
	// Test integration with utility functions

	t.Run("GetCode with HTTP errors", func(t *testing.T) {
		err := New(404)
		assert.Equal(t, int32(404), GetCode(err))
	})

	t.Run("GetMsg with HTTP errors", func(t *testing.T) {
		err := New(500)
		assert.Equal(t, "Internal Server Error", GetMsg(err))
	})

	t.Run("CheckCode with HTTP errors", func(t *testing.T) {
		err := New(403)
		assert.True(t, CheckCode(err, 403))
		assert.False(t, CheckCode(err, 404))
	})

	t.Run("Is comparison with HTTP errors", func(t *testing.T) {
		err1 := New(400)
		err2 := New(400)
		err3 := New(401)

		assert.True(t, Is(err1, err2))
		assert.False(t, Is(err1, err3))
	})
}

func TestHTTPErrorsCloning(t *testing.T) {
	// Test that cloning works correctly for HTTP errors

	t.Run("Clone preserves HTTP error data", func(t *testing.T) {
		original := New(404)
		cloned := original.Clone()

		assert.Equal(t, original.Code, cloned.Code)
		assert.Equal(t, original.Msg, cloned.Msg)
		assert.NotSame(t, original, cloned)

		// Verify modifications don't affect original
		cloned.Msg = "Custom 404 message"
		assert.Equal(t, "Not Found", original.Msg)
		assert.Equal(t, "Custom 404 message", cloned.Msg)
	})
}

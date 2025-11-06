package auth

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestParseBasicAuth(t *testing.T) {
	t.Run("parse valid basic auth credentials", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("testuser:testpass"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		username, password, err := ParseBasicAuth(ctx)
		require.NoError(t, err)
		assert.Equal(t, "testuser", username)
		assert.Equal(t, "testpass", password)
	})

	t.Run("parse credentials with colon in password", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("user:pass:word"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		username, password, err := ParseBasicAuth(ctx)
		require.NoError(t, err)
		assert.Equal(t, "user", username)
		assert.Equal(t, "pass:word", password)
	})

	t.Run("error when authorization header is missing", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}

		username, password, err := ParseBasicAuth(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingCredentials)
		assert.Empty(t, username)
		assert.Empty(t, password)
	})

	t.Run("error when not Basic auth", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Bearer some-token")

		username, password, err := ParseBasicAuth(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Empty(t, username)
		assert.Empty(t, password)
	})

	t.Run("error with invalid base64 encoding", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Basic invalid-base64!@#")

		username, password, err := ParseBasicAuth(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Empty(t, username)
		assert.Empty(t, password)
	})

	t.Run("error when credentials missing colon", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("usernameonly"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		username, password, err := ParseBasicAuth(ctx)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		assert.Empty(t, username)
		assert.Empty(t, password)
	})

	t.Run("parse credentials with empty password", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("user:"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		username, password, err := ParseBasicAuth(ctx)
		require.NoError(t, err)
		assert.Equal(t, "user", username)
		assert.Equal(t, "", password)
	})
}

func TestValidateBasicAuth(t *testing.T) {
	t.Run("validate with user map", func(t *testing.T) {
		config := BasicAuthConfig{
			Users: map[string]string{
				"alice": "secret123",
				"bob":   "password456",
			},
			ContextKey: "username",
		}

		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("alice:secret123"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		err := ValidateBasicAuth(ctx, config)
		require.NoError(t, err)

		// Check username stored in context
		username := ctx.UserValue("username")
		assert.Equal(t, "alice", username)
	})

	t.Run("validate with custom validator", func(t *testing.T) {
		config := BasicAuthConfig{
			Validator: func(username, password string) bool {
				return username == "admin" && password == "admin123"
			},
			ContextKey: "username",
		}

		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("admin:admin123"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		err := ValidateBasicAuth(ctx, config)
		require.NoError(t, err)

		username := ctx.UserValue("username")
		assert.Equal(t, "admin", username)
	})

	t.Run("users map takes precedence over validator", func(t *testing.T) {
		config := BasicAuthConfig{
			Users: map[string]string{
				"user1": "pass1",
			},
			Validator: func(username, password string) bool {
				return true // Always returns true
			},
			ContextKey: "username",
		}

		ctx := &fasthttp.RequestCtx{}
		// Try with credentials not in user map
		credentials := base64.StdEncoding.EncodeToString([]byte("user2:pass2"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		err := ValidateBasicAuth(ctx, config)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("skip validation when skip func returns true", func(t *testing.T) {
		config := BasicAuthConfig{
			Users: map[string]string{
				"user1": "pass1",
			},
			SkipFunc: func(ctx *fasthttp.RequestCtx) bool {
				return true
			},
		}

		ctx := &fasthttp.RequestCtx{}
		// No credentials provided

		err := ValidateBasicAuth(ctx, config)
		require.NoError(t, err)
	})

	t.Run("call success handler after validation", func(t *testing.T) {
		successCalled := false
		var capturedUsername string

		config := BasicAuthConfig{
			Users: map[string]string{
				"test": "pass",
			},
			SuccessHandler: func(ctx *fasthttp.RequestCtx, username string) {
				successCalled = true
				capturedUsername = username
			},
		}

		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("test:pass"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		err := ValidateBasicAuth(ctx, config)
		require.NoError(t, err)
		assert.True(t, successCalled)
		assert.Equal(t, "test", capturedUsername)
	})

	t.Run("error with invalid credentials", func(t *testing.T) {
		config := BasicAuthConfig{
			Users: map[string]string{
				"user1": "correct-password",
			},
		}

		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("user1:wrong-password"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		err := ValidateBasicAuth(ctx, config)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("error when user not in map", func(t *testing.T) {
		config := BasicAuthConfig{
			Users: map[string]string{
				"user1": "pass1",
			},
		}

		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("unknown:pass"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		err := ValidateBasicAuth(ctx, config)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("error when no validator configured", func(t *testing.T) {
		config := BasicAuthConfig{
			ContextKey: "username",
		}

		ctx := &fasthttp.RequestCtx{}
		credentials := base64.StdEncoding.EncodeToString([]byte("user:pass"))
		ctx.Request.Header.Set("Authorization", "Basic "+credentials)

		err := ValidateBasicAuth(ctx, config)
		assert.Error(t, err)
	})

	t.Run("error with missing credentials", func(t *testing.T) {
		config := BasicAuthConfig{
			Users: map[string]string{
				"user1": "pass1",
			},
		}

		ctx := &fasthttp.RequestCtx{}
		// No Authorization header

		err := ValidateBasicAuth(ctx, config)
		assert.Error(t, err)
	})
}

func TestGetUsername(t *testing.T) {
	t.Run("retrieve username from context", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue("username", "testuser")

		username, ok := GetUsername(ctx, "username")
		require.True(t, ok)
		assert.Equal(t, "testuser", username)
	})

	t.Run("return false when username not in context", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}

		username, ok := GetUsername(ctx, "username")
		assert.False(t, ok)
		assert.Empty(t, username)
	})

	t.Run("return false when value is not string", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue("username", 12345)

		username, ok := GetUsername(ctx, "username")
		assert.False(t, ok)
		assert.Empty(t, username)
	})
}

func TestSetWWWAuthenticate(t *testing.T) {
	t.Run("set WWW-Authenticate header with custom realm", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		SetWWWAuthenticate(ctx, "My Realm")

		header := string(ctx.Response.Header.Peek("WWW-Authenticate"))
		assert.Equal(t, `Basic realm="My Realm"`, header)
	})

	t.Run("use default realm when empty", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		SetWWWAuthenticate(ctx, "")

		header := string(ctx.Response.Header.Peek("WWW-Authenticate"))
		assert.Equal(t, `Basic realm="Restricted"`, header)
	})
}

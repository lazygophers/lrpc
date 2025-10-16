package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestGenerateToken(t *testing.T) {
	t.Run("generate valid token with HS256", func(t *testing.T) {
		config := JWTConfig{
			SigningKey:    "test-secret-key",
			SigningMethod: "HS256",
		}

		claims := jwt.MapClaims{
			"user_id": "123",
			"exp":     time.Now().Add(time.Hour).Unix(),
		}

		token, err := GenerateToken(config, claims)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("generate token with custom claims", func(t *testing.T) {
		config := JWTConfig{
			SigningKey:    "test-secret-key",
			SigningMethod: "HS256",
		}

		claims := NewJWTClaims("user123", "testuser", time.Hour)
		claims.Roles = []string{"admin", "user"}

		token, err := GenerateToken(config, claims)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("error when signing key is empty", func(t *testing.T) {
		config := JWTConfig{
			SigningMethod: "HS256",
		}

		claims := jwt.MapClaims{"user_id": "123"}
		token, err := GenerateToken(config, claims)
		assert.Error(t, err)
		assert.Empty(t, token)
	})

	t.Run("error with unsupported signing method", func(t *testing.T) {
		config := JWTConfig{
			SigningKey:    "test-secret-key",
			SigningMethod: "RS256", // Not supported
		}

		claims := jwt.MapClaims{"user_id": "123"}
		token, err := GenerateToken(config, claims)
		assert.Error(t, err)
		assert.Empty(t, token)
	})

	t.Run("generate with different signing methods", func(t *testing.T) {
		methods := []string{"HS256", "HS384", "HS512"}

		for _, method := range methods {
			config := JWTConfig{
				SigningKey:    "test-secret-key",
				SigningMethod: method,
			}

			claims := jwt.MapClaims{"user_id": "123"}
			token, err := GenerateToken(config, claims)
			require.NoError(t, err, "method: %s", method)
			assert.NotEmpty(t, token)
		}
	})
}

func TestParseToken(t *testing.T) {
	config := JWTConfig{
		SigningKey:    "test-secret-key",
		SigningMethod: "HS256",
		Claims:        jwt.MapClaims{},
	}

	t.Run("parse valid token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id": "123",
			"exp":     time.Now().Add(time.Hour).Unix(),
		}

		tokenString, err := GenerateToken(config, claims)
		require.NoError(t, err)

		token, err := ParseToken(config, tokenString)
		require.NoError(t, err)
		assert.True(t, token.Valid)

		mapClaims, ok := token.Claims.(jwt.MapClaims)
		require.True(t, ok)
		assert.Equal(t, "123", mapClaims["user_id"])
	})

	t.Run("error with expired token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id": "123",
			"exp":     time.Now().Add(-time.Hour).Unix(), // Expired
		}

		tokenString, err := GenerateToken(config, claims)
		require.NoError(t, err)

		token, err := ParseToken(config, tokenString)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrExpiredToken)
		assert.Nil(t, token)
	})

	t.Run("error with invalid token", func(t *testing.T) {
		token, err := ParseToken(config, "invalid.token.string")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, token)
	})

	t.Run("error with wrong signing key", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id": "123",
			"exp":     time.Now().Add(time.Hour).Unix(),
		}

		tokenString, err := GenerateToken(config, claims)
		require.NoError(t, err)

		wrongConfig := config
		wrongConfig.SigningKey = "wrong-key"

		token, err := ParseToken(wrongConfig, tokenString)
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("error when signing key is empty", func(t *testing.T) {
		emptyConfig := JWTConfig{
			SigningMethod: "HS256",
			Claims:        jwt.MapClaims{},
		}

		token, err := ParseToken(emptyConfig, "some.token.string")
		assert.Error(t, err)
		assert.Nil(t, token)
	})
}

func TestExtractToken(t *testing.T) {
	t.Run("extract from header with Bearer scheme", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Bearer test-token-123")

		config := JWTConfig{
			TokenLookup: "header:Authorization",
			AuthScheme:  "Bearer",
		}

		token, err := ExtractToken(ctx, config)
		require.NoError(t, err)
		assert.Equal(t, "test-token-123", token)
	})

	t.Run("extract from header without scheme", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("X-Token", "test-token-123")

		config := JWTConfig{
			TokenLookup: "header:X-Token",
			AuthScheme:  "",
		}

		token, err := ExtractToken(ctx, config)
		require.NoError(t, err)
		assert.Equal(t, "test-token-123", token)
	})

	t.Run("extract from query parameter", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.QueryArgs().Set("token", "test-token-456")

		config := JWTConfig{
			TokenLookup: "query:token",
		}

		token, err := ExtractToken(ctx, config)
		require.NoError(t, err)
		assert.Equal(t, "test-token-456", token)
	})

	t.Run("extract from cookie", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.SetCookie("jwt", "test-token-789")

		config := JWTConfig{
			TokenLookup: "cookie:jwt",
		}

		token, err := ExtractToken(ctx, config)
		require.NoError(t, err)
		assert.Equal(t, "test-token-789", token)
	})

	t.Run("error when token is missing", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}

		config := JWTConfig{
			TokenLookup: "header:Authorization",
			AuthScheme:  "Bearer",
		}

		token, err := ExtractToken(ctx, config)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingToken)
		assert.Empty(t, token)
	})

	t.Run("error with invalid auth scheme", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

		config := JWTConfig{
			TokenLookup: "header:Authorization",
			AuthScheme:  "Bearer",
		}

		token, err := ExtractToken(ctx, config)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Empty(t, token)
	})

	t.Run("error with invalid token lookup format", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}

		config := JWTConfig{
			TokenLookup: "invalid-format",
		}

		token, err := ExtractToken(ctx, config)
		assert.Error(t, err)
		assert.Empty(t, token)
	})

	t.Run("error with unsupported source", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}

		config := JWTConfig{
			TokenLookup: "form:token",
		}

		token, err := ExtractToken(ctx, config)
		assert.Error(t, err)
		assert.Empty(t, token)
	})
}

func TestValidateJWT(t *testing.T) {
	config := JWTConfig{
		SigningKey:    "test-secret-key",
		SigningMethod: "HS256",
		TokenLookup:   "header:Authorization",
		AuthScheme:    "Bearer",
		ContextKey:    "user",
		Claims:        jwt.MapClaims{},
	}

	t.Run("validate valid JWT", func(t *testing.T) {
		claims := jwt.MapClaims{
			"user_id": "123",
			"exp":     time.Now().Add(time.Hour).Unix(),
		}

		tokenString, err := GenerateToken(config, claims)
		require.NoError(t, err)

		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Bearer "+tokenString)

		err = ValidateJWT(ctx, config)
		require.NoError(t, err)

		// Check claims stored in context
		storedClaims := ctx.UserValue("user")
		require.NotNil(t, storedClaims)

		mapClaims, ok := storedClaims.(jwt.MapClaims)
		require.True(t, ok)
		assert.Equal(t, "123", mapClaims["user_id"])
	})

	t.Run("skip validation when skip func returns true", func(t *testing.T) {
		skipConfig := config
		skipConfig.SkipFunc = func(ctx *fasthttp.RequestCtx) bool {
			return true
		}

		ctx := &fasthttp.RequestCtx{}
		// No token set

		err := ValidateJWT(ctx, skipConfig)
		require.NoError(t, err)
	})

	t.Run("call success handler after validation", func(t *testing.T) {
		successCalled := false
		successConfig := config
		successConfig.SuccessHandler = func(ctx *fasthttp.RequestCtx, token *jwt.Token) {
			successCalled = true
			assert.NotNil(t, token)
		}

		claims := jwt.MapClaims{
			"user_id": "123",
			"exp":     time.Now().Add(time.Hour).Unix(),
		}

		tokenString, err := GenerateToken(successConfig, claims)
		require.NoError(t, err)

		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Bearer "+tokenString)

		err = ValidateJWT(ctx, successConfig)
		require.NoError(t, err)
		assert.True(t, successCalled)
	})

	t.Run("error with missing token", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		// No Authorization header

		err := ValidateJWT(ctx, config)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingToken)
	})

	t.Run("error with invalid token", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Bearer invalid.token.here")

		err := ValidateJWT(ctx, config)
		assert.Error(t, err)
	})
}

func TestGetClaims(t *testing.T) {
	t.Run("retrieve map claims from context", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		expectedClaims := jwt.MapClaims{
			"user_id": "123",
			"role":    "admin",
		}

		ctx.SetUserValue("user", expectedClaims)

		claims, ok := GetMapClaims(ctx, "user")
		require.True(t, ok)
		assert.Equal(t, "123", claims["user_id"])
		assert.Equal(t, "admin", claims["role"])
	})

	t.Run("retrieve custom claims from context", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		expectedClaims := NewJWTClaims("user123", "testuser", time.Hour)

		ctx.SetUserValue("user", expectedClaims)

		claims, ok := GetCustomClaims(ctx, "user")
		require.True(t, ok)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
	})

	t.Run("return false when claims not in context", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}

		claims, ok := GetMapClaims(ctx, "user")
		assert.False(t, ok)
		assert.Nil(t, claims)
	})

	t.Run("return false when claims have wrong type", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.SetUserValue("user", "not-a-claim")

		claims, ok := GetMapClaims(ctx, "user")
		assert.False(t, ok)
		assert.Nil(t, claims)
	})
}

func TestNewJWTClaims(t *testing.T) {
	t.Run("create claims with basic fields", func(t *testing.T) {
		claims := NewJWTClaims("user123", "testuser", time.Hour)

		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, "testuser", claims.Username)
		assert.NotNil(t, claims.ExpiresAt)
		assert.NotNil(t, claims.IssuedAt)
		assert.NotNil(t, claims.NotBefore)
	})

	t.Run("expiration is set correctly", func(t *testing.T) {
		duration := 2 * time.Hour
		claims := NewJWTClaims("user123", "testuser", duration)

		expectedExpiry := time.Now().Add(duration)
		actualExpiry := claims.ExpiresAt.Time

		// Allow 1 second tolerance
		diff := actualExpiry.Sub(expectedExpiry)
		assert.Less(t, diff, time.Second)
		assert.Greater(t, diff, -time.Second)
	})

	t.Run("can add roles and extra data", func(t *testing.T) {
		claims := NewJWTClaims("user123", "testuser", time.Hour)
		claims.Roles = []string{"admin", "editor"}
		claims.Extra = map[string]interface{}{
			"department": "engineering",
			"level":      5,
		}

		assert.Len(t, claims.Roles, 2)
		assert.Contains(t, claims.Roles, "admin")
		assert.Equal(t, "engineering", claims.Extra["department"])
	})
}

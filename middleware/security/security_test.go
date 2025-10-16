package security

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetSecurityHeaders(t *testing.T) {
	t.Run("default security headers", func(t *testing.T) {
		headers := GetSecurityHeaders(DefaultSecurityHeadersConfig)

		assert.Equal(t, "1; mode=block", headers["X-XSS-Protection"])
		assert.Equal(t, "nosniff", headers["X-Content-Type-Options"])
		assert.Equal(t, "SAMEORIGIN", headers["X-Frame-Options"])
		assert.Contains(t, headers["Strict-Transport-Security"], "max-age=31536000")
		assert.Contains(t, headers["Strict-Transport-Security"], "includeSubDomains")
		assert.Equal(t, "strict-origin-when-cross-origin", headers["Referrer-Policy"])
	})

	t.Run("custom security headers", func(t *testing.T) {
		config := SecurityHeadersConfig{
			XSSProtection:         "0",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "DENY",
			HSTSMaxAge:            7200,
			HSTSIncludeSubdomains: false,
			ContentSecurityPolicy: "default-src 'self'",
			ReferrerPolicy:        "no-referrer",
			PermissionsPolicy:     "geolocation=()",
		}

		headers := GetSecurityHeaders(config)

		assert.Equal(t, "0", headers["X-XSS-Protection"])
		assert.Equal(t, "DENY", headers["X-Frame-Options"])
		assert.Equal(t, "max-age=7200", headers["Strict-Transport-Security"])
		assert.Equal(t, "default-src 'self'", headers["Content-Security-Policy"])
		assert.Equal(t, "no-referrer", headers["Referrer-Policy"])
		assert.Equal(t, "geolocation=()", headers["Permissions-Policy"])
	})

	t.Run("empty config produces empty headers", func(t *testing.T) {
		config := SecurityHeadersConfig{}
		headers := GetSecurityHeaders(config)

		assert.Empty(t, headers)
	})

	t.Run("HSTS with includeSubDomains", func(t *testing.T) {
		config := SecurityHeadersConfig{
			HSTSMaxAge:            3600,
			HSTSIncludeSubdomains: true,
		}

		headers := GetSecurityHeaders(config)

		assert.Equal(t, "max-age=3600; includeSubDomains", headers["Strict-Transport-Security"])
	})

	t.Run("HSTS without includeSubDomains", func(t *testing.T) {
		config := SecurityHeadersConfig{
			HSTSMaxAge:            3600,
			HSTSIncludeSubdomains: false,
		}

		headers := GetSecurityHeaders(config)

		assert.Equal(t, "max-age=3600", headers["Strict-Transport-Security"])
	})
}

func TestRateLimiter(t *testing.T) {
	t.Run("allow requests within rate limit", func(t *testing.T) {
		limiter := NewRateLimiter(5, time.Minute)

		for i := 0; i < 5; i++ {
			assert.True(t, limiter.Allow("user1"), "request %d should be allowed", i+1)
		}
	})

	t.Run("block requests exceeding rate limit", func(t *testing.T) {
		limiter := NewRateLimiter(3, time.Minute)

		// Allow first 3 requests
		for i := 0; i < 3; i++ {
			assert.True(t, limiter.Allow("user1"))
		}

		// Block 4th request
		assert.False(t, limiter.Allow("user1"))
		assert.False(t, limiter.Allow("user1"))
	})

	t.Run("separate limits for different keys", func(t *testing.T) {
		limiter := NewRateLimiter(2, time.Minute)

		// User1 uses 2 requests
		assert.True(t, limiter.Allow("user1"))
		assert.True(t, limiter.Allow("user1"))
		assert.False(t, limiter.Allow("user1"))

		// User2 still has quota
		assert.True(t, limiter.Allow("user2"))
		assert.True(t, limiter.Allow("user2"))
		assert.False(t, limiter.Allow("user2"))
	})

	t.Run("reset bucket after window expires", func(t *testing.T) {
		limiter := NewRateLimiter(2, 10*time.Millisecond)

		// Use all quota
		assert.True(t, limiter.Allow("user1"))
		assert.True(t, limiter.Allow("user1"))
		assert.False(t, limiter.Allow("user1"))

		// Wait for window to expire
		time.Sleep(15 * time.Millisecond)

		// Should be allowed again
		assert.True(t, limiter.Allow("user1"))
		assert.True(t, limiter.Allow("user1"))
	})

	t.Run("concurrent access to same key", func(t *testing.T) {
		limiter := NewRateLimiter(10, time.Minute)
		results := make(chan bool, 20)

		// Launch 20 concurrent requests
		for i := 0; i < 20; i++ {
			go func() {
				results <- limiter.Allow("user1")
			}()
		}

		// Collect results
		allowed := 0
		denied := 0
		for i := 0; i < 20; i++ {
			if <-results {
				allowed++
			} else {
				denied++
			}
		}

		// Should allow exactly 10 requests
		assert.Equal(t, 10, allowed)
		assert.Equal(t, 10, denied)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("Contains - found", func(t *testing.T) {
		slice := []string{"apple", "banana", "cherry"}
		assert.True(t, Contains(slice, "banana"))
	})

	t.Run("Contains - not found", func(t *testing.T) {
		slice := []string{"apple", "banana", "cherry"}
		assert.False(t, Contains(slice, "orange"))
	})

	t.Run("Contains - empty slice", func(t *testing.T) {
		slice := []string{}
		assert.False(t, Contains(slice, "anything"))
	})

	t.Run("JoinStrings", func(t *testing.T) {
		items := []string{"a", "b", "c"}
		result := JoinStrings(items, ", ")
		assert.Equal(t, "a, b, c", result)
	})

	t.Run("JoinStrings - empty", func(t *testing.T) {
		items := []string{}
		result := JoinStrings(items, ", ")
		assert.Equal(t, "", result)
	})
}

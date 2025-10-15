package lrpc

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lazygophers/log"
	"github.com/valyala/fasthttp"
)

// CORSConfig defines the config for CORS middleware
type CORSConfig struct {
	// AllowOrigins defines allowed origins
	AllowOrigins []string

	// AllowMethods defines allowed HTTP methods
	AllowMethods []string

	// AllowHeaders defines allowed headers
	AllowHeaders []string

	// ExposeHeaders defines exposed headers
	ExposeHeaders []string

	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool

	// MaxAge indicates how long preflight results can be cached
	MaxAge int
}

// DefaultCORSConfig is the default CORS configuration
var DefaultCORSConfig = CORSConfig{
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
	ExposeHeaders:    []string{},
	AllowCredentials: false,
	MaxAge:           3600,
}

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware
func CORS(config ...CORSConfig) HandlerFunc {
	cfg := DefaultCORSConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	allowOrigins := strings.Join(cfg.AllowOrigins, ", ")
	allowMethods := strings.Join(cfg.AllowMethods, ", ")
	allowHeaders := strings.Join(cfg.AllowHeaders, ", ")
	exposeHeaders := strings.Join(cfg.ExposeHeaders, ", ")

	return func(ctx *Ctx) error {
		origin := ctx.Header(HeaderOrigin)

		// Set CORS headers
		if allowOrigins == "*" || contains(cfg.AllowOrigins, origin) {
			if allowOrigins == "*" {
				ctx.SetHeader(HeaderAccessControlAllowOrigin, "*")
			} else {
				ctx.SetHeader(HeaderAccessControlAllowOrigin, origin)
			}
		}

		ctx.SetHeader(HeaderAccessControlAllowMethods, allowMethods)
		ctx.SetHeader(HeaderAccessControlAllowHeaders, allowHeaders)

		if len(exposeHeaders) > 0 {
			ctx.SetHeader(HeaderAccessControlExposeHeaders, exposeHeaders)
		}

		if cfg.AllowCredentials {
			ctx.SetHeader(HeaderAccessControlAllowCredentials, "true")
		}

		if cfg.MaxAge > 0 {
			ctx.SetHeader(HeaderAccessControlMaxAge, strconv.Itoa(cfg.MaxAge))
		}

		// Handle preflight request
		if ctx.Method() == "OPTIONS" {
			ctx.Context().SetStatusCode(fasthttp.StatusNoContent)
			return nil
		}

		return ctx.Next()
	}
}

// SecurityHeadersConfig defines security headers configuration
type SecurityHeadersConfig struct {
	// XSSProtection enables XSS protection header
	XSSProtection string

	// ContentTypeNosniff enables X-Content-Type-Options header
	ContentTypeNosniff string

	// XFrameOptions sets X-Frame-Options header
	XFrameOptions string

	// HSTSMaxAge sets Strict-Transport-Security header
	HSTSMaxAge int

	// HSTSIncludeSubdomains includes subdomains in HSTS
	HSTSIncludeSubdomains bool

	// ContentSecurityPolicy sets CSP header
	ContentSecurityPolicy string

	// ReferrerPolicy sets Referrer-Policy header
	ReferrerPolicy string

	// PermissionsPolicy sets Permissions-Policy header
	PermissionsPolicy string
}

// DefaultSecurityHeadersConfig is the default security headers configuration
var DefaultSecurityHeadersConfig = SecurityHeadersConfig{
	XSSProtection:         "1; mode=block",
	ContentTypeNosniff:    "nosniff",
	XFrameOptions:         "SAMEORIGIN",
	HSTSMaxAge:            31536000, // 1 year
	HSTSIncludeSubdomains: true,
	ReferrerPolicy:        "strict-origin-when-cross-origin",
}

// SecurityHeaders returns a middleware that sets security headers
func SecurityHeaders(config ...SecurityHeadersConfig) HandlerFunc {
	cfg := DefaultSecurityHeadersConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *Ctx) error {
		// X-XSS-Protection
		if cfg.XSSProtection != "" {
			ctx.SetHeader(HeaderXXSSProtection, cfg.XSSProtection)
		}

		// X-Content-Type-Options
		if cfg.ContentTypeNosniff != "" {
			ctx.SetHeader(HeaderXContentTypeOptions, cfg.ContentTypeNosniff)
		}

		// X-Frame-Options
		if cfg.XFrameOptions != "" {
			ctx.SetHeader(HeaderXFrameOptions, cfg.XFrameOptions)
		}

		// Strict-Transport-Security
		if cfg.HSTSMaxAge > 0 {
			hsts := "max-age=" + strconv.Itoa(cfg.HSTSMaxAge)
			if cfg.HSTSIncludeSubdomains {
				hsts += "; includeSubDomains"
			}
			ctx.SetHeader(HeaderStrictTransportSecurity, hsts)
		}

		// Content-Security-Policy
		if cfg.ContentSecurityPolicy != "" {
			ctx.SetHeader(HeaderContentSecurityPolicy, cfg.ContentSecurityPolicy)
		}

		// Referrer-Policy
		if cfg.ReferrerPolicy != "" {
			ctx.SetHeader(HeaderReferrerPolicy, cfg.ReferrerPolicy)
		}

		// Permissions-Policy
		if cfg.PermissionsPolicy != "" {
			ctx.SetHeader(HeaderPermissionsPolicy, cfg.PermissionsPolicy)
		}

		return ctx.Next()
	}
}

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int           // requests per window
	window  time.Duration // time window
}

type bucket struct {
	tokens    int
	lastReset time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:    rl.rate - 1,
			lastReset: now,
		}
		rl.buckets[key] = b
		return true
	}

	// Reset bucket if window has passed
	if now.Sub(b.lastReset) >= rl.window {
		b.tokens = rl.rate - 1
		b.lastReset = now
		return true
	}

	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// RateLimitConfig defines rate limit configuration
type RateLimitConfig struct {
	// Rate is the number of requests allowed per window
	Rate int

	// Window is the time window for rate limiting
	Window time.Duration

	// KeyGenerator generates the key for rate limiting (default: IP address)
	KeyGenerator func(ctx *Ctx) string

	// Handler is called when rate limit is exceeded
	Handler func(ctx *Ctx) error
}

// DefaultRateLimitConfig is the default rate limit configuration
var DefaultRateLimitConfig = RateLimitConfig{
	Rate:   100,
	Window: 1 * time.Minute,
	KeyGenerator: func(ctx *Ctx) string {
		return ctx.Context().RemoteIP().String()
	},
	Handler: func(ctx *Ctx) error {
		ctx.Context().SetStatusCode(fasthttp.StatusTooManyRequests)
		return ctx.SendJson(map[string]interface{}{
			"code":    fasthttp.StatusTooManyRequests,
			"message": "Too Many Requests",
		})
	},
}

// RateLimit returns a rate limiting middleware
func RateLimit(config ...RateLimitConfig) HandlerFunc {
	cfg := DefaultRateLimitConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	limiter := NewRateLimiter(cfg.Rate, cfg.Window)

	return func(ctx *Ctx) error {
		key := cfg.KeyGenerator(ctx)

		if !limiter.Allow(key) {
			log.Warnf("Rate limit exceeded for key: %s", key)
			return cfg.Handler(ctx)
		}

		return ctx.Next()
	}
}

// CSRFConfig defines CSRF protection configuration
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token
	TokenLength int

	// TokenLookup defines where to look for the token
	// Format: "<source>:<name>"
	// Possible values: "header:X-CSRF-Token", "form:csrf", "query:csrf"
	TokenLookup string

	// CookieName is the name of the CSRF cookie
	CookieName string

	// CookiePath is the path of the CSRF cookie
	CookiePath string

	// CookieDomain is the domain of the CSRF cookie
	CookieDomain string

	// CookieSecure indicates if the cookie is secure
	CookieSecure bool

	// CookieHTTPOnly indicates if the cookie is HTTP only
	CookieHTTPOnly bool

	// CookieSameSite sets the SameSite attribute
	CookieSameSite string
}

// DefaultCSRFConfig is the default CSRF configuration
var DefaultCSRFConfig = CSRFConfig{
	TokenLength:    32,
	TokenLookup:    "header:X-CSRF-Token",
	CookieName:     "_csrf",
	CookiePath:     "/",
	CookieSecure:   false,
	CookieHTTPOnly: true,
	CookieSameSite: "Strict",
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

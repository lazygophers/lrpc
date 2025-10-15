package security

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lazygophers/log"
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

// GetSecurityHeaders returns security headers based on config
func GetSecurityHeaders(cfg SecurityHeadersConfig) map[string]string {
	headers := make(map[string]string)

	if cfg.XSSProtection != "" {
		headers["X-XSS-Protection"] = cfg.XSSProtection
	}

	if cfg.ContentTypeNosniff != "" {
		headers["X-Content-Type-Options"] = cfg.ContentTypeNosniff
	}

	if cfg.XFrameOptions != "" {
		headers["X-Frame-Options"] = cfg.XFrameOptions
	}

	if cfg.HSTSMaxAge > 0 {
		hsts := "max-age=" + strconv.Itoa(cfg.HSTSMaxAge)
		if cfg.HSTSIncludeSubdomains {
			hsts += "; includeSubDomains"
		}
		headers["Strict-Transport-Security"] = hsts
	}

	if cfg.ContentSecurityPolicy != "" {
		headers["Content-Security-Policy"] = cfg.ContentSecurityPolicy
	}

	if cfg.ReferrerPolicy != "" {
		headers["Referrer-Policy"] = cfg.ReferrerPolicy
	}

	if cfg.PermissionsPolicy != "" {
		headers["Permissions-Policy"] = cfg.PermissionsPolicy
	}

	return headers
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
}

// DefaultRateLimitConfig is the default rate limit configuration
var DefaultRateLimitConfig = RateLimitConfig{
	Rate:   100,
	Window: 1 * time.Minute,
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
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// JoinStrings joins strings with separator
func JoinStrings(items []string, sep string) string {
	return strings.Join(items, sep)
}

// LogRateLimitExceeded logs when rate limit is exceeded
func LogRateLimitExceeded(key string) {
	log.Warnf("Rate limit exceeded for key: %s", key)
}

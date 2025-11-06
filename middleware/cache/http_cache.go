package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/lazygophers/log"
	"github.com/valyala/fasthttp"
)

// CacheConfig defines HTTP caching configuration
type CacheConfig struct {
	// MaxAge sets the max-age directive in seconds (0 = no cache)
	MaxAge int

	// SMaxAge sets the s-maxage directive for shared caches
	SMaxAge int

	// Public makes the response cacheable by any cache
	Public bool

	// Private makes the response cacheable only by browser cache
	Private bool

	// NoCache forces caches to submit request for validation before releasing cached copy
	NoCache bool

	// NoStore disables caching completely
	NoStore bool

	// MustRevalidate forces caches to obey freshness information
	MustRevalidate bool

	// ProxyRevalidate similar to MustRevalidate but for shared caches only
	ProxyRevalidate bool

	// Immutable indicates resource will never change
	Immutable bool

	// StaleWhileRevalidate allows using stale response while revalidating
	StaleWhileRevalidate int

	// StaleIfError allows using stale response if an error occurs
	StaleIfError int

	// Vary specifies request headers that affect the response
	Vary []string

	// ETagGenerator generates ETag for response body
	// If not set, uses SHA256 hash of response body
	ETagGenerator func(body []byte) string

	// WeakETag uses weak ETag (W/"...") instead of strong ETag
	WeakETag bool

	// SkipFunc allows skipping caching for certain requests
	SkipFunc func(ctx *fasthttp.RequestCtx) bool
}

// DefaultCacheConfig is the default cache configuration
var DefaultCacheConfig = CacheConfig{
	MaxAge:     3600,  // 1 hour
	Public:     true,
	WeakETag:   false,
	Vary:       []string{"Accept-Encoding"},
}

// BuildCacheControl builds Cache-Control header value from config
func BuildCacheControl(config CacheConfig) string {
	var directives []string

	if config.NoStore {
		directives = append(directives, "no-store")
		return joinDirectives(directives)
	}

	if config.NoCache {
		directives = append(directives, "no-cache")
	}

	if config.Public {
		directives = append(directives, "public")
	} else if config.Private {
		directives = append(directives, "private")
	}

	if config.MaxAge > 0 {
		directives = append(directives, fmt.Sprintf("max-age=%d", config.MaxAge))
	}

	if config.SMaxAge > 0 {
		directives = append(directives, fmt.Sprintf("s-maxage=%d", config.SMaxAge))
	}

	if config.MustRevalidate {
		directives = append(directives, "must-revalidate")
	}

	if config.ProxyRevalidate {
		directives = append(directives, "proxy-revalidate")
	}

	if config.Immutable {
		directives = append(directives, "immutable")
	}

	if config.StaleWhileRevalidate > 0 {
		directives = append(directives, fmt.Sprintf("stale-while-revalidate=%d", config.StaleWhileRevalidate))
	}

	if config.StaleIfError > 0 {
		directives = append(directives, fmt.Sprintf("stale-if-error=%d", config.StaleIfError))
	}

	return joinDirectives(directives)
}

func joinDirectives(directives []string) string {
	if len(directives) == 0 {
		return ""
	}

	result := directives[0]
	for i := 1; i < len(directives); i++ {
		result += ", " + directives[i]
	}
	return result
}

// GenerateETag generates an ETag for the response body
func GenerateETag(body []byte, weak bool) string {
	hash := sha256.Sum256(body)
	etag := hex.EncodeToString(hash[:])

	if weak {
		return `W/"` + etag + `"`
	}
	return `"` + etag + `"`
}

// CheckETag checks if the request ETag matches the generated ETag
func CheckETag(ctx *fasthttp.RequestCtx, etag string) bool {
	// Check If-None-Match header
	ifNoneMatch := string(ctx.Request.Header.Peek("If-None-Match"))
	if ifNoneMatch == "" {
		return false
	}

	// Check for exact match or wildcard
	if ifNoneMatch == "*" || ifNoneMatch == etag {
		return true
	}

	// Check for multiple ETags (comma-separated)
	// Note: This is a simplified check, proper implementation should handle quoted strings
	// For now, we'll just check if our etag is contained in the header
	return contains(ifNoneMatch, etag)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)+1] == substr+"," || s[len(s)-len(substr)-1:] == ","+substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	target := "," + substr + ","
	for i := 0; i <= len(s)-len(target); i++ {
		if s[i:i+len(target)] == target {
			return true
		}
	}
	return false
}

// CheckModifiedSince checks if resource was modified since the given time
func CheckModifiedSince(ctx *fasthttp.RequestCtx, modTime time.Time) bool {
	ifModifiedSince := string(ctx.Request.Header.Peek("If-Modified-Since"))
	if ifModifiedSince == "" {
		return true
	}

	t, err := time.Parse(time.RFC1123, ifModifiedSince)
	if err != nil {
		log.Errorf("err:%v", err)
		return true
	}

	// Round to second precision for comparison
	modTime = modTime.Truncate(time.Second)
	t = t.Truncate(time.Second)

	return modTime.After(t)
}

// SetCacheHeaders sets HTTP caching headers on the response
func SetCacheHeaders(ctx *fasthttp.RequestCtx, config CacheConfig) {
	// Skip if skip function returns true
	if config.SkipFunc != nil && config.SkipFunc(ctx) {
		return
	}

	// Set Cache-Control header
	cacheControl := BuildCacheControl(config)
	if cacheControl != "" {
		ctx.Response.Header.Set("Cache-Control", cacheControl)
	}

	// Set Vary header
	if len(config.Vary) > 0 {
		vary := config.Vary[0]
		for i := 1; i < len(config.Vary); i++ {
			vary += ", " + config.Vary[i]
		}
		ctx.Response.Header.Set("Vary", vary)
	}

	// Set ETag if not NoStore and response body exists
	if !config.NoStore {
		body := ctx.Response.Body()
		if len(body) > 0 {
			var etag string
			if config.ETagGenerator != nil {
				etag = config.ETagGenerator(body)
			} else {
				etag = GenerateETag(body, config.WeakETag)
			}

			ctx.Response.Header.Set("ETag", etag)

			// Check if client has cached version
			if CheckETag(ctx, etag) {
				ctx.SetStatusCode(fasthttp.StatusNotModified)
				ctx.Response.SetBody(nil)
			}
		}
	}

	// Set Last-Modified header (current time)
	if !config.NoStore && ctx.Response.Header.Peek("Last-Modified") == nil {
		ctx.Response.Header.Set("Last-Modified", time.Now().UTC().Format(time.RFC1123))
	}
}

// NoCache disables all caching
func NoCache() CacheConfig {
	return CacheConfig{
		NoStore: true,
	}
}

// ShortCache creates a config for short-term caching (5 minutes)
func ShortCache() CacheConfig {
	return CacheConfig{
		MaxAge: 300,
		Public: true,
	}
}

// MediumCache creates a config for medium-term caching (1 hour)
func MediumCache() CacheConfig {
	return CacheConfig{
		MaxAge: 3600,
		Public: true,
	}
}

// LongCache creates a config for long-term caching (1 day)
func LongCache() CacheConfig {
	return CacheConfig{
		MaxAge: 86400,
		Public: true,
	}
}

// ImmutableCache creates a config for immutable resources (1 year)
func ImmutableCache() CacheConfig {
	return CacheConfig{
		MaxAge:    31536000, // 1 year
		Public:    true,
		Immutable: true,
	}
}

// PrivateCache creates a config for private (browser-only) caching
func PrivateCache(maxAge int) CacheConfig {
	return CacheConfig{
		MaxAge:  maxAge,
		Private: true,
	}
}

// ConditionalRequest checks conditional request headers and returns appropriate response
func ConditionalRequest(ctx *fasthttp.RequestCtx, etag string, modTime time.Time) bool {
	// Check ETag first
	if etag != "" && CheckETag(ctx, etag) {
		ctx.SetStatusCode(fasthttp.StatusNotModified)
		ctx.Response.SetBody(nil)
		ctx.Response.Header.Set("ETag", etag)
		return true
	}

	// Check Last-Modified
	if !modTime.IsZero() && !CheckModifiedSince(ctx, modTime) {
		ctx.SetStatusCode(fasthttp.StatusNotModified)
		ctx.Response.SetBody(nil)
		ctx.Response.Header.Set("Last-Modified", modTime.UTC().Format(time.RFC1123))
		return true
	}

	return false
}

// ParseCacheControl parses Cache-Control header into a map
func ParseCacheControl(cacheControl string) map[string]string {
	result := make(map[string]string)
	if cacheControl == "" {
		return result
	}

	directives := splitDirectives(cacheControl)
	for _, directive := range directives {
		var key, value string
		if idx := findChar(directive, '='); idx >= 0 {
			key = trimSpace(directive[:idx])
			value = trimSpace(directive[idx+1:])
		} else {
			key = trimSpace(directive)
			value = "true"
		}
		result[key] = value
	}

	return result
}

func splitDirectives(s string) []string {
	var result []string
	var current string
	inQuotes := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuotes = !inQuotes
			current += string(c)
		} else if c == ',' && !inQuotes {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

func findChar(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}

// GetCacheControlMaxAge extracts max-age from Cache-Control header
func GetCacheControlMaxAge(ctx *fasthttp.RequestCtx) (int, error) {
	cacheControl := string(ctx.Request.Header.Peek("Cache-Control"))
	directives := ParseCacheControl(cacheControl)

	maxAgeStr, exists := directives["max-age"]
	if !exists {
		return 0, fmt.Errorf("max-age not found")
	}

	maxAge, err := strconv.Atoi(maxAgeStr)
	if err != nil {
		log.Errorf("err:%v", err)
		return 0, err
	}

	return maxAge, nil
}

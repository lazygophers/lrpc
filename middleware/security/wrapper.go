package security

import (
	"time"

	"github.com/lazygophers/lrpc"
	"github.com/valyala/fasthttp"
)

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware
func CORS(config ...CORSConfig) func(ctx *lrpc.Ctx) error {
	cfg := DefaultCORSConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	allowOrigins := JoinStrings(cfg.AllowOrigins, ", ")
	allowMethods := JoinStrings(cfg.AllowMethods, ", ")
	allowHeaders := JoinStrings(cfg.AllowHeaders, ", ")
	exposeHeaders := JoinStrings(cfg.ExposeHeaders, ", ")

	return func(ctx *lrpc.Ctx) error {
		origin := ctx.Header(lrpc.HeaderOrigin)

		// Set CORS headers
		if allowOrigins == "*" || Contains(cfg.AllowOrigins, origin) {
			if allowOrigins == "*" {
				ctx.SetHeader(lrpc.HeaderAccessControlAllowOrigin, "*")
			} else {
				ctx.SetHeader(lrpc.HeaderAccessControlAllowOrigin, origin)
			}
		}

		ctx.SetHeader(lrpc.HeaderAccessControlAllowMethods, allowMethods)
		ctx.SetHeader(lrpc.HeaderAccessControlAllowHeaders, allowHeaders)

		if len(exposeHeaders) > 0 {
			ctx.SetHeader(lrpc.HeaderAccessControlExposeHeaders, exposeHeaders)
		}

		if cfg.AllowCredentials {
			ctx.SetHeader(lrpc.HeaderAccessControlAllowCredentials, "true")
		}

		if cfg.MaxAge > 0 {
			ctx.SetHeader(lrpc.HeaderAccessControlMaxAge, string(rune(cfg.MaxAge)))
		}

		// Handle preflight request
		if ctx.Method() == "OPTIONS" {
			ctx.Context().SetStatusCode(fasthttp.StatusNoContent)
			return nil
		}

		return ctx.Next()
	}
}

// SecurityHeaders returns a middleware that sets security headers
func SecurityHeaders(config ...SecurityHeadersConfig) func(ctx *lrpc.Ctx) error {
	cfg := DefaultSecurityHeadersConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	headers := GetSecurityHeaders(cfg)

	return func(ctx *lrpc.Ctx) error {
		for key, value := range headers {
			ctx.SetHeader(key, value)
		}
		return ctx.Next()
	}
}

// RateLimitMiddlewareConfig defines rate limit middleware configuration
type RateLimitMiddlewareConfig struct {
	// Rate is the number of requests allowed per window
	Rate int

	// Window is the time window for rate limiting
	Window time.Duration

	// KeyGenerator generates the key for rate limiting (default: IP address)
	KeyGenerator func(ctx *lrpc.Ctx) string

	// Handler is called when rate limit is exceeded
	Handler func(ctx *lrpc.Ctx) error
}

// DefaultRateLimitMiddlewareConfig is the default rate limit middleware configuration
var DefaultRateLimitMiddlewareConfig = RateLimitMiddlewareConfig{
	Rate:   100,
	Window: 1 * time.Minute,
	KeyGenerator: func(ctx *lrpc.Ctx) string {
		return ctx.Context().RemoteIP().String()
	},
	Handler: func(ctx *lrpc.Ctx) error {
		ctx.Context().SetStatusCode(fasthttp.StatusTooManyRequests)
		return ctx.SendJson(map[string]interface{}{
			"code":    fasthttp.StatusTooManyRequests,
			"message": "Too Many Requests",
		})
	},
}

// RateLimit returns a rate limiting middleware
func RateLimit(config ...RateLimitMiddlewareConfig) func(ctx *lrpc.Ctx) error {
	cfg := DefaultRateLimitMiddlewareConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	limiter := NewRateLimiter(cfg.Rate, cfg.Window)

	return func(ctx *lrpc.Ctx) error {
		key := cfg.KeyGenerator(ctx)

		if !limiter.Allow(key) {
			LogRateLimitExceeded(key)
			return cfg.Handler(ctx)
		}

		return ctx.Next()
	}
}

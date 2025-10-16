package lrpc

import (
	"runtime/debug"
	"time"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/auth"
	"github.com/lazygophers/lrpc/middleware/cache"
	"github.com/lazygophers/lrpc/middleware/compress"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/health"
	"github.com/lazygophers/lrpc/middleware/metrics"
	"github.com/lazygophers/lrpc/middleware/security"
	"github.com/valyala/fasthttp"
)

// RecoverConfig defines the config for Recover middleware
type RecoverConfig struct {
	// EnableStackTrace enables printing stack trace
	EnableStackTrace bool

	// StackTraceHandler allows custom stack trace handling
	StackTraceHandler func(ctx *Ctx, err interface{}, stack []byte)

	// ErrorHandler allows custom error response
	ErrorHandler func(ctx *Ctx, err interface{})
}

// DefaultRecoverConfig is the default Recover middleware config
var DefaultRecoverConfig = RecoverConfig{
	EnableStackTrace: true,
	StackTraceHandler: func(ctx *Ctx, err interface{}, stack []byte) {
		log.Errorf("[Panic Recovered] %v\nStack Trace:\n%s", err, string(stack))
	},
	ErrorHandler: nil,
}

// Recover returns a middleware that recovers from panics
func Recover(config ...RecoverConfig) HandlerFunc {
	cfg := DefaultRecoverConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				var stack []byte
				if cfg.EnableStackTrace {
					stack = debug.Stack()
					if cfg.StackTraceHandler != nil {
						cfg.StackTraceHandler(ctx, r, stack)
					}
				}

				if cfg.ErrorHandler != nil {
					cfg.ErrorHandler(ctx, r)
					return
				}

				// Default error handling
				ctx.Context().SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SendJson(map[string]interface{}{
					"code":    int32(core.ErrCode_StatusInternalServerError),
					"message": "Internal Server Error",
				})
			}
		}()

		return ctx.Next()
	}
}

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware
func CORS(config ...security.CORSConfig) HandlerFunc {
	cfg := security.DefaultCORSConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	allowOrigins := security.JoinStrings(cfg.AllowOrigins, ", ")
	allowMethods := security.JoinStrings(cfg.AllowMethods, ", ")
	allowHeaders := security.JoinStrings(cfg.AllowHeaders, ", ")
	exposeHeaders := security.JoinStrings(cfg.ExposeHeaders, ", ")

	return func(ctx *Ctx) error {
		origin := ctx.Header(HeaderOrigin)

		// Set CORS headers
		if allowOrigins == "*" || security.Contains(cfg.AllowOrigins, origin) {
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
			ctx.SetHeader(HeaderAccessControlMaxAge, string(rune(cfg.MaxAge)))
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
func SecurityHeaders(config ...security.SecurityHeadersConfig) HandlerFunc {
	cfg := security.DefaultSecurityHeadersConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	headers := security.GetSecurityHeaders(cfg)

	return func(ctx *Ctx) error {
		for key, value := range headers {
			ctx.SetHeader(key, value)
		}
		return ctx.Next()
	}
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

	limiter := security.NewRateLimiter(cfg.Rate, cfg.Window)

	return func(ctx *Ctx) error {
		key := cfg.KeyGenerator(ctx)

		if !limiter.Allow(key) {
			security.LogRateLimitExceeded(key)
			return cfg.Handler(ctx)
		}

		return ctx.Next()
	}
}

// MetricsMiddleware returns a middleware that collects metrics
func MetricsMiddleware(collector *metrics.Collector) HandlerFunc {
	return func(ctx *Ctx) error {
		collector.IncrementRequest(ctx.Method())

		// Execute next handlers
		err := ctx.Next()

		collector.DecrementInFlight()

		if err != nil {
			collector.IncrementError()
		}
		collector.IncrementResponse()

		return err
	}
}

// SlowRequestLogger returns a middleware that logs slow requests
func SlowRequestLogger(config ...metrics.SlowRequestConfig) HandlerFunc {
	cfg := metrics.DefaultSlowRequestConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *Ctx) error {
		start := time.Now()

		// Execute next handlers
		err := ctx.Next()

		// Calculate duration
		duration := time.Since(start)

		// Check if request is slow
		if duration >= cfg.Threshold {
			metrics.LogSlowRequest(cfg, ctx.Method(), ctx.Path(), ctx.TraceID(), duration)
		}

		return err
	}
}

// Compress returns a middleware that compresses HTTP responses
func Compress(config ...compress.Config) HandlerFunc {
	cfg := compress.DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *Ctx) error {
		// Execute next handlers
		err := ctx.Next()

		// Compress response if applicable
		compress.CompressResponse(ctx.Context(), cfg)

		return err
	}
}

// BodyLimit returns a middleware that limits request body size
func BodyLimit(maxSize int) HandlerFunc {
	return func(ctx *Ctx) error {
		// Check body size
		if ctx.Context().Request.Header.ContentLength() > maxSize {
			ctx.Context().SetStatusCode(fasthttp.StatusRequestEntityTooLarge)
			return ctx.SendJson(map[string]interface{}{
				"code":    fasthttp.StatusRequestEntityTooLarge,
				"message": "Request Entity Too Large",
			})
		}

		return ctx.Next()
	}
}

// Stream creates a stream writer for the context
func (ctx *Ctx) Stream() *compress.StreamWriter {
	return compress.NewStreamWriter(ctx.Context())
}

// JWT returns a JWT authentication middleware
func JWT(config auth.JWTConfig) HandlerFunc {
	// Set defaults
	if config.SigningMethod == "" {
		config.SigningMethod = auth.DefaultJWTConfig.SigningMethod
	}
	if config.TokenLookup == "" {
		config.TokenLookup = auth.DefaultJWTConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = auth.DefaultJWTConfig.AuthScheme
	}
	if config.ContextKey == "" {
		config.ContextKey = auth.DefaultJWTConfig.ContextKey
	}
	if config.Claims == nil {
		config.Claims = auth.DefaultJWTConfig.Claims
	}

	// Default error handler
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(`{"code":401,"message":"Unauthorized"}`)
		}
	}

	return func(ctx *Ctx) error {
		err := auth.ValidateJWT(ctx.Context(), config)
		if err != nil {
			config.ErrorHandler(ctx.Context(), err)
			return err
		}
		return ctx.Next()
	}
}

// BasicAuth returns a basic authentication middleware
func BasicAuth(config auth.BasicAuthConfig) HandlerFunc {
	// Set defaults
	if config.Realm == "" {
		config.Realm = auth.DefaultBasicAuthConfig.Realm
	}
	if config.ContextKey == "" {
		config.ContextKey = auth.DefaultBasicAuthConfig.ContextKey
	}

	// Default error handler
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
			auth.SetWWWAuthenticate(ctx, config.Realm)
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(`{"code":401,"message":"Unauthorized"}`)
		}
	}

	return func(ctx *Ctx) error {
		err := auth.ValidateBasicAuth(ctx.Context(), config)
		if err != nil {
			config.ErrorHandler(ctx.Context(), err)
			return err
		}
		return ctx.Next()
	}
}

// Cache returns an HTTP caching middleware
func Cache(config ...cache.CacheConfig) HandlerFunc {
	cfg := cache.DefaultCacheConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *Ctx) error {
		// Execute next handlers
		err := ctx.Next()

		// Set cache headers if no error
		if err == nil {
			cache.SetCacheHeaders(ctx.Context(), cfg)
		}

		return err
	}
}

// CacheWithETag returns a caching middleware that uses ETags
func CacheWithETag(maxAge int, weak bool) HandlerFunc {
	return Cache(cache.CacheConfig{
		MaxAge:   maxAge,
		Public:   true,
		WeakETag: weak,
	})
}

// AddHealthEndpoints adds standard health check endpoints to the app
func (app *App) AddHealthEndpoints(prefix string, checker *health.Checker) error {
	if prefix == "" {
		prefix = "/"
	}

	// Liveness probe
	err := app.GET(prefix+"health", func(ctx *Ctx) error {
		ctx.Context().SetStatusCode(fasthttp.StatusOK)
		return ctx.SendJson(map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})
	if err != nil {
		return err
	}

	// Readiness probe
	err = app.GET(prefix+"ready", func(ctx *Ctx) error {
		if checker != nil && !checker.IsReady() {
			ctx.Context().SetStatusCode(fasthttp.StatusServiceUnavailable)
			return ctx.SendJson(map[string]interface{}{
				"status": "not ready",
				"time":   time.Now().Unix(),
			})
		}

		ctx.Context().SetStatusCode(fasthttp.StatusOK)
		return ctx.SendJson(map[string]interface{}{
			"status": "ready",
			"time":   time.Now().Unix(),
		})
	})
	if err != nil {
		return err
	}

	// Detailed health check endpoint
	if checker != nil {
		err = app.GET(prefix+"healthz", func(ctx *Ctx) error {
			results := checker.RunChecks()

			status := results["status"]
			if status == health.StatusUnhealthy {
				ctx.Context().SetStatusCode(fasthttp.StatusServiceUnavailable)
			} else {
				ctx.Context().SetStatusCode(fasthttp.StatusOK)
			}

			return ctx.SendJson(results)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// AddMetricsEndpoint adds a metrics endpoint to the app
func (app *App) AddMetricsEndpoint(path string, collector *metrics.Collector) error {
	return app.GET(path, func(ctx *Ctx) error {
		m := collector.GetMetrics()
		return ctx.SendJson(map[string]interface{}{
			"total_requests":     m.TotalRequests,
			"total_responses":    m.TotalResponses,
			"total_errors":       m.TotalErrors,
			"total_panics":       m.TotalPanics,
			"requests_in_flight": m.RequestsInFlight,
			"by_method": map[string]interface{}{
				"get":    m.GetRequests,
				"post":   m.PostRequests,
				"put":    m.PutRequests,
				"delete": m.DeleteRequests,
				"patch":  m.PatchRequests,
				"other":  m.OtherRequests,
			},
		})
	})
}

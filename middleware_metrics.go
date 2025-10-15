package lrpc

import (
	"sync/atomic"
	"time"

	"github.com/lazygophers/log"
)

// Metrics holds request metrics
type Metrics struct {
	TotalRequests   uint64
	TotalResponses  uint64
	TotalErrors     uint64
	TotalPanics     uint64
	RequestsInFlight uint64

	// Method-specific counters
	GetRequests    uint64
	PostRequests   uint64
	PutRequests    uint64
	DeleteRequests uint64
	PatchRequests  uint64
	OtherRequests  uint64
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	metrics Metrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// GetMetrics returns a snapshot of current metrics
func (m *MetricsCollector) GetMetrics() Metrics {
	return Metrics{
		TotalRequests:    atomic.LoadUint64(&m.metrics.TotalRequests),
		TotalResponses:   atomic.LoadUint64(&m.metrics.TotalResponses),
		TotalErrors:      atomic.LoadUint64(&m.metrics.TotalErrors),
		TotalPanics:      atomic.LoadUint64(&m.metrics.TotalPanics),
		RequestsInFlight: atomic.LoadUint64(&m.metrics.RequestsInFlight),
		GetRequests:      atomic.LoadUint64(&m.metrics.GetRequests),
		PostRequests:     atomic.LoadUint64(&m.metrics.PostRequests),
		PutRequests:      atomic.LoadUint64(&m.metrics.PutRequests),
		DeleteRequests:   atomic.LoadUint64(&m.metrics.DeleteRequests),
		PatchRequests:    atomic.LoadUint64(&m.metrics.PatchRequests),
		OtherRequests:    atomic.LoadUint64(&m.metrics.OtherRequests),
	}
}

// MetricsMiddleware returns a middleware that collects metrics
func MetricsMiddleware(collector *MetricsCollector) HandlerFunc {
	return func(ctx *Ctx) error {
		// Increment total requests
		atomic.AddUint64(&collector.metrics.TotalRequests, 1)
		atomic.AddUint64(&collector.metrics.RequestsInFlight, 1)

		// Track method-specific requests
		method := ctx.Method()
		switch method {
		case "GET":
			atomic.AddUint64(&collector.metrics.GetRequests, 1)
		case "POST":
			atomic.AddUint64(&collector.metrics.PostRequests, 1)
		case "PUT":
			atomic.AddUint64(&collector.metrics.PutRequests, 1)
		case "DELETE":
			atomic.AddUint64(&collector.metrics.DeleteRequests, 1)
		case "PATCH":
			atomic.AddUint64(&collector.metrics.PatchRequests, 1)
		default:
			atomic.AddUint64(&collector.metrics.OtherRequests, 1)
		}

		// Execute next handlers
		err := ctx.Next()

		// Decrement in-flight counter
		atomic.AddUint64(&collector.metrics.RequestsInFlight, ^uint64(0)) // Decrement by 1

		// Track responses and errors
		if err != nil {
			atomic.AddUint64(&collector.metrics.TotalErrors, 1)
		}
		atomic.AddUint64(&collector.metrics.TotalResponses, 1)

		return err
	}
}

// SlowRequestConfig defines configuration for slow request logging
type SlowRequestConfig struct {
	// Threshold is the duration after which a request is considered slow
	Threshold time.Duration

	// LogHandler is called for slow requests
	LogHandler func(ctx *Ctx, duration time.Duration)

	// IncludePath enables logging the request path
	IncludePath bool

	// IncludeMethod enables logging the request method
	IncludeMethod bool

	// IncludeHeaders enables logging request headers
	IncludeHeaders bool
}

// DefaultSlowRequestConfig is the default configuration
var DefaultSlowRequestConfig = SlowRequestConfig{
	Threshold:      1 * time.Second,
	LogHandler:     nil,
	IncludePath:    true,
	IncludeMethod:  true,
	IncludeHeaders: false,
}

// SlowRequestLogger returns a middleware that logs slow requests
func SlowRequestLogger(config ...SlowRequestConfig) HandlerFunc {
	cfg := DefaultSlowRequestConfig
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
			if cfg.LogHandler != nil {
				cfg.LogHandler(ctx, duration)
			} else {
				// Default logging
				logFields := []interface{}{
					"duration", duration.String(),
					"threshold", cfg.Threshold.String(),
					"traceID", ctx.TraceID(),
				}

				if cfg.IncludeMethod {
					logFields = append(logFields, "method", ctx.Method())
				}

				if cfg.IncludePath {
					logFields = append(logFields, "path", ctx.Path())
				}

				if cfg.IncludeHeaders {
					logFields = append(logFields, "content-type", ctx.Header(HeaderContentType))
				}

				log.Warnf("Slow request detected: %v", logFields...)
			}
		}

		return err
	}
}

// RequestTimer returns a middleware that logs request duration
func RequestTimer() HandlerFunc {
	return func(ctx *Ctx) error {
		start := time.Now()

		// Execute next handlers
		err := ctx.Next()

		// Log duration
		duration := time.Since(start)
		log.Infof("Request completed: method=%s path=%s duration=%s traceID=%s",
			ctx.Method(), ctx.Path(), duration.String(), ctx.TraceID())

		return err
	}
}

// AddMetricsEndpoint adds a metrics endpoint to the app
func (app *App) AddMetricsEndpoint(path string, collector *MetricsCollector) error {
	return app.GET(path, func(ctx *Ctx) error {
		metrics := collector.GetMetrics()
		return ctx.SendJson(map[string]interface{}{
			"total_requests":     metrics.TotalRequests,
			"total_responses":    metrics.TotalResponses,
			"total_errors":       metrics.TotalErrors,
			"total_panics":       metrics.TotalPanics,
			"requests_in_flight": metrics.RequestsInFlight,
			"by_method": map[string]interface{}{
				"get":    metrics.GetRequests,
				"post":   metrics.PostRequests,
				"put":    metrics.PutRequests,
				"delete": metrics.DeleteRequests,
				"patch":  metrics.PatchRequests,
				"other":  metrics.OtherRequests,
			},
		})
	})
}

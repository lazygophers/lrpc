package metrics

import (
	"time"

	"github.com/lazygophers/lrpc"
)

// MetricsMiddleware returns a middleware that collects metrics
func MetricsMiddleware(collector *Collector) func(ctx *lrpc.Ctx) error {
	return func(ctx *lrpc.Ctx) error {
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
func SlowRequestLogger(config ...SlowRequestConfig) func(ctx *lrpc.Ctx) error {
	cfg := DefaultSlowRequestConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *lrpc.Ctx) error {
		start := time.Now()

		// Execute next handlers
		err := ctx.Next()

		// Calculate duration
		duration := time.Since(start)

		// Check if request is slow
		if duration >= cfg.Threshold {
			LogSlowRequest(cfg, ctx.Method(), ctx.Path(), ctx.TraceID(), duration)
		}

		return err
	}
}

// AddMetricsEndpoint adds a metrics endpoint to the app
func AddMetricsEndpoint(app interface{ GET(path string, handler func(*lrpc.Ctx) error) error }, path string, collector *Collector) error {
	return app.GET(path, func(ctx *lrpc.Ctx) error {
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

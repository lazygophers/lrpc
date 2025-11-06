package metrics

import (
	"sync/atomic"
	"time"

	"github.com/lazygophers/log"
)

// Metrics holds request metrics
type Metrics struct {
	TotalRequests    uint64
	TotalResponses   uint64
	TotalErrors      uint64
	TotalPanics      uint64
	RequestsInFlight uint64

	// Method-specific counters
	GetRequests    uint64
	PostRequests   uint64
	PutRequests    uint64
	DeleteRequests uint64
	PatchRequests  uint64
	OtherRequests  uint64
}

// Collector collects and manages metrics
type Collector struct {
	metrics Metrics
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{}
}

// GetMetrics returns a snapshot of current metrics
func (m *Collector) GetMetrics() Metrics {
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

// IncrementRequest increments request counters
func (m *Collector) IncrementRequest(method string) {
	atomic.AddUint64(&m.metrics.TotalRequests, 1)
	atomic.AddUint64(&m.metrics.RequestsInFlight, 1)

	switch method {
	case "GET":
		atomic.AddUint64(&m.metrics.GetRequests, 1)
	case "POST":
		atomic.AddUint64(&m.metrics.PostRequests, 1)
	case "PUT":
		atomic.AddUint64(&m.metrics.PutRequests, 1)
	case "DELETE":
		atomic.AddUint64(&m.metrics.DeleteRequests, 1)
	case "PATCH":
		atomic.AddUint64(&m.metrics.PatchRequests, 1)
	default:
		atomic.AddUint64(&m.metrics.OtherRequests, 1)
	}
}

// DecrementInFlight decrements in-flight counter
func (m *Collector) DecrementInFlight() {
	atomic.AddUint64(&m.metrics.RequestsInFlight, ^uint64(0))
}

// IncrementResponse increments response counter
func (m *Collector) IncrementResponse() {
	atomic.AddUint64(&m.metrics.TotalResponses, 1)
}

// IncrementError increments error counter
func (m *Collector) IncrementError() {
	atomic.AddUint64(&m.metrics.TotalErrors, 1)
}

// IncrementPanic increments panic counter
func (m *Collector) IncrementPanic() {
	atomic.AddUint64(&m.metrics.TotalPanics, 1)
}

// SlowRequestConfig defines configuration for slow request logging
type SlowRequestConfig struct {
	// Threshold is the duration after which a request is considered slow
	Threshold time.Duration

	// LogHandler is called for slow requests
	LogHandler func(method, path, traceID string, duration time.Duration)

	// IncludePath enables logging the request path
	IncludePath bool

	// IncludeMethod enables logging the request method
	IncludeMethod bool
}

// DefaultSlowRequestConfig is the default configuration
var DefaultSlowRequestConfig = SlowRequestConfig{
	Threshold:     1 * time.Second,
	LogHandler:    nil,
	IncludePath:   true,
	IncludeMethod: true,
}

// LogSlowRequest logs a slow request
func LogSlowRequest(config SlowRequestConfig, method, path, traceID string, duration time.Duration) {
	if config.LogHandler != nil {
		config.LogHandler(method, path, traceID, duration)
	} else {
		// Default logging
		logFields := []interface{}{
			"duration", duration.String(),
			"threshold", config.Threshold.String(),
			"traceID", traceID,
		}

		if config.IncludeMethod {
			logFields = append(logFields, "method", method)
		}

		if config.IncludePath {
			logFields = append(logFields, "path", path)
		}

		log.Warnf("Slow request detected: %v", logFields...)
	}
}

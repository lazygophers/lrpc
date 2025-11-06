package metrics

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	t.Run("create new collector", func(t *testing.T) {
		collector := NewCollector()
		assert.NotNil(t, collector)

		metrics := collector.GetMetrics()
		assert.Equal(t, uint64(0), metrics.TotalRequests)
		assert.Equal(t, uint64(0), metrics.TotalResponses)
	})

	t.Run("increment request counters", func(t *testing.T) {
		collector := NewCollector()

		collector.IncrementRequest("GET")
		collector.IncrementRequest("POST")
		collector.IncrementRequest("PUT")

		metrics := collector.GetMetrics()
		assert.Equal(t, uint64(3), metrics.TotalRequests)
		assert.Equal(t, uint64(1), metrics.GetRequests)
		assert.Equal(t, uint64(1), metrics.PostRequests)
		assert.Equal(t, uint64(1), metrics.PutRequests)
		assert.Equal(t, uint64(3), metrics.RequestsInFlight)
	})

	t.Run("increment specific HTTP methods", func(t *testing.T) {
		collector := NewCollector()

		collector.IncrementRequest("GET")
		collector.IncrementRequest("GET")
		collector.IncrementRequest("POST")
		collector.IncrementRequest("DELETE")
		collector.IncrementRequest("PATCH")
		collector.IncrementRequest("OPTIONS") // Other method

		metrics := collector.GetMetrics()
		assert.Equal(t, uint64(2), metrics.GetRequests)
		assert.Equal(t, uint64(1), metrics.PostRequests)
		assert.Equal(t, uint64(1), metrics.DeleteRequests)
		assert.Equal(t, uint64(1), metrics.PatchRequests)
		assert.Equal(t, uint64(1), metrics.OtherRequests)
	})

	t.Run("decrement in-flight requests", func(t *testing.T) {
		collector := NewCollector()

		collector.IncrementRequest("GET")
		collector.IncrementRequest("POST")
		assert.Equal(t, uint64(2), collector.GetMetrics().RequestsInFlight)

		collector.DecrementInFlight()
		assert.Equal(t, uint64(1), collector.GetMetrics().RequestsInFlight)

		collector.DecrementInFlight()
		assert.Equal(t, uint64(0), collector.GetMetrics().RequestsInFlight)
	})

	t.Run("increment response counter", func(t *testing.T) {
		collector := NewCollector()

		collector.IncrementResponse()
		collector.IncrementResponse()
		collector.IncrementResponse()

		metrics := collector.GetMetrics()
		assert.Equal(t, uint64(3), metrics.TotalResponses)
	})

	t.Run("increment error counter", func(t *testing.T) {
		collector := NewCollector()

		collector.IncrementError()
		collector.IncrementError()

		metrics := collector.GetMetrics()
		assert.Equal(t, uint64(2), metrics.TotalErrors)
	})

	t.Run("increment panic counter", func(t *testing.T) {
		collector := NewCollector()

		collector.IncrementPanic()

		metrics := collector.GetMetrics()
		assert.Equal(t, uint64(1), metrics.TotalPanics)
	})

	t.Run("concurrent operations are thread-safe", func(t *testing.T) {
		collector := NewCollector()
		var wg sync.WaitGroup
		iterations := 100

		// Concurrent requests
		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				collector.IncrementRequest("GET")
				collector.IncrementResponse()
				collector.DecrementInFlight()
			}()
		}

		wg.Wait()

		metrics := collector.GetMetrics()
		assert.Equal(t, uint64(iterations), metrics.TotalRequests)
		assert.Equal(t, uint64(iterations), metrics.TotalResponses)
		assert.Equal(t, uint64(iterations), metrics.GetRequests)
		assert.Equal(t, uint64(0), metrics.RequestsInFlight)
	})

	t.Run("complete request lifecycle", func(t *testing.T) {
		collector := NewCollector()

		// Request comes in
		collector.IncrementRequest("POST")
		assert.Equal(t, uint64(1), collector.GetMetrics().RequestsInFlight)

		// Request completes successfully
		collector.IncrementResponse()
		collector.DecrementInFlight()

		metrics := collector.GetMetrics()
		assert.Equal(t, uint64(1), metrics.TotalRequests)
		assert.Equal(t, uint64(1), metrics.TotalResponses)
		assert.Equal(t, uint64(0), metrics.TotalErrors)
		assert.Equal(t, uint64(0), metrics.RequestsInFlight)
	})

	t.Run("request with error", func(t *testing.T) {
		collector := NewCollector()

		// Request comes in
		collector.IncrementRequest("GET")

		// Request fails
		collector.IncrementError()
		collector.IncrementResponse()
		collector.DecrementInFlight()

		metrics := collector.GetMetrics()
		assert.Equal(t, uint64(1), metrics.TotalRequests)
		assert.Equal(t, uint64(1), metrics.TotalResponses)
		assert.Equal(t, uint64(1), metrics.TotalErrors)
		assert.Equal(t, uint64(0), metrics.RequestsInFlight)
	})
}

func TestSlowRequestConfig(t *testing.T) {
	t.Run("default slow request config", func(t *testing.T) {
		config := DefaultSlowRequestConfig

		assert.Equal(t, time.Second, config.Threshold)
		assert.Nil(t, config.LogHandler)
	})

	t.Run("custom slow request config", func(t *testing.T) {
		called := false
		config := SlowRequestConfig{
			Threshold: 500 * time.Millisecond,
			LogHandler: func(method, path, traceID string, duration time.Duration) {
				called = true
			},
		}

		assert.Equal(t, 500*time.Millisecond, config.Threshold)
		assert.NotNil(t, config.LogHandler)

		config.LogHandler("GET", "/test", "trace-123", time.Second)
		assert.True(t, called)
	})
}

func TestLogSlowRequest(t *testing.T) {
	t.Run("log slow request with default handler", func(t *testing.T) {
		config := DefaultSlowRequestConfig
		// Should not panic
		LogSlowRequest(config, "GET", "/api/test", "trace-123", 2*time.Second)
	})

	t.Run("log slow request with custom handler", func(t *testing.T) {
		var capturedMethod, capturedPath, capturedTraceID string
		var capturedDuration time.Duration

		config := SlowRequestConfig{
			Threshold: time.Second,
			LogHandler: func(method, path, traceID string, duration time.Duration) {
				capturedMethod = method
				capturedPath = path
				capturedTraceID = traceID
				capturedDuration = duration
			},
		}

		LogSlowRequest(config, "POST", "/api/slow", "trace-456", 3*time.Second)

		assert.Equal(t, "POST", capturedMethod)
		assert.Equal(t, "/api/slow", capturedPath)
		assert.Equal(t, "trace-456", capturedTraceID)
		assert.Equal(t, 3*time.Second, capturedDuration)
	})
}

func TestMetricsSnapshot(t *testing.T) {
	t.Run("get metrics returns snapshot", func(t *testing.T) {
		collector := NewCollector()

		collector.IncrementRequest("GET")
		snapshot1 := collector.GetMetrics()

		collector.IncrementRequest("POST")
		snapshot2 := collector.GetMetrics()

		// Snapshots should be independent
		assert.Equal(t, uint64(1), snapshot1.TotalRequests)
		assert.Equal(t, uint64(2), snapshot2.TotalRequests)
	})

	t.Run("modifying snapshot doesn't affect collector", func(t *testing.T) {
		collector := NewCollector()

		collector.IncrementRequest("GET")
		snapshot := collector.GetMetrics()

		// Try to modify snapshot (won't affect collector)
		snapshot.TotalRequests = 999

		// Collector should be unchanged
		assert.Equal(t, uint64(1), collector.GetMetrics().TotalRequests)
	})
}

func BenchmarkCollector(b *testing.B) {
	b.Run("IncrementRequest", func(b *testing.B) {
		collector := NewCollector()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			collector.IncrementRequest("GET")
		}
	})

	b.Run("GetMetrics", func(b *testing.B) {
		collector := NewCollector()
		collector.IncrementRequest("GET")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = collector.GetMetrics()
		}
	})

	b.Run("ConcurrentOperations", func(b *testing.B) {
		collector := NewCollector()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				collector.IncrementRequest("GET")
				collector.IncrementResponse()
				collector.DecrementInFlight()
			}
		})
	})
}

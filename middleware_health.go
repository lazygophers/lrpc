package lrpc

import (
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

// HealthStatus represents the health status of the application
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a health check function
type HealthCheck func() (status HealthStatus, message string)

// HealthChecker manages health checks
type HealthChecker struct {
	mu         sync.RWMutex
	checks     map[string]HealthCheck
	startTime  time.Time
	readyTime  time.Time
	isReady    bool
	customData map[string]interface{}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks:     make(map[string]HealthCheck),
		startTime:  time.Now(),
		customData: make(map[string]interface{}),
	}
}

// AddCheck adds a health check
func (h *HealthChecker) AddCheck(name string, check HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// RemoveCheck removes a health check
func (h *HealthChecker) RemoveCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.checks, name)
}

// SetReady marks the application as ready
func (h *HealthChecker) SetReady() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.isReady {
		h.readyTime = time.Now()
		h.isReady = true
	}
}

// SetNotReady marks the application as not ready
func (h *HealthChecker) SetNotReady() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.isReady = false
}

// SetCustomData sets custom data for health endpoint
func (h *HealthChecker) SetCustomData(key string, value interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.customData[key] = value
}

// RunChecks executes all health checks
func (h *HealthChecker) RunChecks() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make(map[string]interface{})
	overallStatus := HealthStatusHealthy

	for name, check := range h.checks {
		status, message := check()
		results[name] = map[string]interface{}{
			"status":  status,
			"message": message,
		}

		// Determine overall status
		if status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	return map[string]interface{}{
		"status":   overallStatus,
		"checks":   results,
		"uptime":   time.Since(h.startTime).Seconds(),
		"custom":   h.customData,
	}
}

// IsReady returns whether the application is ready
func (h *HealthChecker) IsReady() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isReady
}

// AddHealthEndpoints adds standard health check endpoints to the app
func (app *App) AddHealthEndpoints(prefix string, checker *HealthChecker) error {
	if prefix == "" {
		prefix = "/"
	}

	// Liveness probe - always returns 200 if server is running
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

	// Readiness probe - checks if app is ready to serve traffic
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
			if status == HealthStatusUnhealthy {
				ctx.Context().SetStatusCode(fasthttp.StatusServiceUnavailable)
			} else if status == HealthStatusDegraded {
				ctx.Context().SetStatusCode(fasthttp.StatusOK)
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

// Default health check examples

// DatabaseHealthCheck creates a health check for database connectivity
func DatabaseHealthCheck(pingFunc func() error) HealthCheck {
	return func() (HealthStatus, string) {
		err := pingFunc()
		if err != nil {
			return HealthStatusUnhealthy, "Database connection failed: " + err.Error()
		}
		return HealthStatusHealthy, "Database connected"
	}
}

// CacheHealthCheck creates a health check for cache connectivity
func CacheHealthCheck(pingFunc func() error) HealthCheck {
	return func() (HealthStatus, string) {
		err := pingFunc()
		if err != nil {
			return HealthStatusDegraded, "Cache connection failed: " + err.Error()
		}
		return HealthStatusHealthy, "Cache connected"
	}
}

// ExternalServiceHealthCheck creates a health check for external service
func ExternalServiceHealthCheck(name string, checkFunc func() error) HealthCheck {
	return func() (HealthStatus, string) {
		err := checkFunc()
		if err != nil {
			return HealthStatusDegraded, name + " unavailable: " + err.Error()
		}
		return HealthStatusHealthy, name + " available"
	}
}

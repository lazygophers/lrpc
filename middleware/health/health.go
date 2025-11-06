package health

import (
	"sync"
	"time"
)

// Status represents the health status of the application
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Check represents a health check function
type Check func() (status Status, message string)

// Checker manages health checks
type Checker struct {
	mu         sync.RWMutex
	checks     map[string]Check
	startTime  time.Time
	readyTime  time.Time
	isReady    bool
	customData map[string]interface{}
}

// NewChecker creates a new health checker
func NewChecker() *Checker {
	return &Checker{
		checks:     make(map[string]Check),
		startTime:  time.Now(),
		customData: make(map[string]interface{}),
	}
}

// AddCheck adds a health check
func (h *Checker) AddCheck(name string, check Check) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// RemoveCheck removes a health check
func (h *Checker) RemoveCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.checks, name)
}

// SetReady marks the application as ready
func (h *Checker) SetReady() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.isReady {
		h.readyTime = time.Now()
		h.isReady = true
	}
}

// SetNotReady marks the application as not ready
func (h *Checker) SetNotReady() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.isReady = false
}

// SetCustomData sets custom data for health endpoint
func (h *Checker) SetCustomData(key string, value interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.customData[key] = value
}

// RunChecks executes all health checks
func (h *Checker) RunChecks() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	results := make(map[string]interface{})
	overallStatus := StatusHealthy

	for name, check := range h.checks {
		status, message := check()
		results[name] = map[string]interface{}{
			"status":  status,
			"message": message,
		}

		// Determine overall status
		if status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	return map[string]interface{}{
		"status": overallStatus,
		"checks": results,
		"uptime": time.Since(h.startTime).Seconds(),
		"custom": h.customData,
	}
}

// IsReady returns whether the application is ready
func (h *Checker) IsReady() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isReady
}

// DatabaseCheck creates a health check for database connectivity
func DatabaseCheck(pingFunc func() error) Check {
	return func() (Status, string) {
		err := pingFunc()
		if err != nil {
			return StatusUnhealthy, "Database connection failed: " + err.Error()
		}
		return StatusHealthy, "Database connected"
	}
}

// CacheCheck creates a health check for cache connectivity
func CacheCheck(pingFunc func() error) Check {
	return func() (Status, string) {
		err := pingFunc()
		if err != nil {
			return StatusDegraded, "Cache connection failed: " + err.Error()
		}
		return StatusHealthy, "Cache connected"
	}
}

// ExternalServiceCheck creates a health check for external service
func ExternalServiceCheck(name string, checkFunc func() error) Check {
	return func() (Status, string) {
		err := checkFunc()
		if err != nil {
			return StatusDegraded, name + " unavailable: " + err.Error()
		}
		return StatusHealthy, name + " available"
	}
}

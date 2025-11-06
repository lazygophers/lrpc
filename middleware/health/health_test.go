package health

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewChecker(t *testing.T) {
	t.Run("create new checker", func(t *testing.T) {
		checker := NewChecker()

		assert.NotNil(t, checker)
		assert.NotNil(t, checker.checks)
		assert.NotNil(t, checker.customData)
		assert.False(t, checker.isReady)
		assert.False(t, checker.startTime.IsZero())
	})
}

func TestAddCheck(t *testing.T) {
	t.Run("add single check", func(t *testing.T) {
		checker := NewChecker()

		check := func() (Status, string) {
			return StatusHealthy, "OK"
		}

		checker.AddCheck("test", check)

		checker.mu.RLock()
		assert.Len(t, checker.checks, 1)
		assert.Contains(t, checker.checks, "test")
		checker.mu.RUnlock()
	})

	t.Run("add multiple checks", func(t *testing.T) {
		checker := NewChecker()

		checker.AddCheck("db", func() (Status, string) {
			return StatusHealthy, "Database OK"
		})
		checker.AddCheck("cache", func() (Status, string) {
			return StatusHealthy, "Cache OK"
		})

		checker.mu.RLock()
		assert.Len(t, checker.checks, 2)
		assert.Contains(t, checker.checks, "db")
		assert.Contains(t, checker.checks, "cache")
		checker.mu.RUnlock()
	})

	t.Run("overwrite existing check", func(t *testing.T) {
		checker := NewChecker()

		checker.AddCheck("test", func() (Status, string) {
			return StatusHealthy, "First"
		})
		checker.AddCheck("test", func() (Status, string) {
			return StatusUnhealthy, "Second"
		})

		checker.mu.RLock()
		assert.Len(t, checker.checks, 1)
		checker.mu.RUnlock()

		results := checker.RunChecks()
		checks := results["checks"].(map[string]interface{})
		testCheck := checks["test"].(map[string]interface{})
		assert.Equal(t, StatusUnhealthy, testCheck["status"])
		assert.Equal(t, "Second", testCheck["message"])
	})
}

func TestRemoveCheck(t *testing.T) {
	t.Run("remove existing check", func(t *testing.T) {
		checker := NewChecker()

		checker.AddCheck("test1", func() (Status, string) {
			return StatusHealthy, "OK"
		})
		checker.AddCheck("test2", func() (Status, string) {
			return StatusHealthy, "OK"
		})

		checker.RemoveCheck("test1")

		checker.mu.RLock()
		assert.Len(t, checker.checks, 1)
		assert.NotContains(t, checker.checks, "test1")
		assert.Contains(t, checker.checks, "test2")
		checker.mu.RUnlock()
	})

	t.Run("remove non-existent check", func(t *testing.T) {
		checker := NewChecker()

		checker.AddCheck("test", func() (Status, string) {
			return StatusHealthy, "OK"
		})

		// Should not panic
		checker.RemoveCheck("nonexistent")

		checker.mu.RLock()
		assert.Len(t, checker.checks, 1)
		checker.mu.RUnlock()
	})
}

func TestSetReady(t *testing.T) {
	t.Run("set ready from not ready", func(t *testing.T) {
		checker := NewChecker()

		assert.False(t, checker.IsReady())

		checker.SetReady()

		assert.True(t, checker.IsReady())
		assert.False(t, checker.readyTime.IsZero())
	})

	t.Run("set ready multiple times", func(t *testing.T) {
		checker := NewChecker()

		checker.SetReady()
		firstReadyTime := checker.readyTime

		time.Sleep(10 * time.Millisecond)
		checker.SetReady()

		// Ready time should not change on second call
		assert.Equal(t, firstReadyTime, checker.readyTime)
	})
}

func TestSetNotReady(t *testing.T) {
	t.Run("set not ready from ready", func(t *testing.T) {
		checker := NewChecker()

		checker.SetReady()
		assert.True(t, checker.IsReady())

		checker.SetNotReady()
		assert.False(t, checker.IsReady())
	})

	t.Run("set not ready when already not ready", func(t *testing.T) {
		checker := NewChecker()

		assert.False(t, checker.IsReady())
		checker.SetNotReady()
		assert.False(t, checker.IsReady())
	})
}

func TestSetCustomData(t *testing.T) {
	t.Run("set single custom data", func(t *testing.T) {
		checker := NewChecker()

		checker.SetCustomData("version", "1.0.0")

		checker.mu.RLock()
		assert.Equal(t, "1.0.0", checker.customData["version"])
		checker.mu.RUnlock()
	})

	t.Run("set multiple custom data", func(t *testing.T) {
		checker := NewChecker()

		checker.SetCustomData("version", "1.0.0")
		checker.SetCustomData("environment", "production")
		checker.SetCustomData("region", "us-east-1")

		checker.mu.RLock()
		assert.Len(t, checker.customData, 3)
		assert.Equal(t, "1.0.0", checker.customData["version"])
		assert.Equal(t, "production", checker.customData["environment"])
		assert.Equal(t, "us-east-1", checker.customData["region"])
		checker.mu.RUnlock()
	})

	t.Run("overwrite existing custom data", func(t *testing.T) {
		checker := NewChecker()

		checker.SetCustomData("version", "1.0.0")
		checker.SetCustomData("version", "2.0.0")

		checker.mu.RLock()
		assert.Equal(t, "2.0.0", checker.customData["version"])
		checker.mu.RUnlock()
	})
}

func TestRunChecks(t *testing.T) {
	t.Run("run checks with all healthy", func(t *testing.T) {
		checker := NewChecker()

		checker.AddCheck("db", func() (Status, string) {
			return StatusHealthy, "Database connected"
		})
		checker.AddCheck("cache", func() (Status, string) {
			return StatusHealthy, "Cache connected"
		})

		results := checker.RunChecks()

		assert.Equal(t, StatusHealthy, results["status"])
		checks := results["checks"].(map[string]interface{})
		assert.Len(t, checks, 2)

		dbCheck := checks["db"].(map[string]interface{})
		assert.Equal(t, StatusHealthy, dbCheck["status"])
		assert.Equal(t, "Database connected", dbCheck["message"])
	})

	t.Run("run checks with one degraded", func(t *testing.T) {
		checker := NewChecker()

		checker.AddCheck("db", func() (Status, string) {
			return StatusHealthy, "Database connected"
		})
		checker.AddCheck("cache", func() (Status, string) {
			return StatusDegraded, "Cache slow"
		})

		results := checker.RunChecks()

		assert.Equal(t, StatusDegraded, results["status"])
	})

	t.Run("run checks with one unhealthy", func(t *testing.T) {
		checker := NewChecker()

		checker.AddCheck("db", func() (Status, string) {
			return StatusUnhealthy, "Database down"
		})
		checker.AddCheck("cache", func() (Status, string) {
			return StatusHealthy, "Cache connected"
		})

		results := checker.RunChecks()

		assert.Equal(t, StatusUnhealthy, results["status"])
	})

	t.Run("unhealthy takes precedence over degraded", func(t *testing.T) {
		checker := NewChecker()

		checker.AddCheck("db", func() (Status, string) {
			return StatusUnhealthy, "Database down"
		})
		checker.AddCheck("cache", func() (Status, string) {
			return StatusDegraded, "Cache slow"
		})
		checker.AddCheck("api", func() (Status, string) {
			return StatusHealthy, "API OK"
		})

		results := checker.RunChecks()

		assert.Equal(t, StatusUnhealthy, results["status"])
	})

	t.Run("run checks includes uptime", func(t *testing.T) {
		checker := NewChecker()

		time.Sleep(10 * time.Millisecond)
		results := checker.RunChecks()

		uptime := results["uptime"].(float64)
		assert.Greater(t, uptime, 0.0)
	})

	t.Run("run checks includes custom data", func(t *testing.T) {
		checker := NewChecker()

		checker.SetCustomData("version", "1.0.0")
		checker.SetCustomData("environment", "test")

		results := checker.RunChecks()

		customData := results["custom"].(map[string]interface{})
		assert.Equal(t, "1.0.0", customData["version"])
		assert.Equal(t, "test", customData["environment"])
	})

	t.Run("run checks with no checks", func(t *testing.T) {
		checker := NewChecker()

		results := checker.RunChecks()

		assert.Equal(t, StatusHealthy, results["status"])
		checks := results["checks"].(map[string]interface{})
		assert.Empty(t, checks)
	})
}

func TestIsReady(t *testing.T) {
	t.Run("is ready returns false by default", func(t *testing.T) {
		checker := NewChecker()

		assert.False(t, checker.IsReady())
	})

	t.Run("is ready returns true after SetReady", func(t *testing.T) {
		checker := NewChecker()

		checker.SetReady()
		assert.True(t, checker.IsReady())
	})

	t.Run("is ready returns false after SetNotReady", func(t *testing.T) {
		checker := NewChecker()

		checker.SetReady()
		checker.SetNotReady()
		assert.False(t, checker.IsReady())
	})
}

func TestDatabaseCheck(t *testing.T) {
	t.Run("database check healthy", func(t *testing.T) {
		check := DatabaseCheck(func() error {
			return nil
		})

		status, message := check()

		assert.Equal(t, StatusHealthy, status)
		assert.Equal(t, "Database connected", message)
	})

	t.Run("database check unhealthy", func(t *testing.T) {
		check := DatabaseCheck(func() error {
			return errors.New("connection refused")
		})

		status, message := check()

		assert.Equal(t, StatusUnhealthy, status)
		assert.Contains(t, message, "Database connection failed")
		assert.Contains(t, message, "connection refused")
	})
}

func TestCacheCheck(t *testing.T) {
	t.Run("cache check healthy", func(t *testing.T) {
		check := CacheCheck(func() error {
			return nil
		})

		status, message := check()

		assert.Equal(t, StatusHealthy, status)
		assert.Equal(t, "Cache connected", message)
	})

	t.Run("cache check degraded on error", func(t *testing.T) {
		check := CacheCheck(func() error {
			return errors.New("timeout")
		})

		status, message := check()

		assert.Equal(t, StatusDegraded, status)
		assert.Contains(t, message, "Cache connection failed")
		assert.Contains(t, message, "timeout")
	})
}

func TestExternalServiceCheck(t *testing.T) {
	t.Run("external service healthy", func(t *testing.T) {
		check := ExternalServiceCheck("Payment API", func() error {
			return nil
		})

		status, message := check()

		assert.Equal(t, StatusHealthy, status)
		assert.Equal(t, "Payment API available", message)
	})

	t.Run("external service degraded on error", func(t *testing.T) {
		check := ExternalServiceCheck("Payment API", func() error {
			return errors.New("service unavailable")
		})

		status, message := check()

		assert.Equal(t, StatusDegraded, status)
		assert.Contains(t, message, "Payment API unavailable")
		assert.Contains(t, message, "service unavailable")
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("concurrent add and remove checks", func(t *testing.T) {
		checker := NewChecker()
		var wg sync.WaitGroup

		// Add checks concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				checker.AddCheck(string(rune('a'+id)), func() (Status, string) {
					return StatusHealthy, "OK"
				})
			}(i)
		}

		// Remove checks concurrently
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				checker.RemoveCheck(string(rune('a' + id)))
			}(i)
		}

		wg.Wait()

		// Should not panic and have some checks
		checker.mu.RLock()
		assert.GreaterOrEqual(t, len(checker.checks), 0)
		checker.mu.RUnlock()
	})

	t.Run("concurrent ready state changes", func(t *testing.T) {
		checker := NewChecker()
		var wg sync.WaitGroup

		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				if id%2 == 0 {
					checker.SetReady()
				} else {
					checker.SetNotReady()
				}
			}(i)
		}

		wg.Wait()

		// Should not panic
		_ = checker.IsReady()
	})

	t.Run("concurrent custom data updates", func(t *testing.T) {
		checker := NewChecker()
		var wg sync.WaitGroup

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				checker.SetCustomData("key", id)
			}(i)
		}

		wg.Wait()

		// Should not panic
		results := checker.RunChecks()
		assert.NotNil(t, results["custom"])
	})

	t.Run("concurrent RunChecks", func(t *testing.T) {
		checker := NewChecker()

		checker.AddCheck("db", func() (Status, string) {
			time.Sleep(1 * time.Millisecond)
			return StatusHealthy, "OK"
		})

		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				results := checker.RunChecks()
				assert.NotNil(t, results)
			}()
		}

		wg.Wait()
	})
}

func TestHealthCheckLifecycle(t *testing.T) {
	t.Run("complete lifecycle", func(t *testing.T) {
		checker := NewChecker()

		// Application starts - not ready
		assert.False(t, checker.IsReady())

		// Add health checks
		checker.AddCheck("db", DatabaseCheck(func() error {
			return nil
		}))
		checker.AddCheck("cache", CacheCheck(func() error {
			return nil
		}))

		// Set custom metadata
		checker.SetCustomData("version", "1.0.0")
		checker.SetCustomData("environment", "production")

		// Run initial checks
		results := checker.RunChecks()
		assert.Equal(t, StatusHealthy, results["status"])

		// Mark as ready
		checker.SetReady()
		assert.True(t, checker.IsReady())

		// Simulate cache failure
		checker.AddCheck("cache", CacheCheck(func() error {
			return errors.New("connection lost")
		}))

		results = checker.RunChecks()
		assert.Equal(t, StatusDegraded, results["status"])

		// Simulate database failure
		checker.AddCheck("db", DatabaseCheck(func() error {
			return errors.New("database down")
		}))

		results = checker.RunChecks()
		assert.Equal(t, StatusUnhealthy, results["status"])

		// Mark as not ready
		checker.SetNotReady()
		assert.False(t, checker.IsReady())

		// Remove failed checks
		checker.RemoveCheck("db")
		checker.RemoveCheck("cache")

		results = checker.RunChecks()
		assert.Equal(t, StatusHealthy, results["status"])
	})
}

func BenchmarkChecker(b *testing.B) {
	b.Run("AddCheck", func(b *testing.B) {
		checker := NewChecker()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			checker.AddCheck("test", func() (Status, string) {
				return StatusHealthy, "OK"
			})
		}
	})

	b.Run("RunChecks", func(b *testing.B) {
		checker := NewChecker()
		checker.AddCheck("db", func() (Status, string) {
			return StatusHealthy, "OK"
		})
		checker.AddCheck("cache", func() (Status, string) {
			return StatusHealthy, "OK"
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = checker.RunChecks()
		}
	})

	b.Run("ConcurrentRunChecks", func(b *testing.B) {
		checker := NewChecker()
		checker.AddCheck("db", func() (Status, string) {
			return StatusHealthy, "OK"
		})

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = checker.RunChecks()
			}
		})
	})
}

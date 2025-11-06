package pool

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

// Mock connection for testing
type mockConn struct {
	id     int
	closed bool
	mu     sync.Mutex
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("already closed")
	}
	m.closed = true
	return nil
}

func (m *mockConn) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func TestNewConnection(t *testing.T) {
	t.Run("create new connection", func(t *testing.T) {
		conn := &mockConn{id: 1}
		pooledConn := NewConnection(conn)

		assert.NotNil(t, pooledConn)
		assert.Equal(t, conn, pooledConn.Get())
		assert.Equal(t, int64(0), pooledConn.usageCount)
		assert.False(t, pooledConn.createdAt.IsZero())
		assert.False(t, pooledConn.lastUsedAt.IsZero())
	})
}

func TestConnectionMarkUsed(t *testing.T) {
	t.Run("mark connection as used", func(t *testing.T) {
		conn := NewConnection(&mockConn{id: 1})
		initialLastUsed := conn.lastUsedAt

		time.Sleep(10 * time.Millisecond)
		conn.MarkUsed()

		assert.Equal(t, int64(1), conn.usageCount)
		assert.True(t, conn.lastUsedAt.After(initialLastUsed))
	})

	t.Run("increment usage count", func(t *testing.T) {
		conn := NewConnection(&mockConn{id: 1})

		for i := 0; i < 5; i++ {
			conn.MarkUsed()
		}

		assert.Equal(t, int64(5), conn.usageCount)
	})
}

func TestConnectionIsExpired(t *testing.T) {
	t.Run("not expired when zero lifetime", func(t *testing.T) {
		conn := NewConnection(&mockConn{id: 1})
		assert.False(t, conn.IsExpired(0, 0))
	})

	t.Run("expired due to max lifetime", func(t *testing.T) {
		conn := NewConnection(&mockConn{id: 1})
		conn.createdAt = time.Now().Add(-2 * time.Hour)

		assert.True(t, conn.IsExpired(1*time.Hour, 0))
	})

	t.Run("expired due to max idle time", func(t *testing.T) {
		conn := NewConnection(&mockConn{id: 1})
		conn.lastUsedAt = time.Now().Add(-2 * time.Minute)

		assert.True(t, conn.IsExpired(0, 1*time.Minute))
	})

	t.Run("not expired within both limits", func(t *testing.T) {
		conn := NewConnection(&mockConn{id: 1})

		assert.False(t, conn.IsExpired(1*time.Hour, 1*time.Minute))
	})
}

func TestConnectionStats(t *testing.T) {
	t.Run("get connection statistics", func(t *testing.T) {
		conn := NewConnection(&mockConn{id: 1})
		conn.MarkUsed()
		conn.MarkUsed()

		time.Sleep(10 * time.Millisecond)
		stats := conn.Stats()

		assert.Equal(t, int64(2), stats.UsageCount)
		assert.False(t, stats.CreatedAt.IsZero())
		assert.False(t, stats.LastUsedAt.IsZero())
		assert.Greater(t, stats.Age, time.Duration(0))
		assert.Greater(t, stats.IdleTime, time.Duration(0))
	})
}

func TestDefaultPoolConfig(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		config := DefaultPoolConfig

		assert.Equal(t, 100, config.MaxConns)
		assert.Equal(t, 10, config.MinConns)
		assert.Equal(t, 5*time.Minute, config.MaxIdleTime)
		assert.Equal(t, 1*time.Hour, config.MaxLifetime)
		assert.Equal(t, 10*time.Second, config.WaitTimeout)
	})
}

func TestNewPool(t *testing.T) {
	t.Run("create pool with minimum connections", func(t *testing.T) {
		created := 0
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 3,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				created++
				return &mockConn{id: created}, nil
			},
			func(conn interface{}) error {
				return conn.(*mockConn).Close()
			},
		)

		require.NoError(t, err)
		assert.NotNil(t, pool)
		assert.Equal(t, 3, created)

		stats := pool.Stats()
		assert.Equal(t, int64(3), stats.TotalConns)

		pool.Close()
	})

	t.Run("apply defaults when max conns is zero", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 0,
			MinConns: 0,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)

		require.NoError(t, err)
		assert.Equal(t, DefaultPoolConfig.MaxConns, pool.config.MaxConns)

		pool.Close()
	})

	t.Run("cap min conns to max conns", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 5,
			MinConns: 10,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)

		require.NoError(t, err)
		assert.Equal(t, 5, pool.config.MinConns)

		pool.Close()
	})

	t.Run("handle create error", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 3,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return nil, errors.New("create failed")
			},
			func(conn interface{}) error {
				return nil
			},
		)

		assert.Error(t, err)
		assert.Nil(t, pool)
	})
}

func TestPoolAcquire(t *testing.T) {
	t.Run("acquire connection from pool", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 2,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		conn, err := pool.Acquire()
		require.NoError(t, err)
		assert.NotNil(t, conn)

		stats := pool.Stats()
		assert.Equal(t, int64(1), stats.ActiveConns)
	})

	t.Run("create new connection when pool empty", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 0,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		conn, err := pool.Acquire()
		require.NoError(t, err)
		assert.NotNil(t, conn)

		stats := pool.Stats()
		assert.Equal(t, int64(1), stats.TotalConns)
		assert.Equal(t, int64(1), stats.CreateCount)
	})

	t.Run("wait timeout when pool exhausted", func(t *testing.T) {
		config := PoolConfig{
			MaxConns:    1,
			MinConns:    1, // Start with one connection
			WaitTimeout: 50 * time.Millisecond,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		// Acquire the only connection
		conn1, err := pool.Acquire()
		require.NoError(t, err)

		// Try to acquire another - should timeout since pool is at max and one is in use
		start := time.Now()
		conn2, err := pool.Acquire()
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrWaitTimeout)
		assert.Nil(t, conn2)
		assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)

		stats := pool.Stats()
		assert.Equal(t, int64(1), stats.TimeoutCount)

		// Release the connection
		pool.Release(conn1)
	})

	t.Run("error when pool is closed", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 0,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)

		pool.Close()

		conn, err := pool.Acquire()
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPoolClosed)
		assert.Nil(t, conn)
	})

	t.Run("call on acquire callback", func(t *testing.T) {
		var acquiredConn interface{}
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 1,
			OnAcquire: func(conn interface{}) {
				acquiredConn = conn
			},
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{id: 123}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		conn, err := pool.Acquire()
		require.NoError(t, err)

		assert.NotNil(t, acquiredConn)
		assert.Equal(t, conn.Get(), acquiredConn)
	})
}

func TestPoolRelease(t *testing.T) {
	t.Run("release connection back to pool", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 0,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		conn, err := pool.Acquire()
		require.NoError(t, err)

		pool.Release(conn)

		stats := pool.Stats()
		assert.Equal(t, int64(0), stats.ActiveConns)
		assert.Equal(t, int64(1), stats.IdleConns)
	})

	t.Run("release nil connection", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 0,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		// Should not panic
		pool.Release(nil)
	})

	t.Run("close connection when releasing to closed pool", func(t *testing.T) {
		closed := false
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 0,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				closed = true
				return nil
			},
		)
		require.NoError(t, err)

		conn, err := pool.Acquire()
		require.NoError(t, err)

		pool.Close()
		pool.Release(conn)

		assert.True(t, closed)
	})

	t.Run("call on release callback", func(t *testing.T) {
		var releasedConn interface{}
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 0,
			OnRelease: func(conn interface{}) {
				releasedConn = conn
			},
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{id: 456}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		conn, err := pool.Acquire()
		require.NoError(t, err)

		pool.Release(conn)

		assert.NotNil(t, releasedConn)
		assert.Equal(t, conn.Get(), releasedConn)
	})
}

func TestPoolStats(t *testing.T) {
	t.Run("track pool statistics", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 2,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		// Initial stats
		stats := pool.Stats()
		assert.Equal(t, int64(2), stats.TotalConns)
		assert.Equal(t, int64(2), stats.CreateCount)

		// Acquire and check stats
		conn1, _ := pool.Acquire()
		stats = pool.Stats()
		assert.Equal(t, int64(1), stats.ActiveConns)
		assert.Equal(t, int64(1), stats.ReuseCount)

		// Release and check stats
		pool.Release(conn1)
		stats = pool.Stats()
		assert.Equal(t, int64(0), stats.ActiveConns)
	})
}

func TestPoolClose(t *testing.T) {
	t.Run("close pool and all connections", func(t *testing.T) {
		closed := 0
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 3,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				closed++
				return nil
			},
		)
		require.NoError(t, err)

		err = pool.Close()
		assert.NoError(t, err)
		assert.Equal(t, 3, closed)
	})

	t.Run("error on double close", func(t *testing.T) {
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 0,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)

		err = pool.Close()
		assert.NoError(t, err)

		err = pool.Close()
		assert.ErrorIs(t, err, ErrPoolClosed)
	})
}

func TestPoolCallbacks(t *testing.T) {
	t.Run("call onCreate callback", func(t *testing.T) {
		var createdConn interface{}
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 1,
			OnCreate: func(conn interface{}) {
				createdConn = conn
			},
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{id: 789}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		assert.NotNil(t, createdConn)
		assert.Equal(t, 789, createdConn.(*mockConn).id)
	})

	t.Run("call onClose callback", func(t *testing.T) {
		var closedConn interface{}
		config := PoolConfig{
			MaxConns: 10,
			MinConns: 1,
			OnClose: func(conn interface{}) {
				closedConn = conn
			},
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{id: 999}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)

		pool.Close()

		assert.NotNil(t, closedConn)
		assert.Equal(t, 999, closedConn.(*mockConn).id)
	})
}

func TestConcurrentPoolAccess(t *testing.T) {
	t.Run("concurrent acquire and release", func(t *testing.T) {
		config := PoolConfig{
			MaxConns:    10,
			MinConns:    0,
			WaitTimeout: 5 * time.Second,
		}

		pool, err := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		require.NoError(t, err)
		defer pool.Close()

		var wg sync.WaitGroup
		iterations := 50

		for i := 0; i < iterations; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				conn, err := pool.Acquire()
				if err != nil {
					return
				}

				time.Sleep(1 * time.Millisecond)
				pool.Release(conn)
			}()
		}

		wg.Wait()

		// Give time for stats to settle
		time.Sleep(10 * time.Millisecond)

		// Just verify no panic occurred and pool is still functional
		conn, err := pool.Acquire()
		assert.NoError(t, err)
		assert.NotNil(t, conn)
	})
}

func TestPoolError(t *testing.T) {
	t.Run("pool error messages", func(t *testing.T) {
		assert.Equal(t, "pool is closed", ErrPoolClosed.Error())
		assert.Equal(t, "wait timeout", ErrWaitTimeout.Error())
	})
}

func TestDefaultServerPoolConfig(t *testing.T) {
	t.Run("default server pool config values", func(t *testing.T) {
		config := DefaultServerPoolConfig

		assert.Equal(t, 0, config.MaxConnsPerIP)
		assert.Equal(t, 0, config.MaxRequestsPerConn)
		assert.Equal(t, 10*time.Second, config.MaxIdleWorkerDuration)
		assert.True(t, config.TCPKeepalive)
		assert.Equal(t, 30*time.Second, config.TCPKeepalivePeriod)
		assert.Equal(t, 4096, config.ReadBufferSize)
		assert.Equal(t, 4096, config.WriteBufferSize)
		assert.Equal(t, 30*time.Second, config.ReadTimeout)
		assert.Equal(t, 30*time.Second, config.WriteTimeout)
		assert.Equal(t, 60*time.Second, config.IdleTimeout)
		assert.Equal(t, 4*1024*1024, config.MaxRequestBodySize)
		assert.Equal(t, 256*1024, config.Concurrency)
		assert.False(t, config.DisableKeepalive)
		assert.False(t, config.ReduceMemoryUsage)
	})
}

func TestHighPerformanceConfig(t *testing.T) {
	t.Run("high performance config values", func(t *testing.T) {
		config := HighPerformanceConfig()

		assert.Equal(t, 8192, config.ReadBufferSize)
		assert.Equal(t, 8192, config.WriteBufferSize)
		assert.Equal(t, 10*time.Second, config.ReadTimeout)
		assert.Equal(t, 10*time.Second, config.WriteTimeout)
		assert.Equal(t, 10*1024*1024, config.MaxRequestBodySize)
		assert.Equal(t, 512*1024, config.Concurrency)
		assert.False(t, config.ReduceMemoryUsage)
	})
}

func TestLowMemoryConfig(t *testing.T) {
	t.Run("low memory config values", func(t *testing.T) {
		config := LowMemoryConfig()

		assert.Equal(t, 100, config.MaxConnsPerIP)
		assert.Equal(t, 1000, config.MaxRequestsPerConn)
		assert.Equal(t, 2048, config.ReadBufferSize)
		assert.Equal(t, 2048, config.WriteBufferSize)
		assert.Equal(t, 1*1024*1024, config.MaxRequestBodySize)
		assert.Equal(t, 64*1024, config.Concurrency)
		assert.True(t, config.ReduceMemoryUsage)
	})
}

func TestApplyServerPoolConfig(t *testing.T) {
	t.Run("apply full configuration", func(t *testing.T) {
		server := &fasthttp.Server{}
		config := ServerPoolConfig{
			MaxConnsPerIP:         100,
			MaxRequestsPerConn:    1000,
			MaxIdleWorkerDuration: 15 * time.Second,
			TCPKeepalive:          true,
			TCPKeepalivePeriod:    45 * time.Second,
			ReadBufferSize:        8192,
			WriteBufferSize:       8192,
			ReadTimeout:           20 * time.Second,
			WriteTimeout:          20 * time.Second,
			IdleTimeout:           90 * time.Second,
			MaxRequestBodySize:    5 * 1024 * 1024,
			Concurrency:           128 * 1024,
			DisableKeepalive:      true,
			ReduceMemoryUsage:     true,
		}

		ApplyServerPoolConfig(server, config)

		assert.Equal(t, 100, server.MaxConnsPerIP)
		assert.Equal(t, 1000, server.MaxRequestsPerConn)
		assert.Equal(t, 15*time.Second, server.MaxIdleWorkerDuration)
		assert.True(t, server.TCPKeepalive)
		assert.Equal(t, 45*time.Second, server.TCPKeepalivePeriod)
		assert.Equal(t, 8192, server.ReadBufferSize)
		assert.Equal(t, 8192, server.WriteBufferSize)
		assert.Equal(t, 20*time.Second, server.ReadTimeout)
		assert.Equal(t, 20*time.Second, server.WriteTimeout)
		assert.Equal(t, 90*time.Second, server.IdleTimeout)
		assert.Equal(t, 5*1024*1024, server.MaxRequestBodySize)
		assert.Equal(t, 128*1024, server.Concurrency)
		assert.True(t, server.DisableKeepalive)
		assert.True(t, server.ReduceMemoryUsage)
	})

	t.Run("skip zero values", func(t *testing.T) {
		server := &fasthttp.Server{
			MaxConnsPerIP:  50,
			ReadBufferSize: 4096,
		}

		config := ServerPoolConfig{
			MaxConnsPerIP:  0, // Should be skipped
			ReadBufferSize: 0, // Should be skipped
			Concurrency:    100,
		}

		ApplyServerPoolConfig(server, config)

		// Zero values should not override existing values
		assert.Equal(t, 50, server.MaxConnsPerIP)
		assert.Equal(t, 4096, server.ReadBufferSize)
		// Non-zero value should be applied
		assert.Equal(t, 100, server.Concurrency)
	})
}

func BenchmarkPool(b *testing.B) {
	b.Run("AcquireRelease", func(b *testing.B) {
		config := PoolConfig{
			MaxConns: 100,
			MinConns: 10,
		}

		pool, _ := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		defer pool.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			conn, _ := pool.Acquire()
			pool.Release(conn)
		}
	})

	b.Run("ConcurrentAcquireRelease", func(b *testing.B) {
		config := PoolConfig{
			MaxConns: 100,
			MinConns: 10,
		}

		pool, _ := NewPool(config,
			func() (interface{}, error) {
				return &mockConn{}, nil
			},
			func(conn interface{}) error {
				return nil
			},
		)
		defer pool.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				conn, err := pool.Acquire()
				if err == nil {
					pool.Release(conn)
				}
			}
		})
	})
}

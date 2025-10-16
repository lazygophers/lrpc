package pool

import (
	"sync"
	"time"

	"github.com/lazygophers/log"
)

// PoolConfig defines connection pool configuration
type PoolConfig struct {
	// MaxConns is the maximum number of connections in the pool
	MaxConns int

	// MinConns is the minimum number of connections to keep in the pool
	MinConns int

	// MaxIdleTime is the maximum time a connection can be idle before being closed
	MaxIdleTime time.Duration

	// MaxLifetime is the maximum lifetime of a connection
	MaxLifetime time.Duration

	// WaitTimeout is the maximum time to wait for a connection from the pool
	WaitTimeout time.Duration

	// HealthCheck is a function to check if a connection is healthy
	HealthCheck func(conn interface{}) bool

	// OnCreate is called when a new connection is created
	OnCreate func(conn interface{})

	// OnClose is called when a connection is closed
	OnClose func(conn interface{})

	// OnAcquire is called when a connection is acquired from the pool
	OnAcquire func(conn interface{})

	// OnRelease is called when a connection is released to the pool
	OnRelease func(conn interface{})
}

// DefaultPoolConfig returns default pool configuration
var DefaultPoolConfig = PoolConfig{
	MaxConns:    100,
	MinConns:    10,
	MaxIdleTime: 5 * time.Minute,
	MaxLifetime: 1 * time.Hour,
	WaitTimeout: 10 * time.Second,
}

// Connection represents a pooled connection
type Connection struct {
	conn       interface{}
	createdAt  time.Time
	lastUsedAt time.Time
	usageCount int64
	mu         sync.RWMutex
}

// NewConnection creates a new pooled connection
func NewConnection(conn interface{}) *Connection {
	now := time.Now()
	return &Connection{
		conn:       conn,
		createdAt:  now,
		lastUsedAt: now,
		usageCount: 0,
	}
}

// Get returns the underlying connection
func (c *Connection) Get() interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// MarkUsed marks the connection as recently used
func (c *Connection) MarkUsed() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastUsedAt = time.Now()
	c.usageCount++
}

// IsExpired checks if the connection has expired
func (c *Connection) IsExpired(maxLifetime, maxIdleTime time.Duration) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()

	// Check max lifetime
	if maxLifetime > 0 && now.Sub(c.createdAt) > maxLifetime {
		return true
	}

	// Check max idle time
	if maxIdleTime > 0 && now.Sub(c.lastUsedAt) > maxIdleTime {
		return true
	}

	return false
}

// Stats returns connection statistics
func (c *Connection) Stats() ConnectionStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return ConnectionStats{
		CreatedAt:  c.createdAt,
		LastUsedAt: c.lastUsedAt,
		UsageCount: c.usageCount,
		Age:        time.Since(c.createdAt),
		IdleTime:   time.Since(c.lastUsedAt),
	}
}

// ConnectionStats represents connection statistics
type ConnectionStats struct {
	CreatedAt  time.Time
	LastUsedAt time.Time
	UsageCount int64
	Age        time.Duration
	IdleTime   time.Duration
}

// Pool represents a connection pool
type Pool struct {
	config    PoolConfig
	conns     chan *Connection
	waiting   chan chan *Connection
	mu        sync.RWMutex
	closed    bool
	stats     PoolStats
	statsMu   sync.RWMutex
	createNew func() (interface{}, error)
	closeConn func(interface{}) error
}

// PoolStats represents pool statistics
type PoolStats struct {
	TotalConns     int64
	IdleConns      int64
	ActiveConns    int64
	WaitCount      int64
	CreateCount    int64
	CloseCount     int64
	ReuseCount     int64
	TimeoutCount   int64
	HealthFailures int64
}

// NewPool creates a new connection pool
func NewPool(config PoolConfig, createNew func() (interface{}, error), closeConn func(interface{}) error) (*Pool, error) {
	if config.MaxConns <= 0 {
		config.MaxConns = DefaultPoolConfig.MaxConns
	}
	if config.MinConns < 0 {
		config.MinConns = 0
	}
	if config.MinConns > config.MaxConns {
		config.MinConns = config.MaxConns
	}

	p := &Pool{
		config:    config,
		conns:     make(chan *Connection, config.MaxConns),
		waiting:   make(chan chan *Connection, config.MaxConns),
		createNew: createNew,
		closeConn: closeConn,
	}

	// Pre-create minimum connections
	for i := 0; i < config.MinConns; i++ {
		conn, err := p.createConnection()
		if err != nil {
			log.Errorf("err:%v", err)
			// Close already created connections
			p.Close()
			return nil, err
		}
		p.conns <- conn
	}

	// Start maintenance goroutine
	go p.maintain()

	return p, nil
}

// createConnection creates a new connection
func (p *Pool) createConnection() (*Connection, error) {
	conn, err := p.createNew()
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	pooledConn := NewConnection(conn)

	if p.config.OnCreate != nil {
		p.config.OnCreate(conn)
	}

	p.statsMu.Lock()
	p.stats.CreateCount++
	p.stats.TotalConns++
	p.statsMu.Unlock()

	return pooledConn, nil
}

// Acquire gets a connection from the pool
func (p *Pool) Acquire() (*Connection, error) {
	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		err := ErrPoolClosed
		log.Errorf("err:%v", err)
		return nil, err
	}
	p.mu.RUnlock()

	// Try to get a connection from the pool
	select {
	case conn := <-p.conns:
		// Check if connection is still valid
		if conn.IsExpired(p.config.MaxLifetime, p.config.MaxIdleTime) {
			p.closeConnection(conn)
			return p.Acquire() // Retry
		}

		if p.config.HealthCheck != nil && !p.config.HealthCheck(conn.Get()) {
			p.statsMu.Lock()
			p.stats.HealthFailures++
			p.statsMu.Unlock()
			p.closeConnection(conn)
			return p.Acquire() // Retry
		}

		conn.MarkUsed()

		if p.config.OnAcquire != nil {
			p.config.OnAcquire(conn.Get())
		}

		p.statsMu.Lock()
		p.stats.ReuseCount++
		p.stats.ActiveConns++
		p.stats.IdleConns--
		p.statsMu.Unlock()

		return conn, nil

	default:
		// No idle connection available, try to create a new one
		p.statsMu.RLock()
		totalConns := p.stats.TotalConns
		p.statsMu.RUnlock()

		if totalConns < int64(p.config.MaxConns) {
			conn, err := p.createConnection()
			if err != nil {
				log.Errorf("err:%v", err)
				// Fall through to wait for a connection
			} else {
				conn.MarkUsed()

				if p.config.OnAcquire != nil {
					p.config.OnAcquire(conn.Get())
				}

				p.statsMu.Lock()
				p.stats.ActiveConns++
				p.statsMu.Unlock()

				return conn, nil
			}
		}

		// Wait for a connection to be released
		waiter := make(chan *Connection, 1)

		select {
		case p.waiting <- waiter:
			p.statsMu.Lock()
			p.stats.WaitCount++
			p.statsMu.Unlock()

			select {
			case conn := <-waiter:
				return conn, nil
			case <-time.After(p.config.WaitTimeout):
				p.statsMu.Lock()
				p.stats.TimeoutCount++
				p.statsMu.Unlock()

				err := ErrWaitTimeout
				log.Errorf("err:%v", err)
				return nil, err
			}

		case <-time.After(p.config.WaitTimeout):
			p.statsMu.Lock()
			p.stats.TimeoutCount++
			p.statsMu.Unlock()

			err := ErrWaitTimeout
			log.Errorf("err:%v", err)
			return nil, err
		}
	}
}

// Release returns a connection to the pool
func (p *Pool) Release(conn *Connection) {
	if conn == nil {
		return
	}

	p.mu.RLock()
	if p.closed {
		p.mu.RUnlock()
		p.closeConnection(conn)
		return
	}
	p.mu.RUnlock()

	if p.config.OnRelease != nil {
		p.config.OnRelease(conn.Get())
	}

	p.statsMu.Lock()
	p.stats.ActiveConns--
	p.stats.IdleConns++
	p.statsMu.Unlock()

	// Try to give connection to waiting goroutine
	select {
	case waiter := <-p.waiting:
		waiter <- conn
		return
	default:
		// No one waiting, return to pool
		select {
		case p.conns <- conn:
			// Successfully returned to pool
		default:
			// Pool is full, close the connection
			p.closeConnection(conn)
		}
	}
}

// closeConnection closes a connection
func (p *Pool) closeConnection(conn *Connection) {
	if conn == nil {
		return
	}

	if p.config.OnClose != nil {
		p.config.OnClose(conn.Get())
	}

	if p.closeConn != nil {
		err := p.closeConn(conn.Get())
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}

	p.statsMu.Lock()
	p.stats.CloseCount++
	p.stats.TotalConns--
	p.statsMu.Unlock()
}

// maintain performs periodic maintenance on the pool
func (p *Pool) maintain() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		p.mu.RLock()
		if p.closed {
			p.mu.RUnlock()
			return
		}
		p.mu.RUnlock()

		// Clean up expired connections
		p.cleanupExpired()
	}
}

// cleanupExpired removes expired connections from the pool
func (p *Pool) cleanupExpired() {
	var toClose []*Connection

	// Collect expired connections
	for {
		select {
		case conn := <-p.conns:
			if conn.IsExpired(p.config.MaxLifetime, p.config.MaxIdleTime) {
				toClose = append(toClose, conn)
			} else {
				// Return non-expired connection to pool
				p.conns <- conn
				return // Stop checking
			}
		default:
			// No more connections to check
			goto cleanup
		}
	}

cleanup:
	// Close expired connections
	for _, conn := range toClose {
		p.closeConnection(conn)
	}
}

// Stats returns pool statistics
func (p *Pool) Stats() PoolStats {
	p.statsMu.RLock()
	defer p.statsMu.RUnlock()
	return p.stats
}

// Close closes the pool and all connections
func (p *Pool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return ErrPoolClosed
	}
	p.closed = true
	p.mu.Unlock()

	// Close all connections in pool
	close(p.conns)
	for conn := range p.conns {
		p.closeConnection(conn)
	}

	return nil
}

// Errors
var (
	ErrPoolClosed  = &PoolError{"pool is closed"}
	ErrWaitTimeout = &PoolError{"wait timeout"}
)

// PoolError represents a pool error
type PoolError struct {
	msg string
}

func (e *PoolError) Error() string {
	return e.msg
}

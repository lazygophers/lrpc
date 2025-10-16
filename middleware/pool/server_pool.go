package pool

import (
	"time"

	"github.com/valyala/fasthttp"
)

// ServerPoolConfig defines fasthttp server pool configuration
type ServerPoolConfig struct {
	// MaxConnsPerIP limits the number of concurrent connections per IP
	MaxConnsPerIP int

	// MaxRequestsPerConn limits the number of requests per connection
	MaxRequestsPerConn int

	// MaxIdleWorkerDuration is the maximum idle time for workers
	MaxIdleWorkerDuration time.Duration

	// TCPKeepalive enables TCP keepalive
	TCPKeepalive bool

	// TCPKeepalivePeriod sets the TCP keepalive period
	TCPKeepalivePeriod time.Duration

	// ReadBufferSize sets the read buffer size per connection
	ReadBufferSize int

	// WriteBufferSize sets the write buffer size per connection
	WriteBufferSize int

	// ReadTimeout sets the maximum duration for reading the full request
	ReadTimeout time.Duration

	// WriteTimeout sets the maximum duration for writing the full response
	WriteTimeout time.Duration

	// IdleTimeout sets the maximum idle time before closing connection
	IdleTimeout time.Duration

	// MaxRequestBodySize sets the maximum request body size
	MaxRequestBodySize int

	// Concurrency sets the maximum number of concurrent connections
	Concurrency int

	// DisableKeepalive disables HTTP keep-alive
	DisableKeepalive bool

	// ReduceMemoryUsage enables memory usage reduction mode
	ReduceMemoryUsage bool
}

// DefaultServerPoolConfig returns default server pool configuration
var DefaultServerPoolConfig = ServerPoolConfig{
	MaxConnsPerIP:         0,    // Unlimited
	MaxRequestsPerConn:    0,    // Unlimited
	MaxIdleWorkerDuration: 10 * time.Second,
	TCPKeepalive:          true,
	TCPKeepalivePeriod:    30 * time.Second,
	ReadBufferSize:        4096,
	WriteBufferSize:       4096,
	ReadTimeout:           30 * time.Second,
	WriteTimeout:          30 * time.Second,
	IdleTimeout:           60 * time.Second,
	MaxRequestBodySize:    4 * 1024 * 1024, // 4MB
	Concurrency:           256 * 1024,       // 256K
	DisableKeepalive:      false,
	ReduceMemoryUsage:     false,
}

// HighPerformanceConfig returns a configuration optimized for high performance
func HighPerformanceConfig() ServerPoolConfig {
	return ServerPoolConfig{
		MaxConnsPerIP:         0, // Unlimited
		MaxRequestsPerConn:    0, // Unlimited
		MaxIdleWorkerDuration: 5 * time.Second,
		TCPKeepalive:          true,
		TCPKeepalivePeriod:    30 * time.Second,
		ReadBufferSize:        8192,
		WriteBufferSize:       8192,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           30 * time.Second,
		MaxRequestBodySize:    10 * 1024 * 1024, // 10MB
		Concurrency:           512 * 1024,        // 512K
		DisableKeepalive:      false,
		ReduceMemoryUsage:     false,
	}
}

// LowMemoryConfig returns a configuration optimized for low memory usage
func LowMemoryConfig() ServerPoolConfig {
	return ServerPoolConfig{
		MaxConnsPerIP:         100,
		MaxRequestsPerConn:    1000,
		MaxIdleWorkerDuration: 30 * time.Second,
		TCPKeepalive:          true,
		TCPKeepalivePeriod:    60 * time.Second,
		ReadBufferSize:        2048,
		WriteBufferSize:       2048,
		ReadTimeout:           60 * time.Second,
		WriteTimeout:          60 * time.Second,
		IdleTimeout:           120 * time.Second,
		MaxRequestBodySize:    1 * 1024 * 1024, // 1MB
		Concurrency:           64 * 1024,        // 64K
		DisableKeepalive:      false,
		ReduceMemoryUsage:     true,
	}
}

// ApplyServerPoolConfig applies pool configuration to fasthttp server
func ApplyServerPoolConfig(server *fasthttp.Server, config ServerPoolConfig) {
	if config.MaxConnsPerIP > 0 {
		server.MaxConnsPerIP = config.MaxConnsPerIP
	}

	if config.MaxRequestsPerConn > 0 {
		server.MaxRequestsPerConn = config.MaxRequestsPerConn
	}

	if config.MaxIdleWorkerDuration > 0 {
		server.MaxIdleWorkerDuration = config.MaxIdleWorkerDuration
	}

	if config.TCPKeepalive {
		server.TCPKeepalive = true
		if config.TCPKeepalivePeriod > 0 {
			server.TCPKeepalivePeriod = config.TCPKeepalivePeriod
		}
	}

	if config.ReadBufferSize > 0 {
		server.ReadBufferSize = config.ReadBufferSize
	}

	if config.WriteBufferSize > 0 {
		server.WriteBufferSize = config.WriteBufferSize
	}

	if config.ReadTimeout > 0 {
		server.ReadTimeout = config.ReadTimeout
	}

	if config.WriteTimeout > 0 {
		server.WriteTimeout = config.WriteTimeout
	}

	if config.IdleTimeout > 0 {
		server.IdleTimeout = config.IdleTimeout
	}

	if config.MaxRequestBodySize > 0 {
		server.MaxRequestBodySize = config.MaxRequestBodySize
	}

	if config.Concurrency > 0 {
		server.Concurrency = config.Concurrency
	}

	server.DisableKeepalive = config.DisableKeepalive
	server.ReduceMemoryUsage = config.ReduceMemoryUsage
}

// ConnectionPoolStats represents connection pool statistics
type ConnectionPoolStats struct {
	ActiveConns    int
	IdleConns      int
	TotalRequests  uint64
	TotalResponses uint64
}

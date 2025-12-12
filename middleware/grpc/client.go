package grpc

import (
	"context"
	"time"

	"github.com/lazygophers/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientConfig defines gRPC client configuration
type ClientConfig struct {
	// Address is the server address (host:port)
	Address string

	// Timeout sets the connection timeout
	Timeout time.Duration

	// KeepAlive enables keep-alive
	KeepAlive bool

	// KeepAliveTime is the keep-alive time
	KeepAliveTime time.Duration

	// KeepAliveTimeout is the keep-alive timeout
	KeepAliveTimeout time.Duration

	// MaxRecvMsgSize sets the maximum message size the client can receive
	MaxRecvMsgSize int

	// MaxSendMsgSize sets the maximum message size the client can send
	MaxSendMsgSize int

	// Insecure disables transport security
	Insecure bool

	// WithBlock blocks until connection is established
	WithBlock bool

	// Interceptors for the client
	UnaryInterceptors  []grpc.UnaryClientInterceptor
	StreamInterceptors []grpc.StreamClientInterceptor
}

// DefaultClientConfig returns default client configuration
var DefaultClientConfig = ClientConfig{
	Timeout:          10 * time.Second,
	KeepAlive:        true,
	KeepAliveTime:    30 * time.Second,
	KeepAliveTimeout: 10 * time.Second,
	MaxRecvMsgSize:   4 * 1024 * 1024, // 4MB
	MaxSendMsgSize:   4 * 1024 * 1024, // 4MB
	Insecure:         true,
	WithBlock:        false,
}

// NewClient creates a new gRPC client connection
func NewClient(config ClientConfig) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	// Add insecure credentials if enabled
	if config.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add keep-alive parameters
	if config.KeepAlive {
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                config.KeepAliveTime,
			Timeout:             config.KeepAliveTimeout,
			PermitWithoutStream: true,
		}))
	}

	// Set message size limits
	if config.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(config.MaxRecvMsgSize)))
	}
	if config.MaxSendMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(config.MaxSendMsgSize)))
	}

	// Add block option
	if config.WithBlock {
		opts = append(opts, grpc.WithBlock())
	}

	// Add interceptors
	if len(config.UnaryInterceptors) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(config.UnaryInterceptors...))
	}
	if len(config.StreamInterceptors) > 0 {
		opts = append(opts, grpc.WithChainStreamInterceptor(config.StreamInterceptors...))
	}

	// Create context with timeout
	ctx := context.Background()
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Dial
	conn, err := grpc.DialContext(ctx, config.Address, opts...)
	if err != nil {
		log.Errorf("err:%v", err)
		return nil, err
	}

	return conn, nil
}

// ClientPool manages a pool of gRPC client connections
type ClientPool struct {
	config ClientConfig
	pool   chan *grpc.ClientConn
	create func() (*grpc.ClientConn, error)
}

// NewClientPool creates a new client pool
func NewClientPool(config ClientConfig, size int) (*ClientPool, error) {
	p := &ClientPool{
		config: config,
		pool:   make(chan *grpc.ClientConn, size),
		create: func() (*grpc.ClientConn, error) {
			return NewClient(config)
		},
	}

	// Pre-create connections
	for i := 0; i < size; i++ {
		conn, err := p.create()
		if err != nil {
			log.Errorf("err:%v", err)
			// Close already created connections
			p.Close()
			return nil, err
		}
		p.pool <- conn
	}

	return p, nil
}

// Get gets a connection from the pool
func (p *ClientPool) Get() (*grpc.ClientConn, error) {
	select {
	case conn := <-p.pool:
		return conn, nil
	default:
		// Pool is empty, create new connection
		return p.create()
	}
}

// Put returns a connection to the pool
func (p *ClientPool) Put(conn *grpc.ClientConn) {
	select {
	case p.pool <- conn:
		// Successfully returned to pool
	default:
		// Pool is full, close the connection
		err := conn.Close()
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}
}

// Close closes all connections in the pool
func (p *ClientPool) Close() error {
	close(p.pool)
	for conn := range p.pool {
		err := conn.Close()
		if err != nil {
			log.Errorf("err:%v", err)
		}
	}
	return nil
}

// LoggingUnaryClientInterceptor logs gRPC unary calls
func LoggingUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		if err != nil {
			log.Errorf("gRPC call failed: method=%s duration=%v err=%v", method, duration, err)
		} else {
			log.Infof("gRPC call: method=%s duration=%v", method, duration)
		}

		return err
	}
}

// RetryUnaryClientInterceptor adds retry logic to gRPC unary calls
func RetryUnaryClientInterceptor(maxRetries int, backoff time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		var err error
		for i := 0; i <= maxRetries; i++ {
			err = invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}

			if i < maxRetries {
				log.Warnf("gRPC call failed, retrying: method=%s attempt=%d/%d err=%v", method, i+1, maxRetries, err)
				time.Sleep(backoff * time.Duration(i+1))
			}
		}

		log.Errorf("gRPC call failed after retries: method=%s err=%v", method, err)
		return err
	}
}

// TimeoutUnaryClientInterceptor adds timeout to gRPC unary calls
func TimeoutUnaryClientInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

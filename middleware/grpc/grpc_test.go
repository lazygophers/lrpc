package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestDefaultHTTPtoGRPCBridge(t *testing.T) {
	t.Run("default bridge has header mappings", func(t *testing.T) {
		bridge := DefaultHTTPtoGRPCBridge

		assert.NotNil(t, bridge)
		assert.NotEmpty(t, bridge.HeaderMapping)
		assert.Contains(t, bridge.HeaderMapping, "Content-Type")
		assert.Contains(t, bridge.HeaderMapping, "Authorization")
		assert.Contains(t, bridge.HeaderMapping, "User-Agent")
	})
}

func TestHTTPRequestToGRPCContext(t *testing.T) {
	t.Run("convert HTTP headers to gRPC metadata", func(t *testing.T) {
		bridge := &HTTPtoGRPCBridge{
			HeaderMapping: map[string]string{
				"Authorization": "authorization",
				"X-Request-Id":  "x-request-id",
			},
		}

		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Bearer token123")
		ctx.Request.Header.Set("X-Request-Id", "req-123")

		grpcCtx := bridge.HTTPRequestToGRPCContext(ctx)

		md, ok := metadata.FromOutgoingContext(grpcCtx)
		assert.True(t, ok)
		assert.Equal(t, []string{"Bearer token123"}, md.Get("authorization"))
		assert.Equal(t, []string{"req-123"}, md.Get("x-request-id"))
	})

	t.Run("skip empty headers", func(t *testing.T) {
		bridge := &HTTPtoGRPCBridge{
			HeaderMapping: map[string]string{
				"Authorization": "authorization",
				"X-Request-Id":  "x-request-id",
			},
		}

		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Bearer token123")
		// X-Request-Id not set

		grpcCtx := bridge.HTTPRequestToGRPCContext(ctx)

		md, ok := metadata.FromOutgoingContext(grpcCtx)
		assert.True(t, ok)
		assert.Equal(t, []string{"Bearer token123"}, md.Get("authorization"))
		assert.Empty(t, md.Get("x-request-id"))
	})

	t.Run("use custom metadata extractor", func(t *testing.T) {
		bridge := &HTTPtoGRPCBridge{
			MetadataExtractor: func(ctx *fasthttp.RequestCtx) metadata.MD {
				md := metadata.New(nil)
				md.Set("custom-key", "custom-value")
				return md
			},
		}

		ctx := &fasthttp.RequestCtx{}
		grpcCtx := bridge.HTTPRequestToGRPCContext(ctx)

		md, ok := metadata.FromOutgoingContext(grpcCtx)
		assert.True(t, ok)
		assert.Equal(t, []string{"custom-value"}, md.Get("custom-key"))
	})
}

func TestGRPCResponseToHTTPResponse(t *testing.T) {
	t.Run("convert protobuf response to HTTP", func(t *testing.T) {
		bridge := &HTTPtoGRPCBridge{}
		ctx := &fasthttp.RequestCtx{}

		resp := &emptypb.Empty{}
		header := metadata.New(map[string]string{
			"x-response-id": "resp-123",
		})
		trailer := metadata.New(map[string]string{
			"x-trailer": "trailer-value",
		})

		err := bridge.GRPCResponseToHTTPResponse(ctx, resp, header, trailer)

		assert.NoError(t, err)
		assert.Equal(t, "application/protobuf", string(ctx.Response.Header.Peek("Content-Type")))
		assert.Equal(t, "resp-123", string(ctx.Response.Header.Peek("x-response-id")))
		assert.Equal(t, "trailer-value", string(ctx.Response.Header.Peek("Trailer-x-trailer")))
		// Empty protobuf messages can have empty body, which is valid
		assert.NotNil(t, ctx.Response.Body())
	})

	t.Run("handle multiple header values", func(t *testing.T) {
		bridge := &HTTPtoGRPCBridge{}
		ctx := &fasthttp.RequestCtx{}

		resp := &emptypb.Empty{}
		header := metadata.MD{
			"x-multi": []string{"value1", "value2", "value3"},
		}

		err := bridge.GRPCResponseToHTTPResponse(ctx, resp, header, nil)

		assert.NoError(t, err)
		// fasthttp concatenates multiple values
		headerValue := string(ctx.Response.Header.Peek("x-multi"))
		assert.NotEmpty(t, headerValue)
	})
}

func TestNewStreamHandler(t *testing.T) {
	t.Run("create stream handler with custom bridge", func(t *testing.T) {
		bridge := &HTTPtoGRPCBridge{}
		handler := NewStreamHandler(bridge)

		assert.NotNil(t, handler)
		assert.Equal(t, bridge, handler.bridge)
	})

	t.Run("use default bridge when nil", func(t *testing.T) {
		handler := NewStreamHandler(nil)

		assert.NotNil(t, handler)
		assert.Equal(t, DefaultHTTPtoGRPCBridge, handler.bridge)
	})
}

func TestNewGRPCServiceAdapter(t *testing.T) {
	t.Run("create adapter with custom bridge", func(t *testing.T) {
		bridge := &HTTPtoGRPCBridge{}
		adapter := NewGRPCServiceAdapter(bridge)

		assert.NotNil(t, adapter)
		assert.Equal(t, bridge, adapter.bridge)
	})

	t.Run("use default bridge when nil", func(t *testing.T) {
		adapter := NewGRPCServiceAdapter(nil)

		assert.NotNil(t, adapter)
		assert.Equal(t, DefaultHTTPtoGRPCBridge, adapter.bridge)
	})
}

func TestUnaryHandler(t *testing.T) {
	t.Run("adapt gRPC handler to HTTP", func(t *testing.T) {
		adapter := NewGRPCServiceAdapter(nil)

		// Create a simple gRPC handler
		grpcHandler := func(ctx context.Context, req proto.Message) (proto.Message, error) {
			return &emptypb.Empty{}, nil
		}

		httpHandler := adapter.UnaryHandler(grpcHandler, &emptypb.Empty{})

		// Create HTTP request with valid protobuf body
		ctx := &fasthttp.RequestCtx{}
		reqMsg := &emptypb.Empty{}
		data, _ := proto.Marshal(reqMsg)
		ctx.Request.SetBody(data)

		err := httpHandler(ctx)

		assert.NoError(t, err)
		assert.Equal(t, "application/protobuf", string(ctx.Response.Header.Peek("Content-Type")))
	})

	t.Run("handle unmarshal error", func(t *testing.T) {
		adapter := NewGRPCServiceAdapter(nil)

		grpcHandler := func(ctx context.Context, req proto.Message) (proto.Message, error) {
			return &emptypb.Empty{}, nil
		}

		httpHandler := adapter.UnaryHandler(grpcHandler, &emptypb.Empty{})

		ctx := &fasthttp.RequestCtx{}
		ctx.Request.SetBody([]byte("invalid protobuf data"))

		err := httpHandler(ctx)

		assert.Error(t, err)
	})

	t.Run("handle gRPC handler error", func(t *testing.T) {
		adapter := NewGRPCServiceAdapter(nil)

		grpcHandler := func(ctx context.Context, req proto.Message) (proto.Message, error) {
			return nil, errors.New("handler error")
		}

		httpHandler := adapter.UnaryHandler(grpcHandler, &emptypb.Empty{})

		ctx := &fasthttp.RequestCtx{}
		reqMsg := &emptypb.Empty{}
		data, _ := proto.Marshal(reqMsg)
		ctx.Request.SetBody(data)

		err := httpHandler(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handler error")
	})
}

func TestMetadataCarrier(t *testing.T) {
	t.Run("get header value", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("X-Test-Header", "test-value")

		carrier := NewMetadataCarrier(ctx)
		values := carrier.Get("X-Test-Header")

		assert.Equal(t, []string{"test-value"}, values)
	})

	t.Run("get non-existent header", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		carrier := NewMetadataCarrier(ctx)

		values := carrier.Get("Non-Existent")

		assert.Nil(t, values)
	})

	t.Run("set response header", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		carrier := NewMetadataCarrier(ctx)

		carrier.Set("X-Response-Header", "response-value")

		assert.Equal(t, "response-value", string(ctx.Response.Header.Peek("X-Response-Header")))
	})

	t.Run("append response header", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		carrier := NewMetadataCarrier(ctx)

		carrier.Set("X-Multi", "value1")
		carrier.Append("X-Multi", "value2")

		// fasthttp concatenates values
		headerValue := string(ctx.Response.Header.Peek("X-Multi"))
		assert.NotEmpty(t, headerValue)
	})
}

func TestDefaultStatusMapper(t *testing.T) {
	tests := []struct {
		name           string
		grpcCode       int
		expectedHTTP   int
	}{
		{"OK", 0, 200},
		{"CANCELLED", 1, 499},
		{"UNKNOWN", 2, 500},
		{"INVALID_ARGUMENT", 3, 400},
		{"DEADLINE_EXCEEDED", 4, 504},
		{"NOT_FOUND", 5, 404},
		{"ALREADY_EXISTS", 6, 409},
		{"PERMISSION_DENIED", 7, 403},
		{"RESOURCE_EXHAUSTED", 8, 429},
		{"FAILED_PRECONDITION", 9, 400},
		{"ABORTED", 10, 409},
		{"OUT_OF_RANGE", 11, 400},
		{"UNIMPLEMENTED", 12, 501},
		{"INTERNAL", 13, 500},
		{"UNAVAILABLE", 14, 503},
		{"DATA_LOSS", 15, 500},
		{"UNAUTHENTICATED", 16, 401},
		{"UNKNOWN_CODE", 999, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpCode := defaultStatusMapper(tt.grpcCode)
			assert.Equal(t, tt.expectedHTTP, httpCode)
		})
	}
}

func TestGRPCErrorHandler(t *testing.T) {
	t.Run("default error handler", func(t *testing.T) {
		handler := DefaultGRPCErrorHandler
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.StatusMapper)
	})

	t.Run("handle error", func(t *testing.T) {
		handler := &GRPCErrorHandler{}
		ctx := &fasthttp.RequestCtx{}

		handler.HandleGRPCError(ctx, errors.New("test error"))

		assert.Equal(t, 500, ctx.Response.StatusCode())
		assert.Contains(t, string(ctx.Response.Body()), "Internal Server Error")
	})
}

func TestDefaultClientConfig(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		config := DefaultClientConfig

		assert.Equal(t, 10*time.Second, config.Timeout)
		assert.True(t, config.KeepAlive)
		assert.Equal(t, 30*time.Second, config.KeepAliveTime)
		assert.Equal(t, 10*time.Second, config.KeepAliveTimeout)
		assert.Equal(t, 4*1024*1024, config.MaxRecvMsgSize)
		assert.Equal(t, 4*1024*1024, config.MaxSendMsgSize)
		assert.True(t, config.Insecure)
		assert.False(t, config.WithBlock)
	})
}

func TestLoggingUnaryClientInterceptor(t *testing.T) {
	t.Run("create logging interceptor", func(t *testing.T) {
		interceptor := LoggingUnaryClientInterceptor()
		assert.NotNil(t, interceptor)
	})

	t.Run("log successful call", func(t *testing.T) {
		interceptor := LoggingUnaryClientInterceptor()

		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		}

		err := interceptor(context.Background(), "/test.Service/Method", nil, nil, nil, invoker)
		assert.NoError(t, err)
	})

	t.Run("log failed call", func(t *testing.T) {
		interceptor := LoggingUnaryClientInterceptor()

		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return errors.New("call failed")
		}

		err := interceptor(context.Background(), "/test.Service/Method", nil, nil, nil, invoker)
		assert.Error(t, err)
	})
}

func TestRetryUnaryClientInterceptor(t *testing.T) {
	t.Run("succeed on first attempt", func(t *testing.T) {
		interceptor := RetryUnaryClientInterceptor(3, 10*time.Millisecond)

		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		}

		err := interceptor(context.Background(), "/test.Service/Method", nil, nil, nil, invoker)
		assert.NoError(t, err)
	})

	t.Run("retry and succeed", func(t *testing.T) {
		attempts := 0
		interceptor := RetryUnaryClientInterceptor(3, 1*time.Millisecond)

		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			attempts++
			if attempts < 2 {
				return errors.New("temporary error")
			}
			return nil
		}

		err := interceptor(context.Background(), "/test.Service/Method", nil, nil, nil, invoker)
		assert.NoError(t, err)
		assert.Equal(t, 2, attempts)
	})

	t.Run("exhaust retries", func(t *testing.T) {
		interceptor := RetryUnaryClientInterceptor(2, 1*time.Millisecond)

		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return errors.New("persistent error")
		}

		err := interceptor(context.Background(), "/test.Service/Method", nil, nil, nil, invoker)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "persistent error")
	})
}

func TestTimeoutUnaryClientInterceptor(t *testing.T) {
	t.Run("complete within timeout", func(t *testing.T) {
		interceptor := TimeoutUnaryClientInterceptor(100 * time.Millisecond)

		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		}

		err := interceptor(context.Background(), "/test.Service/Method", nil, nil, nil, invoker)
		assert.NoError(t, err)
	})

	t.Run("timeout exceeded", func(t *testing.T) {
		interceptor := TimeoutUnaryClientInterceptor(10 * time.Millisecond)

		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		}

		err := interceptor(context.Background(), "/test.Service/Method", nil, nil, nil, invoker)
		assert.Error(t, err)
	})
}

func BenchmarkHTTPtoGRPCConversion(b *testing.B) {
	b.Run("HTTPRequestToGRPCContext", func(b *testing.B) {
		bridge := DefaultHTTPtoGRPCBridge
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Authorization", "Bearer token")
		ctx.Request.Header.Set("X-Request-Id", "req-123")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = bridge.HTTPRequestToGRPCContext(ctx)
		}
	})

	b.Run("GRPCResponseToHTTPResponse", func(b *testing.B) {
		bridge := &HTTPtoGRPCBridge{}
		ctx := &fasthttp.RequestCtx{}
		resp := &emptypb.Empty{}
		header := metadata.New(map[string]string{"x-test": "value"})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = bridge.GRPCResponseToHTTPResponse(ctx, resp, header, nil)
		}
	})
}

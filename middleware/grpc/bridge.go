package grpc

import (
	"context"
	"io"

	"github.com/lazygophers/log"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// HTTPtoGRPCBridge provides bridging between HTTP and gRPC
type HTTPtoGRPCBridge struct {
	// HeaderMapping maps HTTP headers to gRPC metadata keys
	HeaderMapping map[string]string

	// ErrorHandler handles errors during conversion
	ErrorHandler func(err error) error

	// MetadataExtractor extracts metadata from HTTP request
	MetadataExtractor func(ctx *fasthttp.RequestCtx) metadata.MD
}

// DefaultHTTPtoGRPCBridge returns default bridge configuration
var DefaultHTTPtoGRPCBridge = &HTTPtoGRPCBridge{
	HeaderMapping: map[string]string{
		"Content-Type":  "content-type",
		"Authorization": "authorization",
		"User-Agent":    "user-agent",
		"X-Request-Id":  "x-request-id",
		"X-Trace-Id":    "x-trace-id",
	},
}

// HTTPRequestToGRPCContext converts HTTP request to gRPC context with metadata
func (b *HTTPtoGRPCBridge) HTTPRequestToGRPCContext(ctx *fasthttp.RequestCtx) context.Context {
	md := metadata.New(nil)

	// Extract headers
	if b.MetadataExtractor != nil {
		md = b.MetadataExtractor(ctx)
	} else {
		// Default: map configured headers
		for httpHeader, grpcKey := range b.HeaderMapping {
			value := string(ctx.Request.Header.Peek(httpHeader))
			if value != "" {
				md.Set(grpcKey, value)
			}
		}
	}

	// Create context with metadata
	grpcCtx := metadata.NewOutgoingContext(context.Background(), md)

	return grpcCtx
}

// GRPCResponseToHTTPResponse converts gRPC response to HTTP response
func (b *HTTPtoGRPCBridge) GRPCResponseToHTTPResponse(
	ctx *fasthttp.RequestCtx,
	resp proto.Message,
	header metadata.MD,
	trailer metadata.MD,
) error {
	// Set response headers from gRPC metadata
	for key, values := range header {
		for _, value := range values {
			ctx.Response.Header.Add(key, value)
		}
	}

	// Marshal protobuf response
	data, err := proto.Marshal(resp)
	if err != nil {
		log.Errorf("err:%v", err)
		if b.ErrorHandler != nil {
			return b.ErrorHandler(err)
		}
		return err
	}

	// Set content type
	ctx.Response.Header.Set("Content-Type", "application/protobuf")
	ctx.Response.SetBody(data)

	// Set trailer headers
	for key, values := range trailer {
		for _, value := range values {
			ctx.Response.Header.Add("Trailer-"+key, value)
		}
	}

	return nil
}

// StreamHandler handles gRPC streaming over HTTP
type StreamHandler struct {
	bridge *HTTPtoGRPCBridge
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(bridge *HTTPtoGRPCBridge) *StreamHandler {
	if bridge == nil {
		bridge = DefaultHTTPtoGRPCBridge
	}
	return &StreamHandler{
		bridge: bridge,
	}
}

// HandleServerStream handles server-side streaming from gRPC to HTTP
func (h *StreamHandler) HandleServerStream(
	ctx *fasthttp.RequestCtx,
	stream grpc.ServerStream,
) error {
	ctx.Response.Header.Set("Content-Type", "application/protobuf")
	ctx.Response.Header.Set("Transfer-Encoding", "chunked")

	for {
		var msg proto.Message
		err := stream.RecvMsg(&msg)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		// Marshal message
		data, err := proto.Marshal(msg)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		// Write chunk
		_, err = ctx.Write(data)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	return nil
}

// GRPCServiceAdapter adapts a gRPC service to HTTP handlers
type GRPCServiceAdapter struct {
	bridge *HTTPtoGRPCBridge
}

// NewGRPCServiceAdapter creates a new service adapter
func NewGRPCServiceAdapter(bridge *HTTPtoGRPCBridge) *GRPCServiceAdapter {
	if bridge == nil {
		bridge = DefaultHTTPtoGRPCBridge
	}
	return &GRPCServiceAdapter{
		bridge: bridge,
	}
}

// UnaryHandler adapts a gRPC unary handler to HTTP
func (a *GRPCServiceAdapter) UnaryHandler(
	handler func(ctx context.Context, req proto.Message) (proto.Message, error),
	reqType proto.Message,
) func(ctx *fasthttp.RequestCtx) error {
	return func(ctx *fasthttp.RequestCtx) error {
		// Parse request body
		req := proto.Clone(reqType)
		err := proto.Unmarshal(ctx.Request.Body(), req)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		// Convert HTTP context to gRPC context
		grpcCtx := a.bridge.HTTPRequestToGRPCContext(ctx)

		// Call gRPC handler
		resp, err := handler(grpcCtx, req)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}

		// Convert response to HTTP
		return a.bridge.GRPCResponseToHTTPResponse(ctx, resp, nil, nil)
	}
}

// MetadataCarrier implements gRPC metadata carrier for HTTP headers
type MetadataCarrier struct {
	ctx *fasthttp.RequestCtx
}

// NewMetadataCarrier creates a new metadata carrier
func NewMetadataCarrier(ctx *fasthttp.RequestCtx) *MetadataCarrier {
	return &MetadataCarrier{ctx: ctx}
}

// Get implements metadata.MD Get
func (m *MetadataCarrier) Get(key string) []string {
	value := m.ctx.Request.Header.Peek(key)
	if value == nil {
		return nil
	}
	return []string{string(value)}
}

// Set implements metadata.MD Set
func (m *MetadataCarrier) Set(key, value string) {
	m.ctx.Response.Header.Set(key, value)
}

// Append implements metadata.MD Append
func (m *MetadataCarrier) Append(key, value string) {
	m.ctx.Response.Header.Add(key, value)
}

// GRPCErrorHandler handles gRPC errors and converts them to HTTP responses
type GRPCErrorHandler struct {
	// StatusMapper maps gRPC status codes to HTTP status codes
	StatusMapper func(code int) int
}

// DefaultGRPCErrorHandler returns default error handler
var DefaultGRPCErrorHandler = &GRPCErrorHandler{
	StatusMapper: defaultStatusMapper,
}

func defaultStatusMapper(code int) int {
	// Map gRPC codes to HTTP status codes
	// Based on https://github.com/grpc/grpc/blob/master/doc/http-grpc-status-mapping.md
	switch code {
	case 0: // OK
		return 200
	case 1: // CANCELLED
		return 499
	case 2: // UNKNOWN
		return 500
	case 3: // INVALID_ARGUMENT
		return 400
	case 4: // DEADLINE_EXCEEDED
		return 504
	case 5: // NOT_FOUND
		return 404
	case 6: // ALREADY_EXISTS
		return 409
	case 7: // PERMISSION_DENIED
		return 403
	case 8: // RESOURCE_EXHAUSTED
		return 429
	case 9: // FAILED_PRECONDITION
		return 400
	case 10: // ABORTED
		return 409
	case 11: // OUT_OF_RANGE
		return 400
	case 12: // UNIMPLEMENTED
		return 501
	case 13: // INTERNAL
		return 500
	case 14: // UNAVAILABLE
		return 503
	case 15: // DATA_LOSS
		return 500
	case 16: // UNAUTHENTICATED
		return 401
	default:
		return 500
	}
}

// HandleGRPCError converts gRPC error to HTTP error response
func (h *GRPCErrorHandler) HandleGRPCError(ctx *fasthttp.RequestCtx, err error) {
	// In real implementation, would extract gRPC status code
	// For now, set a generic error response
	ctx.SetStatusCode(500)
	ctx.SetBodyString(`{"error":"Internal Server Error"}`)
}

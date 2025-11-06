package compress

import (
	"compress/gzip"
	"strings"

	"github.com/valyala/fasthttp"
)

// Level represents compression level
type Level int

const (
	LevelDefault Level = -1
	LevelBest    Level = gzip.BestCompression
	LevelFast    Level = gzip.BestSpeed
)

// Config defines the config for compression middleware
type Config struct {
	// Level sets the compression level
	Level Level

	// MinLength sets minimum response size to compress
	MinLength int

	// SkipFunc allows skipping compression for certain requests
	SkipFunc func(ctx *fasthttp.RequestCtx) bool
}

// DefaultConfig is the default compression config
var DefaultConfig = Config{
	Level:     LevelDefault,
	MinLength: 1024, // 1KB
	SkipFunc:  nil,
}

// CompressResponse compresses response if conditions are met
func CompressResponse(ctx *fasthttp.RequestCtx, cfg Config) {
	// Check if client accepts gzip
	acceptEncoding := string(ctx.Request.Header.Peek("Accept-Encoding"))
	if !strings.Contains(acceptEncoding, "gzip") {
		return
	}

	// Check minimum length
	body := ctx.Response.Body()
	if len(body) < cfg.MinLength {
		return
	}

	// Check skip function
	if cfg.SkipFunc != nil && cfg.SkipFunc(ctx) {
		return
	}

	// Compress
	var w *gzip.Writer
	if cfg.Level == LevelDefault {
		w = gzip.NewWriter(nil)
	} else {
		var err error
		w, err = gzip.NewWriterLevel(nil, int(cfg.Level))
		if err != nil {
			return
		}
	}
	defer w.Close()

	// Create compressed body
	compressed := make([]byte, 0, len(body))
	buffer := &writeBuffer{buf: &compressed}
	w.Reset(buffer)

	_, err := w.Write(body)
	if err != nil {
		return
	}

	err = w.Close()
	if err != nil {
		return
	}

	// Update response
	ctx.Response.SetBody(compressed)
	ctx.Response.Header.Set("Content-Encoding", "gzip")
	ctx.Response.Header.Del("Content-Length")
}

type writeBuffer struct {
	buf *[]byte
}

func (w *writeBuffer) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// StreamWriter wraps an io.Writer for streaming responses
type StreamWriter struct {
	ctx    *fasthttp.RequestCtx
	header bool
}

// NewStreamWriter creates a new stream writer
func NewStreamWriter(ctx *fasthttp.RequestCtx) *StreamWriter {
	return &StreamWriter{
		ctx:    ctx,
		header: false,
	}
}

// Write implements io.Writer
func (s *StreamWriter) Write(p []byte) (n int, err error) {
	if !s.header {
		s.ctx.Response.Header.Set("Transfer-Encoding", "chunked")
		s.header = true
	}

	// Write to response body
	n, err = s.ctx.Write(p)
	return n, err
}

// Flush flushes the stream
func (s *StreamWriter) Flush() error {
	return nil
}

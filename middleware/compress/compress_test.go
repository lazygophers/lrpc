package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestCompressResponse(t *testing.T) {
	t.Run("compress large response", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Accept-Encoding", "gzip")

		// Create a large response body (> 1KB)
		largeBody := bytes.Repeat([]byte("test data "), 200)
		ctx.Response.SetBody(largeBody)

		config := DefaultConfig
		CompressResponse(ctx, config)

		// Check compression happened
		assert.Equal(t, "gzip", string(ctx.Response.Header.Peek("Content-Encoding")))
		assert.Less(t, len(ctx.Response.Body()), len(largeBody))

		// Verify decompression works
		reader, err := gzip.NewReader(bytes.NewReader(ctx.Response.Body()))
		require.NoError(t, err)
		decompressed, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, largeBody, decompressed)
	})

	t.Run("skip compression for small response", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Accept-Encoding", "gzip")

		smallBody := []byte("small")
		ctx.Response.SetBody(smallBody)

		config := DefaultConfig
		CompressResponse(ctx, config)

		// Should not compress
		assert.Empty(t, string(ctx.Response.Header.Peek("Content-Encoding")))
		assert.Equal(t, smallBody, ctx.Response.Body())
	})

	t.Run("skip compression when client doesn't accept gzip", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		// No Accept-Encoding header

		largeBody := bytes.Repeat([]byte("test data "), 200)
		ctx.Response.SetBody(largeBody)

		config := DefaultConfig
		CompressResponse(ctx, config)

		// Should not compress
		assert.Empty(t, string(ctx.Response.Header.Peek("Content-Encoding")))
		assert.Equal(t, largeBody, ctx.Response.Body())
	})

	t.Run("skip compression with custom skip function", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Accept-Encoding", "gzip")

		largeBody := bytes.Repeat([]byte("test data "), 200)
		ctx.Response.SetBody(largeBody)

		config := Config{
			Level:     LevelDefault,
			MinLength: 1024,
			SkipFunc: func(ctx *fasthttp.RequestCtx) bool {
				return true // Always skip
			},
		}
		CompressResponse(ctx, config)

		// Should not compress
		assert.Empty(t, string(ctx.Response.Header.Peek("Content-Encoding")))
		assert.Equal(t, largeBody, ctx.Response.Body())
	})

	t.Run("compress with best compression level", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Accept-Encoding", "gzip")

		largeBody := bytes.Repeat([]byte("test data "), 200)
		ctx.Response.SetBody(largeBody)

		config := Config{
			Level:     LevelBest,
			MinLength: 1024,
		}
		CompressResponse(ctx, config)

		assert.Equal(t, "gzip", string(ctx.Response.Header.Peek("Content-Encoding")))
		assert.Less(t, len(ctx.Response.Body()), len(largeBody))
	})

	t.Run("compress with fast compression level", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Accept-Encoding", "gzip")

		largeBody := bytes.Repeat([]byte("test data "), 200)
		ctx.Response.SetBody(largeBody)

		config := Config{
			Level:     LevelFast,
			MinLength: 1024,
		}
		CompressResponse(ctx, config)

		assert.Equal(t, "gzip", string(ctx.Response.Header.Peek("Content-Encoding")))
		assert.Less(t, len(ctx.Response.Body()), len(largeBody))
	})

	t.Run("custom minimum length", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Accept-Encoding", "gzip")

		// Body size is 100 bytes
		body := bytes.Repeat([]byte("x"), 100)
		ctx.Response.SetBody(body)

		config := Config{
			Level:     LevelDefault,
			MinLength: 50, // Lower threshold
		}
		CompressResponse(ctx, config)

		// Should compress because body > minLength
		assert.Equal(t, "gzip", string(ctx.Response.Header.Peek("Content-Encoding")))
	})

	t.Run("accept-encoding with multiple encodings", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Request.Header.Set("Accept-Encoding", "deflate, gzip, br")

		largeBody := bytes.Repeat([]byte("test data "), 200)
		ctx.Response.SetBody(largeBody)

		config := DefaultConfig
		CompressResponse(ctx, config)

		// Should compress with gzip
		assert.Equal(t, "gzip", string(ctx.Response.Header.Peek("Content-Encoding")))
	})
}

func TestCompressionLevels(t *testing.T) {
	t.Run("compare compression levels", func(t *testing.T) {
		// Create test data
		testData := bytes.Repeat([]byte("This is test data for compression. "), 100)

		levels := []Level{LevelFast, LevelDefault, LevelBest}
		sizes := make([]int, len(levels))

		for i, level := range levels {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.Header.Set("Accept-Encoding", "gzip")
			ctx.Response.SetBody(testData)

			config := Config{
				Level:     level,
				MinLength: 100,
			}
			CompressResponse(ctx, config)

			sizes[i] = len(ctx.Response.Body())
		}

		// Best compression should produce smallest size
		assert.LessOrEqual(t, sizes[2], sizes[1])
		assert.LessOrEqual(t, sizes[2], sizes[0])
	})
}

func TestStreamWriter(t *testing.T) {
	t.Run("create stream writer", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		writer := NewStreamWriter(ctx)

		assert.NotNil(t, writer)
		assert.Equal(t, ctx, writer.ctx)
		assert.False(t, writer.header)
	})

	t.Run("write to stream", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		writer := NewStreamWriter(ctx)

		data := []byte("test data")
		n, err := writer.Write(data)

		require.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.True(t, writer.header)
		// Note: Header might not be set immediately in fasthttp.RequestCtx mock
	})

	t.Run("multiple writes", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		writer := NewStreamWriter(ctx)

		// First write
		n1, err := writer.Write([]byte("first "))
		require.NoError(t, err)
		assert.Equal(t, 6, n1)

		// Second write
		n2, err := writer.Write([]byte("second"))
		require.NoError(t, err)
		assert.Equal(t, 6, n2)

		// Header should only be set once
		assert.True(t, writer.header)
	})

	t.Run("flush does not error", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		writer := NewStreamWriter(ctx)

		err := writer.Flush()
		assert.NoError(t, err)
	})
}

func TestConfig(t *testing.T) {
	t.Run("default config values", func(t *testing.T) {
		config := DefaultConfig

		assert.Equal(t, LevelDefault, config.Level)
		assert.Equal(t, 1024, config.MinLength)
		assert.Nil(t, config.SkipFunc)
	})

	t.Run("custom config", func(t *testing.T) {
		config := Config{
			Level:     LevelBest,
			MinLength: 2048,
			SkipFunc: func(ctx *fasthttp.RequestCtx) bool {
				return false
			},
		}

		assert.Equal(t, LevelBest, config.Level)
		assert.Equal(t, 2048, config.MinLength)
		assert.NotNil(t, config.SkipFunc)
	})
}

func TestWriteBuffer(t *testing.T) {
	t.Run("write to buffer", func(t *testing.T) {
		buf := make([]byte, 0)
		wb := &writeBuffer{buf: &buf}

		data := []byte("test data")
		n, err := wb.Write(data)

		require.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.Equal(t, data, buf)
	})

	t.Run("multiple writes append", func(t *testing.T) {
		buf := make([]byte, 0)
		wb := &writeBuffer{buf: &buf}

		wb.Write([]byte("first "))
		wb.Write([]byte("second"))

		assert.Equal(t, []byte("first second"), buf)
	})
}

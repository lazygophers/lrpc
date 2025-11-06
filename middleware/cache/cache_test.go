package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestBuildCacheControl(t *testing.T) {
	t.Run("basic max-age and public", func(t *testing.T) {
		config := CacheConfig{
			MaxAge: 3600,
			Public: true,
		}

		cc := BuildCacheControl(config)
		assert.Contains(t, cc, "max-age=3600")
		assert.Contains(t, cc, "public")
	})

	t.Run("private cache", func(t *testing.T) {
		config := CacheConfig{
			MaxAge:  1800,
			Private: true,
		}

		cc := BuildCacheControl(config)
		assert.Contains(t, cc, "max-age=1800")
		assert.Contains(t, cc, "private")
		assert.NotContains(t, cc, "public")
	})

	t.Run("no-cache directive", func(t *testing.T) {
		config := CacheConfig{
			NoCache: true,
		}

		cc := BuildCacheControl(config)
		assert.Contains(t, cc, "no-cache")
	})

	t.Run("no-store overrides other directives", func(t *testing.T) {
		config := CacheConfig{
			NoStore: true,
			MaxAge:  3600,
			Public:  true,
		}

		cc := BuildCacheControl(config)
		assert.Equal(t, "no-store", cc)
	})

	t.Run("must-revalidate", func(t *testing.T) {
		config := CacheConfig{
			MaxAge:         3600,
			MustRevalidate: true,
		}

		cc := BuildCacheControl(config)
		assert.Contains(t, cc, "must-revalidate")
	})

	t.Run("s-maxage for shared caches", func(t *testing.T) {
		config := CacheConfig{
			MaxAge:  3600,
			SMaxAge: 7200,
		}

		cc := BuildCacheControl(config)
		assert.Contains(t, cc, "max-age=3600")
		assert.Contains(t, cc, "s-maxage=7200")
	})

	t.Run("immutable directive", func(t *testing.T) {
		config := CacheConfig{
			MaxAge:    31536000,
			Immutable: true,
		}

		cc := BuildCacheControl(config)
		assert.Contains(t, cc, "immutable")
	})

	t.Run("stale-while-revalidate", func(t *testing.T) {
		config := CacheConfig{
			MaxAge:               3600,
			StaleWhileRevalidate: 86400,
		}

		cc := BuildCacheControl(config)
		assert.Contains(t, cc, "stale-while-revalidate=86400")
	})

	t.Run("stale-if-error", func(t *testing.T) {
		config := CacheConfig{
			MaxAge:       3600,
			StaleIfError: 86400,
		}

		cc := BuildCacheControl(config)
		assert.Contains(t, cc, "stale-if-error=86400")
	})
}

func TestSetCacheHeaders(t *testing.T) {
	t.Run("set basic cache headers", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		config := CacheConfig{
			MaxAge: 3600,
			Public: true,
		}

		SetCacheHeaders(ctx, config)

		cc := string(ctx.Response.Header.Peek("Cache-Control"))
		assert.Contains(t, cc, "max-age=3600")
		assert.Contains(t, cc, "public")
	})

	t.Run("set Vary header", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		config := CacheConfig{
			MaxAge: 3600,
			Vary:   []string{"Accept-Encoding", "Accept-Language"},
		}

		SetCacheHeaders(ctx, config)

		vary := string(ctx.Response.Header.Peek("Vary"))
		assert.Contains(t, vary, "Accept-Encoding")
		assert.Contains(t, vary, "Accept-Language")
	})

	t.Run("generate ETag for response body", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Response.SetBodyString("test response body")

		config := CacheConfig{
			MaxAge:   3600,
			WeakETag: false,
		}

		SetCacheHeaders(ctx, config)

		etag := string(ctx.Response.Header.Peek("ETag"))
		assert.NotEmpty(t, etag)
		assert.Contains(t, etag, `"`)
		assert.NotContains(t, etag, "W/")
	})

	t.Run("generate weak ETag", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Response.SetBodyString("test response body")

		config := CacheConfig{
			MaxAge:   3600,
			WeakETag: true,
		}

		SetCacheHeaders(ctx, config)

		etag := string(ctx.Response.Header.Peek("ETag"))
		assert.NotEmpty(t, etag)
		assert.Contains(t, etag, "W/")
	})

	t.Run("custom ETag generator", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Response.SetBodyString("test")

		config := CacheConfig{
			MaxAge: 3600,
			ETagGenerator: func(body []byte) string {
				return "custom-etag-123"
			},
		}

		SetCacheHeaders(ctx, config)

		etag := string(ctx.Response.Header.Peek("ETag"))
		assert.Equal(t, "custom-etag-123", etag)
	})

	t.Run("skip when SkipFunc returns true", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		config := CacheConfig{
			MaxAge: 3600,
			SkipFunc: func(ctx *fasthttp.RequestCtx) bool {
				return true
			},
		}

		SetCacheHeaders(ctx, config)

		cc := string(ctx.Response.Header.Peek("Cache-Control"))
		assert.Empty(t, cc)
	})

	t.Run("handle If-None-Match with matching ETag", func(t *testing.T) {
		ctx := &fasthttp.RequestCtx{}
		ctx.Response.SetBodyString("test response body")

		// First request - get ETag
		config := CacheConfig{
			MaxAge: 3600,
		}
		SetCacheHeaders(ctx, config)
		etag := string(ctx.Response.Header.Peek("ETag"))

		// Second request - with If-None-Match
		ctx2 := &fasthttp.RequestCtx{}
		ctx2.Request.Header.Set("If-None-Match", etag)
		ctx2.Response.SetBodyString("test response body")

		SetCacheHeaders(ctx2, config)

		// Should return 304 Not Modified
		assert.Equal(t, 304, ctx2.Response.StatusCode())
	})
}

func TestGenerateETag(t *testing.T) {
	t.Run("consistent ETag for same content", func(t *testing.T) {
		content := []byte("test content")
		etag1 := GenerateETag(content, false)
		etag2 := GenerateETag(content, false)

		assert.Equal(t, etag1, etag2)
		assert.NotEmpty(t, etag1)
	})

	t.Run("different ETags for different content", func(t *testing.T) {
		content1 := []byte("content one")
		content2 := []byte("content two")

		etag1 := GenerateETag(content1, false)
		etag2 := GenerateETag(content2, false)

		assert.NotEqual(t, etag1, etag2)
	})

	t.Run("weak ETag format", func(t *testing.T) {
		content := []byte("test")
		etag := GenerateETag(content, true)

		assert.Contains(t, etag, "W/")
		assert.Contains(t, etag, `"`)
	})

	t.Run("strong ETag format", func(t *testing.T) {
		content := []byte("test")
		etag := GenerateETag(content, false)

		assert.NotContains(t, etag, "W/")
		assert.Contains(t, etag, `"`)
	})
}

func TestCacheConfig(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		config := DefaultCacheConfig

		assert.Equal(t, 3600, config.MaxAge)
		assert.True(t, config.Public)
		assert.False(t, config.WeakETag)
		assert.NotEmpty(t, config.Vary)
	})

	t.Run("custom config", func(t *testing.T) {
		config := CacheConfig{
			MaxAge:         7200,
			Private:        true,
			NoCache:        false,
			MustRevalidate: true,
		}

		assert.Equal(t, 7200, config.MaxAge)
		assert.True(t, config.Private)
		assert.False(t, config.NoCache)
		assert.True(t, config.MustRevalidate)
	})
}

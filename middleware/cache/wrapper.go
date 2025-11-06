package cache

import "github.com/lazygophers/lrpc"

// Cache returns an HTTP caching middleware
func Cache(config ...CacheConfig) func(ctx *lrpc.Ctx) error {
	cfg := DefaultCacheConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *lrpc.Ctx) error {
		// Execute next handlers
		err := ctx.Next()

		// Set cache headers if no error
		if err == nil {
			SetCacheHeaders(ctx.Context(), cfg)
		}

		return err
	}
}

// CacheWithETag returns a caching middleware that uses ETags
func CacheWithETag(maxAge int, weak bool) func(ctx *lrpc.Ctx) error {
	return Cache(CacheConfig{
		MaxAge:   maxAge,
		Public:   true,
		WeakETag: weak,
	})
}

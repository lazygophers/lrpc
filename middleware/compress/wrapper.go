package compress

import "github.com/lazygophers/lrpc"

// Compress returns a middleware that compresses HTTP responses
func Compress(config ...Config) func(ctx *lrpc.Ctx) error {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *lrpc.Ctx) error {
		// Execute next handlers
		err := ctx.Next()

		// Compress response if applicable
		CompressResponse(ctx.Context(), cfg)

		return err
	}
}

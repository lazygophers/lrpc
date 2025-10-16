package lrpc

import (
	"runtime/debug"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/valyala/fasthttp"
)

// RecoverConfig defines the config for Recover middleware
type RecoverConfig struct {
	// EnableStackTrace enables printing stack trace
	EnableStackTrace bool

	// StackTraceHandler allows custom stack trace handling
	StackTraceHandler func(ctx *Ctx, err interface{}, stack []byte)

	// ErrorHandler allows custom error response
	ErrorHandler func(ctx *Ctx, err interface{})
}

// DefaultRecoverConfig is the default Recover middleware config
var DefaultRecoverConfig = RecoverConfig{
	EnableStackTrace: true,
	StackTraceHandler: func(ctx *Ctx, err interface{}, stack []byte) {
		log.Errorf("[Panic Recovered] %v\nStack Trace:\n%s", err, string(stack))
	},
	ErrorHandler: nil,
}

// Recover returns a middleware that recovers from panics
func Recover(config ...RecoverConfig) HandlerFunc {
	cfg := DefaultRecoverConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				var stack []byte
				if cfg.EnableStackTrace {
					stack = debug.Stack()
					if cfg.StackTraceHandler != nil {
						cfg.StackTraceHandler(ctx, r, stack)
					}
				}

				if cfg.ErrorHandler != nil {
					cfg.ErrorHandler(ctx, r)
					return
				}

				// Default error handling
				ctx.Context().SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SendJson(map[string]interface{}{
					"code":    int32(core.ErrCode_StatusInternalServerError),
					"message": "Internal Server Error",
				})
			}
		}()

		return ctx.Next()
	}
}

// BodyLimit returns a middleware that limits request body size
func BodyLimit(maxSize int) HandlerFunc {
	return func(ctx *Ctx) error {
		// Check body size
		if ctx.Context().Request.Header.ContentLength() > maxSize {
			ctx.Context().SetStatusCode(fasthttp.StatusRequestEntityTooLarge)
			return ctx.SendJson(map[string]interface{}{
				"code":    fasthttp.StatusRequestEntityTooLarge,
				"message": "Request Entity Too Large",
			})
		}

		return ctx.Next()
	}
}

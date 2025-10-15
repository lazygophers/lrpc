package recover

import (
	"fmt"
	"runtime/debug"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/valyala/fasthttp"
)

// Config defines the config for Recover middleware
type Config struct {
	// EnableStackTrace enables printing stack trace
	EnableStackTrace bool

	// StackTraceHandler allows custom stack trace handling
	StackTraceHandler func(ctx *fasthttp.RequestCtx, err interface{}, stack []byte)

	// ErrorHandler allows custom error response
	ErrorHandler func(ctx *fasthttp.RequestCtx, err interface{})
}

// DefaultConfig is the default Recover middleware config
var DefaultConfig = Config{
	EnableStackTrace: true,
	StackTraceHandler: func(ctx *fasthttp.RequestCtx, err interface{}, stack []byte) {
		log.Errorf("[Panic Recovered] %v\nStack Trace:\n%s", err, string(stack))
	},
	ErrorHandler: nil, // Use default error handler
}

// New returns a middleware that recovers from panics
func New(config ...Config) func(next func(*fasthttp.RequestCtx)) func(*fasthttp.RequestCtx) {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(next func(*fasthttp.RequestCtx)) func(*fasthttp.RequestCtx) {
		return func(ctx *fasthttp.RequestCtx) {
			defer func() {
				if r := recover(); r != nil {
					var stack []byte
					if cfg.EnableStackTrace {
						stack = debug.Stack()
						if cfg.StackTraceHandler != nil {
							cfg.StackTraceHandler(ctx, r, stack)
						}
					}

					// Use custom error handler if provided
					if cfg.ErrorHandler != nil {
						cfg.ErrorHandler(ctx, r)
						return
					}

					// Default error handling
					ctx.SetStatusCode(fasthttp.StatusInternalServerError)

					// Try to create a proper error response
					var errMsg string
					switch e := r.(type) {
					case error:
						errMsg = e.Error()
					case string:
						errMsg = e
					default:
						errMsg = fmt.Sprintf("%v", e)
					}

					ctx.SetContentType("application/json")
					ctx.SetBodyString(fmt.Sprintf(`{"code":%d,"message":"Internal Server Error","hint":"%s"}`,
						core.ErrCode_StatusInternalServerError, errMsg))
				}
			}()

			next(ctx)
		}
	}
}

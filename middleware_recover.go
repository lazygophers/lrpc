package lrpc

import (
	"fmt"
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
	ErrorHandler: nil, // Use default error handler
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

				// Use custom error handler if provided
				if cfg.ErrorHandler != nil {
					cfg.ErrorHandler(ctx, r)
					return
				}

				// Default error handling
				ctx.Context().SetStatusCode(fasthttp.StatusInternalServerError)

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

				err := ctx.SendJson(map[string]interface{}{
					"code":    int32(core.ErrCode_StatusInternalServerError),
					"message": "Internal Server Error",
					"hint":    errMsg,
				})
				if err != nil {
					log.Errorf("err:%v", err)
				}
			}
		}()

		return ctx.Next()
	}
}

// RecoverHandler wraps the entire Handler with panic recovery
func (app *App) RecoverHandler(c *fasthttp.RequestCtx) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			log.Errorf("[Panic in Handler] %v\nStack Trace:\n%s", r, string(stack))

			c.SetStatusCode(fasthttp.StatusInternalServerError)
			c.SetBodyString(fmt.Sprintf("Internal Server Error: %v", r))
		}
	}()

	app.Handler(c)
}

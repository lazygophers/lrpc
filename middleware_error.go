package lrpc

import (
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/valyala/fasthttp"
)

// ErrorHandlerConfig defines the config for error handling middleware
type ErrorHandlerConfig struct {
	// CustomHandler allows custom error handling logic
	CustomHandler func(ctx *Ctx, err error) error

	// LogErrors enables error logging
	LogErrors bool

	// SendErrorDetails sends detailed error messages to client
	SendErrorDetails bool
}

// DefaultErrorHandlerConfig is the default error handling config
var DefaultErrorHandlerConfig = ErrorHandlerConfig{
	CustomHandler:    nil,
	LogErrors:        true,
	SendErrorDetails: false, // Don't send details by default for security
}

// ErrorHandler returns a middleware that handles errors uniformly
func ErrorHandler(config ...ErrorHandlerConfig) HandlerFunc {
	cfg := DefaultErrorHandlerConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *Ctx) error {
		// Execute next handlers
		err := ctx.Next()

		// No error, continue
		if err == nil {
			return nil
		}

		// Log error if enabled
		if cfg.LogErrors {
			log.Errorf("Request error: %v | Path: %s %s | TraceID: %s",
				err, ctx.Method(), ctx.Path(), ctx.TraceID())
		}

		// Use custom handler if provided
		if cfg.CustomHandler != nil {
			return cfg.CustomHandler(ctx, err)
		}

		// Default error handling based on error type
		return handleError(ctx, err, cfg.SendErrorDetails)
	}
}

// handleError processes errors and sends appropriate responses
func handleError(ctx *Ctx, err error, sendDetails bool) error {
	// Check if it's an xerror.Error
	if xe, ok := err.(*xerror.Error); ok {
		return handleXError(ctx, xe)
	}

	// Default error handling
	statusCode := fasthttp.StatusInternalServerError
	errCode := int32(core.ErrCode_StatusInternalServerError)
	message := "Internal Server Error"

	if sendDetails {
		message = err.Error()
	}

	ctx.Context().SetStatusCode(statusCode)
	return ctx.SendJson(&BaseResponse{
		Code:    errCode,
		Message: message,
		Hint:    log.GetTrace(),
	})
}

// handleXError handles xerror.Error instances
func handleXError(ctx *Ctx, err *xerror.Error) error {
	statusCode := fasthttp.StatusOK

	// Map error codes to HTTP status codes
	switch err.Code {
	case int32(core.ErrCode_StatusBadRequest):
		statusCode = fasthttp.StatusBadRequest
	case int32(core.ErrCode_StatusUnauthorized):
		statusCode = fasthttp.StatusUnauthorized
	case int32(core.ErrCode_StatusForbidden):
		statusCode = fasthttp.StatusForbidden
	case int32(core.ErrCode_StatusNotFound):
		statusCode = fasthttp.StatusNotFound
	case int32(core.ErrCode_StatusConflict):
		statusCode = fasthttp.StatusConflict
	case int32(core.ErrCode_StatusInternalServerError):
		statusCode = fasthttp.StatusInternalServerError
	default:
		// For custom error codes, use 200 OK
		statusCode = fasthttp.StatusOK
	}

	ctx.Context().SetStatusCode(statusCode)

	return ctx.SendJson(&BaseResponse{
		Code:    err.Code,
		Message: err.Msg,
		Hint:    log.GetTrace(),
	})
}

// SetErrorHandler sets a global error handler for the app
func (app *App) SetErrorHandler(handler func(ctx *Ctx, err error)) {
	app.c.OnError = handler
}

// NotFound sets a custom 404 handler for a specific method
func (app *App) SetNotFound(method string, handler HandlerFunc) error {
	router, ok := app.routers[method]
	if !ok {
		router = NewRouter()
		router.middleware = append(router.middleware, app.globalMiddleware...)
		app.routers[method] = router
	}

	router.notFound = handler
	return nil
}

// SetNotFoundAll sets a custom 404 handler for all methods
func (app *App) SetNotFoundAll(handler HandlerFunc) error {
	for method := range app.routers {
		err := app.SetNotFound(method, handler)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}
	return nil
}

package errors

import (
	"fmt"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/valyala/fasthttp"
)

// Config defines the config for error handling middleware
type Config struct {
	// CustomHandler allows custom error handling logic
	CustomHandler func(ctx *fasthttp.RequestCtx, err error) error

	// LogErrors enables error logging
	LogErrors bool

	// SendErrorDetails sends detailed error messages to client
	SendErrorDetails bool
}

// DefaultConfig is the default error handling config
var DefaultConfig = Config{
	CustomHandler:    nil,
	LogErrors:        true,
	SendErrorDetails: false, // Don't send details by default for security
}

// HandleError processes errors and sends appropriate responses
func HandleError(ctx *fasthttp.RequestCtx, err error, sendDetails bool) {
	// Check if it's an xerror.Error
	if xe, ok := err.(*xerror.Error); ok {
		handleXError(ctx, xe)
		return
	}

	// Default error handling
	statusCode := fasthttp.StatusInternalServerError
	errCode := int32(core.ErrCode_StatusInternalServerError)
	message := "Internal Server Error"

	if sendDetails {
		message = err.Error()
	}

	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	ctx.SetBodyString(fmt.Sprintf(`{"code":%d,"message":"%s","hint":"%s"}`,
		errCode, message, log.GetTrace()))
}

// handleXError handles xerror.Error instances
func handleXError(ctx *fasthttp.RequestCtx, err *xerror.Error) {
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

	ctx.SetStatusCode(statusCode)
	ctx.SetContentType("application/json")
	ctx.SetBodyString(fmt.Sprintf(`{"code":%d,"message":"%s","hint":"%s"}`,
		err.Code, err.Msg, log.GetTrace()))
}

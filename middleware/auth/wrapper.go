package auth

import (
	"github.com/lazygophers/lrpc"
	"github.com/valyala/fasthttp"
)

// JWT returns a JWT authentication middleware
func JWT(config JWTConfig) func(ctx *lrpc.Ctx) error {
	// Set defaults
	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultJWTConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultJWTConfig.AuthScheme
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultJWTConfig.ContextKey
	}
	if config.Claims == nil {
		config.Claims = DefaultJWTConfig.Claims
	}

	// Default error handler
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(`{"code":401,"message":"Unauthorized"}`)
		}
	}

	return func(ctx *lrpc.Ctx) error {
		err := ValidateJWT(ctx.Context(), config)
		if err != nil {
			config.ErrorHandler(ctx.Context(), err)
			return err
		}
		return ctx.Next()
	}
}

// BasicAuth returns a basic authentication middleware
func BasicAuth(config BasicAuthConfig) func(ctx *lrpc.Ctx) error {
	// Set defaults
	if config.Realm == "" {
		config.Realm = DefaultBasicAuthConfig.Realm
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultBasicAuthConfig.ContextKey
	}

	// Default error handler
	if config.ErrorHandler == nil {
		config.ErrorHandler = func(ctx *fasthttp.RequestCtx, err error) {
			SetWWWAuthenticate(ctx, config.Realm)
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(`{"code":401,"message":"Unauthorized"}`)
		}
	}

	return func(ctx *lrpc.Ctx) error {
		err := ValidateBasicAuth(ctx.Context(), config)
		if err != nil {
			config.ErrorHandler(ctx.Context(), err)
			return err
		}
		return ctx.Next()
	}
}

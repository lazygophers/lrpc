package auth

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/lazygophers/log"
	"github.com/valyala/fasthttp"
)

var (
	ErrMissingCredentials = errors.New("missing basic auth credentials")
	ErrInvalidCredentials = errors.New("invalid basic auth credentials")
)

// BasicAuthConfig defines basic authentication configuration
type BasicAuthConfig struct {
	// Validator validates username and password
	// Should return true if credentials are valid
	Validator func(username, password string) bool

	// Users is a map of username to password for simple validation
	// If provided, this takes precedence over Validator
	Users map[string]string

	// Realm is the authentication realm
	Realm string

	// ContextKey is the key to store username in context
	ContextKey string

	// SkipFunc allows skipping authentication for certain requests
	SkipFunc func(ctx *fasthttp.RequestCtx) bool

	// ErrorHandler handles authentication errors
	ErrorHandler func(ctx *fasthttp.RequestCtx, err error)

	// SuccessHandler is called after successful authentication
	SuccessHandler func(ctx *fasthttp.RequestCtx, username string)
}

// DefaultBasicAuthConfig is the default basic auth configuration
var DefaultBasicAuthConfig = BasicAuthConfig{
	Realm:      "Restricted",
	ContextKey: "username",
}

// ParseBasicAuth parses HTTP Basic Authentication credentials from Authorization header
func ParseBasicAuth(ctx *fasthttp.RequestCtx) (username, password string, err error) {
	auth := string(ctx.Request.Header.Peek("Authorization"))
	if auth == "" {
		return "", "", ErrMissingCredentials
	}

	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return "", "", ErrInvalidCredentials
	}

	// Decode base64 credentials
	decoded, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		log.Errorf("err:%v", err)
		return "", "", ErrInvalidCredentials
	}

	// Split into username:password
	credentials := string(decoded)
	colonIndex := strings.Index(credentials, ":")
	if colonIndex < 0 {
		return "", "", ErrInvalidCredentials
	}

	username = credentials[:colonIndex]
	password = credentials[colonIndex+1:]

	return username, password, nil
}

// ValidateBasicAuth validates basic authentication credentials
func ValidateBasicAuth(ctx *fasthttp.RequestCtx, config BasicAuthConfig) error {
	// Skip if skip function returns true
	if config.SkipFunc != nil && config.SkipFunc(ctx) {
		return nil
	}

	// Parse credentials
	username, password, err := ParseBasicAuth(ctx)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// Validate credentials
	var valid bool
	if config.Users != nil {
		// Validate against user map
		expectedPassword, exists := config.Users[username]
		valid = exists && password == expectedPassword
	} else if config.Validator != nil {
		// Use custom validator
		valid = config.Validator(username, password)
	} else {
		err := errors.New("no validator or users configured")
		log.Errorf("err:%v", err)
		return err
	}

	if !valid {
		return ErrInvalidCredentials
	}

	// Store username in context
	if config.ContextKey != "" {
		ctx.SetUserValue(config.ContextKey, username)
	}

	// Call success handler if provided
	if config.SuccessHandler != nil {
		config.SuccessHandler(ctx, username)
	}

	return nil
}

// GetUsername retrieves authenticated username from context
func GetUsername(ctx *fasthttp.RequestCtx, contextKey string) (string, bool) {
	username := ctx.UserValue(contextKey)
	if username == nil {
		return "", false
	}

	usernameStr, ok := username.(string)
	return usernameStr, ok
}

// SetWWWAuthenticate sets WWW-Authenticate header for basic auth challenge
func SetWWWAuthenticate(ctx *fasthttp.RequestCtx, realm string) {
	if realm == "" {
		realm = "Restricted"
	}
	ctx.Response.Header.Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
}

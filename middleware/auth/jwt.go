package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lazygophers/log"
	"github.com/valyala/fasthttp"
)

var (
	ErrMissingToken     = errors.New("missing authorization token")
	ErrInvalidToken     = errors.New("invalid authorization token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
)

// JWTConfig defines JWT authentication configuration
type JWTConfig struct {
	// SigningKey is the secret key for signing tokens
	SigningKey string

	// SigningMethod is the signing method (HS256, HS384, HS512)
	SigningMethod string

	// TokenLookup defines where to look for the token
	// Format: "<source>:<name>"
	// Possible values: "header:Authorization", "query:token", "cookie:jwt"
	TokenLookup string

	// AuthScheme is the authorization scheme (Bearer, JWT)
	AuthScheme string

	// ContextKey is the key to store user info in context
	ContextKey string

	// Claims is the custom claims struct (must implement jwt.Claims)
	Claims jwt.Claims

	// SkipFunc allows skipping authentication for certain requests
	SkipFunc func(ctx *fasthttp.RequestCtx) bool

	// ErrorHandler handles authentication errors
	ErrorHandler func(ctx *fasthttp.RequestCtx, err error)

	// SuccessHandler is called after successful authentication
	SuccessHandler func(ctx *fasthttp.RequestCtx, token *jwt.Token)
}

// DefaultJWTConfig is the default JWT configuration
var DefaultJWTConfig = JWTConfig{
	SigningMethod: "HS256",
	TokenLookup:   "header:Authorization",
	AuthScheme:    "Bearer",
	ContextKey:    "user",
	Claims:        jwt.MapClaims{},
}

// JWTClaims represents standard JWT claims
type JWTClaims struct {
	UserID   string                 `json:"user_id"`
	Username string                 `json:"username"`
	Roles    []string               `json:"roles,omitempty"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token with the given claims
func GenerateToken(config JWTConfig, claims jwt.Claims) (string, error) {
	if config.SigningKey == "" {
		err := errors.New("signing key is required")
		log.Errorf("err:%v", err)
		return "", err
	}

	if config.SigningMethod == "" {
		config.SigningMethod = DefaultJWTConfig.SigningMethod
	}

	var method jwt.SigningMethod
	switch config.SigningMethod {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "HS384":
		method = jwt.SigningMethodHS384
	case "HS512":
		method = jwt.SigningMethodHS512
	default:
		err := errors.New("unsupported signing method")
		log.Errorf("err:%v", err)
		return "", err
	}

	token := jwt.NewWithClaims(method, claims)
	tokenString, err := token.SignedString([]byte(config.SigningKey))
	if err != nil {
		log.Errorf("err:%v", err)
		return "", err
	}

	return tokenString, nil
}

// ParseToken parses and validates a JWT token
func ParseToken(config JWTConfig, tokenString string) (*jwt.Token, error) {
	if config.SigningKey == "" {
		err := errors.New("signing key is required")
		log.Errorf("err:%v", err)
		return nil, err
	}

	token, err := jwt.ParseWithClaims(tokenString, config.Claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if token.Method.Alg() != config.SigningMethod {
			return nil, ErrInvalidSignature
		}
		return []byte(config.SigningKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		log.Errorf("err:%v", err)
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return token, nil
}

// ExtractToken extracts token from request based on TokenLookup config
func ExtractToken(ctx *fasthttp.RequestCtx, config JWTConfig) (string, error) {
	parts := strings.Split(config.TokenLookup, ":")
	if len(parts) != 2 {
		err := errors.New("invalid token lookup format")
		log.Errorf("err:%v", err)
		return "", err
	}

	source := parts[0]
	name := parts[1]

	var token string
	switch source {
	case "header":
		auth := string(ctx.Request.Header.Peek(name))
		if auth == "" {
			return "", ErrMissingToken
		}

		// Extract token from auth scheme (e.g., "Bearer <token>")
		if config.AuthScheme != "" {
			prefix := config.AuthScheme + " "
			if !strings.HasPrefix(auth, prefix) {
				return "", ErrInvalidToken
			}
			token = strings.TrimPrefix(auth, prefix)
		} else {
			token = auth
		}

	case "query":
		token = string(ctx.QueryArgs().Peek(name))
		if token == "" {
			return "", ErrMissingToken
		}

	case "cookie":
		token = string(ctx.Request.Header.Cookie(name))
		if token == "" {
			return "", ErrMissingToken
		}

	default:
		err := errors.New("unsupported token source")
		log.Errorf("err:%v", err)
		return "", err
	}

	if token == "" {
		return "", ErrMissingToken
	}

	return token, nil
}

// ValidateJWT validates JWT token and stores claims in context
func ValidateJWT(ctx *fasthttp.RequestCtx, config JWTConfig) error {
	// Skip if skip function returns true
	if config.SkipFunc != nil && config.SkipFunc(ctx) {
		return nil
	}

	// Extract token
	tokenString, err := ExtractToken(ctx, config)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// Parse and validate token
	token, err := ParseToken(config, tokenString)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// Store claims in context
	if config.ContextKey != "" {
		ctx.SetUserValue(config.ContextKey, token.Claims)
	}

	// Call success handler if provided
	if config.SuccessHandler != nil {
		config.SuccessHandler(ctx, token)
	}

	return nil
}

// GetClaims retrieves JWT claims from context
func GetClaims(ctx *fasthttp.RequestCtx, contextKey string) (jwt.Claims, bool) {
	claims := ctx.UserValue(contextKey)
	if claims == nil {
		return nil, false
	}

	jwtClaims, ok := claims.(jwt.Claims)
	return jwtClaims, ok
}

// GetMapClaims retrieves JWT map claims from context
func GetMapClaims(ctx *fasthttp.RequestCtx, contextKey string) (jwt.MapClaims, bool) {
	claims := ctx.UserValue(contextKey)
	if claims == nil {
		return nil, false
	}

	mapClaims, ok := claims.(jwt.MapClaims)
	return mapClaims, ok
}

// GetCustomClaims retrieves custom JWT claims from context
func GetCustomClaims(ctx *fasthttp.RequestCtx, contextKey string) (*JWTClaims, bool) {
	claims := ctx.UserValue(contextKey)
	if claims == nil {
		return nil, false
	}

	customClaims, ok := claims.(*JWTClaims)
	return customClaims, ok
}

// NewJWTClaims creates new JWT claims with standard fields
func NewJWTClaims(userID, username string, expiration time.Duration) *JWTClaims {
	now := time.Now()
	return &JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
}

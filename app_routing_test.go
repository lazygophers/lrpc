package lrpc

import (
	"fmt"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestAppNewRoutingBasic(t *testing.T) {
	app := NewApp()

	// Register routes
	app.GET("/", func(ctx *Ctx) error {
		ctx.SendString("root")
		return nil
	})

	app.GET("/users", func(ctx *Ctx) error {
		ctx.SendString("users list")
		return nil
	})

	app.GET("/users/{id}", func(ctx *Ctx) error {
		id := ctx.Param("id")
		ctx.SendString("user: " + id)
		return nil
	})

	app.GET("/users/{id:int}", func(ctx *Ctx) error {
		id := ctx.Param("id")
		ctx.SendString("user int: " + id)
		return nil
	})

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/", 200, "root"},
		{"/users", 200, "users list"},
		{"/users/123", 200, "user int: 123"},
		{"/users/abc", 200, "user: abc"},
		{"/notfound", 404, ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI(tt.path)
			ctx.Request.Header.SetMethod("GET")

			app.ServeHTTP(ctx)

			if ctx.Response.StatusCode() != tt.wantStatus {
				t.Errorf("Status: want %d, got %d", tt.wantStatus, ctx.Response.StatusCode())
			}

			if tt.wantBody != "" {
				body := string(ctx.Response.Body())
				if body != tt.wantBody {
					t.Errorf("Body: want %q, got %q", tt.wantBody, body)
				}
			}
		})
	}
}

func TestAppNewRoutingMiddleware(t *testing.T) {
	app := NewApp()

	var executionOrder []string

	// Global middleware
	app.Use(func(ctx *Ctx) error {
		executionOrder = append(executionOrder, "global:before")
		err := ctx.Next()
		executionOrder = append(executionOrder, "global:after")
		return err
	})

	// Route with middleware
	app.GET("/test",
		func(ctx *Ctx) error {
			executionOrder = append(executionOrder, "m1:before")
			err := ctx.Next()
			executionOrder = append(executionOrder, "m1:after")
			return err
		},
		func(ctx *Ctx) error {
			executionOrder = append(executionOrder, "m2:before")
			err := ctx.Next()
			executionOrder = append(executionOrder, "m2:after")
			return err
		},
		func(ctx *Ctx) error {
			executionOrder = append(executionOrder, "handler")
			ctx.SendString("ok")
			return nil
		},
	)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/test")
	ctx.Request.Header.SetMethod("GET")

	executionOrder = []string{}
	app.ServeHTTP(ctx)

	expected := []string{
		"global:before",
		"m1:before",
		"m2:before",
		"handler",
		"m2:after",
		"m1:after",
		"global:after",
	}

	if len(executionOrder) != len(expected) {
		t.Fatalf("Execution order length: want %d, got %d", len(expected), len(executionOrder))
	}

	for i, want := range expected {
		if executionOrder[i] != want {
			t.Errorf("Step %d: want %q, got %q", i, want, executionOrder[i])
		}
	}
}

func TestAppNewRoutingMiddlewareError(t *testing.T) {
	app := NewApp()

	var executionOrder []string

	app.GET("/test",
		func(ctx *Ctx) error {
			executionOrder = append(executionOrder, "m1:before")
			err := ctx.Next()
			executionOrder = append(executionOrder, "m1:after")
			return err
		},
		func(ctx *Ctx) error {
			executionOrder = append(executionOrder, "m2:before")
			// Return error, should stop chain
			return fmt.Errorf("middleware error")
		},
		func(ctx *Ctx) error {
			executionOrder = append(executionOrder, "m3:before")
			err := ctx.Next()
			executionOrder = append(executionOrder, "m3:after")
			return err
		},
		func(ctx *Ctx) error {
			executionOrder = append(executionOrder, "handler")
			return nil
		},
	)

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/test")
	ctx.Request.Header.SetMethod("GET")

	executionOrder = []string{}
	app.ServeHTTP(ctx)

	// m2 returns error, so m3 and handler should not execute
	// But m1:after should still execute
	expected := []string{
		"m1:before",
		"m2:before",
		"m1:after",
	}

	if len(executionOrder) != len(expected) {
		t.Fatalf("Execution order length: want %d, got %d\n%v", len(expected), len(executionOrder), executionOrder)
	}

	for i, want := range expected {
		if executionOrder[i] != want {
			t.Errorf("Step %d: want %q, got %q", i, want, executionOrder[i])
		}
	}
}

func TestAppNewRoutingGroup(t *testing.T) {
	app := NewApp()

	var executionOrder []string

	// API group with middleware
	api := app.Group("/api", func(ctx *Ctx) error {
		executionOrder = append(executionOrder, "api:middleware")
		return ctx.Next()
	})

	// V1 sub-group
	v1 := api.Group("/v1", func(ctx *Ctx) error {
		executionOrder = append(executionOrder, "v1:middleware")
		return ctx.Next()
	})

	v1.GET("/users", func(ctx *Ctx) error {
		executionOrder = append(executionOrder, "handler")
		ctx.SendString("v1 users")
		return nil
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/api/v1/users")
	ctx.Request.Header.SetMethod("GET")

	executionOrder = []string{}
	app.ServeHTTP(ctx)

	expected := []string{
		"api:middleware",
		"v1:middleware",
		"handler",
	}

	if len(executionOrder) != len(expected) {
		t.Fatalf("Execution order length: want %d, got %d", len(expected), len(executionOrder))
	}

	for i, want := range expected {
		if executionOrder[i] != want {
			t.Errorf("Step %d: want %q, got %q", i, want, executionOrder[i])
		}
	}

	body := string(ctx.Response.Body())
	if body != "v1 users" {
		t.Errorf("Body: want %q, got %q", "v1 users", body)
	}
}

func TestAppNewRoutingParameterTypes(t *testing.T) {
	app := NewApp()

	app.GET("/int/{id:int}", func(ctx *Ctx) error {
		id := ctx.Param("id")
		ctx.SendString("int:" + id)
		return nil
	})

	app.GET("/uuid/{id:uuid}", func(ctx *Ctx) error {
		id := ctx.Param("id")
		ctx.SendString("uuid:" + id)
		return nil
	})

	app.GET("/regex/{code:^[A-Z]{3}$}", func(ctx *Ctx) error {
		code := ctx.Param("code")
		ctx.SendString("code:" + code)
		return nil
	})

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{"/int/123", 200, "int:123"},
		{"/int/abc", 404, ""}, // Should not match int constraint
		{"/uuid/550e8400-e29b-41d4-a716-446655440000", 200, "uuid:550e8400-e29b-41d4-a716-446655440000"},
		{"/uuid/invalid", 404, ""},
		{"/regex/ABC", 200, "code:ABC"},
		{"/regex/123", 404, ""},
		{"/regex/ABCD", 404, ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI(tt.path)
			ctx.Request.Header.SetMethod("GET")

			app.ServeHTTP(ctx)

			if ctx.Response.StatusCode() != tt.wantStatus {
				t.Errorf("Status: want %d, got %d", tt.wantStatus, ctx.Response.StatusCode())
			}

			if tt.wantBody != "" {
				body := string(ctx.Response.Body())
				if body != tt.wantBody {
					t.Errorf("Body: want %q, got %q", tt.wantBody, body)
				}
			}
		})
	}
}

func TestAppNewRoutingPriority(t *testing.T) {
	app := NewApp()

	// Register routes in different order to test priority
	app.GET("/users/*", func(ctx *Ctx) error {
		ctx.SendString("wildcard")
		return nil
	})

	app.GET("/users/{id}", func(ctx *Ctx) error {
		ctx.SendString("param")
		return nil
	})

	app.GET("/users/{id:int}", func(ctx *Ctx) error {
		ctx.SendString("typed")
		return nil
	})

	app.GET("/users/list", func(ctx *Ctx) error {
		ctx.SendString("static")
		return nil
	})

	tests := []struct {
		path     string
		wantBody string
	}{
		{"/users/list", "static"},   // Static wins
		{"/users/123", "typed"},     // Typed param wins over generic param
		{"/users/abc", "param"},     // Generic param wins over wildcard
		{"/users/x/y", "wildcard"},  // Only wildcard matches
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI(tt.path)
			ctx.Request.Header.SetMethod("GET")

			app.ServeHTTP(ctx)

			body := string(ctx.Response.Body())
			if body != tt.wantBody {
				t.Errorf("Body: want %q, got %q", tt.wantBody, body)
			}
		})
	}
}

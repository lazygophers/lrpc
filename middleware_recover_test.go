package lrpc

import (
	"fmt"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestRecover(t *testing.T) {
	app := NewApp()

	// Add recover middleware
	app.Use(Recover())

	// Route that panics
	app.GET("/panic", func(ctx *Ctx) error {
		panic("test panic")
	})

	// Route that doesn't panic
	app.GET("/ok", func(ctx *Ctx) error {
		ctx.SendString("ok")
		return nil
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantPanic  bool
	}{
		{"panic route", "/panic", 500, true},
		{"normal route", "/ok", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &fasthttp.RequestCtx{}
			ctx.Request.SetRequestURI(tt.path)
			ctx.Request.Header.SetMethod("GET")

			app.ServeHTTP(ctx)

			if ctx.Response.StatusCode() != tt.wantStatus {
				t.Errorf("Status: want %d, got %d", tt.wantStatus, ctx.Response.StatusCode())
			}
		})
	}
}

func TestRecoverWithCustomHandler(t *testing.T) {
	app := NewApp()

	var customCalled bool
	customHandler := func(ctx *Ctx, err interface{}) {
		customCalled = true
		ctx.SendString(fmt.Sprintf("Custom: %v", err))
	}

	app.Use(Recover(RecoverConfig{
		EnableStackTrace: false,
		ErrorHandler:     customHandler,
	}))

	app.GET("/panic", func(ctx *Ctx) error {
		panic("test panic")
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/panic")
	ctx.Request.Header.SetMethod("GET")

	app.ServeHTTP(ctx)

	if !customCalled {
		t.Error("Custom error handler was not called")
	}

	body := string(ctx.Response.Body())
	expected := "Custom: test panic"
	if body != expected {
		t.Errorf("Body: want %q, got %q", expected, body)
	}
}

func TestRecoverMiddlewareChain(t *testing.T) {
	app := NewApp()

	var executionOrder []string

	app.Use(Recover())
	app.Use(func(ctx *Ctx) error {
		executionOrder = append(executionOrder, "middleware:before")
		err := ctx.Next()
		executionOrder = append(executionOrder, "middleware:after")
		return err
	})

	app.GET("/panic", func(ctx *Ctx) error {
		executionOrder = append(executionOrder, "handler:panic")
		panic("test panic")
	})

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("/panic")
	ctx.Request.Header.SetMethod("GET")

	executionOrder = []string{}
	app.ServeHTTP(ctx)

	// After panic is recovered, execution stops
	expected := []string{
		"middleware:before",
		"handler:panic",
		// Note: "middleware:after" won't execute because panic was recovered
		// and the execution chain was interrupted
	}

	if len(executionOrder) != len(expected) {
		t.Fatalf("Execution order length: want %d, got %d\n%v",
			len(expected), len(executionOrder), executionOrder)
	}

	for i, want := range expected {
		if executionOrder[i] != want {
			t.Errorf("Step %d: want %q, got %q", i, want, executionOrder[i])
		}
	}
}

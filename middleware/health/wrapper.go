package health

import (
	"time"

	"github.com/lazygophers/lrpc"
	"github.com/valyala/fasthttp"
)

// AddHealthEndpoints adds standard health check endpoints to the app
func AddHealthEndpoints(app interface {
	GET(path string, handler func(*lrpc.Ctx) error) error
}, prefix string, checker *Checker) error {
	if prefix == "" {
		prefix = "/"
	}

	// Liveness probe
	err := app.GET(prefix+"health", func(ctx *lrpc.Ctx) error {
		ctx.Context().SetStatusCode(fasthttp.StatusOK)
		return ctx.SendJson(map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})
	if err != nil {
		return err
	}

	// Readiness probe
	err = app.GET(prefix+"ready", func(ctx *lrpc.Ctx) error {
		if checker != nil && !checker.IsReady() {
			ctx.Context().SetStatusCode(fasthttp.StatusServiceUnavailable)
			return ctx.SendJson(map[string]interface{}{
				"status": "not ready",
				"time":   time.Now().Unix(),
			})
		}

		ctx.Context().SetStatusCode(fasthttp.StatusOK)
		return ctx.SendJson(map[string]interface{}{
			"status": "ready",
			"time":   time.Now().Unix(),
		})
	})
	if err != nil {
		return err
	}

	// Detailed health check endpoint
	if checker != nil {
		err = app.GET(prefix+"healthz", func(ctx *lrpc.Ctx) error {
			results := checker.RunChecks()

			status := results["status"]
			if status == StatusUnhealthy {
				ctx.Context().SetStatusCode(fasthttp.StatusServiceUnavailable)
			} else {
				ctx.Context().SetStatusCode(fasthttp.StatusOK)
			}

			return ctx.SendJson(results)
		})
		if err != nil {
			return err
		}
	}

	return nil
}

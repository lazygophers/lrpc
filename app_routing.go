package lrpc

import (
	"net/http"

	"github.com/valyala/fasthttp"
)

// Group represents a route group
type Group struct {
	app        *App
	prefix     string
	middleware []HandlerFunc
}

// Use adds global middleware to all routers
func (app *App) Use(handlers ...HandlerFunc) {
	// Store global middleware
	app.globalMiddleware = append(app.globalMiddleware, handlers...)

	// Also add to existing routers
	for _, router := range app.routers {
		router.middleware = append(router.middleware, handlers...)
	}
}

// Group creates a route group with prefix
func (app *App) Group(prefix string, handlers ...HandlerFunc) *Group {
	return &Group{
		app:        app,
		prefix:     prefix,
		middleware: handlers,
	}
}

// GET registers a GET route
func (app *App) GET(path string, handlers ...HandlerFunc) error {
	return app.addRoute(http.MethodGet, path, handlers)
}

// POST registers a POST route
func (app *App) POST(path string, handlers ...HandlerFunc) error {
	return app.addRoute(http.MethodPost, path, handlers)
}

// PUT registers a PUT route
func (app *App) PUT(path string, handlers ...HandlerFunc) error {
	return app.addRoute(http.MethodPut, path, handlers)
}

// DELETE registers a DELETE route
func (app *App) DELETE(path string, handlers ...HandlerFunc) error {
	return app.addRoute(http.MethodDelete, path, handlers)
}

// PATCH registers a PATCH route
func (app *App) PATCH(path string, handlers ...HandlerFunc) error {
	return app.addRoute(http.MethodPatch, path, handlers)
}

// HEAD registers a HEAD route
func (app *App) HEAD(path string, handlers ...HandlerFunc) error {
	return app.addRoute(http.MethodHead, path, handlers)
}

// OPTIONS registers an OPTIONS route
func (app *App) OPTIONS(path string, handlers ...HandlerFunc) error {
	return app.addRoute(http.MethodOptions, path, handlers)
}

// addRoute adds a route to the routing system
func (app *App) addRoute(method, path string, handlers []HandlerFunc) error {
	// Initialize routers if not done yet
	if app.routers == nil {
		app.routers = make(map[string]*Router)
	}

	router, ok := app.routers[method]
	if !ok {
		router = NewRouter()
		// Add global middleware to new router
		router.middleware = append(router.middleware, app.globalMiddleware...)
		app.routers[method] = router
	}

	return router.addRoute(path, handlers)
}

// ServeHTTP handles HTTP requests using the routing system
func (app *App) ServeHTTP(ctx *fasthttp.RequestCtx) {
	method := string(ctx.Method())
	path := string(ctx.Path())

	// Get router for method
	router, ok := app.routers[method]
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		return
	}

	// Find matching route
	result, err := router.findRoute(path)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	if result == nil {
		// No route found
		if router.notFound != nil {
			appCtx := app.AcquireCtx(ctx)
			defer app.ReleaseCtx(appCtx)
			_ = router.notFound(appCtx)
		} else {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		}
		return
	}

	// Create context
	appCtx := app.AcquireCtx(ctx)
	defer app.ReleaseCtx(appCtx)

	// Set params
	appCtx.params = result.Params

	// Build handler chain: global middleware + route handlers
	appCtx.handlers = make([]HandlerFunc, 0, len(router.middleware)+len(result.Node.handlers))
	appCtx.handlers = append(appCtx.handlers, router.middleware...)
	appCtx.handlers = append(appCtx.handlers, result.Node.handlers...)

	// Execute chain
	if err := appCtx.executeChain(); err != nil {
		// Error handling
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(err.Error())
	}
}

// Use adds middleware to the group
func (g *Group) Use(handlers ...HandlerFunc) *Group {
	g.middleware = append(g.middleware, handlers...)
	return g
}

// Group creates a sub-group
func (g *Group) Group(prefix string, handlers ...HandlerFunc) *Group {
	return &Group{
		app:        g.app,
		prefix:     g.prefix + prefix,
		middleware: append(g.middleware, handlers...),
	}
}

// GET registers a GET route in the group
func (g *Group) GET(path string, handlers ...HandlerFunc) error {
	fullPath := g.prefix + path
	allHandlers := append(g.middleware, handlers...)
	return g.app.addRoute(http.MethodGet, fullPath, allHandlers)
}

// POST registers a POST route in the group
func (g *Group) POST(path string, handlers ...HandlerFunc) error {
	fullPath := g.prefix + path
	allHandlers := append(g.middleware, handlers...)
	return g.app.addRoute(http.MethodPost, fullPath, allHandlers)
}

// PUT registers a PUT route in the group
func (g *Group) PUT(path string, handlers ...HandlerFunc) error {
	fullPath := g.prefix + path
	allHandlers := append(g.middleware, handlers...)
	return g.app.addRoute(http.MethodPut, fullPath, allHandlers)
}

// DELETE registers a DELETE route in the group
func (g *Group) DELETE(path string, handlers ...HandlerFunc) error {
	fullPath := g.prefix + path
	allHandlers := append(g.middleware, handlers...)
	return g.app.addRoute(http.MethodDelete, fullPath, allHandlers)
}

// PATCH registers a PATCH route in the group
func (g *Group) PATCH(path string, handlers ...HandlerFunc) error {
	fullPath := g.prefix + path
	allHandlers := append(g.middleware, handlers...)
	return g.app.addRoute(http.MethodPatch, fullPath, allHandlers)
}

// HEAD registers a HEAD route in the group
func (g *Group) HEAD(path string, handlers ...HandlerFunc) error {
	fullPath := g.prefix + path
	allHandlers := append(g.middleware, handlers...)
	return g.app.addRoute(http.MethodHead, fullPath, allHandlers)
}

// OPTIONS registers an OPTIONS route in the group
func (g *Group) OPTIONS(path string, handlers ...HandlerFunc) error {
	fullPath := g.prefix + path
	allHandlers := append(g.middleware, handlers...)
	return g.app.addRoute(http.MethodOptions, fullPath, allHandlers)
}

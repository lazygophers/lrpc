package plugin

import (
	"github.com/lazygophers/log"
	"github.com/valyala/fasthttp"
)

// ExampleMiddlewarePlugin demonstrates a middleware plugin implementation
type ExampleMiddlewarePlugin struct {
	*BasePlugin
	prefix string
}

// NewExampleMiddlewarePlugin creates a new example middleware plugin
func NewExampleMiddlewarePlugin() *ExampleMiddlewarePlugin {
	return &ExampleMiddlewarePlugin{
		BasePlugin: NewBasePlugin("example-middleware", "1.0.0"),
		prefix:     "[ExamplePlugin]",
	}
}

// Init initializes the plugin with config
func (p *ExampleMiddlewarePlugin) Init(config interface{}) error {
	err := p.BasePlugin.Init(config)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// Extract prefix from config if provided
	if cfg, ok := config.(map[string]interface{}); ok {
		if prefix, ok := cfg["prefix"].(string); ok {
			p.prefix = prefix
		}
	}

	return nil
}

// Handler returns the middleware handler
func (p *ExampleMiddlewarePlugin) Handler() func(ctx *fasthttp.RequestCtx, next func()) {
	return func(ctx *fasthttp.RequestCtx, next func()) {
		// Before request
		log.Infof("%s Before request: %s %s", p.prefix, string(ctx.Method()), string(ctx.Path()))

		// Call next middleware
		next()

		// After request
		log.Infof("%s After request: status=%d", p.prefix, ctx.Response.StatusCode())
	}
}

// ExampleServicePlugin demonstrates a service plugin implementation
type ExampleServicePlugin struct {
	*BasePlugin
	stopChan chan struct{}
}

// NewExampleServicePlugin creates a new example service plugin
func NewExampleServicePlugin() *ExampleServicePlugin {
	return &ExampleServicePlugin{
		BasePlugin: NewBasePlugin("example-service", "1.0.0"),
		stopChan:   make(chan struct{}),
	}
}

// Run runs the service in background
func (p *ExampleServicePlugin) Run() error {
	log.Infof("Service plugin %s is running", p.Name())

	// Simulate background work
	for {
		select {
		case <-p.stopChan:
			log.Infof("Service plugin %s stopped", p.Name())
			return nil
		default:
			// Do background work here
			// time.Sleep(10 * time.Second)
		}
	}
}

// Stop stops the service
func (p *ExampleServicePlugin) Stop() error {
	close(p.stopChan)
	return p.BasePlugin.Stop()
}

// ExampleHookPlugin demonstrates a hook plugin implementation
type ExampleHookPlugin struct {
	*BasePlugin
}

// NewExampleHookPlugin creates a new example hook plugin
func NewExampleHookPlugin() *ExampleHookPlugin {
	return &ExampleHookPlugin{
		BasePlugin: NewBasePlugin("example-hook", "1.0.0"),
	}
}

// OnRequest is called before request is handled
func (p *ExampleHookPlugin) OnRequest(ctx *fasthttp.RequestCtx) {
	log.Infof("Hook: OnRequest %s %s", string(ctx.Method()), string(ctx.Path()))
}

// OnResponse is called after response is sent
func (p *ExampleHookPlugin) OnResponse(ctx *fasthttp.RequestCtx) {
	log.Infof("Hook: OnResponse status=%d", ctx.Response.StatusCode())
}

// OnError is called when an error occurs
func (p *ExampleHookPlugin) OnError(ctx *fasthttp.RequestCtx, err error) {
	log.Errorf("Hook: OnError %v", err)
}

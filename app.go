package lrpc

import (
	"sync"

	"github.com/valyala/fasthttp"
)

type App struct {
	c *Config

	server *fasthttp.Server

	routes map[string]*SearchTree[HandlerFunc]

	// New routing system
	routers            map[string]*Router
	globalMiddleware   []HandlerFunc // Global middleware for new routing

	ctxPool sync.Pool

	hook *Hooks
}

func NewApp(c ...*Config) *App {
	p := &App{
		routes:  make(map[string]*SearchTree[HandlerFunc]),
		routers: make(map[string]*Router),
		ctxPool: sync.Pool{
			New: func() any {
				return newCtx()
			},
		},
		hook: new(Hooks),
	}

	if len(c) > 0 {
		p.c = c[0]
	}

	p.initConfig()
	p.initServer()

	return p
}

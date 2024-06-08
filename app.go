package lrpc

import (
	"github.com/valyala/fasthttp"
	"sync"
)

type App struct {
	c *Config

	server *fasthttp.Server

	routes map[string]*SearchTree[HandlerFunc]

	ctxPool sync.Pool
}

func NewApp(c ...*Config) *App {
	p := &App{
		routes: make(map[string]*SearchTree[HandlerFunc]),
		ctxPool: sync.Pool{
			New: func() any {
				return newCtx()
			},
		},
	}

	if len(c) > 0 {
		p.c = c[0]
	}

	p.initConfig()
	p.initServer()

	return p
}

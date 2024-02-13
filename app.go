package lrpc

import (
	"github.com/lazygophers/log"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"sync"
)

type Config struct {
}

func mergeConfig(c *Config) *Config {

	return c
}

type App struct {
	lock sync.RWMutex

	ctxPool sync.Pool
	ctxMap  sync.Map

	c *Config

	server *fasthttp.Server
	route  *routing.Router

	OnError func(ctx *Ctx, err error)
}

func (p *App) defaultOnError(ctx *Ctx, err error) {
	log.Errorf("err:%v", err)
}

func (p *App) handleOnError(ctx *Ctx, err error) {
	p.OnError(ctx, err)
}

func (p *App) Listen(addr string) (err error) {
	err = p.server.ListenAndServe(addr)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *App) Shutdown() (err error) {
	err = p.server.Shutdown()
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	return nil
}

func (p *App) init() {
	p.ctxPool = sync.Pool{
		New: func() interface{} {
			return p.newCtx()
		},
	}

	p.server.Handler = p.route.HandleRequest

	p.OnError = p.defaultOnError
}

func New(cs ...*Config) *App {
	p := &App{
		server: &fasthttp.Server{
			Logger: &disableLogger{},
		},
		route: routing.New(),
	}

	if len(cs) > 0 {
		p.c = mergeConfig(cs[0])
	} else {
		p.c = mergeConfig(&Config{})
	}

	p.init()

	return p
}

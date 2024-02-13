package lrpc

import (
	"github.com/lazygophers/utils/json"
	"github.com/valyala/fasthttp"
)

type Ctx struct {
	app *App

	context *fasthttp.RequestCtx

	path   string
	method string
}

func (p *Ctx) acquire(fastCtx *fasthttp.RequestCtx) {
	p.context = fastCtx
}

func (p *Ctx) release() {
	p.context = nil
}

func (p *Ctx) Method() string {
	return p.method
}

func (p *Ctx) Path() string {
	return p.path
}

func (p *Ctx) BodyParser(req interface{}) (err error) {
	return json.Unmarshal(p.context.PostBody(), req)
}

func (p *Ctx) ContentType() string {
	return string(p.context.Request.Header.ContentType())
}

func (p *Ctx) SendString(data string) error {
	p.context.SetBodyString(data)
	return nil
}

// -----------App-----------

func (p *App) newCtx() *Ctx {
	return &Ctx{
		app: p,
	}
}

func (p *App) AcquireCtx(fastCtx *fasthttp.RequestCtx) *Ctx {
	ctx := p.ctxPool.Get().(*Ctx)
	ctx.acquire(fastCtx)

	return ctx
}

func (p *App) ReleaseCtx(ctx *Ctx) {
	ctx.release()
	p.ctxPool.Put(ctx)
}

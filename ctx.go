package lrpc

import (
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils"
	"github.com/lazygophers/utils/json"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
	"strings"
)

type Ctx struct {
	ctx *fasthttp.RequestCtx

	params map[string]string
}

func newCtx() *Ctx {
	return &Ctx{
		params: make(map[string]string),
	}
}

func (p *Ctx) Reset() {
	p.ctx = nil
	if len(p.params) > 0 {
		p.params = make(map[string]string)
	}
}

func (p *Ctx) Context() *fasthttp.RequestCtx {
	return p.ctx
}

func (p *Ctx) setParam(params map[string]string) {
	p.params = params
}

func (p *Ctx) SetLocal(key string, value any) {
	p.ctx.SetUserValue(key, value)
}

func (p *Ctx) GetLocal(key string) any {
	return p.ctx.UserValue(key)
}

func (p *Ctx) Method() string {
	return string(p.ctx.Method())
}

func (p *Ctx) Path() string {
	return string(p.ctx.Path())
}

func (p *Ctx) Header(key string) string {
	return string(p.ctx.Request.Header.Peek(key))
}

func (p *Ctx) Query(key string) string {
	return string(p.ctx.QueryArgs().Peek(key))
}

func (p *Ctx) Parame(key string) string {
	return p.params[key]
}

func (p *Ctx) Body() []byte {
	return p.ctx.Request.Body()
}

func (p *Ctx) BodyEmpty() bool {
	return len(p.ctx.Request.Body()) == 0
}

func (p *Ctx) Send(body []byte) {
	p.ctx.SetBody(body)
}

func (p *Ctx) SendString(s string) {
	p.ctx.SetBodyString(s)
}

func (p *Ctx) SendStatus(status int) {
	p.ctx.SetStatusCode(status)
}

func (p *Ctx) SendJson(o any) error {
	buffer, err := json.Marshal(o)
	if err != nil {
		return err
	}

	p.ctx.SetBody(buffer)
	return nil
}

func (p *Ctx) IsBodyStream() bool {
	return p.ctx.IsBodyStream()
}

func (p *Ctx) BodyParser(o any) (err error) {
	// 如果 bdoy 不为空，就解析 body
	body := p.ctx.Request.Body()
	if len(body) == 0 {
		// TODO: 先统一按照 json 的方式处理
		return nil
	}

	contentType := p.Header("Content-Type")
	if strings.Contains(contentType, "application/protobuf") {
		if v, ok := o.(proto.Message); ok {
			err = proto.Unmarshal(body, v)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			err = utils.Validate(v)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			return nil
		}
	}

	err = json.Unmarshal(body, o)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	err = utils.Validate(o)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// TODO: 处理 get 的 args 类型请求

	return nil
}

func (p *App) AcquireCtx(ctx *fasthttp.RequestCtx) *Ctx {
	c := p.ctxPool.Get().(*Ctx)

	if ctx == nil {
		ctx = &fasthttp.RequestCtx{}
	}

	c.ctx = ctx

	return c
}

func (p *App) ReleaseCtx(ctx *Ctx) {
	ctx.Reset()
	p.ctxPool.Put(ctx)
}

package lrpc

import (
	"strings"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils"
	"github.com/lazygophers/utils/json"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

type Ctx struct {
	ctx *fasthttp.RequestCtx

	tranceId string

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

func (p *Ctx) SetHeader(key string, value string) {
	p.ctx.Response.Header.Set(key, value)
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
	contentType := p.Header(HeaderContentType)
	if strings.Contains(contentType, MIMEApplicationProtobuf) {
		if v, ok := o.(proto.Message); ok {
			p.SetHeader(HeaderContentType, MIMEApplicationProtobuf)
			buffer, err := proto.Marshal(v)
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			p.ctx.SetBody(buffer)
			return nil
		}
	}

	p.SetHeader(HeaderContentType, MIMEApplicationJSON)
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

	contentType := p.Header(HeaderContentType)
	if strings.Contains(contentType, MIMEApplicationProtobuf) {
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

func (p *Ctx) TranceId() string {
	return p.tranceId
}

func (p *Ctx) SetTranceId(tranceId ...string) {
	if len(tranceId) > 0 {
		p.tranceId = tranceId[0]
	} else {
		p.tranceId = log.GenTraceId()
	}
}

func (p *Ctx) init() {
	if p.Header(HeaderTrance) != "" {
		p.tranceId = p.Header(HeaderTrance)
	}

	if p.tranceId == "" {
		p.tranceId = log.GetTrace()
	}
}

func (p *App) AcquireCtx(ctx *fasthttp.RequestCtx) *Ctx {
	c := p.ctxPool.Get().(*Ctx)

	if ctx == nil {
		ctx = &fasthttp.RequestCtx{}
	}

	c.ctx = ctx

	c.init()

	return c
}

func (p *App) ReleaseCtx(ctx *Ctx) {
	ctx.Reset()
	p.ctxPool.Put(ctx)
}

func NewCtxTools() *Ctx {
	p := &Ctx{
		ctx: &fasthttp.RequestCtx{},
	}

	p.init()

	return p
}

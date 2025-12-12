package lrpc

import (
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/json"
	"github.com/lazygophers/utils/validator"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

type Ctx struct {
	ctx *fasthttp.RequestCtx

	traceID string

	params map[string]string

	// Middleware chain support
	handlers []HandlerFunc
	index    int
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
	p.handlers = nil
	p.index = -1
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

func (p *Ctx) Param(key string) string {
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

			err = validator.Struct(v)
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

	err = validator.Struct(o)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	// TODO: 处理 get 的 args 类型请求

	return nil
}

func (p *Ctx) TraceID() string {
	return p.traceID
}

func (p *Ctx) SetTraceID(traceID ...string) {
	if len(traceID) > 0 {
		p.traceID = traceID[0]
	} else {
		p.traceID = log.GenTraceId()
	}
}

func (p *Ctx) init() {
	if p.Header(HeaderTrace) != "" {
		p.traceID = p.Header(HeaderTrace)
	}

	if p.traceID == "" {
		p.traceID = log.GetTrace()
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

// Next executes the next handler in the middleware chain
func (p *Ctx) Next() error {
	p.index++
	if p.index < len(p.handlers) {
		return p.handlers[p.index](p)
	}
	return nil
}

// SetParam sets a route parameter (for compatibility with new routing)
func (p *Ctx) SetParam(key, value string) {
	if p.params == nil {
		p.params = make(map[string]string)
	}
	p.params[key] = value
}

// AllParams returns all route parameters
func (p *Ctx) AllParams() map[string]string {
	if p.params == nil {
		return make(map[string]string)
	}
	// Return a copy to prevent external modification
	result := make(map[string]string, len(p.params))
	for k, v := range p.params {
		result[k] = v
	}
	return result
}

// executeChain executes the handler chain starting from index -1
func (p *Ctx) executeChain() error {
	if len(p.handlers) == 0 {
		return nil
	}
	p.index = -1
	return p.Next()
}

// Cookie methods

// Cookie gets a cookie value by name
func (p *Ctx) Cookie(name string) string {
	return string(p.ctx.Request.Header.Cookie(name))
}

// SetCookie sets a cookie
func (p *Ctx) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	cookie := &fasthttp.Cookie{}
	cookie.SetKey(name)
	cookie.SetValue(value)
	cookie.SetMaxAge(maxAge)
	if path != "" {
		cookie.SetPath(path)
	}
	if domain != "" {
		cookie.SetDomain(domain)
	}
	cookie.SetSecure(secure)
	cookie.SetHTTPOnly(httpOnly)
	p.ctx.Response.Header.SetCookie(cookie)
}

// ClearCookie clears a cookie
func (p *Ctx) ClearCookie(name string, path string) {
	p.SetCookie(name, "", -1, path, "", false, false)
}

// Query methods

// QueryInt gets a query parameter as int
func (p *Ctx) QueryInt(key string, defaultValue ...int) int {
	val := p.Query(key)
	if val == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	var result int
	_, err := fmt.Sscanf(val, "%d", &result)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return result
}

// QueryBool gets a query parameter as bool
func (p *Ctx) QueryBool(key string, defaultValue ...bool) bool {
	val := p.Query(key)
	if val == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}

	val = strings.ToLower(val)
	if val == "true" || val == "1" || val == "yes" || val == "on" {
		return true
	}
	if val == "false" || val == "0" || val == "no" || val == "off" {
		return false
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return false
}

// AllQueries returns all query parameters
func (p *Ctx) AllQueries() map[string]string {
	result := make(map[string]string)
	p.ctx.QueryArgs().VisitAll(func(key, value []byte) {
		result[string(key)] = string(value)
	})
	return result
}

// File upload methods

// FormFile gets a file from multipart form
func (p *Ctx) FormFile(key string) (*multipart.FileHeader, error) {
	return p.ctx.FormFile(key)
}

// SaveFile saves an uploaded file to a path
func (p *Ctx) SaveFile(fileHeader *multipart.FileHeader, path string) error {
	return fasthttp.SaveMultipartFile(fileHeader, path)
}

// MultipartForm gets the multipart form
func (p *Ctx) MultipartForm() (*multipart.Form, error) {
	return p.ctx.MultipartForm()
}

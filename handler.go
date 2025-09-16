package lrpc

import (
	"net/http"
	"reflect"

	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/utils/validator"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type BaseResponse struct {
	Code    int32  `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Data    any    `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	Hint    string `protobuf:"bytes,4,opt,name=hint,proto3" json:"hint,omitempty"`
}

type HandlerFunc func(ctx *Ctx) error

func (p *App) onError(ctx *Ctx, err error) {
	if err == nil {
		return
	}

	if p.c.OnError == nil {
		return
	}

	p.c.OnError(ctx, err)
}

func (p *App) afterHandlerWithRef(ctx *Ctx, data reflect.Value, err error) {
	if err != nil {
		p.onError(ctx, err)
		return
	}

	if p.c.AfterHandlerFuncWithRef != nil {
		p.c.AfterHandlerFuncWithRef(ctx, data, err)
		if err != nil {
			p.onError(ctx, err)
			return
		}
	}

	if !ctx.BodyEmpty() || ctx.IsBodyStream() {
		return
	}

	di := data.Interface()
	log.Infof("%s Response %s", ctx.Path(), di)
	if x, ok := di.(proto.Message); ok {
		a, err := anypb.New(x)
		if err != nil {
			log.Errorf("err:%v", err)
			p.onError(ctx, err)
			return
		}

		err = ctx.SendJson(&core.BaseResponse{
			Data: a,
			Hint: log.GetTrace(),
		})
		if err != nil {
			log.Errorf("err:%v", err)
			p.onError(ctx, err)
			return
		}

		return
	}

	err = ctx.SendJson(&BaseResponse{
		Data: data.Interface(),
		Hint: log.GetTrace(),
	})
	if err != nil {
		log.Errorf("err:%v", err)
		p.onError(ctx, err)
		return
	}

	return
}

func (p *App) ToHandlerFunc(logic any) HandlerFunc {
	lt := reflect.TypeOf(logic)
	lv := reflect.ValueOf(logic)
	if lt.Kind() != reflect.Func {
		panic("parameter is not func")
	}

	// 不管怎么样，第一个参数都是一定要存在的
	x := lt.In(0)
	for x.Kind() == reflect.Ptr {
		x = x.Elem()
	}

	if x.Name() != "Ctx" {
		panic("first in is must *github.com/lazygophers/lrpc.Ctx")
	}

	if x.PkgPath() != "github.com/lazygophers/lrpc" {
		panic("first in is must *github.com/lazygophers/lrpc.Ctx")
	}

	// 两个入参，一个出参，那就是需要解析请求参数的
	if lt.NumIn() == 2 && lt.NumOut() == 1 {
		// 先判断出参是否是error
		x = lt.Out(0)
		for x.Kind() == reflect.Ptr {
			x = x.Elem()
		}
		if x.Name() != "error" {
			panic("out is must error")
		}

		// 处理一下入参
		in := lt.In(1)
		for in.Kind() == reflect.Ptr {
			in = in.Elem()
		}

		if in.Kind() != reflect.Struct {
			panic("2rd in is must struct")
		}

		return func(ctx *Ctx) error {
			req := reflect.New(in)
			err := ctx.BodyParser(req.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				p.afterHandlerWithRef(ctx, req, err)
				return err
			}

			err = validator.Struct(req.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				p.afterHandlerWithRef(ctx, req, err)
				return err
			}

			out := lv.Call([]reflect.Value{reflect.ValueOf(ctx), req})
			if !out[0].IsNil() {
				log.Errorf("err:%v", out[0].Interface())
				p.afterHandlerWithRef(ctx, req, err)
				return out[0].Interface().(error)
			}

			p.afterHandlerWithRef(ctx, req, nil)
			return nil
		}
	}

	// 一个入参，两个出参，那就是需要返回数据的
	if lt.NumIn() == 1 && lt.NumOut() == 2 {

		// 先判断第二个出参是否是error
		x = lt.Out(1)
		for x.Kind() == reflect.Ptr {
			x = x.Elem()
		}
		if x.Name() != "error" {
			panic("out 1 is must error")
		}

		// 处理一下返回
		x = lt.Out(0)
		for x.Kind() == reflect.Ptr {
			x = x.Elem()
		}
		if x.Kind() != reflect.Struct {
			panic("out 0 is must struct")
		}

		return func(ctx *Ctx) error {
			out := lv.Call([]reflect.Value{reflect.ValueOf(ctx)})
			if out[1].IsNil() {
				p.afterHandlerWithRef(ctx, out[0], nil)
				return nil
			}

			p.afterHandlerWithRef(ctx, out[0], out[1].Interface().(error))

			return nil
		}
	}

	// 两个入参，两个出参，那就是需要解析请求参数，返回数据的
	if lt.NumIn() == 2 && lt.NumOut() == 2 {

		// 先判断第二个出参是否是error
		x = lt.Out(1)
		for x.Kind() == reflect.Ptr {
			x = x.Elem()
		}
		if x.Name() != "error" {
			panic("out 1 is must error")
		}

		// 处理一下入参
		in := lt.In(1)
		for in.Kind() == reflect.Ptr {
			in = in.Elem()
		}
		if in.Kind() != reflect.Struct {
			panic("2rd in is must struct")
		}

		// 处理一下返回
		out := lt.Out(0)
		for out.Kind() == reflect.Ptr {
			out = out.Elem()
		}
		if out.Kind() != reflect.Struct {
			panic("out 0 is must struct")
		}

		return func(ctx *Ctx) error {
			req := reflect.New(in)
			err := ctx.BodyParser(req.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				return err
			}

			err = validator.Struct(req.Interface())
			if err != nil {
				return err
			}

			out := lv.Call([]reflect.Value{reflect.ValueOf(ctx), req})
			if out[1].IsNil() {
				p.afterHandlerWithRef(ctx, out[0], nil)
				return nil
			}

			p.afterHandlerWithRef(ctx, out[0], out[1].Interface().(error))
			return nil
		}
	}

	// 一个入参，一个出参，那就不需要解析请求参数，也不需要返回数据的
	if lt.NumIn() == 1 && lt.NumOut() == 1 {
		// 先判断出参是否是error
		x = lt.Out(0)
		for x.Kind() == reflect.Ptr {
			x = x.Elem()
		}
		if x.Name() != "error" {
			panic("out is must error")
		}

		return func(ctx *Ctx) error {
			out := lv.Call([]reflect.Value{reflect.ValueOf(ctx)})
			if !out[0].IsNil() {
				log.Errorf("err:%v", out[0].Interface())
				p.afterHandlerWithRef(ctx, reflect.Value{}, out[0].Interface().(error))
				return nil
			}

			p.afterHandlerWithRef(ctx, reflect.Value{}, nil)
			return nil
		}
	}

	panic("func is not support")
}

func MergeHandler(handlers ...HandlerFunc) HandlerFunc {
	return func(ctx *Ctx) error {
		for _, handler := range handlers {
			err := handler(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func (p *App) handlerExtr(extr map[string]any) HandlerFunc {
	return func(ctx *Ctx) error {
		for k, v := range extr {
			ctx.SetLocal(k, v)
		}
		return nil
	}
}

func (p *App) AddRoute(r *Route, opts ...RouteOption) {
	for _, o := range opts {
		o(r)
	}

	if _, ok := p.routes[r.Method]; !ok {
		p.routes[r.Method] = NewSearchTree[HandlerFunc]()
	}

	for _, logic := range p.hook.onRoute {
		err := logic(r)
		if err != nil {
			log.Fatalf("err:%v", err)
			return
		}
	}

	var handlers []HandlerFunc
	if len(r.Extra) > 0 {
		handlers = append(handlers, p.handlerExtr(r.Extra))
	}

	handlers = append(handlers, r.Before...)
	handlers = append(handlers, r.Handler)
	handlers = append(handlers, r.After...)

	p.routes[r.Method].Add(r.Path, MergeHandler(handlers...))
}

func (p *App) AddRoutes(rs []*Route, opts ...RouteOption) {
	for _, r := range rs {
		p.AddRoute(r, opts...)
	}
}

func (p *App) Handle(method, path string, handler HandlerFunc, opts ...RouteOption) {
	p.AddRoute(&Route{
		Method:  method,
		Path:    path,
		Handler: handler,
	}, opts...)
}

func (p *App) Get(path string, handler HandlerFunc, opts ...RouteOption) {
	p.Handle(http.MethodGet, path, handler, opts...)
}

func (p *App) Head(path string, handler HandlerFunc, opts ...RouteOption) {
	p.Handle(http.MethodHead, path, handler, opts...)
}

func (p *App) Post(path string, handler HandlerFunc, opts ...RouteOption) {
	p.Handle(http.MethodPost, path, handler, opts...)
}

func (p *App) Put(path string, handler HandlerFunc, opts ...RouteOption) {
	p.Handle(http.MethodPut, path, handler, opts...)
}

func (p *App) Patch(path string, handler HandlerFunc, opts ...RouteOption) {
	p.Handle(http.MethodPatch, path, handler, opts...)
}

func (p *App) Delete(path string, handler HandlerFunc, opts ...RouteOption) {
	p.Handle(http.MethodDelete, path, handler, opts...)
}

func (p *App) Connect(path string, handler HandlerFunc, opts ...RouteOption) {
	p.Handle(http.MethodConnect, path, handler, opts...)
}

func (p *App) Options(path string, handler HandlerFunc, opts ...RouteOption) {
	p.Handle(http.MethodOptions, path, handler, opts...)
}

func (p *App) Trace(path string, handler HandlerFunc, opts ...RouteOption) {
	p.Handle(http.MethodTrace, path, handler, opts...)
}

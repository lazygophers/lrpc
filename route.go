package lrpc

import (
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils"
	"github.com/lazygophers/utils/common"
	"github.com/lazygophers/utils/json"
	routing "github.com/qiangxue/fasthttp-routing"
	"reflect"
)

const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH"
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

// ----------App----------

func (p *App) toRouteHandle(handler Handler) routing.Handler {
	return func(routeCtx *routing.Context) error {
		ctx := p.AcquireCtx(routeCtx.RequestCtx)
		defer p.ReleaseCtx(ctx)

		return handler(ctx)
	}
}

func (p *App) toHandle(logic any) routing.Handler {
	sendData := func(ctx *Ctx, data reflect.Value) error {
		switch ctx.ContentType() {
		case "application/json":
			buf, err := json.MarshalString(map[string]any{
				"code": 0,
				"data": data.Interface(),
				"hint": log.GetTrace(),
			})
			if err != nil {
				return err
			}
			return ctx.SendString(buf)
		case "application/protobuf":
			buf, err := json.MarshalString(map[string]any{
				"code": 0,
				"data": data.Interface(),
				"hint": log.GetTrace(),
			})
			if err != nil {
				return err
			}
			return ctx.SendString(buf)
		default:
			if data.IsNil() {
				return nil
			} else {
				buf, err := json.MarshalString(data.Interface())
				if err != nil {
					return err
				}
				return ctx.SendString(buf)
			}
		}
	}

	if h, ok := logic.(Handler); ok {
		return p.toRouteHandle(h)
	}

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

		return p.toRouteHandle(func(ctx *Ctx) error {
			req := reflect.New(in)
			err := ctx.BodyParser(req.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				ctx.app.handleOnError(ctx, common.ErrInvalid(err.Error()))
				return nil
			}

			err = utils.Validate(req.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				ctx.app.handleOnError(ctx, common.NewErrorWithError(err))
				return nil
			}

			out := lv.Call([]reflect.Value{reflect.ValueOf(ctx), req})
			if !out[0].IsNil() {
				log.Errorf("err:%v", out[0].Interface())
				ctx.app.handleOnError(ctx, out[0].Interface().(error))
				return nil
			}

			return sendData(ctx, reflect.Value{})
		})
	}

	// 一个入参，两个出参，那就是需要返回数据的
	if lt.NumIn() == 1 && lt.NumOut() == 2 {
		// 先判断第二个出参是否是error
		x = lt.Out(1)
		for x.Kind() == reflect.Ptr {
			x = x.Elem()
		}
		if x.Name() != "error" {
			panic("2rd out is must error")
		}

		return p.toRouteHandle(func(ctx *Ctx) error {
			out := lv.Call([]reflect.Value{reflect.ValueOf(ctx)})
			if !out[1].IsNil() {
				log.Errorf("err:%v", out[1].Interface())
				ctx.app.handleOnError(ctx, out[1].Interface().(error))
				return nil
			}

			return sendData(ctx, out[0])
		})
	}

	// 两个入参，两个出参，那就是需要解析请求参数，返回数据的
	if lt.NumIn() == 2 && lt.NumOut() == 2 {
		// 先判断第二个出参是否是error
		x = lt.Out(1)
		for x.Kind() == reflect.Ptr {
			x = x.Elem()
		}
		if x.Name() != "error" {
			panic("2rd out is must error")
		}

		// 处理一下入参
		in := lt.In(1)
		for in.Kind() == reflect.Ptr {
			in = in.Elem()
		}

		if in.Kind() != reflect.Struct {
			panic("2rd in is must struct")
		}

		return p.toRouteHandle(func(ctx *Ctx) error {
			req := reflect.New(in)
			err := ctx.BodyParser(req.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				ctx.app.handleOnError(ctx, common.ErrInvalid(err.Error()))
				return nil
			}

			err = utils.Validate(req.Interface())
			if err != nil {
				log.Errorf("err:%v", err)
				ctx.app.handleOnError(ctx, common.NewErrorWithError(err))
				return nil
			}

			out := lv.Call([]reflect.Value{reflect.ValueOf(ctx), req})
			if !out[1].IsNil() {
				log.Errorf("err:%v", out[1].Interface())
				ctx.app.handleOnError(ctx, out[1].Interface().(error))
				return nil
			}

			return sendData(ctx, out[0])
		})
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

		return p.toRouteHandle(func(ctx *Ctx) error {
			out := lv.Call([]reflect.Value{reflect.ValueOf(ctx)})
			if !out[0].IsNil() {
				log.Errorf("err:%v", out[0].Interface())
				ctx.app.handleOnError(ctx, out[0].Interface().(error))
				return nil
			}

			return sendData(ctx, reflect.Value{})
		})
	}

	panic("not support")
}

func (p *App) Use(handlers ...Handler) {
	for _, h := range handlers {
		p.route.Use(p.toRouteHandle(h))
	}
}

func (p *App) GET(path string, logic any) {
	p.route.Get(path, p.toHandle(logic))
}

func (p *App) HEAD(path string, logic any) {
	p.route.Head(path, p.toHandle(logic))
}

func (p *App) POST(path string, logic any) {
	p.route.Post(path, p.toHandle(logic))
}

func (p *App) PUT(path string, logic any) {
	p.route.Put(path, p.toHandle(logic))
}

func (p *App) PATCH(path string, logic any) {
	p.route.Patch(path, p.toHandle(logic))
}

func (p *App) DELETE(path string, logic any) {
	p.route.Delete(path, p.toHandle(logic))
}

func (p *App) CONNECT(path string, logic any) {
	p.route.Connect(path, p.toHandle(logic))
}

func (p *App) OPTIONS(path string, logic any) {
	p.route.Options(path, p.toHandle(logic))
}

func (p *App) TRACE(path string, logic any) {
	p.route.Trace(path, p.toHandle(logic))
}

package lrpc

import (
	"errors"
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"reflect"
)

type Config struct {
	Name string

	OnError func(ctx *Ctx, err error)

	// 用于统一的封包、权限等处理
	AfterHandlerFuncWithRef func(ctx *Ctx, data reflect.Value, err error)
	AfterHandlerFunc        func(ctx *Ctx, err error)
}

var defaultOnError = func(ctx *Ctx, err error) {
	var x *xerror.Error
	var ok bool
	if ok = errors.As(err, &x); !ok {
		x = &xerror.Error{
			Code: -1,
			Msg:  err.Error(),
		}
	}

	err = ctx.SendJson(BaseResponse{
		Code: x.Code,
		Msg:  x.Msg,
		Data: nil,
		Hint: string(log.GetTrace()),
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}
}

var defaultAfterHandlerFuncWithDef = func(ctx *Ctx, data reflect.Value, err error) {
	if err != nil {
		defaultOnError(ctx, err)
		return
	}

	if ctx.IsBodyStream() {
		return
	}

	//if len(ctx.Response().Body()) > 0 {
	//	return nil
	//}

	err = ctx.SendJson(map[string]any{
		"code": 0,
		"data": data.Interface(),
		"hint": log.GetTrace(),
	})
	if err != nil {
		log.Errorf("err:%v", err)
		defaultOnError(ctx, err)
		return
	}
}

var defaultConfig = &Config{
	OnError:                 defaultOnError,
	AfterHandlerFuncWithRef: defaultAfterHandlerFuncWithDef,
}

func (p *App) initConfig() {
	if p.c == nil {
		p.c = defaultConfig
	}

	if p.c.OnError == nil {
		p.c.OnError = defaultOnError
	}

	if p.c.AfterHandlerFuncWithRef == nil {
		p.c.AfterHandlerFuncWithRef = defaultAfterHandlerFuncWithDef
	}
}

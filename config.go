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
	AfterHandlerFunc func(ctx *Ctx, data reflect.Value, err error)
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

	err = ctx.SendJson(map[string]any{
		"code": x.Code,
		"msg":  x.Msg,
		"hint": log.GetTrace(),
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}
}

var defaultAfterHandlerFunc = func(ctx *Ctx, data reflect.Value, err error) {
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
	OnError:          defaultOnError,
	AfterHandlerFunc: defaultAfterHandlerFunc,
}

func (p *App) initConfig() {
	if p.c == nil {
		p.c = defaultConfig
	}

	if p.c.OnError == nil {
		p.c.OnError = defaultOnError
	}

	if p.c.AfterHandlerFunc == nil {
		p.c.AfterHandlerFunc = defaultAfterHandlerFunc
	}
}

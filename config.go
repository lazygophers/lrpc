package lrpc

import (
	"errors"
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"reflect"
)

type ListenData struct {
	Host string
	Port string
	TLS  bool
}

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

	err = ctx.SendJson(&core.BaseResponse{
		Code:    x.Code,
		Message: x.Msg,
		Hint:    log.GetTrace(),
	})
	if err != nil {
		log.Errorf("err:%v", err)
		return
	}
}

var defaultAfterHandlerFuncWithDef = func(ctx *Ctx, data reflect.Value, err error) {
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

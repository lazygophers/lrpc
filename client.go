package lrpc

import (
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

type Client interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
}

var DiscoveryClient = func(c *core.ServiceDiscoveryClient) (Client, *fasthttp.Request) {
	panic("you should registe with discovery github.com/lazygophers/lrpc/middleware/service_discovery")
}

func Call(ctx *Ctx, c *core.ServiceDiscoveryClient, req proto.Message, rsp proto.Message) error {
	var response fasthttp.Response
	client, request := DiscoveryClient(c)

	tranceId := ctx.TranceId()
	if tranceId == "" {
		tranceId = log.GetTrace()
		ctx.SetTranceId(tranceId)
	}

	if tranceId == "" {
		tranceId = log.GenTraceId()
		ctx.SetTranceId(tranceId)
	}

	ctx.Context().Request.Header.VisitAll(func(key, value []byte) {
		request.Header.SetBytesKV(key, value)
	})
	ctx.Context().Request.Header.VisitAllCookie(func(key, value []byte) {
		request.Header.SetCookieBytesKV(key, value)
	})

	request.Header.Set(HeaderContentType, MIMEApplicationProtobuf)
	request.Header.Set(HeaderTrance, tranceId)

	if req != nil {
		buffer, err := proto.Marshal(req)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
		request.SetBody(buffer)
	}

	err := client.Do(request, &response)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	tranceId = string(response.Header.Peek(HeaderTrance))
	log.SetTrace(tranceId)
	ctx.SetTranceId(tranceId)

	baseResp := &core.BaseResponse{}
	err = proto.Unmarshal(response.Body(), baseResp)
	if err != nil {
		if response.StatusCode() != fasthttp.StatusOK {
			log.Errorf("status %d", response.StatusCode())
			return xerror.New(int32(response.StatusCode()))
		}
		log.Errorf("err:%v", err)
		return err
	}

	if baseResp.Code != 0 {
		return xerror.NewErrorWithMsg(baseResp.Code, baseResp.Message)
	}

	if rsp != nil {
		err = baseResp.Data.UnmarshalTo(rsp)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	return nil
}

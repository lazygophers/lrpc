package lrpc

import (
	"github.com/lazygophers/log"
	"github.com/lazygophers/lrpc/middleware/core"
	"github.com/lazygophers/lrpc/middleware/xerror"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

var DiscoveryClient = func(c *core.ServiceDiscoveryClient) (*fasthttp.HostClient, *fasthttp.Request) {
	panic("you should registe with discovery github.com/lazygophers/lrpc/middleware/service_discovery")
}

func Call(ctx *Ctx, c *core.ServiceDiscoveryClient, req proto.Message, rsp proto.Message) error {
	var response fasthttp.Response
	client, request := DiscoveryClient(c)

	request.Header.Set(HeaderContentType, MIMEProtobuf)
	request.Header.Set(HeaderTrance, ctx.TranceId())

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

	log.Info(response.Body())
	log.Info(string(response.Header.ContentType()))

	baseResp := &core.BaseResponse{}
	err = proto.Unmarshal(response.Body(), baseResp)
	if err != nil {
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

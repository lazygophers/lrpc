package lrpc

import (
	"fmt"
	"github.com/lazygophers/log"
	"github.com/lazygophers/utils/network"
	"github.com/lazygophers/utils/runtime"
	"github.com/valyala/fasthttp"
	"net"
)

func (p *App) Handler(c *fasthttp.RequestCtx) {
	ctx := p.AcquireCtx(c)
	defer p.ReleaseCtx(ctx)

	route := p.routes[ctx.Method()]
	if route == nil {
		log.Errorf("not found route, method:%s, path:%s", ctx.Method(), ctx.Path())
		c.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	res, ok := route.Search(ctx.Path())
	if !ok {
		log.Errorf("not found route, method:%s, path:%s", ctx.Method(), ctx.Path())
		c.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	ctx.setParam(res.Params)

	err := res.Item(ctx)
	if err != nil {
		log.Errorf("err:%v", err)
		p.onError(ctx, err)
	}

	return
}

func (p *App) ErrorHandler(c *fasthttp.RequestCtx, err error) {
	ctx := p.AcquireCtx(c)
	defer p.ReleaseCtx(ctx)

	p.onError(ctx, err)
}

func (p *App) initServer() {
	p.server = &fasthttp.Server{
		Handler:                            p.Handler,
		ErrorHandler:                       p.ErrorHandler,
		HeaderReceived:                     nil,
		ContinueHandler:                    nil,
		Name:                               p.c.Name,
		Concurrency:                        0,
		ReadBufferSize:                     0,
		WriteBufferSize:                    0,
		ReadTimeout:                        0,
		WriteTimeout:                       0,
		IdleTimeout:                        0,
		MaxConnsPerIP:                      0,
		MaxRequestsPerConn:                 0,
		MaxKeepaliveDuration:               0,
		MaxIdleWorkerDuration:              0,
		TCPKeepalivePeriod:                 0,
		MaxRequestBodySize:                 0,
		DisableKeepalive:                   false,
		TCPKeepalive:                       false,
		ReduceMemoryUsage:                  false,
		GetOnly:                            false,
		DisablePreParseMultipartForm:       false,
		LogAllErrors:                       true,
		SecureErrorLogMessage:              false,
		DisableHeaderNamesNormalizing:      false,
		SleepWhenConcurrencyLimitsExceeded: 0,
		NoDefaultServerHeader:              false,
		NoDefaultDate:                      false,
		NoDefaultContentType:               false,
		KeepHijackedConns:                  true,
		CloseOnShutdown:                    true,
		StreamRequestBody:                  false,
		ConnState:                          nil,
		Logger:                             log.Clone(),
		TLSConfig:                          nil,
		FormValueFunc:                      nil,
	}
}

type listenConfig struct {
	port int

	bindIp  string
	address string
}

func (c *listenConfig) apply() {
	if c.address == "" {
		c.address = fmt.Sprintf(":%d", c.port)
	}
}

type ListenHandler func(a *App, c *listenConfig)

var EmptyListenHandler = func(a *App, c *listenConfig) {}

func ListenWithIp(ip string) ListenHandler {
	return func(a *App, c *listenConfig) {
		c.bindIp = ip
		c.address = fmt.Sprintf("%s:%d", ip, c.port)
	}
}

func ListenWithLocal(usev6 ...bool) ListenHandler {
	if len(usev6) > 0 && usev6[0] {
		return ListenWithIp("[::1]")
	}
	return ListenWithIp("127.0.0.1")
}

func ListenWithLan(prev6 ...bool) ListenHandler {
	// 找到内网 IP
	if ip := network.GetListenIp(prev6...); ip != "" {
		return ListenWithIp(ip)
	}

	log.Error("get interface ip failed")

	return EmptyListenHandler
}

func (p *App) ListenAndServe(port int, handlers ...ListenHandler) (err error) {
	c := &listenConfig{
		port: port,
	}

	for _, handler := range handlers {
		handler(p, c)
	}

	c.apply()

	// TODO 服务发现和服务注册
	conn, err := net.Listen("tcp", c.address)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}
	defer conn.Close()

	err = p.server.Serve(conn)
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	listen := ListenData{
		Host: "",
		Port: "",
		TLS:  false,
	}

	listen.Host, listen.Port, err = net.SplitHostPort(conn.Addr().String())
	if err != nil {
		log.Errorf("err:%v", err)
		return err
	}

	if p.server.TLSConfig != nil {
		listen.TLS = true
	}

	for _, logic := range p.hook.onListen {
		err = logic(listen)
		if err != nil {
			_ = p.server.Shutdown()
			log.Errorf("err:%v", err)
			return err
		}
	}

	runtime.WaitExit()

	for _, logic := range p.hook.onShutdown {
		err = logic(listen)
		if err != nil {
			log.Errorf("err:%v", err)
			return err
		}
	}

	return nil
}

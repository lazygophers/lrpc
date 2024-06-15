package lrpc

type Hooks struct {
	onRoute    []func(*Route) error
	onListen   []func(ListenData) error
	onShutdown []func(ListenData) error
}

func (p *Hooks) OnRoute(route func(*Route) error) {
	p.onRoute = append(p.onRoute, route)
}

func (p *Hooks) OnListen(logic func(ListenData) error) {
	p.onListen = append(p.onListen, logic)
}

func (p *Hooks) OnShutdown(logic func(ListenData) error) {
	p.onShutdown = append(p.onShutdown, logic)
}

func (p *App) Hooks() *Hooks {
	return p.hook
}

func (p *App) OnRoute(route func(*Route) error) {
	p.hook.OnRoute(route)
}

func (p *App) OnListen(logic func(ListenData) error) {
	p.hook.OnListen(logic)
}

func (p *App) OnShutdown(logic func(ListenData) error) {
	p.hook.OnShutdown(logic)
}

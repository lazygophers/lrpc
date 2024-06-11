package lrpc

type Route struct {
	Method string
	Path   string

	Handler HandlerFunc

	Before, After []HandlerFunc

	// 可以存储一些类似于权限等信息，会在调用前写入到 local 中
	Extra map[string]any
}

type RouteOption func(r *Route)

func RouteWithPrefix(prefix string) RouteOption {
	return func(r *Route) {
		r.Path = prefix + r.Path
	}
}

func RouteWithBefore(before HandlerFunc) RouteOption {
	return func(r *Route) {
		r.Before = append(r.Before, before)
	}
}

func RouteWithAfter(after HandlerFunc) RouteOption {
	return func(r *Route) {
		r.After = append(r.After, after)
	}
}

func RouteWithExtra(extra map[string]any) RouteOption {
	return func(r *Route) {
		r.Extra = extra
	}
}

func RouteWithMergeExtra(merge map[string]any) RouteOption {
	return func(r *Route) {
		if r.Extra == nil {
			r.Extra = make(map[string]any)
		}

		for k, v := range merge {
			r.Extra[k] = v
		}
	}
}

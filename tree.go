package lrpc

type SearchResult[M any] struct {
	Item M

	Params map[string]string
}

type searchTreeNode[M any] struct {
	empty bool
	item  M

	chiledren [2]map[string]*searchTreeNode[M]
}

func newSearchTreeNode[M any]() *searchTreeNode[M] {
	return &searchTreeNode[M]{
		empty: true,
		chiledren: [2]map[string]*searchTreeNode[M]{
			{},
			{},
		},
	}
}

func (p *searchTreeNode[M]) setItem(item M) {
	p.empty = false
	p.item = item
}

func (p *searchTreeNode[M]) isEmpty() bool {
	return p.empty
}

func (p *searchTreeNode[M]) getChildren(token string) map[string]*searchTreeNode[M] {
	if token == "" {
		return p.chiledren[0]
	}

	if token[0] != ':' {
		return p.chiledren[0]
	}

	return p.chiledren[1]
}

func (p *searchTreeNode[M]) getOrCreateNode(token string) *searchTreeNode[M] {
	cc := p.getChildren(token)
	if _, ok := cc[token]; !ok {
		cc[token] = newSearchTreeNode[M]()
	}

	return cc[token]
}

func (p *searchTreeNode[M]) forEach(logic func(key string, value *searchTreeNode[M]) bool) bool {
	for _, children := range p.chiledren {
		for k, v := range children {
			if logic(k, v) {
				return true
			}
		}
	}

	return false
}

type SearchTree[M any] struct {
	root *searchTreeNode[M]
}

func (p *SearchTree[M]) Add(route string, item M) {
	if route == "" || route[0] != '/' {
		panic("path should start with /")
	}

	p.add(p.root, route, item)
}

func (p *SearchTree[M]) add(nd *searchTreeNode[M], route string, item M) {
	if route == "" {
		nd.setItem(item)
		return
	}

	for i := range route {
		if route[i] != '/' {
			continue
		}

		token := route[:i]
		child := nd.getOrCreateNode(token)
		p.add(child, route[i+1:], item)
		return
	}

	child := nd.getOrCreateNode(route)
	child.setItem(item)
}

func (p *SearchTree[M]) Search(route string) (*SearchResult[M], bool) {
	if route == "" || route[0] != '/' {
		panic("path should start with /")
	}

	var res SearchResult[M]
	return &res, p.search(p.root, route, &res)
}

func (p *SearchTree[M]) search(nd *searchTreeNode[M], route string, res *SearchResult[M]) bool {
	if route == "" && !nd.isEmpty() {
		res.Item = nd.item
		return true
	}

	for i := range route {
		if route[i] != '/' {
			continue
		}

		token := route[:i]
		return nd.forEach(func(key string, value *searchTreeNode[M]) bool {
			if key != "" && key[0] == ':' {
				if !p.search(value, route[i+1:], res) {
					return false
				}

				if res.Params == nil {
					res.Params = make(map[string]string)
				}

				res.Params[key[1:]] = token

			} else {
				if key != token || !p.search(value, route[i+1:], res) {
					return false
				}
			}

			return true
		})
	}

	return nd.forEach(func(key string, value *searchTreeNode[M]) bool {
		if key != "" && key[0] == ':' {
			if !value.isEmpty() {
				res.Item = value.item

				if res.Params == nil {
					res.Params = make(map[string]string)
				}

				res.Params[key[1:]] = route

				return true
			}

		} else {
			if key == route {
				if !value.isEmpty() {
					res.Item = value.item
					return true
				}
			}
		}

		return false
	})
}

func NewSearchTree[M any]() *SearchTree[M] {
	return &SearchTree[M]{
		root: newSearchTreeNode[M](),
	}
}

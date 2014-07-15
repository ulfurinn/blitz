package blitz

type RoutingTable map[string]*Router

type Router struct {
	routes  RoutingTable
	handler *Instance
}

func NewRouter() *Router {
	return &Router{routes: make(RoutingTable)}
}

func (r *Router) Mount(path []string, handler *Instance) {
	if len(path) == 0 || len(path) == 1 && path[0] == "" {
		if r.handler == nil || r.handler.patch <= handler.patch {
			r.handler = handler
		}
		return
	}
	first := path[0]
	router, ok := r.routes[first]
	if !ok {
		router = NewRouter()
		r.routes[first] = router
	}
	router.Mount(path[1:], handler)
}

func (r *Router) Route(path []string) (handler *Instance) {
	if len(path) == 0 || len(path) == 1 && path[0] == "" {
		return r.handler
	}
	router, ok := r.routes[path[0]]
	if ok {
		handler = router.Route(path[1:])
	}
	if handler != nil {
		return
	}
	router, ok = r.routes["*"]
	if ok {
		handler = router.Route(path[1:])
	}
	return
}

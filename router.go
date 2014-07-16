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
		if r.handler == nil || r.handler.Patch <= handler.Patch {
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

func (r *Router) UsedInstances() (result []*Instance) {
	used := make(map[*Instance]struct{})
	if r.handler != nil {
		used[r.handler] = struct{}{}
	}
	for _, router := range r.routes {
		for _, i := range router.UsedInstances() {
			used[i] = struct{}{}
		}
	}
	for i, _ := range used {
		result = append(result, i)
	}
	return
}

func (r *Router) snapshot(root string) (result []*SnapshotRoute) {
	if r.handler != nil {
		result = append(result, &SnapshotRoute{Path: root, Instance: r.handler})
	}
	for component, router := range r.routes {
		var newroot string
		if root == "/" {
			newroot = root + component
		} else {
			newroot = root + "/" + component
		}
		result = append(result, router.snapshot(newroot)...)
	}
	return
}

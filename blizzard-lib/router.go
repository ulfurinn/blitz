package blizzard

import "sync/atomic"

type RoutingTable map[string]*Router

type Router struct {
	Path          string
	routers       RoutingTable
	handler       *Process
	requests      int64
	totalRequests uint64
	written       uint64
}

func NewRouter() *Router {
	return &Router{routers: make(RoutingTable)}
}

func (r *Router) Mount(path []string, handler *Process, prefix string) {
	if len(path) == 0 || len(path) == 1 && path[0] == "" {
		if r.handler == nil || r.handler.Patch <= handler.Patch {
			r.handler = handler
		}
		return
	}
	first := path[0]
	router, ok := r.routers[first]
	routePath := prefix + "/" + first
	if !ok {
		router = NewRouter()
		router.Path = routePath
		r.routers[first] = router
	}
	router.Mount(path[1:], handler, routePath)
}

func (r *Router) Unmount(proc *Process) {
	routers := make(RoutingTable)
	for key, router := range r.routers {
		if router.handler != proc {
			routers[key] = router
		}
		router.Unmount(proc)
	}
	if r.handler == proc {
		r.handler = nil //	mainly to preserve the root router in master
	}
	r.routers = routers
}

func (r *Router) Route(path []string) (handlingRouter *Router, handler *Process) {
	if len(path) == 0 || len(path) == 1 && path[0] == "" {
		return r, r.handler
	}
	router, ok := r.routers[path[0]]
	if ok {
		handlingRouter, handler = router.Route(path[1:])
	}
	if handler != nil {
		return
	}
	router, ok = r.routers["*"]
	if ok {
		handlingRouter, handler = router.Route(path[1:])
	}
	return
}

func (r *Router) UsedInstances() (result []*Process) {
	used := make(map[*Process]struct{})
	if r.handler != nil {
		used[r.handler] = struct{}{}
	}
	for _, router := range r.routers {
		for _, i := range router.UsedInstances() {
			used[i] = struct{}{}
		}
	}
	for i, _ := range used {
		result = append(result, i)
	}
	return
}

func (r *Router) snapshot() (result []*SnapshotRoute) {
	if r.handler != nil {
		result = append(result, &SnapshotRoute{
			Path:          r.Path,
			Process:       r.handler,
			Requests:      atomic.LoadInt64(&r.requests),
			TotalRequests: atomic.LoadUint64(&r.totalRequests),
			Written:       atomic.LoadUint64(&r.written),
		})
	}
	for _, router := range r.routers {
		result = append(result, router.snapshot()...)
	}
	return
}

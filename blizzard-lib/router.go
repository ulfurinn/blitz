package blizzard

import (
	"fmt"
	"net/http"
	"sync"
	"unsafe"

	"sync/atomic"
)

type RoutingTable map[string]*Router

type RouteSet struct {
	routers map[int]*Router
	lock    sync.RWMutex
}

func NewRouteSet() *RouteSet {
	return &RouteSet{
		routers: make(map[int]*Router),
	}
}

func (r *RouteSet) reading(f func()) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	f()
}

func (r *RouteSet) writing(f func()) {
	r.lock.Lock()
	defer r.lock.Unlock()
	f()
}

func (r *RouteSet) forVersion(v int, create bool) (router *Router) {
	router = r.routers[v]
	if router == nil && create {
		router = NewRouter()
		r.routers[v] = router
	}
	return
}

func (r *RouteSet) UsedInstances() (result ProgGroupSet) {
	result = make(map[*ProcGroup]struct{})
	for _, router := range r.routers {
		for i := range router.UsedInstances() {
			result[i] = struct{}{}
		}
	}
	return
}

func (r *RouteSet) ServeHTTP(resp http.ResponseWriter, req *http.Request) http.HandlerFunc {
	version, path, err := (PathVersionStrategy{}).Version(req)
	if err != nil {
		resp.WriteHeader(400)
		fmt.Fprint(resp, err)
		return nil
	}
	router := r.forVersion(version, false)
	if router == nil {
		resp.WriteHeader(400)
		fmt.Fprintf(resp, "Version %d is not recognised", version)
		return nil
	}
	req.Header.Set("X-Blitz-Path", concatPath(path))
	return router.ServeHTTP(path)
}

type Router struct {
	Path          string
	routers       RoutingTable
	handler       *ProcGroup
	Requests      int64
	TotalRequests uint64
	Written       uint64
}

func NewRouter() *Router {
	return &Router{routers: make(RoutingTable)}
}

func (r *Router) Mount(path []string, handler *ProcGroup, prefix string) {
	if len(path) == 0 || len(path) == 1 && path[0] == "" {
		if r.handler == nil || r.handler.Patch <= handler.Patch {
			log("[router] mounting proc %p under %s\n", handler, prefix)
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

func (r *Router) Unmount(proc *ProcGroup) {
	routers := make(RoutingTable)
	for key, router := range r.routers {
		if router.handler != proc {
			routers[key] = router
		}
		router.Unmount(proc)
	}
	if r.handler == proc {
		log("[router] unmounting %p from %s\n", proc, r.Path)
		r.handler = nil //	mainly to preserve the root router in master
	}
	r.routers = routers
}

//	remove the closure, make ServeHTTP a conventional handler method on Router
func (r *Router) ServeHTTP(path []string) http.HandlerFunc {

	methodRoute, h := r.Route(path)
	if h == nil {
		return notFound
	}
	return func(resp http.ResponseWriter, req *http.Request) {
		counter := &countingResponseWriter{ResponseWriter: resp}
		h.inc()
		methodRoute.inc()
		defer func() {
			h.dec()
			methodRoute.dec()
			counter.update(methodRoute, h)
		}()
		h.ServeHTTP(counter, req)
	}

}

func (r *Router) Route(path []string) (handlingRouter *Router, handler *ProcGroup) {
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

func (r *Router) UsedInstances() (result ProgGroupSet) {
	result = make(map[*ProcGroup]struct{})
	if r.handler != nil {
		result[r.handler] = struct{}{}
	}
	for _, router := range r.routers {
		for i := range router.UsedInstances() {
			result[i] = struct{}{}
		}
	}
	return
}

func (r *Router) inc() {
	atomic.AddInt64(&r.Requests, 1)
	atomic.AddUint64(&r.TotalRequests, 1)
}

func (r *Router) dec() {
	atomic.AddInt64(&r.Requests, -1)
}

func (r *Router) snapshot() (result []*SnapshotRoute) {
	if r.handler != nil {
		result = append(result, &SnapshotRoute{
			Path:          r.Path,
			Process:       uintptr(unsafe.Pointer(r.handler)),
			Requests:      atomic.LoadInt64(&r.Requests),
			TotalRequests: atomic.LoadUint64(&r.TotalRequests),
			Written:       atomic.LoadUint64(&r.Written),
		})
	}
	for _, router := range r.routers {
		result = append(result, router.snapshot()...)
	}
	return
}

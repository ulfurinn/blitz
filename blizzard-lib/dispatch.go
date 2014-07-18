package blizzard

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type VersionStrategy interface {
	Version(*http.Request) (int, []string, error)
}

type PathVersionStrategy struct{}

func (PathVersionStrategy) Version(req *http.Request) (version int, path []string, err error) {
	split := strings.Split(req.URL.Path[1:], "/")
	if len(split) == 0 {
		err = fmt.Errorf("No version provided")
		return
	}
	_, err = fmt.Sscanf(split[0], "v%d", &version)
	if err != nil {
		err = fmt.Errorf("No version provided")
		return
	}
	path = split[1:]
	return
}

func concatPath(path []string) string {
	var buf bytes.Buffer
	for _, c := range path {
		fmt.Fprint(&buf, "/", c)
	}
	return buf.String()
}

func (m *Master) BlitzDispatch(resp http.ResponseWriter, req *http.Request) {
	version, path, err := (PathVersionStrategy{}).Version(req)
	if err != nil {
		resp.WriteHeader(400)
		fmt.Fprint(resp, err)
		return
	}
	m.routeLock.RLock()
	versionRouter, ok := m.routers[version]
	if !ok {
		resp.WriteHeader(400)
		fmt.Fprintf(resp, "Version %d is not recognised", version)
		return
	}
	r, h := versionRouter.Route(path)
	// do this before unlocking so that the collector in announce will see it as busy
	if h != nil {
		atomic.AddInt64(&r.requests, 1)
		atomic.AddUint64(&r.totalRequests, 1)
		atomic.AddInt64(&h.Requests, 1)
		atomic.AddUint64(&h.TotalRequests, 1)
	}
	m.routeLock.RUnlock()
	if h == nil {
		resp.WriteHeader(404)
		return
	}
	counter := &countingResponseWriter{ResponseWriter: resp}
	defer func() {
		atomic.AddInt64(&r.requests, -1)
		atomic.AddInt64(&h.Requests, -1)
		atomic.AddUint64(&r.written, counter.written)
		atomic.AddUint64(&h.Written, counter.written)
	}()
	req.Header.Set("X-Blitz-Path", concatPath(path))
	h.ServeHTTP(counter, req)

}

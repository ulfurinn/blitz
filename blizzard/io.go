package blizzard

import (
	"sync/atomic"

	"net/http"
)

type countingResponseWriter struct {
	http.ResponseWriter
	written uint64
}

func (w *countingResponseWriter) Write(data []byte) (written int, err error) {
	written, err = w.ResponseWriter.Write(data)
	if err == nil {
		w.written += uint64(written)
	}
	return
}

func (w *countingResponseWriter) update(r *Router, pg *ProcGroup) {
	atomic.AddUint64(&r.Written, w.written)
	atomic.AddUint64(&pg.Written, w.written)
}

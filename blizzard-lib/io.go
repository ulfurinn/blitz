package blizzard

import "net/http"

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

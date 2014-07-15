package blitz

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"

	"os"
	"os/signal"
)

var procID string

func init() {
	flag.StringVar(&procID, "blitz-proc-id", "", "internal process ID")
}

type Worker struct {
	conn     net.Conn
	socket   string
	listener net.Listener
	Patch    int64
	Handler  http.Handler
	Paths    []PathSpec
}

func (w *Worker) Run() (err error) {
	err = w.init()
	if err != nil {
		return
	}
	err = w.announce(w.Paths)
	if err != nil {
		return
	}
	w.Serve(w.Handler)
	return
}

func (w *Worker) init() (err error) {
	os.MkdirAll("blitz", os.ModeDir|0775)
	err = w.listen()
	if err != nil {
		return
	}
	err = w.connect()
	return
}

func (w *Worker) connect() (err error) {
	conn, err := net.Dial("unix", "blitz/ctl")
	if err != nil {
		return
	}
	w.conn = conn
	return
}

func (w *Worker) listen() (err error) {
	w.socket = fmt.Sprintf("blitz/%d.worker", os.Getpid())
	listener, err := net.Listen("unix", w.socket)
	if err != nil {
		return
	}
	w.listener = listener
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go w.cleanup(ch)
	return
}

func (w *Worker) cleanup(ch chan os.Signal) {
	<-ch
	err := w.listener.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	err = w.conn.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(0)
}

func (w *Worker) send(data interface{}) error {
	encoder := json.NewEncoder(w.conn)
	return encoder.Encode(data)
}

func (w *Worker) announce(spec []PathSpec) error {
	a := Command{}
	a.Type = "announce"
	a.ProcID = procID
	a.PID = os.Getpid()
	a.Network = "unix"
	a.Address = w.socket
	a.Patch = w.Patch
	a.Paths = spec
	return w.send(a)
}

func (w *Worker) Serve(handler http.Handler) {
	http.Serve(w.listener, handler)
}

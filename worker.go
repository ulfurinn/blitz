package blitz

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"syscall"

	"os"
	"os/signal"
)

var tag string
var bootstrap bool

type Worker struct {
	conn      net.Conn
	socket    string
	listener  net.Listener
	AppName   string
	Patch     int64
	Handler   http.Handler
	Paths     []PathSpec
	Bootstrap Bootstrapper
}

func (w *Worker) Run() (err error) {
	err = w.init()
	if err != nil {
		return
	}
	if bootstrap {
		err = w.bootstrap()
		w.cleanup()
		return
	}
	err = w.listen()
	if err != nil {
		return
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go w.waitForCleanup(ch)

	err = w.announce(w.Paths)
	if err != nil {
		return
	}
	w.Serve(w.Handler)
	return
}

func (w *Worker) init() (err error) {
	CreateDirectoryStructure()
	err = w.connect()
	return
}

func (w *Worker) connect() (err error) {
	conn, err := net.Dial("unix", ControlAddress())
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
	return
}

func (w *Worker) waitForCleanup(ch chan os.Signal) {
	<-ch
	w.cleanup()
}

func (w *Worker) cleanup() {
	if w.listener != nil {
		w.listener.Close()
	}
	if w.conn != nil {
		w.conn.Close()
	}
	os.Exit(0)
}

func (w *Worker) send(data interface{}) error {
	encoder := json.NewEncoder(w.conn)
	return encoder.Encode(data)
}

func (w *Worker) bootstrap() error {
	cmd := BootstrapCommand{}
	cmd.Type = "bootstrap"
	cmd.AppName = w.AppName
	cmd.BinaryTag = tag
	err := w.Bootstrap(&cmd)
	if err != nil {
		return err
	}
	if cmd.Instances == 0 {
		cmd.Instances = 1
	}
	return w.send(cmd)
}

func (w *Worker) announce(spec []PathSpec) error {
	a := AnnounceCommand{}
	a.Type = "announce"
	a.ProcTag = tag
	a.Network = "unix"
	a.Address = w.socket
	a.Patch = w.Patch
	a.Paths = spec
	return w.send(a)
}

func (w *Worker) Serve(handler http.Handler) {
	http.Serve(w.listener, handler)
}

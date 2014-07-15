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
}

func (w *Worker) Init() {
	os.MkdirAll("blitz", os.ModeDir|0775)
}

func (w *Worker) Connect() error {
	conn, err := net.Dial("unix", "blitz/ctl")
	if err != nil {
		return err
	}
	w.conn = conn
	return nil
}

func (w *Worker) Listen() error {
	w.socket = fmt.Sprintf("blitz/%d.worker", os.Getpid())
	listener, err := net.Listen("unix", w.socket)
	w.listener = listener
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go func() {
		<-ch
		err := listener.Close()
		if err != nil {
			fatal(err)
		}
		os.Exit(0)
	}()
	return err
}

func (w *Worker) send(data interface{}) error {
	encoder := json.NewEncoder(w.conn)
	return encoder.Encode(data)
}

func (w *Worker) Announce(spec []PathSpec) error {
	a := Command{}
	a.Type = "announce"
	a.ProcID = procID
	a.Network = "unix"
	a.Address = w.socket
	a.Patch = w.Patch
	a.Paths = spec
	return w.send(a)
}

func (w *Worker) Serve(handler http.Handler) {
	http.Serve(w.listener, handler)
}

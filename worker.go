package blitz

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"syscall"

	"os"
	"os/signal"
)

var tag string
var Config string
var bootstrap bool

type Worker struct {
	conn      net.Conn
	socket    string
	AppName   string
	Patch     uint64
	Handler   http.Handler
	Paths     []PathSpec
	Bootstrap Bootstrapper
	Run       Runner
	shutdown  []func()
	Mainproc  sync.WaitGroup
}

func (w *Worker) Start() (err error) {
	err = w.init()
	if err != nil {
		return
	}
	if bootstrap {
		err = w.bootstrap()
		w.cleanup()
		return
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go w.waitForCleanup(ch)

	err = w.announce()
	if err != nil {
		w.cleanup()
	}
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
	err = w.send(ConnectionTypeCommand{Command{"connection-type"}, "worker"})
	if err != nil {
		return
	}
	return
}

func (w *Worker) OnShutdown(f func()) {
	w.shutdown = append(w.shutdown, f)
}

func StandardRunner(w *Worker, cmd *AnnounceCommand) error {
	listener, err := net.Listen("unix", cmd.Address)
	if err != nil {
		return err
	}
	w.OnShutdown(func() {
		listener.Close()
	})
	w.Mainproc.Add(1)
	go func() {
		http.Serve(listener, w.Handler)
		w.Mainproc.Done()
	}()
	return nil
}

func (w *Worker) waitForCleanup(ch chan os.Signal) {
	<-ch
	fmt.Fprintln(os.Stderr, "shutdown signal")
	w.cleanup()
}

func (w *Worker) cleanup() {
	if w.conn != nil {
		w.conn.Close()
	}
	for _, f := range w.shutdown {
		f()
	}
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
	if w.Bootstrap != nil {
		err := w.Bootstrap(w, &cmd)
		if err != nil {
			return err
		}
	}
	if cmd.Instances == 0 {
		cmd.Instances = 1
	}
	return w.send(cmd)
}

func (w *Worker) announce() error {
	a := AnnounceCommand{}
	a.Type = "announce"
	a.ProcTag = tag
	a.Network = "unix"
	a.Address = fmt.Sprintf("blitz/%d.worker", os.Getpid())
	a.Patch = w.Patch
	a.Paths = w.Paths
	if w.Run != nil {
		err := w.Run(w, &a)
		if err != nil {
			return err
		}
	}
	return w.send(a)
}

func (w *Worker) Wait() {
	w.Mainproc.Wait()
}

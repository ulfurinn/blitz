package blizzard

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"bitbucket.org/ulfurinn/blitz"
)

func closeSocketOnShutdown(listener net.Listener) {
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		err := listener.Close()
		if err != nil {
			fatal(err)
		}
		os.Exit(0)
	}()
}

type workerMonitor interface {
	WorkerClosed(*WorkerConnection)
}

type WorkerConnection struct {
	conn    net.Conn
	server  *Blizzard
	monitor workerMonitor
}

func (w *WorkerConnection) send(data interface{}) error {
	return json.NewEncoder(w.conn).Encode(data)
}

func (w *WorkerConnection) closed() {
	w.conn.Close()
	if w.monitor != nil {
		mon := w.monitor
		w.monitor = nil
		mon.WorkerClosed(w)
	}
}

func (w *WorkerConnection) isDisconnect(err error) bool {
	if err == io.EOF {
		log("[control %p] EOF\n", w)
		return true
	}
	if opError, isOpError := err.(*net.OpError); isOpError {
		errno := opError.Err.(syscall.Errno)
		if errno == syscall.ECONNRESET {
			log("[control %p] ECONNRESET\n", w)
			return true
		}
	}
	return false
}

func (w *WorkerConnection) Run() {
	defer w.closed()
	log("[control %p] opened channel\n", w)
	decoder := json.NewDecoder(w.conn)
	for {
		var raw json.RawMessage
		err := decoder.Decode(&raw)
		if err != nil {
			if !w.isDisconnect(err) {
				log("[control] %v\n", err)
			}
			return
		}
		var base blitz.Command
		json.Unmarshal(raw, &base)
		var parsed interface{}
		log("[control %p] %s\n", w, base.Type)
		switch base.Type {
		case "announce":
			var announce blitz.AnnounceCommand
			json.Unmarshal(raw, &announce)
			parsed = announce
		case "deploy":
			var deploy blitz.DeployCommand
			json.Unmarshal(raw, &deploy)
			parsed = deploy
		case "bootstrap":
			var bootstrap blitz.BootstrapCommand
			json.Unmarshal(raw, &bootstrap)
			parsed = bootstrap
		}
		if parsed != nil {
			cmd := workerCommand{command: parsed, WorkerConnection: w}
			clientResp := w.server.Command(cmd)
			w.send(clientResp)
		}
	}
}

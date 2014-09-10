package blizzard

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"reflect"
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
	connType string
	conn     net.Conn
	server   *Blizzard
	monitor  workerMonitor
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
		return true
	}
	if opError, isOpError := err.(*net.OpError); isOpError {
		errno := opError.Err.(syscall.Errno)
		if errno == syscall.ECONNRESET {
			return true
		}
	}
	return false
}

func blitzCommand(typ string) (cmd interface{}) {
	switch typ {
	case "announce":
		return &blitz.AnnounceCommand{}
	case "deploy":
		return &blitz.DeployCommand{}
	case "bootstrap":
		return &blitz.BootstrapCommand{}
	case "list-apps":
		return &blitz.ListExecutablesCommand{}
	case "restart-takeover":
		return &blitz.RestartTakeoverCommand{}
	}
	return nil
}

func (w *WorkerConnection) Run() {
	defer w.closed()
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
		log("[control %p] %v\n", w, string(raw))
		var base blitz.Command
		json.Unmarshal(raw, &base)
		if base.Type == "connection-type" {
			var ct blitz.ConnectionTypeCommand
			json.Unmarshal(raw, &ct)
			w.connType = ct.ConnectionType
			continue
		}
		var response interface{}
		parsed := blitzCommand(base.Type)
		if parsed != nil {
			err = json.Unmarshal(raw, reflect.ValueOf(parsed).Interface())
			if err == nil {
				response = w.server.Command(workerCommand{command: parsed, WorkerConnection: w})
			} else {
				e := err.Error()
				response = blitz.Response{Error: &e}
			}
		} else {
			e := fmt.Sprintf("unrecognized command: %s", base.Type)
			response = blitz.Response{Error: &e}
		}
		w.send(response)
	}
}

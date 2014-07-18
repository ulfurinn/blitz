package blizzard

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"

	"bitbucket.org/ulfurinn/blitz"
)

func closeSocketOnShutdown(listener net.Listener) {
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

}

func (m *Master) ProcessControl(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fatal(err)
		}
		worker := &WorkerConnection{conn: conn, master: m}
		go worker.Run()
	}
}

func (m *Master) handleCommand(cmd masterRequest) {
	switch cmd.cmd.Type {
	case "announce":
		m.Announce(cmd.cmd, cmd.conn)
		cmd.ret <- masterResponse{}
	case "deploy":
		err := m.Deploy(cmd.cmd.Binary)
		cmd.ret <- masterResponse{err: err}
	default:
		cmd.ret <- masterResponse{err: fmt.Errorf("unknown command %s", cmd.cmd.Type)}
	}
}

func (m *Master) connectionClosed(w *WorkerConnection) {
	var proc *Process
	for _, p := range m.procs {
		if p.connection == w {
			proc = p
			break
		}
	}
	if proc == nil {
		return
	}
	//fmt.Printf("instance left: %v\n", *proc)
	m.routeLock.Lock()
	defer m.routeLock.Unlock()
	proc.dead = true
	m.Unmount(proc)
	m.CollectUnusedInstances()
}

type WorkerConnection struct {
	conn   net.Conn
	master *Master
}

func (w *WorkerConnection) send(data interface{}) error {
	return json.NewEncoder(w.conn).Encode(data)
}

func (w *WorkerConnection) closed() {
	w.conn.Close()
	w.master.connectionClosedCh <- w
}

func (w *WorkerConnection) Run() {
	defer w.closed()
	decoder := json.NewDecoder(w.conn)
	for {
		v := blitz.Command{}
		err := decoder.Decode(&v)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, err)
			}
			return
		}
		ret := make(chan masterResponse, 1)
		w.master.cmdCh <- masterRequest{cmd: v, conn: w, ret: ret}
		resp := <-ret
		clientResp := blitz.Response{}
		if resp.err != nil {
			clientResp.Error = resp.err.Error()
		}
		w.send(clientResp)
	}
}

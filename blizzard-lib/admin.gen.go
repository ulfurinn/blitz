//  GENERATED CODE, DO NOT EDIT
package blizzard

import "time"
import "bitbucket.org/ulfurinn/gen_proc"

type assetServerCh struct {
	chMsg  chan gen_proc.ProcCall
	chStop chan struct{}
}

func NewassetServerCh() *assetServerCh {
	return &assetServerCh{
		chMsg:  make(chan gen_proc.ProcCall),
		chStop: make(chan struct{}, 1),
	}
}

type assetServerEnvelopeRegister struct {
	proc *assetServer
	c    *wsConnection
	ret  chan assetServerRegisterReturn
	gen_proc.Envelope
}

func (msg assetServerEnvelopeRegister) Call() {
	ret := make(chan assetServerRegisterReturn, 1)
	go func(ret chan assetServerRegisterReturn) {
		msg.proc.handleRegister(msg.c)
		ret <- assetServerRegisterReturn{}
	}(ret)
	select {
	case result := <-ret:
		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type assetServerRegisterReturn struct {
}

// Register is a gen_server interface method.
func (proc *assetServer) Register(c *wsConnection) {
	envelope := assetServerEnvelopeRegister{proc, c, make(chan assetServerRegisterReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// RegisterTimeout is a gen_server interface method.
func (proc *assetServer) RegisterTimeout(c *wsConnection, timeout time.Duration) (gen_proc_err error) {
	envelope := assetServerEnvelopeRegister{proc, c, make(chan assetServerRegisterReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type assetServerEnvelopeUnregister struct {
	proc *assetServer
	c    *wsConnection
	ret  chan assetServerUnregisterReturn
	gen_proc.Envelope
}

func (msg assetServerEnvelopeUnregister) Call() {
	ret := make(chan assetServerUnregisterReturn, 1)
	go func(ret chan assetServerUnregisterReturn) {
		msg.proc.handleUnregister(msg.c)
		ret <- assetServerUnregisterReturn{}
	}(ret)
	select {
	case result := <-ret:
		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type assetServerUnregisterReturn struct {
}

// Unregister is a gen_server interface method.
func (proc *assetServer) Unregister(c *wsConnection) {
	envelope := assetServerEnvelopeUnregister{proc, c, make(chan assetServerUnregisterReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// UnregisterTimeout is a gen_server interface method.
func (proc *assetServer) UnregisterTimeout(c *wsConnection, timeout time.Duration) (gen_proc_err error) {
	envelope := assetServerEnvelopeUnregister{proc, c, make(chan assetServerUnregisterReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type assetServerEnvelopeBroadcast struct {
	proc *assetServer
	msg  interface{}
	ret  chan assetServerBroadcastReturn
	gen_proc.Envelope
}

func (msg assetServerEnvelopeBroadcast) Call() {
	ret := make(chan assetServerBroadcastReturn, 1)
	go func(ret chan assetServerBroadcastReturn) {
		msg.proc.handleBroadcast(msg.msg)
		ret <- assetServerBroadcastReturn{}
	}(ret)
	select {
	case result := <-ret:
		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type assetServerBroadcastReturn struct {
}

// Broadcast is a gen_server interface method.
func (proc *assetServer) Broadcast(msg interface{}) {
	envelope := assetServerEnvelopeBroadcast{proc, msg, make(chan assetServerBroadcastReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// BroadcastTimeout is a gen_server interface method.
func (proc *assetServer) BroadcastTimeout(msg interface{}, timeout time.Duration) (gen_proc_err error) {
	envelope := assetServerEnvelopeBroadcast{proc, msg, make(chan assetServerBroadcastReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

func (proc *assetServer) Run() {
	for {
		select {
		case msg := <-proc.chMsg:
			msg.Call()
		case <-proc.chStop:
			close(proc.chMsg)
			close(proc.chStop)
			return
		}
	}
}

func (proc *assetServer) Stop() {
	proc.chStop <- struct{}{}
}
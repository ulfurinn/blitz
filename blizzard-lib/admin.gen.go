//  GENERATED CODE, DO NOT EDIT
package blizzard

import (
	"sync/atomic"
	"time"
)
import "bitbucket.org/ulfurinn/gen_proc"

var assetServerProcCounter int32

type assetServerCh struct {
	chMsg           chan gen_proc.ProcCall
	retChGenCall    chan assetServerGenCallReturn
	retChRegister   chan assetServerRegisterReturn
	retChUnregister chan assetServerUnregisterReturn
	retChBroadcast  chan assetServerBroadcastReturn
	chStop          chan struct{}
}

func NewassetServerCh() *assetServerCh {
	return &assetServerCh{
		chMsg:  make(chan gen_proc.ProcCall),
		chStop: make(chan struct{}, 1),
	}
}

func (proc *assetServer) handleGenCall(f func() interface{}) interface{} {
	return f()
}

type assetServerEnvelopeGenCall struct {
	proc *assetServer
	f    func() interface{}
	ret  chan assetServerGenCallReturn
	gen_proc.Envelope
}

func (msg assetServerEnvelopeGenCall) Call() {
	ret := make(chan assetServerGenCallReturn, 1)
	msg.proc.retChGenCall = ret
	go func(ret chan assetServerGenCallReturn) {
		genret := msg.proc.handleGenCall(msg.f)
		msg.proc.retChGenCall = nil
		ret <- assetServerGenCallReturn{genret}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type assetServerGenCallReturn struct {
	genret interface{}
}

// GenCall is a gen_server interface method.
func (proc *assetServer) GenCall(f func() interface{}) (genret interface{}) {
	envelope := assetServerEnvelopeGenCall{proc, f, make(chan assetServerGenCallReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	genret = retval.genret

	return
}

// GenCallTimeout is a gen_server interface method.
func (proc *assetServer) GenCallTimeout(f func() interface{}, timeout time.Duration) (genret interface{}, gen_proc_err error) {
	envelope := assetServerEnvelopeGenCall{proc, f, make(chan assetServerGenCallReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	genret = retval.genret

	return
}

type assetServerEnvelopeRegister struct {
	proc *assetServer
	c    *wsConnection
	ret  chan assetServerRegisterReturn
	gen_proc.Envelope
}

func (msg assetServerEnvelopeRegister) Call() {
	ret := make(chan assetServerRegisterReturn, 1)
	msg.proc.retChRegister = ret
	go func(ret chan assetServerRegisterReturn) {
		msg.proc.handleRegister(msg.c)
		msg.proc.retChRegister = nil
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
	msg.proc.retChUnregister = ret
	go func(ret chan assetServerUnregisterReturn) {
		msg.proc.handleUnregister(msg.c)
		msg.proc.retChUnregister = nil
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
	msg.proc.retChBroadcast = ret
	go func(ret chan assetServerBroadcastReturn) {
		msg.proc.handleBroadcast(msg.msg)
		msg.proc.retChBroadcast = nil
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
	atomic.AddInt32(&assetServerProcCounter, 1)
	defer atomic.AddInt32(&assetServerProcCounter, -1)
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

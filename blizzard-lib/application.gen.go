//  GENERATED CODE, DO NOT EDIT
package blizzard

import (
	"sync/atomic"
	"time"
)
import "bitbucket.org/ulfurinn/gen_proc"

var ApplicationProcCounter int32

type ApplicationGen struct {
	chMsg          chan gen_proc.ProcCall
	retChGenCall   chan ApplicationGenCallReturn
	retChBootstrap chan ApplicationBootstrapReturn
	chStop         chan struct{}
}

func NewApplicationGen() *ApplicationGen {
	return &ApplicationGen{
		chMsg:  make(chan gen_proc.ProcCall),
		chStop: make(chan struct{}, 1),
	}
}

func (proc *Application) handleGenCall(f func() interface{}) interface{} {
	return f()
}

type ApplicationEnvelopeGenCall struct {
	proc *Application
	f    func() interface{}
	ret  chan ApplicationGenCallReturn
	gen_proc.Envelope
}

func (msg ApplicationEnvelopeGenCall) Call() {
	ret := make(chan ApplicationGenCallReturn, 1)
	msg.proc.retChGenCall = ret
	go func(ret chan ApplicationGenCallReturn) {
		genret := msg.proc.handleGenCall(msg.f)
		msg.proc.retChGenCall = nil
		ret <- ApplicationGenCallReturn{genret}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ApplicationGenCallReturn struct {
	genret interface{}
}

// GenCall is a gen_server interface method.
func (proc *Application) GenCall(f func() interface{}) (genret interface{}) {
	envelope := ApplicationEnvelopeGenCall{proc, f, make(chan ApplicationGenCallReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	genret = retval.genret

	return
}

// GenCallTimeout is a gen_server interface method.
func (proc *Application) GenCallTimeout(f func() interface{}, timeout time.Duration) (genret interface{}, gen_proc_err error) {
	envelope := ApplicationEnvelopeGenCall{proc, f, make(chan ApplicationGenCallReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	genret = retval.genret

	return
}

type ApplicationEnvelopeBootstrap struct {
	proc *Application

	ret chan ApplicationBootstrapReturn
	gen_proc.Envelope
}

func (msg ApplicationEnvelopeBootstrap) Call() {
	ret := make(chan ApplicationBootstrapReturn, 1)
	msg.proc.retChBootstrap = ret
	go func(ret chan ApplicationBootstrapReturn) {
		retval0, retval1 := msg.proc.handleBootstrap()
		msg.proc.retChBootstrap = nil
		ret <- ApplicationBootstrapReturn{retval0, retval1}
	}(ret)
	select {
	case result := <-ret:

		if result.retval0 {
			go func() { msg.ret <- result }()
			return
		}

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ApplicationBootstrapReturn struct {
	retval0 gen_proc.Deferred
	retval1 error
}

// Bootstrap is a gen_server interface method.
func (proc *Application) Bootstrap() (retval1 error) {
	envelope := ApplicationEnvelopeBootstrap{proc, make(chan ApplicationBootstrapReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval1 = retval.retval1

	return
}

// BootstrapTimeout is a gen_server interface method.
func (proc *Application) BootstrapTimeout(timeout time.Duration) (retval1 error, gen_proc_err error) {
	envelope := ApplicationEnvelopeBootstrap{proc, make(chan ApplicationBootstrapReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval1 = retval.retval1

	return
}

func (proc *Application) deferBootstrap(f func(func(retval1 error))) (retval0 gen_proc.Deferred, retval1 error) {
	retfun := func(ret chan ApplicationBootstrapReturn) func(retval1 error) {
		return func(retval1 error) {
			ret <- ApplicationBootstrapReturn{retval1: retval1}
		}
	}(proc.retChBootstrap)
	go f(retfun)
	retval0 = true
	return
}

func (proc *Application) Run() {
	atomic.AddInt32(&ApplicationProcCounter, 1)
	defer atomic.AddInt32(&ApplicationProcCounter, -1)
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

func (proc *Application) Stop() {
	proc.chStop <- struct{}{}
}

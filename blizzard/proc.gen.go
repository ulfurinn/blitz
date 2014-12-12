//  GENERATED CODE, DO NOT EDIT
package blizzard

import (
	"sync/atomic"
	"time"
)
import (
	"bitbucket.org/ulfurinn/blitz"
	"bitbucket.org/ulfurinn/gen_proc"
)

var ProcessProcCounter int32

type ProcessGen struct {
	chMsg               chan gen_proc.ProcCall
	retChGenCall        chan ProcessGenCallReturn
	retChExec           chan ProcessExecReturn
	retChShutdown       chan ProcessShutdownReturn
	retChCleanupProcess chan ProcessCleanupProcessReturn
	chStop              chan struct{}
}

func NewProcessGen() *ProcessGen {
	return &ProcessGen{
		chMsg:  make(chan gen_proc.ProcCall),
		chStop: make(chan struct{}, 1),
	}
}

func (proc *Process) handleGenCall(f func() interface{}) interface{} {
	return f()
}

type ProcessEnvelopeGenCall struct {
	proc *Process
	f    func() interface{}
	ret  chan ProcessGenCallReturn
	gen_proc.Envelope
}

func (msg ProcessEnvelopeGenCall) Call() {
	ret := make(chan ProcessGenCallReturn, 1)
	msg.proc.retChGenCall = ret
	go func(ret chan ProcessGenCallReturn) {
		genret := msg.proc.handleGenCall(msg.f)
		msg.proc.retChGenCall = nil
		ret <- ProcessGenCallReturn{genret}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result

	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcessGenCallReturn struct {
	genret interface{}
}

// GenCall is a gen_server interface method.
func (proc *Process) GenCall(f func() interface{}) (genret interface{}) {
	envelope := ProcessEnvelopeGenCall{proc, f, make(chan ProcessGenCallReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	genret = retval.genret

	return
}

// GenCallTimeout is a gen_server interface method.
func (proc *Process) GenCallTimeout(f func() interface{}, timeout time.Duration) (genret interface{}, gen_proc_err error) {
	envelope := ProcessEnvelopeGenCall{proc, f, make(chan ProcessGenCallReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	genret = retval.genret

	return
}

type ProcessEnvelopeExec struct {
	proc *Process

	ret chan ProcessExecReturn
	gen_proc.Envelope
}

func (msg ProcessEnvelopeExec) Call() {
	ret := make(chan ProcessExecReturn, 1)
	msg.proc.retChExec = ret
	go func(ret chan ProcessExecReturn) {
		retval0, retval1, retval2 := msg.proc.handleExec()
		msg.proc.retChExec = nil
		ret <- ProcessExecReturn{retval0, retval1, retval2}
	}(ret)
	select {
	case result := <-ret:

		if result.retval0 {
			go func() { msg.ret <- (<-ret) }()
		} else {
			msg.ret <- result
		}

	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcessExecReturn struct {
	retval0 gen_proc.Deferred
	retval1 *blitz.AnnounceCommand
	retval2 error
}

// Exec is a gen_server interface method.
func (proc *Process) Exec() (retval1 *blitz.AnnounceCommand, retval2 error) {
	envelope := ProcessEnvelopeExec{proc, make(chan ProcessExecReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval1 = retval.retval1
	retval2 = retval.retval2

	return
}

// ExecTimeout is a gen_server interface method.
func (proc *Process) ExecTimeout(timeout time.Duration) (retval1 *blitz.AnnounceCommand, retval2 error, gen_proc_err error) {
	envelope := ProcessEnvelopeExec{proc, make(chan ProcessExecReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval1 = retval.retval1
	retval2 = retval.retval2

	return
}

func (proc *Process) deferExec(f func(func(retval1 *blitz.AnnounceCommand, retval2 error))) (retval0 gen_proc.Deferred, retval1 *blitz.AnnounceCommand, retval2 error) {
	retfun := func(ret chan ProcessExecReturn) func(retval1 *blitz.AnnounceCommand, retval2 error) {
		return func(retval1 *blitz.AnnounceCommand, retval2 error) {
			ret <- ProcessExecReturn{retval1: retval1, retval2: retval2}
		}
	}(proc.retChExec)
	go f(retfun)
	retval0 = true
	return
}

type ProcessEnvelopeShutdown struct {
	proc *Process
	kill bool
	ret  chan ProcessShutdownReturn
	gen_proc.Envelope
}

func (msg ProcessEnvelopeShutdown) Call() {
	ret := make(chan ProcessShutdownReturn, 1)
	msg.proc.retChShutdown = ret
	go func(ret chan ProcessShutdownReturn) {
		msg.proc.handleShutdown(msg.kill)
		msg.proc.retChShutdown = nil
		ret <- ProcessShutdownReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result

	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcessShutdownReturn struct {
}

// Shutdown is a gen_server interface method.
func (proc *Process) Shutdown(kill bool) {
	envelope := ProcessEnvelopeShutdown{proc, kill, make(chan ProcessShutdownReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// ShutdownTimeout is a gen_server interface method.
func (proc *Process) ShutdownTimeout(kill bool, timeout time.Duration) (gen_proc_err error) {
	envelope := ProcessEnvelopeShutdown{proc, kill, make(chan ProcessShutdownReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type ProcessEnvelopeCleanupProcess struct {
	proc *Process

	ret chan ProcessCleanupProcessReturn
	gen_proc.Envelope
}

func (msg ProcessEnvelopeCleanupProcess) Call() {
	ret := make(chan ProcessCleanupProcessReturn, 1)
	msg.proc.retChCleanupProcess = ret
	go func(ret chan ProcessCleanupProcessReturn) {
		msg.proc.handleCleanupProcess()
		msg.proc.retChCleanupProcess = nil
		ret <- ProcessCleanupProcessReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result

	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcessCleanupProcessReturn struct {
}

// CleanupProcess is a gen_server interface method.
func (proc *Process) CleanupProcess() {
	envelope := ProcessEnvelopeCleanupProcess{proc, make(chan ProcessCleanupProcessReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// CleanupProcessTimeout is a gen_server interface method.
func (proc *Process) CleanupProcessTimeout(timeout time.Duration) (gen_proc_err error) {
	envelope := ProcessEnvelopeCleanupProcess{proc, make(chan ProcessCleanupProcessReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

func (proc *Process) Run() {
	atomic.AddInt32(&ProcessProcCounter, 1)
	defer atomic.AddInt32(&ProcessProcCounter, -1)
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

func (proc *Process) Stop() {
	proc.chStop <- struct{}{}
}

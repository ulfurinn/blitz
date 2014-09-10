//  GENERATED CODE, DO NOT EDIT
package blizzard

import "time"
import (
	"bitbucket.org/ulfurinn/blitz"
	"bitbucket.org/ulfurinn/gen_proc"
)

type ProcessGen struct {
	chMsg  chan gen_proc.ProcCall
	chStop chan struct{}
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
	go func(ret chan ProcessGenCallReturn) {
		genret := msg.proc.handleGenCall(msg.f)
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

type ProcessEnvelopeAnnounced struct {
	proc *Process
	cmd  *blitz.AnnounceCommand
	w    *WorkerConnection
	ret  chan ProcessAnnouncedReturn
	gen_proc.Envelope
}

func (msg ProcessEnvelopeAnnounced) Call() {
	ret := make(chan ProcessAnnouncedReturn, 1)
	go func(ret chan ProcessAnnouncedReturn) {
		msg.proc.handleAnnounced(msg.cmd, msg.w)
		ret <- ProcessAnnouncedReturn{}
	}(ret)
	select {
	case result := <-ret:
		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcessAnnouncedReturn struct {
}

// Announced is a gen_server interface method.
func (proc *Process) Announced(cmd *blitz.AnnounceCommand, w *WorkerConnection) {
	envelope := ProcessEnvelopeAnnounced{proc, cmd, w, make(chan ProcessAnnouncedReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// AnnouncedTimeout is a gen_server interface method.
func (proc *Process) AnnouncedTimeout(cmd *blitz.AnnounceCommand, w *WorkerConnection, timeout time.Duration) (gen_proc_err error) {
	envelope := ProcessEnvelopeAnnounced{proc, cmd, w, make(chan ProcessAnnouncedReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type ProcessEnvelopeShutdown struct {
	proc *Process

	ret chan ProcessShutdownReturn
	gen_proc.Envelope
}

func (msg ProcessEnvelopeShutdown) Call() {
	ret := make(chan ProcessShutdownReturn, 1)
	go func(ret chan ProcessShutdownReturn) {
		msg.proc.handleShutdown()
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
func (proc *Process) Shutdown() {
	envelope := ProcessEnvelopeShutdown{proc, make(chan ProcessShutdownReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// ShutdownTimeout is a gen_server interface method.
func (proc *Process) ShutdownTimeout(timeout time.Duration) (gen_proc_err error) {
	envelope := ProcessEnvelopeShutdown{proc, make(chan ProcessShutdownReturn, 1), gen_proc.Envelope{timeout}}
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
	go func(ret chan ProcessCleanupProcessReturn) {
		msg.proc.handleCleanupProcess()
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

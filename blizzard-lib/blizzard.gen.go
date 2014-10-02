//  GENERATED CODE, DO NOT EDIT
package blizzard

import "time"
import "bitbucket.org/ulfurinn/gen_proc"

type BlizzardCh struct {
	chMsg             chan gen_proc.ProcCall
	retChGenCall      chan BlizzardGenCallReturn
	retChCommand      chan BlizzardCommandReturn
	retChCleanup      chan BlizzardCleanupReturn
	retChWorkerClosed chan BlizzardWorkerClosedReturn
	retChSnapshot     chan BlizzardSnapshotReturn
	chStop            chan struct{}
}

func NewBlizzardCh() *BlizzardCh {
	return &BlizzardCh{
		chMsg:  make(chan gen_proc.ProcCall),
		chStop: make(chan struct{}, 1),
	}
}

func (proc *Blizzard) handleGenCall(f func() interface{}) interface{} {
	return f()
}

type BlizzardEnvelopeGenCall struct {
	proc *Blizzard
	f    func() interface{}
	ret  chan BlizzardGenCallReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeGenCall) Call() {
	ret := make(chan BlizzardGenCallReturn, 1)
	msg.proc.retChGenCall = ret
	go func(ret chan BlizzardGenCallReturn) {
		genret := msg.proc.handleGenCall(msg.f)
		msg.proc.retChGenCall = nil
		ret <- BlizzardGenCallReturn{genret}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardGenCallReturn struct {
	genret interface{}
}

// GenCall is a gen_server interface method.
func (proc *Blizzard) GenCall(f func() interface{}) (genret interface{}) {
	envelope := BlizzardEnvelopeGenCall{proc, f, make(chan BlizzardGenCallReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	genret = retval.genret

	return
}

// GenCallTimeout is a gen_server interface method.
func (proc *Blizzard) GenCallTimeout(f func() interface{}, timeout time.Duration) (genret interface{}, gen_proc_err error) {
	envelope := BlizzardEnvelopeGenCall{proc, f, make(chan BlizzardGenCallReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	genret = retval.genret

	return
}

type BlizzardEnvelopeCommand struct {
	proc *Blizzard
	cmd  workerCommand
	ret  chan BlizzardCommandReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeCommand) Call() {
	ret := make(chan BlizzardCommandReturn, 1)
	msg.proc.retChCommand = ret
	go func(ret chan BlizzardCommandReturn) {
		retval0 := msg.proc.handleCommand(msg.cmd)
		msg.proc.retChCommand = nil
		ret <- BlizzardCommandReturn{retval0}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardCommandReturn struct {
	retval0 interface{}
}

// Command is a gen_server interface method.
func (proc *Blizzard) Command(cmd workerCommand) (retval0 interface{}) {
	envelope := BlizzardEnvelopeCommand{proc, cmd, make(chan BlizzardCommandReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval0 = retval.retval0

	return
}

// CommandTimeout is a gen_server interface method.
func (proc *Blizzard) CommandTimeout(cmd workerCommand, timeout time.Duration) (retval0 interface{}, gen_proc_err error) {
	envelope := BlizzardEnvelopeCommand{proc, cmd, make(chan BlizzardCommandReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval0 = retval.retval0

	return
}

type BlizzardEnvelopeCleanup struct {
	proc *Blizzard

	ret chan BlizzardCleanupReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeCleanup) Call() {
	ret := make(chan BlizzardCleanupReturn, 1)
	msg.proc.retChCleanup = ret
	go func(ret chan BlizzardCleanupReturn) {
		msg.proc.handleCleanup()
		msg.proc.retChCleanup = nil
		ret <- BlizzardCleanupReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardCleanupReturn struct {
}

// Cleanup is a gen_server interface method.
func (proc *Blizzard) Cleanup() {
	envelope := BlizzardEnvelopeCleanup{proc, make(chan BlizzardCleanupReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// CleanupTimeout is a gen_server interface method.
func (proc *Blizzard) CleanupTimeout(timeout time.Duration) (gen_proc_err error) {
	envelope := BlizzardEnvelopeCleanup{proc, make(chan BlizzardCleanupReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type BlizzardEnvelopeWorkerClosed struct {
	proc *Blizzard
	w    *WorkerConnection
	ret  chan BlizzardWorkerClosedReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeWorkerClosed) Call() {
	ret := make(chan BlizzardWorkerClosedReturn, 1)
	msg.proc.retChWorkerClosed = ret
	go func(ret chan BlizzardWorkerClosedReturn) {
		msg.proc.handleWorkerClosed(msg.w)
		msg.proc.retChWorkerClosed = nil
		ret <- BlizzardWorkerClosedReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardWorkerClosedReturn struct {
}

// WorkerClosed is a gen_server interface method.
func (proc *Blizzard) WorkerClosed(w *WorkerConnection) {
	envelope := BlizzardEnvelopeWorkerClosed{proc, w, make(chan BlizzardWorkerClosedReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// WorkerClosedTimeout is a gen_server interface method.
func (proc *Blizzard) WorkerClosedTimeout(w *WorkerConnection, timeout time.Duration) (gen_proc_err error) {
	envelope := BlizzardEnvelopeWorkerClosed{proc, w, make(chan BlizzardWorkerClosedReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type BlizzardEnvelopeSnapshot struct {
	proc *Blizzard
	f    func(interface{})
	ret  chan BlizzardSnapshotReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeSnapshot) Call() {
	ret := make(chan BlizzardSnapshotReturn, 1)
	msg.proc.retChSnapshot = ret
	go func(ret chan BlizzardSnapshotReturn) {
		msg.proc.handleSnapshot(msg.f)
		msg.proc.retChSnapshot = nil
		ret <- BlizzardSnapshotReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardSnapshotReturn struct {
}

// Snapshot is a gen_server interface method.
func (proc *Blizzard) Snapshot(f func(interface{})) {
	envelope := BlizzardEnvelopeSnapshot{proc, f, make(chan BlizzardSnapshotReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// SnapshotTimeout is a gen_server interface method.
func (proc *Blizzard) SnapshotTimeout(f func(interface{}), timeout time.Duration) (gen_proc_err error) {
	envelope := BlizzardEnvelopeSnapshot{proc, f, make(chan BlizzardSnapshotReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

func (proc *Blizzard) Run() {
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

func (proc *Blizzard) Stop() {
	proc.chStop <- struct{}{}
}

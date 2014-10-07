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

var ProcGroupProcCounter int32

type ProcGroupGen struct {
	chMsg                     chan gen_proc.ProcCall
	retChGenCall              chan ProcGroupGenCallReturn
	retChIsReady              chan ProcGroupIsReadyReturn
	retChAdd                  chan ProcGroupAddReturn
	retChAnnounced            chan ProcGroupAnnouncedReturn
	retChRemoveProc           chan ProcGroupRemoveProcReturn
	retChFindProcByConnection chan ProcGroupFindProcByConnectionReturn
	retChGet                  chan ProcGroupGetReturn
	retChGetForRemoval        chan ProcGroupGetForRemovalReturn
	retChGetAll               chan ProcGroupGetAllReturn
	retChSize                 chan ProcGroupSizeReturn
	retChShutdown             chan ProcGroupShutdownReturn
	chStop                    chan struct{}
}

func NewProcGroupGen() *ProcGroupGen {
	return &ProcGroupGen{
		chMsg:  make(chan gen_proc.ProcCall),
		chStop: make(chan struct{}, 1),
	}
}

func (proc *ProcGroup) handleGenCall(f func() interface{}) interface{} {
	return f()
}

type ProcGroupEnvelopeGenCall struct {
	proc *ProcGroup
	f    func() interface{}
	ret  chan ProcGroupGenCallReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeGenCall) Call() {
	ret := make(chan ProcGroupGenCallReturn, 1)
	msg.proc.retChGenCall = ret
	go func(ret chan ProcGroupGenCallReturn) {
		genret := msg.proc.handleGenCall(msg.f)
		msg.proc.retChGenCall = nil
		ret <- ProcGroupGenCallReturn{genret}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupGenCallReturn struct {
	genret interface{}
}

// GenCall is a gen_server interface method.
func (proc *ProcGroup) GenCall(f func() interface{}) (genret interface{}) {
	envelope := ProcGroupEnvelopeGenCall{proc, f, make(chan ProcGroupGenCallReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	genret = retval.genret

	return
}

// GenCallTimeout is a gen_server interface method.
func (proc *ProcGroup) GenCallTimeout(f func() interface{}, timeout time.Duration) (genret interface{}, gen_proc_err error) {
	envelope := ProcGroupEnvelopeGenCall{proc, f, make(chan ProcGroupGenCallReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	genret = retval.genret

	return
}

type ProcGroupEnvelopeIsReady struct {
	proc *ProcGroup

	ret chan ProcGroupIsReadyReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeIsReady) Call() {
	ret := make(chan ProcGroupIsReadyReturn, 1)
	msg.proc.retChIsReady = ret
	go func(ret chan ProcGroupIsReadyReturn) {
		retval0 := msg.proc.handleIsReady()
		msg.proc.retChIsReady = nil
		ret <- ProcGroupIsReadyReturn{retval0}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupIsReadyReturn struct {
	retval0 bool
}

// IsReady is a gen_server interface method.
func (proc *ProcGroup) IsReady() (retval0 bool) {
	envelope := ProcGroupEnvelopeIsReady{proc, make(chan ProcGroupIsReadyReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval0 = retval.retval0

	return
}

// IsReadyTimeout is a gen_server interface method.
func (proc *ProcGroup) IsReadyTimeout(timeout time.Duration) (retval0 bool, gen_proc_err error) {
	envelope := ProcGroupEnvelopeIsReady{proc, make(chan ProcGroupIsReadyReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval0 = retval.retval0

	return
}

type ProcGroupEnvelopeAdd struct {
	proc *ProcGroup
	p    *Process
	ret  chan ProcGroupAddReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeAdd) Call() {
	ret := make(chan ProcGroupAddReturn, 1)
	msg.proc.retChAdd = ret
	go func(ret chan ProcGroupAddReturn) {
		msg.proc.handleAdd(msg.p)
		msg.proc.retChAdd = nil
		ret <- ProcGroupAddReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupAddReturn struct {
}

// Add is a gen_server interface method.
func (proc *ProcGroup) Add(p *Process) {
	envelope := ProcGroupEnvelopeAdd{proc, p, make(chan ProcGroupAddReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// AddTimeout is a gen_server interface method.
func (proc *ProcGroup) AddTimeout(p *Process, timeout time.Duration) (gen_proc_err error) {
	envelope := ProcGroupEnvelopeAdd{proc, p, make(chan ProcGroupAddReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type ProcGroupEnvelopeAnnounced struct {
	proc *ProcGroup
	p    *Process
	cmd  *blitz.AnnounceCommand
	w    *WorkerConnection
	ret  chan ProcGroupAnnouncedReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeAnnounced) Call() {
	ret := make(chan ProcGroupAnnouncedReturn, 1)
	msg.proc.retChAnnounced = ret
	go func(ret chan ProcGroupAnnouncedReturn) {
		ok, first := msg.proc.handleAnnounced(msg.p, msg.cmd, msg.w)
		msg.proc.retChAnnounced = nil
		ret <- ProcGroupAnnouncedReturn{ok, first}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupAnnouncedReturn struct {
	ok    bool
	first bool
}

// Announced is a gen_server interface method.
func (proc *ProcGroup) Announced(p *Process, cmd *blitz.AnnounceCommand, w *WorkerConnection) (ok bool, first bool) {
	envelope := ProcGroupEnvelopeAnnounced{proc, p, cmd, w, make(chan ProcGroupAnnouncedReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	ok = retval.ok
	first = retval.first

	return
}

// AnnouncedTimeout is a gen_server interface method.
func (proc *ProcGroup) AnnouncedTimeout(p *Process, cmd *blitz.AnnounceCommand, w *WorkerConnection, timeout time.Duration) (ok bool, first bool, gen_proc_err error) {
	envelope := ProcGroupEnvelopeAnnounced{proc, p, cmd, w, make(chan ProcGroupAnnouncedReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	ok = retval.ok
	first = retval.first

	return
}

type ProcGroupEnvelopeRemoveProc struct {
	proc *ProcGroup
	p    *Process
	ret  chan ProcGroupRemoveProcReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeRemoveProc) Call() {
	ret := make(chan ProcGroupRemoveProcReturn, 1)
	msg.proc.retChRemoveProc = ret
	go func(ret chan ProcGroupRemoveProcReturn) {
		found := msg.proc.handleRemoveProc(msg.p)
		msg.proc.retChRemoveProc = nil
		ret <- ProcGroupRemoveProcReturn{found}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupRemoveProcReturn struct {
	found bool
}

// RemoveProc is a gen_server interface method.
func (proc *ProcGroup) RemoveProc(p *Process) (found bool) {
	envelope := ProcGroupEnvelopeRemoveProc{proc, p, make(chan ProcGroupRemoveProcReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	found = retval.found

	return
}

// RemoveProcTimeout is a gen_server interface method.
func (proc *ProcGroup) RemoveProcTimeout(p *Process, timeout time.Duration) (found bool, gen_proc_err error) {
	envelope := ProcGroupEnvelopeRemoveProc{proc, p, make(chan ProcGroupRemoveProcReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	found = retval.found

	return
}

type ProcGroupEnvelopeFindProcByConnection struct {
	proc *ProcGroup
	w    *WorkerConnection
	ret  chan ProcGroupFindProcByConnectionReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeFindProcByConnection) Call() {
	ret := make(chan ProcGroupFindProcByConnectionReturn, 1)
	msg.proc.retChFindProcByConnection = ret
	go func(ret chan ProcGroupFindProcByConnectionReturn) {
		retval0 := msg.proc.handleFindProcByConnection(msg.w)
		msg.proc.retChFindProcByConnection = nil
		ret <- ProcGroupFindProcByConnectionReturn{retval0}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupFindProcByConnectionReturn struct {
	retval0 *Process
}

// FindProcByConnection is a gen_server interface method.
func (proc *ProcGroup) FindProcByConnection(w *WorkerConnection) (retval0 *Process) {
	envelope := ProcGroupEnvelopeFindProcByConnection{proc, w, make(chan ProcGroupFindProcByConnectionReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval0 = retval.retval0

	return
}

// FindProcByConnectionTimeout is a gen_server interface method.
func (proc *ProcGroup) FindProcByConnectionTimeout(w *WorkerConnection, timeout time.Duration) (retval0 *Process, gen_proc_err error) {
	envelope := ProcGroupEnvelopeFindProcByConnection{proc, w, make(chan ProcGroupFindProcByConnectionReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval0 = retval.retval0

	return
}

type ProcGroupEnvelopeGet struct {
	proc *ProcGroup

	ret chan ProcGroupGetReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeGet) Call() {
	ret := make(chan ProcGroupGetReturn, 1)
	msg.proc.retChGet = ret
	go func(ret chan ProcGroupGetReturn) {
		retval0 := msg.proc.handleGet()
		msg.proc.retChGet = nil
		ret <- ProcGroupGetReturn{retval0}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupGetReturn struct {
	retval0 *Process
}

// Get is a gen_server interface method.
func (proc *ProcGroup) Get() (retval0 *Process) {
	envelope := ProcGroupEnvelopeGet{proc, make(chan ProcGroupGetReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval0 = retval.retval0

	return
}

// GetTimeout is a gen_server interface method.
func (proc *ProcGroup) GetTimeout(timeout time.Duration) (retval0 *Process, gen_proc_err error) {
	envelope := ProcGroupEnvelopeGet{proc, make(chan ProcGroupGetReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval0 = retval.retval0

	return
}

type ProcGroupEnvelopeGetForRemoval struct {
	proc *ProcGroup

	ret chan ProcGroupGetForRemovalReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeGetForRemoval) Call() {
	ret := make(chan ProcGroupGetForRemovalReturn, 1)
	msg.proc.retChGetForRemoval = ret
	go func(ret chan ProcGroupGetForRemovalReturn) {
		retval0 := msg.proc.handleGetForRemoval()
		msg.proc.retChGetForRemoval = nil
		ret <- ProcGroupGetForRemovalReturn{retval0}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupGetForRemovalReturn struct {
	retval0 *Process
}

// GetForRemoval is a gen_server interface method.
func (proc *ProcGroup) GetForRemoval() (retval0 *Process) {
	envelope := ProcGroupEnvelopeGetForRemoval{proc, make(chan ProcGroupGetForRemovalReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval0 = retval.retval0

	return
}

// GetForRemovalTimeout is a gen_server interface method.
func (proc *ProcGroup) GetForRemovalTimeout(timeout time.Duration) (retval0 *Process, gen_proc_err error) {
	envelope := ProcGroupEnvelopeGetForRemoval{proc, make(chan ProcGroupGetForRemovalReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval0 = retval.retval0

	return
}

type ProcGroupEnvelopeGetAll struct {
	proc *ProcGroup

	ret chan ProcGroupGetAllReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeGetAll) Call() {
	ret := make(chan ProcGroupGetAllReturn, 1)
	msg.proc.retChGetAll = ret
	go func(ret chan ProcGroupGetAllReturn) {
		retval0 := msg.proc.handleGetAll()
		msg.proc.retChGetAll = nil
		ret <- ProcGroupGetAllReturn{retval0}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupGetAllReturn struct {
	retval0 []*Process
}

// GetAll is a gen_server interface method.
func (proc *ProcGroup) GetAll() (retval0 []*Process) {
	envelope := ProcGroupEnvelopeGetAll{proc, make(chan ProcGroupGetAllReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval0 = retval.retval0

	return
}

// GetAllTimeout is a gen_server interface method.
func (proc *ProcGroup) GetAllTimeout(timeout time.Duration) (retval0 []*Process, gen_proc_err error) {
	envelope := ProcGroupEnvelopeGetAll{proc, make(chan ProcGroupGetAllReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval0 = retval.retval0

	return
}

type ProcGroupEnvelopeSize struct {
	proc *ProcGroup

	ret chan ProcGroupSizeReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeSize) Call() {
	ret := make(chan ProcGroupSizeReturn, 1)
	msg.proc.retChSize = ret
	go func(ret chan ProcGroupSizeReturn) {
		retval0 := msg.proc.handleSize()
		msg.proc.retChSize = nil
		ret <- ProcGroupSizeReturn{retval0}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupSizeReturn struct {
	retval0 int
}

// Size is a gen_server interface method.
func (proc *ProcGroup) Size() (retval0 int) {
	envelope := ProcGroupEnvelopeSize{proc, make(chan ProcGroupSizeReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval0 = retval.retval0

	return
}

// SizeTimeout is a gen_server interface method.
func (proc *ProcGroup) SizeTimeout(timeout time.Duration) (retval0 int, gen_proc_err error) {
	envelope := ProcGroupEnvelopeSize{proc, make(chan ProcGroupSizeReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval0 = retval.retval0

	return
}

type ProcGroupEnvelopeShutdown struct {
	proc *ProcGroup

	ret chan ProcGroupShutdownReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeShutdown) Call() {
	ret := make(chan ProcGroupShutdownReturn, 1)
	msg.proc.retChShutdown = ret
	go func(ret chan ProcGroupShutdownReturn) {
		msg.proc.handleShutdown()
		msg.proc.retChShutdown = nil
		ret <- ProcGroupShutdownReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupShutdownReturn struct {
}

// Shutdown is a gen_server interface method.
func (proc *ProcGroup) Shutdown() {
	envelope := ProcGroupEnvelopeShutdown{proc, make(chan ProcGroupShutdownReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// ShutdownTimeout is a gen_server interface method.
func (proc *ProcGroup) ShutdownTimeout(timeout time.Duration) (gen_proc_err error) {
	envelope := ProcGroupEnvelopeShutdown{proc, make(chan ProcGroupShutdownReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

func (proc *ProcGroup) Run() {
	atomic.AddInt32(&ProcGroupProcCounter, 1)
	defer atomic.AddInt32(&ProcGroupProcCounter, -1)
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

func (proc *ProcGroup) Stop() {
	proc.chStop <- struct{}{}
}

//  GENERATED CODE, DO NOT EDIT
package blizzard

import (
	"sync/atomic"
	"time"
)
import "bitbucket.org/ulfurinn/gen_proc"

var ProcGroupProcCounter int32

type ProcGroupGen struct {
	chMsg                     chan gen_proc.ProcCall
	retChGenCall              chan ProcGroupGenCallReturn
	retChSpawn                chan ProcGroupSpawnReturn
	retChIsReady              chan ProcGroupIsReadyReturn
	retChAddPending           chan ProcGroupAddPendingReturn
	retChPendingSpawned       chan ProcGroupPendingSpawnedReturn
	retChWorkerClosed         chan ProcGroupWorkerClosedReturn
	retChRemoveDeadProc       chan ProcGroupRemoveDeadProcReturn
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

type ProcGroupEnvelopeSpawn struct {
	proc *ProcGroup

	ret chan ProcGroupSpawnReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeSpawn) Call() {
	ret := make(chan ProcGroupSpawnReturn, 1)
	msg.proc.retChSpawn = ret
	go func(ret chan ProcGroupSpawnReturn) {
		retval0, retval1 := msg.proc.handleSpawn()
		msg.proc.retChSpawn = nil
		ret <- ProcGroupSpawnReturn{retval0, retval1}
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

type ProcGroupSpawnReturn struct {
	retval0 gen_proc.Deferred
	retval1 error
}

// Spawn is a gen_server interface method.
func (proc *ProcGroup) Spawn() (retval1 error) {
	envelope := ProcGroupEnvelopeSpawn{proc, make(chan ProcGroupSpawnReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval1 = retval.retval1

	return
}

// SpawnTimeout is a gen_server interface method.
func (proc *ProcGroup) SpawnTimeout(timeout time.Duration) (retval1 error, gen_proc_err error) {
	envelope := ProcGroupEnvelopeSpawn{proc, make(chan ProcGroupSpawnReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval1 = retval.retval1

	return
}

func (proc *ProcGroup) deferSpawn(f func(func(retval1 error))) (retval0 gen_proc.Deferred, retval1 error) {
	retfun := func(ret chan ProcGroupSpawnReturn) func(retval1 error) {
		return func(retval1 error) {
			ret <- ProcGroupSpawnReturn{retval1: retval1}
		}
	}(proc.retChSpawn)
	go f(retfun)
	retval0 = true
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

type ProcGroupEnvelopeAddPending struct {
	proc *ProcGroup
	p    *Process
	ret  chan ProcGroupAddPendingReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeAddPending) Call() {
	ret := make(chan ProcGroupAddPendingReturn, 1)
	msg.proc.retChAddPending = ret
	go func(ret chan ProcGroupAddPendingReturn) {
		msg.proc.handleAddPending(msg.p)
		msg.proc.retChAddPending = nil
		ret <- ProcGroupAddPendingReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result

	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupAddPendingReturn struct {
}

// AddPending is a gen_server interface method.
func (proc *ProcGroup) AddPending(p *Process) {
	envelope := ProcGroupEnvelopeAddPending{proc, p, make(chan ProcGroupAddPendingReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// AddPendingTimeout is a gen_server interface method.
func (proc *ProcGroup) AddPendingTimeout(p *Process, timeout time.Duration) (gen_proc_err error) {
	envelope := ProcGroupEnvelopeAddPending{proc, p, make(chan ProcGroupAddPendingReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type ProcGroupEnvelopePendingSpawned struct {
	proc    *ProcGroup
	p       *Process
	spawnOk bool
	ret     chan ProcGroupPendingSpawnedReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopePendingSpawned) Call() {
	ret := make(chan ProcGroupPendingSpawnedReturn, 1)
	msg.proc.retChPendingSpawned = ret
	go func(ret chan ProcGroupPendingSpawnedReturn) {
		msg.proc.handlePendingSpawned(msg.p, msg.spawnOk)
		msg.proc.retChPendingSpawned = nil
		ret <- ProcGroupPendingSpawnedReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result

	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupPendingSpawnedReturn struct {
}

// PendingSpawned is a gen_server interface method.
func (proc *ProcGroup) PendingSpawned(p *Process, spawnOk bool) {
	envelope := ProcGroupEnvelopePendingSpawned{proc, p, spawnOk, make(chan ProcGroupPendingSpawnedReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// PendingSpawnedTimeout is a gen_server interface method.
func (proc *ProcGroup) PendingSpawnedTimeout(p *Process, spawnOk bool, timeout time.Duration) (gen_proc_err error) {
	envelope := ProcGroupEnvelopePendingSpawned{proc, p, spawnOk, make(chan ProcGroupPendingSpawnedReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type ProcGroupEnvelopeWorkerClosed struct {
	proc *ProcGroup
	w    *WorkerConnection
	ret  chan ProcGroupWorkerClosedReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeWorkerClosed) Call() {
	ret := make(chan ProcGroupWorkerClosedReturn, 1)
	msg.proc.retChWorkerClosed = ret
	go func(ret chan ProcGroupWorkerClosedReturn) {
		msg.proc.handleWorkerClosed(msg.w)
		msg.proc.retChWorkerClosed = nil
		ret <- ProcGroupWorkerClosedReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result

	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupWorkerClosedReturn struct {
}

// WorkerClosed is a gen_server interface method.
func (proc *ProcGroup) WorkerClosed(w *WorkerConnection) {
	envelope := ProcGroupEnvelopeWorkerClosed{proc, w, make(chan ProcGroupWorkerClosedReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// WorkerClosedTimeout is a gen_server interface method.
func (proc *ProcGroup) WorkerClosedTimeout(w *WorkerConnection, timeout time.Duration) (gen_proc_err error) {
	envelope := ProcGroupEnvelopeWorkerClosed{proc, w, make(chan ProcGroupWorkerClosedReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type ProcGroupEnvelopeRemoveDeadProc struct {
	proc *ProcGroup
	p    *Process
	ret  chan ProcGroupRemoveDeadProcReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeRemoveDeadProc) Call() {
	ret := make(chan ProcGroupRemoveDeadProcReturn, 1)
	msg.proc.retChRemoveDeadProc = ret
	go func(ret chan ProcGroupRemoveDeadProcReturn) {
		found := msg.proc.handleRemoveDeadProc(msg.p)
		msg.proc.retChRemoveDeadProc = nil
		ret <- ProcGroupRemoveDeadProcReturn{found}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result

	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupRemoveDeadProcReturn struct {
	found bool
}

// RemoveDeadProc is a gen_server interface method.
func (proc *ProcGroup) RemoveDeadProc(p *Process) (found bool) {
	envelope := ProcGroupEnvelopeRemoveDeadProc{proc, p, make(chan ProcGroupRemoveDeadProcReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	found = retval.found

	return
}

// RemoveDeadProcTimeout is a gen_server interface method.
func (proc *ProcGroup) RemoveDeadProcTimeout(p *Process, timeout time.Duration) (found bool, gen_proc_err error) {
	envelope := ProcGroupEnvelopeRemoveDeadProc{proc, p, make(chan ProcGroupRemoveDeadProcReturn, 1), gen_proc.Envelope{timeout}}
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

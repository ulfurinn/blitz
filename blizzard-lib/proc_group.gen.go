//  GENERATED CODE, DO NOT EDIT
package blizzard

import "time"
import (
	"bitbucket.org/ulfurinn/blitz"
	"bitbucket.org/ulfurinn/gen_proc"
)

type ProcGroupGen struct {
	chMsg  chan gen_proc.ProcCall
	chStop chan struct{}
}

func NewProcGroupGen() *ProcGroupGen {
	return &ProcGroupGen{
		chMsg:  make(chan gen_proc.ProcCall),
		chStop: make(chan struct{}, 1),
	}
}

type ProcGroupEnvelopeSpawn struct {
	proc  *ProcGroup
	count int
	cb    SpawnedCallback
	ret   chan ProcGroupSpawnReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeSpawn) Call() {
	ret := make(chan ProcGroupSpawnReturn, 1)
	go func(ret chan ProcGroupSpawnReturn) {
		err := msg.proc.handleSpawn(msg.count, msg.cb)
		ret <- ProcGroupSpawnReturn{err}
	}(ret)
	select {
	case result := <-ret:
		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupSpawnReturn struct {
	err error
}

// Spawn is a gen_server interface method.
func (proc *ProcGroup) Spawn(count int, cb SpawnedCallback) (err error) {
	envelope := ProcGroupEnvelopeSpawn{proc, count, cb, make(chan ProcGroupSpawnReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	err = retval.err

	return
}

// SpawnTimeout is a gen_server interface method.
func (proc *ProcGroup) SpawnTimeout(count int, cb SpawnedCallback, timeout time.Duration) (err error, gen_proc_err error) {
	envelope := ProcGroupEnvelopeSpawn{proc, count, cb, make(chan ProcGroupSpawnReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	err = retval.err

	return
}

type ProcGroupEnvelopeIsReady struct {
	proc *ProcGroup

	ret chan ProcGroupIsReadyReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeIsReady) Call() {
	ret := make(chan ProcGroupIsReadyReturn, 1)
	go func(ret chan ProcGroupIsReadyReturn) {
		retval0 := msg.proc.handleIsReady()
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
	go func(ret chan ProcGroupAddReturn) {
		msg.proc.handleAdd(msg.p)
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
	cmd  blitz.AnnounceCommand
	w    *WorkerConnection
	ret  chan ProcGroupAnnouncedReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeAnnounced) Call() {
	ret := make(chan ProcGroupAnnouncedReturn, 1)
	go func(ret chan ProcGroupAnnouncedReturn) {
		ok, first := msg.proc.handleAnnounced(msg.p, msg.cmd, msg.w)
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
func (proc *ProcGroup) Announced(p *Process, cmd blitz.AnnounceCommand, w *WorkerConnection) (ok bool, first bool) {
	envelope := ProcGroupEnvelopeAnnounced{proc, p, cmd, w, make(chan ProcGroupAnnouncedReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	ok = retval.ok
	first = retval.first

	return
}

// AnnouncedTimeout is a gen_server interface method.
func (proc *ProcGroup) AnnouncedTimeout(p *Process, cmd blitz.AnnounceCommand, w *WorkerConnection, timeout time.Duration) (ok bool, first bool, gen_proc_err error) {
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

type ProcGroupEnvelopeRemove struct {
	proc *ProcGroup
	p    *Process
	ret  chan ProcGroupRemoveReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeRemove) Call() {
	ret := make(chan ProcGroupRemoveReturn, 1)
	go func(ret chan ProcGroupRemoveReturn) {
		found := msg.proc.handleRemove(msg.p)
		ret <- ProcGroupRemoveReturn{found}
	}(ret)
	select {
	case result := <-ret:
		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type ProcGroupRemoveReturn struct {
	found bool
}

// Remove is a gen_server interface method.
func (proc *ProcGroup) Remove(p *Process) (found bool) {
	envelope := ProcGroupEnvelopeRemove{proc, p, make(chan ProcGroupRemoveReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	found = retval.found

	return
}

// RemoveTimeout is a gen_server interface method.
func (proc *ProcGroup) RemoveTimeout(p *Process, timeout time.Duration) (found bool, gen_proc_err error) {
	envelope := ProcGroupEnvelopeRemove{proc, p, make(chan ProcGroupRemoveReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	found = retval.found

	return
}

type ProcGroupEnvelopeGet struct {
	proc *ProcGroup

	ret chan ProcGroupGetReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeGet) Call() {
	ret := make(chan ProcGroupGetReturn, 1)
	go func(ret chan ProcGroupGetReturn) {
		retval0 := msg.proc.handleGet()
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

type ProcGroupEnvelopeGetAll struct {
	proc *ProcGroup

	ret chan ProcGroupGetAllReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeGetAll) Call() {
	ret := make(chan ProcGroupGetAllReturn, 1)
	go func(ret chan ProcGroupGetAllReturn) {
		retval0 := msg.proc.handleGetAll()
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

type ProcGroupEnvelopeShutdown struct {
	proc *ProcGroup

	ret chan ProcGroupShutdownReturn
	gen_proc.Envelope
}

func (msg ProcGroupEnvelopeShutdown) Call() {
	ret := make(chan ProcGroupShutdownReturn, 1)
	go func(ret chan ProcGroupShutdownReturn) {
		msg.proc.handleShutdown()
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

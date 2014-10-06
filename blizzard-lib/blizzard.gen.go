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

var BlizzardProcCounter int32

type BlizzardCh struct {
	chMsg                  chan gen_proc.ProcCall
	retChGenCall           chan BlizzardGenCallReturn
	retChAddTagCallback    chan BlizzardAddTagCallbackReturn
	retChRemoveTagCallback chan BlizzardRemoveTagCallbackReturn
	retChRunTagCallback    chan BlizzardRunTagCallbackReturn
	retChAnnounce          chan BlizzardAnnounceReturn
	retChDeploy            chan BlizzardDeployReturn
	retChBootstrapped      chan BlizzardBootstrappedReturn
	retChTakeover          chan BlizzardTakeoverReturn
	retChCleanup           chan BlizzardCleanupReturn
	retChWorkerClosed      chan BlizzardWorkerClosedReturn
	retChSnapshot          chan BlizzardSnapshotReturn
	chStop                 chan struct{}
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

type BlizzardEnvelopeAddTagCallback struct {
	proc *Blizzard
	tag  string
	cb   func(interface{})
	ret  chan BlizzardAddTagCallbackReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeAddTagCallback) Call() {
	ret := make(chan BlizzardAddTagCallbackReturn, 1)
	msg.proc.retChAddTagCallback = ret
	go func(ret chan BlizzardAddTagCallbackReturn) {
		msg.proc.handleAddTagCallback(msg.tag, msg.cb)
		msg.proc.retChAddTagCallback = nil
		ret <- BlizzardAddTagCallbackReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardAddTagCallbackReturn struct {
}

// AddTagCallback is a gen_server interface method.
func (proc *Blizzard) AddTagCallback(tag string, cb func(interface{})) {
	envelope := BlizzardEnvelopeAddTagCallback{proc, tag, cb, make(chan BlizzardAddTagCallbackReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// AddTagCallbackTimeout is a gen_server interface method.
func (proc *Blizzard) AddTagCallbackTimeout(tag string, cb func(interface{}), timeout time.Duration) (gen_proc_err error) {
	envelope := BlizzardEnvelopeAddTagCallback{proc, tag, cb, make(chan BlizzardAddTagCallbackReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type BlizzardEnvelopeRemoveTagCallback struct {
	proc *Blizzard
	tag  string
	ret  chan BlizzardRemoveTagCallbackReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeRemoveTagCallback) Call() {
	ret := make(chan BlizzardRemoveTagCallbackReturn, 1)
	msg.proc.retChRemoveTagCallback = ret
	go func(ret chan BlizzardRemoveTagCallbackReturn) {
		msg.proc.handleRemoveTagCallback(msg.tag)
		msg.proc.retChRemoveTagCallback = nil
		ret <- BlizzardRemoveTagCallbackReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardRemoveTagCallbackReturn struct {
}

// RemoveTagCallback is a gen_server interface method.
func (proc *Blizzard) RemoveTagCallback(tag string) {
	envelope := BlizzardEnvelopeRemoveTagCallback{proc, tag, make(chan BlizzardRemoveTagCallbackReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// RemoveTagCallbackTimeout is a gen_server interface method.
func (proc *Blizzard) RemoveTagCallbackTimeout(tag string, timeout time.Duration) (gen_proc_err error) {
	envelope := BlizzardEnvelopeRemoveTagCallback{proc, tag, make(chan BlizzardRemoveTagCallbackReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type BlizzardEnvelopeRunTagCallback struct {
	proc *Blizzard
	tag  string
	data interface{}
	ret  chan BlizzardRunTagCallbackReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeRunTagCallback) Call() {
	ret := make(chan BlizzardRunTagCallbackReturn, 1)
	msg.proc.retChRunTagCallback = ret
	go func(ret chan BlizzardRunTagCallbackReturn) {
		msg.proc.handleRunTagCallback(msg.tag, msg.data)
		msg.proc.retChRunTagCallback = nil
		ret <- BlizzardRunTagCallbackReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardRunTagCallbackReturn struct {
}

// RunTagCallback is a gen_server interface method.
func (proc *Blizzard) RunTagCallback(tag string, data interface{}) {
	envelope := BlizzardEnvelopeRunTagCallback{proc, tag, data, make(chan BlizzardRunTagCallbackReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// RunTagCallbackTimeout is a gen_server interface method.
func (proc *Blizzard) RunTagCallbackTimeout(tag string, data interface{}, timeout time.Duration) (gen_proc_err error) {
	envelope := BlizzardEnvelopeRunTagCallback{proc, tag, data, make(chan BlizzardRunTagCallbackReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type BlizzardEnvelopeAnnounce struct {
	proc   *Blizzard
	cmd    *blitz.AnnounceCommand
	worker *WorkerConnection
	ret    chan BlizzardAnnounceReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeAnnounce) Call() {
	ret := make(chan BlizzardAnnounceReturn, 1)
	msg.proc.retChAnnounce = ret
	go func(ret chan BlizzardAnnounceReturn) {
		msg.proc.handleAnnounce(msg.cmd, msg.worker)
		msg.proc.retChAnnounce = nil
		ret <- BlizzardAnnounceReturn{}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardAnnounceReturn struct {
}

// Announce is a gen_server interface method.
func (proc *Blizzard) Announce(cmd *blitz.AnnounceCommand, worker *WorkerConnection) {
	envelope := BlizzardEnvelopeAnnounce{proc, cmd, worker, make(chan BlizzardAnnounceReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	<-envelope.ret

	return
}

// AnnounceTimeout is a gen_server interface method.
func (proc *Blizzard) AnnounceTimeout(cmd *blitz.AnnounceCommand, worker *WorkerConnection, timeout time.Duration) (gen_proc_err error) {
	envelope := BlizzardEnvelopeAnnounce{proc, cmd, worker, make(chan BlizzardAnnounceReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	_, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}

	return
}

type BlizzardEnvelopeDeploy struct {
	proc *Blizzard
	cmd  *blitz.DeployCommand
	ret  chan BlizzardDeployReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeDeploy) Call() {
	ret := make(chan BlizzardDeployReturn, 1)
	msg.proc.retChDeploy = ret
	go func(ret chan BlizzardDeployReturn) {
		retval0, retval1 := msg.proc.handleDeploy(msg.cmd)
		msg.proc.retChDeploy = nil
		ret <- BlizzardDeployReturn{retval0, retval1}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardDeployReturn struct {
	retval0 *Application
	retval1 error
}

// Deploy is a gen_server interface method.
func (proc *Blizzard) Deploy(cmd *blitz.DeployCommand) (retval0 *Application, retval1 error) {
	envelope := BlizzardEnvelopeDeploy{proc, cmd, make(chan BlizzardDeployReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval0 = retval.retval0
	retval1 = retval.retval1

	return
}

// DeployTimeout is a gen_server interface method.
func (proc *Blizzard) DeployTimeout(cmd *blitz.DeployCommand, timeout time.Duration) (retval0 *Application, retval1 error, gen_proc_err error) {
	envelope := BlizzardEnvelopeDeploy{proc, cmd, make(chan BlizzardDeployReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval0 = retval.retval0
	retval1 = retval.retval1

	return
}

type BlizzardEnvelopeBootstrapped struct {
	proc *Blizzard
	cmd  *blitz.BootstrapCommand
	ret  chan BlizzardBootstrappedReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeBootstrapped) Call() {
	ret := make(chan BlizzardBootstrappedReturn, 1)
	msg.proc.retChBootstrapped = ret
	go func(ret chan BlizzardBootstrappedReturn) {
		retval0 := msg.proc.handleBootstrapped(msg.cmd)
		msg.proc.retChBootstrapped = nil
		ret <- BlizzardBootstrappedReturn{retval0}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardBootstrappedReturn struct {
	retval0 error
}

// Bootstrapped is a gen_server interface method.
func (proc *Blizzard) Bootstrapped(cmd *blitz.BootstrapCommand) (retval0 error) {
	envelope := BlizzardEnvelopeBootstrapped{proc, cmd, make(chan BlizzardBootstrappedReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	retval0 = retval.retval0

	return
}

// BootstrappedTimeout is a gen_server interface method.
func (proc *Blizzard) BootstrappedTimeout(cmd *blitz.BootstrapCommand, timeout time.Duration) (retval0 error, gen_proc_err error) {
	envelope := BlizzardEnvelopeBootstrapped{proc, cmd, make(chan BlizzardBootstrappedReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	retval0 = retval.retval0

	return
}

type BlizzardEnvelopeTakeover struct {
	proc *Blizzard
	app  string
	kill bool
	ret  chan BlizzardTakeoverReturn
	gen_proc.Envelope
}

func (msg BlizzardEnvelopeTakeover) Call() {
	ret := make(chan BlizzardTakeoverReturn, 1)
	msg.proc.retChTakeover = ret
	go func(ret chan BlizzardTakeoverReturn) {
		err := msg.proc.handleTakeover(msg.app, msg.kill)
		msg.proc.retChTakeover = nil
		ret <- BlizzardTakeoverReturn{err}
	}(ret)
	select {
	case result := <-ret:

		msg.ret <- result
	case <-msg.TimeoutCh():
		close(msg.ret)
	}

}

type BlizzardTakeoverReturn struct {
	err error
}

// Takeover is a gen_server interface method.
func (proc *Blizzard) Takeover(app string, kill bool) (err error) {
	envelope := BlizzardEnvelopeTakeover{proc, app, kill, make(chan BlizzardTakeoverReturn, 1), gen_proc.Envelope{0}}
	proc.chMsg <- envelope
	retval := <-envelope.ret
	err = retval.err

	return
}

// TakeoverTimeout is a gen_server interface method.
func (proc *Blizzard) TakeoverTimeout(app string, kill bool, timeout time.Duration) (err error, gen_proc_err error) {
	envelope := BlizzardEnvelopeTakeover{proc, app, kill, make(chan BlizzardTakeoverReturn, 1), gen_proc.Envelope{timeout}}
	proc.chMsg <- envelope
	retval, ok := <-envelope.ret
	if !ok {
		gen_proc_err = gen_proc.Timeout
		return
	}
	err = retval.err

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
	atomic.AddInt32(&BlizzardProcCounter, 1)
	defer atomic.AddInt32(&BlizzardProcCounter, -1)
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

package blizzard

import (
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"

	"bitbucket.org/ulfurinn/blitz"
)

type ProcGroup struct {
	*ProcGroupGen `gen_proc:"gen_server"`
	server        *Blizzard
	state         string
	Patch         uint64
	paths         []blitz.PathSpec
	exe           *Executable
	PendingProcs  []*Process
	Procs         []*Process
	RemovedProcs  []*Process
	Requests      int64
	TotalRequests uint64
	Written       uint64
	ShuttingDown  bool
	busyWg        sync.WaitGroup
}

type ProgGroupSet map[*ProcGroup]struct{}

func NewProcGroup(server *Blizzard, exe *Executable) *ProcGroup {
	pg := &ProcGroup{
		state:        "init",
		ProcGroupGen: NewProcGroupGen(),
		server:       server,
		exe:          exe,
	}
	pg.inspect()
	return pg
}

func (pg *ProcGroup) inspect() {
	pg.server.inspect(ProcGroupInspect(pg))
}
func (pg *ProcGroup) inspectDispose() {
	pg.server.inspect(ProcGroupInspectDispose(pg))
}

func (pg *ProcGroup) Spawn(count int, cb SpawnedCallback) (err error) {
	pg.GenCall(func() interface{} {
		pg.state = "spawning"
		pg.inspect()
		return nil
	})

	var spawnWg sync.WaitGroup
	spawnWg.Add(count)

	for i := 0; i < count; i++ {
		p := &Process{ProcessGen: NewProcessGen(), server: pg.server, group: pg, tag: randstr(32), announceCb: func(i *Process, ok bool) { spawnWg.Done() }}
		log("[procgroup %p] spawning proc %p\n", pg, p)
		pg.handleAdd(p)
		p.Exec()
	}
	go func() {
		spawnWg.Wait()
		pg.GenCall(func() interface{} {
			pg.state = "ready"
			pg.inspect()
			return nil
		})
		cb(pg)
	}()
	return
}

func (pg *ProcGroup) handleIsReady() bool {
	return pg.state == "ready"
}

func (pg *ProcGroup) handleAdd(p *Process) {
	pg.PendingProcs = append(pg.PendingProcs, p)
}

func (pg *ProcGroup) handleAnnounced(p *Process, cmd *blitz.AnnounceCommand, w *WorkerConnection) (ok bool, first bool) {
	index, found := findProc(p, pg.PendingProcs)
	if !found {
		log("[procgroup %p] unknown process! %p %s\n", pg, p, p.tag)
		return
	}

	go p.Run()

	first = (pg.Patch == 0)
	pg.PendingProcs = removeProc(index, pg.PendingProcs)

	pg.Procs = append(pg.Procs, p)

	if first {
		pg.Patch = cmd.Patch
		pg.paths = cmd.Paths
	}

	p.Announced(cmd, w)
	if pg.ShuttingDown {
		p.Shutdown()
		p.CleanupProcess()
		p.Stop()
		return
	}
	ok = true
	return
}

func findProc(p *Process, list []*Process) (index int, found bool) {
	for i, pr := range list {
		if pr == p {
			index = i
			found = true
			return
		}
	}
	return
}

func removeProc(index int, list []*Process) (result []*Process) {
	list[index] = nil
	result = append(result, list[:index]...)
	result = append(result, list[index+1:]...)
	return
}

func (pg *ProcGroup) handleRemove(p *Process) (found bool) {

	index, found := findProc(p, pg.Procs)
	if found {
		pg.Procs = removeProc(index, pg.Procs)
		go func() {
			p.CleanupProcess()
			p.Stop()
		}()
	}

	index, found = findProc(p, pg.RemovedProcs)
	if found {
		pg.RemovedProcs = removeProc(index, pg.RemovedProcs)
		go func() {
			p.CleanupProcess()
			p.Stop()
		}()
	}

	if len(pg.Procs) == 0 {
		pg.inspectDispose()
		go pg.server.removeGroup(pg)
	}
	return
}

func (pg *ProcGroup) handleFindProcByConnection(w *WorkerConnection) *Process {
	for _, i := range pg.Procs {
		if i.connection == w {
			return i
		}
	}
	for _, i := range pg.RemovedProcs {
		if i.connection == w {
			return i
		}
	}
	return nil
}

func (pg *ProcGroup) handleGet() *Process {
	if pg.ShuttingDown {
		return nil
	}
	if len(pg.Procs) == 0 {
		return nil
	}
	i := pg.Procs[rand.Intn(len(pg.Procs))]
	i.busyWg.Add(1)
	return i
}

func (pg *ProcGroup) handleGetForRemoval() *Process {
	if len(pg.Procs) == 0 {
		return nil
	}
	i := pg.Procs[0]
	pg.Procs = removeProc(0, pg.Procs)
	pg.RemovedProcs = append(pg.RemovedProcs, i)
	return i
}

func (pg *ProcGroup) handleGetAll() []*Process {
	return pg.Procs
}

func (pg *ProcGroup) handleSize() int {
	return len(pg.Procs)
}

func (pg *ProcGroup) Takeover(old *ProcGroup, cb SpawnedCallback) {
	log("[procgroup %p] starting takeover\n", pg)
	pg.state = "takeover"
	pg.inspect()
	old.GenCall(func() interface{} {
		old.state = "handover"
		old.inspect()
		return nil
	})

	var instancesToMount int32
	instancesToMount = int32(pg.exe.Instances / 2)
	if instancesToMount < 1 {
		instancesToMount = 1
	}

	var started int32

	var spawnWg sync.WaitGroup
	spawnWg.Add(pg.exe.Instances)

	var mountWg sync.WaitGroup
	mountWg.Add(1)
	go func() {
		mountWg.Wait()
		cb(pg)
	}()

	procCh := make(chan struct{})

	for i := 0; i < pg.exe.Instances; i++ {
		p := &Process{
			ProcessGen: NewProcessGen(),
			server:     pg.server,
			group:      pg,
			tag:        randstr(32),
			announceCb: func(i *Process, ok bool) {
				spawnWg.Done()
				currentStarted := atomic.AddInt32(&started, 1)
				if currentStarted == instancesToMount {
					mountWg.Done()
				}
				oldProc := old.GetForRemoval()
				if oldProc != nil {
					oldProc.busyWg.Wait()
					oldProc.Shutdown()
				}
				procCh <- struct{}{}
			},
		}
		log("[procgroup %p] spawning proc %p\n", pg, p)
		pg.Add(p)
		err := p.Exec()
		if err != nil {
			log("[procgroup %p] spawn failed: %v\n", err)
			p.announceCb(p, false)
		}
		<-procCh // TODO: this is fragile wrt. failed starts -- will block forever
	}

	spawnWg.Wait()
	pg.state = "ready"
	pg.inspect()
	log("[procgroup %p] takeover done\n", pg)
}

func (pg *ProcGroup) handleShutdown() {
	pg.ShuttingDown = true
	pg.state = "waiting"
	pg.inspect()
	go func() {
		pg.busyWg.Wait()
		pg.state = "shutdown"
		pg.inspect()
		for _, p := range pg.Procs {
			p.Shutdown()
		}
	}()
}

func (pg *ProcGroup) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	i := pg.Get()
	if i == nil {
		resp.WriteHeader(503)
		return
	}
	defer i.busyWg.Done()
	i.ServeHTTP(resp, req)
}

func (pg *ProcGroup) inc() {
	pg.busyWg.Add(1)
	atomic.AddInt64(&pg.Requests, 1)
	atomic.AddUint64(&pg.TotalRequests, 1)
}

func (pg *ProcGroup) dec() {
	atomic.AddInt64(&pg.Requests, -1)
	pg.busyWg.Done()
}

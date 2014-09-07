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
	Requests      int64
	TotalRequests uint64
	Written       uint64
	ShuttingDown  bool
	busyWg        sync.WaitGroup
	spawnWg       sync.WaitGroup
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

func (pg *ProcGroup) handleSpawn(count int, cb SpawnedCallback) (err error) {
	pg.state = "spawning"
	pg.inspect()
	pg.spawnWg.Add(count)
	for i := 0; i < count; i++ {
		p := &Process{ProcessGen: NewProcessGen(), server: pg.server, group: pg, tag: randstr(32)}
		log("[procgroup %p] spawning proc %p\n", pg, p)
		pg.handleAdd(p)
		err = p.Exec()
		if err != nil {
			break
		}
	}
	go func() {
		pg.spawnWg.Wait()
		pg.state = "ready"
		pg.inspect()
		cb(pg)
	}()
	return
}

func (pg *ProcGroup) handleAdd(p *Process) {
	pg.PendingProcs = append(pg.PendingProcs, p)
}

func (pg *ProcGroup) handleAnnounced(p *Process, cmd blitz.AnnounceCommand, w *WorkerConnection) (ok bool, first bool) {
	index, found := findProc(p, pg.PendingProcs)
	if !found {
		return
	}

	go p.Run()
	p.Announced(cmd, w)

	first = (pg.Patch == 0)
	pg.PendingProcs = removeProc(index, pg.PendingProcs)

	if pg.ShuttingDown {
		p.Shutdown()
		p.CleanupProcess()
		p.Stop()
		return
	}

	ok = true

	pg.Procs = append(pg.Procs, p)

	if first {
		pg.Patch = cmd.Patch
		pg.paths = cmd.Paths
	}
	pg.spawnWg.Done()
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
	if !found {
		return
	}
	pg.Procs = removeProc(index, pg.Procs)
	if len(pg.Procs) == 0 {
		pg.inspectDispose()
		go pg.server.removeGroup(pg)
	}
	go func() {
		p.CleanupProcess()
		p.Stop()
	}()
	return
}

func (pg *ProcGroup) handleGet() *Process {
	if pg.ShuttingDown {
		return nil
	}
	if len(pg.Procs) == 0 {
		return nil
	}
	return pg.Procs[rand.Intn(len(pg.Procs))]
}

func (pg *ProcGroup) handleGetAll() []*Process {
	return pg.Procs
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

package blizzard

import (
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"

	"bitbucket.org/ulfurinn/blitz"
	"bitbucket.org/ulfurinn/gen_proc"
)

type PGState int

const (
	PGInit PGState = iota
	PGSpawning
	PGReady
	PGHandover
	PGTakeover
	PGWaiting
	PGAbortedSpawn
	PGShutdown
)

type ProcGroup struct {
	*ProcGroupGen `gen_proc:"gen_server"`
	server        *Blizzard
	state         PGState
	Patch         uint64
	paths         []blitz.PathSpec
	exe           *Application
	PendingProcs  []*Process
	Procs         []*Process
	RemovedProcs  []*Process
	Requests      int64
	TotalRequests uint64
	Written       uint64
	busyWg        sync.WaitGroup
}

type ProgGroupSet map[*ProcGroup]struct{}

func NewProcGroup(server *Blizzard, exe *Application) *ProcGroup {
	pg := &ProcGroup{
		state:        PGInit,
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

func (pg *ProcGroup) handleSpawn() (gen_proc.Deferred, error) {
	count := pg.exe.Instances
	if count == 0 {
		count = 1
	}
	pg.state = PGSpawning
	pg.inspect()

	return pg.deferSpawn(func(ret func(error)) {
		var err error
		for i := 0; i < count; i++ {
			p := &Process{ProcessGen: NewProcessGen(), server: pg.server, group: pg, tag: randstr(32)}
			go p.Run()
			log("[procgroup %p] spawning proc %p\n", pg, p)
			pg.AddPending(p)
			var announcement *blitz.AnnounceCommand
			announcement, err = p.Exec()
			pg.PendingSpawned(p, err == nil)
			if err == nil {
				if i == 0 {
					pg.Patch = announcement.Patch
					pg.paths = announcement.Paths
				}
			} else {
				p.Stop()
				log("[procgroup %p] %v\n", pg, err)
				break
			}
		}
		pg.GenCall(func() interface{} {
			switch pg.state {
			case PGSpawning:
				pg.state = PGReady
				pg.inspect()
			case PGAbortedSpawn:
				go pg.Shutdown()
			default:
				log("[procgroup %p] don't know how to finalise a spawn in state %v\n", pg, pg.state)
			}
			return nil
		})
		ret(err)
	})
}

func (pg *ProcGroup) handleIsReady() bool {
	return pg.state == PGReady
}

func (pg *ProcGroup) handleAddPending(p *Process) {
	pg.PendingProcs = append(pg.PendingProcs, p)
}

func (pg *ProcGroup) handlePendingSpawned(p *Process, spawnOk bool) {
	if index, found := findProc(p, pg.PendingProcs); found {
		pg.PendingProcs = removeProc(index, pg.PendingProcs)
	}
	if spawnOk {
		pg.Procs = append(pg.Procs, p)
	}
}

func (pg *ProcGroup) handleWorkerClosed(w *WorkerConnection) {
	// TODO: consider removed list?
	var proc *Process
	for _, p := range append(pg.Procs, pg.RemovedProcs...) {
		if p.connection == w {
			proc = p
			break
		}
	}
	if proc == nil {
		log("[procgroup %p] no process found for connection %p\n", pg, w)
		return
	}
	go pg.RemoveDeadProc(proc)
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

func findAndRemoveProcInList(p *Process, list *[]*Process) (found bool) {
	var index int
	index, found = findProc(p, *list)
	if found {
		*list = removeProc(index, *list)
		go func() {
			p.CleanupProcess()
			p.Stop()
		}()
	}
	return
}

func (pg *ProcGroup) handleRemoveDeadProc(p *Process) (found bool) {

	found = findAndRemoveProcInList(p, &pg.Procs)
	if !found {
		found = findAndRemoveProcInList(p, &pg.RemovedProcs)
	}

	//	TODO: consider removed and pending?
	// if no active procs remain, only unmount without removing?
	if len(pg.Procs) == 0 {
		log("[procgroup %p] disposing of itself\n", pg)
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
	if pg.state == PGShutdown {
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

// TODO: test takeover failures extensively

func (pg *ProcGroup) Takeover(old *ProcGroup, cb SpawnedCallback, kill bool) error {
	log("[procgroup %p] starting takeover\n", pg)
	pg.state = PGTakeover
	pg.inspect()
	old.GenCall(func() interface{} {
		old.state = PGHandover
		old.inspect()
		return nil
	})

	var instancesToMount int32
	instancesToMount = int32(pg.exe.Instances / 2)
	if instancesToMount < 1 {
		instancesToMount = 1
	}

	var started int32
	remounted := false
	var spawnErr error

	for i := 0; i < pg.exe.Instances; i++ {
		p := &Process{
			ProcessGen: NewProcessGen(),
			server:     pg.server,
			group:      pg,
			tag:        randstr(32),
		}
		go p.Run()
		log("[procgroup %p] spawning proc %p\n", pg, p)
		pg.AddPending(p)
		cmd, err := p.Exec()
		pg.PendingSpawned(p, err == nil)
		if err == nil {
			pg.Patch = cmd.Patch
			pg.paths = cmd.Paths
			currentStarted := atomic.AddInt32(&started, 1)
			if currentStarted == instancesToMount {
				remounted = true
				cb(pg)
			}
			oldProc := old.GetForRemoval()
			log("[procgroup %p] shutting down old proc %p\n", pg, oldProc)
			if oldProc != nil {
				oldProc.busyWg.Wait()
				oldProc.Shutdown(kill)
			}
		} else {
			log("[procgroup %p] error in takeover: %v\n", pg, err)
			spawnErr = err
			break
		}
	}

	if spawnErr == nil {
		pg.state = PGReady
		pg.inspect()
		log("[procgroup %p] takeover done\n", pg)
		return nil
	} else {
		var disposeGroup *ProcGroup
		if remounted {
			disposeGroup = old
		} else {
			disposeGroup = pg
		}
		log("[procgroup %p] error completing takeover (remounted to new: %v), shutting down group %p\n", pg, remounted, disposeGroup)
		disposeGroup.Shutdown()
		return spawnErr
	}
}

func (pg *ProcGroup) handleShutdown() {
	log("[procgroup %p] shutdown requested\n", pg)
	switch pg.state {
	case PGReady, PGAbortedSpawn, PGHandover, PGTakeover:
		pg.state = PGShutdown
		pg.inspect()
		log("[procgroup %p] waiting for running requests\n", pg)
		pg.busyWg.Wait()
		log("[procgroup %p] shutdown commencing: %d procs\n", pg, len(pg.Procs))
		if len(pg.Procs) == 0 {
			log("[procgroup %p] disposing of itself\n", pg)
			pg.inspectDispose()
			go pg.server.removeGroup(pg)
		}
		for _, p := range pg.Procs {
			log("[procgroup %p] proc %p\n", pg, p)
			p.Shutdown(false)
		}
	case PGSpawning:
		pg.state = PGAbortedSpawn
		pg.inspect()
	case PGShutdown:
		log("[procgroup %p] already shutting down\n", pg)
	default:
		log("[procgroup %p] don't know how to react to shutdown in state %v\n", pg, pg.state)
	}

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

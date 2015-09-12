package distributedvc

import (
    "sync"
    "definition.configinfo"
    "time"
)

/*
    Class representing a merging worker of intranode merge-process. For scalability,
    that several workers work simultaneously is feasible. Supervisor is one scheduler,
    with all the workers on one same file sharing the same, responsible for spawning
    worker, assigning his pinpoint(one node on the linked-list to merge), and permits
    his requirement to merge new pair or forbid this task and kill him.

    Intramerge Process:
    All the patches are stored in a linked-list, with next pointer stored in metadata.
    The first of the list is always patch0, and no explicit last node is defined. When
    there's no file, the end has been reached.
    Each worker insist working on a PINPOINT, that is, one fixed patch-id to be one
    participant of merge. The other one is always its next. After merging it, remove
    the non-pinpoint node out-of linked-list, modifying metadata correspondingly.

    LEGEND:
    0 -> 1 -> 2 -> 5 ->
    ^         ^

    Attetez: It could only be run by one goroutine.
*/
type IntramergeWorker struct{
    supervisor *IntramergeSupervisor
    pinpoint int
    fd *Fd
    havemerged int
}

func NewIntramergeWorker(_supervisor *IntramergeSupervisor, _pinpoint int) *IntramergeWorker {
    return &IntramergeWorker {
        supervisor: _supervisor,
        pinpoint: _pinpoint,
        fd: _supervisor.filed,
        havemerged: 0 ,
    }
}

func (this *IntramergeWorker)run() {
    // TODO
}


type taskLinknode struct {
    status int
    next int
}
const TASKSTATUS_IDLE=0
const TASKSTATUS_WORKING=1
func NewtaskLinknode(_status int, _next int) *taskLinknode {
    return &taskLinknode{
        status: _status,
        next: _next,
    }
}


/*
    Supervisor that manages mergeworkers, creates them, permits their report and kills
    them. Each file should only have one supervisor so it is class-layer-sync-ed.
*/
type IntramergeSupervisor struct {
    taskMap map[int]*taskLinknode
    filed *Fd
    workersAlive int

    locks []*sync.Mutex
}
func NewIntramergeSupervisor(filedes *Fd) *IntramergeSupervisor {
    return &IntramergeSupervisor{
        locks: []*sync.Mutex{&sync.Mutex{},&sync.Mutex{}},
        taskMap: map[int]*taskLinknode{},
        filed: filedes,
        workersAlive: 0,
    }
}

const REPORT_TASK_RESPONSE_CONFIRMED=0        // Approve to continue work
const REPORT_TASK_RESPONSE_REJECT=1           // Reject the work, the worker should commit all the change and suicide
const REPORT_TASK_RESPONSE_COMMIT=2           // Approve, on condition that the status merging be commited first
// Try to declare a task. Return value determines the behavior of worker.
// @Sync(0)
func (this *IntramergeSupervisor)ReportNewTask(worker *IntramergeWorker, patchnum int, oldpatch int /*For none use -1*/) int {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    if oldpatch!=-1 {
        // Remove the merged one from the list. May delete it from Swift.
        this.taskMap[worker.pinpoint].next=patchnum
        delete(this.taskMap.oldpatch)
    }
    if elem, ok:=this.taskMap[patchnum]; !ok || elem.status==TASKSTATUS_WORKING {
        return REPORT_TASK_RESPONSE_REJECT
    }
    this.taskMap[patchnum].status=TASKSTATUS_WORKING
    if worker.havemerged%int(configinfo.GetProperty_Node("auto_commit_per_intramerge").(float64))==0 {
        return REPORT_TASK_RESPONSE_COMMIT
    }
    return REPORT_TASK_RESPONSE_CONFIRMED
}

const REPORT_DEATH_DIEOF_STARVATION=0
const REPORT_DEATH_DIEOF_COMMAND=1
const REPORT_DEATH_DIEOF_EXCEPTION=2
// @Sync(0)
func (this *IntramergeSupervisor)ReportDeath(worker *IntramergeWorker, dieof int) {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    this.workersAlive=this.workersAlive-1
    this.taskMap[worker.pinpoint].status=TASKSTATUS_IDLE
    if this.workersAlive==0 & len(this.taskMap)>1 {
        time.Sleep(time.Second)
        go this.BatchWorker(-1, -1)
    }
}

// Garantee that pinpoint exists!
// @Sync(0)
func (this *IntramergeSupervisor)SpawnWorker(pinpoint int) {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    if this.taskMap[pinpoint].status==TASKSTATUS_WORKING {
        return
    }
    this.taskMap[pinpoint].status=TASKSTATUS_WORKING
    this.workersAlive=this.workersAlive+1

    newWorker:=NewIntramergeWorker(this, pinpoint)
    go newWorker.run()
}

// Nextpatch may be nonexist
// Sync(0)
func (this *IntramergeSupervisor)AnnounceNewTask(patchnum int, nextpatch int) {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    this.taskMap[patchnum]=NewtaskLinknode(TASKSTATUS_IDLE,nextpatch)
}

// Only the two arguments can be specified one, nums indicating the whole number
// and range indicating the interval. The first one has higher priority, with both
// absent, nums=1 is the default.
func max(x1 int, x2 int) int {
    if x1>x2 {
        return x1
    }
    return x2
}
func (this *IntramergeSupervisor)BatchWorker(nums/*=-1*/, ranges/*=-1*/) {
    if this.workersAlive>0 || len(this.taskMap)<=1 {
        return
    }
    if nums<0 && ranges<0 {
        nums=1
    }
    if nums==1 {
        this.SpawnWorker(0)
        return
    }
    if nums>0 {
        ranges=max(len(this.taskMap)/nums, 2)
    }
    p=0
    for {
        this.SpawnWorker(0)
        nums=nums-1
        if nums==0 {
            return
        }
        for i:=0;i<ranges;i++ {
            elem, err:=this.taskMap[p]
            if err!=nil {
                return
            }
            p=elem.next
        }
    }
}

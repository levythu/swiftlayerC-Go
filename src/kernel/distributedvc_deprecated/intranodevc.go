package distributedvc

import (
    "sync"
    "definition/configinfo"
    "time"
    "logger"
    "fmt"
    "strconv"
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
    workingOn int
}

func ____no_use_intra() {
    fmt.Println("Nouse")
}

func NewIntramergeWorker(_supervisor *IntramergeSupervisor, _pinpoint int) *IntramergeWorker {
    return &IntramergeWorker {
        supervisor: _supervisor,
        pinpoint: _pinpoint,
        fd: _supervisor.filed,
        havemerged: 0,
        workingOn: -1,
    }
}

func (this *IntramergeWorker)run() {
    dieof:=REPORT_DEATH_DIEOF_EXCEPTION
    defer this.supervisor.ReportDeath(this, &dieof)

    pinedMeta, pinedFile, err:=this.fd.io.Get(this.fd.GetPatchName(this.pinpoint, -1))
    if err!=nil {
        logger.Secretary.Error("kernel.distributedvc.IntramergeWorker::run()", err)
        return
    }
    prevTask:=-1
    nextTask:=this.supervisor.CheckNext(this.pinpoint)

    uploadFile:=func() error {
        if prevTask==-1 {
            // Nothing has been merged yet. ABORT uploading
            return nil
        }
        pinedMeta[INTRA_PATCH_METAKEY_NEXT_PATCH]=strconv.Itoa(nextTask)
        err:=this.fd.io.Put(this.fd.GetPatchName(this.pinpoint, -1), pinedFile, pinedMeta)
        if err!=nil {
            logger.Secretary.Error("kernel.distributedvc.IntramergeWorker::run.uploadFile()", err)
            return err
        }
        logger.Secretary.Log("kernel.distributedvc.IntramergeWorker::run.uploadFile()", this.fd.GetPatchName(this.pinpoint, -1)+" is uploaded.")
        if this.pinpoint==0 {
            // Start propagation on the splittree
            this.fd.intervisor.PropagateUp()
        }
        return nil
    }
    workOnMerge:=func() error {
        _, nextFile, err:=this.fd.io.Get(this.fd.GetPatchName(nextTask, -1))
        if err!=nil {
            logger.Secretary.Error("kernel.distributedvc.IntramergeWorker::run.workOnMerge()", err)
            return err
        }
        pinedFile, err=pinedFile.MergeWith(nextFile)
        if err!=nil {
            logger.Secretary.Error("kernel.distributedvc.IntramergeWorker::run.workOnMerge()", err)
            return err
        }
        this.havemerged++
        prevTask=nextTask
        nextTask=this.supervisor.CheckNext(nextTask)
        return nil
    }

    for {
        cmd:=this.supervisor.ReportNewTask(this, nextTask, prevTask)

        if cmd==REPORT_TASK_RESPONSE_CONFIRMED {
            if workOnMerge()!=nil {
                return
            }
        } else if cmd==REPORT_TASK_RESPONSE_REJECT {
            if uploadFile()!=nil {
                return
            }
            dieof=REPORT_DEATH_DIEOF_COMMAND
            return
        } else {
            if uploadFile()!=nil {
                return
            }
            if workOnMerge()!=nil {
                return
            }
        }
    }
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

// @Sync(0)
func (this *IntramergeSupervisor)GetWorkersCount() int {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    return this.workersAlive
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
        worker.workingOn=-1
        delete(this.taskMap, oldpatch)
    }
    if elem, ok:=this.taskMap[patchnum]; !ok || elem.status==TASKSTATUS_WORKING {
        return REPORT_TASK_RESPONSE_REJECT
    }
    this.taskMap[patchnum].status=TASKSTATUS_WORKING
    worker.workingOn=patchnum
    if worker.havemerged%configinfo.AUTO_COMMIT_PER_INTRAMERGE==0 {
        return REPORT_TASK_RESPONSE_COMMIT
    }
    return REPORT_TASK_RESPONSE_CONFIRMED
}

const REPORT_DEATH_DIEOF_STARVATION=0
const REPORT_DEATH_DIEOF_COMMAND=1
const REPORT_DEATH_DIEOF_EXCEPTION=2
// @Sync(0)
func (this *IntramergeSupervisor)ReportDeath(worker *IntramergeWorker, dieof *int) {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    this.workersAlive=this.workersAlive-1
    this.taskMap[worker.pinpoint].status=TASKSTATUS_IDLE
    if worker.workingOn>=0 {
        this.taskMap[worker.workingOn].status=TASKSTATUS_IDLE
    }
    if this.workersAlive==0 && len(this.taskMap)>1 {
        time.Sleep(time.Second) //TODO: need to wait?
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

// Nextpatch may be nonexist. Annoucing new task will not cause immediately spawning
// workers.
// Sync(0)
func (this *IntramergeSupervisor)AnnounceNewTask(patchnum int, nextpatch int) {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    this.taskMap[patchnum]=NewtaskLinknode(TASKSTATUS_IDLE,nextpatch)
}

// Lookup in the map to find the next.
// Sync(0)
func (this *IntramergeSupervisor)CheckNext(item int) int {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    e, ok:=this.taskMap[item]
    if ok {
        return e.next
    } else {
        return -1
    }
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
func (this *IntramergeSupervisor)BatchWorker(nums/*=-1*/ int, ranges/*=-1*/ int) {
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
    p:=0
    for {
        this.SpawnWorker(0)
        nums=nums-1
        if nums==0 {
            return
        }
        for i:=0;i<ranges;i++ {
            elem, ok:=this.taskMap[p]
            if !ok {
                return
            }
            p=elem.next
        }
    }
}
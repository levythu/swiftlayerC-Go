package distributedvc

// components for invocating auto-merge automatically. Should be launched in a seperate goroutine
// and schedule mewrging work periodically.

import (
    "sync"
    "errors"
    conf "definition/configinfo"
    . "outapi"
    . "logger"
    "strconv"
    "time"
)


type taskNode struct {
    prev *taskNode
    next *taskNode
    taskID string
}
// This class is used for maintaining marging task order for a lot of merging requests
// from FDs
type MergingScheduler struct {
    lock *sync.RWMutex

    // existMap, if map[key]==true, the key has not been checked-out or has been and
    // no new identical task is checked in during the working time
    // if map[key]==false, an identical task has been checked-in, so when commiting,
    // another inspection is needed.
    existMap map[string]bool
    taskQueueHead taskNode
    taskQueueTail taskNode

    taskInTotal int
}

func NewScheduler() *MergingScheduler {
    var ret=&MergingScheduler {
        lock: &sync.RWMutex{},
        existMap: make(map[string]bool),
    }
    ret.taskQueueHead.prev=nil
    ret.taskQueueHead.next=&ret.taskQueueTail
    ret.taskQueueTail.next=nil
    ret.taskQueueTail.prev=&ret.taskQueueHead

    return ret
}

var QUEUE_CAPACITY_REACHED=errors.New("The task queue is filled up.")
// if err!=nil, the Scheduler simply reject its task check-in request.
func (this *MergingScheduler)CheckInATask(filename string, io Outapi) error {
    this.lock.Lock()
    defer this.lock.Unlock()

    var id=genID_static(filename, io)

    if _, ok:=this.existMap[id]; ok {
        // the task has existed in the queue. DO NOT NEED to check in it again.
        this.existMap[id]=false
        return nil
    }


    if this.taskInTotal>=conf.AUTO_MERGER_TASK_QUEUE_CAPACITY {
        return QUEUE_CAPACITY_REACHED
    }
    this.taskInTotal++

    var newNode=&taskNode {
        next: &this.taskQueueTail,
        prev: this.taskQueueTail.prev,
        taskID: id,
    }
    newNode.next.prev=newNode
    newNode.prev.next=newNode
    this.existMap[id]=true

    return nil
}

var NO_TASK_AVAILABLE=errors.New("No task is available in the queue.")
// if error!=nil, some error was encountered or simply there's no task available
func (this *MergingScheduler)ChechOutATask() (string, error) {
    this.lock.Lock()
    defer this.lock.Unlock()

    if this.taskInTotal==0 {
        return "", NO_TASK_AVAILABLE
    }
    this.taskInTotal--
    var toDel=this.taskQueueHead.next
    toDel.prev.next=toDel.next
    toDel.next.prev=toDel.prev

    return toDel.taskID, nil
}

// if returns==false, it is needed to inspect the task again.
// otherwise, the task is successfully removed.
func (this *MergingScheduler)FinishTask(taskID string) bool {
    this.lock.Lock()
    defer this.lock.Unlock()

    if val, ok:=this.existMap[taskID]; !ok {
        panic("UNEXPECTED LOGICAL FLOW!")
    } else {
        if val {
            delete(this.existMap, taskID)
            return true
        } else {
            this.existMap[taskID]=true
            return false
        }
    }
}

// =============================================================================
// =============================================================================
// =============================================================================

type MergingSupervisor struct {
    lock *sync.RWMutex

    workersAlive int
    scheduler *MergingScheduler
    deamoned bool
}

var MergeManager=&MergingSupervisor {
    lock: &sync.RWMutex{},
    workersAlive: 0,
    scheduler: NewScheduler(),
    deamoned: false,
}


func (this *MergingSupervisor)SubmitTask(filename string, io Outapi) error {
    if err:=this.scheduler.CheckInATask(filename, io); err!=nil {
        return err
    }

    this.spawnWorker()
    return nil
}

func (this *MergingSupervisor)reportDeath() {
    this.lock.Lock()
    defer this.lock.Unlock()
    this.workersAlive--
}

func (this *MergingSupervisor)spawnWorker() {
    this.lock.RLock()
    if this.workersAlive>=conf.MAX_MERGING_WORKER {
        this.lock.RUnlock()
        return
    }
    this.lock.RUnlock()

    this.lock.Lock()
    defer this.lock.Unlock()
    if this.workersAlive>=conf.MAX_MERGING_WORKER {
        return
    }
    this.workersAlive++
    go workerProcess(this)
}

// periodically spawn a worker to finish unadopted tasks
func (this *MergingSupervisor)Deamon() {
    if conf.AUTO_MERGER_DEAMON_PERIOD<=0 {
        return
    }
    Secretary.Log("kernel.distributedvc::Deamon", "Auto merger deamon is running at period "+strconv.Itoa(conf.AUTO_MERGER_DEAMON_PERIOD)+" second(s)")
    var period=time.Second*time.Duration(conf.AUTO_MERGER_DEAMON_PERIOD)
    for {
        // RUN FOREVER
        this.lock.RLock()
        var t=this.workersAlive
        this.lock.RUnlock()
        if t==0 {
            this.spawnWorker()
        }

        time.Sleep(period)
    }
}

func (this *MergingSupervisor)LaunchDeamon() {
    this.lock.Lock()
    defer this.lock.Unlock()

    if this.deamoned {
        return
    }
    this.deamoned=true
    go this.Deamon()
}

// =============================================================================

func workerProcess(supervisor *MergingSupervisor) {
    return
}

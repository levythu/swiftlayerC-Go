package distributedvc

import (
    "sync"
    _ "logger"
    "kernel/filetype"
    . "outapi"
)

type FD struct {
    /*====BEGIN: for fdPool====*/
    lock *sync.Mutex

    filename string
    io Outapi
    reader int
    peeper int

    // 1 for active, 2 for dormant, 0 for uninited, 99 for dead
    status int

    isInTrash bool
    isInDormant bool
    trashNode *fdDLinkedListNode
    dormantNode *fdDLinkedListNode
    /*====END: for fdPool====*/

    /*====BEGIN: for functionality====*/
    updateChainLock *sync.RWMutex
    nextAvailablePosition int
}

const (
    INTRA_PATCH_METAKEY_NEXT_PATCH="next-patch"
)

func newFD(filename string, io Outapi) *FD {
    var ret=&FD {
        filename: filename,
        io: io,
        reader: 0,
        peeper: 0,
        status: 0,
        lock: &sync.Mutex{},
        isInDormant: false,
        isInTrash: false,

        updateChainLock: &sync.RWMutex{},
        nextAvailablePosition: -1,
    }
    ret.trashNode=&fdDLinkedListNode {
        carrier: ret,
    }
    ret.dormantNode=&fdDLinkedListNode {
        carrier: ret,
    }

    return ret
}

func genID_static(filename string, io Outapi) string {
    return filename+"@@"+io.GenerateUniqueID()
}
func (this *FD)ID() string {
    return genID_static(this.filename, this.io)
}
func (this *FD)GoDie() {
    this.lock.Lock()
    defer this.lock.Unlock()

    if this.status!=1 {
        return
    }
    this.status=99

    return
}
func (this *FD)GoDormant() bool {
    this.lock.Lock()
    defer this.lock.Unlock()
    this.isInDormant=false
    //logger.Secretary.LogD("Filehandler "+this.filename+" is going dormant.")
    if this.status!=1 {
        // noe active yet.
        return false
    }
    this.status=2

    return true
}
func (this *FD)Read() {
    this.lock.Lock()
    defer this.lock.Unlock()

    // load in and parse file
    this.status=1
}
func (this *FD)GoGrasped() {
    this.LoadPointerMap()
}

func (this *FD)LoadPointerMap() error {
    this.updateChainLock.Lock()
    defer this.updateChainLock.Unlock()

    if this.nextAvailablePosition>=0 {
        return nil
    }
    return nil
}
func (this *FD)Submit(object *filetype.Kvmap) error {
    return nil
}

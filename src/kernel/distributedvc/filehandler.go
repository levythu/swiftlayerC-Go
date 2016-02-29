package distributedvc

import (
    "sync"
    _ "logger"
)

type FD struct {
    lock *sync.Mutex

    filename string
    reader int
    peeper int

    // 1 for active, 2 for dormant, 0 for uninited, 99 for dead
    status int

    isInTrash bool
    isInDormant bool
    trashNode *fdDLinkedListNode
    dormantNode *fdDLinkedListNode
}

func newFD(filename string) *FD {
    var ret=&FD {
        filename: filename,
        reader: 0,
        peeper: 0,
        status: 0,
        lock: &sync.Mutex{},
        isInDormant: false,
        isInTrash: false,
    }
    ret.trashNode=&fdDLinkedListNode {
        carrier: ret,
    }
    ret.dormantNode=&fdDLinkedListNode {
        carrier: ret,
    }

    return ret
}

func (this *FD)GoDie() {
    this.lock.Lock()
    defer this.lock.Unlock()

    if this.status!=1 {
        return false
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

package distributedvc

import (
    "sync"
    conf "definition/configinfo"
)

/*
** When a fd is in fdPool, it is active, dormant or dead:
** Active:  held by >0 goroutines, while the number of holders reduces to 0, it will be throwed
**          into dormant list, waiting for wiper to change it dormant. In the status, all the
**          information is loaded into memory.
** Dormant: held by arbitrary goroutines
*/

// Lock priority: fdPool lock(locks[0])> dormant working lock(locks[1])> corresponsding fd lock >trash/dormant lock

var fdPool=make(map[string]*FD)

var trash=NewFSDLinkedList()
var dormant=NewFSDLinkedList()

var locks=[]*sync.RWMutex{&sync.RWMutex{}, &sync.RWMutex{}}

func GetFD(filename string) *FD {
    locks[0].RLock()
    var elem, ok=fdPool[filename]
    if ok {
        elem.Grasp()
        locks[0].RUnlock()
        return elem
    }
    locks[0].RUnlock()

    locks[0].Lock()
    elem, ok=fdPool[filename]
    if ok {
        elem.Grasp()
        locks[0].Unlock()
        return elem
    }
    // New a FD
    var ret=newFD(filename)
    fdPool[filename]=ret
    ret.Grasp()
    return ret
}
func GetFDWithoutModifying(filename string) *FD {
    locks[0].RLock()
    defer locks[0].RUnlock()
    var elem, ok=fdPool[filename]
    if ok {
        elem.Grasp()
        return elem
    }
    return nil
}

func ClearTrash() {
    locks[0].Lock()
    defer locks[0].Unlock()

    trash.Lock.Lock()

    var nLimit int=conf.MAX_NUMBER_OF_CACHED_DORMANT_FD/2
    if trash.Length<=nLimit {
        trash.Lock.Unlock()
        return
    }

    // peel half of the list to del-list, and delete all the elements in it one by one.

    var p=trash.Head
    for i:=0; i<nLimit; i++ {
        p=p.next
    }

    var delHead=trash.Head.next
    var delTail=p
    trash.Head.next=p.next
    trash.Head.next.prev=trash.Head

    trash.Length-=nLimit

    trash.Lock.Unlock()

    // In such circumstance, all the FDs in the del-list is ungrasped. So lock is
    // not needed to operate on them.

    delTail.next=nil
    for delHead!=nil {
        delete(fdPool, delHead.carrier.filename)
        delHead=delHead.next
    }
}
func clearDormant() {
    // different from ClearTrash(), this function does not need to modify the pool, so
    // a global lock is not essential. However, to prevent re-adding, a dormant working
    // lock is used.
    locks[1].Lock()
    defer locks[1].Unlock()

    dormant.Lock.Lock()
    var nLimit int=conf.MAX_NUMBER_OF_CACHED_ACTIVE_FD/2
    if dormant.Length<=nLimit {
        dormant.Lock.Unlock()
        return
    }

    var p=dormant.Head
    for i:=0; i<nLimit; i++ {
        p=p.next
    }

    var delHead=dormant.Head.next
    var delTail=p
    dormant.Head.next=p.next
    dormant.Head.next.prev=dormant.Head

    dormant.Length-=nLimit

    dormant.Lock.Unlock()

    // At this moment, all the elements in del-list is no-reader-grasped and no new reader-grasping
    // is allowed.
    delTail.next=nil
    for delHead!=nil {
        delHead.carrier.GoDormant()
        delHead=delHead.next
    }
}

func (this *FD)Grasp() {
    // If in trashlist, remove it.
    this.lock.Lock()
    defer this.lock.Unlock()
    this.peeper++
    if this.isInTrash {
        this.isInTrash=false
        trash.Cut(this.trashNode)
    }
}
func (this *FD)Release() {
    // If peeper==0, throw into trashlist and check capacity
    this.lock.Lock()
    defer this.lock.Unlock()
    this.peeper--
    if this.peeper==0 {
        this.isInTrash=true
        trash.Lock.Lock()
        trash.AppendWithoutLock(this.trashNode)
        if trash.Length>=conf.MAX_NUMBER_OF_CACHED_DORMANT_FD {
            go ClearTrash()
        }
        trash.Lock.Unlock()
    }
}
func (this *FD)GraspReader() {
    locks[1].RLock()
    defer locks[1].RUnlock()

    this.lock.Lock()
    defer this.lock.Unlock()

    this.reader++
    if this.isInDormant {
        this.isInDormant=false
        dormant.Cut(this.dormantNode)
    }
}
func (this *FD)ReleaseReader() {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.reader--
    if this.reader==0 && this.status==1 {
        this.isInDormant=true
        dormant.Lock.Lock()
        dormant.AppendWithoutLock(this.dormantNode)
        if dormant.Length>=conf.MAX_NUMBER_OF_CACHED_ACTIVE_FD {
            go clearDormant()
        }
        dormant.Lock.Unlock()
    }
}

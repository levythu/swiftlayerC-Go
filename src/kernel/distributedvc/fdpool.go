package distributedvc

import (
    "sync"
)

/*
** When a fd is in fdPool, it is active, dormant or dead:
** Active:  held by >0 goroutines, while the number of holders reduces to 0, it will be throwed
**          into dormant list, waiting for wiper to change it dormant. In the status, all the
**          information is loaded into memory.
** Dormant: held by arbitrary goroutines
*/

// Lock priority: fdPool lock> trash/dormant lock

var fdPool=make(map[string]*FD)

var trash=NewFSDLinkedList()
var dormant=NewFSDLinkedList()

var locks []*sync.RWMutex{&sync.RWMutex{}, &sync.RWMutex{}}

func GetFD(filename string) *Fd {
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
    var ret=newFD()
    fdPool[filename]=ret
    ret.Grasp()
    return ret
}
func GetFDWithoutModifying(filename string) *Fd {
    locks[0].RLock()
    defer locks[0].RUnlock()
    var elem, ok=fdPool[filename]
    if ok {
        elem.Grasp()
        return elem
    }
    return nil
}

func (this *Fd)Grasp() {
    // If in trashlist, remove it.
    this.lock.Lock()
    defer this.lock.Unlock()
    this.peeper++
    if this.trashNode!=nil {
        
    }
}
func (this *Fd)Release() {

}
func (this *Fd)CatchReader() {

}
func (this *Fd)ReleaseReader() {

}

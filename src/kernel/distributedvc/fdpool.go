package distributedvc

/*
** Global pool to manage out-of-date filehandler. Besides the global map, a linked
** list is maintained to record all the non-refered filehandler. When there're no
** enough space, a sync-clean is called to clean up all the outdated.
*/

import (
    "container/list"
    "sync"
)

var uselessList=list.New()
var gblock=&sync.Mutex{}

// @Sync
func AddToUseless(obj *Fd) {
    gblock.Lock()
    defer gblock.Unlock()
    if obj.uselessPosition!=nil {
        return
    }
    obj.uselessPosition=uselessList.PushBack(obj)
}

// @Sync
func RemoveFromUseless(obj *Fd) bool {
    gblock.Lock()
    defer gblock.Unlock()
    if obj.uselessPosition==nil {
        return false
    }
    uselessList.Remove(obj.uselessPosition)
    obj.uselessPosition=nil
    return true
}

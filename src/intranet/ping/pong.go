package ping

import (
    . "utils/timestamp"
    "sync"
    . "intranet/gossipd/interactive"
)

var activeList=map[int]ClxTimestamp{}
var aLLock sync.RWMutex

// nonexist returns zero
func QueryConn(nodenum int) ClxTimestamp {
    aLLock.RLock()
    defer aLLock.RUnlock()

    return activeList[nodenum]
}

func DumpConn(nodenum int) map[int]ClxTimestamp {
    aLLock.RLock()
    defer aLLock.RUnlock()

    var ret=map[int]ClxTimestamp{}
    for k, v:=range activeList {
        ret[k]=v
    }

    return ret
}

// returning value indicates whether the gossip should be passed on
func Pong(context *GossipEntry) bool {
    aLLock.Lock()
    defer aLLock.Unlock()

    if activeList[context.NodeNumber]<context.UpdateTime {
        activeList[context.NodeNumber]=context.UpdateTime
        return true
    } else {
        // it has been notified before. there is no need to propagate the gossip
        return false
    }
}

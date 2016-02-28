package distributedvc

import (
    "sync"
)

type FD struct {
    lock *sync.Mutex

    filename string
    reader int
    peeper int

    // 1 for active, 2 for dormant, 0 for uninited
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
        isInTrash: false
    }
    ret.trashNode=&fdDLinkedListNode {
        carrier: ret
    }
    ret.dormantNode=&fdDLinkedListNode {
        carrier: ret
    }

    return ret
}

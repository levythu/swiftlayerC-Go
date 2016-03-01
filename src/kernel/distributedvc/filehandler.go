package distributedvc

import (
    "sync"
    _ "logger"
    "kernel/filetype"
    . "outapi"
    "strconv"
    . "definition/configinfo"
    . "utils/timestamp"
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

    // only available when active
    numberZero *filetype.Kvmap
    contentLock *sync.RWMutex
}

// Lock priority: lock > updateChainLock > contentLock

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

        numberZero: nil,
        contentLock: &sync.RWMutex{},
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

func (this *FD)ReadInNumberZero() error {

}

/*
** Patch list: 0(the combined version) -> 1 -> 2 -> ...
** If the #0 patch does not exist, the file does not have a separate version in the node.
** otherwise, "INTRA_PATCH_METAKEY_NEXT_PATCH" in the meta will form a linked list
** to chain all the uncombined patch.
**
** As soon as the file is loaded into system, its uncombined patch will start to combine
** and the dormant fd will store the next available patch number.
*/

func (this *Fd)GetPatchName(patchnumber int, nodenumber int/*-1*/) string {
    if nodenumber<0 {
        nodenumber=NODE_NUMBER
    }
    return this.filename+".node"+strconv.Itoa(nodenumber)+".patch"+strconv.Itoa(patchnumber)
}

func (this *FD)LoadPointerMap() error {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.updateChainLock.Lock()
    defer this.updateChainLock.Unlock()

    if this.nextAvailablePosition>=0 {
        return nil
    }

    var tmpPos=0
    for {
        tMeta, tErr:=this.io.Getinfo(this.GetPatchName(tmpPos, -1))
        if tErr!=nil {
            return tErr
        }
        if tMeta==nil {
            // the file does not exist
            this.nextAvailablePosition=tmpPos
            return nil
        }
        if tNum, ok:=tMeta[INTRA_PATCH_METAKEY_NEXT_PATCH]; !ok {
            Secretary.WarnD("File "+this.filename+"'s patch #"+tmpPos+" has broken/invalid metadata. All the patches after it will get lost.")
            this.nextAvailablePosition=tmpPos+1
            return nil
        } else {
            tmpPos, tErr=strconv.Atoi(tNum)
            // TODO: consider add it into merge list
        }
    }

    return nil
}
func (this *FD)Submit(object *filetype.Kvmap) error {
    this.updateChainLock.RLock()
    defer this.updateChainLock.RUnlock()

    if this.nextAvailablePosition<0 {
        panic("Fatal logic error!")
    }
    var err=this.io.Put(this.GetPatchName(this.nextAvailablePosition, -1),
                object,
                FileMeta(map[string]string {
                    INTRA_PATCH_METAKEY_NEXT_PATCH: strconv.Itoa(this.nextAvailablePosition+1),
                    METAKEY_TIMESTAMP: GetTimestamp().String()
                })
    )
    if err!=nil {
        Secretary.Warn("Fail in putting file "+this.GetPatchName(this.nextAvailablePosition, -1), "FD.Submit()")
        return err
    }
    this.nextAvailablePosition++
    // TODO: consider add it into merge list
}

package distributedvc

import (
    "sync"
    _ "logger"
    "kernel/filetype"
    . "outapi"
    "strconv"
    . "definition/configinfo"
    . "utils/timestamp"
    "errors"
)

type FD struct {
    /*====BEGIN: for fdPool====*/
    lock *sync.Mutex

    filename string
    io Outapi
    reader int
    peeper int

    // 1 for active, 2 for dormant, 0 for uninited, -1 for dead
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
    nextToBeMerge int
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
        nextToBeMerge: -1,
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
    this.status=-1

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

// If not active yet, will fetch the data from storage.
// With fetching failure, a nil will be returned.
// @ Must be Grasped Reader to use
func (this *FD)Read() *filetype.Kvmap {
    this.contentLock.RLock()
    var t=this.numberZero
    this.contentLock.RUnlock()
    if t!=nil {
        return t
    }
    if ReadInNumberZero()!=nil {
        return nil
    }
    this.contentLock.RLock()
    t=this.numberZero
    this.contentLock.RUnlock()
    return t
}

// Attentez: this method is asynchonously invoked
func (this *FD)GoGrasped() {
    this.LoadPointerMap()
}

// Attentez: this method is asynchonously invoked
func (this *FD)GoRead() {
    this.ReadInNumberZero()
}

// @ indeed static
func (this *FD)GetTSFromMeta(meta FileMeta) ClxTimestamp {
    if tTS, tOK:=meta[METAKEY_TIMESTAMP]; !tOK {
        Secretary.WarnD("File "+this.filename+"'s patch #0 has invalid timestamp.")
        return 0
    } else {
        return String2ClxTimestamp(tTS)
    }
}
var READ_ZERO_NONEXISTENCE=errors.New("Patch#0 does not exist.")
// @ Must be Grasped Reader to use
func (this *FD)ReadInNumberZero() error {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.contentLock.Lock()
    defer this.contentLock.Unlock()

    if this.numberZero!=nil {
        return nil
    }

    var tMeta, tFile, tErr=this.io.Get(this.GetPatchName(0, -1))
    if tErr!=nil {
        return tErr
    }
    if tFile==nil || tMeta==nil {
        return READ_ZERO_NONEXISTENCE
    }
    var tKvmap, ok=tFile.(*filetype.Kvmap)
    if !ok {
        Secretary.WarnD("File "+this.filename+"'s patch #0 has invalid filetype. Its content will get ignored.")
        this.numberZero=filetype.NewKvMap()
    } else {
        this.numberZero=tKvmap
    }
    this.numberZero.TSet(this.GetTSFromMeta(tMeta))
    if tNext, ok2:=tMeta[INTRA_PATCH_METAKEY_NEXT_PATCH]; !ok2 {
        Secretary.WarnD("File "+this.filename+"'s patch #0 has invalid next-patch. Its precedents will get ignored.")
        this.nextToBeMerge=1
    } else {
        if nextNum, errx:=strconv.Atoi(tNext); errx!=nil {
            Secretary.WarnD("File "+this.filename+"'s patch #0 has invalid next-patch. Its precedents will get ignored.")
            this.nextToBeMerge=1
        } else {
            this.nextToBeMerge=nextNum
        }
    }
    this.status=1
    return nil
}

// @ Must be Grasped Reader to use
func (this *FD)MergeNext() error {
    var
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

// @ Get Normally Grasped
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
            if this.status<=0 {
                this.status=2
            }
            this.nextAvailablePosition=tmpPos
            return nil
        }
        if tNum, ok:=tMeta[INTRA_PATCH_METAKEY_NEXT_PATCH]; !ok {
            Secretary.WarnD("File "+this.filename+"'s patch #"+tmpPos+" has broken/invalid metadata. All the patches after it will get lost.")
            if this.status<=0 {
                this.status=2
            }
            this.nextAvailablePosition=tmpPos+1
            return nil
        } else {
            var oldPos=tmpPos
            tmpPos, tErr=strconv.Atoi(tNum)
            if tErr!=nil {
                Secretary.WarnD("File "+this.filename+"'s patch #"+tmpPos+" has broken/invalid metadata. All the patches after it will get lost.")
                tmpPos=oldPos
            }
            // TODO: consider add it into merge list
        }
    }

    return nil
}
// @ Get Normally Grasped
func (this *FD)Submit(object *filetype.Kvmap) error {
    this.updateChainLock.Lock()
    if this.nextAvailablePosition<0 {
        this.updateChainLock.Unlock()
        if err:=this.LoadPointerMap(); err!=nil {
            return nil
        }
        this.updateChainLock.Lock()
    }
    defer this.updateChainLock.Unlock()

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

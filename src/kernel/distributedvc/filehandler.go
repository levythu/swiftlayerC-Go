package distributedvc

import (
    "sync"
    . "logger"
    "kernel/filetype"
    . "kernel/distributedvc/filemeta"
    . "kernel/distributedvc/constdef"
    . "outapi"
    "strconv"
    . "definition/configinfo"
    . "utils/timestamp"
    "time"
    "errors"
)

/*
** FD: File Descriptor
** File Descriptor is the core data structure of the S-H2, which is responsible for
** directory meta info management.
** Each FD represents a separate directory meta and is unique in the memory. It controls
** submission of patches and auto-merging. Also, any LS operation will execute it
** to read & merge all the data, while notifying random number of peers to update their
** own patch chain.
** The first segment of member variables are used for fdPool to keep it unique and supporting
** automatically wiped out to control memory cost. It has several phases:
** 1. uninited phase:   neither the file content nor the chain info is loaded into memory
** 2. dormant phase:    when .grasp() gets invoked it will load chain info into memory,
**                      then functions like .MergeNext(), .ReadInNumberZero() and .Read()
**                      could get executed
** 3. active phase:     when .graspReader() gets invoked it will loadthe file into memory,
**                      then function .Submit could get executed.
** So always GetFD()->[GraspReader()->ReleaseReader()]->Release() in use
**
*/

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
    lastSyncTime int64
    latestReadableVersionTS ClxTimestamp
    modified bool
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
        lastSyncTime: 0,
        latestReadableVersionTS: 0,     // This version is for written version
        modified: false,
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

func (this *FD)__clearContentSansLock() {
    if this.status!=1 {
        return
    }

    this.contentLock.Lock()
    this.numberZero=nil
    this.nextToBeMerge=-1
    // consider write-back for unpersisted data
    this.contentLock.Unlock()

    this.status=2
}
func (this *FD)GoDie() {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.WriteBack()
    this.__clearContentSansLock()
    this.status=-1

    return
}
func (this *FD)GoDormant() {
    this.lock.Lock()
    defer this.lock.Unlock()
    this.isInDormant=false

    this.WriteBack()
    //logger.Secretary.LogD("Filehandler "+this.filename+" is going dormant.")
    this.__clearContentSansLock()
}

// If not active yet, will fetch the data from storage.
// With fetching failure, a nil will be returned.
// @ Must be Grasped Reader to use
func (this *FD)Read() (map[string]*filetype.KvmapEntry, error) {
    this.contentLock.RLock()
    var t=this.numberZero
    if t!=nil {
        var q=t.CheckOutReadOnly()
        this.contentLock.RUnlock()
        return q, nil
    }
    this.contentLock.RUnlock()

    if err:=this.ReadInNumberZero(); err!=nil {
        return nil, err
    }
    this.contentLock.RLock()
    defer this.contentLock.RUnlock()
    t=this.numberZero
    return t.CheckOutReadOnly(), nil
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
    this.modified=false
    return nil
}

var FORMAT_EXCEPTION=errors.New("Kvmap file not suitable.")
// if return (nil, nil), the file just does not exist.
// a nil for file and an error instance will be returned for other errors
// if the file is not nil, the function is invoked successfully
func readInKvMapfile(io Outapi, filename string) (*filetype.Kvmap, FileMeta, error) {
    var meta, file, err=io.Get(filename)
    if err!=nil {
        return nil, nil, err
    }
    if file==nil || meta==nil {
        Secretary.Log("distributedvc::readInKvMapfile()", "File "+filename+" does not exist.")
        return nil, nil, nil
    }
    var result, ok=file.(*filetype.Kvmap)
    if !ok {
        Secretary.Warn("distributedvc::readInKvMapfile()", "Fail in reading file "+filename)
        return nil, nil, FORMAT_EXCEPTION
    }

    if tTS, tOK:=meta[METAKEY_TIMESTAMP]; !tOK {
        Secretary.WarnD("File "+filename+"'s patch #0 has invalid timestamp.")
        result.TSet(0)
    } else {
        result.TSet(String2ClxTimestamp(tTS))
    }

    return result, meta, nil
}

var MERGE_ERROR=errors.New("Merging error")

// @ Must be Grasped Reader to use
var NOTHING_TO_MERGE=errors.New("Nothing to merge.")
func (this *FD)MergeNext() error {
    if tmpErr:=this.ReadInNumberZero(); tmpErr!=nil {
        return tmpErr
    }
    // Read one patch file , get ready for merge
    this.updateChainLock.RLock()
    var nextEmptyPatch=this.nextAvailablePosition
    this.updateChainLock.RUnlock()

    this.contentLock.Lock()
    defer this.contentLock.Unlock()

    if nextEmptyPatch==this.nextToBeMerge {
        return NOTHING_TO_MERGE
    }

    var oldMerged=this.nextToBeMerge
    var thePatch, meta, err=readInKvMapfile(this.io, this.GetPatchName(this.nextToBeMerge, -1))
    // may happen due to the unsubmission of Submit() function
    if thePatch==nil {
        Secretary.Warn("distributedvc::FD.MergeNext()", "Fail to get a supposed-to-be patch for file "+this.filename)
        if err==nil {
            return MERGE_ERROR
        } else {
            return err
        }
    }
    var theNext int
    if tNext, ok:=meta[INTRA_PATCH_METAKEY_NEXT_PATCH]; !ok {
        Secretary.Warn("distributedvc::FD.MergeNext()", "Fail to get INTRA_PATCH_METAKEY_NEXT_PATCH for file "+this.filename)
        theNext=this.nextToBeMerge+1
    } else {
        if intTNext, err:=strconv.Atoi(tNext); err!=nil {
            Secretary.Warn("distributedvc::FD.MergeNext()", "Fail to get INTRA_PATCH_METAKEY_NEXT_PATCH for file "+this.filename)
            theNext=this.nextToBeMerge+1
        } else {
            theNext=intTNext
        }
    }
    tNew, err:=this.numberZero.MergeWith(thePatch)
    if err!=nil {
        Secretary.Warn("distributedvc::FD.MergeNext()", "Fail to merge patches for file "+this.filename)
        return err
    }

    this.numberZero=tNew
    this.nextToBeMerge=theNext
    this.modified=true

    Secretary.Log("distributedvc::FD.MergeNext()", "Successfully merged in patch #"+strconv.Itoa(oldMerged)+" for "+this.filename)
    return nil
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

func (this *FD)GetPatchName(patchnumber int, nodenumber int/*-1*/) string {
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
    var needMerge=false
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
            if needMerge {
                MergeManager.SubmitTask(this.filename, this.io)
            }
            return nil
        }
        if tNum, ok:=tMeta[INTRA_PATCH_METAKEY_NEXT_PATCH]; !ok {
            Secretary.WarnD("File "+this.filename+"'s patch #"+strconv.Itoa(tmpPos)+" has broken/invalid metadata. All the patches after it will get lost.")
            if this.status<=0 {
                this.status=2
            }
            this.nextAvailablePosition=tmpPos+1
            return nil
        } else {
            var oldPos=tmpPos
            tmpPos, tErr=strconv.Atoi(tNum)
            if tErr!=nil {
                Secretary.WarnD("File "+this.filename+"'s patch #"+strconv.Itoa(tmpPos)+" has broken/invalid metadata. All the patches after it will get lost.")
                tmpPos=oldPos
            } else {
                if oldPos!=0 {
                    needMerge=true
                }
            }
        }
    }

    return nil
}

// object need not have its Timestamp set, 'cause the function will set it to
// the current systime
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
    var nAP=this.nextAvailablePosition
    this.nextAvailablePosition=nAP+1
    this.updateChainLock.Unlock()

    var selfName=CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(NODE_NUMBER)
    var nowTime=GetTimestamp()
    if object.Kvm==nil {
        object.CheckOut()
    }
    object.Kvm[selfName]=&filetype.KvmapEntry {
        Key: selfName,
        Val: "",
        Timestamp: nowTime,
    }
    object.CheckIn()

    var err=this.io.Put(this.GetPatchName(nAP, -1),
                object,
                FileMeta(map[string]string {
                    INTRA_PATCH_METAKEY_NEXT_PATCH: strconv.Itoa(nAP+1),
                    METAKEY_TIMESTAMP: nowTime.String(),
                }))
    if err!=nil {
        Secretary.Warn("distributedvc::FD.Submit()", "Fail in putting file "+this.GetPatchName(nAP, -1))
        go (func() {
            // failure rollback
            this.updateChainLock.Lock()
            if nAP+1==this.nextAvailablePosition {
                // up to now, no new patch has been submitted. Just rollback the number.
                this.nextAvailablePosition--
                this.updateChainLock.Unlock()
                return
            }
            this.updateChainLock.Unlock()
            Secretary.Error("distributedvc::FD.Submit()", "Submission gap occurs! Trying to fix it: "+this.GetPatchName(nAP, -1)+" TRIAL ")

            //TODO: write in auto fix local log.

        })()
        return err
    }

    if nAP!=1 {
        MergeManager.SubmitTask(this.filename, this.io)
    }

    return nil
}

const CONF_FLAG_PREFIX="/*CONF-FLAG*/"
// NOT for header, so can be camaralized
const NODE_SYNC_TIME_PREFIX="Node-Sync-"

// Will not require any lock in the process, so the function invocation must be
// strictly protected by lock in caller.
// @ only can be invoked by this.Sync()
func (this *FD)combineNodeX(nodenumber int) error {
    if nodenumber==NODE_NUMBER {
        return nil
    }
    // First, check whether the corresponding version exists or newer than currently
    // merged version.
    var keyStoreName=CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(nodenumber)
    if this.numberZero.Kvm==nil {
        this.numberZero.CheckOut()
    }
    var lastTime ClxTimestamp
    if elem, ok:=this.numberZero.Kvm[keyStoreName]; ok {
        lastTime=elem.Timestamp
    } else {
        lastTime=0
    }
    if lastTime>0 {
        var meta, err=this.io.Getinfo(this.GetPatchName(0, nodenumber))
        if meta==nil || err!=nil {
            // The file does not exist. Combining ends.
            return nil
        }
        var res, ok=meta[METAKEY_TIMESTAMP]
        if !ok {
            // The file does not exist. Combining ends.
            return nil
        }
        var existRecentTS=String2ClxTimestamp(res)
        if existRecentTS<=lastTime {
            // no need to fetch the file
            return nil
        }
    }

    var file, _, err=readInKvMapfile(this.io, this.GetPatchName(0, nodenumber))
    if err!=nil {
        return err
    }
    if file==nil {
        return nil
    }
    this.numberZero.MergeWith(file)
    this.numberZero.CheckOut()
    var newTS=GetTimestamp()
    var selfName=CONF_FLAG_PREFIX+NODE_SYNC_TIME_PREFIX+strconv.Itoa(NODE_NUMBER)
    this.numberZero.Kvm[selfName]=&filetype.KvmapEntry {
        Key: selfName,
        Val: "",
        Timestamp: newTS,
    }
    this.numberZero.TSet(newTS)
    this.numberZero.CheckIn()
    this.modified=true

    return nil
}
// Read and combine all the version from other nodes, providing the combined version.
// @ Get Reader Grasped
func (this *FD)Sync() error {
    this.ReadInNumberZero()
    this.contentLock.Lock()
    defer this.contentLock.Unlock()

    if this.lastSyncTime+SINGLE_FILE_SYNC_INTERVAL_MIN>time.Now().Unix() {
        // interval is too small, abort the sync.
        return nil
    }

    if this.numberZero==nil {
        this.numberZero=filetype.NewKvMap()
        this.nextToBeMerge=1
    }
    for searchNode:=0; searchNode<NODE_NUMS_IN_ALL; searchNode++ {
        this.combineNodeX(searchNode)
    }
    this.lastSyncTime=time.Now().Unix()

    return nil
}

// can be invoked after MergeWith(), Sync() or the moment that the FD goes dormant.
// @ async
func (this *FD)WriteBack() error {
    this.contentLock.Lock()
    defer this.contentLock.Unlock()

    if this.numberZero==nil {
        return nil
    }
    if !this.modified {
        return nil
    }

    var meta4Set=NewMeta()
    meta4Set[METAKEY_TIMESTAMP]=this.numberZero.TGet().String()
    meta4Set[INTRA_PATCH_METAKEY_NEXT_PATCH]=strconv.Itoa(this.nextToBeMerge)
    if err:=this.io.Put(this.GetPatchName(0, -1), this.numberZero, meta4Set); err!=nil {
        return err
    }

    this.modified=false
    this.latestReadableVersionTS=this.numberZero.TGet()

    return nil
}

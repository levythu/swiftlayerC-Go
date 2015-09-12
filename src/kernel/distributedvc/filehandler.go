package distributedvc

/*
    Kernel descriptor of one file(patch excluded), the filename should be unique both in swift and in
    memory, thus providing exclusive control on it.
    Responsible for scheduling intra- and inter- node merging work.
    - filename: the filename in SWIFT OBJECT
    - io: storage io interface

    Attentez: when Construct with a stream, its get all the data from the stream and writeBack returns
*/

import (
    "utils/datastructure/syncdict"
    "kernel/filetype"
    "outapi"
    "definition/configinfo"
    "strconv"
    "errors"
    . "kernel/distributedvc/filemeta"
    "sync"
)

type Fd struct {
    filename string
    io outapi.Outapi
    metadata FileMeta
    intravisor *IntramergeSupervisor
    intervisor *IntermergeSupervisor
    latestPatch int

    locks []*sync.Mutex
}

global_file_dict:=NewSyncdict()

// METADATA must not contain "_" and only lowercase is permitted
// ============Constants in the mainfile's metadata=============
METAKEY_TIMESTAMP="timestamp"
METAKEY_TYPE="typestamp"

// ============Constants in the intra-patch's metadata=============
INTRA_PATCH_METAKEY_NEXT_PATCH="next-patch"

// ============Constants in the inter-patch's metadata=============
INTER_PATCH_METAKEY_SYNCTIME1="sync-time-l"
INTER_PATCH_METAKEY_SYNCTIME2="sync-time-r"

CANONICAL_VERSION_METAKEY_SYNCTIME="sync-time"

func GetFD(fn string, _io outapi.Outapi) *Fd {
    return global_file_dict.Declare(fn+_io.generateUniqueID(), &Fd{
        filename: fn,
        io: _io,
        latestPatch: -10,
        // The number of locks may be changed here.
        locks: []*sync.Mutex{&sync.Mutex{},&sync.Mutex{},&sync.Mutex{}},
    })
}

func (this *Fd)GetPatchName(patchnumber int, nodenumber int) string {
    if nodenumber<0 {
        nodenumber=int(configinfo.GetProperty_Node("node_number").(float64))
    }
    return this.filename+".proxy"+strconv.Itoa(nodenumber)+".patch"+strconv.Itoa(patchnumber)
}

func (this *Fd)GetGlobalPatchName(splittreeid uint32) string {
    return this.filename+".splittree"+strconv.FormatUint(uint64(splittreeid),10)+".patch"
}

func (this *Fd)GetCanonicalVersionName() string {
    return this.filename+".cversion"
}

// @Sync(0)
func (this *Fd)GetLatestPatch() int {
    // LatestPatch means the next available PatchID
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    if this.latestPatch==-10 {
        prg:=0
        prgto, err:=this.io.getinfo(this.GetPatchName(prg))
        if err!=nil {
            return -10
        }
        for prgto!=nil {
            nprg:=strconv.ParseInt(prgto[INTRA_PATCH_METAKEY_NEXT_PATCH])
            this.intravisor.AnnounceNewTask(prg,nprg)       // Attetez: may announce empty file (nprg)
            prg=nprg
            prgto, err:=this.io.getinfo(this.GetPatchName(prg))
            if err!=nil {
                return -10
            }
        }
        this.latestPatch=prg
        this.intravisor.BatchWorker(-1, -1)
    }
    return this.latestPatch
}

// Pay attention that this commit donot support streaming.
// Given that it is mostly used for folder patch. If there's need for streaming,
// it will be added in the future.
// @Sync(1)
func (this *Fd)CommitPatch(patchfile filetype.Filetype) err {
    this.locks[1].Lock()
    defer this.locks[1].Unlock()

    latestAvailable:=GetLatestPatch()
    if latestAvailable<0 {
        return errors.New(exception.EX_FAIL_TO_FETCH_INTRALINK)
    }
    mata:=NewMeta()
    meta[INTRA_PATCH_METAKEY_NEXT_PATCH]=strconv.FormatInt(int64(latestAvailable+1), 10)
    this.io.Put(this.GetPatchName(latestAvailable), patchfile, meta)
    this.latestPatch++
    this.intravisor.AnnounceNewTask(latestAvailable, latestAvailable+1)

    if this.latestPatch==1 {
        // TODO: start intermerge.
    } else {
        this.intravisor.BatchWorker(-1, -1)
    }
    return nil
}

// Get the whole file of latest version. The order is, canonical version, then original
// file. If neither of them exists, a nil will be returned indicating the file not
// existing.
func (this *Fd)GetFile() filetype.Filetype {
    tFile, _, err:=this.io.Get(this.GetCanonicalVersionName())
    if tFile==nil || err!=nil {
        tFile, _, err=this.io.Get(this.filename)
        if tFile==nil || err!=nil {
            return nil
        }
    }
    return tFile
}

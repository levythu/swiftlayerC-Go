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
    "outapi"
    "definition/configinfo"
    "strconv"
    . "kernel/distributedvc/filemeta"
    "sync"
)

type fd struct {
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

func GetFD(fn string, _io outapi.Outapi) *fd {
    return global_file_dict.Declare(fn+_io.generateUniqueID(), &fd{
        filename: fn,
        io: _io,
        latestPatch: -1,
        // The number of locks may be changed here.
        locks: []*sync.Mutex{&sync.Mutex{},&sync.Mutex{},&sync.Mutex{}},
    })
}

func (this *fd)GetPatchName(patchnumber int, nodenumber int) string {
    if nodenumber<0 {
        nodenumber=configinfo.GetProperty_Node("node_number")
    }
    return this.filename+".proxy"+strconv.Itoa(nodenumber)+".patch"+strconv.Itoa(patchnumber)
}

func (this *fd)GetGlobalPatchName(splittreeid uint32) string {
    return this.filename+".splittree"+strconv.FormatUint(uint64(splittreeid),10)+".patch"
}

func (this *fd)GetCanonicalVersionName() string {
    return this.filename+".cversion"
}

// @Sync(0)
func (this *fd)GetLatestPatch() int {
    this.locks[0].Lock()
    defer this.locks[0].Unlock()

    if this.latestPatch==-1 {
        prg:=0
        prgto, err:=this.io.getinfo(this.GetPatchName(prg))
        if err!=nil {
            return -1
        }
        for prgto!=nil {
            // TODO
        }
    }
}

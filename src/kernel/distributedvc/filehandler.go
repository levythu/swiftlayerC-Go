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
    . "kernel/distributedvc/filemeta"
    "sync"
    "definition/exception"
    "fmt"
    "logger"
    "io"
    "kernel/distributedvc/constdef"
    "container/list"
    "utils/iomidware"
)

type Fd struct {
    filename string
    io outapi.Outapi
    metadata FileMeta
    intravisor *IntramergeSupervisor
    intervisor *IntermergeSupervisor
    latestPatch int
    grabbed int
    uselessPosition *list.Element

    locks []*sync.Mutex
}

func ____no_use() {
    fmt.Println("Nouse")
}

var global_file_dict=syncdict.NewSyncdict()

// METADATA must not contain "_" and only lowercase is permitted
// ============Constants in the mainfile's metadata=============
const METAKEY_TIMESTAMP=constdef.METAKEY_TIMESTAMP
const METAKEY_TYPE=constdef.METAKEY_TYPE

// ============Constants in the intra-patch's metadata=============
const INTRA_PATCH_METAKEY_NEXT_PATCH="next-patch"

// ============Constants in the inter-patch's metadata=============
const INTER_PATCH_METAKEY_SYNCTIME1="sync-time-l"
const INTER_PATCH_METAKEY_SYNCTIME2="sync-time-r"

const CANONICAL_VERSION_METAKEY_SYNCTIME="sync-time"

func GetFD(fn string, _io outapi.Outapi) *Fd {
    ret:=&Fd{
        filename: fn,
        io: _io,
        latestPatch: -10,
        grabbed: 0,
        uselessPosition: nil,
        // The number of locks may be changed here.
        locks: []*sync.Mutex{&sync.Mutex{},&sync.Mutex{},&sync.Mutex{},&sync.Mutex{}},
    }
    ret.intervisor=NewIntermergeSupervisor(ret)
    ret.intravisor=NewIntramergeSupervisor(ret)
    var result=global_file_dict.Declare(fn+_io.GenerateUniqueID(), ret).(*Fd)
    result.Grab()
    return result
}

func (this *Fd)GetPatchName(patchnumber int, nodenumber int/*-1*/) string {
    if nodenumber<0 {
        nodenumber=configinfo.NODE_NUMBER
    }
    return this.filename+".proxy"+strconv.Itoa(nodenumber)+".patch"+strconv.Itoa(patchnumber)
}

// @Sync(3)
func (this *Fd)Grab() bool {
    this.locks[3].Lock()
    defer this.locks[3].UnLock()
    if this.grabbed<0 {
        return false
    }
    this.grabbed++
    return true
}
// @Sync(3)
func (this *Fd)Release() {
    this.locks[3].Lock()
    this.grabbed--
    this.locks[3].UnLock()

    this.checkInactive();
}
// @Sync(3)
func (this *Fd)checkInactive() {
    // Only run when grabbed==0 and synced. So no more other calls are there.
    this.locks[3].Lock()
    defer this.locks[3].UnLock()
    if this.grabbed!=0 {
        return
    }

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
        prgto, err:=this.io.Getinfo(this.GetPatchName(prg, -1))
        if err!=nil {
            return -10
        }
        for prgto!=nil {
            nprg, _:=strconv.Atoi(prgto[INTRA_PATCH_METAKEY_NEXT_PATCH])
            this.intravisor.AnnounceNewTask(prg,nprg)       // Attetez: may announce empty file (nprg)
            prg=nprg
            prgto, err=this.io.Getinfo(this.GetPatchName(prg, -1))
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
func (this *Fd)CommitPatch(patchfile filetype.Filetype) error {
    this.locks[1].Lock()
    defer this.locks[1].Unlock()

    latestAvailable:=this.GetLatestPatch()
    if latestAvailable<0 {
        return exception.EX_FAIL_TO_FETCH_INTRALINK
    }
    meta:=NewMeta()
    meta[INTRA_PATCH_METAKEY_NEXT_PATCH]=strconv.FormatInt(int64(latestAvailable+1), 10)
    this.io.Put(this.GetPatchName(latestAvailable, -1), patchfile, meta)
    this.latestPatch++
    this.intravisor.AnnounceNewTask(latestAvailable, latestAvailable+1)
    //fmt.Println("123")
    if this.latestPatch==1 {
        this.intervisor.PropagateUp()
    } else {
        this.intravisor.BatchWorker(-1, -1)
    }
    return nil
}

// Get the whole file of latest version. The order is, canonical version, then original
// file. If neither of them exists, a nil will be returned indicating the file not
// existing.
func (this *Fd)GetFile() filetype.Filetype {
    _, tFile, err:=this.io.Get(this.GetCanonicalVersionName())
    if tFile==nil || err!=nil {
        _, tFile, err=this.io.Get(this.filename)
        if tFile==nil || err!=nil {
            return nil
        }
    }
    return tFile
}

// @Sync(2)
func (this *Fd)PutOriginalFile(content filetype.Filetype, meta FileMeta/*=nil*/) error {
    this.locks[2].Lock()
    defer this.locks[2].Unlock()

    if meta==nil {
        meta=NewMeta()
    }
    if err:=this.io.Put(this.filename, content, meta); err!=nil {
        return err
    }
    logger.Secretary.Log("kernel.dvc.fd::PutOriginalFile", "Upload original file successfully")
    return nil
}

// @Sync(2)
func (this *Fd)PutOriginalFileStream(meta FileMeta) (io.WriteCloser, error) {
    this.locks[2].Lock()

    var wc, err=this.io.PutStream(this.filename, meta)
    if err!=nil {
        this.locks[2].Unlock()
        return nil, err
    }

    var hackedwc=iomidware.NewWritecloserHackerPost(wc, func(tErr error) error {
        this.locks[2].Unlock()
        return tErr
    })
    return hackedwc, nil
}

// nil will be returned indicating that the file does not exist
func (this *Fd)GetFileStream() (io.ReadCloser, FileMeta, error) {
    fm, rc, err:=this.io.GetStream(this.GetCanonicalVersionName())
    if rc==nil || err!=nil {
        fm, rc, err=this.io.GetStream(this.filename)
        if rc==nil || err!=nil {
            return nil, nil, nil
        }
    }
    return rc, fm, nil
}

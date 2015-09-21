package distributedvc

import (
    "utils/datastructure/splittree"
    "definition/configinfo"
    "kernel/filetype"
    . "utils/timestamp"
    . "kernel/distributedvc/filemeta"
    "sync"
)

/*  Class for inter-node merge's sake. Considering all the nodes and build a segment
**  tree on them. Each node's data changed, propagate its change in the segment tree
**  from bottom to up.
**  Attentez: There exist some risks - when two nodes are trying to modify one segment-
**  tree-node simultaneously, an unexpected result may occur: an earlier one may override
**  the later one. So periodic overhaul or correction is needed.
**
**  It's identical between different nodes.
*/

var rootnodeid=int(splittree.GetRootLable(uint32(configinfo.GetProperty_Node("node_nums_in_all").(float64))))
var overhaulOrder=func() []int {
    ret:=[]int{}
    splittree.Traverse(uint32(configinfo.GetProperty_Node("node_nums_in_all").(float64)), func(nodeid uint32, layer uint32){
        ret=append(ret, int(nodeid))
    })
    return ret
}()

type IntermergeWorker struct {
    supervisor *IntermergeSupervisor
    fd *Fd
    pinpoint int
}
// If pinpoint==-1, use nodenumber as default pinpoint
func NewIntermergeWorker(_supervisor *IntermergeSupervisor, _pinpoint int/*=-1*/) *IntermergeWorker {
    var ret IntermergeWorker
    ret.supervisor=_supervisor
    ret.fd=_supervisor.filed
    if _pinpoint==-1 {
        ret.pinpoint=int(configinfo.GetProperty_Node("node_number").(float64))
    } else {
        ret.pinpoint=_pinpoint
    }

    return &ret
}

type FetchRecord struct {
    file filetype.Filetype
    fetchTime uint64
}

// Read info in one split-tree node. For leaf nodes, return its intra-node top patch.
func (this *IntermergeWorker)ReadInfo(nodeid int) (*FetchRecord, error) {
    var filename string
    if splittree.IsLeaf(uint32(nodeid)) {
        filename=this.fd.GetPatchName(0, int(splittree.FromLeaftoNode(uint32(nodeid))))
    } else {
        filename=this.fd.GetGlobalPatchName(uint32(nodeid))
    }

    _, file, err:=this.fd.io.Get(filename)
    if err!=nil {
        return nil, err
    }
    if file==nil {
        // The file do not exist, return a nonexist with max timestamp
        file=filetype.MAX_NONEXIST
    }

    return &FetchRecord{
        file: file,
        fetchTime: GetABSTimestamp(),
    }, nil
}

// Glean info from one nodes on splittree. It may be combined from its children
// or just come from one single child. If nil is returned, some error has happened
// and merge work should be terminated.
func (this *IntermergeWorker)GleanInfo(nodeid int, cacheDict map[int]*FetchRecord/*=nil*/) (*FetchRecord, uint64, uint64) {
    if cacheDict==nil {
        cacheDict=map[int]*FetchRecord{}
    }
    if splittree.IsLeaf(uint32(nodeid)) {
        if elem, ok:=cacheDict[nodeid]; ok {
            return elem, elem.fetchTime, elem.fetchTime
        }
        res, err:=this.ReadInfo(nodeid)
        if err!=nil {
            return nil, 0, 0
        }
        return res, res.fetchTime, res.fetchTime
    }

    resl, ok:=cacheDict[int(splittree.Left(uint32(nodeid)))]
    if !ok {
        var errl error
        resl, errl=this.ReadInfo(int(splittree.Left(uint32(nodeid))))
        if errl!=nil {
            return nil, 0, 0
        }
    }

    resr, ok:=cacheDict[int(splittree.Right(uint32(nodeid)))]
    if !ok {
        var errr error
        resr, errr=this.ReadInfo(int(splittree.Right(uint32(nodeid))))
        if errr!=nil {
            return nil, 0, 0
        }
    }

    combined, err:=resl.file.MergeWith(resr.file)
    if err!=nil {
        return nil, 0, 0
    }
    if filetype.IsNonexist(combined) {
        return nil, 0, 0
    }

    return &FetchRecord{
        file: combined,
        fetchTime: GetABSTimestamp(),   // Will be modified after rewrited.
    }, resl.fetchTime, resr.fetchTime
}

// Similar to GleanInfo, but datasource differs a little bit.
func (this *IntermergeWorker)MakeCanonicalVersion(cacheDict map[int]*FetchRecord/*=nil*/) {
    if cacheDict==nil {
        cacheDict=map[int]*FetchRecord{}
    }

    res, ok:=cacheDict[rootnodeid]
    if !ok {
        var err error
        res, err=this.ReadInfo(rootnodeid)
        if err!=nil {
            return
        }
    }
    if filetype.IsNonexist(res.file) {
        return
    }
    _, oriFile, err:=this.fd.io.Get(this.fd.filename)
    if oriFile!=nil && err==nil {
        res.file, err=oriFile.MergeWith(res.file)
        if err!=nil {
            return
        }
    }

    uploadMeta:=map[string]string{}
    uploadMeta[CANONICAL_VERSION_METAKEY_SYNCTIME]=ABSTimestamp2String(res.fetchTime)

    onlineMeta, err:=this.fd.io.Getinfo(this.fd.GetCanonicalVersionName())
    if err==nil {
        if elem, ok:=onlineMeta[CANONICAL_VERSION_METAKEY_SYNCTIME]; ok && String2ABSTimestamp(elem)>=res.fetchTime {
            // indicating the online version is newer than the local one. ABANDON submit.
            return
        }
    }

    if err=this.fd.io.Put(this.fd.GetCanonicalVersionName(), res.file, FileMeta(uploadMeta)); err!=nil {
        // TODO:Log the error
    }
}

// Responsible for getting gleaned data and update it to the server. If the version is out-of-date, returns
// nil to prevent further operation. With any error a nil will be returned, too.
// If the local version and the online one cannot override each other, reglean it until successfully.
func (this *IntermergeWorker)GleanAndUpdate(nodeid int, cacheDict map[int]*FetchRecord/*=nil*/) *FetchRecord {
    for {
        res, lt, rt:=this.GleanInfo(nodeid, cacheDict)
        if res==nil {
            return nil
        }
        if splittree.IsLeaf(uint32(nodeid)) {
            return res
        }
        uploadMeta:=map[string]string{}
        uploadMeta[INTER_PATCH_METAKEY_SYNCTIME1]=ABSTimestamp2String(lt)
        uploadMeta[INTER_PATCH_METAKEY_SYNCTIME2]=ABSTimestamp2String(rt)

        onlineMeta, err:=this.fd.io.Getinfo(this.fd.GetGlobalPatchName(uint32(nodeid)))
        if err==nil {
            onlt, ok1:=onlineMeta[INTER_PATCH_METAKEY_SYNCTIME1]
            onrt, ok2:=onlineMeta[INTER_PATCH_METAKEY_SYNCTIME2]
            if ok1 && ok2 {
                ltOnline:=String2ABSTimestamp(onlt)
                rtOnline:=String2ABSTimestamp(onrt)
                if ltOnline>=lt && rtOnline>=rt {
                    // Local version out of date. ABORT.
                    return nil
                }
                if ltOnline>lt || rtOnline>rt {
                    // Conflicted. Clear the cache and reglean data.
                    cacheDict=map[int]*FetchRecord{}
                    continue
                }
            }
        }
        // Check passed. Now update:
        // (Attentez: it is also probable that a out-of-date version be committed in high concurrency)
        if err=this.fd.io.Put(this.fd.GetGlobalPatchName(uint32(nodeid)), res.file, FileMeta(uploadMeta)); err!=nil {
            // TODO:Log the error
            return nil
        }
        return &FetchRecord{
            file: res.file,
            fetchTime: GetABSTimestamp(),
        }
    }
}

// Propagate modification from one node to the whole tree bottom-up.
// @Async
func (this *IntermergeWorker)BubbleUp() {
    cache:=map[int]*FetchRecord{}
    workNode:=this.pinpoint
    for {
        tmpRes:=this.GleanAndUpdate(workNode, cache)
        if tmpRes==nil {
            return
        }
        if workNode==rootnodeid {
            break
        }
        cache=map[int]*FetchRecord{}
        cache[workNode]=tmpRes
        workNode=int(splittree.Parent(uint32(workNode)))
    }
    this.MakeCanonicalVersion(cache)

    this.supervisor.ReportDeath(this)
}

/*
    Manager of intermerge workers. Comparing to intra-node merge, this process does
    not have lock and sync mechanism, being much easier.
    It has only two modes:
    - BubblePropagate Mode: the proxy has got some modification and it needs to propagate
    up on the tree.
    - Overhaul Mode: the content on the tree seems to be out-of-date, all the info should
    be gleaned another time, layer by layer.
*/

type IntermergeSupervisor struct {
    filed *Fd
    workersAlive int
    needsRefresh bool

    lock *sync.Mutex
}
func NewIntermergeSupervisor(filed *Fd) *IntermergeSupervisor {
    return &IntermergeSupervisor {
        filed: filed,
        workersAlive: 0,
        needsRefresh: false,
        lock: &sync.Mutex{},
    }
}

// @Sync
func (this *IntermergeSupervisor)ReportDeath(worker *IntermergeWorker) {
    this.lock.Lock()
    defer this.lock.Unlock()

    if this.workersAlive==1 && this.needsRefresh {
        go worker.BubbleUp()
        this.needsRefresh=false
    } else {
        this.workersAlive=this.workersAlive-1
    }
}
// @Sync
func (this *IntermergeSupervisor)PropagateUp() {
    this.lock.Lock()
    defer this.lock.Unlock()

    if this.workersAlive>0 {
        this.needsRefresh=true
        return
    }
    this.workersAlive=this.workersAlive+1
    worker:=NewIntermergeWorker(this, -1)
    go worker.BubbleUp()
}

func (this *IntermergeSupervisor)Overhaul() {
    go this._overhaul()
}

func (this *IntermergeSupervisor)_overhaul() {
    worker:=NewIntermergeWorker(this, -1)
    for _, elem:=range overhaulOrder {
        worker.GleanAndUpdate(elem, nil)
    }
}

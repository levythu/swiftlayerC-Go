package distributedvc

import (
    "utils/datastructure/splittree"
    "definition/configinfo"
    "kernel/filetype"
    . "utils/timestamp"
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

type IntermergeWorker struct {
    supervisor *IntermergeSupervisor
    fd *Fd
    pinpoint int
    isBubble bool
}
// If pinpoint==-1, use nodenumber as default pinpoint
func NewIntermergeWorker(_supervisor *IntermergeSupervisor, _pinpoint int/*=-1*/, _isBubble bool) *IntermergeWorker {
    var ret IntermergeWorker
    ret.supervisor=_supervisor
    ret.fd=_supervisor.filed
    if _pinpoint==-1 {
        ret.pinpoint=int(configinfo.GetProperty_Node("node_number").(float64))
    } else {
        ret.pinpoint=_pinpoint
    }
    ret.isBubble=_isBubble

    return &ret
}

type FetchRecord struct {
    file filetype.Filetype
    fetchTime uint64
}

// Read info in one split-tree node. For leaf nodes, return its intra-node top patch.
func (this *IntermergeWorker)ReadInfo(nodeid int) (*FetchRecord, error) {
    var filename string
    if splittree.IsLeaf(nodeid) {
        filename=this.fd.GetPatchName(0, int(splittree.FromLeaftoNode(uint32(nodeid))))
    } else {
        filename=this.fd.GetGlobalPatchName(nodeid)
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
func (this *IntermergeWorker)GleanInfo(nodeid int, cacheDict map[int]*FetchRecord/*=nil*/) (FetchRecord*, uint64, uint64) {
    if cacheDict==nil {
        cacheDict=map[int]*FetchRecord{}
    }
    if splittree.IsLeaf(nodeid) {
        if elem, ok:=cacheDict[nodeid]; ok {
            return elem, elem.fetchTime, elem.fetchTime
        }
        res, err:=this.ReadInfo(nodeid)
    }
}

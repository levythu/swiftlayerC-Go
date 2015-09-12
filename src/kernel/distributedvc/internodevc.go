package distributedvc

import (

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

type IntermergeWorker struct {
    
}

package uniqueid

import (
    "time"
    "strconv"
    "definition/configinfo"
)

const _BASESHOW=36

var lauchTime=strconv.FormatInt(time.Now().UnixNano(), _BASESHOW)
var nodeNum=strconv.FormatUint(uint64(configinfo.GetProperty_Node("node_number").(float64)), _BASESHOW)
var globalCounter=SyncCounter{counter:0}

func GenGlobalUniqueName() string  {
    return nodeNum+"~"+lauchTime+"~"+
        strconv.FormatInt(globalCounter.Inc(), _BASESHOW)
}

func GenGlobalUniqueNameWithTag(tag string) string  {
    return tag+"~"+nodeNum+"~"+lauchTime+"~"+
        strconv.FormatInt(globalCounter.Inc(), _BASESHOW)
}

package uniqueid

import (
    "time"
    "strconv"
    "definition/configinfo"
)

const _BASESHOW=36

var lauchTime=strconv.FormatInt(time.Now().UnixNano(), _BASESHOW)
var globalCounter=SyncCounter{counter:0}

func GenGlobalUniqueName() string  {
    return strconv.FormatUint(uint64(configinfo.Node_number()), _BASESHOW)+"~"+
        lauchTime+"~"+
        strconv.FormatInt(globalCounter.Inc(), _BASESHOW)
}

// Relative Timestamp is a 8-bytes integer (unsigned long), indeed, with
// higher 20 bits of version and lower 36 bits of time(UNIX TIME IN SECONDS).
// HIGHEST 7 bits are reserved.

package timestamp

import (
    "time"
    "strconv"
)

type ClxTimestamp uint64

func GetVersionNumber(timestamp ClxTimestamp) uint32 {
    return uint32((timestamp>>36)&0xfffff)
}

func GetExacttime(timestamp ClxTimestamp) uint64 {
    return uint64(timestamp)&0xfffffffff
}

func processRawTime(unixTime uint64) ClxTimestamp {
    unixTime=unixTime&0xfffffffff
    unixTime=0xfffffffff-unixTime
    return ClxTimestamp(unixTime)
}

func GetTimestamp(baseTime ClxTimestamp) ClxTimestamp {
    return (((ClxTimestamp(GetVersionNumber(baseTime))+1)&0xfffff)<<36)+processRawTime(uint64(time.Now().Unix()))
}

func String2ClxTimestamp(val string) ClxTimestamp {
    res, err:=strconv.ParseUint(val, 10, 64)
    if err!=nil {
        return 0
    }
    return ClxTimestamp(res)
}
func ClxTimestamp2String(val ClxTimestamp) string {
    return strconv.FormatUint(uint64(val), 10)
}

// identical to ClxTimestamp2String
func (this ClxTimestamp)String() string {
    return ClxTimestamp2String(this)
}

func MergeTimestamp(ts1, ts2 ClxTimestamp) ClxTimestamp {
    if ts1>ts2 {
        return ts1
    } else {
        return ts2
    }
}

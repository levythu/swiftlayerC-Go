// Relative Timestamp is a 8-bytes integer (unsigned long), indeed, with
// higher 20 bits of version and lower 36 bits of time(UNIX TIME IN SECONDS).
// HIGHEST 7 bits are reserved.

package timestamp

import (
    "time"
)

type ClxTimestamp uint64

func GetVersionNumber(timestamp ClxTimestamp) uint32 {
    return uint32((timestamp>>36)&0xfffff)
}

func GetExacttime(timestamp ClxTimestamp) uint64 {
    return uint64(timestamp)&0xfffffffff
}

func GetTimestamp(baseTime ClxTimestamp) ClxTimestamp {
    return (((ClxTimestamp(GetVersionNumber(baseTime))+1)&0xfffff)<<36)+ClxTimestamp(time.Now().Unix())
}

func MergeTimestamp(ts1, ts2 ClxTimestamp) ClxTimestamp {
    if ts1>ts2 {
        return ts1
    } else {
        return ts2
    }
}

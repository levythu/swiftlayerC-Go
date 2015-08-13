// Relative Timestamp is a 8-bytes integer (unsigned long), indeed, with
// higher 20 bits of version and lower 36 bits of time(UNIX TIME IN SECONDS).
// HIGHEST 7 bits are reserved.

package timestamp

import (
    "time"
)

func GetVersionNumber(timestamp uint64) uint32 {
    return uint32((timestamp>>36)&0xfffff)
}

func GetExacttime(timestamp uint64) uint64 {
    return timestamp&0xfffffffff
}

func GetTimestamp(baseTime uint64) uint64 {
    return (((uint64(GetVersionNumber(baseTime))+1)&0xfffff)<<36)+uint64(time.Now().Unix())
}

func MergeTimestamp(ts1, ts2 uint64) uint64 {
    if ts1>ts2 {
        return ts1
    } else {
        return ts2
    }
}

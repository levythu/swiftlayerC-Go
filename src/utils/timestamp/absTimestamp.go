package timestamp

import (
    "time"
)

func GetABSTimestamp() uint64 {
    return uint64(time.Now().UnixNano())
}

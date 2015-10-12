package filemeta

import (
    . "kernel/distributedvc/constdef"
)

type FileMeta map[string]string

func NewMeta() FileMeta {
    return FileMeta(map[string]string{})
}

func CheckIntegrity(obj FileMeta) bool {
    if _, ok:=obj[METAKEY_TIMESTAMP]; !ok {
        return false
    }
    if _, ok:=obj[METAKEY_TYPE]; !ok {
        return false
    }
    return true
}

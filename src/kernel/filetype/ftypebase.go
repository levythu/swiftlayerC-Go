package filetype

import (
    . "utils/timestamp"
    "io"
)

type Filetype interface {
    Init(dtSource io.Reader, dtTimestamp ClxTimestamp)
    WriteBack(dtDes io.Writer) error
    GetTS() ClxTimestamp
    SetTS(val ClxTimestamp)

    // The ret result is the merged version. However, the invocation of this func
    // may result in alternation in this*
    MergeWith(file2 Filetype) (Filetype, error)

    GetType() string
    EnsureRead() error
}

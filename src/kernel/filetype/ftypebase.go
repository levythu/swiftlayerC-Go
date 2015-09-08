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
}

// BLOB file. Only compare timestamp.

package filetype

import (
    "io"
    . "utils/timestamp"
    "reflect"
    "definition/exception"
    "errors"
)

type Blob struct {
    fileTS ClxTimestamp
    pointTo string
}


func (this *Blob)IsPointer() bool {
    return true
}
func (this *Blob)SetPointer(val string) {
    this.pointTo=val
}
func (this *Blob)GetPointer() string {
    return this.pointTo;
}


func (this *Blob)Init(dtSource io.Reader, dtTimestamp ClxTimestamp) {
    this.fileTS=dtTimestamp
}
func (this *Blob)GetTS() ClxTimestamp {
    return this.fileTS
}
func (this *Blob)SetTS(val ClxTimestamp) {
    this.fileTS=val
}
func (this *Blob)WriteBack(dtDes io.Writer) error {
    return nil
}

func (this *Blob)GetType() string {
    return "Binary Large OBject"
}
func (this *Blob)EnsureRead() error {
    return nil
}
func (this *Blob)MergeWith(file2 Filetype) (Filetype, error) {
    if IsNonexist(file2) {
        return this, nil
    }
    if reflect.TypeOf(this)!=reflect.TypeOf(file2) {
        return nil, errors.New(exception.EX_UNMATCHED_MERGE)
    }

    if this.fileTS>=(file2.(*Blob)).fileTS {
        return this, nil
    }
    return file2, nil
}

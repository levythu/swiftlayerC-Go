// nonexist file is a place holder file used to replace nil file.
// The behavior is, it will not read from any data source, Writeback will end in
// nothing either. It cannot be SetTs, and GetTs will return a fixed value predefined
// when constructed.
//
// If merged with any other kind of filetype q, q will be the result.

import (
    . "utils/timestamp"
    "io"
)

type Nonexist struct {
    fixedTS ClxTimestamp
}

func NewNonexist(ts ClxTimestamp) *Nonexist {
    return &Nonexist {
        fixedTS: ts
    }
}

var MAX_NONEXIST=NewNonexist()
var MIN_NONEXIST=NewNonexist(ClxTimestamp(^uint64(0)))

const NONEXIST_TYPESTAMP="Nonexist file"

func IsNonexist(this *NewNonexist) bool {
    return this.GetType()==NONEXIST_TYPESTAMP
}

func (this *NewNonexist)Init(_ io.Reader, _ ClxTimestamp) {
    return
}
func (this *NewNonexist)WriteBack(_ io.Writer) error {
    return nil
}
func (this *NewNonexist)GetTS() ClxTimestamp {
    return this.fixedTS
}
func (this *NewNonexist)SetTS(_ ClxTimestamp) {
    return
}
func (this *NewNonexist)MergeWith(file2 Filetype) (Filetype, error) {
    return file2, nil
}
func (this *NewNonexist)GetType() string {
    return NONEXIST_TYPESTAMP
}
func (this *NewNonexist)EnsureRead() error {
    return nil
}

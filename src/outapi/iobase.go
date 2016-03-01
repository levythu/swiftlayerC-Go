// Define the output interface to contect with Openstack Swift.
// Also, it can be rewrite to connect to local disk or other storage media.
// Pay attention that the input/output data is assembled, for streaming version
// Please refer to put/getStream

package outapi

import (
    "kernel/filetype"
    . "kernel/distributedvc/filemeta"
    "io"
)

type Outapi interface {

    GenerateUniqueID() string

    // Need not have typestamp in FileMeta. It will be set according to content's record automatically.
    // Filemeta could be nil.
    Put(filename string, content filetype.Filetype, info FileMeta) error

    // If file does not exist, a nil will be returned. No error occurs.
    Get(filename string) (FileMeta, filetype.Filetype, error)

    Putinfo(filename string, info FileMeta) error

    // If file does not exist, a nil will be returned. No error occurs.
    Getinfo(filename string) (FileMeta, error)

    Delete(filename string) error

    // If file does not exist, a nil will be returned. No error occurs.
    // Pay attention that io.ReadCloser should be closed.
    GetStream(filename string) (FileMeta, io.ReadCloser, error)

    PutStream(filename string, info FileMeta) (io.WriteCloser, error)

    // If the space is not available, create it and return (TRUE, nil);
    // If the space is already available, return (FALSE, nil);
    // Otherwise, (space is not available and fail to create), return a non-nil error.
    EnsureSpace() (bool, error)

}

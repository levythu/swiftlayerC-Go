// Define the output interface to contect with Openstack Swift.
// Also, it can be rewrite to connect to local disk or other storage media.
// Pay attention that the input/output data is assembled, for streaming version
// Please refer to put/getStream

package outapi

import (
    "kernel/filetype"
    . "kernel/distributedvc/filemeta"
)

type Outapi interface {

    GenerateUniqueID() string

    Put(filename string, content filetype.Filetype, info FileMeta) error

    // If file does not exist, a nil will be returned. No error occurs.
    Get(filename string) (FileMeta, filetype.Filetype, error)

    Putinfo(filename string, info FileMeta) error

    // If file does not exist, a nil will be returned. No error occurs.
    Getinfo(filename string) (FileMeta, error)

    Delete(filename string) error

    //TODO: Setup streaming api.
}

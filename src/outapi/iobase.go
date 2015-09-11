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
    generateUniqueID() string

    put(filename string, content filetype.Filetype, info FileMeta) error

    // If file does not exist, a nil will be returned. No error occurs.
    get(filename string) (FileMeta, filetype.Filetype, error)

    putinfo(filename string, info FileMeta) error

    // If file does not exist, a nil will be returned. No error occurs.
    getinfo(filename string) (FileMeta, error)

    delete(filename string) error

    //TODO: Setup streaming api.
}

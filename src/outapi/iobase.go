// Define the output interface to contect with Openstack Swift.
// Also, it can be rewrite to connect to local disk or other storage media.
// Pay attention that the input/output data is assembled, for streaming version
// Please refer to [TODO:Set it up!]

package outapi

import (
    "kernel/filetype"
)

type FileMeta map[string]string

type Outapi interface {
    generateUniqueID() string
    put(filename string, content filetype.Filetype, info FileMeta) error
    get(filename string) (FileMeta, filetype.Filetype, error)
    putinfo(filename string, info FileMeta) error
    getinfo(filename string) (FileMeta, error)
    delete(filename string) error
}

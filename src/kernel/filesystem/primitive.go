package filesystem

// the implementation of mimicking file system and index maintainance.
//
// Folder index file is a key-value map file that record the mapping from filename
// to its real name in swift, which could be either a file or another folder index
// file.
// Every search is from the root node '/', which name is fixed in the code. Pls pay attention
// not to modify it directly.
//
// Generally, a folder file's original file is empty(or nonexist), but patches help to
// maintain its real information.


import (
    "strings"
    "outapi"
    "definition/exception"
    dvc "kernel/distributedvc"
    "kernel/filetype"
)

// It is the primary function of filesystem, responsible for basic fs operation

// Check whether a direct filename/foldername is valid (not containing invalid chars)
var invalidSet=[]string{"/"}
func CheckValidFilename(filename string) bool {
    for _, e:=range invalidSet {
        if strings.Contains(filename, e) {
            return false
        }
    }
    return true
}

// From the parent inode, consult the vfilename and return its corresponding filename
// With any error, the string returned will be "".
// However, when the file does not exist, error WILL BE nil
func lookUp(inode string, vfilename string, io outapi.Outapi) (string, error) {
    if !CheckValidFilename(vfilename) {
        return "", exception.EX_INVALID_FILENAME
    }
    var tmp=dvc.GetFD(inode, io)
    inodefile, _:=tmp.GetFile().(*filetype.Kvmap)
    tmp.Release()
    if inodefile==nil {
        return "", exception.EX_INODE_NONEXIST
    }
    inodefile.CheckOut()
    elem, ok:=inodefile.Kvm[vfilename]
    if !ok {
        return "", nil
    }
    return elem.Val, nil
}

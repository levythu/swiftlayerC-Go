package filesystem
// Most impotant implementation of pseudo-filesystem layer.

import (
    "outapi"
    "definition/exception"
    dvc "kernel/distributedvc"
    "kernel/filetype"
    "utils/uniqueid"
    "strings"
    . "kernel/distributedvc/filemeta"
    . "kernel/distributedvc/constdef"
    . "logger"
    "io"
    "sync"
    "fmt"
)

const ROOT_INODE_NAME="rootNode"

type Fs struct {
    io outapi.Outapi
    rootName string

    cLock *sync.RWMutex
    trashInode string
}
func __nouse__() {
    fmt.Println("123")
}

func NewFs(_io outapi.Outapi) *Fs {
    return &Fs{
        io: _io,
        rootName: ROOT_INODE_NAME,

        cLock: &sync.RWMutex{},
        trashInode: "",
    }
}

const FOLDER_MAP="/$Folder-Map/"
const TRASH_BOX=".trash"

func (this *Fs)GetTrashInode() string {
    this.cLock.RLock()
    if t:=this.trashInode; t!="" {
        this.cLock.RUnlock()
        return t
    }
    this.cLock.RUnlock()
    this.cLock.Lock()
    defer this.cLock.Unlock()
    if this.trashInode!="" {
        return this.trashInode
    }
    // fetch it from storage
    var _, file, _=this.io.Get(GenFileName(this.rootName, TRASH_BOX))
    var filen, _=file.(*filetype.Nnode)
    if filen==nil {
        return ""
    } else {
        return filen.DesName
    }
}

//==============================================================================
// Followings are filesystem functions:

// path is a unix-like path string. If path starts with "/", search begins at
// root node. Otherwise in the frominode folder, when the frominode must exist.
// For any error, a blank string and error will be returned.
func (this *Fs)Locate(path string, frominode string/*=""*/) (string, error) {
    if strings.HasPrefix(path, "/") || frominode=="" {
        frominode=ROOT_INODE_NAME
    }
    var rawResult=strings.Split(path, "/")
    for _, e:=range rawResult {
        if e!="" {
            frominode, _=lookUp(frominode, e, this.io)
            if frominode=="" {
                // It is correct only to check result without referring to error.
                return "", exception.EX_FAIL_TO_LOOKUP
            }
        }
    }

    return frominode, nil
}

// If the file exist and forceMake==false, an error will be returned
func (this *Fs)Mkdir(foldername string, frominode string, forceMake bool) error {
    if !CheckValidFilename(foldername) {
        return exception.EX_INVALID_FILENAME
    }

    // nnodeName: parentInode::foldername
    var nnodeName=GenFileName(frominode, foldername)
    if !forceMake {
        if tmeta, _:=this.io.Getinfo(nnodeName); tmeta!=nil {
            return exception.EX_FOLDER_ALREADY_EXIST
        }
    }

    // newDomainname: <GENERATED>
    var newDomainname=uniqueid.GenGlobalUniqueName()
    var newNnode=filetype.NewNnode(newDomainname)
    if err:=this.io.Put(nnodeName, newNnode, nil); err!=nil {
        return err
    }
    // initialize two basic element
    if err:=this.io.Put(GenFileName(newDomainname, ".."), filetype.NewNnode(frominode), nil); err!=nil {
        Secretary.Error("kernel.filesystem::Mkdir", "Fail to create .. link for new folder "+nnodeName+".")
        return err
    }

    if err:=this.io.Put(GenFileName(newDomainname, "."), filetype.NewNnode(newDomainname), nil); err!=nil {
        Secretary.Error("kernel.filesystem::Mkdir", "Fail to create . link for new folder "+nnodeName+".")
        return err
    }

    // write new folder's map
    {
        var newFolderMapFD=dvc.GetFD(GenFileName(newDomainname, FOLDER_MAP), this.io)
        if newFolderMapFD==nil {
            Secretary.Error("kernel.filesystem::Mkdir", "Fail to create foldermap fd for new folder "+nnodeName+".")
            return exception.EX_IO_ERROR
        }
        if err:=newFolderMapFD.Submit(filetype.FastMake(".", "..")); err!=nil {
            Secretary.Error("kernel.filesystem::Mkdir", "Fail to init foldermap for new folder "+nnodeName+".")
            newFolderMapFD.Release()
            return err
        }
        newFolderMapFD.Release()
    }

    // submit patch to parent folder's map
    {
        var parentFolderMapFD=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
        if parentFolderMapFD==nil {
            Secretary.Error("kernel.filesystem::Mkdir", "Fail to create foldermap fd for new folder "+nnodeName+"'s parent map'")
            return exception.EX_IO_ERROR
        }
        if err:=parentFolderMapFD.Submit(filetype.FastMake(foldername)); err!=nil {
            Secretary.Error("kernel.filesystem::Mkdir", "Fail to submit patch to foldermap for new folder "+nnodeName+"'s parent map'")
            parentFolderMapFD.Release()
            return err
        }
        parentFolderMapFD.Release()
    }

    return nil
}

// Format the filesystem.
// TODO: Setup clear old fs info?
func (this *Fs)FormatFS() error {
    if err:=this.io.Put(GenFileName(this.rootName, ".."), filetype.NewNnode(this.rootName), nil); err!=nil {
        Secretary.Error("kernel.filesystem::FormatFS", "Fail to create .. link for Root.")
        return err
    }

    if err:=this.io.Put(GenFileName(this.rootName, "."), filetype.NewNnode(this.rootName), nil); err!=nil {
        Secretary.Error("kernel.filesystem::FormatFS", "Fail to create . link for Root.")
        return err
    }

    {
        var rootFD=dvc.GetFD(GenFileName(this.rootName, FOLDER_MAP), this.io)
        if rootFD==nil {
            Secretary.Error("kernel.filesystem::FormatFS", "Fail to get FD for Root.")
            return exception.EX_IO_ERROR
        }
        if err:=rootFD.Submit(filetype.FastMake(".", "..")); err!=nil {
            Secretary.Error("kernel.filesystem::FormatFS", "Fail to submit format patch for Root.")
            rootFD.Release()
            return nil
        }
        rootFD.Release()
    }
    // setup Trash for users
    return this.Mkdir(TRASH_BOX, this.rootName, true)
}

// Only returns file name list of one inode. Innername excluded.
func (this *Fs)List(frominode string) ([]string, error) {
    // TODO

    return nil, nil
}

// All the folder will be removed. No matter if it is empty or not.
// Move it to the trash
func (this *Fs)Rm(foldername string, frominode string) error {
    if !CheckValidFilename(foldername) {
        return exception.EX_INVALID_FILENAME
    }

    if tsinode:=this.GetTrashInode(); tsinode=="" {
        Secretary.ErrorD("IO: "+this.io.GenerateUniqueID()+" has an invalid trashbox, which leads to removing failure.")
        return exception.EX_TRASHBOX_NOT_INITED
    } else {
        return this.MvX(foldername, frominode, tsinode, uniqueid.GenGlobalUniqueNameWithTag("removed"), false)
        // TODO: logging the original position for recovery
    }
}

// Attentez: It is not atomic
// If byForce set to false and the destination file exists, an EX_FOLDER_ALREADY_EXIST will be returned
func (this *Fs)MvX(srcName, srcInode, desName, desInode string, byForce bool) error {
    // Create a mirror at destination position.
    // Then, remove the old one.
    // Third, modify the .. pointer.

    if !CheckValidFilename(srcName) || !CheckValidFilename(desName) {
        return exception.EX_INVALID_FILENAME
    }
    if !byForce && outapi.ForceCheckExist(this.io.CheckExist(GenFileName(desInode, desName))) {
        return exception.EX_FOLDER_ALREADY_EXIST
    }
    if err:=this.io.Copy(GenFileName(srcInode, srcName), GenFileName(desInode, desName), nil); err!=nil {
        Secretary.Error("kernel.filesystem::MvX", "Fail to issue a copy from "+GenFileName(srcInode, srcName)+" to "+GenFileName(desInode, desName))
        return err
    }

    {
        var desParentMap=dvc.GetFD(GenFileName(desInode, FOLDER_MAP), this.io)
        if desParentMap==nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to get foldermap fd for folder "+desInode)
            return exception.EX_IO_ERROR
        }
        if err:=desParentMap.Submit(filetype.FastMake(desName)); err!=nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to submit foldermap patch for folder "+desInode)
            desParentMap.Release()
            return err
        }
        desParentMap.Release()
    }

    // remove the old one.
    this.io.Delete(GenFileName(srcInode, srcName))

    {
        var srcParentMap=dvc.GetFD(GenFileName(srcInode, FOLDER_MAP), this.io)
        if srcParentMap==nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to get foldermap fd for folder "+srcInode)
            return exception.EX_IO_ERROR
        }
        if err:=srcParentMap.Submit(filetype.FastAntiMake(srcName)); err!=nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to submit foldermap patch for folder "+srcInode)
            srcParentMap.Release()
            return err
        }
        srcParentMap.Release()
    }

    // modify the .. pointer
    var _, dstFileNnodeOriginal, _=this.io.Get(GenFileName(desInode, desName))
    var dstFileNnode, _=dstFileNnodeOriginal.(*filetype.Nnode)
    if dstFileNnode==nil {
        Secretary.Error("kernel.filesystem::MvX", "Fail to read nnode "+GenFileName(desInode, desName)+".")
        return exception.EX_IO_ERROR
    }
    if err:=this.io.Put(GenFileName(dstFileNnode.DesName, ".."), filetype.NewNnode(desInode), nil); err!=nil {
        Secretary.Error("kernel.filesystem::MvX", "Fail to modify .. link for "+dstFileNnode.DesName+".")
        return err
    }

    // ALL DONE!
    return nil

}

// To put a large file and modify its corresponding index.
// Note that the function is synchronous, which means that it
// will block until data are fully written.
// It will try to put a file at destination, no matter whether
// there's already one file, which will be replaced then.

// if filename!="", a new filename will be assigned and frominode::filename will be set
// otherwise, frominode indicates the target fileinode and the target file will override it
const STREAM_TYPE="stream type file"
func (this *Fs)Put(filename string, frominode string, meta FileMeta/*=nil*/, dataSource io.Reader, typeoffile string/*=""*/) error {
    var targetFileinode string
    if filename!="" {
        if !CheckValidFilename(filename) {
            return exception.EX_INVALID_FILENAME
        }
        targetFileinode=uniqueid.GenGlobalUniqueNameWithTag("Stream")
        if err:=this.io.Put(GenFileName(frominode, filename), filetype.NewNnode(targetFileinode), nil); err!=nil {
            Secretary.Warn("kernel.filesystem::Put", "Put nnode for new file "+GenFileName(frominode, filename)+" failed.")
            return err
        }
    } else {
        targetFileinode=frominode
    }

    if meta==nil {
        meta=NewMeta()
    }
    meta=meta.Clone()
    meta[METAKEY_TYPE]=STREAM_TYPE
    if wc, err:=this.io.PutStream(targetFileinode, meta); err!=nil {
        Secretary.Error("kernel.filesystem::Put", "Put stream for new file "+GenFileName(frominode, filename)+" failed.")
        return err
    } else {
        if _, err2:=io.Copy(wc, dataSource); err2!=nil {
            wc.Close()
            Secretary.Error("kernel.filesystem::Put", "Piping stream for new file "+GenFileName(frominode, filename)+" failed.")
            return err2
        }
        if err2:=wc.Close(); err2!=nil {
            Secretary.Error("kernel.filesystem::Put", "Close writer for new file "+GenFileName(frominode, filename)+" failed.")
            return err2
        }
    }

    return nil
}

// If the file does not exist, an EX_FILE_NOT_EXIST will be returned.
func (this *Fs)Get(filename string, frominode string, w io.Writer) error {
    var targetFileinode string
    if filename!="" {
        if !CheckValidFilename(filename) {
            return exception.EX_INVALID_FILENAME
        }
        var _, file, _=this.io.Get(GenFileName(frominode, filename))
        var filen, _=file.(*filetype.Nnode)
        if filen==nil {
            return exception.EX_FILE_NOT_EXIST
        }
        targetFileinode=filen.DesName
    } else {
        targetFileinode=frominode
    }

    var meta, rc, _=this.io.GetStream(targetFileinode)
    if meta==nil || rc==nil {
        return exception.EX_FILE_NOT_EXIST
    }
    if val, ok:=meta[METAKEY_TYPE]; !ok || val!=STREAM_TYPE {
        rc.Close()
        return exception.EX_WRONG_FILEFORMAT
    }

    if _, copyErr:=io.Copy(w, rc); copyErr!=nil {
        rc.Close()
        return copyErr
    }
    if err2:=rc.Close(); err2!=nil {
        return err2
    }
    return nil
}

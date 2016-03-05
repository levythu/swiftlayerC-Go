package filesystem
// Most impotant implementation of pseudo-filesystem layer.

import (
    "outapi"
    "definition/exception"
    dvc "kernel/distributedvc"
    "kernel/filetype"
    "utils/uniqueid"
    "strings"
    . "utils/timestamp"
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
        trashInode: ""
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
        Secretary.ErrorD("kernel.filesystem::Mkdir", "Fail to create .. link for new folder "+nnodeName+".")
        return err
    }

    if err:=this.io.Put(GenFileName(newDomainname, "."), filetype.NewNnode(newDomainname), nil); err!=nil {
        Secretary.ErrorD("kernel.filesystem::Mkdir", "Fail to create . link for new folder "+nnodeName+".")
        return err
    }

    // write new folder's map
    {
        var newFolderMapFD=dvc.GetFD(GenFileName(newDomainname, FOLDER_MAP), this.io)
        if newFolderMapFD==nil {
            Secretary.ErrorD("kernel.filesystem::Mkdir", "Fail to create foldermap fd for new folder "+nnodeName+".")
            return exception.EX_IO_ERROR
        }
        if err:=newFolderMapFD.Submit(filetype.FastMake(".", "..")); err!=nil {
            Secretary.ErrorD("kernel.filesystem::Mkdir", "Fail to init foldermap for new folder "+nnodeName+".")
            newFolderMapFD.Release()
            return err
        }
        newFolderMapFD.Release()
    }

    // submit patch to parent folder's map
    {
        var parentFolderMapFD=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
        if parentFolderMapFD==nil {
            Secretary.ErrorD("kernel.filesystem::Mkdir", "Fail to create foldermap fd for new folder "+nnodeName+"'s parent map'")
            return exception.EX_IO_ERROR
        }
        if err:=parentFolderMapFD.Submit(filetype.FastMake(foldername)); err!=nil {
            Secretary.ErrorD("kernel.filesystem::Mkdir", "Fail to submit patch to foldermap for new folder "+nnodeName+"'s parent map'")
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
        Secretary.ErrorD("kernel.filesystem::FormatFS", "Fail to create .. link for Root.")
        return err
    }

    if err:=this.io.Put(GenFileName(this.rootName, "."), filetype.NewNnode(this.rootName), nil); err!=nil {
        Secretary.ErrorD("kernel.filesystem::FormatFS", "Fail to create . link for Root.")
        return err
    }

    {
        var rootFD=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
        if rootFD==nil {
            Secretary.ErrorD("kernel.filesystem::FormatFS", "Fail to get FD for Root.")
            return exception.EX_IO_ERROR
        }
        if err:=rootFD.Submit(filetype.FastMake(".", "..")); err!=nil {
            Secretary.ErrorD("kernel.filesystem::FormatFS", "Fail to submit format patch for Root.")
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
        Secretary.Error("IO: "+this.io.GenerateUniqueID()+" has an invalid trashbox, which leads to removing failure.")
        return exception.EX_TRASHBOX_NOT_INITED
    } else {
        return this.MvX(foldername, frominode, tsinode, uniqueid.GenGlobalUniqueNameWithTag("removed"))
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
    if !byForce && this.io.CheckExist(GenFileName(desInode, desName)) {
        return exception.EX_FOLDER_ALREADY_EXIST
    }
    if err:=this.io.Copy(GenFileName(srcInode, srcName), GenFileName(desInode, desName), nil); err!=nil {
        Secretary.ErrorD("kernel.filesystem::MvX", "Fail to issue a copy from "+GenFileName(srcInode, srcName)+" to "+GenFileName(desInode, desName))
        return err
    }

    {
        var desParentMap=dvc.GetFD(GenFileName(desInode, FOLDER_MAP), this.io)
        if desParentMap==nil {
            Secretary.ErrorD("kernel.filesystem::MvX", "Fail to get foldermap fd for folder "+desInode)
            return exception.EX_IO_ERROR
        }
        if err:=desParentMap.Submit(filetype.FastMake(desName)); err!=nil {
            Secretary.ErrorD("kernel.filesystem::MvX", "Fail to submit foldermap patch for folder "+desInode)
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
            Secretary.ErrorD("kernel.filesystem::MvX", "Fail to get foldermap fd for folder "+srcInode)
            return exception.EX_IO_ERROR
        }
        if err:=srcParentMap.Submit(filetype.FastAntiMake(srcName)); err!=nil {
            Secretary.ErrorD("kernel.filesystem::MvX", "Fail to submit foldermap patch for folder "+srcInode)
            srcParentMap.Release()
            return err
        }
        srcParentMap.Release()
    }

    // modify the .. pointer
    var _, dstFileNnodeOriginal, _=this.io.Get(GenFileName(desInode, desName))
    var dstFileNnode, _=dstFileNnodeOriginal.(*filetype.Nnode)
    if dstFileNnode==nil {
        Secretary.ErrorD("kernel.filesystem::MvX", "Fail to read nnode "+GenFileName(desInode, desName)+".")
        return exception.EX_IO_ERROR
    }
    if err:=this.io.Put(GenFileName(dstFileNnode.DesName, ".."), filetype.NewNnode(desInode), nil); err!=nil {
        Secretary.ErrorD("kernel.filesystem::MvX", "Fail to modify .. link for "+dstFileNnode.DesName+".")
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
func (this *Fs)Put(fiename string, frominode string, meta FileMeta/*=nil*/, dataSource io.Reader, typeoffile string/*=""*/) error {
    var basenode=frominode

    var newFileName=uniqueid.GenGlobalUniqueNameWithTag("Stream")
    var newFilefd=dvc.GetFD(newFileName, this.io)
    defer newFilefd.Release()
    if meta==nil {
        meta=NewMeta()
    }
    var newFilemeta=meta
    newFilemeta[METAKEY_TIMESTAMP]=GetTimestamp(0).String()
    newFilemeta[METAKEY_TYPE]=typeoffile
    wc, err:=newFilefd.PutOriginalFileStream(newFilemeta)
    if err!=nil {
        return err
    }

    // Streaming
    _, err=io.Copy(wc, dataSource)
    var err2=wc.Close()
    if err!=nil {
        return err
    }
    if err2!=nil {
        logger.Secretary.Error("kernel.filesystem.Fs::Put", "Error when closing: "+err2.Error())
        return err2
    }
    // Copy successfully. Update the index.

    var par=dvc.GetFD(basenode, this.io)
    defer par.Release()
    var flist, _=par.GetFile().(*filetype.Kvmap)
    if flist==nil {
        return exception.EX_INODE_NONEXIST
    }
    flist.CheckOut()

    var patcher=filetype.NewKvMap()
    patcher.SetTS(GetTimestamp(flist.GetTS()))
    patcher.CheckOut()
    patcher.Kvm[filename]=&filetype.KvmapEntry{
        Timestamp: GetTimestamp(flist.GetRelativeTS(filename)),
        Key: filename,
        Val: newFileName,
    }
    patcher.CheckIn()
    if err:=par.CommitPatch(patcher); err!=nil {
        return err
    }
    logger.Secretary.Log("kernel.filesystem.Fs::put", "Put stream file "+destination+" successfully.")

    return nil
}


// To get a large file by streaming.
// Attentez: it is a two-phase function.
// The first phase is try to locate the file in repository, and commit the possible
// error to phase1callback, which will determine the next step by returning a nil or
// available WriteCloser. In this phase an error code could be returned in the form
// of HTTP response code.
// The second phase is data transmission, by returning HTTP 200 the webserver just pipe
// data to the client. It will be run in another goroutine. So use it synchronously.

// Unlike Put, which handles upstream, Get func must use downstream to return error or
// valid stream. So the architectures are different.

// Phase1Callback is called when data transmission is ready.
type Phase1Callback func(error, FileMeta) io.Writer
// Phase2Callback is called when transmission completed.
type Phase2Callback func(error)

func (this *Fs)Get(source string, frominode string/*=""*/, phase1 Phase1Callback, phase2 Phase2Callback, isSychronous bool) {
    // Use this.locate, but not for finding folder. Note the semantic discrepancy.
    var obj, err=this.Locate(source, frominode)
    if err!=nil {
        phase1(exception.EX_FILE_NOT_EXIST, nil)
        return
    }

    var objFD=dvc.GetFD(obj, this.io)
    rc, fm, err:=objFD.GetFileStream()
    if err!=nil{
        phase1(err, nil)
        objFD.Release()
        return
    }
    if rc==nil {
        // 404
        phase1(exception.EX_FILE_NOT_EXIST, nil)
        objFD.Release()
        return
    }

    var wc=phase1(nil, fm)
    if wc==nil {
        // Phase1 requires to terminate.
        rc.Close()
        objFD.Release()
        return
    }

    if isSychronous {
        _, copyError:=io.Copy(wc, rc)
        rc.Close()
        objFD.Release()
        phase2(copyError)
    } else {
        go func() {
            _, copyError:=io.Copy(wc, rc)
            rc.Close()
            phase2(copyError)
            objFD.Release()
        }()
    }
}

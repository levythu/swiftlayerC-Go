package filesystem

// Paralleled version of fs functions. Also members of *Fs
// `Paralleled` mainly focuses on reducing io rtt, not on cutting time by parallel
// computing of required data.

import (
    "sync"
    "definition/errorgroup"
    "outapi"
    "definition/exception"
    dvc "kernel/distributedvc"
    "kernel/filetype"
    "utils/uniqueid"
    . "kernel/distributedvc/filemeta"
    . "logger"
)

// If the file exist and forceMake==false, an error EX_FOLDER_ALREADY_EXIST will be returned
func (this *Fs)MkdirParalleled(foldername string, frominode string, forceMake bool) error {
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
    var newDomainname=uniqueid.GenGlobalUniqueName()

    var globalError *errorgroup.ErrorAssembly=nil
    var geLock sync.Mutex
    var wg sync.WaitGroup
    // create target inode and write parent folder map
    var routineToCreateInode=func() {
        defer wg.Done()

        var newNnode=filetype.NewNnode(newDomainname)
        var initMeta=FileMeta(map[string]string {
            META_INODE_TYPE:    META_INODE_TYPE_FOLDER,
            META_PARENT_INODE:  frominode,
        })
        if err:=this.io.Put(nnodeName, newNnode, initMeta); err!=nil {
            geLock.Lock()
            globalError=errorgroup.AddIn(globalError, err)
            geLock.Unlock()
            return
        }

        // submit patch to parent folder's map
        {
            var parentFolderMapFD=dvc.GetFD(GenFileName(frominode, FOLDER_MAP), this.io)
            if parentFolderMapFD==nil {
                Secretary.Error("kernel.filesystem::MkdirParalleled", "Fail to create foldermap fd for new folder "+nnodeName+"'s parent map'")
                geLock.Lock()
                globalError=errorgroup.AddIn(globalError, exception.EX_IO_ERROR)
                geLock.Unlock()
                return
            }
            if err:=parentFolderMapFD.Submit(filetype.FastMake(foldername)); err!=nil {
                Secretary.Error("kernel.filesystem::MkdirParalleled", "Fail to submit patch to foldermap for new folder "+nnodeName+"'s parent map'")
                parentFolderMapFD.Release()
                geLock.Lock()
                globalError=errorgroup.AddIn(globalError, err)
                geLock.Unlock()
                return
            }
            parentFolderMapFD.Release()
        }
    }

    var initMetaSelf=FileMeta(map[string]string {
        META_INODE_TYPE:    META_INODE_TYPE_FOLDER,
        META_PARENT_INODE:  newDomainname,
    })
    var routineToCreateDotDot=func() {
        defer wg.Done()

        if err:=this.io.Put(GenFileName(newDomainname, ".."), filetype.NewNnode(frominode), initMetaSelf); err!=nil {
            Secretary.Error("kernel.filesystem::MkdirParalleled", "Fail to create .. link for new folder "+nnodeName+".")
            geLock.Lock()
            globalError=errorgroup.AddIn(globalError, err)
            geLock.Unlock()
            return
        }
    }
    var routineToCreateDot=func() {
        defer wg.Done()

        if err:=this.io.Put(GenFileName(newDomainname, "."), filetype.NewNnode(newDomainname), initMetaSelf); err!=nil {
            Secretary.Error("kernel.filesystem::MkdirParalleled", "Fail to create . link for new folder "+nnodeName+".")
            geLock.Lock()
            globalError=errorgroup.AddIn(globalError, err)
            geLock.Unlock()
            return
        }
    }

    var routineToWriteNewMap=func() {
        defer wg.Done()

        {
            var newFolderMapFD=dvc.GetFD(GenFileName(newDomainname, FOLDER_MAP), this.io)
            if newFolderMapFD==nil {
                Secretary.Error("kernel.filesystem::MkdirParalleled", "Fail to create foldermap fd for new folder "+nnodeName+".")
                geLock.Lock()
                globalError=errorgroup.AddIn(globalError, exception.EX_IO_ERROR)
                geLock.Unlock()
                return
            }
            if err:=newFolderMapFD.Submit(filetype.FastMake(".", "..")); err!=nil {
                Secretary.Error("kernel.filesystem::MkdirParalleled", "Fail to init foldermap for new folder "+nnodeName+".")
                newFolderMapFD.Release()
                geLock.Lock()
                globalError=errorgroup.AddIn(globalError, err)
                geLock.Unlock()
                return
            }
            newFolderMapFD.Release()
        }
    }

    wg.Add(4)
    go routineToCreateInode()
    go routineToCreateDotDot()
    go routineToCreateDot()
    go routineToWriteNewMap()
    wg.Wait()

    // NOW All the routines have returned and globalError stores all the errors
    // TODO: consider clearing roll-back

    return globalError
}

// All the folder will be removed. No matter if it is empty or not.
// Move it to the trash
func (this *Fs)RmParalleled(foldername string, frominode string) error {
    if tsinode:=this.GetTrashInode(); tsinode=="" {
        Secretary.ErrorD("IO: "+this.io.GenerateUniqueID()+" has an invalid trashbox, which leads to removing failure.")
        return exception.EX_TRASHBOX_NOT_INITED
    } else {
        return this.MvXParalleled(foldername, frominode, uniqueid.GenGlobalUniqueNameWithTag("removed"), tsinode, true)
        // TODO: logging the original position for recovery
    }
}

// Attentez: It is not atomic
// If byForce set to false and the destination file exists, an EX_FOLDER_ALREADY_EXIST will be returned
func (this *Fs)MvXParalleled(srcName, srcInode, desName, desInode string, byForce bool) error {
    if !CheckValidFilename(srcName) || !CheckValidFilename(desName) {
        return exception.EX_INVALID_FILENAME
    }
    if !byForce && outapi.ForceCheckExist(this.io.CheckExist(GenFileName(desInode, desName))) {
        return exception.EX_FOLDER_ALREADY_EXIST
    }

    var modifiedMeta=FileMeta(map[string]string {
        META_PARENT_INODE: desInode,
    })
    if err:=this.io.Copy(GenFileName(srcInode, srcName), GenFileName(desInode, desName), modifiedMeta); err!=nil {
        return exception.EX_FILE_NOT_EXIST
    }

    var globalError *errorgroup.ErrorAssembly=nil
    var geLock sync.Mutex
    var wg sync.WaitGroup

    var routineToUpdateDesParentMap=func() {
        defer wg.Done()

        {
            var desParentMap=dvc.GetFD(GenFileName(desInode, FOLDER_MAP), this.io)
            if desParentMap==nil {
                Secretary.Error("kernel.filesystem::MvXParalleled", "Fail to get foldermap fd for folder "+desInode)
                geLock.Lock()
                globalError=errorgroup.AddIn(globalError, exception.EX_IO_ERROR)
                geLock.Unlock()
                return
            }
            if err:=desParentMap.Submit(filetype.FastMake(desName)); err!=nil {
                Secretary.Error("kernel.filesystem::MvXParalleled", "Fail to submit foldermap patch for folder "+desInode)
                desParentMap.Release()
                geLock.Lock()
                globalError=errorgroup.AddIn(globalError, err)
                geLock.Unlock()
                return
            }
            desParentMap.Release()
        }
    }

    var routineToRemoveOldNnode=func() {
        defer wg.Done()

        this.io.Delete(GenFileName(srcInode, srcName))
    }

    var routineToUpdateSrcParentMap=func() {
        defer wg.Done()

        {
            var srcParentMap=dvc.GetFD(GenFileName(srcInode, FOLDER_MAP), this.io)
            if srcParentMap==nil {
                Secretary.Error("kernel.filesystem::MvXParalleled", "Fail to get foldermap fd for folder "+srcInode)
                geLock.Lock()
                globalError=errorgroup.AddIn(globalError, exception.EX_IO_ERROR)
                geLock.Unlock()
                return
            }
            if err:=srcParentMap.Submit(filetype.FastAntiMake(srcName)); err!=nil {
                Secretary.Error("kernel.filesystem::MvXParalleled", "Fail to submit foldermap patch for folder "+srcInode)
                srcParentMap.Release()
                geLock.Lock()
                globalError=errorgroup.AddIn(globalError, err)
                geLock.Unlock()
                return
            }
            srcParentMap.Release()
        }
    }

    var routineToUpdateDotDot=func() {
        defer wg.Done()

        var _, dstFileNnodeOriginal, _=this.io.Get(GenFileName(desInode, desName))
        var dstFileNnode, _=dstFileNnodeOriginal.(*filetype.Nnode)
        if dstFileNnode==nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to read nnode "+GenFileName(desInode, desName)+".")
            geLock.Lock()
            globalError=errorgroup.AddIn(globalError, exception.EX_IO_ERROR)
            geLock.Unlock()
            return
        }
        var target=GenFileName(dstFileNnode.DesName, "..")
        if err:=this.io.Put(target, filetype.NewNnode(desInode), nil); err!=nil {
            Secretary.Error("kernel.filesystem::MvX", "Fail to modify .. link for "+dstFileNnode.DesName+".")
            geLock.Lock()
            globalError=errorgroup.AddIn(globalError, err)
            geLock.Unlock()
            return
        } else {
            // Secretary.Log("kernel.filesystem::MvX", "Update file "+target)
        }
    }

    wg.Add(4)
    go routineToUpdateDesParentMap()
    go routineToRemoveOldNnode()
    go routineToUpdateSrcParentMap()
    go routineToUpdateDotDot()
    wg.Wait()

    return globalError
}

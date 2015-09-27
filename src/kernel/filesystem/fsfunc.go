package filesystem
// Most impotant implementation of pseudo-filesystem layer.

import (
    "outapi"
    "errors"
    "definition/exception"
    dvc "kernel/distributedvc"
    "kernel/filetype"
    "utils/uniqueid"
    . "utils/timestamp"
    "logger"
)

const ROOT_INODE_NAME="rootNode"

type Fs struct {
    io outapi.Outapi
}

func NewFs(_io outapi.Outapi) *Fs {
    return &Fs{
        io: _io
    }
}

//==============================================================================
// Followings are filesystem functions:

// path is a unix-like path string. If path starts with "/", search begins at
// root node. Otherwise in the frominode folder, when the frominode must exist.
// For any errors, a blank string and error will be returned.
func (this *Fs)Locate(path string, frominode string/*=-""*/) (string, error) {
    if strings.HasPrefix(path, "/") {
        frominode=ROOT_INODE_NAME
    }
    rawResult:=strings.Split(path, "/")
    for _, e:=range rawResult {
        if e!="" {
            frominode, err:=lookUp(frominode, e, this.io)
            if frominode=="" {
                return "", errors.New(exception.EX_FAIL_TO_LOOKUP)
            }
        }
    }

    return frominode, nil
}

func (this *Fs)Mkdir(foldername string, frominode string) error {
    if !CheckValidFilename(foldername) {
        return errors.New(exception.EX_INVALID_FILENAME)
    }

    var par=dvc.GetFD(frominode, this.io)
    var flist=par.GetFile().(*filetype.Kvm)
    if flist==nil {
        return errors.New(exception.EX_INODE_NONEXIST)
    }
    flist.CheckOut()
    if _, ok:=flist.Kvm[foldername]; ok {
        return errors.New(exception.EX_FOLDER_ALREADY_EXIST)
    }

    var newFileName=uniqueid.GenGlobalUniqueName()
    var newFile=dvc.FD(newFileName, this.io)

    fmap:=filetype.NewKvMap()
    fmap.CheckOut()
    fmap.Kvm[".."]=&filetype.KvmapEntry{
        Timestamp: GetTimestamp(0),
        Key: "..",
        Val: frominode,
    }
    fmap.Kvm["."]=&filetype.KvmapEntry{
        Timestamp: GetTimestamp(0),
        Key: ".",
        Val: newFileName,
    }
    fmap.CheckIn()

    if err:=newFile.PutOriginalFile(fmap, nil); err!=nil {
        return err
    }

    patcher:=filetype.NewKvMap()
    patcher.SetTS(GetTimestamp(flist.GetTS()))
    patcher.CheckOut()
    patcher.Kvm[foldername]=&filetype.KvmapEntry{
        Timestamp: GetTimestamp(0),
        Key: foldername,
        Val: newFileName,
    }
    patcher.CheckIn()
    if err:=par.CommitPatch(patcher); err!=nil {
        return err
    }

    return nil
}

// Format the filesystem.
// TODO: Setup clear old fs info
func (this *Fs)FormatFS() error {
    var nf=dvc.GetFD(ROOT_INODE_NAME, this.io)
    //nf.clear()

    fmap:=filetype.NewKvMap()
    fmap.CheckOut()
    fmap.Kvm[".."]=&filetype.KvmapEntry{
        Timestamp: GetTimestamp(0),
        Key: "..",
        Val: ROOT_INODE_NAME,
    }
    fmap.Kvm["."]=&filetype.KvmapEntry{
        Timestamp: GetTimestamp(0),
        Key: ".",
        Val: ROOT_INODE_NAME,
    }
    fmap.CheckIn()

    if err:=nf.PutOriginalFile(fmap, nil); err!=nil {
        return err
    }

    logger.Secretary.Log("kernel.filesystem.Fs::formatfs", "Formatted!")
    reutrn nil
}

// Only returns file name list of one inode. Innername excluded.
func (this *Fs)List(frominode string) ([]string, error) {
    var inodefile=dvc.GetFD(frominode, this.io).GetFile().(*filetype.Kvmap)
    if inodefile==nil {
        return nil, errors.New(exception.EX_INODE_NONEXIST)
    }
    inodefile.CheckOut()

    var ret=[]string{}
    for k, _:=range inodefile.Kvm {
        if CheckValidFilename(k) {
            ret=append(ret, k)
        }
    }

    return ret, nil
}

// All the folder will be removed. No matter if it is empty or not.
func (this *Fs)Rm(foldername string, frominode string) error {
    if !CheckValidFilename(foldername) {
        return errors.New(exception.EX_INVALID_FILENAME)
    }

    var par=dvc.GetFD(frominode, this.io)
    var flist=par.GetFile().(*filetype.Kvmap)
    if flist==nil {
        return errors.New(exception.EX_INODE_NONEXIST)
    }
    if _, ok:=flist.Kvm[foldername]; !ok {
        return
    }

    var patcher=filetype.NewKvMap()
    patcher.SetTS(GetTimestamp(flist.GetTS()))
    patcher.CheckOut()
    patcher.Kvm[foldername]=&filetype.KvmapEntry{
        Timestamp: GetTimestamp(0),
        Key: foldername,
        Val: filetype.REMOVE_SPECIFIED,
    }
    patcher.CheckIn()
    if err:=par.CommitPatch(patcher); err!=nil {
        return err
    }

}

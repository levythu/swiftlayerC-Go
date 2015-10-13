package filesystem
// Most impotant implementation of pseudo-filesystem layer.

import (
    "outapi"
    "errors"
    "definition/exception"
    dvc "kernel/distributedvc"
    "kernel/filetype"
    "utils/uniqueid"
    "strings"
    . "utils/timestamp"
    . "kernel/distributedvc/filemeta"
    . "kernel/distributedvc/constdef"
    "logger"
    "io"
    "fmt"
)

const ROOT_INODE_NAME="rootNode"

type Fs struct {
    io outapi.Outapi
}
func __nouse__() {
    fmt.Println("123")
}

func NewFs(_io outapi.Outapi) *Fs {
    return &Fs{
        io: _io,
    }
}

//==============================================================================
// Followings are filesystem functions:

// path is a unix-like path string. If path starts with "/", search begins at
// root node. Otherwise in the frominode folder, when the frominode must exist.
// For any errors, a blank string and error will be returned.
func (this *Fs)Locate(path string, frominode string/*=""*/) (string, error) {
    if strings.HasPrefix(path, "/") || frominode=="" {
        frominode=ROOT_INODE_NAME
    }
    rawResult:=strings.Split(path, "/")
    for _, e:=range rawResult {
        if e!="" {
            frominode, _=lookUp(frominode, e, this.io)
            if frominode=="" {
                // It is correct only to check result without referring to error.
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
    var flist, _=par.GetFile().(*filetype.Kvmap)
    if flist==nil {
        return errors.New(exception.EX_INODE_NONEXIST)
    }
    flist.CheckOut()
    if _, ok:=flist.Kvm[foldername]; ok {
        return errors.New(exception.EX_FOLDER_ALREADY_EXIST)
    }

    var newFileName=uniqueid.GenGlobalUniqueName()
    var newFile=dvc.GetFD(newFileName, this.io)

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
        Timestamp: GetTimestamp(flist.GetRelativeTS(foldername)),
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
    return nil
}

// Only returns file name list of one inode. Innername excluded.
func (this *Fs)List(frominode string) ([]string, error) {
    var inodefile, _=dvc.GetFD(frominode, this.io).GetFile().(*filetype.Kvmap)
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
    var flist, _=par.GetFile().(*filetype.Kvmap)
    if flist==nil {
        return errors.New(exception.EX_INODE_NONEXIST)
    }
    if _, ok:=flist.Kvm[foldername]; !ok {
        return nil
    }

    var patcher=filetype.NewKvMap()
    patcher.SetTS(GetTimestamp(flist.GetTS()))
    patcher.CheckOut()
    patcher.Kvm[foldername]=&filetype.KvmapEntry{
        Timestamp: GetTimestamp(flist.GetRelativeTS(foldername)),
        Key: foldername,
        Val: filetype.REMOVE_SPECIFIED,
    }
    patcher.CheckIn()
    if err:=par.CommitPatch(patcher); err!=nil {
        return err
    }

    return nil
}

// To put a large file and modify its corresponding index.
// Note that the function is synchronous, which means that it
// will block until data are fully written.
// It will try to put a file at destination, no matter whether
// there's already one file, which will be replaced then.

// The para frominode could be used to accerlerate index access.
// Note that the type/time will be specified in the func, so no need to provide in meta.
func (this *Fs)Put(destination string, frominode string/*=""*/, meta FileMeta/*=nil*/, dataSource io.Reader, typeoffile string/*=""*/) error {
    var lastPos=strings.LastIndex(destination, "/")

    var path=destination[:lastPos+1]
    var filename=destination[lastPos+1:]
    var basenode, err=this.Locate(path, frominode)
    if err!=nil {
        return err
    }

    if typeoffile=="" {
        typeoffile=(&filetype.Blob{}).GetType()
    }

    var newFileName=uniqueid.GenGlobalUniqueNameWithTag("Stream")
    var newFilefd=dvc.GetFD(newFileName, this.io)
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
    var flist, _=par.GetFile().(*filetype.Kvmap)
    if flist==nil {
        return errors.New(exception.EX_INODE_NONEXIST)
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
type Phase1Callback func(error, FileMeta) io.WriteCloser
// Phase2Callback is called when transmission completed.
type Phase2Callback func(error)

func (this *Fs)Get(source string, frominode string/*=""*/, phase1 Phase1Callback, phase2 Phase2Callback) {
    // Use this.locate, but not for finding folder. Note the semantic discrepancy.
    var obj, err=this.Locate(source, frominode)
    if err!=nil {
        phase1(err, nil)
        return
    }

    var objFD=dvc.GetFD(obj, this.io)
    rc, fm, err:=objFD.GetFileStream()
    if err!=nil{
        phase1(err, nil)
        return
    }
    if rc==nil {
        // 404
        phase1(errors.New(exception.EX_FILE_NOT_EXIST), nil)
        return
    }

    var wc=phase1(nil, fm)
    if wc==nil {
        // Phase1 requires to terminate.
        rc.Close()
        return
    }

    go func() {
        _, copyError:=io.Copy(wc, rc)
        rc.Close()
        wc.Close()
        phase2(copyError)
    }()
}

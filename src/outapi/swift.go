package outapi

import (
    "github.com/ncw/swift"
    "definition/configinfo"
    "fmt"
    . "kernel/distributedvc/filemeta"
    "definition/exception"
    "kernel/filetype"
    "bytes"
    "io"
    . "kernel/distributedvc/constdef"
)

type SwiftConnector struct {
    c *swift.Connection
}

func _no__use_1_() {
    fmt.Println("nosue")

}
// If auth failed, return nil
func ConnectbyAuth(username string, passwd string, tenant string) *SwiftConnector {
    swc:=&swift.Connection{
        UserName: username,
        ApiKey: passwd,
        Tenant: tenant,
        AuthUrl: configinfo.SWIFT_AUTH_URL,
        //AuthVersion: 2,
    }
    if err:=swc.Authenticate();err!=nil {
        panic(exception.EX_KEYSTONE_AUTH_ERROR)
        return nil
    }

    return &SwiftConnector{
        c: swc,
    }
}
func ConnectbyPreauth(account string, token string) *SwiftConnector {
    // TODO: set it up!
    panic(exception.EX_KEYSTONE_AUTH_ERROR)
    return nil
}

type Swiftio struct {
    //Implementing outapi.Outapi
    conn *SwiftConnector
    container string
}

// 2 ways to setup a new swift io.
func NewSwiftio(_conn *SwiftConnector, _container string) *Swiftio {
    return &Swiftio{
        conn: _conn,
        container: _container,
    }
}
func DupSwiftio(oldio *Swiftio, _container string) *Swiftio {
    return &Swiftio{
        conn: oldio.conn,
        container: _container,
    }
}

func (this *Swiftio)GenerateUniqueID() string {
    return "outapi.Swiftio: "+this.container
}

func (this *Swiftio)Getinfo(filename string) (FileMeta, error) {
    _, headers, err:=this.conn.c.Object(this.container, filename)
    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil
        }
        return nil, err
    }
    //fmt.Println(headers)
    return FileMeta(headers.ObjectMetadata()), nil
}

func (this *Swiftio)Putinfo(filename string, info FileMeta) error {
    head4Put:=swift.Metadata(info).ObjectHeaders()
    return this.conn.c.ObjectUpdate(this.container, filename, head4Put)
}

func (this *Swiftio)Delete(filename string) error {
    err:=this.conn.c.ObjectDelete(this.container, filename)
    if err!=nil && err!=swift.ObjectNotFound {
        return err
    }
    return nil
}

// Get file and automatically check the MD5
func (this *Swiftio)Get(filename string) (FileMeta, filetype.Filetype, error) {
    contents:=&bytes.Buffer{}
    header, err:=this.conn.c.ObjectGet(
        this.container, filename, contents,
        configinfo.INDEX_FILE_CHECK_MD5,
        nil)

    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil, nil
        }
        return nil, nil, err
    }
    meta:=header.ObjectMetadata()

    resFile:=filetype.Makefile(meta[METAKEY_TYPE])
    if resFile==nil {
        return nil, nil, exception.EX_UNSUPPORTED_TYPESTAMP
    }
    resFile.LoadIn(contents)

    return FileMeta(meta), resFile, nil
}

func (this *Swiftio)Put(filename string, content filetype.Filetype, info FileMeta) error {
    if info==nil {
        info=FileMeta(map[string]string{})
    }
    meta:=swift.Metadata(info)
    meta[METAKEY_TYPE]=content.GetType()

    buffer:=&bytes.Buffer{}
    content.WriteBack(buffer)

    _, err:=this.conn.c.ObjectPut(this.container, filename, buffer, false, "", "", meta.ObjectHeaders())
    return err
}

// If pointerfile, returns the actual data of the pointed one.
func (this *Swiftio)GetStream(filename string) (FileMeta, io.ReadCloser, error) {
    file, header, err:=this.conn.c.ObjectOpen(this.container, filename, false, nil)
    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil, nil
        }
        return nil, nil, err
    }
    meta:=header.ObjectMetadata()

    return FileMeta(meta), file, nil
}

func (this *Swiftio)PutStream(filename string, info FileMeta) (io.WriteCloser, error) {
    if info==nil || !CheckIntegrity(info) {
        return nil, exception.EX_METADATA_NEEDS_TO_BE_SPECIFIED
    }
    meta:=swift.Metadata(info)

    fileW, err:=this.conn.c.ObjectCreate(this.container, filename, false, "", "", meta.ObjectHeaders())
    if err!=nil {
        return nil, err
    }
    return fileW, nil
}

func (this *Swiftio)EnsureSpace() (bool, error) {
    _, _, err:=this.conn.c.Container(this.container)
    if err==swift.ContainerNotFound {
        err=this.conn.c.ContainerCreate(this.container, nil)
        return true, err
    }
    return false, err
}

func (this *Swiftio)Copy(srcname string, desname string, overrideMeta FileMeta) error {
    var _, err=this.conn.c.ObjectCopy(this.container, srcname, this.container, desname, overrideMeta)
    return err
}

func (this *Swiftio)CheckExist(filename string) (bool, error) {
    _, _, err:=this.conn.c.Object(this.container, filename)
    if err!=nil {
        if err==swift.ObjectNotFound {
            return false, nil
        }
        return false, err
    }
    return true, nil
}

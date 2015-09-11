package outapi

import (
    "github.com/ncw/swift"
    "definition/configinfo"
    "fmt"
    . "kernel/distributedvc/filemeta"
    "definition/exception"
    "errors"
    "kernel/filetype"
    . "utils/timestamp"
    "bytes"
)

type SwiftConnector struct {
    c *swift.Connection
}
// If auth failed, return nil
func ConnectbyAuth(username string, passwd string, tenant string) *SwiftConnector {
    swc:=&swift.Connection{
        UserName: username,
        ApiKey: passwd,
        Tenant: tenant,
        AuthUrl: configinfo.GetProperty_Node("swift_auth_url").(string),
        //AuthVersion: 2,
    }
    if err:=swc.Authenticate();err!=nil {
        fmt.Println(err.Error())
        panic(errors.New(exception.EX_KEYSTONE_AUTH_ERROR))
        return nil
    }

    return &SwiftConnector{
        c: swc,
    }
}
func ConnectbyPreauth(account string, token string) *SwiftConnector {
    // TODO: set it up!
    panic(errors.New(exception.EX_KEYSTONE_AUTH_ERROR))
    return nil
}

type Swiftio struct {
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
    fmt.Println(headers)
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
    header, err:=this.conn.c. ObjectGet(
        this.container, filename, contents,
        configinfo.GetProperty_Node("index_file_check_md5").(bool),
        nil)

    if err!=nil {
        if err==swift.ObjectNotFound {
            return nil, nil, nil
        }
        return nil, nil, err
    }
    meta:=header.ObjectMetadata()

    //TODO: change the typestamp to the constant in kernel.distributedvc.filehandler
    resFile:=filetype.Makefile(meta["typestamp"])
    if resFile==nil {
        return nil, nil, errors.New(exception.EX_UNSUPPORTED_TYPESTAMP)
    }
    //TODO: change the timestamp to the constant in kernel.distributedvc.filehandler
    resFile.Init(contents, String2ClxTimestamp(meta["timestamp"]))

    return FileMeta(meta), resFile, nil
}
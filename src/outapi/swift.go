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
        AuthUrl: configinfo.GetProperty_Node("swift_auth_url").(string),
        //AuthVersion: 2,
    }
    if err:=swc.Authenticate();err!=nil {
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
    var filem, metaerror=this.Getinfo(filename)
    if metaerror!=nil {
        if metaerror==swift.ObjectNotFound {
            return nil, nil, nil
        }
        return nil, nil, metaerror
    }

    if filetype.CheckPointerMap[filem[METAKEY_TYPE]] {
        // is pointer file
        var metaFile=filetype.Makefile(filem[METAKEY_TYPE])
        if metaFile==nil {
            return nil, nil, errors.New(exception.EX_UNSUPPORTED_TYPESTAMP)
        }

        var targetFile=metaFile.(filetype.PointerFileType)
        targetFile.Init(nil, String2ClxTimestamp(filem[METAKEY_TIMESTAMP]))
        targetFile.SetPointer(filem[filetype.META_POINT_TO])
        return filem, targetFile, nil
    }

    contents:=&bytes.Buffer{}
    header, err:=this.conn.c.ObjectGet(
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

    resFile:=filetype.Makefile(meta[METAKEY_TYPE])
    if resFile==nil {
        return nil, nil, errors.New(exception.EX_UNSUPPORTED_TYPESTAMP)
    }
    if resFile.IsPointer() {
        return nil, nil, errors.New(exception.EX_INCONSISTENT_TYPE)
    }
    resFile.Init(contents, String2ClxTimestamp(meta[METAKEY_TIMESTAMP]))

    return FileMeta(meta), resFile, nil
}

func (this *Swiftio)Put(filename string, content filetype.Filetype, info FileMeta) error {
    if info==nil {
        info=FileMeta(map[string]string{})
    }
    meta:=swift.Metadata(info)
    meta[METAKEY_TIMESTAMP]=content.GetTS().String()
    meta[METAKEY_TYPE]=content.GetType()

    if content.IsPointer() {
        var contentInPointer=content.(filetype.PointerFileType)
        meta[filetype.META_POINT_TO]=contentInPointer.GetPointer()
        this.Putinfo(filename, FileMeta(meta))
        return nil
    }

    buffer:=&bytes.Buffer{}
    content.WriteBack(buffer)

    _, err:=this.conn.c.ObjectPut(this.container, filename, buffer, false, "", "", meta.ObjectHeaders())
    return err
}

// If pointerfile, returns the actual data of the pointed one.
func (this *Swiftio)GetStream(filename string) (FileMeta, io.ReadCloser, error) {
    // Check whether it is pointer type first.
    var filem, metaerror=this.Getinfo(filename)
    if metaerror!=nil {
        if metaerror==swift.ObjectNotFound {
            return nil, nil, nil
        }
        return nil, nil, metaerror
    }
    if filetype.CheckPointerMap[filem[METAKEY_TYPE]] {
        if filem[filetype.META_POINT_TO]!=filename {
            // jump to the pointed file
            return this.GetStream(filem[filetype.META_POINT_TO])
        }
    }

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

// If PointerFile, automatically set to pointer to itself.
func (this *Swiftio)PutStream(filename string, info FileMeta) (io.WriteCloser, error) {
    if info==nil || !CheckIntegrity(info) {
        return nil, errors.New(exception.EX_METADATA_NEEDS_TO_BE_SPECIFIED)
    }
    if filetype.CheckPointerMap[info[METAKEY_TYPE]] {
        info[filetype.META_POINT_TO]=filename
    }
    meta:=swift.Metadata(info)

    fileW, err:=this.conn.c.ObjectCreate(this.container, filename, false, "", "", meta.ObjectHeaders())
    if err!=nil {
        return nil, err
    }
    return fileW, nil
}

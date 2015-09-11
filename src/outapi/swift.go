package outapi

import (
    "github.com/ncw/swift"
    "definition/configinfo"
    "fmt"
    . "kernel/distributedvc/filemeta"
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
        return nil
    }

    return &SwiftConnector{
        c: swc,
    }
}
func ConnectbyPreauth(account string, token string) *SwiftConnector {
    // TODO: set it up!
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
    fmt.Println(headers.ObjectMetadata())
    return nil, nil
}

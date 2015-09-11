package outapi

import (
    "github.com/ncw/swift"
    "definition/configinfo"
)

struct SwiftConnector {
    c *swift.Connection
}
// If auth failed, return nil
func ConnectbyAuth(username string, passwd string, tenant string) *SwiftConnector {
    swc:=&swift.Connection{
        UserName: username,
        ApiKey: passwd,
        Tenant: tenant,
        AuthUrl: configinfo.GetProperty_Node("swift_proxy_url").(string),
    }
    if swc.Authenticate()!=nil {
        return nil
    }

    return &SwiftConnector{
        c: swc
    }
}
func ConnectbyPreauth(account string, token string) *SwiftConnector {
    // TODO: set it up!
    return nil
}

struct Swiftio {
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

func (this *Swiftio)generateUniqueID string {
    return "outapi.Swiftio: "+this.container
}

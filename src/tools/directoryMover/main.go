package main

import (
    "github.com/ncw/swift"
    "outapi"
    "flag"
    "fmt"
    "time"
    "os"
)

var c=outapi.DefaultConnector.DumpConn()

func main() {
    var pContainer=flag.String("container", "", "The container to manipulate.")
    var pFromPath=flag.String("from", "", "The from path.")
    var pToPath=flag.String("to", "", "The to path")
    flag.Parse()

    if *pContainer=="" {
        fmt.Println("Container must be specified.")
        os.Exit(1)
    }
    if *pFromPath==*pToPath {
        fmt.Println("FromPath==ToPath, abort.")
        return
    }
    var nowTime=time.Now().UnixNano()

    var objList, err=c.ObjectsAll(*pContainer, &swift.ObjectsOpts {
        Prefix: *pFromPath,
    })
    if err!=nil {
        fmt.Println(err)
        os.Exit(1)
    }

    var io=outapi.NewSwiftio(outapi.DefaultConnector, *pContainer)
    for _, e:=range objList {
        var fromName=e.Name
        var toName=*pToPath+e.Name[len(*pFromPath):]
        fmt.Println("Moving", fromName, "->", toName)

        if err:=io.Copy(fromName, toName, nil); err!=nil {
            fmt.Println("Error:", err, "when trying to copy", fromName)
            os.Exit(1)
        }
        if err:=io.Delete(fromName); err!=nil {
            fmt.Println("Error:", err, "when trying to delete", fromName)
            os.Exit(1)
        }
    }
    fmt.Println("Time consumed:", time.Now().UnixNano()-nowTime, "ns")
    return
}

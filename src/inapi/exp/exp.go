package exp


import (
    . "github.com/levythu/gurgling"
    "kernel/filesystem"
    "fmt"
    "outapi"
    "strconv"
    //"logger"
)


func __expgo_nouse() {
    fmt.Println("nouse")
}

func ExpRouter() Router {
    var rootRouter=ARegexpRouter()

    rootRouter.Put(`/batchput`, batchPutHandler)

    return rootRouter
}

func batchPutHandler(req Request, res Response) {
    var container=req.Get("P-Container")
    var frominode=req.Get("P-From-Inode")
    var fromn=req.Get("P-From")
    var ton=req.Get("P-To")
    var prefix=req.Get("P-Prefix")

    var content="The quick brown fox jumps over the lazy dog"

    var fs=filesystem.GetFs(outapi.NewSwiftio(outapi.DefaultConnector, container))
    i, _:=strconv.Atoi(fromn)
    j, _:=strconv.Atoi(ton)
    if err:=fs.BatchPutDir(prefix, frominode, i, j, content); err!=nil {
        res.Send(err.Error())
    } else {
        res.Send("OK")
    }
}

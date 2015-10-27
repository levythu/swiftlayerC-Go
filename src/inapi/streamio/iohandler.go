package streamio

// the router implementing restAPIs that support streaming upload/download

import (
    . "github.com/levythu/gurgling"
    //"outapi"
    //"kernel/filesystem"
    //"kernel/filetype"
    //"definition/exception"
    "fmt"
    //"logger"
)

func __iohandlergo_nouse() {
    fmt.Println("nouse")
}

func IORouter() Router {
    var rootRouter=ARegexpRouter()
    rootRouter.Use(`/([^/]+)/\[\[SC\](.+)\]/(.*)`, handlingShortcut)

    rootRouter.Get(`/([^/]+)/(.*)`, downloader)
    rootRouter.Put(`/([^/]+)/(.*)`, uploader)

    return rootRouter
}

const LAST_PARENT_NODE="Manipulated-Node"

// Handling shortcut retrieve. It's applied to all the api in the field
// format: /fs/{contianer}/[[SC]{rootnode}]/{followingpath}
func handlingShortcut(req Request, res Response) bool {
    // After the midware,
    // req.F()["HandledRR"][1]=={container},
    // req.F()["HandledRR"][2]=={followingpath},
    // req.F()["HandledRR"][3]=={rootnode},
    // note that if no shortcut is specified, there should not be [3]
    var matchRes=req.F()["RR"].([]string)
    var t=matchRes[2]
    matchRes[2]=matchRes[3]
    matchRes[3]=t
    req.F()["HandledRR"]=matchRes

    return true
}

func downloader(req Request, res Response) {

}

func uploader(req Request, res Response) {
    
}

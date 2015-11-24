package streamio

// the router implementing restAPIs that support streaming upload/download

import (
    . "github.com/levythu/gurgling"
    . "kernel/distributedvc/filemeta"
    //"outapi"
    //"kernel/filesystem"
    //"kernel/filetype"
    "definition/exception"
    "io"
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

// ==========================API DOCS=======================================
// API Name: Stream data from specified path
// Action: Read the destination data and return it as a file by streaming
// API URL: /io/{contianer}/{followingpath}
// REQUEST: GET
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be listed
// Returns:
//      - HTTP 200: No error and the result will be returned in raw-data streaming.
//      - HTTP 404: Either the container or the filepath does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func downloader(req Request, res Response) {
    var pathDetail, _=req.F()["HandledRR"].([]string)
    if pathDetail==nil {
        pathDetail=req.F()["RR"].([]string)
        pathDetail=append(pathDetail, filesystem.ROOT_INODE_NAME)
    }

    var fs=filesystem.NewFs(outapi.NewSwiftio(outapi.DefaultConnector, pathDetail[1]))
    fs.Get(pathDetail[2], pathDetail[3], func(err error, fm FileMeta) io.Write {
        if err!=nil {
            if err==exception.EX_FILE_NOT_EXIST {
                res.Status("Nonexist container or path.", 404)
                return nil
            }
            res.Status("Internal Error: "+err.Error(), 500)
            return nil
        }
        // TODO: consider setting Content-Type
        res.SendCode(200)
        return res.R()
    }, func(err error) {
        if err!=nil {
            // TODO: logging the file sending error
        }
    }, true)
}

// ==========================API DOCS=======================================
// API Name: Stream data to specified file
// Action: Write the data from http to the file streamingly
// API URL: /io/{contianer}/{followingpath}
// REQUEST: PUT
// Parameters:
//      - contianer(in URL): the container name
// Returns:
//      - HTTP 200: No error, the file is written by force
//              When success, 'Manipulated-Node' will indicate the parent directory.
//      - HTTP 404: Either the container or the filepath does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func uploader(req Request, res Response) {
    var pathDetail, _=req.F()["HandledRR"].([]string)
    if pathDetail==nil {
        pathDetail=req.F()["RR"].([]string)
        pathDetail=append(pathDetail, filesystem.ROOT_INODE_NAME)
    }

    var fs=filesystem.NewFs(outapi.NewSwiftio(outapi.DefaultConnector, pathDetail[1]))
    var err=fs.Put(pathDetail[2], pathDetail[3], nil, "")
    if err!=nil {
        if err==exception.EX_FILE_NOT_EXIST {
            res.Status("Nonexist container or path.", 404)
            return
        }
        res.Status("Internal Error: "+err.Error(), 500)
        return
    }
    res.SendCode(200)
}

package fsmanage

// APIs for managing pseudo-filesystem, not for managing files

import (
    . "github.com/levythu/gurgling"
    "outapi"
    "kernel/filesystem"
    "kernel/filetype"
    //"strings"
    //"fmt"
    //"logger"
)

func FMRouter() Router {
    var rootRouter=ARegexpRouter()
    rootRouter.Use(`/([^/]+)/\[\[SC\](.+)\]/(.*)`, handlingShortcut)

    rootRouter.Get(`/([^/]+)/(.*)`, lsDirectory)
    rootRouter.Put(`/([^/]+)/(.*)`, mkDirectory)

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

// ==========================API DOCS=======================================
// API Name: List all the object in the directory
// Action: Return all the file in the format of JSON
// API URL: /fs/{contianer}/{followingpath}
// REQUEST: GET
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be listed
// Returns:
//      - HTTP 200: No error and the result will be returned in JSON in the body.
//              When success, 'Manipulated-Node' will indicate the listed directory.
//      - HTTP 404: Either the container or the filepath does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func lsDirectory(req Request, res Response) {
    var pathDetail, _=req.F()["HandledRR"].([]string)
    if pathDetail==nil {
        pathDetail=req.F()["RR"].([]string)
        pathDetail=append(pathDetail, filesystem.ROOT_INODE_NAME)
    }

    var fs=filesystem.NewFs(outapi.NewSwiftio(outapi.DefaultConnector, pathDetail[1]))
    var nodeName, err=fs.Locate(pathDetail[2], pathDetail[3])
    if err!=nil {
        res.Status("Nonexist container or path. "+err.Error(), 404)
        return
    }
    var resultList []*filetype.KvmapEntry
    resultList, err=fs.ListDetail(nodeName)
    if err!=nil {
        res.Status("Nonexist container or path. "+err.Error(), 404)
        return
    }

    res.Set(LAST_PARENT_NODE, nodeName)
    res.JSON(resultList)
}


// ==========================API DOCS=======================================
// API Name: Make one directory
// Action: make the directory only if it does not exist and its parent path exists
// API URL: /fs/{contianer}/{followingpath}
// REQUEST: PUT
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be create. Please guarantee its parent node exists.
// Returns:
//      - HTTP 201: No error and the directory creation application has been submitted.
//        to ensure created, another list operation should be carried.
//              When success, 'Manipulated-Node' will indicate the created directory.
//      - HTTP 202: No error but the directory has existed before.
//              When success, 'Manipulated-Node' will indicate the already exist directory.
//      - HTTP 404: Either the container or the parent filepath does not exist.
//      - HTTP 405: Parameters not specifed. Info will be provided in body.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func mkDirectory(req Request, res Response) {

}

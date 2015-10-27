package fsmanage

// APIs for managing pseudo-filesystem, not for managing files

import (
    . "github.com/levythu/gurgling"
    "outapi"
    "kernel/filesystem"
    "kernel/filetype"
    "definition/exception"
    "fmt"
    //"logger"
)


func __fsmanagego_nouse() {
    fmt.Println("nouse")
}

func FMRouter() Router {
    var rootRouter=ARegexpRouter()
    rootRouter.Use(`/([^/]+)/\[\[SC\](.+)\]/(.*)`, handlingShortcut)

    rootRouter.Get(`/([^/]+)/(.*)`, lsDirectory)
    rootRouter.Put(`/([^/]+)/(.*)`, mkDirectory)
    rootRouter.Delete(`/([^/]+)/(.*)`, rmDirectory)

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
//              When success, 'Manipulated-Node' will indicate the parent of created directory.
//      - HTTP 202: No error but the directory has existed before.
//              When success, 'Manipulated-Node' will indicate the parent of already exist directory.
//      - HTTP 404: Either the container or the parent filepath does not exist.
//      - HTTP 405: Parameters not specifed. Info will be provided in body.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func mkDirectory(req Request, res Response) {
    var pathDetail, _=req.F()["HandledRR"].([]string)
    if pathDetail==nil {
        pathDetail=req.F()["RR"].([]string)
        pathDetail=append(pathDetail, filesystem.ROOT_INODE_NAME)
    }

    var trimer=pathDetail[2]
    var i int
    for i=len(trimer)-1; i>=0; i-- {
        if trimer[i]!='/' {
            break
        }
    }
    if i<0 {
        res.Status("The directory to create should be specified.", 405)
        return
    }
    trimer=trimer[:i+1]
    var j int
    for j=i; j>=0; j-- {
        if trimer[j]=='/' {
            break
        }
    }
    var base=trimer[:j+1]
    trimer=trimer[j+1:]
    // now trimer holds the last foldername
    // base holds the parent folder path

    var fs=filesystem.NewFs(outapi.NewSwiftio(outapi.DefaultConnector, pathDetail[1]))
    var nodeName, err=fs.Locate(base, pathDetail[3])
    if err!=nil {
        res.Status("Nonexist container or path. "+err.Error(), 404)
        return
    }

    res.Set(LAST_PARENT_NODE, nodeName)
    err=fs.Mkdir(trimer, nodeName)
    if err!=nil {
        if err==exception.EX_INODE_NONEXIST {
            res.Status("Nonexist container or path.", 404)
            return
        }
        if err==exception.EX_FOLDER_ALREADY_EXIST {
            res.SendCode(202)
            return
        }
        res.Status("Internal Error: "+err.Error(), 500)
        return
    }

    res.SendCode(201)
}


// ==========================API DOCS=======================================
// API Name: Remove one directory
// Action: remove the directory only if it exists and its parent path exists
// API URL: /fs/{contianer}/{followingpath}
// REQUEST: DELETE
// Parameters:
//      - contianer(in URL): the container name
//      - followingpath(in URL): the path to be removed. Please guarantee its parent node exists.
// Returns:
//      - HTTP 204: The deletion succeeds but it is only a patch. to ensure created, another list
//        operation should be carried.
//              When success, 'Manipulated-Node' will indicate the parent of removed directory.
//      - HTTP 404: Either the container or the parent filepath does not exist.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
func rmDirectory(req Request, res Response) {
    var pathDetail, _=req.F()["HandledRR"].([]string)
    if pathDetail==nil {
        pathDetail=req.F()["RR"].([]string)
        pathDetail=append(pathDetail, filesystem.ROOT_INODE_NAME)
    }

    var trimer=pathDetail[2]
    var i int
    for i=len(trimer)-1; i>=0; i-- {
        if trimer[i]!='/' {
            break
        }
    }
    if i<0 {
        res.Status("The directory to create should be specified.", 405)
        return
    }
    trimer=trimer[:i+1]
    var j int
    for j=i; j>=0; j-- {
        if trimer[j]=='/' {
            break
        }
    }
    var base=trimer[:j+1]
    trimer=trimer[j+1:]
    // now trimer holds the last foldername
    // base holds the parent folder path

    var fs=filesystem.NewFs(outapi.NewSwiftio(outapi.DefaultConnector, pathDetail[1]))
    var nodeName, err=fs.Locate(base, pathDetail[3])
    if err!=nil {
        res.Status("Nonexist container or path. "+err.Error(), 404)
        return
    }

    res.Set(LAST_PARENT_NODE, nodeName)
    err=fs.Rm(trimer, nodeName)
    if err!=nil {
        if err==exception.EX_INODE_NONEXIST {
            res.Status("Nonexist container or path.", 404)
            return
        }
        res.Status("Internal Error: "+err.Error(), 500)
        return
    }

    res.SendCode(204)
}

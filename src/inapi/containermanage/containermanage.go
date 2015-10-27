package containermanage

// APIs for managing containers

import (
    . "github.com/levythu/gurgling"
    "outapi"
    "kernel/filesystem"
    "logger"
)

func CMRouter() Router {
    var rootRouter=ARouter()
    rootRouter.UseSpecified("/", "PUT", createContainerHandler, false)

    return rootRouter
}

// ==========================API DOCS=======================================
// API Name: Create & Initiate Container
// Action: Create a new container and format it for pseudo-fs
// API URL: /cn/{Container-Name}
// REQUEST: PUT
// Parameters:
//      - Container-Name(in URL): the container name to create
// Returns:
//      - HTTP 201: No problem and the container has been created.
//      - HTTP 202: Container already exist. No modification has made.
//      - HTTP 405: Parameters not specifed. Info will be provided in body.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
const HEADER_CONTAINER_NAME="Container-Name"
func createContainerHandler(req Request, res Response) {
    var containerName=req.Path()
    if containerName=="" || containerName[0]!='/' {
        res.Status("Path /container/{Container-Name} is required.", 405)
        return
    }
    containerName=containerName[1:]

    var ioAPI=outapi.NewSwiftio(outapi.DefaultConnector, containerName)
    var isNew, err=ioAPI.EnsureSpace()
    if err!=nil {
        logger.Secretary.Error("inapi.container.create", err)
        res.Status("Internal Error: "+err.Error(), 500)
        return
    }
    if !isNew {
        res.SendCode(202)
        return
    }
    // The container is newly created. Now format it.
    if err=filesystem.NewFs(ioAPI).FormatFS(); err!=nil {
        logger.Secretary.Error("inapi.container.create", err)
        res.Status("Internal Error: "+err.Error(), 500)
        return
    }

    res.SendCode(201)
}

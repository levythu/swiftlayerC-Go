package containermanage

// APIs for managing containers

import (
    "net/http"
    "outapi"
    "strings"
    "io"
    _ "fmt"
    "logger"
    "kernel/filesystem"
)

func RootRouter(w http.ResponseWriter, r *http.Request) {
    var methodInUpper=strings.ToUpper(r.Method)
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")

    if (methodInUpper=="PUT") {
        createContainerHandler(w, r)
        return
    }

    // Unrecognized.
    w.WriteHeader(400)
}

// ==========================API DOCS=======================================
// API Name: Create & Initiate Container
// Action: Create a new container and format it for pseudo-fs
// API URL: /containermng
// REQUEST: PUT
// Parameters:
//      - Container-Name(in Header): the container name to create
// Returns:
//      - HTTP 201: No problem and the container has been created.
//      - HTTP 202: Container already exist. No modification has made.
//      - HTTP 405: Parameters not specifed. Info will be provided in body.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS END===================================
const HEADER_CONTAINER_NAME="Container-Name"
func createContainerHandler(w http.ResponseWriter, r *http.Request) {
    var containerName string
    if containerHeader, ok:=r.Header[HEADER_CONTAINER_NAME]; !ok {
        w.WriteHeader(405)
        io.WriteString(w, "Header "+HEADER_CONTAINER_NAME+" is required.")
        return
    } else {
        if len(containerHeader)!=1 {
            w.WriteHeader(405)
            io.WriteString(w, "Format Error: Header "+HEADER_CONTAINER_NAME+".")
            return
        }
        containerName=containerHeader[0]
    }

    var ioAPI=outapi.NewSwiftio(outapi.DefaultConnector, containerName)
    var isNew, err=ioAPI.EnsureSpace()
    if err!=nil {
        logger.Secretary.Error("inapi.container.create", err)
        w.WriteHeader(500)
        io.WriteString(w, "Internal Error: "+err.Error())
        return
    }
    if !isNew {
        w.WriteHeader(202)
        return
    }
    // The container is newly created. Now format it.
    if err=filesystem.NewFs(ioAPI).FormatFS(); err!=nil {
        logger.Secretary.Error("inapi.container.create", err)
        w.WriteHeader(500)
        io.WriteString(w, "Internal Error: "+err.Error())
        return
    }

    w.WriteHeader(201)
}

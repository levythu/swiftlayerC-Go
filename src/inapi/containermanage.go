package inapi

// APIs for managing containers

import (
    "net/http"
    "outapi"
)

// ==========================API DOCS=======================================
// API Name: Create & Initiate Container
// Action: Create a new container and format it for pseudo-fs
// API URL: /containermng
// REQUEST: PUT
// Parameters:
//      - CONTAINER_NAME(in Header): the container name to create
// Returns:
//      - HTTP 201: No problem and the container has been created.
//      - HTTP 202: Container already exist. No modification has made.
//      - HTTP 500: Error. The body is supposed to return error info.
// ==========================API DOCS=======================================
func createContainerHandler(w http.ResponseWriter, r *http.Request) {
    
    res:=make([]byte, 3)
    ds:=iomidware.Blockify(r.Body)
    for {
        n, err:=ds.Read(res)
        fmt.Println(string(res[:n]),n)
        if err==io.EOF {
            break
        }
    }
}

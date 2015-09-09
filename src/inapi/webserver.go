package inapi

import (
    "net/http"
    //"io/ioutil"
    "fmt"
    "io"
    "utils/iomidware"
)

func uploadhandler(w http.ResponseWriter, r *http.Request) {
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

func Entry() {
    http.HandleFunc("/upload", uploadhandler)
    http.ListenAndServe(":9144", nil)
}

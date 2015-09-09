package inapi

import (
    "net/http"
    //"io/ioutil"
    "fmt"
    "io"
)

func uploadhandler(w http.ResponseWriter, r *http.Request) {
    res:=make([]byte, 10)
    for {
        n, err:=r.Body.Read(res)
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

package filesystem

import (
    "testing"
    "time"
    "fmt"
    "io"
    . "kernel/distributedvc/filemeta"
)

func _TestGet(t *testing.T) {
    var fs4test=NewFs(Testio)
    var r, w=io.Pipe()
    go func() {
        var buf=make([]byte, 5)
        for {
            n, err:=r.Read(buf)
            fmt.Print(string(buf[:n]))
            if n==0 || err!=nil {
                return
            }
        }
    }()

    fs4test.Get("/file1.txt", "", func(err error, _ FileMeta) io.WriteCloser {
        if err!=nil {
            fmt.Println("Error:", err)
            return nil
        }
        return w
    }, func(err error) {
        fmt.Println(err)
        fmt.Println("DONE.")
    })

    for {
        time.Sleep(time.Hour)
    }
}

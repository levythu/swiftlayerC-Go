package filesystem

import (
    "time"
    "fmt"
    "io"
)

func TestUpstream() {
    var fs4test=NewFs(Testio)
    r, w:=io.Pipe()
    go func() {
        fmt.Println(fs4test.Put("/file1.txt", "", r, ""))
    }()
    var str string
    fmt.Println("input:")
    for str!="END." {
        fmt.Scan(&str)
        w.Write([]byte(str))
    }
    w.Close()

    fmt.Println("SENT DONE.")
    for {
        time.Sleep(time.Hour)
    }
}

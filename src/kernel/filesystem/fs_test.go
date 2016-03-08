package filesystem

import (
    "testing"
    "time"
    "fmt"
)

var fs4test=NewFs(Testio)

func _TestFormat(t *testing.T) {
    fmt.Println(fs4test.FormatFS())

    for {
        time.Sleep(time.Hour)
    }
}

func _TestMkDir(t *testing.T) {
    fmt.Println(fs4test.Mkdir("directory1", fs4test.rootName, false))

    for {
        time.Sleep(time.Hour)
    }
}

func TestLS(t *testing.T) {
    fmt.Println(fs4test.List(fs4test.rootName))

    for {
        time.Sleep(time.Hour)
    }
}

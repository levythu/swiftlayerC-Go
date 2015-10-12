package filesystem

import (
    "testing"
    "time"
    "fmt"
)

func TestLS(t *testing.T) {
    var ses4test=NewSession(Testio)
    res, _:=ses4test.Ls()
    fmt.Println(res);
}

func _TestSession(t *testing.T) {
    var ses4test=NewSession(Testio)
    res, _:=ses4test.Ls()
    fmt.Println(res);

    fmt.Println(ses4test.Mkdir("花花花"))

    ses4test.Cd("..")
    res, _=ses4test.Ls()
    fmt.Println(res);

    ses4test.Cd("huahua")
    res, _=ses4test.Ls()
    fmt.Println(res);

    ses4test.Cd("实验2")
    res, _=ses4test.Ls()
    fmt.Println(res);
    fmt.Println(ses4test.Mkdir("asd asdjld asd"))

    ses4test.Cd("..")
    res, _=ses4test.Ls()
    fmt.Println(res);

    for {
        time.Sleep(time.Hour)
    }
}

func _TestErrSession(t *testing.T) {
    var ses4test=NewSession(Testio)
    ses4test.Cd("xle")
    for {
        time.Sleep(time.Hour)
    }
}

func _TestFormat(t *testing.T) {
    var fs4test=NewFs(Testio)
    fs4test.FormatFS()
    for {
        time.Sleep(time.Hour)
    }
}

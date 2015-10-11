package filesystem

import (
    "testing"
    "time"
    "fmt"
)

func TestSession(t *testing.T) {
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

package filesystem

import (
    "testing"
)

func TestSession(t *testing.T) {
    var ses4test=NewSession(Testio)
    res, _:=ses4test.Ls()
    t.Log(res);

    ses4test.Cd("..")
    res, _=ses4test.Ls()
    t.Log(res);

    ses4test.Cd("huahua")
    res, _=ses4test.Ls()
    t.Log(res);

    ses4test.Cd("..")
    res, _=ses4test.Ls()
    t.Log(res);
}

package distributedvc

// Unit test for kernel/distributedvc

import (
    "testing"
    //"kernel/filetype"
)

func _TestOverhaulOrder(t *testing.T) {
    t.Log(overhaulOrder)
}

func TestFDLatestPatch(t *testing.T) {
    fd:=GetFD("rootNode", Testio)
    t.Log(fd.GetLatestPatch())
}

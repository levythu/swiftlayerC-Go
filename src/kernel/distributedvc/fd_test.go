package distributedvc

// Unit test for kernel/distributedvc

import (
    "testing"
    "kernel/filetype"
    "time"
)

func _TestOverhaulOrder(t *testing.T) {
    t.Log(overhaulOrder)
}

func _TestFDLatestPatch(t *testing.T) {
    fd:=GetFD("only4test", Testio)
    t.Log(fd.GetLatestPatch())
}

func _TestFDGetFile(t *testing.T) {
    fd:=GetFD("only4test", Testio)
    fileGot:=fd.GetFile().(*filetype.Kvmap)
    fileGot.CheckOut()
    for i, e:=range fileGot.Kvm {
        t.Log(i,": ",e)
    }
}

func TestFDAddPatch(t *testing.T) {
    fd:=GetFD("only4test", Testio)
    newPatch:=filetype.NewKvMap()
    newPatch.CheckOut()
    newPatch.Kvm["Olllo"]=&filetype.KvmapEntry{
        Timestamp: 2333833,
        Key: "Olllo",
        Val: "Upload2",
    }
    newPatch.Kvm["ps"]=&filetype.KvmapEntry{
        Timestamp: 123,
        Key: "ps",
        Val: "爆米花",
    }
    newPatch.CheckIn()
    fd.CommitPatch(newPatch)
    t.Log("Checked in")
    for {
        time.Sleep(time.Hour)
    }
}

package filetype

import (
    "testing"
    "io/ioutil"
    . "definition"
    "bytes"
    "os"
)

func readFile(filename string) ([]byte, error) {
    var err error
    filename, err=GetABSPath(filename)
    if err!=nil {
        return nil, err
    }
    return ioutil.ReadFile(filename)
}

func _TestREAD(t *testing.T) {
    datas, _:=readFile("noupload/rootNode.cversion")
    datainput:=bytes.NewReader(datas)
    var kvm4test Kvmap
    kvm4test.Init(datainput,5)
    kvm4test.LoadIntoMem()

    for _, elem:=range kvm4test.readData {
        t.Log(elem)
    }

    t.Log("Finish")
}
func _TestWRITE(t *testing.T) {
    datas, _:=readFile("noupload/rootNode.cversion")
    datainput:=bytes.NewReader(datas)
    var kvm4test Kvmap
    kvm4test.Init(datainput,5)
    kvm4test.LoadIntoMem()

    var plc bytes.Buffer
    p2plc:=&plc
    kvm4test.WriteBack(p2plc)

    oName, _:=GetABSPath("noupload/outp.txt")
    ioutil.WriteFile(oName,p2plc.Bytes(),os.ModePerm)

    t.Log("Finish")
}
func _TestCHECKINnOUT(t *testing.T) {
    kvm4test:=NewKvMap()
    kvm4test.CheckOut()
    kvm4test.Kvm["asd阿斯顿"]=&KvmapEntry{123321123,"asd阿斯顿","valu1eHu阿萨德撒的"}
    kvm4test.CheckIn()

    var plc bytes.Buffer
    p2plc:=&plc
    kvm4test.WriteBack(p2plc)

    oName, _:=GetABSPath("noupload/outp2.txt")
    ioutil.WriteFile(oName,p2plc.Bytes(),os.ModePerm)

    t.Log(kvm4test.Kvm)
}

func TestMERGE(t *testing.T) {
    kvm4test:=NewKvMap()
    kvm4test.CheckOut()
    kvm4test.Kvm["asd阿斯顿"]=&KvmapEntry{123321123,"asd阿斯顿","valu1eHu阿萨德撒的"}
    kvm4test.Kvm["a2x"]=&KvmapEntry{13,"a2x","dsasadsad"}
    kvm4test.CheckIn()

    kvm4test2:=NewKvMap()
    kvm4test2.CheckOut()
    kvm4test2.Kvm["asd阿斯顿"]=&KvmapEntry{1232321123,"asd阿斯顿","vds"}
    kvm4test2.Kvm["qq"]=&KvmapEntry{1223,"qq","dsasadsad"}
    kvm4test2.CheckIn()

    kvm4test.MergeWith(kvm4test2)

    kvm4test.CheckOut()
    t.Log(kvm4test.Kvm)
    t.Log(kvm4test.Kvm["asd阿斯顿"])
}

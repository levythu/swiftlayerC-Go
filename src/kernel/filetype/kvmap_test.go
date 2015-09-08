package filetype

import (
    "testing"
    "io/ioutil"
    . "definition"
    "bytes"
)

func readFile(filename string) ([]byte, error) {
    var err error
    filename, err=GetABSPath(filename)
    if err!=nil {
        return nil, err
    }
    return ioutil.ReadFile(filename)
}

func TestJSONGet(t *testing.T) {
    datas, _:=readFile("rootNode.cversion")
    datainput:=bytes.NewReader(datas)
    var kvm4test Kvmap
    kvm4test.Init(datainput,5)
    kvm4test.LoadIntoMem()
    for _, elem:=range kvm4test.readData {
        t.Log(elem)
    }

    t.Log("123")
}

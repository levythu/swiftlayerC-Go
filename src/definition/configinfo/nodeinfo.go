package configinfo

import (
    "log"
    . "definition"
)

func errcomb(err1, err2 error) error {
    if err1!=nil {
        return err1
    }
    return err2
}

var conf map[string]Tout=make(map[string]Tout)
var err1=AppendFileToJSON("conf/nodeinfo.json", conf)
var err2=errcomb(err1, AppendFileToJSON("conf/accountinfo.json", conf))

func GetProperty_Node(proname string) Tout {
    const errPrefix="<Nodeinfo::GetProperty_Node> "
    if err2!=nil {
        log.Fatal(errPrefix+err2.Error())
    }
    var elem, ok=conf[proname]
    if ok==false {
        log.Fatal(errPrefix+"No such a property named "+proname)
    }

    return elem
}

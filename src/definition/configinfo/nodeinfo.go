package configinfo

import (
    "log"
    . "definition"
)

var conf,err=ReadFileToJSON("conf/nodeinfo.json")

func GetProperty_Node(proname string) Tout {
    const errPrefix="<Nodeinfo::GetProperty_Node> "
    if err!=nil {
        log.Fatal(errPrefix+err.Error())
    }
    var elem, ok=conf[proname]
    if ok==false {
        log.Fatal(errPrefix+"No such a property named "+proname)
    }

    return elem
}

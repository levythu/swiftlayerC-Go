package configinfo

import (
    "log"
    . "definition"
    "logger"
)

func errcomb(err1, err2 error) error {
    if err1!=nil {
        return err1
    }
    return err2
}

var conf map[string]Tout=make(map[string]Tout)

// WARN: Deprecated!
func GetProperty_Node(proname string) Tout {
    const errPrefix="<Nodeinfo::GetProperty_Node> "
    var elem, ok=conf[proname]
    if ok==false {
        log.Fatal(errPrefix+"No such a property named "+proname)
    }

    return elem
}

func errorAssert(err error, reason string) bool {
    if err!=nil {
        logger.Secretary.Error(reason, err)
        panic("EXIT DUE TO ASSERTION FAILURE.")
    }
    return false
}
func extractProperty(key string) Tout {
    var elem, ok=conf[key]
    if !ok {
        logger.Secretary.ErrorD("Fail to extract property <"+key+">")
        panic("EXIT DUE TO ASSERTION FAILURE.")
    }
    return elem
}

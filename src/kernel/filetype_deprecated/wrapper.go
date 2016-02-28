// Wrapper for all the filetypes, offering a quick way to setup one class.
// The id of different types is GetType(). Keep it unique.
package filetype

import (
    "reflect"
    //"fmt"
)
//============================================================
// Modify this to add new filetype.
var prototypeList=[]Filetype{&Kvmap{}, &Nonexist{}, &Blob{}}
//============================================================

var typeMap=makeTypeMap()
var CheckPointerMap=func() map[string]bool {
    var ret=make(map[string]bool)
    for _, elem:=range prototypeList {
        ret[elem.GetType()]=elem.IsPointer()
    }
    return ret
}()

func makeTypeMap() map[string]reflect.Type {
    ret:=make(map[string]reflect.Type)
    for _, elem:=range prototypeList {
        ret[elem.GetType()]=reflect.TypeOf(elem).Elem()
    }
    return ret
}

func Makefile(typeid string) Filetype {
    if res, ok:=typeMap[typeid]; ok {
        return reflect.New(res).Interface().(Filetype)
    }
    return nil
}

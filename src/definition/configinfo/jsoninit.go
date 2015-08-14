package configinfo

import (
    "io/ioutil"
    . "definition"
    "encoding/json"
)

//Pay attention that filename is a relative path
func ReadFileToJSON(filename string) (map[string]Tout, error) {
    var err error
    var res []byte
    filename, err=GetABSPath(filename)
    if err!=nil {
        return nil, err
    }
    res, err=ioutil.ReadFile(filename)
    if err!=nil {
        return nil, err
    }

    var ret map[string]Tout
    err=json.Unmarshal(res, &ret)
    if err!=nil {
        return nil, err
    }

    return ret, nil
}

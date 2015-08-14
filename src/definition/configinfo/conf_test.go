package configinfo

import (
    "testing"
)

func TestJSONGet(t *testing.T) {
    t.Log(ReadFileToJSON("conf/nodeinfo.json"))
    t.Log(GetProperty_Node("node_number"))
}

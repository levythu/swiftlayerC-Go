package configinfo

import (
    "testing"
)

func TestJSONGet(t *testing.T) {
    t.Log(GetProperty_Node("node_number"))
    t.Log(GetProperty_Node("keystone_username"))
}

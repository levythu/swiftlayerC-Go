package configinfo

import (
    "testing"
)

func TestJSONGet(t *testing.T) {
    InitAll()
    t.Log(NODE_NUMBER)
    t.Log(KEYSTONE_TENANT)
}

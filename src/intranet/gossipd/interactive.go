package gossipd

import (
    . "utils/timestamp"
    "strcov"
    . "definition"
    "fmt"
    "strings"
    "errors"
)

// implementation for the format to gossip

type GossipEntry struct {
    FDID string
    UpdateTime ClxTimestamp
    NodeNumber int
}

// Just segment them with \n
func (this *GossipEntry)Stringify() string {
    return this.FDID+"\n"+this.UpdateTime.String()+"\n"+strconv.Itoa(this.NodeNumber)
}

func BatchStringify(src []Tout) (string, error) {
    var ret=""
    for i, e:=range src {
        if p, ok:=e.(*GossipEntry); !ok {
            return "", errors.New("Format error")
        } else {
            ret+=p.Stringify()
            if i!=len(src)-1 {
                ret+="\n"
            }
        }
    }

    return ret, nil
}

// For errors returns nil
func ParseOne(src string) *GossipEntry {
    var res=strings.SplitN(src, "\n", 3)
    if len(res)!=3 {
        return nil
    }
    var pInt, err:=strconv.Atoi(res[2])
    if err!=nil {
        return nil
    }

    return &GossipEntry {
        FDID: res[0],
        UpdateTime: String2ClxTimestamp(res[1]),
        NodeNumber: pInt,
    }
}

// For errors returns nil
func ParseAll(src string) []*GossipEntry {
    var res=strings.SplitN(src, "\n")
    if len(res)%3!=0 {
        return nil
    }
    var result=[]*GossipEntry{}
    for i:=0; i<len(res); i+=3 {
        var pInt, err:=strconv.Atoi(res[i+2])
        if err!=nil {
            return nil
        }
        result=append(result, &GossipEntry {
            FDID: res[i],
            UpdateTime: String2ClxTimestamp(res[i+1]),
            NodeNumber: pInt,
        })
    }

    return result
}

package gossipd

import (
    conf "definition/configinfo"
    gsp "intranet/gossip"
    . "logger"
    . "definition"
    "errors"
    "stringss"
    "fmt"
)

func convStrToTout(src []string) []Tout {
    var ret=make([]Tout, len(src))
    for i, e:=range src {
        ret[i]=e
    }
    return ret
}

var (
    ADDR_FORMAT_ERROR=errors.New("Addr has wrong format.")
    CONT_FORMAT_ERROR=errors.New("Content has wrong format.")
    CONNECT_ERROR=errors.New("Connection error.")
    HTTP_ERROR=errors.New("HTTP Status code error: a non-2xx code.")
)

var gossipHTTPclient=&http.Client{}
func GossipViaHTTP(addr Tout, content []Tout) error {
    var addrStr, ok:=addr.(string)
    if !ok {
        return ADDR_FORMAT_ERROR
    }
    if content, err:=BatchStringify(content); err!=nil {
        return CONT_FORMAT_ERROR
    } else {
        // DO HTTP REQUEST
        var res, err=gossipHTTPclient.Post("http://"+addr+"/gossip", "test/plain; charset=utf-8", strings.NewReader(content))
        if err!=nil {
            return CONNECT_ERROR
        }
        // TODO: check result.
        res.Body.Close()
        if res.StatusCode%100!=2 {
            return HTTP_ERROR
        }

        return nil
    }
}

func Init() {
    var ret=gsp.NewBufferedGossiper(conf.GOSSIP_BUFFER_SIZE)
    ret.PeriodInMillisecond=conf.GOSSIP_PERIOD_IN_MS
    ret.EnsureTellCount=conf.GOSSIP_RETELL_TIMES
    ret.TellMaxCount=conf.GOSSIP_MAX_DELIVERED_IN_ONE_TICK
    ret.ParallelTell=conf.GOSSIP_MAX_TELLING_IN_ONE_TICK
    ret.SetGossiperList(convStrToTout(conf.FilterSelf(conf.SH2_MAP)))
    ret.SetGossipingFunc(GossipViaHTTP)

    gsp.GlobalGossiper=ret
}

func Entry(exit chan bool) {
    defer (func(){
        exit<-false
    })()

    Secretary.Log("gossipd::Entry", "Backgroud Gossiper is going to launch.")
    gsp.GlobalGossiper.Launch()
}
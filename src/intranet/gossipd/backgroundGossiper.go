package gossipd

import (
    conf "definition/configinfo"
    gsp "intranet/gossip"
    . "definition"
    "fmt"
)

func convStrToTout(src []string) []Tout {
    var ret=make([]Tout, len(src))
    for i, e:=range src {
        ret[i]=e
    }
    return ret
}
func Init() {
    var ret=gsp.NewBufferedGossiper(conf.GOSSIP_BUFFER_SIZE)
    ret.PeriodInMillisecond=conf.GOSSIP_PERIOD_IN_MS
    ret.EnsureTellCount=conf.GOSSIP_RETELL_TIMES
    ret.TellMaxCount=conf.GOSSIP_MAX_DELIVERED_IN_ONE_TICK
    ret.ParallelTell=conf.GOSSIP_MAX_TELLING_IN_ONE_TICK
    ret.SetGossiperList(convStrToTout(conf.FilterSelf(conf.SH2_MAP)))
    ret.SetGossipingFunc(func(addr Tout, content []Tout) error {
        fmt.Println(addr, ": ", content)
        return nil
    })

    gsp.GlobalGossiper=ret
}

func Entry(exit chan bool) {
    defer (func(){
        exit<-false
    })()

    gsp.GlobalGossiper.Launch()
}

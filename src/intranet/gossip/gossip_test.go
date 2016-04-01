package gossip

import (
    "testing"
    "time"
    "fmt"
)

func TestBatRand(t *testing.T) {
    go GlobalGossiper.Launch()
    var i=0
    for {
        time.Sleep(time.Second)
        if err:=GlobalGossiper.PostGossip(i); err==nil {
            fmt.Println("POSTED:", i)
        } else {
            fmt.Println("ERROR:", err)
        }
        i++
    }
}

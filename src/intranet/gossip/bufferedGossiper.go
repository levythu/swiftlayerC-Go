package gossip

import (
    . "definition"
    "sync"
    "errors"
)

type BufferedGossiper struct {
    *stdGossiperListImplementation

    // <0 for no launch
    PeriodInMillisecond int

    BufferSize int

    EnsureTellCount int

    TellMaxCount int

    ParallelTell int

    // ===============================

    do func(addr Tout, content []Tout) error

    // the next of the last
    tail int
    head int

    lenLock sync.RWMutex
    len int

    buffer []Tout
    gCount []int
}
func (this *BufferedGossiper)SetGossipingFunc(do func(addr Tout, content []Tout) error) {
    this.do=do
}

func NewBufferedGossiper(bufferSize int) *BufferedGossiper {
    return &*BufferedGossiper {
        BufferSize: bufferSize,
        head: 0,
        tail: 0,
        len: 0,
        buffer: make([]Tout, bufferSize),
        gCount: make([]int, bufferSize),
    }
}

var BUFFER_IS_FULL=errors.New("The buffer for buffered gossiper is full. New gossip cannot be checked in.")
func (this *BufferedGossiper)PostGossip(content Tout) error {
    this.lenLock.Lock()
    defer this.lenLock.Unlock()

    if this.len==this.BufferSize {
        return BUFFER_IS_FULL
    }
    this.len++

    buffer[this.tail]=content
    gCount[this.tail]=this.EnsureTellCount
    this.tail++
    if this.tail>=this.BufferSize {
        this.tail-=this.BufferSize
    }

    return nil
}

func (this *BufferedGossiper)gossip(content []Tout) {
    var c=
}
func (this *BufferedGossiper)onTick() {
    this.lenLock.Lock()
    defer this.lenLock.Unlock()

    var c=this.len
    if c>this.TellMaxCount {
        c=this.TellMaxCount
    }
    var res=make([]Tout, c)

    var p=this.head
    for i:=0; i<c; i++ {
        res[i]=this.buffer[p]
        this.gCount[p]-=this.ParallelTell
        p++
        if p>=this.BufferSize {
            p-=this.BufferSize
        }
    }
    for this.len>0 && this.gCount[this.head]<=0 {
        this.len--
        this.head++
        if this.head>=this.BufferSize {
            this.head-=this.BufferSize
        }
    }

    go this.gossip(res)
}

func (this *BufferedGossiper)Launch() error {
    if this.PeriodInMillisecond<0 {
        return nil
    }

}

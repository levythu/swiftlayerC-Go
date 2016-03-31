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

    buffer[tail]=content
    gCount[tail]=this.EnsureTellCount
    tail++
    if tail>=this.BufferSize {
        tail-=this.BufferSize
    }

    return nil
}

func (this *BufferedGossiper)Launch() error {
    if this.PeriodInMillisecond<0 {
        return nil
    }
    
}

package gossip

import (
    . "definition"
    "sync"
    "errors"
)

type Gossiper interface {
    PostGossip(content Tout) error

    // the list passed in will be replicated.
    SetGossiperList(list []Tout) error
    SetGossipingFunc(do func(addr Tout, content []Tout) error)

    // a deamon function
    Launch() error
}

type stdGossiperListImplementation struct {
    list []Tout
    lock sync.RWMutex
}

func (this *stdGossiperListImplementation)SetGossiperList(list []Tout) error {
    this.lock.Lock()
    defer this.lock.Unlock()

    this.list=[]Tout{}
    for _, e:=range list {
        this.list=append(this.list, e)
    }
    return nil
}
func (this *stdGossiperListImplementation)Get(numberInList int) (Tout, error) {
    this.lock.RLock()
    defer this.lock.RUnlock()

    if this.list==nil {
        return nil, errors.New("The gossiper list has not been set yet.")
    }
    if numberInList<0 || numberInList>=len(this.list) {
        return nil, errors.New("Invalid Access to gossiper list.")
    }

    return this.list[numberInList], nil
}
func (this *stdGossiperListImplementation)GetLen() (int, error) {
    this.lock.RLock()
    defer this.lock.RUnlock()

    if this.list==nil {
        return nil, errors.New("The gossiper list has not been set yet.")
    }

    return len(this.list), nil
}

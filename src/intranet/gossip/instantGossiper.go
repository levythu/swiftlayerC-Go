package gossip

import (
    . "definition"
)

// DEPRECATED

// No waiting list for this gossiper and no gossiping batch is implented. Nothing is
// controlled.

type InstantGossiper struct {
    *stdGossiperListImplementation
}

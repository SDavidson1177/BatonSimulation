package simulator

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
)

// GetShortestPath returns the shortest path from the source chain to the destination
// chain. This uses hop count to determine the length of a path.
// Hub: hub chains
// Direct: If true, connect only directly or through hub chains
func GetShortestPath(ctx context.Context, src string, dst string, hubs map[string]bool) ([]string, error) {
	len_hubs := len(hubs)
	is_hub := func(chain string) bool {
		if len_hubs == 0 {
			return true
		}

		for k := range hubs {
			if chain == k {
				return true
			}
		}

		return false
	}

	// Get the state
	state, err := GetStateFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Load the event queue
	src_found, dst_found := false, false
	const inf = 100000000
	event_queue := &EventHeap{}
	for chain := range state.Chains {
		var de *DijkstraEvent
		if chain == src {
			src_found = true
			de = NewDijkstraEvent(1, chain)
		} else {
			de = NewDijkstraEvent(inf, chain)
			if chain == dst {
				dst_found = true
			}
		}

		event_queue.Insert(de)
	}

	if !src_found || !dst_found {
		return nil, errors.New("could not find source and destination chain")
	}

	// Find the shortest path
	cmp := func(this_event Event, cmp_event Event) bool {
		dijk_event, ok := cmp_event.(*DijkstraEvent)
		if !ok {
			return false
		}

		base_event, ok := this_event.(*DijkstraEvent)
		if !ok {
			return false
		}

		return dijk_event.Chain == base_event.Chain
	}

	type Prev struct {
		Chain_id string
		Amount   int // number of chains that could be this chain's previous hop
	}
	prev_chain := make(map[string]Prev)

	sp := make([]string, 0)
	node := event_queue.Pop().(*DijkstraEvent)
	prev_chain[node.Chain] = Prev{}
	for node.Chain != dst {
		// Check unreachability
		if node.Distance == inf {
			return nil, errors.New("unreachable")
		}

		// Check for hubs
		if !(len_hubs == 0 || node.Distance == 1 || is_hub(node.Chain)) {
			node = event_queue.Pop().(*DijkstraEvent)
			continue
		}

		// Update all neighbours
		for n := range state.Chains[node.Chain].neighbours {
			c_event, c_index := event_queue.Find(&DijkstraEvent{Chain: n}, cmp)

			if c_event != nil {
				c_dijk_event := c_event.(*DijkstraEvent)
				if c_event.(*DijkstraEvent).Distance > node.Distance+1 {
					c_dijk_event.Distance = node.Distance + 1
					event_queue.Update(c_index)
					prev_chain[c_dijk_event.Chain] = Prev{Chain_id: node.Chain, Amount: 1}
				} else if c_event.(*DijkstraEvent).Distance == node.Distance+1 {
					// Since we assign every chain an infinite distance at first,
					// this can only happen if the distance is no longer infinite.
					// We can therefore assume that there is an entry in prev_chain

					// Replace with probability 1/amount
					p := prev_chain[c_dijk_event.Chain]
					p.Amount++
					if do_replace, err := rand.Int(rand.Reader, big.NewInt(int64(p.Amount))); err == nil && do_replace.Int64() == 0 {
						p.Chain_id = node.Chain
					}
					prev_chain[c_dijk_event.Chain] = p
				}
			}
		}

		// Get next
		node = event_queue.Pop().(*DijkstraEvent)
	}

	// create path
	sp = append(sp, dst)
	next_chain := prev_chain[dst]
	for next_chain.Chain_id != "" {
		sp = append([]string{next_chain.Chain_id}, sp...)
		next_chain = prev_chain[next_chain.Chain_id]
	}

	return sp, nil
}

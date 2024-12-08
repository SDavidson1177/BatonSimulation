package simulator

import (
	"context"
	"errors"
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

	prev_chain := make(map[string]string)

	sp := make([]string, 0)
	node := event_queue.Pop().(*DijkstraEvent)
	prev_chain[node.Chain] = ""
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
			if c_event != nil && c_event.(*DijkstraEvent).Distance > node.Distance+1 {
				c_dijk_event := c_event.(*DijkstraEvent)
				c_dijk_event.Distance = node.Distance + 1
				event_queue.Update(c_index)
				prev_chain[c_dijk_event.Chain] = node.Chain
			}
		}

		// Get next
		node = event_queue.Pop().(*DijkstraEvent)
	}

	// create path
	sp = append(sp, dst)
	next_chain := prev_chain[dst]
	for next_chain != "" {
		sp = append([]string{next_chain}, sp...)
		next_chain = prev_chain[next_chain]
	}

	return sp, nil
}

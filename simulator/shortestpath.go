package simulator

import (
	"context"
	"errors"
)

// GetShortestPath returns the shortest path from the source chain to the destination
// chain. This uses hop count to determine the length of a path.
func GetShortestPath(ctx context.Context, src string, dst string) ([]string, error) {
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

	sp := make([]string, 1)
	node := event_queue.Pop().(*DijkstraEvent)
	sp[0] = node.Chain
	for node.Chain != dst {
		// Update all neighbours
		for n := range state.Chains[node.Chain].neighbours {
			c_event, c_index := event_queue.Find(&DijkstraEvent{Chain: n}, cmp)
			if c_event != nil {
				c_dijk_event := c_event.(*DijkstraEvent)
				c_dijk_event.Distance = node.Distance + 1
				event_queue.Update(c_index)
			}
		}

		// Get next
		node = event_queue.Pop().(*DijkstraEvent)
		sp = append(sp, node.Chain)
	}

	return sp, nil
}

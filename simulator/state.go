package simulator

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const StateContextKey = "CTX_State"

const (
	IMPLICIT_HEIGHT          = 0
	IMPLICIT_HEIGHT_INTERVAL = 4000 // 4 seconds
)

// Global simulator state
type State struct {
	Seq        uint64
	Chains     map[string]*Chain
	Neighbours map[string][]*Chain

	// Add periodic events for implicit event loading
	implicit_intervals []uint32 // interval between implicit events in milliseconds
	implicit_tracker   []uint32 // time until next event in milliseconds

}

func NewState() *State {
	s := &State{Seq: 0, Chains: make(map[string]*Chain), Neighbours: make(map[string][]*Chain)}

	// Add info for implicit events
	s.implicit_intervals = make([]uint32, 1)
	s.implicit_tracker = make([]uint32, 1)

	s.implicit_intervals[IMPLICIT_HEIGHT] = IMPLICIT_HEIGHT_INTERVAL
	s.implicit_tracker[IMPLICIT_HEIGHT] = IMPLICIT_HEIGHT_INTERVAL

	return s
}

func (s *State) AddChain(ch *Chain) {
	s.Chains[ch.GetID()] = ch
}

func GetStateFromContext(ctx context.Context) (*State, error) {
	val := ctx.Value(GetContextKey(StateContextKey))
	if val == nil {
		return nil, errors.New("state context not present")
	}

	state, ok := val.(*State)
	if !ok {
		return nil, errors.New("cannot get state from context")
	}

	return state, nil
}

// Returns the time and type of next implicit event. Will return an error if there are no
// events that should be added to the loader.
func (s *State) GetNextImplicit(curr time.Time, max time.Time) (time.Time, int, error) {
	// find the minimum time
	min_time := -1
	min_event := -1
	for i, t := range s.implicit_tracker {
		if min_time == -1 || int(t) < min_time {
			min_time = int(t)
			min_event = i
		}
	}

	// Check if the next event can be added
	d, _ := time.ParseDuration(fmt.Sprintf("%dms", min_time))
	if curr.Add(d).After(max) {
		return time.Time{}, -1, errors.New("cannot add event")
	}

	// Update all of the event trackers
	for i := range s.implicit_tracker {
		if i == min_event {
			s.implicit_tracker[i] = s.implicit_intervals[i]
			continue
		}

		s.implicit_tracker[i] -= uint32(min_time)
	}

	return curr.Add(d), min_event, nil
}

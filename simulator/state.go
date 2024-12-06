package simulator

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"
)

const StateContextKey = "CTX_State"

const (
	IMPLICIT_HEIGHT          = 0
	IMPLICIT_HEIGHT_INTERVAL = 4000 // 4 seconds
)

type ImplicitEventTracker struct {
	Type     uint32
	Interval uint32
	Evnt     Event
}

// Global simulator state
type State struct {
	Seq    uint64
	Chains map[string]*Chain
	Time   time.Time

	// Add periodic events for implicit event loading
	implicit_tracker []ImplicitEventTracker // time until next event in milliseconds
}

func NewState() *State {
	s := &State{Seq: 0, Chains: make(map[string]*Chain)}
	return s
}

func (s *State) AddChain(ch *Chain) {
	s.Chains[ch.GetID()] = ch
}

// Initializes the implicit events. This must be called after all
// blockchains have been added.
func (s *State) InitializeImplicitEvents() {
	// Add info for implicit events: Height Update x Number of chains.
	num_chains := len(s.Chains)
	s.implicit_tracker = make([]ImplicitEventTracker, num_chains)

	i := 0
	for chain_name := range s.Chains {
		ri, _ := rand.Int(rand.Reader, big.NewInt(IMPLICIT_HEIGHT_INTERVAL))
		s.implicit_tracker[i] = ImplicitEventTracker{
			Type:     IMPLICIT_HEIGHT,
			Interval: uint32(ri.Int64()),
			Evnt:     NewHeightEvent(time.Now(), chain_name),
		}
		i++
	}

	fmt.Printf("Height Tracker: %v\n", s.implicit_tracker)
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
func (s *State) GetNextImplicit(curr time.Time, max time.Time) (Event, error) {
	// find the minimum time
	min_time := -1
	min_event := -1
	for i, t := range s.implicit_tracker {
		if min_time == -1 || int(t.Interval) < min_time {
			min_time = int(t.Interval)
			min_event = i
		}
	}

	// Check if the next event can be added
	d, _ := time.ParseDuration(fmt.Sprintf("%dms", min_time))
	if curr.Add(d).After(max) {
		return nil, errors.New("cannot add event")
	}

	var evnt Event

	// Update all of the event trackers
	for i := range s.implicit_tracker {
		if i == min_event {
			switch s.implicit_tracker[i].Type {
			case IMPLICIT_HEIGHT:
				s.implicit_tracker[i].Interval = IMPLICIT_HEIGHT_INTERVAL
				s.implicit_tracker[i].Evnt.AdjustTime(curr.Add(d))
			}

			evnt = s.implicit_tracker[i].Evnt.Copy()
			continue
		}

		s.implicit_tracker[i].Interval -= uint32(min_time)
	}
	return evnt, nil
}

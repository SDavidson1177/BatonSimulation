package simulator

import (
	"context"
	"errors"
)

const StateContextKey = "CTX_State"

// Global simulator state
type State struct {
	Seq        uint64
	Chains     map[string]*Chain
	Neighbours map[string][]*Chain
}

func NewState() *State {
	return &State{Seq: 0, Chains: make(map[string]*Chain), Neighbours: make(map[string][]*Chain)}
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

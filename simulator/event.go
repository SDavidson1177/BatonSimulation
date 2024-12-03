package simulator

import (
	"context"
	"fmt"
	"time"
)

type Event interface {
	Execute(ctx context.Context)
	Time() time.Time
	AddMsg()
}

// Test event

type TestEvent struct {
	event_time time.Time
}

func NewTestEvent() *TestEvent {
	return &TestEvent{event_time: time.Now()}
}

func (t *TestEvent) Execute(ctx context.Context) {
	fmt.Printf("test: %v\n", t.Time())
}

func (t *TestEvent) Time() time.Time {
	return t.event_time
}

func (t *TestEvent) AddMsg() {
	fmt.Printf("Adding test event with time: %v\n", t.Time())
}

// Update event
type UpdateEvent struct {
	event_time time.Time
	chain      string
	neighbour  string
}

func NewUpdateEvent(t time.Time, chain_id string, neighbour_id string) *UpdateEvent {
	return &UpdateEvent{event_time: t, chain: chain_id, neighbour: neighbour_id}
}

func (e *UpdateEvent) Execute(ctx context.Context) {

}

func (e *UpdateEvent) Time() time.Time {
	return e.event_time
}

func (e *UpdateEvent) AddMsg() {
	fmt.Printf("Adding update event with time: %v\n", e.Time())
}

// Height event
type HeightEvent struct {
	event_time time.Time
	chain      string
}

func NewHeightEvent(t time.Time, chain_id string) *HeightEvent {
	return &HeightEvent{event_time: t, chain: chain_id}
}

func (e *HeightEvent) Execute(ctx context.Context) {
	state, err := GetStateFromContext(ctx)
	if err != nil {
		return
	}

	if chain, ok := state.Chains[e.chain]; ok {
		fmt.Printf("Height of chain %s increased to %d at time %v\n", chain.GetID(), chain.IncHeight(), e.Time())
	}
}

func (e *HeightEvent) Time() time.Time {
	return e.event_time
}

func (e *HeightEvent) AddMsg() {
	fmt.Printf("Adding height event with time: %v\n", e.Time())
}

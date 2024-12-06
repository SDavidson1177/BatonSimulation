package simulator

import (
	"context"
	"fmt"
	"math"
	"time"
)

const (
	TEST_EVENT_TYPE    = 0
	UPDATE_EVENT_TYPE  = 1
	HEIGHT_EVENT_TYPE  = 2
	SEND_EVENT_TYPE    = 3
	DELIVER_EVENT_TYPE = 4
)

type Event interface {
	Execute(ctx context.Context)
	Type() uint64
	Copy() Event
	Time() time.Time
	AddMsg()
	SubEvents() []Event
	Following() []Event
	SetFollowing([]Event)
	AdjustTime(time.Time)
}

// Test event

type TestEvent struct {
	event_time time.Time
	following  []Event
}

func NewTestEvent(t time.Time) *TestEvent {
	return &TestEvent{event_time: t}
}

func (t *TestEvent) Execute(ctx context.Context) {
	fmt.Printf("test: %v\n", t.Time())
}

func (t *TestEvent) Type() uint64 {
	return TEST_EVENT_TYPE
}

func (t *TestEvent) Copy() Event {
	return NewTestEvent(t.Time())
}

func (t *TestEvent) Time() time.Time {
	return t.event_time
}

func (t *TestEvent) AddMsg() {
	fmt.Printf("Adding test event with time: %v\n", t.Time())
}

func (t *TestEvent) SubEvents() []Event {
	return nil
}

func (t *TestEvent) SetFollowing(events []Event) {
	t.following = events
}

func (t *TestEvent) Following() []Event {
	return t.following
}

func (t *TestEvent) AdjustTime(et time.Time) {
	t.event_time = et
}

// Update event
type UpdateEvent struct {
	event_time time.Time
	following  []Event
	chain      string
	neighbour  string
}

func NewUpdateEvent(t time.Time, chain_id string, neighbour_id string) *UpdateEvent {
	return &UpdateEvent{event_time: t, following: make([]Event, 0), chain: chain_id, neighbour: neighbour_id}
}

func (e *UpdateEvent) Execute(ctx context.Context) {
	state, err := GetStateFromContext(ctx)
	if err != nil {
		return
	}

	ch, ok := state.Chains[e.chain]
	if !ok {
		fmt.Printf("failed to update. Could not find chain %s\n", e.chain)
		return
	}

	var updated bool
	if updated, err = ch.UpdateView(e.neighbour); err != nil {
		fmt.Printf("could not update view. %s\n", err.Error())
	}

	// Since update happened, neighbour should exist
	n, _ := ch.GetNeighbour(e.neighbour)
	fmt.Printf("Updated chain %s to view chain %s at height %d: %v\n", e.chain, e.neighbour, n.GetHeight(), e.Time())

	// Enqueue next update if there is one to follow
	for _, follow := range e.Following() {
		// Enqueue event immediately if it is deliver, as the send
		// event should have scheduled the deliver event for the same
		// time as this update event. Otherwise, we may want to schdule
		// the next update event to run immediately after if this event
		// did not trigger an update.
		if (follow.Type() == UPDATE_EVENT_TYPE && !updated) || follow.Type() == DELIVER_EVENT_TYPE {
			// adjust time of next update event so that it is triggered immediately
			follow.AdjustTime(e.Time())
		}
		MainEventQueue.Enqueue(follow)
	}
}

func (e *UpdateEvent) Type() uint64 {
	return UPDATE_EVENT_TYPE
}

func (e *UpdateEvent) Copy() Event {
	return NewUpdateEvent(e.Time(), e.chain, e.neighbour)
}

func (e *UpdateEvent) Time() time.Time {
	return e.event_time
}

func (e *UpdateEvent) AddMsg() {
	fmt.Printf("Adding update event with time: %v\n", e.Time())
}

func (e *UpdateEvent) SubEvents() []Event {
	return nil
}

func (e *UpdateEvent) Following() []Event {
	return e.following
}

func (e *UpdateEvent) SetFollowing(events []Event) {
	e.following = events
}

func (e *UpdateEvent) AdjustTime(t time.Time) {
	e.event_time = t
}

// Height event
type HeightEvent struct {
	event_time time.Time
	following  []Event
	chain      string
}

func NewHeightEvent(t time.Time, chain_id string) *HeightEvent {
	return &HeightEvent{event_time: t, following: make([]Event, 0), chain: chain_id}
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

func (e *HeightEvent) Type() uint64 {
	return HEIGHT_EVENT_TYPE
}

func (e *HeightEvent) Copy() Event {
	return NewHeightEvent(e.Time(), e.chain)
}

func (e *HeightEvent) Time() time.Time {
	return e.event_time
}

func (e *HeightEvent) AddMsg() {
	fmt.Printf("Adding height event with time: %v\n", e.Time())
}

func (t *HeightEvent) SubEvents() []Event {
	return nil
}

func (e *HeightEvent) Following() []Event {
	return e.following
}

func (e *HeightEvent) SetFollowing(events []Event) {
	e.following = events
}

func (e *HeightEvent) AdjustTime(t time.Time) {
	e.event_time = t
}

// Send event
type SendEvent struct {
	event_time time.Time
	following  []Event
	src_chain  string
	hops       []string // chain hops not including the source chain
}

func NewSendEvent(t time.Time, src_chain string, hops []string) *SendEvent {
	return &SendEvent{event_time: t, following: make([]Event, 0), src_chain: src_chain, hops: hops}
}

func (e *SendEvent) Execute(ctx context.Context) {
	// Create update events
	if len(e.hops) < 1 {
		return
	}

	update_events := make([]Event, len(e.hops))
	a := e.src_chain
	for i := range e.hops {
		b := e.hops[i]
		d, _ := time.ParseDuration(fmt.Sprintf("%dms", int(math.Round(float64(i)*1.233*IMPLICIT_HEIGHT_INTERVAL))))
		update_events[i] = NewUpdateEvent(e.Time().Add(d), a, b)
		a = b

		// Add the following update event
		if i > 0 {
			update_events[i-1].SetFollowing([]Event{update_events[i]})
		}
	}

	// Add the deliver event
	update_events[len(update_events)-1].SetFollowing([]Event{NewDeliverEvent(
		update_events[len(update_events)-1].Time(),
		e.hops[len(e.hops)-1],
	)})

	// Only enqueue the first update event. The rest will be triggered as needed
	MainEventQueue.Enqueue(update_events[0])
}

func (e *SendEvent) Type() uint64 {
	return SEND_EVENT_TYPE
}

func (e *SendEvent) Copy() Event {
	return NewSendEvent(e.Time(), e.src_chain, e.hops)
}

func (e *SendEvent) Time() time.Time {
	return e.event_time
}

func (e *SendEvent) AddMsg() {
	fmt.Printf("Adding send event with time: %v\n", e.Time())
}

func (t *SendEvent) SubEvents() []Event {
	return nil
}

func (e *SendEvent) Following() []Event {
	return e.following
}

func (e *SendEvent) SetFollowing(events []Event) {
	e.following = events
}

func (e *SendEvent) AdjustTime(t time.Time) {
	e.event_time = t
}

// Deliver event
type DeliverEvent struct {
	event_time time.Time
	following  []Event
	chain      string
}

func NewDeliverEvent(t time.Time, chain_id string) *DeliverEvent {
	return &DeliverEvent{event_time: t, following: make([]Event, 0), chain: chain_id}
}

func (e *DeliverEvent) Execute(ctx context.Context) {
	state, err := GetStateFromContext(ctx)
	if err != nil {
		return
	}

	if chain, ok := state.Chains[e.chain]; ok {
		fmt.Printf("Delivering messages to chain %s at time %v\n", chain.GetID(), e.Time())
	}
}

func (e *DeliverEvent) Type() uint64 {
	return DELIVER_EVENT_TYPE
}

func (e *DeliverEvent) Copy() Event {
	return NewDeliverEvent(e.Time(), e.chain)
}

func (e *DeliverEvent) Time() time.Time {
	return e.event_time
}

func (e *DeliverEvent) AddMsg() {
	fmt.Printf("Adding deliver event with time: %v\n", e.Time())
}

func (t *DeliverEvent) SubEvents() []Event {
	return nil
}

func (e *DeliverEvent) Following() []Event {
	return e.following
}

func (e *DeliverEvent) SetFollowing(events []Event) {
	e.following = events
}

func (e *DeliverEvent) AdjustTime(t time.Time) {
	e.event_time = t
}

// Dijkstra event. Not to be loaded into main event queue. Just so that we can use event heap.
type DijkstraEvent struct {
	Distance int
	Chain    string
}

func NewDijkstraEvent(distance int, chain_id string) *DijkstraEvent {
	return &DijkstraEvent{Distance: distance, Chain: chain_id}
}

func (e *DijkstraEvent) Execute(ctx context.Context) {
}

func (e *DijkstraEvent) Type() uint64 {
	return DELIVER_EVENT_TYPE
}

func (e *DijkstraEvent) Copy() Event {
	return NewDijkstraEvent(e.Distance, e.Chain)
}

func (e *DijkstraEvent) Time() time.Time {
	return time.Unix(int64(e.Distance), 0)
}

func (e *DijkstraEvent) AddMsg() {
	fmt.Printf("Adding dijkstra event with distance and chain: %d, %s\n", e.Distance, e.Chain)
}

func (t *DijkstraEvent) SubEvents() []Event {
	return nil
}

func (e *DijkstraEvent) Following() []Event {
	return nil
}

func (e *DijkstraEvent) SetFollowing(events []Event) {
}

func (e *DijkstraEvent) AdjustTime(t time.Time) {
	e.Distance = int(t.Unix())
}

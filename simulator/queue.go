package simulator

import (
	"context"
	"errors"
	"time"
)

var MainEventQueue EventQueue
var EventLoader EventHeap

// Event Heap
type EventHeap struct {
	heap []Event
}

func NewEventHeap() *EventHeap {
	return &EventHeap{
		heap: make([]Event, 0),
	}
}

func (eh *EventHeap) parent(i int) int {
	if i >= len(eh.heap) {
		return -1
	}

	return (i - 1) / 2
}

func (eh *EventHeap) left(i int) int {
	if i*2+1 >= len(eh.heap) {
		return -1
	}

	return i*2 + 1
}

func (eh *EventHeap) right(i int) int {
	if i*2+2 >= len(eh.heap) {
		return -1
	}

	return i*2 + 2
}

func (eh *EventHeap) bubbleUp(i int) {
	if i >= len(eh.heap) {
		return
	}

	parent := eh.parent(i)
	if parent < 0 {
		return
	}

	for eh.heap[parent].Time().After(eh.heap[i].Time()) {
		tmp := eh.heap[i]
		eh.heap[i] = eh.heap[parent]
		eh.heap[parent] = tmp

		i = parent
		parent = eh.parent(i)
	}
}

func (eh *EventHeap) bubbleDown(i int) {
	heap_len := len(eh.heap)

	if i > heap_len {
		return
	}

	left, right := eh.left(i), eh.right(i)

	// Child exists
	for !(left < 0 && right < 0) {
		var min_event Event
		min_index := left
		if left < 0 || (right > 0 && eh.heap[left].Time().After(eh.heap[right].Time())) {
			min_index = right
			min_event = eh.heap[right]
		} else {
			min_event = eh.heap[left]
		}

		// no smaller child
		if eh.heap[i].Time().Before(min_event.Time()) {
			break
		}

		tmp := eh.heap[min_index]
		eh.heap[min_index] = eh.heap[i]
		eh.heap[i] = tmp
		i = min_index
		left, right = eh.left(i), eh.right(i)
	}
}

func (eh *EventHeap) Insert(event Event) {
	eh.heap = append(eh.heap, event)
	eh.bubbleUp(len(eh.heap) - 1)
}

func (eh *EventHeap) Top() Event {
	if len(eh.heap) == 0 {
		return nil
	}

	return eh.heap[0]
}

func (eh *EventHeap) Pop() Event {
	if len(eh.heap) == 0 {
		return nil
	}

	// fmt.Printf("HEAP POP: ")
	// for _, e := range eh.heap {
	// 	fmt.Printf("%v ", e.Time())
	// }
	// fmt.Printf("\n\n")

	top := eh.heap[0]
	eh.heap[0] = eh.heap[len(eh.heap)-1]
	eh.heap = eh.heap[:len(eh.heap)-1]
	eh.bubbleDown(0)

	return top
}

func (eh *EventHeap) Find(this_event Event, cmp func(Event, Event) bool) (Event, int) {
	for i := range eh.heap {
		if cmp(this_event, eh.heap[i]) {
			return eh.heap[i], i
		}
	}
	return nil, -1
}

func (eh *EventHeap) Update(index int) {
	// Check whether to bubble up or down
	parent := eh.parent(index)
	if parent < 0 || eh.heap[parent].Time().Before(eh.heap[index].Time()) {
		eh.bubbleDown(index)
		return
	}
	eh.bubbleUp(index)
}

// Event Queue
type EventQueue struct {
	queue *EventHeap

	BatonState *State
}

func NewQueue() *EventQueue {
	MainEventQueue = EventQueue{queue: &EventHeap{}, BatonState: NewState()}
	EventLoader = EventHeap{}
	return &MainEventQueue
}

// Should be called after adding all chains
func (e *EventQueue) Init() {
	e.BatonState.InitializeImplicitEvents()
}

func (e *EventQueue) Enqueue(event Event) {
	e.queue.Insert(event)
}

func (e *EventQueue) Step(ctx context.Context) error {
	event := e.queue.Pop()
	if event == nil {
		return errors.New("empty")
	}

	e.BatonState.Time = event.Time()
	event.Execute(ctx)

	return nil
}

// Add & Load events
func AddEventToLoad(event Event) {
	EventLoader.Insert(event)
	event.AddMsg()

	// Load sub events
	sub_events := event.SubEvents()
	for _, e := range sub_events {
		AddEventToLoad(e)
	}
}

// LoadEventsIntoQueue will load all the events added to the
// event loader into the event queue. This function will
// also add any necessary implicit event. For example, this
// will add events to increment the height of each blockchain.
func LoadEventsIntoQueue() error {
	var implicit_timer time.Time
	started := false

	for {
		event := EventLoader.Pop()
		if event == nil {
			// Check in case empty
			break
		}

		if !started {
			// initialize implicit timer to start at the same time as the first event
			implicit_timer = event.Time()
			started = true
		}

		MainEventQueue.Enqueue(event)

		// Get the next event
		next := EventLoader.Top()
		if next == nil {
			// no more events
			break
		}

		// Add implicit events
		evnt, err := MainEventQueue.BatonState.GetNextImplicit(implicit_timer, next.Time())
		for err == nil {
			MainEventQueue.Enqueue(evnt)
			implicit_timer = evnt.Time()
			next = EventLoader.Top()
			evnt, err = MainEventQueue.BatonState.GetNextImplicit(implicit_timer, next.Time())
		}
	}

	return nil
}

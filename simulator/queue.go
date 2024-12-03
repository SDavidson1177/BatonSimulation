package simulator

import (
	"context"
	"errors"
)

var MainEventQueue EventQueue
var EventLoader EventHeap

// Event Heap
type EventHeap struct {
	heap []Event
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

		tmp := min_event
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

func (eh *EventHeap) Pop() Event {
	if len(eh.heap) == 0 {
		return nil
	}

	top := eh.heap[0]
	eh.heap[0] = eh.heap[len(eh.heap)-1]
	eh.heap = eh.heap[:len(eh.heap)-1]
	eh.bubbleDown(0)

	return top
}

// Event Queue
type EventQueue struct {
	queue *EventHeap

	BatonState *State
}

func InitQueue() *EventQueue {
	MainEventQueue = EventQueue{queue: &EventHeap{}, BatonState: NewState()}
	EventLoader = EventHeap{}
	return &MainEventQueue
}

func (e *EventQueue) Enqueue(event Event) {
	e.queue.Insert(event)
}

func (e *EventQueue) Step(ctx context.Context) error {
	event := e.queue.Pop()
	if event == nil {
		return errors.New("empty")
	}
	event.Execute(ctx)

	return nil
}

// Add & Load events
func AddEventToLoad(event Event) {
	EventLoader.Insert(event)
	event.AddMsg()
}

// LoadEventsIntoQueue will load all the events added to the
// event loader into the event queue. This function will
// also add any necessary implicit event. For example, this
// will add events to increment the height of each blockchain.
func LoadEventsIntoQueue() error {
	for {
		event := EventLoader.Pop()
		if event == nil {
			break
		}

		MainEventQueue.Enqueue(event)
	}

	return nil
}

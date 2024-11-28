package simulator

var MainEvent EventQueue

type EventQueue struct {
	queue []Event
}

func InitQueue() *EventQueue {
	MainEvent = EventQueue{}
	return &MainEvent
}

func (e *EventQueue) Enqueue(event Event) {
	e.queue = append(e.queue, event)
}

func (e *EventQueue) Step() {
	if len(e.queue) == 0 {
		return
	}

	event := e.queue[0]
	event.Execute()
	e.queue = e.queue[1:]
}

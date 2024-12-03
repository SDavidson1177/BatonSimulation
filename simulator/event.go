package simulator

import (
	"fmt"
	"time"
)

type Event interface {
	Execute()
	Time() time.Time
}

type TestEvent struct {
	event_time time.Time
}

func NewTestEvent() *TestEvent {
	return &TestEvent{event_time: time.Now()}
}

func (t *TestEvent) Execute() {
	fmt.Printf("test: %v\n", t.Time())
}

func (t *TestEvent) Time() time.Time {
	return t.event_time
}

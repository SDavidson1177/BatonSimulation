package simulator

import "fmt"

type Event interface {
	Execute()
}

type TestEvent struct {
}

func (t *TestEvent) Execute() {
	fmt.Printf("test\n")
}

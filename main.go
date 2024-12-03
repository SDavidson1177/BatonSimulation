package main

import (
	"time"

	"github.com/SDavidson1177/ThroughputSim/simulator"
)

func main() {
	main_event := simulator.InitQueue()

	main_event.Enqueue(simulator.NewTestEvent())

	time.Sleep(2 * time.Second)

	main_event.Enqueue(simulator.NewTestEvent())

	main_event.Step()
	main_event.Step()
}

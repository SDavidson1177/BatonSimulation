package main

import "github.com/SDavidson1177/ThroughputSim/simulator"

func main() {
	main_event := simulator.InitQueue()

	main_event.Enqueue(&simulator.TestEvent{})
	main_event.Step()
}

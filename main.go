package main

import (
	"context"
	"time"

	"github.com/SDavidson1177/ThroughputSim/simulator"
)

func main() {
	ctx := context.Background()

	main_event := simulator.InitQueue()
	ctx = context.WithValue(ctx, simulator.GetContextKey(simulator.StateContextKey), main_event.BatonState)

	for i := 0; i < 4; i++ {
		simulator.AddEventToLoad(simulator.NewTestEvent())
		time.Sleep(time.Second)
	}

	simulator.LoadEventsIntoQueue()

	for main_event.Step(ctx) == nil {
	}
}

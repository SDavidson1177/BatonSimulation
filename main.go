package main

import (
	"context"
	"fmt"
	"time"

	"github.com/SDavidson1177/ThroughputSim/simulator"
)

func main() {
	ctx := context.Background()

	main_event := simulator.InitQueue()
	ctx = context.WithValue(ctx, simulator.GetContextKey(simulator.StateContextKey), main_event.BatonState)

	// Add blockchains
	for i := 0; i < 5; i++ {
		main_event.BatonState.AddChain(simulator.NewChain(fmt.Sprintf("baton-%d", i)))
	}

	// Add events
	for i := 0; i < 5; i++ {
		simulator.AddEventToLoad(simulator.NewTestEvent())
		time.Sleep(3 * time.Second)
	}

	simulator.LoadEventsIntoQueue()

	for main_event.Step(ctx) == nil {
	}
}

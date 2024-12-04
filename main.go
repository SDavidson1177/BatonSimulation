package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/SDavidson1177/ThroughputSim/simulator"
)

func readTopology(filename string) (map[string]*simulator.Chain, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)

	GetChainID := func(id string) string {
		return fmt.Sprintf("baton-%s", id)
	}

	chains := make(map[string]*simulator.Chain)

	// Iterate over every edge
	for scanner.Scan() {
		chain_pair := strings.Split(scanner.Text(), ",")
		if len(chain_pair) != 2 {
			return nil, errors.New("not enough chain pairs")
		}

		// Add both chains
		if _, ok := chains[GetChainID(chain_pair[0])]; !ok {
			chains[GetChainID(chain_pair[0])] = simulator.NewChain(GetChainID(chain_pair[0]))
		}

		if _, ok := chains[GetChainID(chain_pair[1])]; !ok {
			chains[GetChainID(chain_pair[1])] = simulator.NewChain(GetChainID(chain_pair[1]))
		}

		// Make chains neighbours of each other
		chains[GetChainID(chain_pair[0])].AddNeighbour(chains[GetChainID(chain_pair[1])])
		chains[GetChainID(chain_pair[1])].AddNeighbour(chains[GetChainID(chain_pair[0])])
	}

	return chains, nil
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("missing edges csv file")
		return
	}

	chains, err := readTopology(args[1])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	ctx := context.Background()

	main_event := simulator.InitQueue()
	ctx = context.WithValue(ctx, simulator.GetContextKey(simulator.StateContextKey), main_event.BatonState)

	// Add blockchains
	for _, chain := range chains {
		main_event.BatonState.AddChain(chain)
	}

	// Add events
	start_time := time.Now()
	d, _ := time.ParseDuration("5s")

	// Test shortest paths
	sp, err := simulator.GetShortestPath(ctx, "baton-1", "baton-3")
	if err != nil {
		return
	}

	simulator.AddEventToLoad(simulator.NewSendEvent(
		start_time,
		sp[0],
		sp[1:],
	))

	simulator.AddEventToLoad(simulator.NewSendEvent(
		start_time.Add(d),
		sp[0],
		sp[1:],
	))

	simulator.LoadEventsIntoQueue()

	fmt.Printf("SP: %v\n", sp)

	for main_event.Step(ctx) == nil {
	}
}

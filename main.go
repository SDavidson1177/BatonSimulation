package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/SDavidson1177/ThroughputSim/simulator"
)

// Reads in the blockchain topology from edges csv file.
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

// Generates and prints a list of send events
func genSends(ctx context.Context, send_interval uint32, jitter uint32, num_sends int) ([]*simulator.SendEvent, error) {
	if jitter >= send_interval {
		return nil, errors.New("jitter cannot be >= than send interval")
	}

	state, err := simulator.GetStateFromContext(ctx)
	if err != nil {
		return nil, err
	}

	base_time := time.Now()

	gen_start_time := func() time.Time {
		start_r := big.NewInt(int64(send_interval))
		r, err := rand.Int(rand.Reader, start_r)
		if err != nil {
			return base_time
		}

		d, _ := time.ParseDuration(fmt.Sprintf("%dms", r))
		return base_time.Add(d)
	}

	gen_send_time := func() time.Time {
		jitter_r := big.NewInt(int64(jitter))
		r, err := rand.Int(rand.Reader, jitter_r)
		if err != nil {
			return base_time
		}

		d, _ := time.ParseDuration(fmt.Sprintf("%dms", r.Int64()+int64(send_interval)))
		return base_time.Add(d)
	}

	// Create a priority queue for send event timing
	queue := simulator.NewEventHeap()
	for c1 := range state.Chains {
		for c2 := range state.Chains {
			if c1 != c2 {
				// Enqueue event
				queue.Insert(simulator.NewGenSendEvent(
					gen_start_time(),
					c1,
					c2,
				))
			}
		}
	}

	// Generate the events
	hops := make(map[string][]string)
	retval := make([]*simulator.SendEvent, 0)
	for i := 0; i < num_sends; i++ {
		next := queue.Pop()
		if next == nil {
			return retval, errors.New("queue empty")
		}

		gs_evnt := next.(*simulator.GenSendEvent)
		sp, ok := hops[fmt.Sprintf("%s-%s", gs_evnt.Src, gs_evnt.Dst)]
		if !ok {
			spi, err := simulator.GetShortestPath(ctx, gs_evnt.Src, gs_evnt.Dst)
			if err != nil {
				return retval, errors.New("unreachable chain")
			}
			sp = spi
		}

		new_event := simulator.NewSendEvent(
			gs_evnt.Time(),
			sp[0],
			sp[1:],
		)
		fmt.Printf("Scheduling Send: %s --> %s | Time %v | Path %v\n",
			gs_evnt.Src,
			gs_evnt.Dst,
			new_event.Time(),
			sp)
		retval = append(retval, new_event)

		base_time = gs_evnt.Time()
		gs_evnt.AdjustTime(gen_send_time())
		queue.Insert(gs_evnt)
	}

	return retval, nil
}

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Printf("Format: main.go [edges csv file] [command]\nCommand can be either 'sim' or 'gen_sends'")
		return
	}

	chains, err := readTopology(args[1])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	ctx := context.Background()

	main_event := simulator.NewQueue()
	ctx = context.WithValue(ctx, simulator.GetContextKey(simulator.StateContextKey), main_event.BatonState)

	// Add blockchains
	for _, chain := range chains {
		main_event.BatonState.AddChain(chain)
	}
	main_event.Init()

	sends, err := genSends(ctx, 5000, 2500, 10)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	// Add events
	for _, e := range sends {
		fmt.Printf("%v\n", e)
		simulator.AddEventToLoad(e)
	}

	simulator.LoadEventsIntoQueue()

	for main_event.Step(ctx) == nil {
	}
}

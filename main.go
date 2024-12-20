package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/SDavidson1177/ThroughputSim/simulator"
)

// Reads in the blockchain topology from edges csv file.
// The csv file should be structured as follows:
//
//	1,2
//	2,3
//	3,1
//
// Where the integers represent blockchain IDs, and the pairing
// represents an IBC connection.
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

// Generates a list of send events
// If the channel type is 'multi', the event type will be  simulator.SendEvent
// If the channel type is 'single', the event type will be simulator.SendSingleEvent
func genSends(ctx context.Context, send_interval uint32, jitter uint32, num_sends int, is_multi_channel bool) ([]simulator.Event, error) {
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

	// Get direct and hubs from context
	direct := ctx.Value(simulator.GetContextKey(simulator.DirectContextKey)).(bool)
	hub_chains := ctx.Value(simulator.GetContextKey(simulator.HubsContextKey)).(map[string]bool)

	fmt.Printf("Hub chains: %v\n", hub_chains)

	// Generate the events
	hops := make(map[string][]string)
	retval := make([]simulator.Event, 0)
	for i := 0; i < num_sends; i++ {
		next := queue.Pop()
		if next == nil {
			return retval, errors.New("queue empty")
		}

		gs_evnt := next.(*simulator.GenSendEvent)
		sp, ok := hops[fmt.Sprintf("%s-%s", gs_evnt.Src, gs_evnt.Dst)]
		if !ok {
			spi, err := simulator.GetShortestPath(ctx, gs_evnt.Src, gs_evnt.Dst, hub_chains)
			if err != nil {
				// Unreachable. Try another pair.
				i--
				continue
			}

			if !direct {
				// We are using baton. Therefore, get the Baton shortest path
				spi, _ = simulator.GetShortestPath(ctx, gs_evnt.Src, gs_evnt.Dst, make(map[string]bool))
			}
			sp = spi
			hops[fmt.Sprintf("%s-%s", gs_evnt.Src, gs_evnt.Dst)] = sp
		}

		var new_event simulator.Event

		if is_multi_channel {
			new_event = simulator.NewSendEvent(
				gs_evnt.Time(),
				sp[0],
				sp[1:],
			)
		} else {
			new_event = simulator.NewSendSingleEvent(
				gs_evnt.Time(),
				sp[0],
				sp[1:],
			)
		}

		retval = append(retval, new_event)

		base_time = gs_evnt.Time()
		gs_evnt.AdjustTime(gen_send_time())
		queue.Insert(gs_evnt)
	}

	return retval, nil
}

func main() {
	args := os.Args
	if len(args) < 4 {
		fmt.Printf(`Format: main.go [edges csv file] [channel_type] [send interval] [jitter] [number of sends] [direct] [hubs...]
		Channel type can be either 'single' or 'multi'
		'single' will assume single-hop channels, but 'multi' will allow for multi-hop channels
`)
		return
	}

	chains, err := readTopology(args[1])
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	channel_type := args[2]
	if channel_type != "multi" && channel_type != "single" {
		panic("channel type must be 'single' or 'multi'")
	}

	send_interval, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		panic("send interval not the correct format")
	}
	jitter, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		panic("jitter not the correct format")
	}
	number_of_sends, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		panic("'number of sends' not the correct format")
	}

	ctx := context.Background()

	main_event := simulator.NewQueue()
	ctx = context.WithValue(ctx, simulator.GetContextKey(simulator.StateContextKey), main_event.BatonState)

	direct := false
	if len(args) >= 6 && args[6] == "true" {
		direct = true
	}

	ctx = context.WithValue(ctx, simulator.GetContextKey(simulator.DirectContextKey), direct)

	hub_chains := make(map[string]bool)
	if len(args) >= 7 {
		for _, c := range args[7:] {
			hub_chains[c] = true
		}
	}
	ctx = context.WithValue(ctx, simulator.GetContextKey(simulator.HubsContextKey), hub_chains)

	// Add blockchains
	for _, chain := range chains {
		main_event.BatonState.AddChain(chain)
	}
	main_event.Init()

	sends, err := genSends(ctx, uint32(send_interval), uint32(jitter), int(number_of_sends), channel_type == "multi")
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return
	}

	// Add events
	for _, e := range sends {
		simulator.AddEventToLoad(e)
	}

	simulator.LoadEventsIntoQueue()

	for main_event.Step(ctx) == nil {
	}

	// Get all the max tx counts for each chain.
	// This indicates congestion
	max_congestion := 0
	max_con_chain := ""
	all_tx := 0
	for _, chain := range main_event.BatonState.Chains {
		fmt.Printf("Congestion: %s -- %d| total %d\n", chain.GetID(), chain.GetMaxTxCount(), chain.TotalTx())
		all_tx += chain.TotalTx()
		if chain.GetMaxTxCount() > max_congestion {
			max_con_chain = chain.GetID()
			max_congestion = chain.GetMaxTxCount()
		}
	}

	fmt.Printf("MOST congestion chain: %s -- %d\n", max_con_chain, max_congestion)
	fmt.Printf("Total Transactions: %d\n", all_tx)
}

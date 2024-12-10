package simulator

import "fmt"

type Chain struct {
	id     string
	height uint64

	// This chain's view of its neighbour
	view       map[string]uint64
	neighbours map[string]*Chain

	// Keep track of congestion
	maxTxCount int
	txCount    int
	totalTx    int
}

func NewChain(id string) *Chain {
	return &Chain{id: id, view: make(map[string]uint64), neighbours: make(map[string]*Chain)}
}

func (c *Chain) GetID() string {
	return c.id
}

func (c *Chain) GetView(chain_id string) uint64 {
	if v, ok := c.view[chain_id]; ok {
		return v
	}

	return 0
}

func (c *Chain) IncreaseTxCount() {
	c.txCount++
	c.totalTx++
}

func (c *Chain) ResetTxCount() {
	if c.txCount > c.maxTxCount {
		c.maxTxCount = c.txCount
	}
	c.txCount = 0
}

func (c *Chain) TotalTx() int {
	return c.totalTx
}

func (c *Chain) GetMaxTxCount() int {
	return c.maxTxCount
}

// UpdateView returns true when a client update was necessary
// to track the neighbour's new height. Otherwise, return false.
func (c *Chain) UpdateView(chain_id string) (bool, error) {
	if _, ok := c.view[chain_id]; !ok {
		return false, fmt.Errorf("cannot find chain %s for view update", chain_id)
	}

	if c.GetHeight() == c.neighbours[chain_id].GetView(c.GetID()) {
		return false, nil
	}
	c.neighbours[chain_id].view[c.GetID()] = c.GetHeight()

	return true, nil
}

func (c *Chain) GetHeight() uint64 {
	return c.height
}

func (c *Chain) IncHeight() uint64 {
	c.height++
	return c.height
}

func (c *Chain) AddNeighbour(ch *Chain) {
	c.neighbours[ch.GetID()] = ch
	c.view[ch.GetID()] = ch.GetHeight()
}

func (c *Chain) GetNeighbour(id string) (*Chain, bool) {
	n, ok := c.neighbours[id]
	if !ok {
		return nil, ok
	}

	return n, true
}

func (c *Chain) GetNeighbours() map[string]*Chain {
	return c.neighbours
}

package simulator

import "fmt"

type Chain struct {
	id     string
	height uint64

	// This chain's view of its neighbour
	view       map[string]uint64
	neighbours map[string]*Chain
}

func NewChain(id string) *Chain {
	return &Chain{id: id}
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

func (c *Chain) UpdateView(chain_id string) error {
	if _, ok := c.view[chain_id]; !ok {
		return fmt.Errorf("cannot find chain %s for view update", chain_id)
	}

	c.view[chain_id] = c.neighbours[chain_id].GetHeight()
	return nil
}

func (c *Chain) GetHeight() uint64 {
	return c.height
}

func (c *Chain) IncHeight() uint64 {
	c.height++
	return c.height
}

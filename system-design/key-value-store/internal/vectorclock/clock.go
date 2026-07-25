package vectorclock

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Clock struct {
	Counters map[string]uint64 `json:"counters"`
}

func New() *Clock {
	return &Clock{Counters: make(map[string]uint64)}
}

func (c *Clock) Increment(nodeID string) {
	c.Counters[nodeID]++
}

func (c *Clock) Merge(other *Clock) {
	for nodeID, counter := range other.Counters {
		if c.Counters[nodeID] < counter {
			c.Counters[nodeID] = counter
		}
	}
}

type Relation int

const (
	Before Relation = iota
	After
	Concurrent
	Equal
)

func (c *Clock) Compare(other *Clock) Relation {
	allNodes := make(map[string]bool)
	for n := range c.Counters {
		allNodes[n] = true
	}
	for n := range other.Counters {
		allNodes[n] = true
	}

	allGreaterOrEqual := true
	allLessOrEqual := true
	for n := range allNodes {
		selfVal := c.Counters[n]
		otherVal := other.Counters[n]
		if selfVal < otherVal {
			allGreaterOrEqual = false
		}
		if selfVal > otherVal {
			allLessOrEqual = false
		}
	}

	if allGreaterOrEqual && allLessOrEqual {
		return Equal
	}
	if allLessOrEqual {
		return Before
	}
	if allGreaterOrEqual {
		return After
	}
	return Concurrent
}

func (c *Clock) Copy() *Clock {
	cp := New()
	for k, v := range c.Counters {
		cp.Counters[k] = v
	}
	return cp
}

func (c *Clock) Marshal() ([]byte, error) {
	return json.Marshal(c.Counters)
}

func (c *Clock) Unmarshal(data []byte) error {
	return json.Unmarshal(data, &c.Counters)
}

func (c *Clock) String() string {
	keys := make([]string, 0, len(c.Counters))
	for k := range c.Counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%d", k, c.Counters[k]))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

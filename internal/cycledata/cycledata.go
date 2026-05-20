package cycledata

import (
	"sync"

	"github.com/tanenking/gsframe/gsinf"
)

type Data struct {
	C    chan struct{}
	Data interface{}
}

func (d *Data) GetData() interface{} {
	return d.Data
}
func (d *Data) Trigger() <-chan struct{} {
	return d.C
}

type Cycle struct {
	sync.Mutex
	seq      int32
	list     []*Data
	maxCount int32
}

func NewCycle(maxCount int32) gsinf.ICycle {
	cycle := &Cycle{Mutex: sync.Mutex{}, maxCount: maxCount}
	cycle.list = make([]*Data, maxCount)

	cycle.seq = 0
	cycle.list[cycle.seq] = &Data{C: make(chan struct{})}

	return cycle
}
func (c *Cycle) CurrSeq() int32 {
	return c.seq
}
func (c *Cycle) MaxCount() int32 {
	return c.maxCount
}
func (c *Cycle) GetCycleData(seq int32) gsinf.ICycleData {
	return c.list[seq]
}
func (c *Cycle) PublishData(data interface{}) {
	c.Lock()
	d := c.list[c.seq]
	c.seq++
	if c.seq >= c.maxCount {
		c.seq = 0
	}
	c.list[c.seq] = &Data{C: make(chan struct{})}
	c.Unlock()

	d.Data = data

	close(d.C)
}

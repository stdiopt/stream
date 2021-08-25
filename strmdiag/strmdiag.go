package strmdiag

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	strm "github.com/stdiopt/stream"
)

// Counter provides a count diagnostics of message passed
type Counter struct {
	mu       sync.Mutex
	mark     time.Time
	duration time.Duration
	counts   map[string]int
	dcount   int
	w        io.Writer
}

func NewCounter(w io.Writer, d time.Duration) *Counter {
	return &Counter{
		mark:     time.Now(),
		w:        w,
		duration: d,
	}
}

func (c *Counter) Add(v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.counts == nil {
		c.counts = map[string]int{}
	}
	now := time.Now()
	if now.After(c.mark.Add(c.duration)) {
		c.mark = now
		c.writeCount()
		c.dcount = 0
	}
	c.dcount++
	c.counts[fmt.Sprintf("%T", v)]++
}

func (c *Counter) WriteCount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeCount()
}

func (c *Counter) writeCount() {
	perSec := float64(c.dcount) / float64(c.duration) * float64(time.Second)
	fmt.Fprintf(c.w, "Processed messages: %v %.2f/s\n", c.counts, perSec)
}

func Count(d time.Duration) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		c := NewCounter(os.Stderr, d)
		defer c.WriteCount()
		return p.Consume(func(v interface{}) error {
			c.Add(v)
			return p.Send(v)
		})
	})
}

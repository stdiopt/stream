package strmdiag

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	strm "github.com/stdiopt/stream"
)

// TypeCounter provides a count diagnostics of message passed
type TypeCounter struct {
	mu       sync.Mutex
	mark     time.Time
	duration time.Duration
	counts   map[string]int
	dcount   int

	lastCheck time.Time
	w         io.Writer
}

func NewTypeCounter(w io.Writer, d time.Duration) *TypeCounter {
	return &TypeCounter{
		mark:      time.Now(),
		lastCheck: time.Now(),
		w:         w,
		duration:  d,
	}
}

func (c *TypeCounter) Add(v interface{}) {
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

func (c *TypeCounter) WriteCount() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeCount()
}

func (c *TypeCounter) writeCount() {
	dur := time.Since(c.lastCheck)
	c.lastCheck = time.Now()
	perSec := float64(c.dcount) / float64(dur) * float64(time.Second)
	buf := bytes.Buffer{}
	for k, v := range c.counts {
		fmt.Fprintf(&buf, "(%v: %s)", k, humanize(float64(v)))
	}
	fmt.Fprintf(c.w, "Processed messages: %v - %v/s\n", buf.String(), humanize(perSec))
}

func TypeCount(d time.Duration) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		c := NewTypeCounter(os.Stderr, d)
		defer c.WriteCount()
		return p.Consume(func(v interface{}) error {
			c.Add(v)
			return p.Send(v)
		})
	})
}

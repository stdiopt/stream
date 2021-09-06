package strmdiag

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	strm "github.com/stdiopt/stream"
)

type Metrics struct {
	sync.Mutex
	collectors []*collector
	set        map[string]*collector
	writer     io.Writer
	duration   time.Duration
	start      time.Time
	stop       chan struct{}
	done       chan struct{}
}

func NewMetrics(w io.Writer, d time.Duration) *Metrics {
	m := &Metrics{
		writer:   w,
		duration: d,
	}
	m.Start()
	return m
}

func (m *Metrics) Start() {
	if m.stop != nil {
		return
	}
	m.start = time.Now()

	m.stop = make(chan struct{})
	m.done = make(chan struct{})
	lastCheck := time.Now()
	go func() {
		ticker := time.NewTicker(m.duration)
		defer func() {
			ticker.Stop()
			m.tick(time.Since(lastCheck))
			close(m.done)
		}()
		for {
			select {
			case <-m.stop:
				return
			case <-ticker.C:
				m.tick(time.Since(lastCheck))
				lastCheck = time.Now()
			}
		}
	}()
}

func (m *Metrics) tick(dur time.Duration) {
	m.Lock()
	defer m.Unlock()

	if len(m.collectors) == 0 {
		return
	}

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "Metrics:\n")
	for _, c := range m.collectors {
		fmt.Fprintf(buf, "\t%v\n", c.Get(dur))
	}

	fmt.Fprint(m.writer, buf.String())
}

func (m *Metrics) Stop() {
	close(m.stop)
	<-m.done

	m.Lock()
	defer m.Unlock()
	m.collectors = nil
	m.set = nil

	fmt.Fprintf(m.writer, "Time elapsed: %v\n", time.Since(m.start).Round(time.Millisecond))
}

func (m *Metrics) collector(name, desc string, opts ...MetricOpt) *collector {
	m.Lock()
	defer m.Unlock()
	if m.set == nil {
		m.set = map[string]*collector{}
	}
	c, ok := m.set[name]
	if !ok {
		c = &collector{name: name, desc: desc}
		m.collectors = append(m.collectors, c)
		m.set[name] = c
	}
	for _, fn := range opts {
		fn(c)
	}
	return c
}

// Do we need to remove?
func (m *Metrics) remove(c *collector) {
	m.Lock()
	defer m.Unlock()
	for i, cc := range m.collectors {
		if c == cc {
			m.collectors = append(m.collectors[:i], m.collectors[i+1:]...)
			break
		}
	}
}

func (m *Metrics) Count(name string, opts ...MetricOpt) strm.Pipe {
	c := m.collector(name, "messages processed", opts...)
	return strm.Func(func(p strm.Proc) error {
		// defer c.Close()
		return p.Consume(func(v interface{}) error {
			c.Add(1)
			return p.Send(v)
		})
	})
}

func (m *Metrics) CountBytes(name string, opts ...MetricOpt) strm.Pipe {
	c := m.collector(name, "bytes processed", opts...)
	return strm.Func(func(p strm.Proc) error {
		return p.Consume(func(b []byte) error {
			c.Add(len(b))
			return p.Send(b)
		})
	})
}

type collector struct {
	name      string
	desc      string
	lastCount uint64
	count     uint64
	total     uint64
}

func (c *collector) Name() string {
	return c.name
}

func (c *collector) Get(dur time.Duration) string {
	count := atomic.LoadUint64(&c.count)
	delta := count - c.lastCount
	c.lastCount = count
	perSec := float64(delta) / float64(dur) * float64(time.Second)

	var scount string
	fc := float64(count)
	if c.total != 0 {
		fsz := float64(c.total)
		scount = fmt.Sprintf("%v/%v (%.2f%%)", humanize(fc), humanize(fsz), fc/fsz*100)
	} else {
		scount = fmt.Sprintf("%v", humanize(fc))
	}

	return fmt.Sprintf("[%s] %s: %v - %v/s",
		c.name, c.desc,
		scount,
		humanize(perSec),
	)
}

func (c *collector) Add(n int) {
	atomic.AddUint64(&c.count, uint64(n))
}

type MetricOpt func(c *collector)

func WithTotal(n uint64) MetricOpt {
	return func(c *collector) {
		c.total = n
	}
}

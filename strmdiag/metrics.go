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

	start     time.Time
	lastCheck time.Time
}

func NewMetrics(w io.Writer, d time.Duration) *Metrics {
	m := &Metrics{
		writer:   w,
		duration: d,
		start:    time.Now(),
	}
	return m
}

func (m *Metrics) tick() {
	m.Lock()
	defer m.Unlock()
	dur := time.Since(m.lastCheck)
	if dur < m.duration {
		return
	}
	m.lastCheck = time.Now()

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

func (m *Metrics) Done() {
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
		c = &collector{
			metrics: m,
			name:    name,
			desc:    desc,
		}
		m.collectors = append(m.collectors, c)
		m.set[name] = c
	}
	for _, fn := range opts {
		fn(c)
	}
	return c
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
	metrics   *Metrics
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
	c.metrics.tick()
}

type MetricOpt func(c *collector)

func WithTotal(n uint64) MetricOpt {
	return func(c *collector) {
		c.total = n
	}
}

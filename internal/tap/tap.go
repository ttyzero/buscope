// Package tap records ttybus traffic per channel.
package tap

import (
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	SparkBuckets = 24
	Bucket       = 150 * time.Millisecond
	LogCap       = 200
	PulseFrames  = 8
)

// Msg is one payload on a channel.
type Msg struct {
	At   time.Time
	Body string
}

// Chan is live stats for one channel.
type Chan struct {
	Name  string
	Total int
	Last  time.Time
	Pulse int // frames of highlight remaining
	Log   []Msg

	buckets []int
	cur     int
	stepAt  time.Time
}

// Tap is a set of channels plus a global sparkline.
type Tap struct {
	mu       sync.Mutex
	chans    map[string]*Chan
	order    []string // first-seen order, plus activity bump
	global   *Chan
	selected string
}

func New() *Tap {
	now := time.Now()
	return &Tap{
		chans:  map[string]*Chan{},
		global: newChan("*", now),
	}
}

func newChan(name string, now time.Time) *Chan {
	return &Chan{
		Name:    name,
		buckets: make([]int, SparkBuckets),
		stepAt:  now,
	}
}

// Ensure makes sure name exists (from LIST / --watch) without counting a hit.
func (t *Tap) Ensure(name string) {
	if name == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.ensureLocked(name, time.Now())
}

func (t *Tap) ensureLocked(name string, now time.Time) *Chan {
	c, ok := t.chans[name]
	if ok {
		return c
	}
	c = newChan(name, now)
	t.chans[name] = c
	t.order = append(t.order, name)
	if t.selected == "" {
		t.selected = name
	}
	return c
}

// Hit records a message.
func (t *Tap) Hit(name, body string, at time.Time) {
	if at.IsZero() {
		at = time.Now()
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	c := t.ensureLocked(name, at)
	c.hit(body, at)
	t.global.hit(body, at)
}

func (c *Chan) hit(body string, at time.Time) {
	c.advance(at)
	c.Total++
	c.Last = at
	c.Pulse = PulseFrames
	c.buckets[c.cur]++
	c.Log = append(c.Log, Msg{At: at, Body: body})
	if len(c.Log) > LogCap {
		c.Log = c.Log[len(c.Log)-LogCap:]
	}
}

func (c *Chan) advance(now time.Time) {
	if c.stepAt.IsZero() {
		c.stepAt = now
		return
	}
	n := int(now.Sub(c.stepAt) / Bucket)
	if n <= 0 {
		return
	}
	if n > SparkBuckets {
		n = SparkBuckets
		for i := range c.buckets {
			c.buckets[i] = 0
		}
		c.cur = 0
		c.stepAt = now
		return
	}
	for i := 0; i < n; i++ {
		c.cur = (c.cur + 1) % SparkBuckets
		c.buckets[c.cur] = 0
	}
	c.stepAt = c.stepAt.Add(time.Duration(n) * Bucket)
}

// Tick advances buckets and decays pulses.
func (t *Tap) Tick(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.global.advance(now)
	if t.global.Pulse > 0 {
		t.global.Pulse--
	}
	for _, c := range t.chans {
		c.advance(now)
		if c.Pulse > 0 {
			c.Pulse--
		}
	}
}

// Names returns channel names, most recently active first, then alpha.
func (t *Tap) Names() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := append([]string{}, t.order...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := t.chans[out[i]], t.chans[out[j]]
		if a.Last.Equal(b.Last) {
			return a.Name < b.Name
		}
		return a.Last.After(b.Last)
	})
	return out
}

// Snapshot copies one channel (Log is a shallow copy of the slice header).
func (t *Tap) Snapshot(name string) Chan {
	t.mu.Lock()
	defer t.mu.Unlock()
	c := t.chans[name]
	if c == nil {
		return Chan{Name: name}
	}
	return copyChan(c)
}

func (t *Tap) Global() Chan {
	t.mu.Lock()
	defer t.mu.Unlock()
	return copyChan(t.global)
}

func copyChan(c *Chan) Chan {
	out := *c
	out.buckets = append([]int{}, c.buckets...)
	out.Log = append([]Msg{}, c.Log...)
	return out
}

func (t *Tap) Selected() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.selected
}

func (t *Tap) Select(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.chans[name]; ok {
		t.selected = name
	}
}

func (t *Tap) SelectDelta(delta int) {
	names := t.Names()
	if len(names) == 0 {
		return
	}
	cur := t.Selected()
	idx := 0
	for i, n := range names {
		if n == cur {
			idx = i
			break
		}
	}
	idx = (idx + delta) % len(names)
	if idx < 0 {
		idx += len(names)
	}
	t.Select(names[idx])
}

func (t *Tap) ClearLog(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if c := t.chans[name]; c != nil {
		c.Log = nil
	}
}

// Spark renders a Unicode sparkline of the last SparkBuckets buckets.
func (c Chan) Spark(width int) string {
	if width < 1 {
		return ""
	}
	data := c.orderedBuckets()
	if width > len(data) {
		width = len(data)
	}
	data = data[len(data)-width:]
	max := 0
	for _, n := range data {
		if n > max {
			max = n
		}
	}
	bars := []rune("▁▂▃▄▅▆▇█")
	var b strings.Builder
	for _, n := range data {
		if max == 0 || n == 0 {
			b.WriteRune('▁')
			continue
		}
		i := (n * (len(bars) - 1)) / max
		if i >= len(bars) {
			i = len(bars) - 1
		}
		if i < 0 {
			i = 0
		}
		b.WriteRune(bars[i])
	}
	return b.String()
}

func (c Chan) orderedBuckets() []int {
	out := make([]int, SparkBuckets)
	if len(c.buckets) != SparkBuckets {
		return out
	}
	for i := 0; i < SparkBuckets; i++ {
		out[i] = c.buckets[(c.cur+1+i)%SparkBuckets]
	}
	return out
}

// Rate is messages in the last ~1s (7 buckets).
func (c Chan) Rate() int {
	n := 0
	b := c.orderedBuckets()
	k := 7
	if k > len(b) {
		k = len(b)
	}
	for _, v := range b[len(b)-k:] {
		n += v
	}
	return n
}

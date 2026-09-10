package tap

import (
	"strings"
	"testing"
	"time"
)

func TestHitAndSpark(t *testing.T) {
	tp := New()
	now := time.Now()
	tp.Hit("files", "/tmp/a.go", now)
	tp.Hit("files", "/tmp/b.go", now)
	tp.Hit("cwd", "/tmp", now.Add(10*time.Millisecond))
	c := tp.Snapshot("files")
	if c.Total != 2 {
		t.Fatalf("total %d", c.Total)
	}
	if len(c.Log) != 2 || c.Log[1].Body != "/tmp/b.go" {
		t.Fatalf("log %+v", c.Log)
	}
	sp := c.Spark(8)
	if len([]rune(sp)) != 8 {
		t.Fatalf("spark %q", sp)
	}
	if !strings.ContainsAny(sp, "▂▃▄▅▆▇█") && c.Total > 0 {
		// last bucket should be non-empty → some bar above ▁
		if sp[len(sp)-1:] == "▁" {
			t.Fatalf("expected activity in last bucket: %q", sp)
		}
	}
	g := tp.Global()
	if g.Total != 3 {
		t.Fatalf("global %d", g.Total)
	}
}

func TestEnsureSelect(t *testing.T) {
	tp := New()
	tp.Ensure("files")
	tp.Ensure("cwd")
	if tp.Selected() != "files" {
		t.Fatalf("selected %q", tp.Selected())
	}
	tp.SelectDelta(1)
	if tp.Selected() != "cwd" {
		t.Fatalf("delta %q", tp.Selected())
	}
	tp.SelectDelta(1)
	if tp.Selected() != "files" {
		t.Fatalf("wrap %q", tp.Selected())
	}
}

func TestTickDecaysPulse(t *testing.T) {
	tp := New()
	now := time.Now()
	tp.Hit("files", "x", now)
	if tp.Snapshot("files").Pulse != PulseFrames {
		t.Fatal("pulse")
	}
	for i := 0; i < PulseFrames; i++ {
		tp.Tick(now.Add(time.Duration(i) * Bucket))
	}
	if p := tp.Snapshot("files").Pulse; p != 0 {
		t.Fatalf("pulse %d", p)
	}
}

func TestLogCap(t *testing.T) {
	tp := New()
	now := time.Now()
	for i := 0; i < LogCap+10; i++ {
		tp.Hit("files", "m", now)
	}
	if n := len(tp.Snapshot("files").Log); n != LogCap {
		t.Fatalf("cap %d", n)
	}
}

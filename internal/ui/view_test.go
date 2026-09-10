package ui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/ttyzero/buscope/internal/tap"
	"github.com/ttyzero/commons/theme"
	ttybus "github.com/ttyzero/ttybus/client"
)

func TestRenderFits(t *testing.T) {
	m := New(Options{Theme: theme.Parse("dark"), Watch: []string{"files"}})
	m.width, m.height = 48, 24
	m.busOK = true
	m.peers = []ttybus.Peer{
		{Name: "gitwing", Subs: []string{"files"}},
		{Name: "nvim"},
	}
	now := time.Now()
	m.tap.Hit("files", "/code/ttybus/README.md", now)
	m.tap.Hit("files", "/code/navehnet/README.md", now.Add(time.Millisecond))
	m.tap.Ensure("cwd")
	out := m.render()
	t.Log("\n" + out)
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 48 {
			t.Errorf("line %d width %d > 48: %q", i, w, line)
		}
	}
	for _, need := range []string{"buscope", "files", "gitwing"} {
		if !strings.Contains(out, need) {
			t.Errorf("missing %q", need)
		}
	}
}

func TestSparklineWidth(t *testing.T) {
	c := tap.Chan{}
	c = func() tap.Chan {
		tp := tap.New()
		tp.Hit("files", "a", time.Now())
		return tp.Snapshot("files")
	}()
	s := c.Spark(12)
	if lipgloss.Width(s) != 12 && len([]rune(s)) != 12 {
		t.Fatalf("%q width %d", s, lipgloss.Width(s))
	}
}

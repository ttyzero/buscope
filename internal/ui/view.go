package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/ttyzero/buscope/internal/tap"
	ttybus "github.com/ttyzero/ttybus/client"
)

func (m model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = "buscope"
	return v
}

func (m model) render() string {
	w, h := m.width, m.height
	if w < 24 {
		w = 24
	}
	if h < 12 {
		h = 12
	}
	st := newStyles(m.pal, w, h, m.borderless)
	inner := innerWidth(w, m.borderless)
	innerH := innerHeight(h, m.borderless)

	var b strings.Builder
	write := func(s string) {
		if s == "" {
			return
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s)
	}

	write(m.headerLine(st, inner))
	write(m.peerLine(st, inner))
	g := m.tap.Global()
	write(st.spark.Render(g.Spark(inner)))
	write(st.rule.Render(hairline(inner)))

	names := m.tap.Names()
	listH := m.listHeight(innerH, len(names))
	write(m.channelList(st, inner, names, listH))
	write(st.rule.Render(hairline(inner)))

	sel := m.tap.Selected()
	snap := m.tap.Snapshot(sel)
	used := 1 + 1 + 1 + 1 + listH + 1 // header, peers, spark, rule, list, rule
	remain := innerH - used - 2       // input + help
	if remain < 3 {
		remain = 3
	}
	write(m.logBlock(st, inner, snap, remain))
	write(m.inputLine(st, inner, sel))

	help := st.subtle.Render("j/k  i pub  c clr  q")
	body := b.String()
	used = strings.Count(body, "\n") + 1
	if used < innerH {
		body += strings.Repeat("\n", innerH-used) + help
	}
	return st.box.Render(body)
}

func (m model) headerLine(st styles, inner int) string {
	g := m.tap.Global()
	dot := st.subtle.Render("·")
	if m.busOK {
		if g.Pulse > 0 {
			dot = st.accent.Render("●")
		} else {
			dot = st.ok.Render("●")
		}
	} else if m.busErr != "" {
		dot = st.bad.Render("●")
	}
	title := st.title.Render("buscope")
	n := compactInt(g.Total)
	tail := st.muted.Render(n)
	room := inner - lipgloss.Width(title) - 2 - lipgloss.Width(n)
	if room < 0 {
		return title + " " + dot
	}
	return title + " " + dot + strings.Repeat(" ", room) + tail
}

func (m model) peerLine(st styles, inner int) string {
	if !m.busOK {
		return st.subtle.Render("no bus yet")
	}
	names := peerNames(m.peers)
	if len(names) == 0 {
		return st.muted.Render("0 peers")
	}
	line := strings.Join(names, " · ")
	prefix := fmt.Sprintf("%d  ", len(names))
	return st.muted.Render(ellipsize(prefix+line, inner))
}

func peerNames(peers []ttybus.Peer) []string {
	var out []string
	for _, p := range peers {
		n := p.Name
		if n == "" {
			n = p.ID
		}
		if n == "buscope" {
			continue
		}
		out = append(out, n)
	}
	return out
}

func (m model) listHeight(innerH, nchan int) int {
	// Reserve ~half the pane below the list for the log.
	h := innerH / 3
	if h < 3 {
		h = 3
	}
	if h > 10 {
		h = 10
	}
	if nchan > 0 && h > nchan {
		h = nchan
	}
	if h < 1 {
		h = 1
	}
	return h
}

func (m model) channelList(st styles, inner int, names []string, rows int) string {
	if len(names) == 0 {
		return st.subtle.Render("(no channels)")
	}
	sel := m.tap.Selected()
	idx := 0
	for i, n := range names {
		if n == sel {
			idx = i
			break
		}
	}
	start := m.offset
	if idx < start {
		start = idx
	}
	if idx >= start+rows {
		start = idx - rows + 1
	}
	if start < 0 {
		start = 0
	}

	var lines []string
	for i := start; i < start+rows && i < len(names); i++ {
		c := m.tap.Snapshot(names[i])
		lines = append(lines, m.channelRow(st, inner, c, names[i] == sel))
	}
	return strings.Join(lines, "\n")
}

func (m model) channelRow(st styles, inner int, c tap.Chan, selected bool) string {
	mark := " "
	if selected {
		mark = "▸"
	}
	name := c.Name
	count := compactInt(c.Total)
	sparkW := 8
	if inner > 36 {
		sparkW = 12
	}
	spark := c.Spark(sparkW)
	// mark(1) space name count space spark
	fixed := 1 + 1 + 1 + lipgloss.Width(count) + 1 + sparkW
	nameW := inner - fixed
	if nameW < 4 {
		nameW = 4
		spark = ""
	}
	name = ellipsize(name, nameW)
	pad := nameW - lipgloss.Width(name)
	if pad < 0 {
		pad = 0
	}
	row := mark + " " + name + strings.Repeat(" ", pad) + " " + count
	if spark != "" {
		row += " " + spark
	}
	if selected {
		if c.Pulse > 0 {
			return st.sel.Foreground(m.pal.Accent).Render(ellipsize(row, inner))
		}
		return st.sel.Width(inner).Render(ellipsize(row, inner))
	}
	if c.Pulse > 0 {
		return st.accent.Render(ellipsize(row, inner))
	}
	// colour the spark
	plain := mark + " " + st.text.Render(name) + strings.Repeat(" ", pad) + " " + st.muted.Render(count)
	if spark != "" {
		plain += " " + st.spark.Render(spark)
	}
	return plain
}

func (m model) logBlock(st styles, inner int, c tap.Chan, rows int) string {
	title := c.Name
	if title == "" {
		title = "—"
	}
	head := st.accent2.Render(ellipsize(title, inner))
	if rows <= 1 {
		return head
	}
	bodyRows := rows - 1
	logs := c.Log
	var lines []string
	lines = append(lines, head)
	if len(logs) == 0 {
		lines = append(lines, st.subtle.Render("waiting for messages"))
	} else {
		start := len(logs) - bodyRows - m.logOff
		if start < 0 {
			start = 0
		}
		end := start + bodyRows
		if end > len(logs) {
			end = len(logs)
		}
		for _, msg := range logs[start:end] {
			lines = append(lines, formatLog(st, inner, msg))
		}
	}
	for len(lines) < rows {
		lines = append(lines, "")
	}
	if len(lines) > rows {
		lines = lines[:rows]
	}
	return strings.Join(lines, "\n")
}

func formatLog(st styles, inner int, msg tap.Msg) string {
	ts := msg.At.Format("15:04:05")
	prefix := ts + "  "
	room := inner - lipgloss.Width(prefix)
	body := ellipsize(strings.ReplaceAll(msg.Body, "\t", " "), room)
	return st.subtle.Render(ts) + "  " + st.payload.Render(body)
}

func (m model) inputLine(st styles, inner int, sel string) string {
	prompt := st.input.Render("▸")
	if m.focus == "input" {
		cur := "█"
		if m.frame%6 < 3 {
			cur = "▓"
		}
		text := m.input + cur
		return prompt + " " + st.text.Render(ellipsize(text, inner-2))
	}
	ph := "i to pub"
	if sel != "" {
		ph = "i pub · " + sel
	}
	return prompt + " " + st.subtle.Render(ellipsize(ph, inner-2))
}

package ui

import (
	"charm.land/lipgloss/v2"

	"github.com/ttyzero/commons/theme"
)

type styles struct {
	box     lipgloss.Style
	title   lipgloss.Style
	text    lipgloss.Style
	muted   lipgloss.Style
	subtle  lipgloss.Style
	ok      lipgloss.Style
	bad     lipgloss.Style
	warn    lipgloss.Style
	accent  lipgloss.Style
	accent2 lipgloss.Style
	sel     lipgloss.Style
	spark   lipgloss.Style
	rule    lipgloss.Style
	input   lipgloss.Style
	payload lipgloss.Style
}

func newStyles(p theme.Palette, width, height int, borderless bool) styles {
	w, h := width, height
	if w < 20 {
		w = 20
	}
	if h < 10 {
		h = 10
	}
	box := lipgloss.NewStyle().
		Width(w).
		Height(h).
		MaxHeight(h).
		Padding(0, 1)
	if !borderless {
		box = box.Border(lipgloss.RoundedBorder()).BorderForeground(p.Border)
	}
	return styles{
		box:     box,
		title:   lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		text:    lipgloss.NewStyle().Foreground(p.Text),
		muted:   lipgloss.NewStyle().Foreground(p.Muted),
		subtle:  lipgloss.NewStyle().Foreground(p.Subtle),
		ok:      lipgloss.NewStyle().Foreground(p.Success),
		bad:     lipgloss.NewStyle().Foreground(p.Danger),
		warn:    lipgloss.NewStyle().Foreground(p.Warn),
		accent:  lipgloss.NewStyle().Foreground(p.Accent),
		accent2: lipgloss.NewStyle().Foreground(p.Accent2),
		sel: lipgloss.NewStyle().
			Foreground(p.Text).
			Bold(true).
			Background(p.Heat[1]),
		spark:   lipgloss.NewStyle().Foreground(p.Success),
		rule:    lipgloss.NewStyle().Foreground(p.Subtle),
		input:   lipgloss.NewStyle().Foreground(p.Accent2),
		payload: lipgloss.NewStyle().Foreground(p.Muted),
	}
}

func innerWidth(termWidth int, borderless bool) int {
	w := termWidth - 2
	if !borderless {
		w -= 2
	}
	if w < 8 {
		return 8
	}
	return w
}

func innerHeight(termHeight int, borderless bool) int {
	h := termHeight
	if !borderless {
		h -= 2
	}
	if h < 6 {
		return 6
	}
	return h
}

package ui

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

func compactInt(n int) string {
	switch {
	case n < 1000:
		return fmt.Sprintf("%d", n)
	case n < 10_000:
		return trim1(float64(n)/1000) + "k"
	case n < 1_000_000:
		return fmt.Sprintf("%dk", n/1000)
	default:
		return trim1(float64(n)/1_000_000) + "m"
	}
}

func trim1(f float64) string {
	s := fmt.Sprintf("%.1f", f)
	return strings.TrimSuffix(s, ".0")
}

func ellipsize(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return runewidth.Truncate(s, max, "…")
}

func hairline(width int) string {
	if width < 1 {
		return ""
	}
	return strings.Repeat("─", width)
}

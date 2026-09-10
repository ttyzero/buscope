package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/ttyzero/buscope/internal/ui"
	"github.com/ttyzero/commons/theme"
)

var version = "dev"

func main() {
	fs := flag.NewFlagSet("buscope", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	themeFlag := fs.String("theme", "", "auto|dark|light|nord|… (TTYTHEME, then CLITHEME)")
	borderless := fs.Bool("borderless", false, "no rounded frame (TTYBORDERLESS=1)")
	socket := fs.String("socket", "", "ttybus socket path")
	bus := fs.String("bus", "", "named ttybus")
	watch := fs.String("watch", "files", "comma-separated channels to always subscribe")
	showVersion := fs.Bool("version", false, "print version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `buscope — visual inspector for ttybus

Discovers channels from connected peers, tails their traffic, and
lets you publish a test event. Same $TTYTHEME as gitwing.

Usage:
  buscope [flags]

Flags:
`)
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
Theme (all companion panes honor this):
  --theme, $TTYTHEME, $CLITHEME, then OSC 11 / $COLORFGBG
  Values: auto, dark, light, or a palette (nord, catppuccin, …)
  --borderless / $TTYBORDERLESS=1  no rounded frame

Keys:
  j/k  select channel     i  publish to selected
  c    clear log          r  refresh peers
  q    quit
`)
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}
	if *showVersion {
		fmt.Println("buscope", version)
		return
	}
	if *socket != "" {
		_ = os.Setenv("TTYBUS_SOCKET", *socket)
	}
	if *bus != "" {
		_ = os.Setenv("TTYBUS_BUS", *bus)
	}

	var chans []string
	for _, p := range strings.Split(*watch, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			chans = append(chans, p)
		}
	}

	m := ui.New(ui.Options{
		Theme:      theme.Resolve(*themeFlag),
		Watch:      chans,
		Borderless: theme.Borderless(*borderless),
	})
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

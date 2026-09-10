package ui

import (
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/ttyzero/buscope/internal/tap"
	"github.com/ttyzero/commons/connect"
	"github.com/ttyzero/commons/theme"
	ttybus "github.com/ttyzero/ttybus/client"
)

const (
	tickEvery      = 150 * time.Millisecond
	listEvery      = 1 * time.Second
	reconnectEvery = 3 * time.Second
)

// Options configure the inspector.
type Options struct {
	Theme      theme.Spec
	Watch      []string
	Borderless bool
}

type model struct {
	spec       theme.Spec
	pal        theme.Palette
	borderless bool

	width, height int

	watch []string

	client   *ttybus.Client
	busOK    bool
	busErr   string
	lastDial time.Time
	lastList time.Time

	tap *tap.Tap

	subs map[string]<-chan string

	peers []ttybus.Peer

	focus  string // "list" or "input"
	input  string
	offset int // channel list scroll
	logOff int // 0 = follow tail; >0 lines up

	frame int
}

type connectedMsg struct {
	client *ttybus.Client
}
type connectErr struct{ err error }
type listMsg struct {
	peers []ttybus.Peer
	err   error
}
type subsStarted struct {
	items []subItem
}
type subItem struct {
	name string
	ch   <-chan string
}
type traffic struct {
	name string
	body string
	at   time.Time
}
type subClosed struct{ name string }
type tickMsg time.Time
type busClosed struct{}

func New(opts Options) model {
	watch := opts.Watch
	if len(watch) == 0 {
		watch = []string{"files"}
	}
	tp := tap.New()
	for _, w := range watch {
		tp.Ensure(w)
	}
	tp.Ensure(theme.Channel)
	return model{
		spec:       opts.Theme,
		pal:        theme.New(opts.Theme.InitialDark()),
		borderless: opts.Borderless,
		width:      48,
		height:     24,
		watch:      watch,
		tap:        tp,
		subs:       map[string]<-chan string{},
		focus:      "list",
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		connectCmd(),
		tickCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.spec = m.spec.ApplyOSC(msg.IsDark())
		m.pal = theme.New(m.spec.Polarity != "light")
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)

	case connectedMsg:
		if m.client != nil {
			_ = m.client.Close()
		}
		m.client = msg.client
		m.busOK = true
		m.busErr = ""
		m.subs = map[string]<-chan string{}
		return m, tea.Batch(listCmd(m.client), m.subMissing())

	case connectErr:
		m.busOK = false
		m.busErr = "no bus"
		m.lastDial = time.Now()
		return m, nil

	case listMsg:
		if msg.err != nil {
			return m, nil
		}
		m.peers = msg.peers
		m.lastList = time.Now()
		for _, p := range msg.peers {
			for _, s := range p.Subs {
				m.tap.Ensure(s)
			}
		}
		return m, m.subMissing()

	case subsStarted:
		var waits []tea.Cmd
		for _, it := range msg.items {
			m.subs[it.name] = it.ch
			m.tap.Ensure(it.name)
			waits = append(waits, waitMsg(it.name, it.ch))
		}
		return m, tea.Batch(waits...)

	case traffic:
		m.tap.Hit(msg.name, msg.body, msg.at)
		wait := waitMsg(msg.name, m.subs[msg.name])
		if msg.name == theme.Channel {
			if c, ok := theme.ParseBus(msg.body); ok {
				return m, tea.Batch(wait, connect.ApplyLook(c, &m.spec, &m.pal, &m.borderless))
			}
		}
		return m, wait

	case subClosed:
		delete(m.subs, msg.name)
		return m, nil

	case tickMsg:
		now := time.Time(msg)
		m.tap.Tick(now)
		m.frame++
		cmds := []tea.Cmd{tickCmd()}
		if m.busOK && now.Sub(m.lastList) >= listEvery {
			cmds = append(cmds, listCmd(m.client))
		}
		if !m.busOK && now.Sub(m.lastDial) >= reconnectEvery {
			cmds = append(cmds, connectCmd())
		}
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if m.focus == "input" {
		switch k {
		case "esc":
			m.focus = "list"
			return m, nil
		case "enter":
			text := strings.TrimSpace(m.input)
			m.input = ""
			sel := m.tap.Selected()
			if text == "" || sel == "" || m.client == nil {
				return m, nil
			}
			c, ch := m.client, sel
			return m, func() tea.Msg {
				_ = c.Pub(ch, text)
				return nil
			}
		case "backspace":
			if m.input != "" {
				r := []rune(m.input)
				m.input = string(r[:len(r)-1])
			}
			return m, nil
		case "ctrl+c":
			return m.quit()
		default:
			if k == "space" {
				m.input += " "
			} else if len(msg.Text) > 0 {
				m.input += msg.Text
			}
			return m, nil
		}
	}

	switch k {
	case "q", "ctrl+c", "esc":
		return m.quit()
	case "j", "down":
		m.tap.SelectDelta(1)
		m.logOff = 0
		m.clampOffset()
	case "k", "up":
		m.tap.SelectDelta(-1)
		m.logOff = 0
		m.clampOffset()
	case "pgdown", "ctrl+d":
		m.logOff -= 5
		if m.logOff < 0 {
			m.logOff = 0
		}
	case "pgup", "ctrl+u":
		m.logOff += 5
	case "i", "enter":
		m.focus = "input"
	case "c":
		m.tap.ClearLog(m.tap.Selected())
	case "r":
		if m.client != nil {
			return m, listCmd(m.client)
		}
	}
	return m, nil
}

func (m model) quit() (tea.Model, tea.Cmd) {
	if m.client != nil {
		_ = m.client.Close()
	}
	return m, tea.Quit
}

func (m *model) clampOffset() {
	n := len(m.tap.Names())
	if m.offset < 0 {
		m.offset = 0
	}
	if m.offset >= n && n > 0 {
		m.offset = n - 1
	}
}

func (m model) subMissing() tea.Cmd {
	if m.client == nil {
		return nil
	}
	var missing []string
	for _, name := range m.tap.Names() {
		if _, ok := m.subs[name]; ok {
			continue
		}
		missing = append(missing, name)
	}
	for _, w := range m.watch {
		if _, ok := m.subs[w]; !ok {
			found := false
			for _, n := range missing {
				if n == w {
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, w)
			}
		}
	}
	if len(missing) == 0 {
		return nil
	}
	c := m.client
	return func() tea.Msg {
		var items []subItem
		for _, name := range missing {
			ch, err := c.Sub(name)
			if err != nil {
				continue
			}
			items = append(items, subItem{name: name, ch: ch})
		}
		return subsStarted{items: items}
	}
}

func waitMsg(name string, ch <-chan string) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		s, ok := <-ch
		if !ok {
			return subClosed{name: name}
		}
		return traffic{name: name, body: s, at: time.Now()}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(tickEvery, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func listCmd(c *ttybus.Client) tea.Cmd {
	if c == nil {
		return nil
	}
	return func() tea.Msg {
		peers, err := c.List()
		return listMsg{peers: peers, err: err}
	}
}

func connectCmd() tea.Cmd {
	return func() tea.Msg {
		c, err := connect.Dial()
		if err != nil {
			return connectErr{err}
		}
		cwd, _ := os.Getwd()
		if err := c.Hello(ttybus.Info{Name: "buscope", CWD: cwd}); err != nil {
			_ = c.Close()
			return connectErr{err}
		}
		return connectedMsg{client: c}
	}
}

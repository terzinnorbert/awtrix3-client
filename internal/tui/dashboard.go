package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/terzi/awtrix3-client/internal/api"
)

// --- messages ---

type dashboardStatsMsg struct {
	stats *api.Stats
	err   error
}

type dashboardLoopMsg struct {
	loop []api.LoopItem
	err  error
}

type dashboardTickMsg struct{}

type dashboardActionMsg struct {
	err error
}

// --- model ---

type dashboardTab struct {
	client      *api.Client
	width       int
	height      int
	stats       *api.Stats
	loop        []api.LoopItem
	loopCursor  int // selected row in the App Loop list
	loading     bool
	err         string
	success     string
	switchInput textinput.Model
	focusSwitch bool
}

func newDashboardTab(client *api.Client) *dashboardTab {
	ti := textinput.New()
	ti.Placeholder = "App name (e.g. Time)"
	ti.CharLimit = 64
	ti.Width = 24
	return &dashboardTab{
		client:      client,
		switchInput: ti,
	}
}

func (t *dashboardTab) Title() string { return "Dashboard" }

func (t *dashboardTab) InputFocused() bool {
	return t.switchInput.Focused()
}

func (t *dashboardTab) SetSize(width, height int) {
	t.width = width
	t.height = height
}

func (t *dashboardTab) Init() tea.Cmd {
	return tea.Batch(
		t.loadStats(),
		t.scheduleTick(),
	)
}

func (t *dashboardTab) loadStats() tea.Cmd {
	return func() tea.Msg {
		stats, err := t.client.GetStats()
		return dashboardStatsMsg{stats: stats, err: err}
	}
}

func (t *dashboardTab) loadLoop() tea.Cmd {
	return func() tea.Msg {
		loop, err := t.client.GetLoop()
		return dashboardLoopMsg{loop: loop, err: err}
	}
}

func (t *dashboardTab) scheduleTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
		return dashboardTickMsg{}
	})
}

func (t dashboardTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			if len(t.loop) > 0 {
				t.loopCursor = (t.loopCursor - 1 + len(t.loop)) % len(t.loop)
			}
		case tea.MouseButtonWheelDown:
			if len(t.loop) > 0 {
				t.loopCursor = (t.loopCursor + 1) % len(t.loop)
			}
		case tea.MouseButtonLeft:
			if msg.Action == tea.MouseActionPress && len(t.loop) > 0 {
				// click on a loop row loads and switches to that app
				name := t.loop[t.loopCursor].Name
				client := t.client
				cmds = append(cmds, func() tea.Msg {
					return dashboardActionMsg{err: client.SwitchApp(name)}
				})
			}
		}
		return &t, tea.Batch(cmds...)

	case dashboardStatsMsg:
		t.loading = false
		if msg.err != nil {
			t.err = msg.err.Error()
		} else {
			t.stats = msg.stats
			t.err = ""
			// Also propagate to root for the header.
			cmds = append(cmds, func() tea.Msg {
				return statsResultMsg{stats: msg.stats}
			})
			cmds = append(cmds, t.loadLoop())
		}

	case dashboardLoopMsg:
		if msg.err == nil {
			t.loop = msg.loop
		}

	case dashboardTickMsg:
		cmds = append(cmds, t.loadStats(), t.scheduleTick())

	case dashboardActionMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = "Done"
			t.err = ""
		}

	case tea.KeyMsg:
		if t.switchInput.Focused() {
			switch msg.String() {
			case "esc":
				t.switchInput.Blur()
				return &t, nil
			case "enter":
				name := strings.TrimSpace(t.switchInput.Value())
				if name != "" {
					client := t.client
					cmds = append(cmds, func() tea.Msg {
						return dashboardActionMsg{err: client.SwitchApp(name)}
					})
				}
				t.switchInput.Blur()
				return &t, tea.Batch(cmds...)
			}
			var cmd tea.Cmd
			t.switchInput, cmd = t.switchInput.Update(msg)
			return &t, cmd
		}

		switch msg.String() {
		case "r":
			t.loading = true
			cmds = append(cmds, t.loadStats())
		case "n":
			client := t.client
			cmds = append(cmds, func() tea.Msg {
				return dashboardActionMsg{err: client.NextApp()}
			})
		case "p":
			client := t.client
			cmds = append(cmds, func() tea.Msg {
				return dashboardActionMsg{err: client.PrevApp()}
			})
		case "s":
			cmd := t.switchInput.Focus()
			cmds = append(cmds, cmd)
		case "j", "down":
			if len(t.loop) > 0 {
				t.loopCursor = (t.loopCursor + 1) % len(t.loop)
			}
		case "k", "up":
			if len(t.loop) > 0 {
				t.loopCursor = (t.loopCursor - 1 + len(t.loop)) % len(t.loop)
			}
		case "enter":
			if len(t.loop) > 0 {
				name := t.loop[t.loopCursor].Name
				client := t.client
				cmds = append(cmds, func() tea.Msg {
					return dashboardActionMsg{err: client.SwitchApp(name)}
				})
			}
		}
	}

	return &t, tea.Batch(cmds...)
}

func (t dashboardTab) View() string {
	var b strings.Builder

	if t.loading && t.stats == nil {
		b.WriteString("\n  Loading…\n")
		return b.String()
	}
	if t.err != "" && t.stats == nil {
		b.WriteString("\n  " + styleError.Render("Connection error: "+t.err) + "\n")
		b.WriteString("  " + styleMuted.Render("Press r to retry") + "\n")
		return b.String()
	}

	leftW := (t.width / 2) - 2
	if leftW < 20 {
		leftW = 20
	}

	// Left: status panel
	var left strings.Builder
	left.WriteString(styleSectionTitle.Render("Status") + "\n")
	if t.stats != nil {
		s := t.stats
		left.WriteString(row("IP", s.IP))
		left.WriteString(row("RAM free", fmt.Sprintf("%d KB", s.RAM/1024)))
		left.WriteString(row("Battery", fmt.Sprintf("%d%%", s.Battery)))
		left.WriteString(row("WiFi RSSI", fmt.Sprintf("%d dBm", s.Wifi)))
		left.WriteString(row("Uptime", formatUptime(s.Uptime)))
		left.WriteString(row("Firmware", "v"+s.Version))
		left.WriteString(row("Current App", s.App))
		left.WriteString(row("Temperature", fmt.Sprintf("%.1f°C", s.Temp)))
		left.WriteString(row("Humidity", fmt.Sprintf("%d%%", s.Humidity)))
		left.WriteString(row("Matrix", boolLabel(s.Matrix)))
	}

	if len(t.loop) > 0 {
		left.WriteString("\n" + styleSectionTitle.Render("App Loop") + "\n")
		for i, item := range t.loop {
			if i >= 8 {
				left.WriteString(styleMuted.Render(fmt.Sprintf("  … and %d more", len(t.loop)-8)) + "\n")
				break
			}
			cursor := "  "
			if i == t.loopCursor {
				cursor = styleAccent("► ")
			}
			left.WriteString(cursor + styleMuted.Render(fmt.Sprintf("%d  ", i)) + styleValue.Render(item.Name) + "\n")
		}
		left.WriteString(styleMuted.Render("j/k=move  Enter=switch") + "\n")
	}

	// Right: quick controls
	var right strings.Builder
	right.WriteString(styleSectionTitle.Render("Quick Controls") + "\n")
	right.WriteString(styleButton.Render("◀ Prev") + "  " + styleKey("p") + "\n\n")
	right.WriteString(styleButton.Render("▶ Next") + "  " + styleKey("n") + "\n\n")
	right.WriteString(styleButton.Render("↺ Refresh") + "  " + styleKey("r") + "\n\n")

	right.WriteString("\n" + styleSectionTitle.Render("Switch to App") + "\n")
	right.WriteString(t.switchInput.View() + "\n")
	right.WriteString(styleMuted.Render("Press s to edit, Enter to switch") + "\n")

	if t.success != "" {
		right.WriteString("\n" + styleSuccess.Render("✓ "+t.success) + "\n")
	}
	if t.err != "" {
		right.WriteString("\n" + styleError.Render("✗ "+t.err) + "\n")
	}

	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(left.String()),
		lipgloss.NewStyle().PaddingLeft(3).Render(right.String()),
	)
	return "\n" + lipgloss.NewStyle().PaddingLeft(2).Render(columns)
}

// --- helpers ---

func row(label, value string) string {
	return styleLabel.Render(label+":") + "  " + styleValue.Render(value) + "\n"
}

func styleKey(k string) string {
	return styleDim.Render("[" + k + "]")
}

func boolLabel(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func formatUptime(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

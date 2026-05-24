// Package tui implements the Bubbletea terminal UI for the AWTRIX 3 client.
package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/terzi/awtrix3-client/internal/api"
)

// tabModel is implemented by every tab in the application.
type tabModel interface {
	tea.Model
	// Title returns the short display name shown in the tab bar.
	Title() string
	// InputFocused reports whether a text-input widget currently has keyboard focus,
	// allowing the root model to suppress tab-switching shortcut keys while typing.
	InputFocused() bool
	// SetSize passes the available content area dimensions to the tab.
	SetSize(width, height int)
}

// tabNames defines the display labels shown in the tab bar.
var tabNames = []string{
	"1 Dashboard",
	"2 Apps",
	"3 Notify",
	"4 Indicators",
	"5 Mood",
	"6 Sound",
	"7 Settings",
	"8 System",
}

// AppModel is the root Bubbletea model that manages the tab bar, chrome, and
// delegates all events to the currently active tab.
type AppModel struct {
	tabs      []tabModel
	activeTab int
	width     int
	height    int
	client    *api.Client
	connected bool
	version   string
	statusMsg string
	statusErr bool
}

// NewAppModel constructs the root model and all tab sub-models.
func NewAppModel(client *api.Client) AppModel {
	tabs := []tabModel{
		newDashboardTab(client),
		newAppsTab(client),
		newNotifyTab(client),
		newIndicatorsTab(client),
		newMoodTab(client),
		newSoundTab(client),
		newSettingsTab(client),
		newSystemTab(client),
	}
	return AppModel{
		tabs:   tabs,
		client: client,
	}
}

// Init starts the application by loading initial dashboard stats.
func (m AppModel) Init() tea.Cmd {
	return m.tabs[0].Init()
}

// statsResultMsg is sent when the initial stats check completes.
type statsResultMsg struct {
	stats *api.Stats
	err   error
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Pass content area size to all tabs.
		contentH := m.height - 6 // header(2) + tabbar(2) + footer(2)
		if contentH < 1 {
			contentH = 1
		}
		for i := range m.tabs {
			m.tabs[i].SetSize(m.width, contentH)
		}
		// Also forward to active tab so it can re-init layout.
		var cmd tea.Cmd
		newModel, cmd := m.tabs[m.activeTab].Update(msg)
		if t, ok := newModel.(tabModel); ok {
			m.tabs[m.activeTab] = t
		}
		return m, cmd

	case statsResultMsg:
		if msg.err == nil && msg.stats != nil {
			m.connected = true
			m.version = msg.stats.Version
		} else {
			m.connected = false
		}

	case tea.KeyMsg:
		// Only intercept navigation keys when no text input is focused.
		if !m.tabs[m.activeTab].InputFocused() {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "1", "2", "3", "4", "5", "6", "7", "8":
				idx := int(msg.String()[0] - '1')
				if idx >= 0 && idx < len(m.tabs) && idx != m.activeTab {
					m.activeTab = idx
					return m, m.tabs[m.activeTab].Init()
				}
				return m, nil
			}
		} else {
			// Always allow ctrl+c to quit.
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}

	case tea.MouseMsg:
		// Scroll wheel is always delegated to the active tab.
		if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
			newModel, cmd := m.tabs[m.activeTab].Update(msg)
			if t, ok := newModel.(tabModel); ok {
				m.tabs[m.activeTab] = t
			}
			return m, cmd
		}
		// Left-click on the tab bar row (row 2) switches tabs.
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft && msg.Y == 2 {
			if idx := m.tabIndexAtX(msg.X); idx >= 0 && idx != m.activeTab {
				m.activeTab = idx
				return m, m.tabs[m.activeTab].Init()
			}
			return m, nil
		}
		// All other clicks are delegated to the active tab.
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			newModel, cmd := m.tabs[m.activeTab].Update(msg)
			if t, ok := newModel.(tabModel); ok {
				m.tabs[m.activeTab] = t
			}
			return m, cmd
		}
	}

	// Delegate the message to the active tab.
	newModel, cmd := m.tabs[m.activeTab].Update(msg)
	if t, ok := newModel.(tabModel); ok {
		m.tabs[m.activeTab] = t
	}
	return m, cmd
}

func (m AppModel) View() string {
	if m.width == 0 {
		return "Loading…"
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.renderTabBar(),
		m.renderContent(),
		m.renderFooter(),
	)
}

func (m AppModel) renderHeader() string {
	title := styleHeader.Render("AWTRIX 3 Client")

	var statusStr string
	if m.connected {
		dot := styleSuccess.Render("●")
		host := styleMuted.Render(m.client.Host())
		ver := styleDim.Render("v" + m.version)
		statusStr = fmt.Sprintf("%s Connected · %s  %s", dot, host, ver)
	} else {
		dot := styleError.Render("●")
		statusStr = dot + styleError.Render(" Disconnected")
	}

	// Right-align the status string.
	gap := m.width - lipgloss.Width(title) - lipgloss.Width(statusStr)
	if gap < 1 {
		gap = 1
	}
	header := title + strings.Repeat(" ", gap) + statusStr
	divider := styleDim.Render(strings.Repeat("─", m.width))
	return header + "\n" + divider
}

func (m AppModel) renderTabBar() string {
	var tabs []string
	for i, name := range tabNames {
		if i == m.activeTab {
			tabs = append(tabs, styleTabActive.Render("["+name+"]"))
		} else {
			tabs = append(tabs, styleTabInactive.Render("["+name+"]"))
		}
	}
	bar := strings.Join(tabs, " ")
	divider := styleDim.Render(strings.Repeat("─", m.width))
	return bar + "\n" + divider
}

func (m AppModel) renderContent() string {
	return m.tabs[m.activeTab].View()
}

func (m AppModel) renderFooter() string {
	hints := "  1-8 Switch tab   Tab/Enter Navigate   Esc Unfocus   q Quit"
	if m.statusMsg != "" {
		var s string
		if m.statusErr {
			s = "  " + styleError.Render("✗ "+m.statusMsg)
		} else {
			s = "  " + styleSuccess.Render("✓ "+m.statusMsg)
		}
		hints = s + strings.Repeat(" ", max(0, m.width-lipgloss.Width(s)-1))
	}
	return styleFooter.Width(m.width).Render(hints)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// tabIndexAtX returns the tab index whose label occupies the given terminal
// column in the tab bar, or -1 if no tab is at that column.
func (m AppModel) tabIndexAtX(x int) int {
	col := 0
	for i, name := range tabNames {
		w := len("[" + name + "]") // ASCII only; ANSI codes don't affect visual width here
		if x >= col && x < col+w {
			return i
		}
		col += w + 1 // +1 for the space separator between tabs
	}
	return -1
}

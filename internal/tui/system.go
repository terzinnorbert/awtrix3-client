package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/terzinnorbert/awtrix3-client/internal/api"
)

const (
	sysBtnPowerOn       = 0
	sysBtnPowerOff      = 1
	sysBtnMatrixOff     = 2
	sysBtnSleep         = 3
	sysBtnReboot        = 4
	sysBtnUpdate        = 5
	sysBtnResetSettings = 6
	sysBtnErase         = 7
	sysFocusTotal       = 8
)

type sysActionMsg struct {
	label string
	err   error
}

// confirm tracks if the user has pressed a "dangerous" button once already.
type systemTab struct {
	client     *api.Client
	width      int
	height     int
	sleepInput textinput.Model
	focus      int
	confirmBtn int // which danger button is awaiting confirmation (-1 = none)
	zones      []clickZone
	err        string
	success    string
}

func newSystemTab(client *api.Client) *systemTab {
	ti := textinput.New()
	ti.Placeholder = "60"
	ti.Width = 6
	ti.CharLimit = 5
	return &systemTab{
		client:     client,
		sleepInput: ti,
		focus:      sysBtnReboot,
		confirmBtn: -1,
	}
}

func (t *systemTab) Title() string { return "System" }

func (t *systemTab) InputFocused() bool {
	return t.sleepInput.Focused()
}

func (t *systemTab) SetSize(w, h int) { t.width = w; t.height = h }

func (t *systemTab) Init() tea.Cmd { return nil }

func (t systemTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch hitZone(t.zones, msg.X, msg.Y) {
			case "reboot":
				return &t, t.rebootCmd()
			case "poweron":
				return &t, t.powerCmd(true)
			case "poweroff":
				return &t, t.powerCmd(false)
			}
		}

	case sysActionMsg:
		t.confirmBtn = -1
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = msg.label + " OK"
			t.err = ""
		}

	case tea.KeyMsg:
		if t.sleepInput.Focused() {
			switch msg.String() {
			case "esc":
				t.sleepInput.Blur()
				return &t, nil
			case "enter":
				t.sleepInput.Blur()
				return &t, t.sleepCmd()
			}
			var cmd tea.Cmd
			t.sleepInput, cmd = t.sleepInput.Update(msg)
			return &t, cmd
		}

		switch msg.String() {
		case "tab", "down":
			t.focus = (t.focus + 1) % sysFocusTotal
		case "shift+tab", "up":
			t.focus = (t.focus - 1 + sysFocusTotal) % sysFocusTotal
		case "s":
			cmd := t.sleepInput.Focus()
			cmds = append(cmds, cmd)
		case "enter", " ":
			return &t, t.handleFocused()
		}
	}

	return &t, tea.Batch(cmds...)
}

func (t *systemTab) handleFocused() tea.Cmd {
	switch t.focus {
	case sysBtnPowerOn:
		return t.powerCmd(true)
	case sysBtnPowerOff:
		return t.powerCmd(false)
	case sysBtnMatrixOff:
		return t.matrixCmd(false)
	case sysBtnSleep:
		return t.sleepCmd()
	case sysBtnReboot:
		return t.rebootCmd()
	case sysBtnUpdate:
		return t.updateCmd()
	case sysBtnResetSettings:
		return t.dangerCmd(sysBtnResetSettings, "Reset Settings", func() error {
			return t.client.ResetSettings()
		})
	case sysBtnErase:
		return t.dangerCmd(sysBtnErase, "Factory Erase", func() error {
			return t.client.Erase()
		})
	}
	return nil
}

func (t *systemTab) powerCmd(on bool) tea.Cmd {
	client := t.client
	label := "Power off"
	if on {
		label = "Power on"
	}
	return func() tea.Msg {
		return sysActionMsg{label: label, err: client.SetPower(on)}
	}
}

func (t *systemTab) matrixCmd(on bool) tea.Cmd {
	client := t.client
	return func() tea.Msg {
		s := api.Settings{MATP: api.BoolPtr(on)}
		return sysActionMsg{label: "Matrix disabled", err: client.SetSettings(s)}
	}
}

func (t *systemTab) sleepCmd() tea.Cmd {
	secs, _ := strconv.Atoi(t.sleepInput.Value())
	if secs <= 0 {
		secs = 60
	}
	client := t.client
	return func() tea.Msg {
		return sysActionMsg{label: "Sleep", err: client.Sleep(secs)}
	}
}

func (t *systemTab) rebootCmd() tea.Cmd {
	client := t.client
	return func() tea.Msg {
		return sysActionMsg{label: "Reboot", err: client.Reboot()}
	}
}

func (t *systemTab) updateCmd() tea.Cmd {
	client := t.client
	return func() tea.Msg {
		return sysActionMsg{label: "OTA update triggered", err: client.DoUpdate()}
	}
}

// dangerCmd requires a second confirmation press before executing.
func (t *systemTab) dangerCmd(btn int, label string, fn func() error) tea.Cmd {
	if t.confirmBtn != btn {
		// First press: arm confirmation.
		t.confirmBtn = btn
		t.success = ""
		t.err = "Press Enter again to confirm: " + label
		return nil
	}
	// Second press: execute.
	t.confirmBtn = -1
	return func() tea.Msg {
		return sysActionMsg{label: label, err: fn()}
	}
}

func (t *systemTab) View() string {
	var b strings.Builder
	b.WriteString("\n")

	btn := func(label string, idx int, danger bool) string {
		focused := t.focus == idx
		confirming := t.confirmBtn == idx
		var s lipgloss.Style
		switch {
		case danger && confirming:
			s = styleButtonDangerFocused
		case danger:
			s = styleButtonDanger
		case focused:
			s = styleButtonFocused
		default:
			s = styleButton
		}
		if confirming {
			label = "⚠ " + label + " (confirm)"
		}
		return s.Render("[ " + label + " ]")
	}

	// Power section
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("Power") + "\n" +
			btn("⏻ Power On", sysBtnPowerOn, false) + "  " +
			btn("⏻ Power Off", sysBtnPowerOff, false) + "  " +
			btn("Matrix Off", sysBtnMatrixOff, false) + "\n\n",
	))

	// Sleep section
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("Deep Sleep") + "\n" +
			"  Duration: " + t.sleepInput.View() + "s   " +
			btn("Enter Sleep", sysBtnSleep, false) + "\n" +
			styleMuted.Render("  (s to edit duration; wakes after timer or middle button press)\n\n"),
	))

	// Firmware section
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("Firmware") + "\n" +
			btn("↺ Reboot", sysBtnReboot, false) + "  " +
			btn("⟳ OTA Update", sysBtnUpdate, false) + "\n\n",
	))

	// Danger zone
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("⚠ Danger Zone") + "\n" +
			btn("Reset Settings", sysBtnResetSettings, true) +
			styleMuted.Render("  Resets API settings (keeps WiFi & flash)\n\n") +
			"  " + btn("Factory Erase", sysBtnErase, true) +
			styleMuted.Render("  Formats flash + EEPROM (keeps WiFi)\n"),
	))

	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleMuted.Render("Tab=focus  Enter/Space=activate  s=edit sleep"),
	))
	b.WriteString("\n")

	if t.success != "" {
		b.WriteString("\n" + lipgloss.NewStyle().PaddingLeft(2).Render(styleSuccess.Render("✓ "+t.success)))
		b.WriteString("\n")
	}
	if t.err != "" {
		var style lipgloss.Style
		if t.confirmBtn >= 0 {
			style = styleWarning
		} else {
			style = styleError
		}
		b.WriteString("\n" + lipgloss.NewStyle().PaddingLeft(2).Render(style.Render(t.err)))
		b.WriteString("\n")
	}

	return b.String()
}

// Ensure strings package is used.
var _ = strings.Builder{}

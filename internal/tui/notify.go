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
	notifyFieldText     = 0
	notifyFieldColor    = 1
	notifyFieldIcon     = 2
	notifyFieldDuration = 3
	notifyFieldSound    = 4
	notifyFieldRTTTL    = 5
	notifyFieldClients  = 6
	notifyNumInputs     = 7
)

const (
	notifyTogHold      = 0
	notifyTogWakeup    = 1
	notifyTogStack     = 2
	notifyTogLoopSound = 3
	notifyNumToggles   = 4
)

type notifyActionMsg struct{ err error }
type notifyDismissMsg struct{ err error }

type notifyTab struct {
	client  *api.Client
	width   int
	height  int
	inputs  []textinput.Model
	focus   int // 0..numInputs-1 = inputs; numInputs = send btn; +1 = dismiss btn
	toggles [notifyNumToggles]bool
	zones   []clickZone
	err     string
	success string
}

func newNotifyTab(client *api.Client) *notifyTab {
	specs := []struct {
		placeholder string
		width       int
		limit       int
	}{
		{"Alert! Motion detected", 40, 512},
		{"#FF0000", 10, 16},
		{"1001", 8, 32},
		{"10", 5, 5},
		{"alarm", 20, 64},
		{"Mario:d=4,o=5,b=125:e,e,e", 40, 512},
		{"192.168.1.101", 40, 256},
	}
	inputs := make([]textinput.Model, len(specs))
	for i, s := range specs {
		ti := textinput.New()
		ti.Placeholder = s.placeholder
		ti.Width = s.width
		ti.CharLimit = s.limit
		inputs[i] = ti
	}
	inputs[notifyFieldText].Focus()

	return &notifyTab{
		client:  client,
		inputs:  inputs,
		focus:   notifyFieldText,
		toggles: [notifyNumToggles]bool{false, true, true, false},
	}
}

func (t *notifyTab) Title() string { return "Notify" }

func (t *notifyTab) InputFocused() bool {
	for _, in := range t.inputs {
		if in.Focused() {
			return true
		}
	}
	return false
}

func (t *notifyTab) SetSize(w, h int) { t.width = w; t.height = h }

func (t *notifyTab) Init() tea.Cmd { return nil }

func (t notifyTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch hitZone(t.zones, msg.X, msg.Y) {
			case "send":
				return &t, t.sendCmd()
			case "dismiss":
				return &t, t.dismissCmd()
			}
		}

	case notifyActionMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = "Notification sent"
			t.err = ""
		}
	case notifyDismissMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = "Notification dismissed"
			t.err = ""
		}

	case tea.KeyMsg:
		if t.InputFocused() {
			switch msg.String() {
			case "esc":
				t.blurAll()
				t.focus = -1
				return &t, nil
			case "tab", "down":
				t.blurAll()
				total := notifyNumInputs + 2
				t.focus = (t.focus + 1) % total
				if t.focus < notifyNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "shift+tab", "up":
				t.blurAll()
				total := notifyNumInputs + 2
				t.focus = (t.focus - 1 + total) % total
				if t.focus < notifyNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "enter":
				if t.focus == notifyNumInputs {
					return &t, t.sendCmd()
				}
				if t.focus == notifyNumInputs+1 {
					return &t, t.dismissCmd()
				}
			}
			for i := range t.inputs {
				if t.inputs[i].Focused() {
					var cmd tea.Cmd
					t.inputs[i], cmd = t.inputs[i].Update(msg)
					cmds = append(cmds, cmd)
					return &t, tea.Batch(cmds...)
				}
			}
		}

		switch msg.String() {
		case "tab":
			t.blurAll()
			t.focus = 0
			cmds = append(cmds, t.inputs[0].Focus())
		case "enter":
			if t.focus == notifyNumInputs {
				return &t, t.sendCmd()
			}
			if t.focus == notifyNumInputs+1 {
				return &t, t.dismissCmd()
			}
		case "1":
			t.toggles[notifyTogHold] = !t.toggles[notifyTogHold]
		case "2":
			t.toggles[notifyTogWakeup] = !t.toggles[notifyTogWakeup]
		case "3":
			t.toggles[notifyTogStack] = !t.toggles[notifyTogStack]
		case "4":
			t.toggles[notifyTogLoopSound] = !t.toggles[notifyTogLoopSound]
		case "s":
			return &t, t.sendCmd()
		case "d":
			return &t, t.dismissCmd()
		}
	}

	return &t, tea.Batch(cmds...)
}

func (t *notifyTab) blurAll() {
	for i := range t.inputs {
		t.inputs[i].Blur()
	}
}

func (t *notifyTab) buildNotification() api.CustomApp {
	n := api.CustomApp{}
	if v := t.inputs[notifyFieldText].Value(); v != "" {
		n.Text = v
	}
	if v := t.inputs[notifyFieldColor].Value(); v != "" {
		n.Color = v
	}
	if v := t.inputs[notifyFieldIcon].Value(); v != "" {
		n.Icon = v
	}
	if v := t.inputs[notifyFieldSound].Value(); v != "" {
		n.Sound = v
	}
	if v := t.inputs[notifyFieldRTTTL].Value(); v != "" {
		n.Rtttl = v
	}
	if v, err := strconv.Atoi(t.inputs[notifyFieldDuration].Value()); err == nil && v > 0 {
		n.Duration = api.IntPtr(v)
	}
	if raw := strings.TrimSpace(t.inputs[notifyFieldClients].Value()); raw != "" {
		for _, c := range strings.Split(raw, ",") {
			if addr := strings.TrimSpace(c); addr != "" {
				n.Clients = append(n.Clients, addr)
			}
		}
	}
	if t.toggles[notifyTogHold] {
		n.Hold = api.BoolPtr(true)
	}
	if t.toggles[notifyTogWakeup] {
		n.Wakeup = api.BoolPtr(true)
	}
	if t.toggles[notifyTogStack] {
		n.Stack = api.BoolPtr(true)
	}
	if t.toggles[notifyTogLoopSound] {
		n.LoopSound = api.BoolPtr(true)
	}
	return n
}

func (t *notifyTab) sendCmd() tea.Cmd {
	if strings.TrimSpace(t.inputs[notifyFieldText].Value()) == "" {
		t.err = "text is required"
		t.success = ""
		return nil
	}
	if t.inputs[notifyFieldSound].Value() != "" && t.inputs[notifyFieldRTTTL].Value() != "" {
		t.err = "sound and RTTTL are mutually exclusive"
		t.success = ""
		return nil
	}
	n := t.buildNotification()
	client := t.client
	return func() tea.Msg {
		return notifyActionMsg{err: client.SendNotification(n)}
	}
}

func (t *notifyTab) dismissCmd() tea.Cmd {
	client := t.client
	return func() tea.Msg {
		return notifyDismissMsg{err: client.DismissNotification()}
	}
}

func (t *notifyTab) View() string {
	var b strings.Builder
	b.WriteString("\n")

	labels := []string{"Text", "Color", "Icon", "Duration(s)", "Sound", "RTTTL", "Forward to (IPs)"}
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(styleSectionTitle.Render("Send Notification") + "\n"))
	for i, lbl := range labels {
		focused := t.focus == i
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(formRow(lbl, t.inputs[i].View(), focused)))
	}
	b.WriteString("\n")

	// Toggles
	togLabels := []string{"Hold", "Wakeup if off", "Stack", "Loop sound"}
	parts := make([]string, notifyNumToggles)
	for i, lbl := range togLabels {
		parts[i] = checkbox(t.toggles[i], lbl)
	}
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		strings.Join(parts, "   ") + "\n" +
			styleMuted.Render("(1-4 to toggle flags when not typing)") + "\n",
	))
	b.WriteString("\n")

	sendBtn := styleButton.Render("[ Send Notification ]")
	dismissBtn := styleButtonDanger.Render("[ Dismiss Current ]")
	if t.focus == notifyNumInputs {
		sendBtn = styleButtonFocused.Render("[ Send Notification ]")
	}
	if t.focus == notifyNumInputs+1 {
		dismissBtn = styleButtonDangerFocused.Render("[ Dismiss Current ]")
	}

	const outerPad = 2
	btnY := zoneLine(b.String())
	sendW := lipgloss.Width(sendBtn)
	dismissW := lipgloss.Width(dismissBtn)
	t.zones = []clickZone{
		{YMin: btnY, YMax: btnY, XMin: outerPad, XMax: outerPad + sendW - 1, ID: "send"},
		{YMin: btnY, YMax: btnY, XMin: outerPad + sendW + 2, XMax: outerPad + sendW + 2 + dismissW - 1, ID: "dismiss"},
	}

	b.WriteString(lipgloss.NewStyle().PaddingLeft(outerPad).Render(
		sendBtn + "  " + dismissBtn + "   " +
			styleMuted.Render("s=send  d=dismiss  Tab=next field"),
	))
	b.WriteString("\n")

	if t.success != "" {
		b.WriteString("\n" + lipgloss.NewStyle().PaddingLeft(2).Render(styleSuccess.Render("✓ "+t.success)))
		b.WriteString("\n")
	}
	if t.err != "" {
		b.WriteString("\n" + lipgloss.NewStyle().PaddingLeft(2).Render(styleError.Render("✗ "+t.err)))
		b.WriteString("\n")
	}

	return b.String()
}

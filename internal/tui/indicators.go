package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/terzi/awtrix3-client/internal/api"
)

// Each indicator has 3 inputs: color, blink, fade.
// Total = 3 indicators × 3 fields = 9 inputs.
// Focus indices beyond 9 map to buttons: set1, clear1, set2, clear2, set3, clear3, clearAll.
const (
	indInputsPerInd = 3 // color, blink, fade
	indNumInds      = 3
	indTotalInputs  = indNumInds * indInputsPerInd // 9
	// Button focus indices start at indTotalInputs.
	indBtnSet1     = indTotalInputs
	indBtnClear1   = indTotalInputs + 1
	indBtnSet2     = indTotalInputs + 2
	indBtnClear2   = indTotalInputs + 3
	indBtnSet3     = indTotalInputs + 4
	indBtnClear3   = indTotalInputs + 5
	indBtnClearAll = indTotalInputs + 6
	indFocusTotal  = indTotalInputs + 7
)

type indActionMsg struct {
	label string
	err   error
}

type indicatorsTab struct {
	client  *api.Client
	width   int
	height  int
	inputs  []textinput.Model // 9 inputs: [ind0_color, ind0_blink, ind0_fade, ind1_color, ...]
	focus   int
	zones   []clickZone
	err     string
	success string
}

func newIndicatorsTab(client *api.Client) *indicatorsTab {
	inputs := make([]textinput.Model, indTotalInputs)
	for i := 0; i < indNumInds; i++ {
		base := i * indInputsPerInd

		color := textinput.New()
		color.Placeholder = "#FF0000"
		color.Width = 10
		color.CharLimit = 16
		inputs[base+0] = color

		blink := textinput.New()
		blink.Placeholder = "ms (e.g. 500)"
		blink.Width = 12
		blink.CharLimit = 8
		inputs[base+1] = blink

		fade := textinput.New()
		fade.Placeholder = "ms (e.g. 300)"
		fade.Width = 12
		fade.CharLimit = 8
		inputs[base+2] = fade
	}
	inputs[0].Focus() // start with ind1 color focused

	return &indicatorsTab{
		client: client,
		inputs: inputs,
		focus:  0,
	}
}

func (t *indicatorsTab) Title() string { return "Indicators" }

func (t *indicatorsTab) InputFocused() bool {
	for _, in := range t.inputs {
		if in.Focused() {
			return true
		}
	}
	return false
}

func (t *indicatorsTab) SetSize(w, h int) { t.width = w; t.height = h }

func (t *indicatorsTab) Init() tea.Cmd { return nil }

func (t indicatorsTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch hitZone(t.zones, msg.X, msg.Y) {
			case "clearall":
				return &t, t.clearAllCmd()
			}
		}

	case indActionMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = msg.label + " OK"
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
				t.focus = (t.focus + 1) % indFocusTotal
				if t.focus < indTotalInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "shift+tab", "up":
				t.blurAll()
				t.focus = (t.focus - 1 + indFocusTotal) % indFocusTotal
				if t.focus < indTotalInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "enter":
				return &t, t.handleButton()
			}
			// Forward to focused input.
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
			return &t, t.handleButton()
		case "a":
			return &t, t.clearAllCmd()
		}
	}

	return &t, tea.Batch(cmds...)
}

func (t *indicatorsTab) blurAll() {
	for i := range t.inputs {
		t.inputs[i].Blur()
	}
}

func (t *indicatorsTab) handleButton() tea.Cmd {
	switch t.focus {
	case indBtnSet1:
		return t.setCmd(1)
	case indBtnClear1:
		return t.clearCmd(1)
	case indBtnSet2:
		return t.setCmd(2)
	case indBtnClear2:
		return t.clearCmd(2)
	case indBtnSet3:
		return t.setCmd(3)
	case indBtnClear3:
		return t.clearCmd(3)
	case indBtnClearAll:
		return t.clearAllCmd()
	}
	return nil
}

func (t *indicatorsTab) buildIndicator(num int) api.Indicator {
	base := (num - 1) * indInputsPerInd
	ind := api.Indicator{}
	if v := t.inputs[base+0].Value(); v != "" {
		ind.Color = v
	}
	if v, err := strconv.Atoi(t.inputs[base+1].Value()); err == nil && v > 0 {
		ind.Blink = api.IntPtr(v)
	}
	if v, err := strconv.Atoi(t.inputs[base+2].Value()); err == nil && v > 0 {
		ind.Fade = api.IntPtr(v)
	}
	return ind
}

func (t *indicatorsTab) setCmd(num int) tea.Cmd {
	ind := t.buildIndicator(num)
	client := t.client
	label := fmt.Sprintf("Indicator %d set", num)
	return func() tea.Msg {
		return indActionMsg{label: label, err: client.SetIndicator(num, ind)}
	}
}

func (t *indicatorsTab) clearCmd(num int) tea.Cmd {
	client := t.client
	label := fmt.Sprintf("Indicator %d cleared", num)
	return func() tea.Msg {
		return indActionMsg{label: label, err: client.ClearIndicator(num)}
	}
}

func (t *indicatorsTab) clearAllCmd() tea.Cmd {
	client := t.client
	return func() tea.Msg {
		return indActionMsg{label: "All indicators cleared", err: client.ClearAllIndicators()}
	}
}

func (t *indicatorsTab) View() string {
	var b strings.Builder
	b.WriteString("\n")

	positions := []string{
		"Indicator 1  (Upper Right)",
		"Indicator 2  (Right Side)",
		"Indicator 3  (Lower Right)",
	}
	setBtnFocus := []int{indBtnSet1, indBtnSet2, indBtnSet3}
	clearBtnFocus := []int{indBtnClear1, indBtnClear2, indBtnClear3}

	for i := 0; i < indNumInds; i++ {
		base := i * indInputsPerInd

		colorFocused := t.focus == base+0
		blinkFocused := t.focus == base+1
		fadeFocused := t.focus == base+2

		body := strings.Builder{}
		body.WriteString(formRow("Color", t.inputs[base+0].View(), colorFocused))
		body.WriteString(formRow("Blink (ms)", t.inputs[base+1].View(), blinkFocused))
		body.WriteString(formRow("Fade (ms)", t.inputs[base+2].View(), fadeFocused))

		setBtn := styleButton.Render("[ Set ]")
		clearBtn := styleButtonDanger.Render("[ Clear ]")
		if t.focus == setBtnFocus[i] {
			setBtn = styleButtonFocused.Render("[ Set ]")
		}
		if t.focus == clearBtnFocus[i] {
			clearBtn = styleButtonDangerFocused.Render("[ Clear ]")
		}
		body.WriteString("  " + setBtn + "  " + clearBtn + "\n")

		box := styleBox.Width(50).Render(body.String())
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			styleSectionTitle.Render(positions[i]) + "\n" + box,
		))
		b.WriteString("\n\n")
	}

	clearAllBtn := styleButtonDanger.Render("[ Clear All Indicators ]")
	if t.focus == indBtnClearAll {
		clearAllBtn = styleButtonDangerFocused.Render("[ Clear All Indicators ]")
	}

	const outerPad = 2
	caY := zoneLine(b.String())
	caW := lipgloss.Width(clearAllBtn)
	t.zones = []clickZone{
		{YMin: caY, YMax: caY, XMin: outerPad, XMax: outerPad + caW - 1, ID: "clearall"},
	}

	b.WriteString(lipgloss.NewStyle().PaddingLeft(outerPad).Render(
		clearAllBtn + "   " + styleMuted.Render("a=clear all  Tab=next field"),
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

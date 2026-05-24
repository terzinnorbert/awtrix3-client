package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/terzi/awtrix3-client/internal/api"
)

const (
	moodFieldColor      = 0
	moodFieldBrightness = 1
	moodFieldKelvin     = 2
	moodNumInputs       = 3
	moodBtnApply        = moodNumInputs
	moodBtnDisable      = moodNumInputs + 1
	moodFocusTotal      = moodNumInputs + 2
)

type moodActionMsg struct{ err error }

type moodTab struct {
	client    *api.Client
	width     int
	height    int
	inputs    []textinput.Model
	focus     int
	useKelvin bool // false=RGB color, true=kelvin
	zones     []clickZone
	err       string
	success   string
}

func newMoodTab(client *api.Client) *moodTab {
	specs := []struct {
		placeholder string
		width       int
		limit       int
	}{
		{"#FF00FF", 12, 16},
		{"170", 5, 5},
		{"2300", 6, 6},
	}
	inputs := make([]textinput.Model, moodNumInputs)
	for i, s := range specs {
		ti := textinput.New()
		ti.Placeholder = s.placeholder
		ti.Width = s.width
		ti.CharLimit = s.limit
		inputs[i] = ti
	}
	inputs[moodFieldColor].Focus()

	return &moodTab{
		client: client,
		inputs: inputs,
		focus:  moodFieldColor,
	}
}

func (t *moodTab) Title() string { return "Mood" }

func (t *moodTab) InputFocused() bool {
	for _, in := range t.inputs {
		if in.Focused() {
			return true
		}
	}
	return false
}

func (t *moodTab) SetSize(w, h int) { t.width = w; t.height = h }

func (t *moodTab) Init() tea.Cmd { return nil }

func (t moodTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch hitZone(t.zones, msg.X, msg.Y) {
			case "apply":
				return &t, t.applyCmd()
			case "disable":
				return &t, t.disableCmd()
			}
		}

	case moodActionMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = "Applied"
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
				t.focus = (t.focus + 1) % moodFocusTotal
				if t.focus < moodNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "shift+tab", "up":
				t.blurAll()
				t.focus = (t.focus - 1 + moodFocusTotal) % moodFocusTotal
				if t.focus < moodNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "enter":
				if t.focus == moodBtnApply {
					return &t, t.applyCmd()
				}
				if t.focus == moodBtnDisable {
					return &t, t.disableCmd()
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
		case "m":
			t.useKelvin = !t.useKelvin
		case "a":
			return &t, t.applyCmd()
		case "d":
			return &t, t.disableCmd()
		}
	}

	return &t, tea.Batch(cmds...)
}

func (t *moodTab) blurAll() {
	for i := range t.inputs {
		t.inputs[i].Blur()
	}
}

func (t *moodTab) buildMoodLight() api.MoodLight {
	ml := api.MoodLight{}
	if v, err := strconv.Atoi(t.inputs[moodFieldBrightness].Value()); err == nil {
		ml.Brightness = v
	} else {
		ml.Brightness = 128
	}
	if t.useKelvin {
		if v, err := strconv.Atoi(t.inputs[moodFieldKelvin].Value()); err == nil {
			ml.Kelvin = api.IntPtr(v)
		}
	} else {
		if v := t.inputs[moodFieldColor].Value(); v != "" {
			ml.Color = v
		}
	}
	return ml
}

func (t *moodTab) applyCmd() tea.Cmd {
	ml := t.buildMoodLight()
	client := t.client
	return func() tea.Msg {
		return moodActionMsg{err: client.SetMoodLight(ml)}
	}
}

func (t *moodTab) disableCmd() tea.Cmd {
	client := t.client
	return func() tea.Msg {
		return moodActionMsg{err: client.DisableMoodLight()}
	}
}

func (t *moodTab) View() string {
	var b strings.Builder
	b.WriteString("\n")

	// Mode selector
	colorMode := radioButton(!t.useKelvin, "RGB Color")
	kelvinMode := radioButton(t.useKelvin, "Color Temperature (Kelvin)")
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("Mood Lighting") + "\n" +
			styleMuted.Render("Mode (m to toggle): ") + colorMode + "   " + kelvinMode + "\n\n",
	))

	// Fields
	bri := t.inputs[moodFieldBrightness].Value()
	briVal, _ := strconv.Atoi(bri)
	briBar := progressBar(briVal, 255, 20)

	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		formRow("Brightness", t.inputs[moodFieldBrightness].View()+" "+
			styleMuted.Render(briBar+" "+bri+"/255"), t.focus == moodFieldBrightness),
	))

	if !t.useKelvin {
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			formRow("Color (hex)", t.inputs[moodFieldColor].View(), t.focus == moodFieldColor),
		))
		// Color preview strip
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			styleLabel.Render("Preview:") + "  " + colorPreviewStrip(t.inputs[moodFieldColor].Value()) + "\n",
		))
	} else {
		kelvinVal, _ := strconv.Atoi(t.inputs[moodFieldKelvin].Value())
		kelvinBar := progressBar(kelvinVal-1000, 9000, 20)
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			formRow("Kelvin (1000-10000)", t.inputs[moodFieldKelvin].View()+" "+
				styleMuted.Render(kelvinBar+" warm→cool"), t.focus == moodFieldKelvin),
		))
	}
	b.WriteString("\n")

	applyBtn := styleButton.Render("[ Apply Moodlight ]")
	disableBtn := styleButtonDanger.Render("[ Disable ]")
	if t.focus == moodBtnApply {
		applyBtn = styleButtonFocused.Render("[ Apply Moodlight ]")
	}
	if t.focus == moodBtnDisable {
		disableBtn = styleButtonDangerFocused.Render("[ Disable ]")
	}

	const outerPad = 2
	btnY := zoneLine(b.String())
	applyW := lipgloss.Width(applyBtn)
	disableW := lipgloss.Width(disableBtn)
	t.zones = []clickZone{
		{YMin: btnY, YMax: btnY, XMin: outerPad, XMax: outerPad + applyW - 1, ID: "apply"},
		{YMin: btnY, YMax: btnY, XMin: outerPad + applyW + 2, XMax: outerPad + applyW + 2 + disableW - 1, ID: "disable"},
	}

	b.WriteString(lipgloss.NewStyle().PaddingLeft(outerPad).Render(
		applyBtn + "  " + disableBtn + "   " +
			styleMuted.Render("a=apply  d=disable  m=toggle mode  Tab=next field"),
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

// colorPreviewStrip renders a colored block representing a hex color value.
// When the terminal supports true color the background will match the hex.
func colorPreviewStrip(hex string) string {
	hex = strings.TrimSpace(hex)
	if len(hex) == 0 || hex[0] != '#' {
		return styleDim.Render("(enter a hex color above)")
	}
	strip := strings.Repeat("█", 20)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex)).Render(strip)
}

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
	soundFieldName   = 0
	soundFieldRTTTL  = 1
	soundFieldVolume = 2
	soundNumInputs   = 3
	soundBtnPlay     = soundNumInputs
	soundFocusTotal  = soundNumInputs + 1
)

type soundActionMsg struct{ err error }

type soundTab struct {
	client   *api.Client
	width    int
	height   int
	inputs   []textinput.Model
	focus    int
	useRTTTL bool // false=melody file name, true=RTTTL string
	zones    []clickZone
	err      string
	success  string
}

func newSoundTab(client *api.Client) *soundTab {
	specs := []struct {
		placeholder string
		width       int
		limit       int
	}{
		{"alarm", 24, 64},
		{"d=4,o=5,b=125:e,e,e,c,e,g,g", 50, 512},
		{"15", 4, 4},
	}
	inputs := make([]textinput.Model, soundNumInputs)
	for i, s := range specs {
		ti := textinput.New()
		ti.Placeholder = s.placeholder
		ti.Width = s.width
		ti.CharLimit = s.limit
		inputs[i] = ti
	}
	inputs[soundFieldName].Focus()

	return &soundTab{
		client: client,
		inputs: inputs,
		focus:  soundFieldName,
	}
}

func (t *soundTab) Title() string { return "Sound" }

func (t *soundTab) InputFocused() bool {
	for _, in := range t.inputs {
		if in.Focused() {
			return true
		}
	}
	return false
}

func (t *soundTab) SetSize(w, h int) { t.width = w; t.height = h }

func (t *soundTab) Init() tea.Cmd { return nil }

func (t soundTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if hitZone(t.zones, msg.X, msg.Y) == "play" {
				return &t, t.playCmd()
			}
		}

	case soundActionMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = "Playing"
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
				t.focus = (t.focus + 1) % soundFocusTotal
				if t.focus < soundNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "shift+tab", "up":
				t.blurAll()
				t.focus = (t.focus - 1 + soundFocusTotal) % soundFocusTotal
				if t.focus < soundNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "enter":
				if t.focus == soundBtnPlay {
					return &t, t.playCmd()
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
			t.useRTTTL = !t.useRTTTL
			t.blurAll()
			if t.useRTTTL {
				t.focus = soundFieldRTTTL
			} else {
				t.focus = soundFieldName
			}
			cmds = append(cmds, t.inputs[t.focus].Focus())
		case "p":
			return &t, t.playCmd()
		}
	}

	return &t, tea.Batch(cmds...)
}

func (t *soundTab) blurAll() {
	for i := range t.inputs {
		t.inputs[i].Blur()
	}
}

func (t *soundTab) playCmd() tea.Cmd {
	client := t.client
	useRTTTL := t.useRTTTL
	name := t.inputs[soundFieldName].Value()
	rtttl := t.inputs[soundFieldRTTTL].Value()
	volStr := t.inputs[soundFieldVolume].Value()
	vol, _ := strconv.Atoi(volStr)

	return func() tea.Msg {
		// Apply volume via settings if specified.
		if vol > 0 {
			_ = client.SetSettings(api.Settings{VOL: api.IntPtr(vol)})
		}
		if useRTTTL {
			if rtttl == "" {
				return soundActionMsg{err: nil}
			}
			return soundActionMsg{err: client.PlayRTTTL(rtttl)}
		}
		if name == "" {
			return soundActionMsg{err: nil}
		}
		return soundActionMsg{err: client.PlaySound(name)}
	}
}

func (t *soundTab) View() string {
	var b strings.Builder
	b.WriteString("\n")

	nameMode := radioButton(!t.useRTTTL, "Melody file name")
	rtttlMode := radioButton(t.useRTTTL, "RTTTL string")
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("Sound Playback") + "\n" +
			styleMuted.Render("Mode (m to toggle): ") + nameMode + "   " + rtttlMode + "\n\n",
	))

	if !t.useRTTTL {
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			formRow("Melody name", t.inputs[soundFieldName].View(), t.focus == soundFieldName),
		))
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			styleMuted.Render("  (filename without extension from the MELODIES folder)\n"),
		))
	} else {
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			formRow("RTTTL string", t.inputs[soundFieldRTTTL].View(), t.focus == soundFieldRTTTL),
		))
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			styleMuted.Render("  e.g. d=4,o=5,b=125:e,e,e,c,e,g,g\n"),
		))
	}

	volVal, _ := strconv.Atoi(t.inputs[soundFieldVolume].Value())
	volBar := progressBar(volVal, 30, 15)
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		formRow("Volume (0-30)", t.inputs[soundFieldVolume].View()+" "+
			styleMuted.Render(volBar+" "+t.inputs[soundFieldVolume].Value()+"/30"),
			t.focus == soundFieldVolume),
	))
	b.WriteString("\n")

	playBtn := styleButton.Render("[ ▶ Play ]")
	if t.focus == soundBtnPlay {
		playBtn = styleButtonFocused.Render("[ ▶ Play ]")
	}

	const outerPad = 2
	btnY := zoneLine(b.String())
	playW := lipgloss.Width(playBtn)
	t.zones = []clickZone{
		{YMin: btnY, YMax: btnY, XMin: outerPad, XMax: outerPad + playW - 1, ID: "play"},
	}

	b.WriteString(lipgloss.NewStyle().PaddingLeft(outerPad).Render(
		playBtn + "   " + styleMuted.Render("p=play  m=toggle mode  Tab=next field"),
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

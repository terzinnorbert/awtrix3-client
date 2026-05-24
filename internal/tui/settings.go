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
	setFieldBRI     = 0
	setFieldATIME   = 1
	setFieldTSPEED  = 2
	setFieldSSPEED  = 3
	setFieldTCOL    = 4
	setFieldTFORMAT = 5
	setFieldDFORMAT = 6
	setNumInputs    = 7
	setBtnApply     = setNumInputs
	setBtnReload    = setNumInputs + 1
	setFocusTotal   = setNumInputs + 2
)

const (
	setTogABRI    = 0
	setTogATRANS  = 1
	setTogSOM     = 2
	setTogCEL     = 3
	setTogUPPER   = 4
	setNumToggles = 5
)

type settingsLoadMsg struct {
	settings *api.Settings
	err      error
}

type settingsApplyMsg struct{ err error }

type settingsTab struct {
	client  *api.Client
	width   int
	height  int
	inputs  []textinput.Model
	focus   int
	toggles [setNumToggles]bool
	zones   []clickZone
	loaded  bool
	err     string
	success string
}

func newSettingsTab(client *api.Client) *settingsTab {
	specs := []struct {
		placeholder string
		width       int
		limit       int
	}{
		{"128", 5, 5},
		{"7", 4, 4},
		{"500", 6, 6},
		{"100", 5, 5},
		{"#FFFFFF", 10, 16},
		{"%H:%M", 16, 32},
		{"%d.%m.%y", 16, 32},
	}
	inputs := make([]textinput.Model, setNumInputs)
	for i, s := range specs {
		ti := textinput.New()
		ti.Placeholder = s.placeholder
		ti.Width = s.width
		ti.CharLimit = s.limit
		inputs[i] = ti
	}
	inputs[setFieldBRI].Focus()

	return &settingsTab{
		client:  client,
		inputs:  inputs,
		focus:   setFieldBRI,
		toggles: [setNumToggles]bool{false, true, true, true, true},
	}
}

func (t *settingsTab) Title() string { return "Settings" }

func (t *settingsTab) InputFocused() bool {
	for _, in := range t.inputs {
		if in.Focused() {
			return true
		}
	}
	return false
}

func (t *settingsTab) SetSize(w, h int) { t.width = w; t.height = h }

func (t *settingsTab) Init() tea.Cmd {
	return t.loadCmd()
}

func (t *settingsTab) loadCmd() tea.Cmd {
	client := t.client
	return func() tea.Msg {
		s, err := client.GetSettings()
		return settingsLoadMsg{settings: s, err: err}
	}
}

func (t settingsTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			switch hitZone(t.zones, msg.X, msg.Y) {
			case "apply":
				return &t, t.applyCmd()
			case "reload":
				return &t, t.loadCmd()
			}
		}

	case settingsLoadMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
		} else if msg.settings != nil {
			t.loaded = true
			t.populateFromSettings(msg.settings)
			t.err = ""
		}

	case settingsApplyMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = "Settings applied"
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
				t.focus = (t.focus + 1) % setFocusTotal
				if t.focus < setNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "shift+tab", "up":
				t.blurAll()
				t.focus = (t.focus - 1 + setFocusTotal) % setFocusTotal
				if t.focus < setNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "enter":
				if t.focus == setBtnApply {
					return &t, t.applyCmd()
				}
				if t.focus == setBtnReload {
					return &t, t.loadCmd()
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
		case "1":
			t.toggles[setTogABRI] = !t.toggles[setTogABRI]
		case "2":
			t.toggles[setTogATRANS] = !t.toggles[setTogATRANS]
		case "3":
			t.toggles[setTogSOM] = !t.toggles[setTogSOM]
		case "4":
			t.toggles[setTogCEL] = !t.toggles[setTogCEL]
		case "5":
			t.toggles[setTogUPPER] = !t.toggles[setTogUPPER]
		case "a":
			return &t, t.applyCmd()
		case "r":
			return &t, t.loadCmd()
		}
	}

	return &t, tea.Batch(cmds...)
}

func (t *settingsTab) blurAll() {
	for i := range t.inputs {
		t.inputs[i].Blur()
	}
}

func (t *settingsTab) populateFromSettings(s *api.Settings) {
	if s.BRI != nil {
		t.inputs[setFieldBRI].SetValue(strconv.Itoa(*s.BRI))
	}
	if s.ATIME != nil {
		t.inputs[setFieldATIME].SetValue(strconv.Itoa(*s.ATIME))
	}
	if s.TSPEED != nil {
		t.inputs[setFieldTSPEED].SetValue(strconv.Itoa(*s.TSPEED))
	}
	if s.SSPEED != nil {
		t.inputs[setFieldSSPEED].SetValue(strconv.Itoa(*s.SSPEED))
	}
	if s.TFORMAT != "" {
		t.inputs[setFieldTFORMAT].SetValue(s.TFORMAT)
	}
	if s.DFORMAT != "" {
		t.inputs[setFieldDFORMAT].SetValue(s.DFORMAT)
	}
	if s.ABRI != nil {
		t.toggles[setTogABRI] = *s.ABRI
	}
	if s.ATRANS != nil {
		t.toggles[setTogATRANS] = *s.ATRANS
	}
	if s.SOM != nil {
		t.toggles[setTogSOM] = *s.SOM
	}
	if s.CEL != nil {
		t.toggles[setTogCEL] = *s.CEL
	}
	if s.UPPERCASE != nil {
		t.toggles[setTogUPPER] = *s.UPPERCASE
	}
}

func (t *settingsTab) buildSettings() api.Settings {
	s := api.Settings{}
	if v, err := strconv.Atoi(t.inputs[setFieldBRI].Value()); err == nil {
		s.BRI = api.IntPtr(v)
	}
	if v, err := strconv.Atoi(t.inputs[setFieldATIME].Value()); err == nil {
		s.ATIME = api.IntPtr(v)
	}
	if v, err := strconv.Atoi(t.inputs[setFieldTSPEED].Value()); err == nil {
		s.TSPEED = api.IntPtr(v)
	}
	if v, err := strconv.Atoi(t.inputs[setFieldSSPEED].Value()); err == nil {
		s.SSPEED = api.IntPtr(v)
	}
	if v := t.inputs[setFieldTCOL].Value(); v != "" {
		s.TCOL = v
	}
	if v := t.inputs[setFieldTFORMAT].Value(); v != "" {
		s.TFORMAT = v
	}
	if v := t.inputs[setFieldDFORMAT].Value(); v != "" {
		s.DFORMAT = v
	}
	s.ABRI = api.BoolPtr(t.toggles[setTogABRI])
	s.ATRANS = api.BoolPtr(t.toggles[setTogATRANS])
	s.SOM = api.BoolPtr(t.toggles[setTogSOM])
	s.CEL = api.BoolPtr(t.toggles[setTogCEL])
	s.UPPERCASE = api.BoolPtr(t.toggles[setTogUPPER])
	return s
}

func (t *settingsTab) applyCmd() tea.Cmd {
	s := t.buildSettings()
	client := t.client
	return func() tea.Msg {
		return settingsApplyMsg{err: client.SetSettings(s)}
	}
}

func (t *settingsTab) View() string {
	var b strings.Builder
	b.WriteString("\n")

	if !t.loaded {
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			styleMuted.Render("Loading settings…  (r to reload)"),
		))
		b.WriteString("\n")
	}

	// Display section
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("Display") + "\n" +
			formRow("Brightness (0-255)", t.inputs[setFieldBRI].View(), t.focus == setFieldBRI) +
			formRow("App duration (s)", t.inputs[setFieldATIME].View(), t.focus == setFieldATIME) +
			formRow("Transition speed (ms)", t.inputs[setFieldTSPEED].View(), t.focus == setFieldTSPEED) +
			formRow("Scroll speed (%)", t.inputs[setFieldSSPEED].View(), t.focus == setFieldSSPEED) +
			formRow("Global text color", t.inputs[setFieldTCOL].View(), t.focus == setFieldTCOL),
	))
	b.WriteString("\n")

	// Time & Date section
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("Time & Date") + "\n" +
			formRow("Time format", t.inputs[setFieldTFORMAT].View(), t.focus == setFieldTFORMAT) +
			styleMuted.Render("  e.g. %H:%M  %l:%M %p  %H:%M:%S\n") +
			formRow("Date format", t.inputs[setFieldDFORMAT].View(), t.focus == setFieldDFORMAT) +
			styleMuted.Render("  e.g. %d.%m.%y  %m/%d/%y  %y-%m-%d\n"),
	))
	b.WriteString("\n")

	// Toggles
	togLabels := []string{"Auto brightness", "Auto transition", "Week starts Mon", "Celsius", "Uppercase"}
	parts := make([]string, setNumToggles)
	for i, lbl := range togLabels {
		parts[i] = checkbox(t.toggles[i], lbl)
	}
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		strings.Join(parts[:3], "   ") + "\n" +
			strings.Join(parts[3:], "   ") + "\n" +
			styleMuted.Render("(1-5 to toggle when not typing)\n"),
	))
	b.WriteString("\n")

	applyBtn := styleButton.Render("[ Apply Settings ]")
	reloadBtn := styleButton.Render("[ Reload from Device ]")
	if t.focus == setBtnApply {
		applyBtn = styleButtonFocused.Render("[ Apply Settings ]")
	}
	if t.focus == setBtnReload {
		reloadBtn = styleButtonFocused.Render("[ Reload from Device ]")
	}

	const outerPad = 2
	btnY := zoneLine(b.String())
	applyW := lipgloss.Width(applyBtn)
	reloadW := lipgloss.Width(reloadBtn)
	t.zones = []clickZone{
		{YMin: btnY, YMax: btnY, XMin: outerPad, XMax: outerPad + applyW - 1, ID: "apply"},
		{YMin: btnY, YMax: btnY, XMin: outerPad + applyW + 2, XMax: outerPad + applyW + 2 + reloadW - 1, ID: "reload"},
	}

	b.WriteString(lipgloss.NewStyle().PaddingLeft(outerPad).Render(
		applyBtn + "  " + reloadBtn + "   " +
			styleMuted.Render("a=apply  r=reload  Tab=next field"),
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

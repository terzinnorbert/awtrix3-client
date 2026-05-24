package tui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/terzi/awtrix3-client/internal/api"
)

// field indices for the apps form
const (
	appFieldName     = 0
	appFieldText     = 1
	appFieldColor    = 2
	appFieldIcon     = 3
	appFieldDuration = 4
	appFieldRepeat   = 5
	appFieldScroll   = 6
	appFieldBg       = 7
	appFieldEffect   = 8
	appFieldPos      = 9
	appNumInputs     = 10
)

// toggle indices
const (
	appTogRainbow  = 0
	appTogCenter   = 1
	appTogNoScroll = 2
	appTogTopText  = 3
	appTogSave     = 4
	appNumToggles  = 5
)

type appsActionMsg struct{ err error }
type appsLoopMsg struct {
	loop []api.LoopItem
	err  error
}
type appsDeleteMsg struct{ err error }

type appsTab struct {
	client      *api.Client
	width       int
	height      int
	inputs      []textinput.Model
	focus       int // 0..appNumInputs-1 = inputs; appNumInputs = push btn; +1 = del btn
	toggles     [appNumToggles]bool
	loop        []api.LoopItem
	loopCursor  int // selected row in the App Loop list
	zones       []clickZone
	err         string
	success     string
	jsonPreview string
}

func newAppsTab(client *api.Client) *appsTab {
	specs := []struct {
		placeholder string
		width       int
		limit       int
	}{
		{"myapp", 20, 64},
		{"Hello World!", 30, 256},
		{"#FFFFFF", 10, 16},
		{"1234", 8, 32},
		{"5", 5, 5},
		{"-1", 5, 5},
		{"100", 5, 5},
		{"#000000", 10, 16},
		{"", 20, 64},
		{"0", 5, 5},
	}
	inputs := make([]textinput.Model, len(specs))
	for i, s := range specs {
		ti := textinput.New()
		ti.Placeholder = s.placeholder
		ti.Width = s.width
		ti.CharLimit = s.limit
		inputs[i] = ti
	}
	// Focus the name field initially.
	inputs[appFieldName].Focus()

	t := &appsTab{
		client: client,
		inputs: inputs,
		toggles: [appNumToggles]bool{
			false, // rainbow
			true,  // center
			false, // noScroll
			false, // topText
			false, // save
		},
	}
	t.updatePreview()
	return t
}

func (t *appsTab) Title() string { return "Apps" }

func (t *appsTab) InputFocused() bool {
	for _, in := range t.inputs {
		if in.Focused() {
			return true
		}
	}
	return false
}

func (t *appsTab) SetSize(w, h int) { t.width = w; t.height = h }

func (t *appsTab) Init() tea.Cmd {
	return t.loadLoop()
}

func (t *appsTab) loadLoop() tea.Cmd {
	client := t.client
	return func() tea.Msg {
		loop, err := client.GetLoop()
		return appsLoopMsg{loop: loop, err: err}
	}
}

func (t appsTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case appsLoopMsg:
		if msg.err == nil {
			t.loop = msg.loop
		}
	case appsActionMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = "App pushed successfully"
			t.err = ""
			cmds = append(cmds, t.loadLoop())
		}
	case appsDeleteMsg:
		if msg.err != nil {
			t.err = msg.err.Error()
			t.success = ""
		} else {
			t.success = "App deleted"
			t.err = ""
			cmds = append(cmds, t.loadLoop())
		}

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
			if msg.Action == tea.MouseActionPress {
				switch hitZone(t.zones, msg.X, msg.Y) {
				case "push":
					return &t, t.pushCmd()
				case "delete":
					return &t, t.deleteCmd()
				}
			}
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
				t.focus = (t.focus + 1) % (appNumInputs + 2)
				if t.focus < appNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "shift+tab", "up":
				t.blurAll()
				t.focus = (t.focus - 1 + appNumInputs + 2) % (appNumInputs + 2)
				if t.focus < appNumInputs {
					cmds = append(cmds, t.inputs[t.focus].Focus())
				}
				return &t, tea.Batch(cmds...)
			case "enter":
				if t.focus == appNumInputs {
					return &t, t.pushCmd()
				}
				if t.focus == appNumInputs+1 {
					return &t, t.deleteCmd()
				}
			}
			// Forward to focused input.
			for i := range t.inputs {
				if t.inputs[i].Focused() {
					var cmd tea.Cmd
					t.inputs[i], cmd = t.inputs[i].Update(msg)
					t.updatePreview()
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
			if t.focus == appNumInputs {
				return &t, t.pushCmd()
			}
			if t.focus == appNumInputs+1 {
				return &t, t.deleteCmd()
			}
			// Load cursor-selected app name into form.
			if len(t.loop) > 0 {
				t.inputs[appFieldName].SetValue(t.loop[t.loopCursor].Name)
				t.updatePreview()
			}
		case "j", "down":
			if len(t.loop) > 0 {
				t.loopCursor = (t.loopCursor + 1) % len(t.loop)
			}
		case "k", "up":
			if len(t.loop) > 0 {
				t.loopCursor = (t.loopCursor - 1 + len(t.loop)) % len(t.loop)
			}
		// Toggle shortcuts (when not in input mode).
		case "1":
			t.toggles[appTogRainbow] = !t.toggles[appTogRainbow]
			t.updatePreview()
		case "2":
			t.toggles[appTogCenter] = !t.toggles[appTogCenter]
			t.updatePreview()
		case "3":
			t.toggles[appTogNoScroll] = !t.toggles[appTogNoScroll]
			t.updatePreview()
		case "4":
			t.toggles[appTogTopText] = !t.toggles[appTogTopText]
			t.updatePreview()
		case "5":
			t.toggles[appTogSave] = !t.toggles[appTogSave]
			t.updatePreview()
		case "p":
			return &t, t.pushCmd()
		case "d":
			return &t, t.deleteCmd()
		case "r":
			return &t, t.loadLoop()
		}
	}

	return &t, tea.Batch(cmds...)
}

func (t *appsTab) blurAll() {
	for i := range t.inputs {
		t.inputs[i].Blur()
	}
}

func (t *appsTab) pushCmd() tea.Cmd {
	name := strings.TrimSpace(t.inputs[appFieldName].Value())
	if name == "" {
		name = "myapp"
	}
	app := t.buildApp()
	client := t.client
	return func() tea.Msg {
		return appsActionMsg{err: client.PushCustomApp(name, app)}
	}
}

func (t *appsTab) deleteCmd() tea.Cmd {
	name := strings.TrimSpace(t.inputs[appFieldName].Value())
	if name == "" {
		return nil
	}
	client := t.client
	return func() tea.Msg {
		return appsDeleteMsg{err: client.DeleteCustomApp(name)}
	}
}

func (t *appsTab) buildApp() api.CustomApp {
	app := api.CustomApp{}
	if v := t.inputs[appFieldText].Value(); v != "" {
		app.Text = v
	}
	if v := t.inputs[appFieldColor].Value(); v != "" {
		app.Color = v
	}
	if v := t.inputs[appFieldIcon].Value(); v != "" {
		app.Icon = v
	}
	if v := t.inputs[appFieldBg].Value(); v != "" {
		app.Background = v
	}
	if v := t.inputs[appFieldEffect].Value(); v != "" {
		app.Effect = v
	}
	if v, err := strconv.Atoi(t.inputs[appFieldDuration].Value()); err == nil {
		app.Duration = api.IntPtr(v)
	}
	if v, err := strconv.Atoi(t.inputs[appFieldRepeat].Value()); err == nil {
		app.Repeat = api.IntPtr(v)
	}
	if v, err := strconv.Atoi(t.inputs[appFieldScroll].Value()); err == nil {
		app.ScrollSpeed = api.IntPtr(v)
	}
	if v, err := strconv.Atoi(t.inputs[appFieldPos].Value()); err == nil {
		app.Pos = api.IntPtr(v)
	}
	if t.toggles[appTogRainbow] {
		app.Rainbow = api.BoolPtr(true)
	}
	if !t.toggles[appTogCenter] {
		app.Center = api.BoolPtr(false)
	}
	if t.toggles[appTogNoScroll] {
		app.NoScroll = api.BoolPtr(true)
	}
	if t.toggles[appTogTopText] {
		app.TopText = api.BoolPtr(true)
	}
	if t.toggles[appTogSave] {
		app.Save = api.BoolPtr(true)
	}
	return app
}

func (t *appsTab) updatePreview() {
	app := t.buildApp()
	data, _ := json.MarshalIndent(app, "  ", "  ")
	t.jsonPreview = string(data)
}

func (t *appsTab) View() string {
	var b strings.Builder
	b.WriteString("\n")

	// Form
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("App Builder") + "\n" +
			formRow("App Name", t.inputs[appFieldName].View(), t.focus == appFieldName) +
			formRow("Text", t.inputs[appFieldText].View(), t.focus == appFieldText) +
			formRow("Color", t.inputs[appFieldColor].View(), t.focus == appFieldColor) +
			formRow("Icon", t.inputs[appFieldIcon].View(), t.focus == appFieldIcon) +
			formRowInline(
				"Duration", t.inputs[appFieldDuration].View(),
				"Repeat", t.inputs[appFieldRepeat].View(),
				"Scroll%", t.inputs[appFieldScroll].View(),
			) +
			formRowInline(
				"Background", t.inputs[appFieldBg].View(),
				"Effect", t.inputs[appFieldEffect].View(),
				"Position", t.inputs[appFieldPos].View(),
			) +
			"\n" +
			toggleRow(t.toggles, t.focus) +
			"\n",
	))

	// JSON preview
	previewLines := strings.Split(t.jsonPreview, "\n")
	maxLines := 5
	preview := ""
	for i, l := range previewLines {
		if i >= maxLines {
			preview += styleMuted.Render(fmt.Sprintf("  … %d more lines", len(previewLines)-maxLines))
			break
		}
		preview += styleMuted.Render(l) + "\n"
	}
	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
		styleSectionTitle.Render("JSON Preview") + "\n" +
			styleBox.Width(60).Render(preview),
	))
	b.WriteString("\n\n")

	// Buttons — record click zones before rendering.
	pushBtn := styleButton.Render("[ Push Custom App ]")
	delBtn := styleButtonDanger.Render("[ Delete App ]")
	if t.focus == appNumInputs {
		pushBtn = styleButtonFocused.Render("[ Push Custom App ]")
	}
	if t.focus == appNumInputs+1 {
		delBtn = styleButtonDangerFocused.Render("[ Delete App ]")
	}

	const outerPad = 2
	btnY := zoneLine(b.String())
	pushW := lipgloss.Width(pushBtn)
	delW := lipgloss.Width(delBtn)
	t.zones = []clickZone{
		{YMin: btnY, YMax: btnY, XMin: outerPad, XMax: outerPad + pushW - 1, ID: "push"},
		{YMin: btnY, YMax: btnY, XMin: outerPad + pushW + 2, XMax: outerPad + pushW + 2 + delW - 1, ID: "delete"},
	}

	b.WriteString(lipgloss.NewStyle().PaddingLeft(outerPad).Render(
		pushBtn + "  " + delBtn + "   " +
			styleMuted.Render("p=push  d=delete  Tab=next field  j/k=cursor  Enter=load"),
	))
	b.WriteString("\n")

	if t.success != "" {
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(styleSuccess.Render("✓ " + t.success)))
		b.WriteString("\n")
	}
	if t.err != "" {
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(styleError.Render("✗ " + t.err)))
		b.WriteString("\n")
	}

	// App loop
	if len(t.loop) > 0 {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(styleSectionTitle.Render("App Loop") + "\n"))
		b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
			styleMuted.Render(fmt.Sprintf("%-4s  %-20s\n", "#", "Name")),
		))
		for i, item := range t.loop {
			if i >= 10 {
				b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
					styleMuted.Render(fmt.Sprintf("… %d more entries (r to reload)", len(t.loop)-10))))
				b.WriteString("\n")
				break
			}
			cursor := "  "
			if i == t.loopCursor {
				cursor = styleAccent("► ")
			}
			b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(
				cursor+
					styleMuted.Render(fmt.Sprintf("%-4d", i))+
					styleValue.Render(item.Name),
			))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func formRow(label, input string, focused bool) string {
	lbl := styleLabel.Render(label + ":")
	indicator := " "
	if focused {
		indicator = styleAccent("›")
	}
	return indicator + " " + lbl + " " + input + "\n"
}

func formRowInline(label1, in1, label2, in2, label3, in3 string) string {
	l := func(l, v string) string {
		return styleLabel.Width(12).Render(l+":") + " " + v
	}
	return "  " + l(label1, in1) + "  " + l(label2, in2) + "  " + l(label3, in3) + "\n"
}

func toggleRow(toggles [appNumToggles]bool, focus int) string {
	labels := []string{"Rainbow", "Center", "NoScroll", "TopText", "Save to flash"}
	parts := make([]string, len(labels))
	for i, lbl := range labels {
		parts[i] = checkbox(toggles[i], lbl)
	}
	return "  " + strings.Join(parts, "   ") + "\n" +
		styleMuted.Render("  (when not typing: 1-5 toggles the flags above)") + "\n"
}

func styleAccent(s string) string {
	return lipgloss.NewStyle().Foreground(colorAccent).Render(s)
}

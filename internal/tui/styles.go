package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent  = lipgloss.Color("#7C3AED")
	colorSuccess = lipgloss.Color("#10B981")
	colorError   = lipgloss.Color("#EF4444")
	colorWarning = lipgloss.Color("#F59E0B")
	colorMuted   = lipgloss.Color("#9CA3AF")
	colorText    = lipgloss.Color("#F9FAFB")
	colorBorder  = lipgloss.Color("#374151")
	colorDim     = lipgloss.Color("#4B5563")

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText)

	styleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Underline(true)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(colorMuted)

	styleSection = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(colorAccent).
			PaddingLeft(1).
			MarginBottom(1)

	styleSectionTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent).
				MarginBottom(1)

	styleLabel = lipgloss.NewStyle().
			Foreground(colorMuted).
			Width(18)

	styleValue = lipgloss.NewStyle().
			Foreground(colorText)

	styleButton = lipgloss.NewStyle().
			Background(colorAccent).
			Foreground(colorText).
			Padding(0, 2).
			Bold(true)

	styleButtonFocused = lipgloss.NewStyle().
				Background(lipgloss.Color("#9333EA")).
				Foreground(colorText).
				Padding(0, 2).
				Bold(true).
				Underline(true)

	styleButtonDanger = lipgloss.NewStyle().
				Background(colorError).
				Foreground(colorText).
				Padding(0, 2).
				Bold(true)

	styleButtonDangerFocused = lipgloss.NewStyle().
					Background(lipgloss.Color("#DC2626")).
					Foreground(colorText).
					Padding(0, 2).
					Bold(true).
					Underline(true)

	styleSuccess = lipgloss.NewStyle().Foreground(colorSuccess)
	styleError   = lipgloss.NewStyle().Foreground(colorError)
	styleWarning = lipgloss.NewStyle().Foreground(colorWarning)
	styleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	styleDim     = lipgloss.NewStyle().Foreground(colorDim)

	styleFooter = lipgloss.NewStyle().
			Foreground(colorMuted).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorBorder)

	styleBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)
)

// progressBar renders a simple ASCII progress bar.
// value and max define the fill fraction; width is the total bar width in chars.
func progressBar(value, max, width int) string {
	if max <= 0 || width <= 0 {
		return ""
	}
	if value > max {
		value = max
	}
	filled := value * width / max
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

// checkbox renders a checkbox label.
func checkbox(checked bool, label string) string {
	if checked {
		return styleSuccess.Render("[✓]") + " " + label
	}
	return styleDim.Render("[ ]") + " " + label
}

// radioButton renders a single radio option.
func radioButton(selected bool, label string) string {
	if selected {
		return styleSuccess.Render("(●)") + " " + label
	}
	return styleDim.Render("( )") + " " + label
}

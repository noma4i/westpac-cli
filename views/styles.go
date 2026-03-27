package views

import "github.com/charmbracelet/lipgloss"

var (
	// Westpac brand colors
	WpacRed      = lipgloss.Color("#DA1710")
	WpacDarkRed  = lipgloss.Color("#9E1209")
	WpacWhite    = lipgloss.Color("#FFFFFF")
	WpacGray     = lipgloss.Color("#F2F2F2")
	WpacDarkGray = lipgloss.Color("#666666")
	WpacBlack    = lipgloss.Color("#1A1A1A")
	WpacGreen    = lipgloss.Color("#008A00")
	WpacNegRed   = lipgloss.Color("#D0021B")
	WpacYellow   = lipgloss.Color("#FFA500")

	// Base styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WpacWhite).
			Background(WpacRed).
			Padding(0, 2)

	SubtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WpacRed)

	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WpacWhite).
			Background(WpacDarkRed).
			Padding(0, 1)

	NormalStyle = lipgloss.NewStyle().
			Padding(0, 1)

	DimStyle = lipgloss.NewStyle().
			Foreground(WpacDarkGray)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(WpacNegRed).
			Bold(true)

	PositiveStyle = lipgloss.NewStyle().
			Foreground(WpacGreen).
			Bold(true)

	NegativeStyle = lipgloss.NewStyle().
			Foreground(WpacNegRed)

	HelpStyle = lipgloss.NewStyle().
			Foreground(WpacDarkGray).
			Padding(1, 0)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(WpacRed).
			Padding(1, 2)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WpacWhite).
			Background(WpacRed).
			Padding(0, 1).
			MarginBottom(1)

	LabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WpacRed).
			Width(16)

	ValueStyle = lipgloss.NewStyle().
			Foreground(WpacGray)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(WpacRed)
)

func MoneyStyle(amount float64) lipgloss.Style {
	if amount >= 0 {
		return PositiveStyle
	}
	return NegativeStyle
}

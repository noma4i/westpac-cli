package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type KeyBinding struct {
	Key  string
	Desc string
}

func RenderHelp(bindings []KeyBinding, width int) string {
	var parts []string
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(WpacRed)
	greenStyle := lipgloss.NewStyle().Bold(true).Foreground(WpacGreen)
	descStyle := lipgloss.NewStyle().Foreground(WpacDarkGray)

	for _, b := range bindings {
		if b.Key == "Copied!" {
			parts = append(parts, greenStyle.Render(b.Key))
		} else if b.Desc == "" {
			parts = append(parts, keyStyle.Render(b.Key))
		} else {
			parts = append(parts, keyStyle.Render(b.Key)+" "+descStyle.Render(b.Desc))
		}
	}

	help := strings.Join(parts, "  |  ")
	return HelpStyle.Width(width).Render(help)
}

var LoginHelp = []KeyBinding{
	{"Tab", "next"},
	{"Enter", "login"},
	{"Esc Esc", "quit"},
}

var AccountsHelp = []KeyBinding{
	{"Up/Down", "navigate"},
	{"Enter", "transactions"},
	{"Tab", "summary"},
	{"R", "refresh"},
	{"L", "logout"},
	{"Esc Esc", "quit"},
}

var TransactionsHelp = []KeyBinding{
	{"Up/Down", "navigate"},
	{"Enter", "details"},
	{"type", "search"},
	{"Esc", "back"},
}

var TransactionDetailHelp = []KeyBinding{
	{"Up/Down", "scroll"},
	{"C", "copy"},
	{"Esc", "back"},
}

var TransactionDetailCopiedHelp = []KeyBinding{
	{"Up/Down", "scroll"},
	{"Copied!", ""},
	{"Esc", "back"},
}

var SummaryHelp = []KeyBinding{
	{"Tab", "accounts"},
	{"R", "refresh"},
	{"Esc", "back"},
}

package views

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/noma4i/westpac-cli/utils"
)

func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle
	return s
}

func RenderError(msg string) string {
	return ErrorStyle.Render("Error: " + msg)
}

func RenderLoading(s spinner.Model, msg string) string {
	return fmt.Sprintf("%s %s", s.View(), msg)
}

func RenderMoney(amount float64) string {
	formatted := utils.FormatMoney(amount, "AUD")
	return MoneyStyle(amount).Render(formatted)
}

func RenderKeyValue(label, value string) string {
	return LabelStyle.Render(label) + "  " + ValueStyle.Render(value)
}

func RenderHeader(title string, width int) string {
	if width < 10 {
		width = 10
	}
	return HeaderStyle.Width(width).Render(title)
}

func RenderBox(content string, width int) string {
	style := BorderStyle.Width(width - 4) // account for border + padding
	return style.Render(content)
}

func CenterHorizontal(s string, width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, s)
}

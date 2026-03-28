package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/noma4i/westpac-cli/models"
	"github.com/noma4i/westpac-cli/utils"
)

type AccountsModel struct {
	accounts []models.Account
	cursor   int
	loading  bool
	errorMsg string
	spinner  spinner.Model
	width    int
	height   int
}

func NewAccountsModel() AccountsModel {
	return AccountsModel{
		spinner: NewSpinner(),
	}
}

type AccountSelectedMsg struct {
	Account models.Account
}

func (m AccountsModel) Update(msg tea.Msg) (AccountsModel, tea.Cmd) {
	if m.loading {
		if smsg, ok := msg.(spinner.TickMsg); ok {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(smsg)
			return m, cmd
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.accounts)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.accounts) > 0 && m.cursor < len(m.accounts) {
				return m, func() tea.Msg {
					return AccountSelectedMsg{Account: m.accounts[m.cursor]}
				}
			}
		}
	}
	return m, nil
}

func (m AccountsModel) View() string {
	var b strings.Builder

	b.WriteString(RenderHeader(" Accounts ", m.width))
	b.WriteString("\n")

	if m.loading {
		b.WriteString(RenderLoading(m.spinner, "Loading accounts..."))
		return b.String()
	}

	if m.errorMsg != "" {
		b.WriteString(RenderError(m.errorMsg))
		b.WriteString("\n\n")
			return b.String()
	}

	if len(m.accounts) == 0 {
		b.WriteString(DimStyle.Render("No accounts found"))
		return b.String()
	}

	// Account list
	nameWidth := 30
	balWidth := 15

	for i, acc := range m.accounts {
		name := utils.TruncateString(acc.DisplayName, nameWidth)
		balStr := utils.FormatMoney(acc.CurrentBalance, "AUD")

		accNum := ""
		if acc.BSB != "" {
			if utils.MaskMode {
				accNum = fmt.Sprintf("  %s %s", utils.MaskStars(acc.BSB), utils.MaskStars(acc.AccountNumber))
			} else {
				accNum = fmt.Sprintf("  %s %s", acc.BSB, acc.AccountNumber)
			}
		}

		line := fmt.Sprintf("%-*s  %*s%s", nameWidth, name, balWidth, balStr, accNum)

		if i == m.cursor {
			b.WriteString(SelectedStyle.Width(m.width).Render(line))
		} else {
			namePart := fmt.Sprintf("%-*s  ", nameWidth, name)
			balPart := MoneyStyle(acc.CurrentBalance).Render(fmt.Sprintf("%*s", balWidth, balStr))
			b.WriteString(NormalStyle.Render(namePart + balPart + DimStyle.Render(accNum)))
		}
		b.WriteString("\n")
	}

	// Total
	if len(m.accounts) > 1 {
		b.WriteString("\n")
		total := 0.0
		for _, acc := range m.accounts {
			total += acc.CurrentBalance
		}
		totalLine := fmt.Sprintf("%-*s  %*s", nameWidth, "Total", balWidth, utils.FormatMoney(total, "AUD"))
		b.WriteString(lipgloss.NewStyle().Bold(true).Padding(0, 1).Render(totalLine))
	}

	b.WriteString("\n")

	return b.String()
}

func (m AccountsModel) SetAccounts(accounts []models.Account) AccountsModel {
	m.accounts = accounts
	m.loading = false
	m.errorMsg = ""
	if m.cursor >= len(accounts) {
		m.cursor = 0
	}
	return m
}

func (m AccountsModel) SetLoading(loading bool) AccountsModel {
	m.loading = loading
	return m
}

func (m AccountsModel) SetError(err error) AccountsModel {
	if err != nil {
		m.errorMsg = fmt.Sprintf("%v", err)
	} else {
		m.errorMsg = ""
	}
	m.loading = false
	return m
}

func (m AccountsModel) SetSize(w, h int) AccountsModel {
	m.width = w
	m.height = h
	return m
}

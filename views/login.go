package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	fieldCustomerID = iota
	fieldPassword
	fieldSubmit
)

type LoginModel struct {
	inputs     []textinput.Model
	focusIndex int
	loading    bool
	errorMsg   string
	spinner    spinner.Model
	width      int
	height     int
}

func NewLoginModel() LoginModel {
	inputs := make([]textinput.Model, 2)

	inputs[fieldCustomerID] = textinput.New()
	inputs[fieldCustomerID].Placeholder = "Customer ID"
	inputs[fieldCustomerID].CharLimit = 20
	inputs[fieldCustomerID].Focus()

	inputs[fieldPassword] = textinput.New()
	inputs[fieldPassword].Placeholder = "Password"
	inputs[fieldPassword].EchoMode = textinput.EchoPassword
	inputs[fieldPassword].CharLimit = 64

	return LoginModel{
		inputs:     inputs,
		focusIndex: 0,
		spinner:    NewSpinner(),
	}
}

func (m LoginModel) Update(msg tea.Msg) (LoginModel, tea.Cmd) {
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
		m.errorMsg = ""
		key := msg.String()

		switch key {
		case "enter":
			if m.focusIndex == fieldSubmit {
				return m, m.submit()
			}
			m.focusIndex++
			return m, m.syncFocus()
		case "tab", "down":
			m.focusIndex++
			if m.focusIndex > fieldSubmit {
				m.focusIndex = 0
			}
			return m, m.syncFocus()
		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = fieldSubmit
			}
			return m, m.syncFocus()
		}
	}

	// Only update the focused input
	if m.focusIndex < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *LoginModel) syncFocus() tea.Cmd {
	var cmd tea.Cmd
	for i := range m.inputs {
		if i == m.focusIndex {
			cmd = m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
	return cmd
}

type LoginSubmitMsg struct {
	CustomerID string
	Password   string
}

func (m LoginModel) submit() tea.Cmd {
	cid := strings.TrimSpace(m.inputs[fieldCustomerID].Value())
	pw := strings.TrimSpace(m.inputs[fieldPassword].Value())

	if cid == "" || pw == "" {
		return nil
	}

	return func() tea.Msg {
		return LoginSubmitMsg{CustomerID: cid, Password: pw}
	}
}

func (m LoginModel) CustomerID() string {
	return strings.TrimSpace(m.inputs[fieldCustomerID].Value())
}

func (m LoginModel) View() string {
	var b strings.Builder

	logo := lipgloss.NewStyle().
		Bold(true).
		Foreground(WpacWhite).
		Background(WpacRed).
		Padding(0, 3).
		Render(" WESTPAC ")

	b.WriteString("\n")
	b.WriteString(CenterHorizontal(logo, m.width))
	b.WriteString("\n\n")
	b.WriteString(CenterHorizontal(DimStyle.Render("Terminal Banking Client"), m.width))
	b.WriteString("\n\n\n")

	if m.loading {
		loading := RenderLoading(m.spinner, "Logging in...")
		b.WriteString(CenterHorizontal(loading, m.width))
		return b.String()
	}

	inputWidth := 40
	if m.width > 0 && m.width < 50 {
		inputWidth = m.width - 10
	}
	if inputWidth < 20 {
		inputWidth = 20
	}
	inputStyle := lipgloss.NewStyle().Width(inputWidth)

	label1 := LabelStyle.Render("Customer ID:")
	field1 := inputStyle.Render(m.inputs[fieldCustomerID].View())

	label2 := LabelStyle.Render("Password:")
	field2 := inputStyle.Render(m.inputs[fieldPassword].View())

	var btnStyle lipgloss.Style
	if m.focusIndex == fieldSubmit {
		btnStyle = lipgloss.NewStyle().
			Padding(1, 4).
			Bold(true).
			Foreground(WpacWhite).
			Background(WpacRed)
	} else {
		btnStyle = lipgloss.NewStyle().
			Padding(0, 3).
			Border(lipgloss.NormalBorder()).
			BorderForeground(WpacDarkGray).
			Foreground(WpacRed)
	}
	button := btnStyle.Render("Login")

	form := lipgloss.JoinVertical(lipgloss.Left,
		label1,
		field1,
		"",
		label2,
		field2,
		"",
		button,
	)

	b.WriteString(CenterHorizontal(form, m.width))

	if m.errorMsg != "" {
		b.WriteString("\n\n")
		b.WriteString(CenterHorizontal(RenderError(m.errorMsg), m.width))
	}

	return b.String()
}

func (m LoginModel) SetSize(w, h int) LoginModel {
	m.width = w
	m.height = h
	return m
}

func (m LoginModel) SetLoading(loading bool) LoginModel {
	m.loading = loading
	return m
}

func (m LoginModel) SetError(err error) LoginModel {
	if err != nil {
		m.errorMsg = fmt.Sprintf("%v", err)
	} else {
		m.errorMsg = ""
	}
	return m
}
